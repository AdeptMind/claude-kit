package sandbox

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// RuntimeInfo describes a detected OCI container runtime.
type RuntimeInfo struct {
	Name             string `json:"name"`
	Path             string `json:"path"`
	Version          string `json:"version"`
	ComposeAvailable bool   `json:"compose_available"`
}

// Package-level function vars for testability.
var lookPath = exec.LookPath

var runCommand = func(name string, args ...string) ([]byte, error) {
	return exec.Command(name, args...).Output()
}

// runtimes defines the detection priority order.
var runtimes = []string{"docker", "podman", "nerdctl"}

// versionRe matches version strings like "24.0.7", "4.9.1", "1.7.2".
var versionRe = regexp.MustCompile(`(\d+\.\d+\.\d+)`)

// Detect finds an available OCI container runtime.
// If override is non-empty, only that specific runtime is checked.
// Returns nil with a descriptive error when no runtime is found.
func Detect(override string) (*RuntimeInfo, error) {
	candidates := runtimes
	if override != "" {
		candidates = []string{override}
	}

	var checked []string
	for _, name := range candidates {
		info, err := probe(name)
		if err != nil {
			checked = append(checked, name)
			continue
		}
		return info, nil
	}

	return nil, fmt.Errorf("no container runtime found; checked %s", strings.Join(checked, ", "))
}

// probe checks if a runtime binary exists, parses its version, and tests compose support.
func probe(name string) (*RuntimeInfo, error) {
	path, err := lookPath(name)
	if err != nil {
		return nil, err
	}

	out, err := runCommand(name, "--version")
	if err != nil {
		return nil, fmt.Errorf("%s found but --version failed: %w", name, err)
	}

	version := parseVersion(out)

	compose := checkCompose(name)

	return &RuntimeInfo{
		Name:             name,
		Path:             path,
		Version:          version,
		ComposeAvailable: compose,
	}, nil
}

// parseVersion extracts a semver-like version string from runtime --version output.
// Examples:
//
//	"Docker version 24.0.7, build afdd53b" -> "24.0.7"
//	"podman version 4.9.1"                 -> "4.9.1"
//	"nerdctl version 1.7.2"                -> "1.7.2"
func parseVersion(output []byte) string {
	m := versionRe.Find(output)
	if m == nil {
		return "unknown"
	}
	return string(m)
}

// checkCompose tests whether the runtime supports a compose subcommand.
func checkCompose(name string) bool {
	// Try native compose subcommand first (docker compose, podman compose, nerdctl compose).
	if _, err := runCommand(name, "compose", "version"); err == nil {
		return true
	}
	// For docker, also try the standalone docker-compose binary.
	if name == "docker" {
		if _, err := runCommand("docker-compose", "--version"); err == nil {
			return true
		}
	}
	return false
}
