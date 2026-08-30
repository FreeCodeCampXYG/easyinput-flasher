package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Settings struct {
	ProxyMode string   `json:"proxyMode"`
	ProxyURL  string   `json:"proxyUrl"`
	Sources   []Source `json:"sources"`
}

type Source struct {
	Repository string `json:"repository"`
	Trusted    bool   `json:"trusted"`
	Enabled    bool   `json:"enabled"`
}

func DefaultSettings() Settings {
	return Settings{
		ProxyMode: "inherit",
		Sources:   []Source{{Repository: "FreeCodeCampXYG/easy-input-maker", Trusted: true, Enabled: true}},
	}
}

func Load() (Settings, error) {
	path, err := settingsPath()
	if err != nil {
		return Settings{}, err
	}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return DefaultSettings(), nil
	}
	if err != nil {
		return Settings{}, fmt.Errorf("读取应用设置失败: %w", err)
	}
	var settings Settings
	if err := json.Unmarshal(data, &settings); err != nil {
		return Settings{}, fmt.Errorf("设置文件格式无效，未覆盖原文件: %w", err)
	}
	return settings, nil
}

// Save 先写入完整临时文件再替换，避免进程中断留下半截信任来源配置。
func Save(settings Settings) error {
	path, err := settingsPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	temporary := path + ".tmp"
	if err := os.WriteFile(temporary, data, 0o600); err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return os.Rename(temporary, path)
}

func settingsPath() (string, error) {
	root, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "easyinput-flasher", "settings.json"), nil
}
