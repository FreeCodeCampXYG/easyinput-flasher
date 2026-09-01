package device

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"sync/atomic"

	"github.com/karalabe/hid"
)

const (
	easyInputVID              = 0x303a
	easyInputNormalPID        = 0x1006
	statusRequestReportID     = 0x13
	statusResponseReportID    = 0x11
	statusRequestPayloadLen   = 16
	statusResponsePayloadLen  = 63
	statusResponseHeaderLen   = 13
	statusResponseCommandKind = 0x04
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

// ReadHIDDiagnostics 通过现有 Vendor HID 状态报告读取一次运行态诊断快照。
func ReadHIDDiagnostics(ctx context.Context) (HIDDiagnostics, error) {
	infos := hid.Enumerate(easyInputVID, easyInputNormalPID)
	if len(infos) == 0 {
		return HIDDiagnostics{}, errors.New("未发现正常模式 EasyInput HID")
	}
	dev, err := infos[0].Open()
	if err != nil {
		return HIDDiagnostics{}, fmt.Errorf("打开 EasyInput HID 失败: %w", err)
	}
	defer dev.Close()
	requestID := atomic.AddUint32(&statusRequestSequence, 1)
	request := make([]byte, statusRequestPayloadLen+1)
	request[0] = statusRequestReportID
	copy(request[1:4], []byte{'S', '3', 'R'})
	request[4] = 1
	binary.LittleEndian.PutUint32(request[5:9], requestID)
	request[9] = 0x03 // fresh + diagnostics
	if _, err := dev.Write(request); err != nil {
		return HIDDiagnostics{}, fmt.Errorf("请求 EasyInput 状态失败: %w", err)
	}
	result := make(chan struct {
		value HIDDiagnostics
		err   error
	}, 1)
	go func() {
		value, readErr := readStatusResponse(dev, requestID)
		result <- struct {
			value HIDDiagnostics
			err   error
		}{value, readErr}
	}()
	select {
	case <-ctx.Done():
		_ = dev.Close()
		return HIDDiagnostics{}, ctx.Err()
	case result := <-result:
		return result.value, result.err
	}
}

var statusRequestSequence uint32

func readStatusResponse(dev *hid.Device, requestID uint32) (HIDDiagnostics, error) {
	var jsonBytes []byte
	var expectedLen int
	var expectedCRC uint16
	var next byte
	buffer := make([]byte, 64)
	for chunks := 0; chunks < 32; chunks++ {
		count, err := dev.Read(buffer)
		if err != nil {
			return HIDDiagnostics{}, fmt.Errorf("读取 EasyInput 状态失败: %w", err)
		}
		if count < statusResponsePayloadLen+1 || buffer[0] != statusResponseReportID || buffer[1] != statusResponseCommandKind || buffer[2] != next {
			continue
		}
		total := buffer[3]
		if total == 0 || next >= total || buffer[4] < statusResponseHeaderLen || binary.LittleEndian.Uint32(buffer[5:9]) != requestID {
			continue
		}
		if next == 0 {
			expectedLen = int(binary.LittleEndian.Uint16(buffer[9:11]))
			expectedCRC = binary.LittleEndian.Uint16(buffer[11:13])
			jsonBytes = make([]byte, 0, expectedLen)
		}
		chunkLen := int(buffer[4]) - statusResponseHeaderLen
		if chunkLen < 0 || 13+chunkLen > count {
			return HIDDiagnostics{}, errors.New("EasyInput 状态分片长度无效")
		}
		jsonBytes = append(jsonBytes, buffer[13:13+chunkLen]...)
		next++
		if next == total {
			if len(jsonBytes) != expectedLen || crc16CCITT(jsonBytes) != expectedCRC {
				return HIDDiagnostics{}, errors.New("EasyInput 状态校验失败")
			}
			var payload struct {
				Firmware     string `json:"firmware"`
				BatteryMV    uint16 `json:"battery_mv"`
				BatteryPct   uint8  `json:"battery_percent"`
				BatteryState string `json:"battery_state"`
				Diag         struct {
					Board        string `json:"board"`
					Last         string `json:"last"`
					InputEvents  uint32 `json:"in_evt"`
					EncoderSteps uint32 `json:"enc_step"`
					VIN          int    `json:"vin"`
					Charge       int    `json:"chrg"`
					LEDGPIO      int    `json:"led_gpio"`
				} `json:"diag"`
			}
			if err := json.Unmarshal(jsonBytes, &payload); err != nil {
				return HIDDiagnostics{}, err
			}
			return HIDDiagnostics{Supported: payload.Diag.Board != "", Firmware: payload.Firmware, LastInput: payload.Diag.Last, InputEvents: payload.Diag.InputEvents, EncoderSteps: payload.Diag.EncoderSteps, BatteryMV: payload.BatteryMV, BatteryPct: payload.BatteryPct, BatteryState: payload.BatteryState, VIN: payload.Diag.VIN, Charge: payload.Diag.Charge, LEDGPIO: payload.Diag.LEDGPIO}, nil
		}
	}
	return HIDDiagnostics{}, errors.New("等待 EasyInput 状态响应超时")
}

func crc16CCITT(data []byte) uint16 {
	crc := uint16(0xffff)
	for _, value := range data {
		crc ^= uint16(value) << 8
		for bit := 0; bit < 8; bit++ {
			if crc&0x8000 != 0 {
				crc = (crc << 1) ^ 0x1021
			} else {
				crc <<= 1
			}
		}
	}
	return crc
}
