package packaging

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// InstallResult holds the outcome of an install operation.
type InstallResult struct {
	Type         string   // "skill", "agent", "claude-kit"
	Name         string
	Version      string
	FilesCreated []string
	Conflicts    []string // files that were overwritten (when force=true)
}

// Install extracts an archive (.skill, .agent, .claude-kit) into targetDir.
// If force is false and existing files would be overwritten, it returns an error
// listing the conflicts. If force is true, conflicts are overwritten and recorded
// in the result.
func Install(archivePath, targetDir string, force bool) (*InstallResult, error) {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return nil, fmt.Errorf("opening archive: %w", err)
	}
	defer r.Close()

	manifest, err := readManifestFromArchive(&r.Reader)
	if err != nil {
		return nil, fmt.Errorf("reading manifest: %w", err)
	}

	if err := validateManifest(manifest); err != nil {
		return nil, err
	}

	extractDir := extractTarget(targetDir, manifest)

	conflicts := listConflicts(&r.Reader, extractDir)
	if len(conflicts) > 0 && !force {
		return nil, &ConflictError{Conflicts: conflicts}
	}

	created, err := extractFiles(&r.Reader, extractDir)
	if err != nil {
		return nil, fmt.Errorf("extracting files: %w", err)
	}

	result := &InstallResult{
		Type:         manifest.Type,
		Name:         manifest.Name,
		Version:      manifest.Version,
		FilesCreated: created,
	}
	if force {
		result.Conflicts = conflicts
	}
	return result, nil
}

// DetectArchiveType reads the manifest from a zip archive and returns its type.
// Falls back to the file extension if manifest is missing or unreadable.
func DetectArchiveType(archivePath string) (string, error) {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return "", fmt.Errorf("opening archive: %w", err)
	}
	defer r.Close()

	manifest, err := readManifestFromArchive(&r.Reader)
	if err == nil && manifest.Type != "" {
		return manifest.Type, nil
	}

	return typeFromExtension(archivePath)
}

// ConflictError is returned when an install would overwrite existing files
// and force is not enabled.
type ConflictError struct {
	Conflicts []string
}

func (e *ConflictError) Error() string {
	return fmt.Sprintf("install would overwrite %d existing file(s): %s",
		len(e.Conflicts), strings.Join(e.Conflicts, ", "))
}

// readManifestFromArchive finds and decodes manifest.json inside a zip.Reader.
func readManifestFromArchive(r *zip.Reader) (Manifest, error) {
	for _, f := range r.File {
		if f.Name != "manifest.json" {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return Manifest{}, fmt.Errorf("opening manifest.json: %w", err)
		}
		defer rc.Close()

		var m Manifest
		if err := json.NewDecoder(rc).Decode(&m); err != nil {
			return Manifest{}, fmt.Errorf("decoding manifest.json: %w", err)
		}
		return m, nil
	}
	return Manifest{}, fmt.Errorf("manifest.json not found in archive")
}

// validateManifest checks that required fields are present.
func validateManifest(m Manifest) error {
	if m.Name == "" {
		return fmt.Errorf("manifest missing required field: name")
	}
	if m.Type == "" {
		return fmt.Errorf("manifest missing required field: type")
	}
	switch m.Type {
	case "skill", "agent", "claude-kit":
		// valid
	default:
		return fmt.Errorf("unknown manifest type: %q", m.Type)
	}
	return nil
}

// extractTarget determines the destination directory based on the manifest type.
func extractTarget(targetDir string, m Manifest) string {
	switch m.Type {
	case "skill":
		return filepath.Join(targetDir, "skills", m.Name)
	case "agent":
		return filepath.Join(targetDir, "agents")
	default: // claude-kit
		return targetDir
	}
}

// listConflicts returns relative paths of archive entries that already exist on disk.
func listConflicts(r *zip.Reader, targetDir string) []string {
	var conflicts []string
	for _, f := range r.File {
		if f.Name == "manifest.json" {
			continue
		}
		if strings.HasSuffix(f.Name, "/") {
			continue
		}
		dest := filepath.Join(targetDir, filepath.FromSlash(f.Name))
		if _, err := os.Stat(dest); err == nil {
			conflicts = append(conflicts, f.Name)
		}
	}
	return conflicts
}

// extractFiles writes all archive entries (except manifest.json) to targetDir.
func extractFiles(r *zip.Reader, targetDir string) ([]string, error) {
	var created []string

	for _, f := range r.File {
		if f.Name == "manifest.json" {
			continue
		}

		dest := filepath.Join(targetDir, filepath.FromSlash(f.Name))

		// Prevent zip slip
		if !strings.HasPrefix(filepath.Clean(dest), filepath.Clean(targetDir)+string(os.PathSeparator)) {
			return created, fmt.Errorf("illegal file path in archive: %s", f.Name)
		}

		if strings.HasSuffix(f.Name, "/") {
			if err := os.MkdirAll(dest, 0o755); err != nil {
				return created, fmt.Errorf("creating directory %s: %w", f.Name, err)
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return created, fmt.Errorf("creating parent directory for %s: %w", f.Name, err)
		}

		if err := extractFile(f, dest); err != nil {
			return created, err
		}
		created = append(created, dest)
	}

	return created, nil
}

// extractFile writes a single zip entry to disk.
func extractFile(f *zip.File, dest string) error {
	rc, err := f.Open()
	if err != nil {
		return fmt.Errorf("opening %s: %w", f.Name, err)
	}
	defer rc.Close()

	out, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("creating %s: %w", dest, err)
	}
	defer out.Close()

	if _, err := io.Copy(out, rc); err != nil {
		return fmt.Errorf("writing %s: %w", f.Name, err)
	}
	return nil
}

// typeFromExtension infers archive type from the file extension.
func typeFromExtension(path string) (string, error) {
	ext := strings.TrimPrefix(filepath.Ext(path), ".")
	switch ext {
	case "skill":
		return "skill", nil
	case "agent":
		return "agent", nil
	case "claude-kit":
		return "claude-kit", nil
	}
	return "", fmt.Errorf("cannot determine archive type from extension: %q", ext)
}
