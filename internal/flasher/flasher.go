package flasher

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/FreeCodeCampXYG/easyinput-flasher/internal/domain"
)

type Runner struct {
	helper string
}

func NewRunner() (*Runner, error) {
	executable, err := os.Executable()
	if err != nil {
		return nil, err
	}
	name := "esptool"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	return &Runner{helper: filepath.Join(filepath.Dir(executable), "tools", "esptool", name)}, nil
}

func (r *Runner) Inspect(ctx context.Context, port string) (string, error) {
	if _, err := os.Stat(r.helper); err != nil {
		return "", fmt.Errorf("烧录辅助程序未就绪: %w", err)
	}
	var results []string
	// esptool 的子命令在不同版本不能假设可串联；逐项只读执行可保持验身证据明确。
	for _, operation := range []string{"chip-id", "read-mac", "flash-id"} {
		output, err := r.run(ctx, "--chip", domain.ChipType, "--port", port, operation)
		if err != nil {
			return strings.Join(results, "\n"), err
		}
		results = append(results, output)
	}
	return strings.Join(results, "\n"), nil
}

// FlashArgs 由清单的固定偏移生成，不接受前端传入的任意命令或路径。
func FlashArgs(port, bundleRoot string, manifest domain.FirmwareManifest) ([]string, error) {
	if manifest.Board != domain.BoardID || manifest.Chip != domain.ChipType || len(manifest.Files) != 3 {
		return nil, fmt.Errorf("固件清单未通过目标设备检查")
	}
	arguments := []string{"--chip", domain.ChipType, "--port", port, "--baud", "460800", "write-flash"}
	for _, file := range manifest.Files {
		arguments = append(arguments, file.Offset, filepath.Join(bundleRoot, file.Name))
	}
	return arguments, nil
}

func (r *Runner) Flash(ctx context.Context, port, bundleRoot string, manifest domain.FirmwareManifest) (string, error) {
	if _, err := os.Stat(r.helper); err != nil {
		return "", fmt.Errorf("烧录辅助程序未就绪: %w", err)
	}
	arguments, err := FlashArgs(port, bundleRoot, manifest)
	if err != nil {
		return "", err
	}
	return r.run(ctx, arguments...)
}

func (r *Runner) run(ctx context.Context, arguments ...string) (string, error) {
	command := exec.CommandContext(ctx, r.helper, arguments...)
	command.SysProcAttr = hiddenWindowAttributes()
	output, err := command.CombinedOutput()
	redacted := redact(string(output))
	if err != nil {
		return redacted, fmt.Errorf("烧录工具执行失败: %w", err)
	}
	return redacted, nil
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
