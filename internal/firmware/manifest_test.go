package firmware

import "testing"

func TestParseManifestRejectsWrongTarget(t *testing.T) {
	_, err := ParseManifest([]byte(`{"schemaVersion":1,"product":"easyinput-firmware","board":"other","chip":"esp32s3","tag":"v1","commit":"abc","files":[]}`))
	if err == nil {
		t.Fatal("expected wrong board to be rejected")
	}
}
