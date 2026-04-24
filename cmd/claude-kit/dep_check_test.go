package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseVersionOutput(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{"rtk format", "rtk 0.37.2", "0.37.2"},
		{"prefixed v", "v1.2.3", "1.2.3"},
		{"verbose output", "tool version 2.0.1\n", "2.0.1"},
		{"no version", "some random output", ""},
		{"empty", "", ""},
		{"multiline", "name 1.0.0\nother stuff", "1.0.0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseVersionOutput(tt.input)
			if got != tt.expect {
				t.Errorf("parseVersionOutput(%q) = %q, want %q", tt.input, got, tt.expect)
			}
		})
	}
}

func TestIsPluginInstalled(t *testing.T) {
	t.Run("returns false for empty key", func(t *testing.T) {
		if isPluginInstalled("") {
			t.Error("expected false for empty key")
		}
	})
}

func TestIsPluginInstalledWithFile(t *testing.T) {
	// Create a temp directory mimicking ~/.claude/plugins/
	tmpHome := t.TempDir()
	pluginsDir := filepath.Join(tmpHome, ".claude", "plugins")
	os.MkdirAll(pluginsDir, 0o755)

	content := `{"version":2,"plugins":{"claude-mem@thedotmack":[{"scope":"user"}]}}`
	os.WriteFile(filepath.Join(pluginsDir, "installed_plugins.json"), []byte(content), 0o644)

	// Override HOME for the test
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	t.Cleanup(func() { os.Setenv("HOME", origHome) })

	t.Run("finds installed plugin", func(t *testing.T) {
		if !isPluginInstalled("claude-mem@thedotmack") {
			t.Error("expected true for installed plugin")
		}
	})

	t.Run("returns false for missing plugin", func(t *testing.T) {
		if isPluginInstalled("nonexistent@marketplace") {
			t.Error("expected false for missing plugin")
		}
	})
}
