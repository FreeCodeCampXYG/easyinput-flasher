package firmware

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/FreeCodeCampXYG/easyinput-flasher/internal/domain"
)

// ParseManifest 只接受固定板型和芯片，防止社区来源的通用描述绕过硬件合同。
func ParseManifest(data []byte) (domain.FirmwareManifest, error) {
	var manifest domain.FirmwareManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return manifest, fmt.Errorf("解析固件清单失败: %w", err)
	}
	if manifest.SchemaVersion != 1 || manifest.Product != "easyinput-firmware" ||
		manifest.Board != domain.BoardID || manifest.Chip != domain.ChipType {
		return manifest, fmt.Errorf("固件清单不适用于 EasyInput V2.0 / ESP32-S3")
	}
	if manifest.Tag == "" || manifest.Commit == "" || len(manifest.Files) != 3 {
		return manifest, fmt.Errorf("固件清单缺少版本或完整写入范围")
	}
	seen := make(map[string]bool, len(manifest.Files))
	for _, file := range manifest.Files {
		if file.Name == "" || file.Offset == "" || len(file.SHA256) != 64 || file.Size <= 0 || seen[file.Name] {
			return manifest, fmt.Errorf("固件清单包含无效文件记录")
		}
		seen[file.Name] = true
	}
	return manifest, nil
}

// VerifyBundle 在写入前逐文件校验，下载完成不等于固件包可信。
func VerifyBundle(root string, manifest domain.FirmwareManifest) error {
	for _, expected := range manifest.Files {
		path := filepath.Join(root, expected.Name)
		info, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf("缺少固件文件 %s: %w", expected.Name, err)
		}
		if info.Size() != expected.Size {
			return fmt.Errorf("固件文件大小不匹配: %s", expected.Name)
		}
		actual, err := fileSHA256(path)
		if err != nil {
			return err
		}
		if !strings.EqualFold(actual, expected.SHA256) {
			return fmt.Errorf("固件哈希校验失败: %s", expected.Name)
		}
	}
	return nil
}

func fileSHA256(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
