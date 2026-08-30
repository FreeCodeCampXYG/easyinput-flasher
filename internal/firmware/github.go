package firmware

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/FreeCodeCampXYG/easyinput-flasher/internal/config"
	"github.com/FreeCodeCampXYG/easyinput-flasher/internal/domain"
)

type GitHubClient struct {
	client *http.Client
}

type githubRelease struct {
	ID          int64  `json:"id"`
	TagName     string `json:"tag_name"`
	Name        string `json:"name"`
	Draft       bool   `json:"draft"`
	Prerelease  bool   `json:"prerelease"`
	PublishedAt string `json:"published_at"`
	Assets      []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func NewGitHubClient(settings config.Settings) (*GitHubClient, error) {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if settings.ProxyMode == "disabled" {
		transport.Proxy = nil
	} else if settings.ProxyMode == "custom" {
		proxyURL, err := url.Parse(settings.ProxyURL)
		if err != nil || proxyURL.Scheme == "" || proxyURL.Host == "" {
			return nil, fmt.Errorf("自定义代理地址无效")
		}
		transport.Proxy = http.ProxyURL(proxyURL)
	}
	return &GitHubClient{client: &http.Client{Transport: transport, Timeout: 20 * time.Second}}, nil
}

// ListReleases 只读取公开 Release；来源是否能烧录仍由本地信任目录和 manifest 决定。
func (c *GitHubClient) ListReleases(ctx context.Context, source config.Source) ([]domain.FirmwareRelease, error) {
	if !source.Enabled || !strings.Contains(source.Repository, "/") {
		return nil, nil
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/repos/"+source.Repository+"/releases", nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("User-Agent", "EasyInput-Flasher")
	response, err := c.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("读取 GitHub Release 失败: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub Release 请求失败: HTTP %d", response.StatusCode)
	}
	var releases []githubRelease
	if err := json.NewDecoder(response.Body).Decode(&releases); err != nil {
		return nil, err
	}
	result := make([]domain.FirmwareRelease, 0, len(releases))
	for _, release := range releases {
		if release.Draft || release.Prerelease || !hasManifest(release) {
			continue
		}
		result = append(result, domain.FirmwareRelease{
			ID: fmt.Sprintf("%d", release.ID), Repository: source.Repository, Tag: release.TagName,
			Name: release.Name, PublishedAt: release.PublishedAt, Trusted: source.Trusted,
		})
	}
	return result, nil
}

func hasManifest(release githubRelease) bool {
	for _, asset := range release.Assets {
		if asset.Name == "firmware-manifest.json" {
			return true
		}
	}
	return false
}

// DownloadBundle 先下载并解析 manifest，再按其白名单下载三段镜像；Release 中的其他附件不会被执行或写入设备。
func (c *GitHubClient) DownloadBundle(ctx context.Context, repository, tag, destination string) (domain.FirmwareManifest, error) {
	var empty domain.FirmwareManifest
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/repos/"+repository+"/releases/tags/"+url.PathEscape(tag), nil)
	if err != nil {
		return empty, err
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("User-Agent", "EasyInput-Flasher")
	response, err := c.client.Do(request)
	if err != nil {
		return empty, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return empty, fmt.Errorf("读取指定 Release 失败: HTTP %d", response.StatusCode)
	}
	var release githubRelease
	if err := json.NewDecoder(response.Body).Decode(&release); err != nil {
		return empty, err
	}
	assets := make(map[string]string, len(release.Assets))
	for _, asset := range release.Assets {
		assets[asset.Name] = asset.BrowserDownloadURL
	}
	manifestURL := assets["firmware-manifest.json"]
	if manifestURL == "" {
		return empty, fmt.Errorf("该 Release 缺少 firmware-manifest.json")
	}
	if err := os.MkdirAll(destination, 0o700); err != nil {
		return empty, err
	}
	manifestPath := filepath.Join(destination, "firmware-manifest.json")
	if err := c.download(ctx, manifestURL, manifestPath); err != nil {
		return empty, err
	}
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return empty, err
	}
	manifest, err := ParseManifest(data)
	if err != nil {
		return empty, err
	}
	if manifest.Tag != tag {
		return empty, fmt.Errorf("Release 标签与清单标签不一致")
	}
	for _, file := range manifest.Files {
		assetURL := assets[file.Name]
		if assetURL == "" {
			return empty, fmt.Errorf("Release 缺少清单声明的文件: %s", file.Name)
		}
		if err := c.download(ctx, assetURL, filepath.Join(destination, file.Name)); err != nil {
			return empty, err
		}
	}
	if err := VerifyBundle(destination, manifest); err != nil {
		return empty, err
	}
	return manifest, nil
}

func (c *GitHubClient) download(ctx context.Context, source, destination string) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, source, nil)
	if err != nil {
		return err
	}
	response, err := c.client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("下载固件资源失败: HTTP %d", response.StatusCode)
	}
	temporary := destination + ".part"
	output, err := os.OpenFile(temporary, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(output, io.LimitReader(response.Body, 32<<20))
	closeErr := output.Close()
	if copyErr != nil {
		return copyErr
	}
	if closeErr != nil {
		return closeErr
	}
	return os.Rename(temporary, destination)
}
