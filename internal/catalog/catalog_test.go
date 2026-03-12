package catalog

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// writeTestFile creates a .md file with the given content in a temp directory.
func writeTestFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestExtractInterfaces_BothProducesAndConsumes(t *testing.T) {
	dir := t.TempDir()
	path := writeTestFile(t, dir, "backend.md", `---
name: backend
description: Backend agent
interfaces:
  produces:
    - "API endpoints"
    - "**/*.go"
  consumes:
    - "architecture.yaml"
    - "API contracts"
---
# Backend Agent
`)

	got := ExtractInterfaces(path)
	want := AgentInterfaces{
		Produces: []string{"API endpoints", "**/*.go"},
		Consumes: []string{"architecture.yaml", "API contracts"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ExtractInterfaces() = %+v, want %+v", got, want)
	}
}

func TestExtractInterfaces_OnlyProduces(t *testing.T) {
	dir := t.TempDir()
	path := writeTestFile(t, dir, "frontend.md", `---
name: frontend
description: Frontend agent
interfaces:
  produces:
    - "React components"
---
# Frontend
`)

	got := ExtractInterfaces(path)
	want := AgentInterfaces{
		Produces: []string{"React components"},
		Consumes: []string{},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ExtractInterfaces() = %+v, want %+v", got, want)
	}
}

func TestExtractInterfaces_NoInterfaces(t *testing.T) {
	dir := t.TempDir()
	path := writeTestFile(t, dir, "simple.md", `---
name: simple
description: No interfaces here
---
# Simple
`)

	got := ExtractInterfaces(path)
	if len(got.Produces) != 0 {
		t.Errorf("expected empty Produces, got %v", got.Produces)
	}
	if len(got.Consumes) != 0 {
		t.Errorf("expected empty Consumes, got %v", got.Consumes)
	}
}

func TestExtractInterfaces_EmptyInterfaces(t *testing.T) {
	dir := t.TempDir()
	path := writeTestFile(t, dir, "empty.md", `---
name: empty
description: Empty interfaces
interfaces:
skills:
  - code-reviewer
---
# Empty
`)

	got := ExtractInterfaces(path)
	if len(got.Produces) != 0 {
		t.Errorf("expected empty Produces, got %v", got.Produces)
	}
	if len(got.Consumes) != 0 {
		t.Errorf("expected empty Consumes, got %v", got.Consumes)
	}
}

func TestExtractInterfaces_QuotedValues(t *testing.T) {
	dir := t.TempDir()
	path := writeTestFile(t, dir, "quoted.md", `---
name: quoted
description: Quoted values
interfaces:
  produces:
    - "double quoted"
    - 'single quoted'
    - unquoted value
---
`)

	got := ExtractInterfaces(path)
	want := AgentInterfaces{
		Produces: []string{"double quoted", "single quoted", "unquoted value"},
		Consumes: []string{},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ExtractInterfaces() = %+v, want %+v", got, want)
	}
}

func TestExtractInterfaces_FileNotFound(t *testing.T) {
	got := ExtractInterfaces("/nonexistent/file.md")
	if len(got.Produces) != 0 || len(got.Consumes) != 0 {
		t.Errorf("expected empty AgentInterfaces for missing file, got %+v", got)
	}
}

func TestComponent_HasInterfaces(t *testing.T) {
	dir := t.TempDir()
	agentsDir := filepath.Join(dir, "agents")
	if err := os.MkdirAll(agentsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, dir, "agents/backend.md", `---
name: backend
description: Backend API agent
interfaces:
  produces:
    - "API endpoints"
  consumes:
    - "architecture.yaml"
---
# Backend
`)

	components := scanMarkdownDir(agentsDir, "agents")
	if len(components) != 1 {
		t.Fatalf("expected 1 component, got %d", len(components))
	}
	comp := components[0]
	if len(comp.Interfaces.Produces) != 1 || comp.Interfaces.Produces[0] != "API endpoints" {
		t.Errorf("unexpected Produces: %v", comp.Interfaces.Produces)
	}
	if len(comp.Interfaces.Consumes) != 1 || comp.Interfaces.Consumes[0] != "architecture.yaml" {
		t.Errorf("unexpected Consumes: %v", comp.Interfaces.Consumes)
	}
}

func TestExtractDescription(t *testing.T) {
	dir := t.TempDir()
	path := writeTestFile(t, dir, "agent.md", `---
name: test
description: This is a test agent
---
# Test
`)

	got := ExtractDescription(path)
	if got != "This is a test agent" {
		t.Errorf("ExtractDescription() = %q, want %q", got, "This is a test agent")
	}
}

func TestExtractDescription_Quoted(t *testing.T) {
	dir := t.TempDir()
	path := writeTestFile(t, dir, "agent.md", `---
name: test
description: "Quoted description"
---
`)

	got := ExtractDescription(path)
	if got != "Quoted description" {
		t.Errorf("ExtractDescription() = %q, want %q", got, "Quoted description")
	}
}

func TestExtractDescription_Missing(t *testing.T) {
	dir := t.TempDir()
	path := writeTestFile(t, dir, "agent.md", `---
name: test
---
`)

	got := ExtractDescription(path)
	if got != "" {
		t.Errorf("ExtractDescription() = %q, want empty string", got)
	}
}

func TestExtractVersion(t *testing.T) {
	dir := t.TempDir()
	path := writeTestFile(t, dir, "agent.md", `---
name: test
version: 1.2.3
---
`)

	got := ExtractVersion(path)
	if got != "1.2.3" {
		t.Errorf("ExtractVersion() = %q, want %q", got, "1.2.3")
	}
}

func TestExtractVersion_Missing(t *testing.T) {
	dir := t.TempDir()
	path := writeTestFile(t, dir, "agent.md", `---
name: test
---
`)

	got := ExtractVersion(path)
	if got != "" {
		t.Errorf("ExtractVersion() = %q, want empty string", got)
	}
}

func TestScanMarkdownDir_NonAgentNoInterfaces(t *testing.T) {
	dir := t.TempDir()
	rulesDir := filepath.Join(dir, "rules")
	if err := os.MkdirAll(rulesDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, dir, "rules/security.md", `---
name: security
description: Security rules
---
# Security
`)

	components := scanMarkdownDir(rulesDir, "rules")
	if len(components) != 1 {
		t.Fatalf("expected 1 component, got %d", len(components))
	}
	// Rules should have zero-value interfaces (not populated)
	comp := components[0]
	if len(comp.Interfaces.Produces) != 0 || len(comp.Interfaces.Consumes) != 0 {
		t.Errorf("rules should not have interfaces populated, got %+v", comp.Interfaces)
	}
}
