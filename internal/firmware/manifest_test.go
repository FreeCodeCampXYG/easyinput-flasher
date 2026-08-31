package firmware

import "testing"

func TestParseManifestRejectsWrongTarget(t *testing.T) {
	_, err := ParseManifest([]byte(`{"schemaVersion":1,"product":"easyinput-firmware","board":"other","chip":"esp32s3","tag":"v1","commit":"abc","files":[]}`))
	if err == nil {
		t.Fatal("expected wrong board to be rejected")
	}
}

func TestParseManifestRejectsUnexpectedFlashRange(t *testing.T) {
	manifest := []byte(`{
  "schemaVersion": 1,
  "product": "easyinput-firmware",
  "board": "easyinput-v2",
  "chip": "esp32s3",
  "tag": "firmware-main-v1",
  "commit": "0123456789abcdef",
  "files": [
    {"name": "bootloader.bin", "offset": "0x0", "sha256": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "size": 1},
    {"name": "partition-table.bin", "offset": "0x8000", "sha256": "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", "size": 1},
    {"name": "easy_input_keyboard.bin", "offset": "0x11000", "sha256": "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc", "size": 1}
  ]
}`)
	if _, err := ParseManifest(manifest); err == nil {
		t.Fatal("expected unexpected application offset to be rejected")
	}
}
