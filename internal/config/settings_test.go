package config

import "testing"

func TestDefaultSettingsUseConfiguredLocalProxy(t *testing.T) {
	settings := DefaultSettings()
	if settings.ProxyMode != "custom" || settings.ProxyURL != DefaultProxyURL {
		t.Fatalf("unexpected default proxy: %#v", settings)
	}
}
