package flasher

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/FreeCodeCampXYG/easyinput-flasher/internal/domain"
)

func TestLoadImagesRejectsWrongBoard(t *testing.T) {
	_, err := loadImages(t.TempDir(), domain.FirmwareManifest{Board: "other", Chip: domain.ChipType})
	if err == nil {
		t.Fatal("expected wrong board to fail")
	}
}

func TestLoadImagesRejectsUnsafeFileName(t *testing.T) {
	manifest := validManifest([]domain.FlashFile{
		{Name: "../outside.bin", Offset: "0x0"},
		{Name: "partition-table.bin", Offset: "0x8000"},
		{Name: "easy_input_keyboard.bin", Offset: "0x10000"},
	})
	_, err := loadImages(t.TempDir(), manifest)
	if err == nil || !strings.Contains(err.Error(), "无效文件名") {
		t.Fatalf("unexpected unsafe path result: %v", err)
	}
}

func TestLoadImagesUsesManifestOffsets(t *testing.T) {
	directory := t.TempDir()
	files := []domain.FlashFile{
		{Name: "bootloader.bin", Offset: "0x0"},
		{Name: "partition-table.bin", Offset: "0x8000"},
		{Name: "easy_input_keyboard.bin", Offset: "0x10000"},
	}
	for _, file := range files {
		if err := os.WriteFile(filepath.Join(directory, file.Name), []byte(file.Name), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	images, err := loadImages(directory, validManifest(files))
	if err != nil {
		t.Fatal(err)
	}
	if len(images) != 3 || images[1].Offset != 0x8000 || string(images[2].Data) != "easy_input_keyboard.bin" {
		t.Fatalf("unexpected image mapping: %#v", images)
	}
}

func TestRedactHidesMAC(t *testing.T) {
	actual := redact("MAC: aa:bb:cc:dd:ee:ff")
	if strings.Contains(actual, "aa:bb:cc") || !strings.Contains(actual, "ee:ff") {
		t.Fatalf("unexpected redaction: %q", actual)
	}
}

func TestRunnerDoesNotRequireExternalHelper(t *testing.T) {
	if _, err := NewRunner(); err != nil {
		t.Fatalf("pure Go runner should not require an external helper: %v", err)
	}
}

func validManifest(files []domain.FlashFile) domain.FirmwareManifest {
	return domain.FirmwareManifest{Board: domain.BoardID, Chip: domain.ChipType, Files: files}
}
