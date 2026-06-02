package policy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseBytes(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr string // empty means no error expected
		check   func(t *testing.T, spec *PolicySpec)
	}{
		{
			name: "full policy with profile",
			yaml: `
profile: strict
filesystem:
  deny:
    - "/etc/**"
  allow:
    - "/tmp/**"
network:
  mode: allowlist
  allowlist:
    - "api.example.com"
hooks:
  pre_tool_use:
    - matcher: "Bash"
      command: "echo pre"
  post_tool_use:
    - matcher: "Write"
      command: "echo post"
sandbox:
  enabled: true
  runtime: docker
  writable_mounts:
    - "/workspace"
  env_passthrough:
    - "HOME"
  image_base: "ubuntu:22.04"
`,
			check: func(t *testing.T, spec *PolicySpec) {
				if spec.Profile != "strict" {
					t.Errorf("profile = %q, want %q", spec.Profile, "strict")
				}
				if spec.Extends != "" {
					t.Error("extends should be empty")
				}
				if len(spec.Filesystem.Deny) != 1 || spec.Filesystem.Deny[0] != "/etc/**" {
					t.Errorf("filesystem.deny = %v", spec.Filesystem.Deny)
				}
				if len(spec.Filesystem.Allow) != 1 || spec.Filesystem.Allow[0] != "/tmp/**" {
					t.Errorf("filesystem.allow = %v", spec.Filesystem.Allow)
				}
				if spec.Network.Mode != "allowlist" {
					t.Errorf("network.mode = %q", spec.Network.Mode)
				}
				if len(spec.Network.Allowlist) != 1 {
					t.Errorf("network.allowlist = %v", spec.Network.Allowlist)
				}
				if len(spec.Hooks.PreToolUse) != 1 || spec.Hooks.PreToolUse[0].Matcher != "Bash" {
					t.Errorf("hooks.pre_tool_use = %v", spec.Hooks.PreToolUse)
				}
				if len(spec.Hooks.PostToolUse) != 1 || spec.Hooks.PostToolUse[0].Command != "echo post" {
					t.Errorf("hooks.post_tool_use = %v", spec.Hooks.PostToolUse)
				}
				if !spec.Sandbox.Enabled || spec.Sandbox.Runtime != "docker" {
					t.Errorf("sandbox = %+v", spec.Sandbox)
				}
				if len(spec.Sandbox.WritableMounts) != 1 || spec.Sandbox.WritableMounts[0] != "/workspace" {
					t.Errorf("sandbox.writable_mounts = %v", spec.Sandbox.WritableMounts)
				}
				if spec.Sandbox.ImageBase != "ubuntu:22.04" {
					t.Errorf("sandbox.image_base = %q", spec.Sandbox.ImageBase)
				}
			},
		},
		{
			name: "minimal valid policy with profile only",
			yaml: `profile: permissive`,
			check: func(t *testing.T, spec *PolicySpec) {
				if spec.Profile != "permissive" {
					t.Errorf("profile = %q, want %q", spec.Profile, "permissive")
				}
			},
		},
		{
			name: "valid policy with extends",
			yaml: `
extends: moderate
network:
  mode: block_all
`,
			check: func(t *testing.T, spec *PolicySpec) {
				if spec.Extends != "moderate" {
					t.Errorf("extends = %q, want %q", spec.Extends, "moderate")
				}
				if spec.Network.Mode != "block_all" {
					t.Errorf("network.mode = %q", spec.Network.Mode)
				}
			},
		},
		{
			name:    "missing both profile and extends",
			yaml:    `network:\n  mode: block_all`,
			wantErr: "neither is set",
		},
		{
			name: "both profile and extends set",
			yaml: `
profile: strict
extends: moderate
`,
			wantErr: "mutually exclusive",
		},
		{
			name:    "invalid profile value",
			yaml:    `profile: paranoid`,
			wantErr: `profile "paranoid" is not valid`,
		},
		{
			name:    "invalid extends value",
			yaml:    `extends: custom`,
			wantErr: `extends "custom" is not valid`,
		},
		{
			name: "invalid network mode",
			yaml: `
profile: strict
network:
  mode: open
`,
			wantErr: `network.mode "open" is not valid`,
		},
		{
			name: "invalid sandbox runtime",
			yaml: `
profile: strict
sandbox:
  enabled: true
  runtime: lxc
`,
			wantErr: `sandbox.runtime "lxc" is not valid`,
		},
		{
			name: "sandbox without runtime is valid",
			yaml: `
profile: moderate
sandbox:
  enabled: true
`,
			check: func(t *testing.T, spec *PolicySpec) {
				if !spec.Sandbox.Enabled {
					t.Error("sandbox should be enabled")
				}
				if spec.Sandbox.Runtime != "" {
					t.Errorf("sandbox.runtime = %q, want empty", spec.Sandbox.Runtime)
				}
			},
		},
		{
			name:    "invalid YAML syntax",
			yaml:    `profile: [unclosed`,
			wantErr: "invalid YAML",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec, err := ParseBytes([]byte(tt.yaml))

			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error %q does not contain %q", err.Error(), tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.check != nil {
				tt.check(t, spec)
			}
		})
	}
}

func TestParseFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "policy.yaml")

	content := []byte("profile: strict\n")
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatal(err)
	}

	spec, err := ParseFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if spec.Profile != "strict" {
		t.Errorf("profile = %q, want %q", spec.Profile, "strict")
	}
}

func TestParseFile_NotFound(t *testing.T) {
	_, err := ParseFile("/nonexistent/policy.yaml")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
	if !strings.Contains(err.Error(), "policy: reading") {
		t.Errorf("error %q does not contain expected prefix", err.Error())
	}
}
