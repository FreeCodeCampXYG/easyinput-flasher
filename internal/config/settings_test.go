package config

import "testing"

func TestDefaultSettingsUseConfiguredLocalProxy(t *testing.T) {
	settings := DefaultSettings()
	if settings.ProxyMode != "custom" || settings.ProxyURL != DefaultProxyURL {
		t.Fatalf("unexpected default proxy: %#v", settings)
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
