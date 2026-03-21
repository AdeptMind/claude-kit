package sandbox

import (
	"strings"
	"testing"

	"github.com/AdeptMind/infra-tool/claude-cli/internal/policy"
)

func TestFallback_NilSandbox(t *testing.T) {
	t.Run("nil spec returns nil", func(t *testing.T) {
		out, warnings := FallbackPolicy(nil)
		if out != nil {
			t.Fatal("expected nil spec back")
		}
		if warnings != nil {
			t.Fatalf("expected no warnings, got %v", warnings)
		}
	})

	t.Run("no sandbox returns original spec", func(t *testing.T) {
		spec := &policy.PolicySpec{Profile: "default"}
		out, warnings := FallbackPolicy(spec)
		if out != spec {
			t.Fatal("expected same spec pointer when no sandbox")
		}
		if warnings != nil {
			t.Fatalf("expected no warnings, got %v", warnings)
		}
	})

	t.Run("disabled sandbox returns original spec", func(t *testing.T) {
		spec := &policy.PolicySpec{
			Sandbox: &policy.SandboxPolicy{Enabled: false},
		}
		out, warnings := FallbackPolicy(spec)
		if out != spec {
			t.Fatal("expected same spec pointer when sandbox disabled")
		}
		if warnings != nil {
			t.Fatalf("expected no warnings, got %v", warnings)
		}
	})
}

func TestFallback_DenyFromWritableMounts(t *testing.T) {
	spec := &policy.PolicySpec{
		Sandbox: &policy.SandboxPolicy{
			Enabled:        true,
			WritableMounts: []string{"src", "build"},
		},
	}

	out, _ := FallbackPolicy(spec)

	if out.Filesystem == nil {
		t.Fatal("expected Filesystem to be set")
	}

	if len(out.Filesystem.Deny) == 0 {
		t.Fatal("expected deny rules to be generated")
	}

	// Should contain the catch-all deny pattern.
	found := false
	for _, d := range out.Filesystem.Deny {
		if d == "**" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected '**' deny pattern, got %v", out.Filesystem.Deny)
	}

	// Should have a PreToolUse hook.
	if out.Hooks == nil || len(out.Hooks.PreToolUse) == 0 {
		t.Fatal("expected PreToolUse hook to be generated")
	}

	// Original spec should not be modified.
	if spec.Filesystem != nil {
		t.Fatal("original spec should not be mutated")
	}
}

func TestFallback_ExistingDenyMerged(t *testing.T) {
	spec := &policy.PolicySpec{
		Filesystem: &policy.FilesystemPolicy{
			Deny: []string{"/etc/passwd"},
		},
		Sandbox: &policy.SandboxPolicy{
			Enabled:        true,
			WritableMounts: []string{"src"},
		},
	}

	out, _ := FallbackPolicy(spec)

	if len(out.Filesystem.Deny) < 2 {
		t.Fatalf("expected existing deny + new deny, got %v", out.Filesystem.Deny)
	}

	hasEtc := false
	hasStar := false
	for _, d := range out.Filesystem.Deny {
		if d == "/etc/passwd" {
			hasEtc = true
		}
		if d == "**" {
			hasStar = true
		}
	}
	if !hasEtc || !hasStar {
		t.Fatalf("expected both existing and new deny rules, got %v", out.Filesystem.Deny)
	}

	// Original must not be mutated.
	if len(spec.Filesystem.Deny) != 1 {
		t.Fatal("original filesystem deny was mutated")
	}
}

func TestFallback_NetworkWarning(t *testing.T) {
	tests := []struct {
		name        string
		mode        string
		expectWarn  bool
	}{
		{"block_all triggers warning", "block_all", true},
		{"allowlist triggers warning", "allowlist", true},
		{"allow_all no warning", "allow_all", false},
		{"empty mode no warning", "", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			spec := &policy.PolicySpec{
				Sandbox: &policy.SandboxPolicy{Enabled: true},
				Network: &policy.NetworkPolicy{Mode: tc.mode},
			}

			_, warnings := FallbackPolicy(spec)

			hasNetworkWarn := false
			for _, w := range warnings {
				if strings.Contains(w, "network restrictions") {
					hasNetworkWarn = true
					break
				}
			}

			if tc.expectWarn && !hasNetworkWarn {
				t.Fatalf("expected network warning for mode=%q, got warnings: %v", tc.mode, warnings)
			}
			if !tc.expectWarn && hasNetworkWarn {
				t.Fatalf("did not expect network warning for mode=%q, got warnings: %v", tc.mode, warnings)
			}
		})
	}
}

func TestFallback_WarningsListComplete(t *testing.T) {
	spec := &policy.PolicySpec{
		Sandbox: &policy.SandboxPolicy{
			Enabled:        true,
			WritableMounts: []string{"src"},
			EnvPassthrough: []string{"API_KEY"},
		},
		Network: &policy.NetworkPolicy{Mode: "block_all"},
	}

	_, warnings := FallbackPolicy(spec)

	expected := []string{
		"filesystem isolation",
		"network restrictions",
		"env_passthrough",
		"process isolation",
	}

	for _, exp := range expected {
		found := false
		for _, w := range warnings {
			if strings.Contains(w, exp) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected warning containing %q, got: %v", exp, warnings)
		}
	}
}

func TestFallback_PostToolUseAuditHook(t *testing.T) {
	spec := &policy.PolicySpec{
		Sandbox: &policy.SandboxPolicy{Enabled: true},
	}

	out, _ := FallbackPolicy(spec)

	if out.Hooks == nil || len(out.Hooks.PostToolUse) == 0 {
		t.Fatal("expected PostToolUse audit hook")
	}

	hook := out.Hooks.PostToolUse[0]
	if hook.Matcher != "*" {
		t.Errorf("expected matcher='*', got %q", hook.Matcher)
	}
	if !strings.Contains(hook.Command, "fallback-audit") {
		t.Errorf("expected audit command, got %q", hook.Command)
	}
}

func TestFormatWarnings(t *testing.T) {
	t.Run("empty warnings returns empty", func(t *testing.T) {
		result := FormatWarnings(nil)
		if result != "" {
			t.Fatalf("expected empty, got %q", result)
		}
	})

	t.Run("formats warnings", func(t *testing.T) {
		result := FormatWarnings([]string{"warn1", "warn2"})
		if !strings.Contains(result, "warn1") || !strings.Contains(result, "warn2") {
			t.Fatalf("expected both warnings in output, got %q", result)
		}
		if !strings.Contains(result, "ENFORCED") {
			t.Fatal("expected ENFORCED summary section")
		}
		if !strings.Contains(result, "NOT ENFORCED") {
			t.Fatal("expected NOT ENFORCED summary section")
		}
	})
}
