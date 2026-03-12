package packaging

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/AdeptMind/infra-tool/claude-cli/internal/validate"
)

// Manifest holds metadata written as manifest.json inside every package.
type Manifest struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	Type      string `json:"type"` // "skill", "agent", "claude-kit"
	CreatedAt string `json:"created_at"`
}

// PackageOptions controls packaging behavior.
type PackageOptions struct {
	SkipValidation  bool
	ExcludePatterns []string // additional exclude patterns
}

// defaultExcludes lists paths and patterns excluded from every package.
var defaultExcludes = []string{
	".git",
	"node_modules",
	"__pycache__",
	".DS_Store",
	"*.pyc",
	".env",
	".venv",
	"dist",
	"build",
}

// PackageSkill validates and packages a skill directory into a .skill archive.
// outputPath is the directory where the archive will be written.
func PackageSkill(skillDir, outputPath string, opts PackageOptions) error {
	if !opts.SkipValidation {
		if errs := validate.ValidateSkill(skillDir); len(errs) > 0 {
			return fmt.Errorf("validation failed: %s", errs[0].Error())
		}
	}

	fields, err := readFrontmatter(filepath.Join(skillDir, "SKILL.md"))
	if err != nil {
		return fmt.Errorf("reading skill frontmatter: %w", err)
	}

	name := fields["name"]
	if name == "" {
		name = filepath.Base(skillDir)
	}

	manifest := Manifest{
		Name:      name,
		Version:   fieldOrDefault(fields, "version", "0.0.0"),
		Type:      "skill",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	dest := filepath.Join(outputPath, name+".skill")
	return createZip(dest, manifest, skillDir, mergeExcludes(opts.ExcludePatterns))
}

// PackageAgent validates and packages an agent file into a .agent archive.
// agentPath is the .md file; outputPath is the destination directory.
func PackageAgent(agentPath, outputPath string, opts PackageOptions) error {
	if !opts.SkipValidation {
		skillsDir := filepath.Join(filepath.Dir(agentPath), "..", "skills")
		if errs := validate.ValidateAgent(agentPath, skillsDir); len(errs) > 0 {
			return fmt.Errorf("validation failed: %s", errs[0].Error())
		}
	}

	fields, err := readFrontmatter(agentPath)
	if err != nil {
		return fmt.Errorf("reading agent frontmatter: %w", err)
	}

	name := fields["name"]
	if name == "" {
		name = strings.TrimSuffix(filepath.Base(agentPath), ".md")
	}

	manifest := Manifest{
		Name:      name,
		Version:   fieldOrDefault(fields, "version", "0.0.0"),
		Type:      "agent",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	dest := filepath.Join(outputPath, name+".agent")
	rootDir := filepath.Dir(agentPath)
	return createZip(dest, manifest, rootDir, mergeExcludes(opts.ExcludePatterns))
}

// PackageTemplate packages an entire template directory into a .claude-kit archive.
func PackageTemplate(templateDir, outputPath string, opts PackageOptions) error {
	name := filepath.Base(templateDir)

	manifest := Manifest{
		Name:      name,
		Version:   "0.0.0",
		Type:      "claude-kit",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	dest := filepath.Join(outputPath, name+".claude-kit")
	return createZip(dest, manifest, templateDir, mergeExcludes(opts.ExcludePatterns))
}

// createZip builds a zip archive at dest with manifest.json at the root,
// then walks rootDir adding all files that pass exclusion checks.
func createZip(dest string, manifest Manifest, rootDir string, excludes []string) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	f, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("creating archive file: %w", err)
	}
	defer f.Close()

	w := zip.NewWriter(f)
	defer w.Close()

	// Write manifest.json first
	manifestData, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling manifest: %w", err)
	}

	mw, err := w.Create("manifest.json")
	if err != nil {
		return fmt.Errorf("creating manifest entry: %w", err)
	}
	if _, err := mw.Write(manifestData); err != nil {
		return fmt.Errorf("writing manifest: %w", err)
	}

	// Walk and add files
	return filepath.Walk(rootDir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		rel, err := filepath.Rel(rootDir, path)
		if err != nil {
			return err
		}

		if rel == "." {
			return nil
		}

		if shouldExclude(rel, excludes) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if info.IsDir() {
			return nil
		}

		fw, err := w.Create(filepath.ToSlash(rel))
		if err != nil {
			return fmt.Errorf("creating zip entry %s: %w", rel, err)
		}

		src, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("opening %s: %w", path, err)
		}
		defer src.Close()

		_, err = io.Copy(fw, src)
		return err
	})
}

// shouldExclude returns true if the relative path matches any exclude pattern.
func shouldExclude(rel string, excludes []string) bool {
	base := filepath.Base(rel)
	for _, pattern := range excludes {
		// Exact directory/file name match (e.g. ".git", "node_modules")
		if base == pattern {
			return true
		}
		// Check each path component
		for _, part := range strings.Split(rel, string(filepath.Separator)) {
			if part == pattern {
				return true
			}
		}
		// Glob match on basename (e.g. "*.pyc")
		if matched, _ := filepath.Match(pattern, base); matched {
			return true
		}
	}
	return false
}

// mergeExcludes combines default excludes with additional patterns.
func mergeExcludes(additional []string) []string {
	excludes := make([]string, len(defaultExcludes))
	copy(excludes, defaultExcludes)
	excludes = append(excludes, additional...)
	return excludes
}

// readFrontmatter extracts key-value pairs from YAML frontmatter in a markdown file.
func readFrontmatter(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	content := string(data)
	if !strings.HasPrefix(strings.TrimSpace(content), "---") {
		return map[string]string{}, nil
	}

	lines := strings.Split(content, "\n")
	fields := make(map[string]string)
	inFrontmatter := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "---" {
			if inFrontmatter {
				break
			}
			inFrontmatter = true
			continue
		}
		if !inFrontmatter {
			continue
		}
		if len(line) > 0 && (line[0] == ' ' || line[0] == '\t' || line[0] == '-') {
			continue
		}
		idx := strings.Index(line, ":")
		if idx < 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])
		if len(val) >= 2 && ((val[0] == '"' && val[len(val)-1] == '"') || (val[0] == '\'' && val[len(val)-1] == '\'')) {
			val = val[1 : len(val)-1]
		}
		fields[key] = val
	}

	return fields, nil
}

// fieldOrDefault returns the field value or a default if empty/missing.
func fieldOrDefault(fields map[string]string, key, fallback string) string {
	if v := fields[key]; v != "" {
		return v
	}
	return fallback
}
