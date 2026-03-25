package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// AppSettings holds the persisted UI settings.
type AppSettings struct {
	CoworkFolder  string `json:"coworkFolder"`
	WorkspaceRoot string `json:"workspaceRoot"`
}

// SettingsService reads/writes UI settings to ~/.claude-kit/settings.json.
type SettingsService struct{}

func (s *SettingsService) configPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".claude-kit", "settings.json")
}

// Load returns saved settings, falling back to sensible defaults.
func (s *SettingsService) Load() AppSettings {
	defaults := AppSettings{
		CoworkFolder:  filepath.Join(homeDir(), ".claude", "cowork-profiles"),
		WorkspaceRoot: filepath.Join(homeDir(), "workspace"),
	}

	data, err := os.ReadFile(s.configPath())
	if err != nil {
		return defaults
	}

	var settings AppSettings
	if err := json.Unmarshal(data, &settings); err != nil {
		return defaults
	}

	// Fill empty fields with defaults
	if settings.CoworkFolder == "" {
		settings.CoworkFolder = defaults.CoworkFolder
	}
	if settings.WorkspaceRoot == "" {
		settings.WorkspaceRoot = defaults.WorkspaceRoot
	}
	return settings
}

// Save persists the settings to disk.
func (s *SettingsService) Save(settings AppSettings) error {
	dir := filepath.Dir(s.configPath())
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.configPath(), append(data, '\n'), 0o644)
}

func homeDir() string {
	home, _ := os.UserHomeDir()
	return home
}
