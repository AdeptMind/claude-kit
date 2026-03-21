package policy

import (
	"encoding/json"
	"os"
	"path/filepath"
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

	// MergeSettings returns diff only, doesn't write
	diff, err := MergeSettings(path, spec)
	if err != nil {
		t.Fatal(err)
	}
	if diff == "" {
		t.Error("expected non-empty diff")
	}

	// ApplySettings writes the file — check preservation
	if err := ApplySettings(path, spec); err != nil {
		t.Fatal(err)
	}

	result := readSettingsMap(t, path)
	if result["teammateMode"] != "auto" {
		t.Errorf("teammateMode not preserved: got %v", result["teammateMode"])
	}
	env, ok := result["env"].(map[string]interface{})
	if !ok || env["FOO"] != "bar" {
		t.Errorf("env not preserved: got %v", result["env"])
	}
}

func TestMergeSettings_EmptySettings(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	// Write empty JSON so readSettingsRaw can read it
	os.WriteFile(path, []byte("{}"), 0o644)

	spec := &PolicySpec{
		Filesystem: &FilesystemPolicy{
			Deny: []string{".env", "secrets/**"},
		},
	}

	if err := ApplySettings(path, spec); err != nil {
		t.Fatal(err)
	}

	result := readSettingsMap(t, path)
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

	if err := ApplySettings(path, spec); err != nil {
		t.Fatal(err)
	}

	result := readSettingsMap(t, path)
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
		t.Errorf("expected Read(.env) in deny, got %v", deny)
	}

	allow := perms["allow"].([]interface{})
	foundAllow := false
	for _, a := range allow {
		if a.(string) == "Read(src/**)" {
			foundAllow = true
		}
	}
	if !foundAllow {
		t.Errorf("expected Read(src/**) in allow, got %v", allow)
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

	if err := ApplySettings(path, spec); err != nil {
		t.Fatal(err)
	}

	result := readSettingsMap(t, path)
	hooks := result["hooks"].(map[string]interface{})
	pre, ok := hooks["PreToolUse"].([]interface{})
	if !ok || len(pre) == 0 {
		t.Error("PreToolUse hooks should be preserved")
	}
}

func TestMergeSettings_AuditHookAdded(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	os.WriteFile(path, []byte("{}"), 0o644)

	// Audit hooks are added via the Hooks field in the resolved policy
	// (the resolver converts Audit.Enabled into a PostToolUse hook)
	spec := &PolicySpec{
		Hooks: &HooksPolicy{
			PostToolUse: []HookDef{
				{Matcher: "*", Command: "echo audit"},
			},
		},
	}

	diff, err := MergeSettings(path, spec)
	if err != nil {
		t.Fatal(err)
	}

	if diff == "" {
		t.Error("diff should be non-empty when adding hooks")
	}
}

func TestApplySettings_WritesFile(t *testing.T) {
	dir := t.TempDir()
	subDir := filepath.Join(dir, "sub")
	os.MkdirAll(subDir, 0o755)
	path := filepath.Join(subDir, "settings.json")
	os.WriteFile(path, []byte("{}"), 0o644)

	spec := &PolicySpec{
		Filesystem: &FilesystemPolicy{
			Deny: []string{".env"},
		},
	}

	err := ApplySettings(path, spec)
	if err != nil {
		t.Fatal(err)
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
