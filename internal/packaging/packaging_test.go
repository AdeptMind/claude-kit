package packaging

import (
	"archive/zip"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPackageSkill_CreatesValidZip(t *testing.T) {
	skillDir := createTestSkill(t, "my-skill", "1.0.0")
	outDir := t.TempDir()

	if err := PackageSkill(skillDir, outDir, PackageOptions{}); err != nil {
		t.Fatalf("PackageSkill failed: %v", err)
	}

	archivePath := filepath.Join(outDir, "my-skill.skill")
	if _, err := os.Stat(archivePath); err != nil {
		t.Fatalf("archive not created: %v", err)
	}

	r, err := zip.OpenReader(archivePath)
	if err != nil {
		t.Fatalf("cannot open zip: %v", err)
	}
	defer r.Close()

	names := zipFileNames(r)
	if !contains(names, "manifest.json") {
		t.Error("manifest.json not found in archive")
	}
	if !contains(names, "SKILL.md") {
		t.Error("SKILL.md not found in archive")
	}
}

func TestPackageSkill_IncludesManifest(t *testing.T) {
	skillDir := createTestSkill(t, "test-skill", "2.1.0")
	outDir := t.TempDir()

	if err := PackageSkill(skillDir, outDir, PackageOptions{}); err != nil {
		t.Fatalf("PackageSkill failed: %v", err)
	}

	m := readManifestFromZip(t, filepath.Join(outDir, "test-skill.skill"))

	if m.Name != "test-skill" {
		t.Errorf("manifest name = %q, want %q", m.Name, "test-skill")
	}
	if m.Version != "2.1.0" {
		t.Errorf("manifest version = %q, want %q", m.Version, "2.1.0")
	}
	if m.Type != "skill" {
		t.Errorf("manifest type = %q, want %q", m.Type, "skill")
	}
	if m.CreatedAt == "" {
		t.Error("manifest created_at is empty")
	}
}

func TestPackageSkill_ExcludesArtifacts(t *testing.T) {
	skillDir := createTestSkill(t, "clean-skill", "1.0.0")

	// Create directories and files that should be excluded
	for _, dir := range []string{".git", "node_modules", "__pycache__"} {
		os.MkdirAll(filepath.Join(skillDir, dir), 0o755)
		writeTestFile(t, filepath.Join(skillDir, dir, "file.txt"), "junk")
	}
	writeTestFile(t, filepath.Join(skillDir, ".DS_Store"), "junk")
	writeTestFile(t, filepath.Join(skillDir, "cache.pyc"), "junk")

	outDir := t.TempDir()
	if err := PackageSkill(skillDir, outDir, PackageOptions{}); err != nil {
		t.Fatalf("PackageSkill failed: %v", err)
	}

	r, err := zip.OpenReader(filepath.Join(outDir, "clean-skill.skill"))
	if err != nil {
		t.Fatalf("cannot open zip: %v", err)
	}
	defer r.Close()

	names := zipFileNames(r)
	for _, excluded := range []string{".git/file.txt", "node_modules/file.txt", "__pycache__/file.txt", ".DS_Store", "cache.pyc"} {
		if contains(names, excluded) {
			t.Errorf("archive should not contain %q", excluded)
		}
	}
}

func TestPackageSkill_FailsOnInvalidSkill(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "invalid-skill")
	os.MkdirAll(skillDir, 0o755)
	// No SKILL.md — validation should fail

	outDir := t.TempDir()
	err := PackageSkill(skillDir, outDir, PackageOptions{})
	if err == nil {
		t.Fatal("expected error for invalid skill, got nil")
	}
	if !strings.Contains(err.Error(), "validation failed") {
		t.Errorf("error should mention validation: %v", err)
	}
}

func TestPackageSkill_SkipValidation(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "no-frontmatter")
	os.MkdirAll(skillDir, 0o755)
	writeTestFile(t, filepath.Join(skillDir, "SKILL.md"), "no frontmatter here")

	outDir := t.TempDir()
	err := PackageSkill(skillDir, outDir, PackageOptions{SkipValidation: true})
	if err != nil {
		t.Fatalf("expected no error with SkipValidation, got: %v", err)
	}

	archivePath := filepath.Join(outDir, "no-frontmatter.skill")
	if _, err := os.Stat(archivePath); err != nil {
		t.Fatalf("archive not created: %v", err)
	}
}

func TestPackageSkill_CustomExcludePatterns(t *testing.T) {
	skillDir := createTestSkill(t, "custom-exclude", "1.0.0")
	writeTestFile(t, filepath.Join(skillDir, "secret.key"), "private")
	writeTestFile(t, filepath.Join(skillDir, "notes.txt"), "keep me")

	outDir := t.TempDir()
	opts := PackageOptions{ExcludePatterns: []string{"*.key"}}
	if err := PackageSkill(skillDir, outDir, opts); err != nil {
		t.Fatalf("PackageSkill failed: %v", err)
	}

	r, err := zip.OpenReader(filepath.Join(outDir, "custom-exclude.skill"))
	if err != nil {
		t.Fatalf("cannot open zip: %v", err)
	}
	defer r.Close()

	names := zipFileNames(r)
	if contains(names, "secret.key") {
		t.Error("archive should not contain secret.key")
	}
	if !contains(names, "notes.txt") {
		t.Error("archive should contain notes.txt")
	}
}

func TestPackageAgent_CreatesValidZip(t *testing.T) {
	dir := t.TempDir()
	agentPath := filepath.Join(dir, "agents", "backend.md")
	writeTestFile(t, agentPath, `---
name: backend
description: Backend agent
version: "1.0.0"
---
# Backend
`)

	outDir := t.TempDir()
	err := PackageAgent(agentPath, outDir, PackageOptions{SkipValidation: true})
	if err != nil {
		t.Fatalf("PackageAgent failed: %v", err)
	}

	archivePath := filepath.Join(outDir, "backend.agent")
	if _, err := os.Stat(archivePath); err != nil {
		t.Fatalf("archive not created: %v", err)
	}

	m := readManifestFromZip(t, archivePath)
	if m.Type != "agent" {
		t.Errorf("manifest type = %q, want %q", m.Type, "agent")
	}
	if m.Name != "backend" {
		t.Errorf("manifest name = %q, want %q", m.Name, "backend")
	}
}

func TestPackageTemplate_CreatesValidZip(t *testing.T) {
	dir := t.TempDir()
	templateDir := filepath.Join(dir, "my-template")
	os.MkdirAll(filepath.Join(templateDir, "agents"), 0o755)
	os.MkdirAll(filepath.Join(templateDir, "skills", "example"), 0o755)
	writeTestFile(t, filepath.Join(templateDir, "agents", "dev.md"), "# Dev agent")
	writeTestFile(t, filepath.Join(templateDir, "skills", "example", "SKILL.md"), "# Example")

	outDir := t.TempDir()
	err := PackageTemplate(templateDir, outDir, PackageOptions{})
	if err != nil {
		t.Fatalf("PackageTemplate failed: %v", err)
	}

	archivePath := filepath.Join(outDir, "my-template.claude-kit")
	if _, err := os.Stat(archivePath); err != nil {
		t.Fatalf("archive not created: %v", err)
	}

	m := readManifestFromZip(t, archivePath)
	if m.Type != "claude-kit" {
		t.Errorf("manifest type = %q, want %q", m.Type, "claude-kit")
	}

	r, err := zip.OpenReader(archivePath)
	if err != nil {
		t.Fatalf("cannot open zip: %v", err)
	}
	defer r.Close()

	names := zipFileNames(r)
	if !contains(names, "agents/dev.md") {
		t.Error("archive should contain agents/dev.md")
	}
	if !contains(names, "skills/example/SKILL.md") {
		t.Error("archive should contain skills/example/SKILL.md")
	}
}

func TestShouldExclude_Patterns(t *testing.T) {
	tests := []struct {
		rel      string
		excludes []string
		want     bool
	}{
		{".git", []string{".git"}, true},
		{"src/.git/config", []string{".git"}, true},
		{"node_modules/pkg/index.js", []string{"node_modules"}, true},
		{"src/main.go", []string{".git", "node_modules"}, false},
		{"cache.pyc", []string{"*.pyc"}, true},
		{"src/deep/file.pyc", []string{"*.pyc"}, true},
		{".DS_Store", []string{".DS_Store"}, true},
		{"src/.env", []string{".env"}, true},
		{"README.md", []string{".git"}, false},
		{"dist/bundle.js", []string{"dist"}, true},
	}

	for _, tt := range tests {
		got := shouldExclude(tt.rel, tt.excludes)
		if got != tt.want {
			t.Errorf("shouldExclude(%q, %v) = %v, want %v", tt.rel, tt.excludes, got, tt.want)
		}
	}
}

func TestCreateManifest_Fields(t *testing.T) {
	skillDir := createTestSkill(t, "manifest-test", "3.0.0")
	outDir := t.TempDir()

	if err := PackageSkill(skillDir, outDir, PackageOptions{}); err != nil {
		t.Fatalf("PackageSkill failed: %v", err)
	}

	m := readManifestFromZip(t, filepath.Join(outDir, "manifest-test.skill"))

	if m.Name != "manifest-test" {
		t.Errorf("name = %q, want %q", m.Name, "manifest-test")
	}
	if m.Version != "3.0.0" {
		t.Errorf("version = %q, want %q", m.Version, "3.0.0")
	}
	if m.Type != "skill" {
		t.Errorf("type = %q, want %q", m.Type, "skill")
	}
	// Verify ISO 8601 format
	if !strings.Contains(m.CreatedAt, "T") || !strings.HasSuffix(m.CreatedAt, "Z") {
		t.Errorf("created_at should be ISO 8601 UTC, got %q", m.CreatedAt)
	}
}

func TestPackageSkill_DefaultVersion(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "no-version")
	os.MkdirAll(skillDir, 0o755)
	writeTestFile(t, filepath.Join(skillDir, "SKILL.md"), `---
name: no-version
description: Skill without version
---
Content.
`)

	outDir := t.TempDir()
	if err := PackageSkill(skillDir, outDir, PackageOptions{}); err != nil {
		t.Fatalf("PackageSkill failed: %v", err)
	}

	m := readManifestFromZip(t, filepath.Join(outDir, "no-version.skill"))
	if m.Version != "0.0.0" {
		t.Errorf("version = %q, want %q (default)", m.Version, "0.0.0")
	}
}

// --- helpers ---

func createTestSkill(t *testing.T, name, version string) string {
	t.Helper()
	dir := t.TempDir()
	skillDir := filepath.Join(dir, name)
	os.MkdirAll(skillDir, 0o755)

	content := "---\nname: " + name + "\ndescription: Test skill " + name + "\n"
	if version != "" {
		content += "version: \"" + version + "\"\n"
	}
	content += "---\n# " + name + "\nContent.\n"

	writeTestFile(t, filepath.Join(skillDir, "SKILL.md"), content)
	return skillDir
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func zipFileNames(r *zip.ReadCloser) []string {
	var names []string
	for _, f := range r.File {
		names = append(names, f.Name)
	}
	return names
}

func contains(ss []string, target string) bool {
	for _, s := range ss {
		if s == target {
			return true
		}
	}
	return false
}

func readManifestFromZip(t *testing.T, archivePath string) Manifest {
	t.Helper()
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		t.Fatalf("cannot open zip %s: %v", archivePath, err)
	}
	defer r.Close()

	for _, f := range r.File {
		if f.Name != "manifest.json" {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			t.Fatalf("cannot open manifest.json: %v", err)
		}
		defer rc.Close()

		var m Manifest
		if err := json.NewDecoder(rc).Decode(&m); err != nil {
			t.Fatalf("cannot decode manifest.json: %v", err)
		}
		return m
	}

	t.Fatal("manifest.json not found in archive")
	return Manifest{}
}
