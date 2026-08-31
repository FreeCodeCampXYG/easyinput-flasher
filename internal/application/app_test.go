package application

import (
	"strings"
	"testing"

	"github.com/FreeCodeCampXYG/easyinput-flasher/internal/domain"
)

func TestStartFlashRejectsUnknownOrUntrustedFirmware(t *testing.T) {
	app := New("test")
	app.devices["COM3"] = domain.DeviceInfo{
		ID: "COM3", Chip: domain.ChipType, MACSuffix: "A1B2C3", Verified: true,
	}
	app.firmware = []domain.FirmwareRelease{{
		Repository: "community/easyinput", Tag: "firmware-v1", Trusted: false,
	}}

	request := FlashRequest{DeviceID: "COM3", Confirmation: "确认烧录 A1B2C3"}
	request.FirmwareID = "unknown/repository@firmware-v1"
	if err := app.StartFlash(request); err == nil || !strings.Contains(err.Error(), "选择已过期") {
		t.Fatalf("unknown release error = %v, want expired selection rejection", err)
	}

	request.FirmwareID = "community/easyinput@firmware-v1"
	if err := app.StartFlash(request); err == nil || !strings.Contains(err.Error(), "尚未审核") {
		t.Fatalf("untrusted release error = %v, want trust rejection", err)
	}
}
