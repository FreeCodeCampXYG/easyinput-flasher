//go:build windows

package device

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/FreeCodeCampXYG/easyinput-flasher/internal/domain"
	"go.bug.st/serial"
)

type pnpDevice struct {
	InstanceID   string
	FriendlyName string
	Class        string
}

// ListPorts 同时扫描 HID 与下载串口，避免正常固件只枚举 HID 时被误报为未连接。
func ListPorts() ([]domain.DeviceInfo, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	devices := make([]domain.DeviceInfo, 0)
	seen := make(map[string]bool)

	if pnp, err := listPnPDevices(); err == nil {
		for _, item := range pnp {
			id := item.InstanceID
			if seen[id] {
				continue
			}
			seen[id] = true
			mode := "normal"
			if strings.Contains(id, "PID_1001") {
				mode = "download"
			}
			devices = append(devices, domain.DeviceInfo{
				ID: id, Port: friendlyPort(item.FriendlyName), Label: friendlyLabel(item.FriendlyName),
				Mode: mode, Chip: chipForMode(mode), ObservedAt: now,
			})
		}
	}
	ports, err := serial.GetPortsList()
	if err != nil && len(devices) == 0 {
		return nil, fmt.Errorf("读取设备列表失败: %w", err)
	}
	for _, port := range ports {
		if seen[port] {
			continue
		}
		devices = append(devices, domain.DeviceInfo{ID: port, Port: port, Label: port, Mode: "download", ObservedAt: now})
	}
	sort.Slice(devices, func(i, j int) bool { return devices[i].Port < devices[j].Port })
	return devices, nil
}

func chipForMode(mode string) string {
	if mode == "download" {
		return "ESP32-S3"
	}
	// 正常 HID/BLE 只证明设备在线；芯片型号必须等进入下载模式后由 esptool 读取。
	return ""
}

func listPnPDevices() ([]pnpDevice, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// 正常模式可能是原生 USB HID，也可能是已配对的 EasyInput AI BLE；两者都只能作为 BOOT 前确认，不能代替芯片验身。
	command := exec.CommandContext(ctx, "powershell.exe", "-NoProfile", "-NonInteractive", "-Command", "Get-PnpDevice -PresentOnly | Where-Object { $_.InstanceId -match 'VID_303A&PID_100[16]' -or $_.FriendlyName -match 'EasyInput AI' } | Select-Object InstanceId,FriendlyName,Class | ConvertTo-Json -Compress")
	// 设备刷新会频繁执行；隐藏系统 PowerShell 窗口，避免把只读 PnP 扫描打断用户的 BOOT 操作。
	command.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	output, err := command.Output()
	if err != nil {
		return nil, err
	}
	trimmed := strings.TrimSpace(string(output))
	if trimmed == "" {
		return nil, nil
	}
	var many []pnpDevice
	if strings.HasPrefix(trimmed, "{") {
		var one pnpDevice
		if err := json.Unmarshal([]byte(trimmed), &one); err != nil {
			return nil, err
		}
		return []pnpDevice{one}, nil
	}
	if err := json.Unmarshal([]byte(trimmed), &many); err != nil {
		return nil, err
	}
	return many, nil
}

func friendlyPort(name string) string {
	if strings.Contains(strings.ToLower(name), "easyinput") {
		return "正常模式"
	}
	if index := strings.Index(strings.ToUpper(name), "(COM"); index >= 0 {
		end := strings.Index(name[index:], ")")
		if end > 0 {
			return name[index+1 : index+end]
		}
	}
	return "HID"
}

func friendlyLabel(name string) string {
	if name == "" {
		return "EasyInput 设备"
	}
	return name
}
