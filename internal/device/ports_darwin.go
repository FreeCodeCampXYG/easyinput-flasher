//go:build darwin

package device

import (
	"context"
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/FreeCodeCampXYG/easyinput-flasher/internal/domain"
	"go.bug.st/serial"
)

// ListPorts 枚举 macOS 串口和正常模式 USB HID；串口先作为下载候选，后续仍由 ESP32-S3 ROM 验身确认。
func ListPorts() ([]domain.DeviceInfo, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	devices := make([]domain.DeviceInfo, 0)
	probeCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if hasExpectedMacDevice(probeCtx) {
		devices = append(devices, domain.DeviceInfo{ID: "macos-usb-hid", Port: "正常模式", Label: "EasyInput AI（正常模式）", Mode: "normal", ObservedAt: now})
	}
	ports, err := serial.GetPortsList()
	if err != nil && len(devices) == 0 {
		return nil, fmt.Errorf("读取串口列表失败: %w", err)
	}
	sort.Strings(ports)
	for _, port := range ports {
		devices = append(devices, domain.DeviceInfo{ID: port, Port: port, Label: port, Mode: "download", ObservedAt: now})
	}
	return devices, nil
}

func hasExpectedMacDevice(ctx context.Context) bool {
	if hasExpectedUSBHID(ctx) {
		return true
	}
	command := exec.CommandContext(ctx, "system_profiler", "SPBluetoothDataType", "-json")
	output, err := command.Output()
	if err != nil {
		return false
	}
	return strings.Contains(strings.ToLower(string(output)), "easyinput ai")
}

func hasExpectedUSBHID(ctx context.Context) bool {
	command := exec.CommandContext(ctx, "ioreg", "-p", "IOUSB", "-l", "-w", "0")
	output, err := command.Output()
	if err != nil {
		return false
	}
	text := strings.ToLower(string(output))
	// macOS ioreg 可能以十六进制或十进制展示 ID；同时匹配产品名，避免把普通 USB 设备当成键盘。
	hasVendor := strings.Contains(text, "0x303a") || strings.Contains(text, "303a") || strings.Contains(text, "12346")
	hasProduct := strings.Contains(text, "0x1006") || strings.Contains(text, "1006") || strings.Contains(text, "4102")
	return hasVendor && hasProduct && strings.Contains(text, "easyinput")
}
