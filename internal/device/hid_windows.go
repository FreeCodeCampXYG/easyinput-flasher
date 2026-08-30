//go:build windows

package device

import (
	"context"
	"os/exec"
	"syscall"
)

// HasExpectedHID 仅判断公开的 VID/PID 是否重新枚举，不读取用户输入或设备私有数据。
func HasExpectedHID(ctx context.Context) (bool, error) {
	command := exec.CommandContext(ctx, "powershell.exe", "-NoProfile", "-NonInteractive", "-Command", "Get-PnpDevice -PresentOnly | Where-Object { $_.InstanceId -match 'VID_303A&PID_1006' } | Select-Object -First 1 -ExpandProperty InstanceId")
	command.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	output, err := command.Output()
	if err != nil {
		return false, err
	}
	return len(output) > 0, nil
}
