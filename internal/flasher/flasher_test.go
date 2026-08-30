package flasher

import (
	"strings"
	"testing"

	"github.com/FreeCodeCampXYG/easyinput-flasher/internal/domain"
)

func TestFlashArgsRejectsWrongBoard(t *testing.T) {
	_, err := FlashArgs("COM3", ".", domain.FirmwareManifest{Board: "other", Chip: domain.ChipType})
	if err == nil {
		t.Fatal("expected wrong board to fail")
	}
}

func TestRedactHidesMAC(t *testing.T) {
	actual := redact("MAC: aa:bb:cc:dd:ee:ff")
	if strings.Contains(actual, "aa:bb:cc") || !strings.Contains(actual, "ee:ff") {
		t.Fatalf("unexpected redaction: %q", actual)
	}
}
