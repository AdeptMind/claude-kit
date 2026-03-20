package sandbox

import (
	"fmt"
	"testing"
)

// saveFuncs saves and returns a restore function for the package-level function vars.
func saveFuncs() func() {
	origLookPath := lookPath
	origRunCommand := runCommand
	return func() {
		lookPath = origLookPath
		runCommand = origRunCommand
	}
}

func TestDetect_Docker(t *testing.T) {
	restore := saveFuncs()
	defer restore()

	lookPath = func(file string) (string, error) {
		if file == "docker" {
			return "/usr/bin/docker", nil
		}
		return "", fmt.Errorf("not found: %s", file)
	}
	runCommand = func(name string, args ...string) ([]byte, error) {
		if name == "docker" && len(args) == 1 && args[0] == "--version" {
			return []byte("Docker version 24.0.7, build afdd53b"), nil
		}
		if name == "docker" && len(args) == 2 && args[0] == "compose" && args[1] == "version" {
			return []byte("Docker Compose version v2.21.0"), nil
		}
		return nil, fmt.Errorf("unexpected: %s %v", name, args)
	}

	info, err := Detect("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Name != "docker" {
		t.Errorf("expected name=docker, got %s", info.Name)
	}
	if info.Path != "/usr/bin/docker" {
		t.Errorf("expected path=/usr/bin/docker, got %s", info.Path)
	}
	if info.Version != "24.0.7" {
		t.Errorf("expected version=24.0.7, got %s", info.Version)
	}
	if !info.ComposeAvailable {
		t.Error("expected compose_available=true")
	}
}

func TestDetect_Podman(t *testing.T) {
	restore := saveFuncs()
	defer restore()

	lookPath = func(file string) (string, error) {
		if file == "podman" {
			return "/usr/bin/podman", nil
		}
		return "", fmt.Errorf("not found: %s", file)
	}
	runCommand = func(name string, args ...string) ([]byte, error) {
		if name == "podman" && len(args) == 1 && args[0] == "--version" {
			return []byte("podman version 4.9.1"), nil
		}
		if name == "podman" && len(args) == 2 && args[0] == "compose" && args[1] == "version" {
			return []byte("podman compose version 1.0.6"), nil
		}
		return nil, fmt.Errorf("unexpected: %s %v", name, args)
	}

	info, err := Detect("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Name != "podman" {
		t.Errorf("expected name=podman, got %s", info.Name)
	}
	if info.Version != "4.9.1" {
		t.Errorf("expected version=4.9.1, got %s", info.Version)
	}
}

func TestDetect_NerdCtl(t *testing.T) {
	restore := saveFuncs()
	defer restore()

	lookPath = func(file string) (string, error) {
		if file == "nerdctl" {
			return "/usr/local/bin/nerdctl", nil
		}
		return "", fmt.Errorf("not found: %s", file)
	}
	runCommand = func(name string, args ...string) ([]byte, error) {
		if name == "nerdctl" && len(args) == 1 && args[0] == "--version" {
			return []byte("nerdctl version 1.7.2"), nil
		}
		if name == "nerdctl" && len(args) == 2 && args[0] == "compose" && args[1] == "version" {
			return []byte("nerdctl compose version 1.7.2"), nil
		}
		return nil, fmt.Errorf("unexpected: %s %v", name, args)
	}

	info, err := Detect("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Name != "nerdctl" {
		t.Errorf("expected name=nerdctl, got %s", info.Name)
	}
	if info.Path != "/usr/local/bin/nerdctl" {
		t.Errorf("expected path=/usr/local/bin/nerdctl, got %s", info.Path)
	}
	if info.Version != "1.7.2" {
		t.Errorf("expected version=1.7.2, got %s", info.Version)
	}
}

func TestDetect_NoRuntime(t *testing.T) {
	restore := saveFuncs()
	defer restore()

	lookPath = func(file string) (string, error) {
		return "", fmt.Errorf("not found: %s", file)
	}

	info, err := Detect("")
	if info != nil {
		t.Errorf("expected nil info, got %+v", info)
	}
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	msg := err.Error()
	if msg != "no container runtime found; checked docker, podman, nerdctl" {
		t.Errorf("unexpected error message: %s", msg)
	}
}

func TestDetect_Override(t *testing.T) {
	restore := saveFuncs()
	defer restore()

	lookPath = func(file string) (string, error) {
		switch file {
		case "docker":
			return "/usr/bin/docker", nil
		case "podman":
			return "/usr/bin/podman", nil
		}
		return "", fmt.Errorf("not found: %s", file)
	}
	runCommand = func(name string, args ...string) ([]byte, error) {
		if name == "podman" && len(args) == 1 && args[0] == "--version" {
			return []byte("podman version 4.9.1"), nil
		}
		if name == "podman" && len(args) == 2 && args[0] == "compose" {
			return nil, fmt.Errorf("compose not available")
		}
		return nil, fmt.Errorf("unexpected: %s %v", name, args)
	}

	info, err := Detect("podman")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Name != "podman" {
		t.Errorf("expected name=podman, got %s", info.Name)
	}
}

func TestDetect_ComposeUnavailable(t *testing.T) {
	restore := saveFuncs()
	defer restore()

	lookPath = func(file string) (string, error) {
		if file == "docker" {
			return "/usr/bin/docker", nil
		}
		return "", fmt.Errorf("not found: %s", file)
	}
	runCommand = func(name string, args ...string) ([]byte, error) {
		if name == "docker" && len(args) == 1 && args[0] == "--version" {
			return []byte("Docker version 24.0.7, build afdd53b"), nil
		}
		// Both compose checks fail.
		return nil, fmt.Errorf("compose not available")
	}

	info, err := Detect("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Name != "docker" {
		t.Errorf("expected name=docker, got %s", info.Name)
	}
	if info.ComposeAvailable {
		t.Error("expected compose_available=false")
	}
}
