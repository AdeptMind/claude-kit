package roles

import (
	"os"
	"path/filepath"
	"testing"
)

const testRolesYAML = `roles:
  dev:
    description: "Solo developer"
    workflow:
      inter_round_gate: false
      client_challenge: false
      readiness_check: none
      done_check: technical
    reporting:
      language: technical
      proof: tests
    analysis:
      traceability: shallow
      challenge_so_that: false

  po:
    description: "Product owner"
    workflow:
      inter_round_gate: true
      client_challenge: true
      readiness_check: full
      done_check: functional
    reporting:
      language: product
      proof: visual
    analysis:
      traceability: deep
      challenge_so_that: true

  architect:
    description: "Software architect"
    workflow:
      inter_round_gate: false
      client_challenge: false
      readiness_check: light
      done_check: technical
    reporting:
      language: technical
      proof: tests
    analysis:
      traceability: shallow
      challenge_so_that: false

  all:
    description: "Full role orchestration"
    workflow:
      inter_round_gate: true
      client_challenge: true
      readiness_check: full
      done_check: combined
    reporting:
      language: product
      proof: combined
    analysis:
      traceability: deep
      challenge_so_that: true
    phases:
      break:
        - po
      model:
        - architect
      analyze:
        - po
        - architect
      act:
        - dev
        - architect
      deliver:
        - po
`

func writeTestRoles(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "roles.yaml")
	if err := os.WriteFile(path, []byte(testRolesYAML), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

// setupEnvRole sets CK_USER_ROLE and BMAD_TEMPLATE_DIR for a test, restoring them on cleanup.
func setupEnvRole(t *testing.T, role, templateDir string) {
	t.Helper()

	origRole := os.Getenv(envRoleKey)
	origDir := os.Getenv("BMAD_TEMPLATE_DIR")

	os.Setenv(envRoleKey, role)
	os.Setenv("BMAD_TEMPLATE_DIR", templateDir)

	t.Cleanup(func() {
		os.Setenv(envRoleKey, origRole)
		os.Setenv("BMAD_TEMPLATE_DIR", origDir)
	})
}

func TestLoadRoles_ValidFile(t *testing.T) {
	path := writeTestRoles(t)

	cfg, err := LoadRoles(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cfg.Roles) != 4 {
		t.Errorf("expected 4 roles, got %d", len(cfg.Roles))
	}

	dev, ok := cfg.Roles["dev"]
	if !ok {
		t.Fatal("expected 'dev' role")
	}
	if dev.Description != "Solo developer" {
		t.Errorf("unexpected description: %s", dev.Description)
	}
	if dev.Workflow.InterRoundGate {
		t.Error("dev inter_round_gate should be false")
	}
	if dev.Reporting.Language != "technical" {
		t.Errorf("unexpected reporting language: %s", dev.Reporting.Language)
	}
}

func TestLoadRoles_InvalidPath(t *testing.T) {
	_, err := LoadRoles("/nonexistent/roles.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadRoles_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "roles.yaml")
	os.WriteFile(path, []byte(":::bad yaml"), 0644)

	_, err := LoadRoles(path)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestLoadRoles_EmptyRoles(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "roles.yaml")
	os.WriteFile(path, []byte("roles: {}"), 0644)

	_, err := LoadRoles(path)
	if err == nil {
		t.Fatal("expected error for empty roles")
	}
}

func TestActiveRole_Default(t *testing.T) {
	path := writeTestRoles(t)
	dir := filepath.Dir(path)
	setupEnvRole(t, "", dir)

	role, name, err := ActiveRole()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "dev" {
		t.Errorf("expected default role 'dev', got %q", name)
	}
	if role.Description != "Solo developer" {
		t.Errorf("unexpected description: %s", role.Description)
	}
}

func TestActiveRole_Set(t *testing.T) {
	path := writeTestRoles(t)
	dir := filepath.Dir(path)
	setupEnvRole(t, "po", dir)

	role, name, err := ActiveRole()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "po" {
		t.Errorf("expected role 'po', got %q", name)
	}
	if !role.Workflow.InterRoundGate {
		t.Error("po inter_round_gate should be true")
	}
	if !role.Workflow.ClientChallenge {
		t.Error("po client_challenge should be true")
	}
}

func TestActiveRole_Invalid(t *testing.T) {
	path := writeTestRoles(t)
	dir := filepath.Dir(path)
	setupEnvRole(t, "unknown", dir)

	_, _, err := ActiveRole()
	if err == nil {
		t.Fatal("expected error for unknown role")
	}
	// Should list valid roles in error message.
	errMsg := err.Error()
	if !contains(errMsg, "unknown") || !contains(errMsg, "valid roles") {
		t.Errorf("error should mention unknown role and valid roles: %s", errMsg)
	}
}

func TestIsGateActive_DevMode(t *testing.T) {
	path := writeTestRoles(t)
	dir := filepath.Dir(path)
	setupEnvRole(t, "dev", dir)

	tests := []struct {
		gate   string
		expect bool
	}{
		{"inter_round_gate", false},
		{"client_challenge", false},
		{"readiness_check", false},  // "none" => false
		{"done_check", false},       // "technical" => false (base level)
		{"nonexistent_gate", false},
	}

	for _, tt := range tests {
		t.Run(tt.gate, func(t *testing.T) {
			if got := IsGateActive(tt.gate); got != tt.expect {
				t.Errorf("IsGateActive(%q) = %v, want %v", tt.gate, got, tt.expect)
			}
		})
	}
}

func TestIsGateActive_POMode(t *testing.T) {
	path := writeTestRoles(t)
	dir := filepath.Dir(path)
	setupEnvRole(t, "po", dir)

	tests := []struct {
		gate   string
		expect bool
	}{
		{"inter_round_gate", true},
		{"client_challenge", true},
		{"readiness_check", true},  // "full" => true
		{"done_check", true},       // "functional" => true (enhanced)
	}

	for _, tt := range tests {
		t.Run(tt.gate, func(t *testing.T) {
			if got := IsGateActive(tt.gate); got != tt.expect {
				t.Errorf("IsGateActive(%q) = %v, want %v", tt.gate, got, tt.expect)
			}
		})
	}
}

func TestIsGateActive_ArchitectMode(t *testing.T) {
	path := writeTestRoles(t)
	dir := filepath.Dir(path)
	setupEnvRole(t, "architect", dir)

	// readiness_check: "light" => true
	if !IsGateActive("readiness_check") {
		t.Error("architect readiness_check (light) should be active")
	}
	// done_check: "technical" => false
	if IsGateActive("done_check") {
		t.Error("architect done_check (technical) should not be active")
	}
}

func TestGetPhaseRoles_Act(t *testing.T) {
	path := writeTestRoles(t)
	dir := filepath.Dir(path)
	setupEnvRole(t, "", dir)

	roles, err := GetPhaseRoles("act")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{"dev", "architect"}
	if len(roles) != len(expected) {
		t.Fatalf("expected %d roles, got %d: %v", len(expected), len(roles), roles)
	}
	for i, r := range expected {
		if roles[i] != r {
			t.Errorf("role[%d] = %q, want %q", i, roles[i], r)
		}
	}
}

func TestGetPhaseRoles_AllPhases(t *testing.T) {
	path := writeTestRoles(t)
	dir := filepath.Dir(path)
	setupEnvRole(t, "", dir)

	phases := []string{"break", "model", "analyze", "act", "deliver"}
	for _, phase := range phases {
		t.Run(phase, func(t *testing.T) {
			roles, err := GetPhaseRoles(phase)
			if err != nil {
				t.Fatalf("unexpected error for phase %q: %v", phase, err)
			}
			if len(roles) == 0 {
				t.Errorf("expected at least one role for phase %q", phase)
			}
		})
	}
}

func TestGetPhaseRoles_InvalidPhase(t *testing.T) {
	path := writeTestRoles(t)
	dir := filepath.Dir(path)
	setupEnvRole(t, "", dir)

	_, err := GetPhaseRoles("nonexistent")
	if err == nil {
		t.Fatal("expected error for unknown phase")
	}
	if !contains(err.Error(), "unknown phase") {
		t.Errorf("error should mention unknown phase: %s", err.Error())
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
