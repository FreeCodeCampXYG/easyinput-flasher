package flasher

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/FreeCodeCampXYG/easyinput-flasher/internal/domain"
	"tinygo.org/x/espflasher/pkg/espflasher"
)

// Runner 是纯 Go 烧录器适配层；前端只能请求固定流程，不能传入串口命令或偏移。
type Runner struct{}

func NewRunner() (*Runner, error) { return &Runner{}, nil }

func (r *Runner) Inspect(ctx context.Context, port string) (string, error) {
	flasher, err := openROM(ctx, port)
	if err != nil {
		return "", err
	}
	defer flasher.Close()
	mac, err := flasher.MAC()
	if err != nil {
		return "", fmt.Errorf("读取设备 MAC 失败: %w", err)
	}
	if flasher.ChipType() != espflasher.ChipESP32S3 {
		return "", fmt.Errorf("目标端口不是 ESP32-S3，已停止")
	}
	return fmt.Sprintf("Chip type: %s\nMAC: %s", flasher.ChipName(), mac), nil
}

func (r *Runner) Flash(ctx context.Context, port, bundleRoot string, manifest domain.FirmwareManifest, progress func(current, total int)) (string, error) {
	images, err := loadImages(bundleRoot, manifest)
	if err != nil {
		return "", err
	}
	flasher, err := openForFlash(ctx, port)
	if err != nil {
		return "", err
	}
	defer flasher.Close()
	if err := flashWithContext(ctx, flasher, images, progress); err != nil {
		return "", err
	}
	return "纯 Go 烧录器已完成受控三段写入", nil
}

// FlashWithDetails 在保留旧接口的同时报告当前镜像索引和字节进度，供桌面诊断面板展示。
func (r *Runner) FlashWithDetails(ctx context.Context, port, bundleRoot string, manifest domain.FirmwareManifest, progress func(index, current, total int)) (string, error) {
	images, err := loadImages(bundleRoot, manifest)
	if err != nil {
		return "", err
	}
	flasher, err := openForFlash(ctx, port)
	if err != nil {
		return "", err
	}
	defer flasher.Close()
	done := make(chan error, 1)
	go func() {
		done <- flasher.FlashImages(images, func(current, total int) {
			remaining := current
			index := 0
			for index+1 < len(images) && remaining >= len(images[index].Data) {
				remaining -= len(images[index].Data)
				index++
			}
			progress(index, remaining, len(images[index].Data))
		})
	}()
	select {
	case err := <-done:
		if err != nil {
			return "", fmt.Errorf("纯 Go 烧录失败: %w", err)
		}
		return "纯 Go 烧录器已完成受控写入", nil
	case <-ctx.Done():
		_ = flasher.Close()
		<-done
		return "", ctx.Err()
	}
}

func openROM(ctx context.Context, port string) (*espflasher.Flasher, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	options := espflasher.DefaultOptions()
	options.ChipType = espflasher.ChipESP32S3
	options.ResetMode = espflasher.ResetNoReset
	options.SkipStub = true
	options.ConnectAttempts = 3
	return espflasher.New(port, options)
}

func openForFlash(ctx context.Context, port string) (*espflasher.Flasher, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	options := espflasher.DefaultOptions()
	options.ChipType = espflasher.ChipESP32S3
	options.ResetMode = espflasher.ResetNoReset
	options.FlashBaudRate = 460800
	options.ConnectAttempts = 3
	// 用户已手动进入下载模式；内嵌 stub 只用于本次会话的可靠写入与校验，结束后仍要求完整关机再开机。
	options.SkipStub = false
	return espflasher.New(port, options)
}

func flashWithContext(ctx context.Context, flasher *espflasher.Flasher, images []espflasher.ImagePart, progress espflasher.ProgressFunc) error {
	done := make(chan error, 1)
	go func() { done <- flasher.FlashImages(images, progress) }()
	select {
	case err := <-done:
		if err != nil {
			return fmt.Errorf("纯 Go 烧录失败: %w", err)
		}
		return nil
	case <-ctx.Done():
		// 关闭本会话串口可中断阻塞读写；不会清理下载缓存，也不会伪造写入成功状态。
		_ = flasher.Close()
		<-done
		return ctx.Err()
	}
}

func loadImages(bundleRoot string, manifest domain.FirmwareManifest) ([]espflasher.ImagePart, error) {
	if manifest.Product == "easyinput-factory" && len(manifest.Files) == 1 && manifest.Files[0].Name == "factory.bin" && manifest.Files[0].Offset == "0x0" {
		data, err := os.ReadFile(filepath.Join(bundleRoot, "factory.bin"))
		if err != nil {
			return nil, fmt.Errorf("读取 Factory 镜像失败: %w", err)
		}
		return []espflasher.ImagePart{{Data: data, Offset: 0}}, nil
	}
	if manifest.Board != domain.BoardID || manifest.Chip != domain.ChipType || len(manifest.Files) != 3 {
		return nil, fmt.Errorf("固件清单未通过目标设备检查")
	}
	images := make([]espflasher.ImagePart, 0, len(manifest.Files))
	for _, file := range manifest.Files {
		if filepath.Base(file.Name) != file.Name || file.Name == "." || filepath.IsAbs(file.Name) {
			return nil, fmt.Errorf("固件清单包含无效文件名: %s", file.Name)
		}
		offset, err := strconv.ParseUint(file.Offset, 0, 32)
		if err != nil {
			return nil, fmt.Errorf("固件清单偏移无效: %s", file.Offset)
		}
		data, err := os.ReadFile(filepath.Join(bundleRoot, file.Name))
		if err != nil {
			return nil, fmt.Errorf("读取已校验固件段失败: %w", err)
		}
		images = append(images, espflasher.ImagePart{Data: data, Offset: uint32(offset)})
	}
	return images, nil
}

func redact(value string) string {
	// 设备完整 MAC 只用于本轮授权比对；持久日志仅保留脱敏后的尾号。
	parts := strings.Fields(value)
	for index, part := range parts {
		if strings.Count(part, ":") == 5 && len(part) >= 17 {
			parts[index] = "**:**:**:**:" + part[len(part)-5:]
		}
	}
	return strings.Join(parts, " ")
}
