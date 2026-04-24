package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// CheckDependencies verifies that all registered dependencies are installed
// and up to date. Prints warnings for missing or outdated dependencies.
func CheckDependencies() {
	fmt.Println(sectionHeader("Dependency Check"))

	allOk := true
	for _, dep := range depRegistry {
		switch {
		case dep.Type == DepTypePlugin:
			if !isPluginInstalled(dep.PluginKey) {
				allOk = false
				fmt.Println(warnStyle.Render(fmt.Sprintf("  ⚠ %s is not installed", dep.Name)))
				fmt.Println(dimStyle.Render(fmt.Sprintf("    Run in Claude Code: %s", dep.PluginMarketplaceCmd)))
				fmt.Println(dimStyle.Render(fmt.Sprintf("    Then:               %s", dep.PluginInstallCmd)))
			}
		default:
			installed, version := getInstalledVersion(dep)
			if !installed {
				allOk = false
				fmt.Println(warnStyle.Render(fmt.Sprintf("  ⚠ %s is not installed", dep.Name)))
				fmt.Println(dimStyle.Render("    Run 'ck dep install' to install it."))
			} else if dep.MinVersion != "" && version != "" && semverGT(dep.MinVersion, version) {
				allOk = false
				fmt.Println(warnStyle.Render(fmt.Sprintf("  ⚠ %s %s is outdated (minimum: %s)", dep.Name, version, dep.MinVersion)))
				if dep.Type == DepTypeBrew {
					fmt.Println(dimStyle.Render(fmt.Sprintf("    Run 'brew upgrade %s' or 'ck dep install'.", dep.Source)))
				}
			}
		}
	}

	if allOk {
		fmt.Println(fmt.Sprintf("  %s %s", checkMark, dimStyle.Render("All dependencies OK")))
	}
}

// getInstalledVersion runs the dependency's VersionCmd and extracts the semver.
// Returns (false, "") if not installed, (true, version) otherwise.
func getInstalledVersion(dep Dependency) (bool, string) {
	if dep.VersionCmd == "" {
		// No version command — just check if binary exists
		name := strings.Fields(dep.Source)[0]
		_, err := exec.LookPath(name)
		return err == nil, ""
	}

	parts := strings.Fields(dep.VersionCmd)
	out, err := exec.Command(parts[0], parts[1:]...).Output()
	if err != nil {
		return false, ""
	}

	return true, parseVersionOutput(string(out))
}

// parseVersionOutput extracts a semver (X.Y.Z) from command output.
// Handles formats like "rtk 0.37.2", "v1.2.3", "tool version 1.0.0".
func parseVersionOutput(output string) string {
	for _, field := range strings.Fields(output) {
		v := strings.TrimPrefix(field, "v")
		if parseSemver(v) != nil {
			return v
		}
	}
	return ""
}

// isPluginInstalled checks ~/.claude/plugins/installed_plugins.json for the given key.
func isPluginInstalled(pluginKey string) bool {
	if pluginKey == "" {
		return false
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	data, err := os.ReadFile(filepath.Join(home, ".claude", "plugins", "installed_plugins.json"))
	if err != nil {
		return false
	}

	var file struct {
		Plugins map[string]json.RawMessage `json:"plugins"`
	}
	if err := json.Unmarshal(data, &file); err != nil {
		return false
	}

	_, exists := file.Plugins[pluginKey]
	return exists
}

// isBrewInstalled checks if a formula is already installed via Homebrew.
func isBrewInstalled(formula string) bool {
	err := exec.Command("brew", "list", "--formula", formula).Run()
	return err == nil
}
