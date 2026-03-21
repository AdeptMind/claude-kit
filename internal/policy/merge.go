package policy

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AdeptMind/infra-tool/claude-cli/internal/audit"
)

// MergeSettings reads settings.json, generates policy permissions/hooks,
// returns the merged JSON and a human-readable diff string.
func MergeSettings(settingsPath string, resolved *PolicySpec) ([]byte, string, error) {
	existing := make(map[string]interface{})

	data, err := os.ReadFile(settingsPath)
	if err != nil && !os.IsNotExist(err) {
		return nil, "", fmt.Errorf("policy: reading settings: %w", err)
	}
	if len(data) > 0 {
		if err := json.Unmarshal(data, &existing); err != nil {
			return nil, "", fmt.Errorf("policy: parsing settings: %w", err)
		}
	}

	beforeJSON, _ := json.MarshalIndent(existing, "", "  ")

	applyPermissions(existing, resolved)
	applyHooks(existing, resolved)

	afterJSON, err := json.MarshalIndent(existing, "", "  ")
	if err != nil {
		return nil, "", fmt.Errorf("policy: marshaling merged settings: %w", err)
	}
	afterJSON = append(afterJSON, '\n')

	diff := computeDiff(string(beforeJSON), string(afterJSON))

	return afterJSON, diff, nil
}

// ApplySettings writes the merged settings to disk.
func ApplySettings(settingsPath string, resolved *PolicySpec) (string, error) {
	merged, diff, err := MergeSettings(settingsPath, resolved)
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(filepath.Dir(settingsPath), 0o755); err != nil {
		return "", fmt.Errorf("policy: creating settings dir: %w", err)
	}

	if err := os.WriteFile(settingsPath, merged, 0o644); err != nil {
		return "", fmt.Errorf("policy: writing settings: %w", err)
	}

	return diff, nil
}

func applyPermissions(settings map[string]interface{}, resolved *PolicySpec) {
	if resolved.Filesystem == nil {
		return
	}

	perms, ok := settings["permissions"].(map[string]interface{})
	if !ok {
		perms = make(map[string]interface{})
		settings["permissions"] = perms
	}

	if len(resolved.Filesystem.Deny) > 0 {
		perms["deny"] = toClaudePermPatterns(resolved.Filesystem.Deny)
	}
	if len(resolved.Filesystem.Allow) > 0 {
		perms["allow"] = toClaudePermPatterns(resolved.Filesystem.Allow)
	}
}

func applyHooks(settings map[string]interface{}, resolved *PolicySpec) {
	if resolved.Audit == nil || !resolved.Audit.Enabled {
		return
	}

	hooks, ok := settings["hooks"].(map[string]interface{})
	if !ok {
		hooks = make(map[string]interface{})
		settings["hooks"] = hooks
	}

	auditPath := resolved.Audit.Path
	if auditPath == "" {
		auditPath = ".claude/audit.jsonl"
	}

	hookEntry := audit.GenerateHookEntry(auditPath)
	hooks["PostToolUse"] = []interface{}{
		map[string]interface{}{
			"matcher": ".*",
			"hooks":   []interface{}{hookEntry},
		},
	}
}

// toClaudePermPatterns converts policy glob patterns into Claude settings
// permission patterns (Read/Bash prefixed).
func toClaudePermPatterns(patterns []string) []interface{} {
	result := make([]interface{}, 0, len(patterns))
	for _, p := range patterns {
		result = append(result, "Read("+p+")")
		result = append(result, "Bash(cat "+p+")")
	}
	return result
}

func computeDiff(before, after string) string {
	beforeLines := strings.Split(before, "\n")
	afterLines := strings.Split(after, "\n")

	var sb strings.Builder

	// Removals: lines in before but not in after
	afterCounts := make(map[string]int, len(afterLines))
	for _, l := range afterLines {
		afterCounts[l]++
	}
	for _, l := range beforeLines {
		if afterCounts[l] > 0 {
			afterCounts[l]--
			continue
		}
		sb.WriteString("- " + l + "\n")
	}

	// Additions: lines in after but not in before
	beforeCounts := make(map[string]int, len(beforeLines))
	for _, l := range beforeLines {
		beforeCounts[l]++
	}
	for _, l := range afterLines {
		if beforeCounts[l] > 0 {
			beforeCounts[l]--
			continue
		}
		sb.WriteString("+ " + l + "\n")
	}

	return sb.String()
}
