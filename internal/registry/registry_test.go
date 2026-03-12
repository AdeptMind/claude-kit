package registry

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeAgentFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name+".md")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

const agentWithInterfaces = `---
name: backend
description: Activate for backend API work
model: claude-sonnet-4-6
skills:
  - code-reviewer
  - test-generator
interfaces:
  produces:
    - "API endpoints"
    - "**/*.go"
  consumes:
    - "architecture.yaml"
    - "API contracts"
---

## Rules
- DRY, KISS, SOLID
`

const agentWithoutInterfaces = `---
name: simple
description: A simple agent without interfaces
model: claude-sonnet-4-6
skills:
  - code-reviewer
---

## Rules
- Keep it simple
`

const agentPartialInterfaces = `---
name: partial
description: Agent with only produces
model: claude-sonnet-4-6
skills: []
interfaces:
  produces:
    - "reports"
---

## Content
`

const agentInlineSkills = `---
name: inline
description: Agent with inline skills list
model: claude-sonnet-4-6
skills: [reader, writer]
interfaces:
  produces: []
  consumes:
    - "data files"
---
`

func TestScanAgents_MultipleAgents(t *testing.T) {
	dir := t.TempDir()
	writeAgentFile(t, dir, "backend", agentWithInterfaces)
	writeAgentFile(t, dir, "simple", agentWithoutInterfaces)
	writeAgentFile(t, dir, "partial", agentPartialInterfaces)

	agents, err := ScanAgents(dir)
	if err != nil {
		t.Fatalf("ScanAgents returned error: %v", err)
	}

	if len(agents) != 3 {
		t.Fatalf("expected 3 agents, got %d", len(agents))
	}

	// Verify sorted by name
	if agents[0].Name != "backend" || agents[1].Name != "partial" || agents[2].Name != "simple" {
		t.Errorf("agents not sorted: got %s, %s, %s", agents[0].Name, agents[1].Name, agents[2].Name)
	}

	// Verify backend
	be := agents[0]
	if be.Description != "Activate for backend API work" {
		t.Errorf("backend description = %q", be.Description)
	}
	if be.Model != "claude-sonnet-4-6" {
		t.Errorf("backend model = %q", be.Model)
	}
	if len(be.Skills) != 2 || be.Skills[0] != "code-reviewer" || be.Skills[1] != "test-generator" {
		t.Errorf("backend skills = %v", be.Skills)
	}
	if len(be.Interfaces.Produces) != 2 || be.Interfaces.Produces[0] != "API endpoints" {
		t.Errorf("backend produces = %v", be.Interfaces.Produces)
	}
	if len(be.Interfaces.Consumes) != 2 || be.Interfaces.Consumes[0] != "architecture.yaml" {
		t.Errorf("backend consumes = %v", be.Interfaces.Consumes)
	}
}

func TestScanAgents_EmptyDir(t *testing.T) {
	dir := t.TempDir()

	agents, err := ScanAgents(dir)
	if err != nil {
		t.Fatalf("ScanAgents returned error: %v", err)
	}

	if len(agents) != 0 {
		t.Errorf("expected 0 agents, got %d", len(agents))
	}
}

func TestScanAgents_SkipsNonMdFiles(t *testing.T) {
	dir := t.TempDir()
	writeAgentFile(t, dir, "backend", agentWithInterfaces)
	// Write a non-.md file
	os.WriteFile(filepath.Join(dir, "README.txt"), []byte("not an agent"), 0o644)
	// Write a subdirectory
	os.MkdirAll(filepath.Join(dir, "subdir"), 0o755)

	agents, err := ScanAgents(dir)
	if err != nil {
		t.Fatalf("ScanAgents returned error: %v", err)
	}

	if len(agents) != 1 {
		t.Fatalf("expected 1 agent, got %d", len(agents))
	}
	if agents[0].Name != "backend" {
		t.Errorf("expected backend, got %s", agents[0].Name)
	}
}

func TestScanAgents_InvalidDir(t *testing.T) {
	_, err := ScanAgents("/nonexistent/path")
	if err == nil {
		t.Fatal("expected error for nonexistent directory")
	}
}

func TestExtractInterfaces_WithInterfaces(t *testing.T) {
	dir := t.TempDir()
	writeAgentFile(t, dir, "backend", agentWithInterfaces)

	ifaces := ExtractInterfaces(filepath.Join(dir, "backend.md"))

	if len(ifaces.Produces) != 2 {
		t.Fatalf("expected 2 produces, got %d", len(ifaces.Produces))
	}
	if ifaces.Produces[0] != "API endpoints" {
		t.Errorf("produces[0] = %q", ifaces.Produces[0])
	}
	if ifaces.Produces[1] != "**/*.go" {
		t.Errorf("produces[1] = %q", ifaces.Produces[1])
	}
	if len(ifaces.Consumes) != 2 {
		t.Fatalf("expected 2 consumes, got %d", len(ifaces.Consumes))
	}
	if ifaces.Consumes[0] != "architecture.yaml" {
		t.Errorf("consumes[0] = %q", ifaces.Consumes[0])
	}
}

func TestExtractInterfaces_WithoutInterfaces(t *testing.T) {
	dir := t.TempDir()
	writeAgentFile(t, dir, "simple", agentWithoutInterfaces)

	ifaces := ExtractInterfaces(filepath.Join(dir, "simple.md"))

	if len(ifaces.Produces) != 0 {
		t.Errorf("expected empty produces, got %v", ifaces.Produces)
	}
	if len(ifaces.Consumes) != 0 {
		t.Errorf("expected empty consumes, got %v", ifaces.Consumes)
	}
	// Verify non-nil
	if ifaces.Produces == nil {
		t.Error("produces should be empty slice, not nil")
	}
	if ifaces.Consumes == nil {
		t.Error("consumes should be empty slice, not nil")
	}
}

func TestExtractInterfaces_PartialInterfaces(t *testing.T) {
	dir := t.TempDir()
	writeAgentFile(t, dir, "partial", agentPartialInterfaces)

	ifaces := ExtractInterfaces(filepath.Join(dir, "partial.md"))

	if len(ifaces.Produces) != 1 || ifaces.Produces[0] != "reports" {
		t.Errorf("produces = %v", ifaces.Produces)
	}
	if len(ifaces.Consumes) != 0 {
		t.Errorf("expected empty consumes, got %v", ifaces.Consumes)
	}
	if ifaces.Consumes == nil {
		t.Error("consumes should be empty slice, not nil")
	}
}

func TestExtractInterfaces_NonexistentFile(t *testing.T) {
	ifaces := ExtractInterfaces("/nonexistent/file.md")

	if ifaces.Produces == nil || ifaces.Consumes == nil {
		t.Error("should return non-nil empty slices for nonexistent file")
	}
	if len(ifaces.Produces) != 0 || len(ifaces.Consumes) != 0 {
		t.Error("should return empty slices for nonexistent file")
	}
}

func TestScanAgents_InlineSkills(t *testing.T) {
	dir := t.TempDir()
	writeAgentFile(t, dir, "inline", agentInlineSkills)

	agents, err := ScanAgents(dir)
	if err != nil {
		t.Fatalf("ScanAgents returned error: %v", err)
	}

	if len(agents) != 1 {
		t.Fatalf("expected 1 agent, got %d", len(agents))
	}

	a := agents[0]
	if len(a.Skills) != 2 || a.Skills[0] != "reader" || a.Skills[1] != "writer" {
		t.Errorf("inline skills = %v", a.Skills)
	}
	if len(a.Interfaces.Produces) != 0 {
		t.Errorf("expected empty produces, got %v", a.Interfaces.Produces)
	}
	if len(a.Interfaces.Consumes) != 1 || a.Interfaces.Consumes[0] != "data files" {
		t.Errorf("consumes = %v", a.Interfaces.Consumes)
	}
}

func TestGenerateYAML_ValidOutput(t *testing.T) {
	agents := []AgentEntry{
		{
			Name:        "backend",
			Description: "Activate for backend API work",
			Model:       "claude-sonnet-4-6",
			Skills:      []string{"code-reviewer"},
			Interfaces: AgentInterfaces{
				Produces: []string{"API endpoints", "**/*.go"},
				Consumes: []string{"architecture.yaml"},
			},
		},
		{
			Name:        "frontend",
			Description: "Activate for UI work",
			Model:       "claude-sonnet-4-6",
			Skills:      []string{},
			Interfaces: AgentInterfaces{
				Produces: []string{},
				Consumes: []string{},
			},
		},
	}

	data := GenerateYAML(agents)
	output := string(data)

	// Verify header
	if !strings.HasPrefix(output, "# Auto-generated by ck agents registry\n") {
		t.Error("missing header comment")
	}

	// Verify structure
	if !strings.Contains(output, "agents:\n") {
		t.Error("missing agents: key")
	}
	if !strings.Contains(output, "  - name: backend\n") {
		t.Error("missing backend entry")
	}
	if !strings.Contains(output, "  - name: frontend\n") {
		t.Error("missing frontend entry")
	}

	// Verify description quoting
	if !strings.Contains(output, `    description: "Activate for backend API work"`) {
		t.Error("description not properly quoted")
	}

	// Verify empty lists use inline []
	if !strings.Contains(output, "    skills: []\n") {
		t.Errorf("empty skills should use []\nGot:\n%s", output)
	}
	if !strings.Contains(output, "      produces: []\n") {
		t.Errorf("empty produces should use []\nGot:\n%s", output)
	}

	// Verify non-empty lists
	if !strings.Contains(output, "    skills:\n") {
		t.Errorf("non-empty skills should be multi-line\nGot:\n%s", output)
	}
	if !strings.Contains(output, `      - "API endpoints"`) {
		t.Errorf("produces items missing\nGot:\n%s", output)
	}
}

func TestGenerateYAML_NoModel(t *testing.T) {
	agents := []AgentEntry{
		{
			Name:        "minimal",
			Description: "Minimal agent",
			Skills:      []string{},
			Interfaces:  AgentInterfaces{Produces: []string{}, Consumes: []string{}},
		},
	}

	output := string(GenerateYAML(agents))

	if strings.Contains(output, "model:") {
		t.Error("model field should be omitted when empty")
	}
}

func TestWriteRegistry_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "output", "agent-registry.yaml")

	agents := []AgentEntry{
		{
			Name:        "test",
			Description: "Test agent",
			Model:       "claude-sonnet-4-6",
			Skills:      []string{"reviewer"},
			Interfaces: AgentInterfaces{
				Produces: []string{"reports"},
				Consumes: []string{},
			},
		},
	}

	err := WriteRegistry(outPath, agents)
	if err != nil {
		t.Fatalf("WriteRegistry returned error: %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	content := string(data)
	if !strings.HasPrefix(content, "# Auto-generated by ck agents registry\n") {
		t.Error("output file missing header")
	}
	if !strings.Contains(content, "  - name: test\n") {
		t.Error("output file missing agent entry")
	}
}

func TestWriteRegistry_EmptyAgents(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "agent-registry.yaml")

	err := WriteRegistry(outPath, []AgentEntry{})
	if err != nil {
		t.Fatalf("WriteRegistry returned error: %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	content := string(data)
	expected := "# Auto-generated by ck agents registry\nagents:\n"
	if content != expected {
		t.Errorf("expected:\n%s\ngot:\n%s", expected, content)
	}
}
