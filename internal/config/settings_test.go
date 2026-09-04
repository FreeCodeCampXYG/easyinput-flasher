package config

import "testing"

func TestDefaultSettingsUseAutomaticProxyProbe(t *testing.T) {
	settings := DefaultSettings()
	if settings.ProxyMode != "auto" || settings.ProxyURL != DefaultProxyURL {
		t.Fatalf("unexpected default proxy: %#v", settings)
	}
}

func TestNormalizeProxyURLAcceptsPortOnly(t *testing.T) {
	value, err := NormalizeProxyURL("127.0.0.1:10808")
	if err != nil || value != "http://127.0.0.1:10808" {
		t.Fatalf("NormalizeProxyURL() = %q, %v", value, err)
	}
}

func TestNormalizeRepositoryAcceptsGitHubRepositoryOnly(t *testing.T) {
	repository, err := NormalizeRepository("https://github.com/example/easyinput-fork.git")
	if err != nil || repository != "example/easyinput-fork" {
		t.Fatalf("NormalizeRepository() = %q, %v", repository, err)
	}
	if _, err := NormalizeRepository("https://github.com/example/easyinput/releases/tag/v1"); err == nil {
		t.Fatal("expected release URL to be rejected as a repository")
	}
}
