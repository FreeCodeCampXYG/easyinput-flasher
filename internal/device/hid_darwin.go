//go:build darwin

package device

import "context"

// HasExpectedHID 只确认 EasyInput 正常模式 USB HID 已重新枚举，不读取用户输入或设备私有数据。
func HasExpectedHID(ctx context.Context) (bool, error) {
	return hasExpectedMacDevice(ctx), nil
}
