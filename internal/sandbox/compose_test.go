package sandbox

import (
	"strings"
	"testing"

	"github.com/AdeptMind/infra-tool/claude-cli/internal/policy"
)

func TestGenerateCompose_ReadOnlyMount(t *testing.T) {
	spec := &policy.PolicySpec{
		Sandbox: &policy.SandboxPolicy{Enabled: true},
	}

	out, err := GenerateCompose(spec, "/home/user/myproject")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, `"/home/user/myproject:/workspace:ro"`) {
		t.Error("expected read-only project mount")
	}
	if !strings.Contains(out, "container_name: ck-sandbox-myproject") {
		t.Error("expected container name derived from project dir basename")
	}
	if strings.Contains(out, ":rw") {
		t.Error("expected no writable mounts")
	}
}

func TestGenerateCompose_WritableMounts(t *testing.T) {
	spec := &policy.PolicySpec{
		Sandbox: &policy.SandboxPolicy{
			Enabled:        true,
			WritableMounts: []string{".claude/output", "tmp"},
		},
	}

	out, err := GenerateCompose(spec, "/home/user/myproject")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, `"/home/user/myproject/.claude/output:/workspace/.claude/output:rw"`) {
		t.Error("expected writable mount for .claude/output")
	}
	if !strings.Contains(out, `"/home/user/myproject/tmp:/workspace/tmp:rw"`) {
		t.Error("expected writable mount for tmp")
	}
}

func TestGenerateCompose_NetworkNone(t *testing.T) {
	spec := &policy.PolicySpec{
		Sandbox: &policy.SandboxPolicy{Enabled: true},
		Network: &policy.NetworkPolicy{Mode: "block_all"},
	}

	out, err := GenerateCompose(spec, "/home/user/myproject")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, `network_mode: "none"`) {
		t.Error("expected network_mode none for block_all")
	}
}

func TestGenerateCompose_NetworkNone_NotSetForOtherModes(t *testing.T) {
	spec := &policy.PolicySpec{
		Sandbox: &policy.SandboxPolicy{Enabled: true},
		Network: &policy.NetworkPolicy{Mode: "allow_all"},
	}

	out, err := GenerateCompose(spec, "/home/user/myproject")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Contains(out, "network_mode") {
		t.Error("expected no network_mode for allow_all")
	}
}

func TestGenerateCompose_EnvPassthrough(t *testing.T) {
	spec := &policy.PolicySpec{
		Sandbox: &policy.SandboxPolicy{
			Enabled:        true,
			EnvPassthrough: []string{"HOME", "GITHUB_TOKEN"},
		},
	}

	out, err := GenerateCompose(spec, "/home/user/myproject")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "environment:") {
		t.Error("expected environment section")
	}
	if !strings.Contains(out, `"HOME"`) {
		t.Error("expected HOME in environment")
	}
	if !strings.Contains(out, `"GITHUB_TOKEN"`) {
		t.Error("expected GITHUB_TOKEN in environment")
	}
}

func TestGenerateCompose_FullConfig(t *testing.T) {
	spec := &policy.PolicySpec{
		Sandbox: &policy.SandboxPolicy{
			Enabled:        true,
			WritableMounts: []string{"dist"},
			EnvPassthrough: []string{"PATH"},
		},
		Network: &policy.NetworkPolicy{Mode: "block_all"},
	}

	out, err := GenerateCompose(spec, "/opt/projects/webapp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	checks := []struct {
		name string
		want string
	}{
		{"ro mount", `"/opt/projects/webapp:/workspace:ro"`},
		{"rw mount", `"/opt/projects/webapp/dist:/workspace/dist:rw"`},
		{"network none", `network_mode: "none"`},
		{"env var", `"PATH"`},
		{"working dir", "working_dir: /workspace"},
		{"stdin", "stdin_open: true"},
		{"tty", "tty: true"},
		{"container name", "ck-sandbox-webapp"},
	}

	for _, c := range checks {
		if !strings.Contains(out, c.want) {
			t.Errorf("%s: expected %q in output", c.name, c.want)
		}
	}
}

func TestGenerateCompose_NilSpec(t *testing.T) {
	_, err := GenerateCompose(nil, "/tmp/project")
	if err == nil {
		t.Error("expected error for nil spec")
	}
}
