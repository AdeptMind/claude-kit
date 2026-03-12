package packaging

import (
	"archive/zip"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// createTestArchive builds a zip archive in outDir with the given manifest and files.
// Returns the path to the created archive.
func createTestArchive(t *testing.T, outDir string, manifest Manifest, files map[string]string) string {
	t.Helper()

	ext := manifest.Type
	archivePath := filepath.Join(outDir, manifest.Name+"."+ext)

	f, err := os.Create(archivePath)
	if err != nil {
		t.Fatalf("creating archive file: %v", err)
	}
	defer f.Close()

	w := zip.NewWriter(f)
	defer w.Close()

	manifestData, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		t.Fatalf("marshaling manifest: %v", err)
	}
	mw, err := w.Create("manifest.json")
	if err != nil {
		t.Fatalf("creating manifest entry: %v", err)
	}
	if _, err := mw.Write(manifestData); err != nil {
		t.Fatalf("writing manifest: %v", err)
	}

	for name, content := range files {
		fw, err := w.Create(name)
		if err != nil {
			t.Fatalf("creating zip entry %s: %v", name, err)
		}
		if _, err := fw.Write([]byte(content)); err != nil {
			t.Fatalf("writing zip entry %s: %v", name, err)
		}
	}

	return archivePath
}

func TestInstall_Skill(t *testing.T) {
	archiveDir := t.TempDir()
	targetDir := t.TempDir()

	manifest := Manifest{Name: "my-skill", Version: "1.0.0", Type: "skill", CreatedAt: "2025-01-01T00:00:00Z"}
	files := map[string]string{
		"SKILL.md":     "# My Skill\nContent.",
		"sub/extra.md": "extra content",
	}
	archivePath := createTestArchive(t, archiveDir, manifest, files)

	result, err := Install(archivePath, targetDir, false)
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	if result.Type != "skill" {
		t.Errorf("type = %q, want %q", result.Type, "skill")
	}
	if result.Name != "my-skill" {
		t.Errorf("name = %q, want %q", result.Name, "my-skill")
	}
	if result.Version != "1.0.0" {
		t.Errorf("version = %q, want %q", result.Version, "1.0.0")
	}
	if len(result.FilesCreated) != 2 {
		t.Errorf("files created = %d, want 2", len(result.FilesCreated))
	}

	// Verify files landed in skills/{name}/
	skillMD := filepath.Join(targetDir, "skills", "my-skill", "SKILL.md")
	if _, err := os.Stat(skillMD); err != nil {
		t.Errorf("expected %s to exist: %v", skillMD, err)
	}

	extraMD := filepath.Join(targetDir, "skills", "my-skill", "sub", "extra.md")
	if _, err := os.Stat(extraMD); err != nil {
		t.Errorf("expected %s to exist: %v", extraMD, err)
	}
}

func TestInstall_Agent(t *testing.T) {
	archiveDir := t.TempDir()
	targetDir := t.TempDir()

	manifest := Manifest{Name: "backend", Version: "2.0.0", Type: "agent", CreatedAt: "2025-01-01T00:00:00Z"}
	files := map[string]string{
		"backend.md": "# Backend Agent\nRules here.",
	}
	archivePath := createTestArchive(t, archiveDir, manifest, files)

	result, err := Install(archivePath, targetDir, false)
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	if result.Type != "agent" {
		t.Errorf("type = %q, want %q", result.Type, "agent")
	}
	if result.Name != "backend" {
		t.Errorf("name = %q, want %q", result.Name, "backend")
	}

	// Agent files go to agents/
	agentMD := filepath.Join(targetDir, "agents", "backend.md")
	if _, err := os.Stat(agentMD); err != nil {
		t.Errorf("expected %s to exist: %v", agentMD, err)
	}
}

func TestInstall_ClaudeKit(t *testing.T) {
	archiveDir := t.TempDir()
	targetDir := t.TempDir()

	manifest := Manifest{Name: "my-template", Version: "1.0.0", Type: "claude-kit", CreatedAt: "2025-01-01T00:00:00Z"}
	files := map[string]string{
		"agents/dev.md":             "# Dev",
		"skills/example/SKILL.md":   "# Example",
		"rules/code-style.md":       "# Style",
	}
	archivePath := createTestArchive(t, archiveDir, manifest, files)

	result, err := Install(archivePath, targetDir, false)
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	if result.Type != "claude-kit" {
		t.Errorf("type = %q, want %q", result.Type, "claude-kit")
	}

	// claude-kit extracts directly into targetDir
	for _, rel := range []string{"agents/dev.md", "skills/example/SKILL.md", "rules/code-style.md"} {
		p := filepath.Join(targetDir, filepath.FromSlash(rel))
		if _, err := os.Stat(p); err != nil {
			t.Errorf("expected %s to exist: %v", p, err)
		}
	}
}

func TestInstall_DetectsType(t *testing.T) {
	archiveDir := t.TempDir()
	targetDir := t.TempDir()

	for _, tt := range []struct {
		archiveType string
	}{
		{"skill"},
		{"agent"},
		{"claude-kit"},
	} {
		t.Run(tt.archiveType, func(t *testing.T) {
			manifest := Manifest{Name: "test-" + tt.archiveType, Version: "1.0.0", Type: tt.archiveType, CreatedAt: "2025-01-01T00:00:00Z"}
			files := map[string]string{"readme.md": "content"}
			archivePath := createTestArchive(t, archiveDir, manifest, files)

			result, err := Install(archivePath, targetDir, true)
			if err != nil {
				t.Fatalf("Install failed: %v", err)
			}
			if result.Type != tt.archiveType {
				t.Errorf("type = %q, want %q", result.Type, tt.archiveType)
			}
		})
	}
}

func TestInstall_ConflictWarning(t *testing.T) {
	archiveDir := t.TempDir()
	targetDir := t.TempDir()

	manifest := Manifest{Name: "conflict-skill", Version: "1.0.0", Type: "skill", CreatedAt: "2025-01-01T00:00:00Z"}
	files := map[string]string{"SKILL.md": "# Conflict Skill"}
	archivePath := createTestArchive(t, archiveDir, manifest, files)

	// Pre-create the file that would conflict
	conflictPath := filepath.Join(targetDir, "skills", "conflict-skill", "SKILL.md")
	writeTestFile(t, conflictPath, "existing content")

	_, err := Install(archivePath, targetDir, false)
	if err == nil {
		t.Fatal("expected conflict error, got nil")
	}

	conflictErr, ok := err.(*ConflictError)
	if !ok {
		t.Fatalf("expected *ConflictError, got %T: %v", err, err)
	}
	if len(conflictErr.Conflicts) != 1 {
		t.Errorf("conflicts = %d, want 1", len(conflictErr.Conflicts))
	}
	if !contains(conflictErr.Conflicts, "SKILL.md") {
		t.Errorf("expected SKILL.md in conflicts, got %v", conflictErr.Conflicts)
	}
}

func TestInstall_ForceOverwrite(t *testing.T) {
	archiveDir := t.TempDir()
	targetDir := t.TempDir()

	manifest := Manifest{Name: "force-skill", Version: "2.0.0", Type: "skill", CreatedAt: "2025-01-01T00:00:00Z"}
	files := map[string]string{"SKILL.md": "# New Content"}
	archivePath := createTestArchive(t, archiveDir, manifest, files)

	// Pre-create conflicting file
	conflictPath := filepath.Join(targetDir, "skills", "force-skill", "SKILL.md")
	writeTestFile(t, conflictPath, "old content")

	result, err := Install(archivePath, targetDir, true)
	if err != nil {
		t.Fatalf("Install with force failed: %v", err)
	}

	if len(result.Conflicts) != 1 {
		t.Errorf("conflicts = %d, want 1", len(result.Conflicts))
	}

	// Verify file was overwritten
	data, err := os.ReadFile(conflictPath)
	if err != nil {
		t.Fatalf("reading overwritten file: %v", err)
	}
	if string(data) != "# New Content" {
		t.Errorf("file content = %q, want %q", string(data), "# New Content")
	}
}

func TestDetectArchiveType(t *testing.T) {
	archiveDir := t.TempDir()

	tests := []struct {
		name     string
		manifest Manifest
		want     string
	}{
		{"skill from manifest", Manifest{Name: "s", Version: "1.0.0", Type: "skill"}, "skill"},
		{"agent from manifest", Manifest{Name: "a", Version: "1.0.0", Type: "agent"}, "agent"},
		{"claude-kit from manifest", Manifest{Name: "t", Version: "1.0.0", Type: "claude-kit"}, "claude-kit"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			archivePath := createTestArchive(t, archiveDir, tt.manifest, map[string]string{"f.md": "x"})
			got, err := DetectArchiveType(archivePath)
			if err != nil {
				t.Fatalf("DetectArchiveType failed: %v", err)
			}
			if got != tt.want {
				t.Errorf("type = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDetectArchiveType_FallbackToExtension(t *testing.T) {
	archiveDir := t.TempDir()

	// Create a zip without manifest.json
	archivePath := filepath.Join(archiveDir, "nomanifest.skill")
	f, err := os.Create(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	w := zip.NewWriter(f)
	fw, _ := w.Create("readme.md")
	fw.Write([]byte("content"))
	w.Close()
	f.Close()

	got, err := DetectArchiveType(archivePath)
	if err != nil {
		t.Fatalf("DetectArchiveType failed: %v", err)
	}
	if got != "skill" {
		t.Errorf("type = %q, want %q", got, "skill")
	}
}

func TestInstall_InvalidManifest_MissingName(t *testing.T) {
	archiveDir := t.TempDir()
	targetDir := t.TempDir()

	manifest := Manifest{Name: "", Version: "1.0.0", Type: "skill"}
	archivePath := createTestArchive(t, archiveDir, manifest, map[string]string{"f.md": "x"})

	_, err := Install(archivePath, targetDir, false)
	if err == nil {
		t.Fatal("expected error for missing name")
	}
	if !strings.Contains(err.Error(), "name") {
		t.Errorf("error should mention name: %v", err)
	}
}

func TestInstall_InvalidManifest_UnknownType(t *testing.T) {
	archiveDir := t.TempDir()
	targetDir := t.TempDir()

	manifest := Manifest{Name: "x", Version: "1.0.0", Type: "unknown"}
	archivePath := createTestArchive(t, archiveDir, manifest, map[string]string{"f.md": "x"})

	_, err := Install(archivePath, targetDir, false)
	if err == nil {
		t.Fatal("expected error for unknown type")
	}
	if !strings.Contains(err.Error(), "unknown") {
		t.Errorf("error should mention unknown type: %v", err)
	}
}
