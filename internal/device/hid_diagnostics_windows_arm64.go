//go:build windows && arm64

package device

import (
	"context"
	"errors"
)

// HIDDiagnostics 是运行态固件回报的最小诊断遥测；不读取用户文本或完整设备身份。
type HIDDiagnostics struct {
	Supported    bool
	Firmware     string
	LastInput    string
	InputEvents  uint32
	EncoderSteps uint32
	BatteryMV    uint16
	BatteryPct   uint8
	BatteryState string
	VIN          int
	Charge       int
	LEDGPIO      int
}

// ReadHIDDiagnostics 在 Windows ARM64 无原生 hidapi 编译器时安全降级；不伪造设备已支持诊断。
func ReadHIDDiagnostics(context.Context) (HIDDiagnostics, error) {
	return HIDDiagnostics{}, errors.New("Windows ARM64 构建未包含 hidapi 诊断读取能力")
}
