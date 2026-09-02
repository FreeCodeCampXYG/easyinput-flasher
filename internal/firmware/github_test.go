package firmware

import (
	"reflect"
	"testing"

	"github.com/FreeCodeCampXYG/easyinput-flasher/internal/domain"
)

func TestSortFirmwareReleasesUsesSemanticVersionOrder(t *testing.T) {
	releases := []domain.FirmwareRelease{
		{Tag: "firmware-v0.2.9", PublishedAt: "2026-09-01T00:00:00Z"},
		{Tag: "firmware-v0.2.8", PublishedAt: "2026-09-01T00:00:00Z"},
		{Tag: "firmware-v0.2.10", PublishedAt: "2026-08-31T00:00:00Z"},
		{Tag: "firmware-v0.2.3-main", PublishedAt: "2026-09-02T00:00:00Z"},
		{Tag: "firmware-v0.2.3", PublishedAt: "2026-09-01T00:00:00Z"},
	}

	sortFirmwareReleases(releases)

	got := make([]string, len(releases))
	for index, release := range releases {
		got[index] = release.Tag
	}
	want := []string{"firmware-v0.2.10", "firmware-v0.2.9", "firmware-v0.2.8", "firmware-v0.2.3", "firmware-v0.2.3-main"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("sorted tags = %v, want %v", got, want)
	}
}

func TestSortFirmwareReleasesFallsBackToPublishedAt(t *testing.T) {
	releases := []domain.FirmwareRelease{
		{Tag: "factory-recovery", PublishedAt: "2026-09-01T00:00:00Z"},
		{Tag: "community-build", PublishedAt: "2026-09-02T00:00:00Z"},
	}

	sortFirmwareReleases(releases)

	if releases[0].Tag != "community-build" {
		t.Fatalf("first fallback tag = %q, want community-build", releases[0].Tag)
	}
}
