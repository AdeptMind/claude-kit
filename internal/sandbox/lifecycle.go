package sandbox

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// execCommand is an injectable function var for testability (same pattern as detect.go).
var execCommand = func(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	return out, err
}

// containerName returns the deterministic container name for a sandbox directory.
// The sandbox dir is expected to be inside the project dir (e.g. <projectDir>/.claude/sandbox).
func containerName(sandboxDir string) string {
	projectDir := filepath.Dir(filepath.Dir(sandboxDir))
	return "ck-sandbox-" + filepath.Base(projectDir)
}

// Build builds the container image from the Containerfile in sandboxDir.
func Build(runtime *RuntimeInfo, sandboxDir string) error {
	name := containerName(sandboxDir)
	projectDir := filepath.Dir(filepath.Dir(sandboxDir))
	containerfile := filepath.Join(sandboxDir, "Containerfile")

	args := []string{"build", "-t", name, "-f", containerfile, projectDir}
	_, err := execCommand(runtime.Path, args...)
	if err != nil {
		return fmt.Errorf("%s %s: %w", runtime.Path, strings.Join(args, " "), err)
	}
	return nil
}

// Run starts the container using compose.
func Run(runtime *RuntimeInfo, sandboxDir string) error {
	composeFile := filepath.Join(sandboxDir, "compose.yaml")

	args := []string{"compose", "-f", composeFile, "up", "-d"}
	_, err := execCommand(runtime.Path, args...)
	if err != nil {
		return fmt.Errorf("%s %s: %w", runtime.Path, strings.Join(args, " "), err)
	}
	return nil
}

// Stop stops and removes the container gracefully (10s timeout then force kill).
func Stop(runtime *RuntimeInfo, sandboxDir string) error {
	composeFile := filepath.Join(sandboxDir, "compose.yaml")

	args := []string{"compose", "-f", composeFile, "down", "--timeout", "10"}
	_, err := execCommand(runtime.Path, args...)
	if err != nil {
		return fmt.Errorf("%s %s: %w", runtime.Path, strings.Join(args, " "), err)
	}
	return nil
}
