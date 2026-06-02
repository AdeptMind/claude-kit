package policy

import (
	"os"
	"path/filepath"
	"testing"
)

// writeProfile writes a YAML profile file to the given directory.
func writeProfile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name+".yaml"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestResolve_ProfileOnly(t *testing.T) {
	spec := &PolicySpec{
		Profile: "strict",
		Filesystem: &FilesystemPolicy{
			Deny: []string{".env"},
		},
	}
	got, err := Resolve(spec, "/nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != spec {
		t.Error("expected same pointer back when no extends")
	}
}

func TestResolve_ExtendsStrict(t *testing.T) {
	dir := t.TempDir()
	writeProfile(t, dir, "strict", `
profile: strict
filesystem:
  deny:
    - ".env"
    - "secrets/**"
  allow:
    - "src/**"
    - ".claude/**"
network:
  mode: block_all
`)

	spec := &PolicySpec{
		Extends: "strict",
		Filesystem: &FilesystemPolicy{
			Allow: []string{"custom/**"},
		},
	}

	got, err := Resolve(spec, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Allow should be replaced entirely by custom
	if len(got.Filesystem.Allow) != 1 || got.Filesystem.Allow[0] != "custom/**" {
		t.Errorf("allow should be replaced by custom, got %v", got.Filesystem.Allow)
	}

	// Deny should still be the base (no custom deny added)
	if len(got.Filesystem.Deny) != 2 {
		t.Errorf("deny should have base entries, got %v", got.Filesystem.Deny)
	}

	// Network should be inherited from base
	if got.Network == nil || got.Network.Mode != "block_all" {
		t.Error("network should be inherited from base")
	}
}

func TestResolve_DenyAdditive(t *testing.T) {
	dir := t.TempDir()
	writeProfile(t, dir, "strict", `
profile: strict
filesystem:
  deny:
    - ".env"
    - "secrets/**"
  allow:
    - "src/**"
network:
  mode: block_all
`)

	spec := &PolicySpec{
		Extends: "strict",
		Filesystem: &FilesystemPolicy{
			Deny: []string{"secrets/**", "vendor/**"},
		},
	}

	got, err := Resolve(spec, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Union of base + custom deny, deduplicated
	expected := map[string]bool{".env": true, "secrets/**": true, "vendor/**": true}
	if len(got.Filesystem.Deny) != len(expected) {
		t.Fatalf("expected %d deny entries, got %d: %v", len(expected), len(got.Filesystem.Deny), got.Filesystem.Deny)
	}
	for _, d := range got.Filesystem.Deny {
		if !expected[d] {
			t.Errorf("unexpected deny entry: %s", d)
		}
	}
}

func TestResolve_HooksAppended(t *testing.T) {
	dir := t.TempDir()
	writeProfile(t, dir, "moderate", `
profile: moderate
hooks:
  pre_tool_use:
    - matcher: "Bash"
      command: "base-hook"
network:
  mode: unrestricted
`)

	spec := &PolicySpec{
		Extends: "moderate",
		Hooks: &HooksPolicy{
			PreToolUse: []HookDef{
				{Matcher: "Write", Command: "custom-hook"},
			},
		},
	}

	got, err := Resolve(spec, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got.Hooks.PreToolUse) != 2 {
		t.Fatalf("expected 2 pre_tool_use hooks, got %d", len(got.Hooks.PreToolUse))
	}
	if got.Hooks.PreToolUse[0].Command != "base-hook" {
		t.Error("first hook should be base")
	}
	if got.Hooks.PreToolUse[1].Command != "custom-hook" {
		t.Error("second hook should be custom")
	}
}

func TestResolve_InvalidProfile(t *testing.T) {
	dir := t.TempDir()

	spec := &PolicySpec{
		Extends: "strict",
	}

	_, err := Resolve(spec, dir)
	if err == nil {
		t.Fatal("expected error for missing profile file")
	}
}
