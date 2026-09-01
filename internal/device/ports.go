//go:build !windows && !darwin

package device

import (
	"fmt"
	"sort"
	"time"

	"github.com/FreeCodeCampXYG/easyinput-flasher/internal/domain"
	"go.bug.st/serial"
)

// ListPorts 只枚举宿主串口，不连接设备也不发送复位或写入指令。
func ListPorts() ([]domain.DeviceInfo, error) {
	ports, err := serial.GetPortsList()
	if err != nil {
		return nil, fmt.Errorf("读取串口列表失败: %w", err)
	}
	sort.Strings(ports)
	now := time.Now().UTC().Format(time.RFC3339)
	devices := make([]domain.DeviceInfo, 0, len(ports))
	for _, port := range ports {
		devices = append(devices, domain.DeviceInfo{ID: port, Port: port, Label: port, Mode: "serial-candidate", ObservedAt: now})
	}
	return devices, nil
}
