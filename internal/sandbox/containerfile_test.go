package sandbox

import (
	"strings"
	"testing"

	"github.com/AdeptMind/infra-tool/claude-cli/internal/policy"
)

func TestGenerateContainerfile_DefaultImage(t *testing.T) {
	out, err := GenerateContainerfile(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(out, "FROM node:22-slim\n") {
		t.Errorf("expected default base image, got:\n%s", out)
	}
}

func TestGenerateContainerfile_CustomImage(t *testing.T) {
	p := &policy.SandboxPolicy{ImageBase: "ubuntu:24.04"}
	out, err := GenerateContainerfile(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(out, "FROM ubuntu:24.04\n") {
		t.Errorf("expected custom base image, got:\n%s", out)
	}
}

func TestGenerateContainerfile_NonRootUser(t *testing.T) {
	out, err := GenerateContainerfile(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "USER claude") {
		t.Error("expected USER directive for non-root user")
	}
	if !strings.Contains(out, "useradd -r -g claude") {
		t.Error("expected useradd for claude user")
	}
}

func TestGenerateContainerfile_InstallsClaude(t *testing.T) {
	out, err := GenerateContainerfile(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "npm install -g @anthropic-ai/claude-code") {
		t.Error("expected npm install of claude-code")
	}
}
