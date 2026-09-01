package firmware

import (
	"context"
	"crypto/sha256"
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
		Size               int64  `json:"size"`
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
		if release.Draft || release.Prerelease || (!hasManifest(release) && !hasFactoryAsset(release)) {
			continue
		}
		isFactory := !hasManifest(release) && hasFactoryAsset(release)
		manifest := domain.FirmwareManifest{}
		if isFactory {
			name, _, size, _ := factoryAsset(release)
			manifest = domain.FirmwareManifest{SchemaVersion: 1, Product: "easyinput-factory", Board: domain.BoardID, Chip: domain.ChipType, Tag: release.TagName, IDFVersion: "未知（Factory 镜像）", Files: []domain.FlashFile{{Name: name, Offset: "0x0", Size: size}}}
		}
		result = append(result, domain.FirmwareRelease{
			ID: fmt.Sprintf("%d", release.ID), Repository: source.Repository, Tag: release.TagName,
			Name: release.Name, PublishedAt: release.PublishedAt, Trusted: source.Trusted, Manifest: manifest, IsFactory: isFactory,
		})
	}
	return result, nil
}

// AuditSource 只读取公开仓库的发布合同，帮助贡献者在接入前补齐自动构建与发布资产。
func (c *GitHubClient) AuditSource(ctx context.Context, repository string) (domain.FirmwareSourceAudit, error) {
	audit := domain.FirmwareSourceAudit{Repository: repository, Checks: make([]domain.FirmwareSourceCheck, 0, 3)}
	releases, err := c.listGitHubReleases(ctx, repository)
	if err != nil {
		return audit, err
	}
	valid := 0
	for _, release := range releases {
		if !release.Draft && !release.Prerelease && hasRequiredFirmwareAssets(release) {
			valid++
		}
	}
	audit.ValidReleases = valid
	audit.Checks = append(audit.Checks, domain.FirmwareSourceCheck{
		Name: "可烧录 Release", Passed: valid > 0,
		Message: fmt.Sprintf("发现 %d 个同时包含 manifest 与三段镜像的公开 Release", valid),
	})
	factoryOnly := false
	for _, release := range releases {
		if release.Draft || release.Prerelease || hasRequiredFirmwareAssets(release) {
			continue
		}
		for _, asset := range release.Assets {
			if strings.HasSuffix(strings.ToLower(asset.Name), ".bin") && strings.Contains(strings.ToLower(asset.Name), "factory") {
				factoryOnly = true
				break
			}
		}
	}
	if factoryOnly {
		// 整片 factory 镜像可能覆盖配置分区，不能在桌面端猜测拆分或扩大写入范围。
		audit.Checks = append(audit.Checks, domain.FirmwareSourceCheck{
			Name: "整片 factory.bin", Passed: false,
			Message: "已发现整片镜像，但不能安全自动拆分；请用仓库 Actions 生成 manifest、bootloader、分区表和应用三段资产",
		})
	}
	for _, file := range []struct{ path, name string }{
		{".github/workflows/firmware-release.yml", "自动发布工作流"},
		{"scripts/build_firmware_manifest.py", "manifest 生成脚本"},
	} {
		present, checkErr := c.repositoryFileExists(ctx, repository, file.path)
		if checkErr != nil {
			return audit, checkErr
		}
		audit.Checks = append(audit.Checks, domain.FirmwareSourceCheck{
			Name: file.name, Passed: present,
			Message: map[bool]string{true: "已发现，可复用自动编译合同", false: "缺失，请按发布教程补充"}[present],
		})
	}
	// Factory-only Release 允许进入“受信来源”候选，但写入时仍要求独立的恢复出厂确认口令。
	audit.Ready = valid > 0 || factoryOnly
	return audit, nil
}

func (c *GitHubClient) listGitHubReleases(ctx context.Context, repository string) ([]githubRelease, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/repos/"+repository+"/releases", nil)
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
	return releases, nil
}

func (c *GitHubClient) repositoryFileExists(ctx context.Context, repository, path string) (bool, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/repos/"+repository+"/contents/"+path, nil)
	if err != nil {
		return false, err
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("User-Agent", "EasyInput-Flasher")
	response, err := c.client.Do(request)
	if err != nil {
		return false, fmt.Errorf("检查仓库构建文件失败: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusNotFound {
		return false, nil
	}
	if response.StatusCode != http.StatusOK {
		return false, fmt.Errorf("检查仓库构建文件失败: HTTP %d", response.StatusCode)
	}
	return true, nil
}

func hasManifest(release githubRelease) bool {
	for _, asset := range release.Assets {
		if asset.Name == "firmware-manifest.json" {
			return true
		}
	}
	return false
}

func factoryAsset(release githubRelease) (string, string, int64, bool) {
	for _, asset := range release.Assets {
		name := strings.ToLower(asset.Name)
		if strings.HasSuffix(name, ".bin") && strings.Contains(name, "factory") {
			return asset.Name, asset.BrowserDownloadURL, asset.Size, true
		}
	}
	return "", "", 0, false
}

func hasFactoryAsset(release githubRelease) bool { _, _, _, ok := factoryAsset(release); return ok }

func sha256Hex(data []byte) string { return fmt.Sprintf("%x", sha256.Sum256(data)) }

func hasRequiredFirmwareAssets(release githubRelease) bool {
	assets := make(map[string]bool, len(release.Assets))
	for _, asset := range release.Assets {
		assets[asset.Name] = true
	}
	return assets["firmware-manifest.json"] && assets["bootloader.bin"] && assets["partition-table.bin"] && assets["easy_input_keyboard.bin"]
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
		_, factoryURL, factorySize, ok := factoryAsset(release)
		if !ok {
			return empty, fmt.Errorf("该 Release 缺少 firmware-manifest.json")
		}
		if err := os.MkdirAll(destination, 0o700); err != nil {
			return empty, err
		}
		if err := c.download(ctx, factoryURL, filepath.Join(destination, "factory.bin")); err != nil {
			return empty, err
		}
		data, err := os.ReadFile(filepath.Join(destination, "factory.bin"))
		if err != nil {
			return empty, err
		}
		return domain.FirmwareManifest{SchemaVersion: 1, Product: "easyinput-factory", Board: domain.BoardID, Chip: domain.ChipType, Tag: tag, IDFVersion: "未知（Factory 镜像）", Files: []domain.FlashFile{{Name: "factory.bin", Offset: "0x0", SHA256: sha256Hex(data), Size: factorySize}}}, nil
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
