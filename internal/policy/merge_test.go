package policy

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeSettings(t *testing.T, dir string, content map[string]interface{}) string {
	t.Helper()
	path := filepath.Join(dir, "settings.json")
	data, err := json.MarshalIndent(content, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func readSettingsMap(t *testing.T, path string) map[string]interface{} {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	return m
}

func TestMergeSettings_WithExisting(t *testing.T) {
	dir := t.TempDir()
	existing := map[string]interface{}{
		"teammateMode": "auto",
		"env": map[string]interface{}{
			"FOO": "bar",
		},
	}
	path := writeSettings(t, dir, existing)

	spec := &PolicySpec{
		Filesystem: &FilesystemPolicy{
			Deny:  []string{".env"},
			Allow: []string{"src/**"},
		},
	}

	merged, diff, err := MergeSettings(path, spec)
	if err != nil {
		t.Fatal(err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(merged, &result); err != nil {
		t.Fatal(err)
	}

	if result["teammateMode"] != "auto" {
		t.Errorf("teammateMode not preserved: got %v", result["teammateMode"])
	}
	env, ok := result["env"].(map[string]interface{})
	if !ok || env["FOO"] != "bar" {
		t.Errorf("env not preserved: got %v", result["env"])
	}

	if diff == "" {
		t.Error("expected non-empty diff")
	}
}

func TestMergeSettings_EmptySettings(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")

	spec := &PolicySpec{
		Filesystem: &FilesystemPolicy{
			Deny: []string{".env", "secrets/**"},
		},
	}

	merged, _, err := MergeSettings(path, spec)
	if err != nil {
		t.Fatal(err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(merged, &result); err != nil {
		t.Fatal(err)
	}

	perms, ok := result["permissions"].(map[string]interface{})
	if !ok {
		t.Fatal("permissions not created")
	}
	deny, ok := perms["deny"].([]interface{})
	if !ok || len(deny) == 0 {
		t.Error("deny patterns not set")
	}
}

func TestMergeSettings_ReplacesPermissions(t *testing.T) {
	dir := t.TempDir()
	existing := map[string]interface{}{
		"permissions": map[string]interface{}{
			"deny":  []interface{}{"Read(old_pattern)"},
			"allow": []interface{}{"Read(old_allow)"},
		},
	}
	path := writeSettings(t, dir, existing)

	spec := &PolicySpec{
		Filesystem: &FilesystemPolicy{
			Deny:  []string{".env"},
			Allow: []string{"src/**"},
		},
	}

	merged, _, err := MergeSettings(path, spec)
	if err != nil {
		t.Fatal(err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(merged, &result); err != nil {
		t.Fatal(err)
	}

	perms := result["permissions"].(map[string]interface{})
	deny := perms["deny"].([]interface{})

	for _, d := range deny {
		if d.(string) == "Read(old_pattern)" {
			t.Error("old deny pattern should have been replaced")
		}
	}

	found := false
	for _, d := range deny {
		if d.(string) == "Read(.env)" {
			found = true
		}
	}
	if !found {
		t.Error("new deny pattern not found")
	}
}

func TestMergeSettings_PreservesPreToolUse(t *testing.T) {
	dir := t.TempDir()
	preToolUse := []interface{}{
		map[string]interface{}{
			"matcher": "Bash",
			"hooks": []interface{}{
				map[string]interface{}{
					"type":    "command",
					"command": "echo staged",
				},
			},
		},
	}
	existing := map[string]interface{}{
		"hooks": map[string]interface{}{
			"PreToolUse": preToolUse,
		},
	}
	path := writeSettings(t, dir, existing)

	spec := &PolicySpec{
		Audit: &AuditPolicy{
			Enabled: true,
			Path:    "/tmp/audit.jsonl",
		},
	}

	merged, _, err := MergeSettings(path, spec)
	if err != nil {
		t.Fatal(err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(merged, &result); err != nil {
		t.Fatal(err)
	}

	hooks := result["hooks"].(map[string]interface{})
	pre, ok := hooks["PreToolUse"].([]interface{})
	if !ok || len(pre) == 0 {
		t.Error("PreToolUse hooks should be preserved")
	}

	post, ok := hooks["PostToolUse"].([]interface{})
	if !ok || len(post) == 0 {
		t.Error("PostToolUse hooks should be added")
	}
}

func TestMergeSettings_AuditHookAdded(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")

	spec := &PolicySpec{
		Audit: &AuditPolicy{
			Enabled: true,
			Path:    "/tmp/audit.jsonl",
		},
	}

	merged, diff, err := MergeSettings(path, spec)
	if err != nil {
		t.Fatal(err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(merged, &result); err != nil {
		t.Fatal(err)
	}

	hooks, ok := result["hooks"].(map[string]interface{})
	if !ok {
		t.Fatal("hooks not created")
	}

	post, ok := hooks["PostToolUse"].([]interface{})
	if !ok || len(post) == 0 {
		t.Fatal("PostToolUse not added")
	}

	if !strings.Contains(diff, "PostToolUse") {
		t.Error("diff should mention PostToolUse")
	}
}

func TestApplySettings_WritesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "settings.json")

	spec := &PolicySpec{
		Filesystem: &FilesystemPolicy{
			Deny: []string{".env"},
		},
	}

	diff, err := ApplySettings(path, spec)
	if err != nil {
		t.Fatal(err)
	}

	if diff == "" {
		t.Error("expected non-empty diff")
	}

	result := readSettingsMap(t, path)
	perms, ok := result["permissions"].(map[string]interface{})
	if !ok {
		t.Fatal("permissions not written")
	}
	deny, ok := perms["deny"].([]interface{})
	if !ok || len(deny) == 0 {
		t.Error("deny patterns not written to file")
	}
}
