package sandbox

import (
	"fmt"
	"strings"
	"testing"
)

// capturedCall records arguments passed to the mocked execCommand.
type capturedCall struct {
	name string
	args []string
}

func mockExec(calls *[]capturedCall, err error) func() {
	orig := execCommand
	execCommand = func(name string, args ...string) ([]byte, error) {
		*calls = append(*calls, capturedCall{name: name, args: args})
		return nil, err
	}
	return func() { execCommand = orig }
}

func TestBuild_InvokesCorrectCommand(t *testing.T) {
	var calls []capturedCall
	restore := mockExec(&calls, nil)
	defer restore()

	rt := &RuntimeInfo{Name: "docker", Path: "/usr/bin/docker"}
	sandboxDir := "/home/user/myproject/.claude/sandbox"

	if err := Build(rt, sandboxDir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(calls))
	}

	c := calls[0]
	if c.name != "/usr/bin/docker" {
		t.Errorf("expected binary /usr/bin/docker, got %s", c.name)
	}
	if c.args[0] != "build" {
		t.Errorf("expected first arg 'build', got %s", c.args[0])
	}
	if c.args[2] != "ck-sandbox-myproject" {
		t.Errorf("expected image name ck-sandbox-myproject, got %s", c.args[2])
	}
	if c.args[4] != "/home/user/myproject/.claude/sandbox/Containerfile" {
		t.Errorf("expected containerfile path, got %s", c.args[4])
	}
	if c.args[5] != "/home/user/myproject" {
		t.Errorf("expected project dir /home/user/myproject, got %s", c.args[5])
	}
}

func TestRun_UsesCompose(t *testing.T) {
	var calls []capturedCall
	restore := mockExec(&calls, nil)
	defer restore()

	rt := &RuntimeInfo{Name: "podman", Path: "/usr/bin/podman"}
	sandboxDir := "/home/user/myproject/.claude/sandbox"

	if err := Run(rt, sandboxDir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(calls))
	}

	c := calls[0]
	if c.name != "/usr/bin/podman" {
		t.Errorf("expected binary /usr/bin/podman, got %s", c.name)
	}

	expected := []string{"compose", "-f", "/home/user/myproject/.claude/sandbox/compose.yaml", "up", "-d"}
	for i, want := range expected {
		if c.args[i] != want {
			t.Errorf("arg[%d]: expected %s, got %s", i, want, c.args[i])
		}
	}
}

func TestStop_GracefulTimeout(t *testing.T) {
	var calls []capturedCall
	restore := mockExec(&calls, nil)
	defer restore()

	rt := &RuntimeInfo{Name: "docker", Path: "/usr/bin/docker"}
	sandboxDir := "/home/user/myproject/.claude/sandbox"

	if err := Stop(rt, sandboxDir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(calls))
	}

	c := calls[0]
	expected := []string{"compose", "-f", "/home/user/myproject/.claude/sandbox/compose.yaml", "down", "--timeout", "10"}
	for i, want := range expected {
		if c.args[i] != want {
			t.Errorf("arg[%d]: expected %s, got %s", i, want, c.args[i])
		}
	}
}

func TestBuild_ErrorIncludesCommand(t *testing.T) {
	var calls []capturedCall
	restore := mockExec(&calls, fmt.Errorf("exit status 1"))
	defer restore()

	rt := &RuntimeInfo{Name: "docker", Path: "/usr/bin/docker"}
	sandboxDir := "/home/user/myproject/.claude/sandbox"

	err := Build(rt, sandboxDir)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "/usr/bin/docker") {
		t.Errorf("error should contain runtime path, got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "build") {
		t.Errorf("error should contain 'build' command, got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "exit status 1") {
		t.Errorf("error should contain original error, got: %s", errMsg)
	}
}
