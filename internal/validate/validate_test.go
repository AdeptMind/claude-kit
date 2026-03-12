package validate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateSkill_Valid(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "my-skill")
	os.MkdirAll(skillDir, 0o755)

	writeFile(t, filepath.Join(skillDir, "SKILL.md"), `---
name: my-skill
description: A valid skill for testing
---
# My Skill
Content here.
`)

	errs := ValidateSkill(skillDir)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %d: %v", len(errs), errs)
	}
}

func TestValidateSkill_MissingSkillMd(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "empty-skill")
	os.MkdirAll(skillDir, 0o755)

	errs := ValidateSkill(skillDir)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	if !strings.Contains(errs[0].Message, "SKILL.md not found") {
		t.Errorf("unexpected error message: %s", errs[0].Message)
	}
}

func TestValidateSkill_MissingName(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "no-name")
	os.MkdirAll(skillDir, 0o755)

	writeFile(t, filepath.Join(skillDir, "SKILL.md"), `---
description: Has description but no name
---
Content.
`)

	errs := ValidateSkill(skillDir)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	if errs[0].Field != "name" {
		t.Errorf("expected field 'name', got %q", errs[0].Field)
	}
	if !strings.Contains(errs[0].Message, "missing or empty") {
		t.Errorf("unexpected message: %s", errs[0].Message)
	}
	// Line should be 0 since the field is absent
	if errs[0].Line != 0 {
		t.Errorf("expected line 0 for missing field, got %d", errs[0].Line)
	}
}

func TestValidateSkill_MissingDescription(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "no-desc")
	os.MkdirAll(skillDir, 0o755)

	writeFile(t, filepath.Join(skillDir, "SKILL.md"), `---
name: no-desc
---
Content.
`)

	errs := ValidateSkill(skillDir)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	if errs[0].Field != "description" {
		t.Errorf("expected field 'description', got %q", errs[0].Field)
	}
}

func TestValidateSkill_InvalidEvalsJson(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "bad-evals")
	os.MkdirAll(skillDir, 0o755)

	writeFile(t, filepath.Join(skillDir, "SKILL.md"), `---
name: bad-evals
description: Skill with broken evals
---
`)
	writeFile(t, filepath.Join(skillDir, "evals.json"), `{not valid json`)

	errs := ValidateSkill(skillDir)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	if errs[0].Field != "evals.json" {
		t.Errorf("expected field 'evals.json', got %q", errs[0].Field)
	}
	if !strings.Contains(errs[0].Message, "invalid JSON") {
		t.Errorf("unexpected message: %s", errs[0].Message)
	}
}

func TestValidateSkill_ValidEvalsJson(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "good-evals")
	os.MkdirAll(skillDir, 0o755)

	writeFile(t, filepath.Join(skillDir, "SKILL.md"), `---
name: good-evals
description: Skill with valid evals
---
`)
	writeFile(t, filepath.Join(skillDir, "evals.json"), `[{"name": "test"}]`)

	errs := ValidateSkill(skillDir)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %d: %v", len(errs), errs)
	}
}

func TestValidateAgent_Valid(t *testing.T) {
	dir := t.TempDir()
	skillsDir := filepath.Join(dir, "skills")
	createSkillDir(t, skillsDir, "code-reviewer")
	createSkillDir(t, skillsDir, "test-generator")

	agentPath := filepath.Join(dir, "backend.md")
	writeFile(t, agentPath, `---
name: backend
description: Activate for backend API tasks
model: claude-sonnet-4-6
version: "1.0.0"
tools: [Read, Write, Edit, Bash, Grep, Glob]
skills:
  - code-reviewer
  - test-generator
---
# Backend Agent
`)

	errs := ValidateAgent(agentPath, skillsDir)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %d: %v", len(errs), errs)
	}
}

func TestValidateAgent_MissingDescription(t *testing.T) {
	dir := t.TempDir()
	agentPath := filepath.Join(dir, "nodesc.md")
	writeFile(t, agentPath, `---
name: nodesc
---
`)

	errs := ValidateAgent(agentPath, filepath.Join(dir, "skills"))
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	if errs[0].Field != "description" {
		t.Errorf("expected field 'description', got %q", errs[0].Field)
	}
	if errs[0].Line != 0 {
		t.Errorf("expected line 0 for missing field, got %d", errs[0].Line)
	}
}

func TestValidateAgent_MissingName(t *testing.T) {
	dir := t.TempDir()
	agentPath := filepath.Join(dir, "noname.md")
	writeFile(t, agentPath, `---
description: Has description but no name
---
`)

	errs := ValidateAgent(agentPath, filepath.Join(dir, "skills"))
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	if errs[0].Field != "name" {
		t.Errorf("expected field 'name', got %q", errs[0].Field)
	}
}

func TestValidateAgent_BrokenSkillRef(t *testing.T) {
	dir := t.TempDir()
	skillsDir := filepath.Join(dir, "skills")
	os.MkdirAll(skillsDir, 0o755)

	agentPath := filepath.Join(dir, "broken.md")
	writeFile(t, agentPath, `---
name: broken
description: Agent with broken skill ref
skills:
  - nonexistent-skill
---
`)

	errs := ValidateAgent(agentPath, skillsDir)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	if errs[0].Field != "skills" {
		t.Errorf("expected field 'skills', got %q", errs[0].Field)
	}
	if !strings.Contains(errs[0].Message, "nonexistent-skill") {
		t.Errorf("error should mention the skill name: %s", errs[0].Message)
	}
	if !strings.Contains(errs[0].Message, "not found") {
		t.Errorf("error should say 'not found': %s", errs[0].Message)
	}
}

func TestValidateAgent_InvalidTool(t *testing.T) {
	dir := t.TempDir()
	agentPath := filepath.Join(dir, "badtool.md")
	writeFile(t, agentPath, `---
name: badtool
description: Agent with invalid tool
tools: [Read, FakeTool, Write]
---
`)

	errs := ValidateAgent(agentPath, filepath.Join(dir, "skills"))
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	if errs[0].Field != "tools" {
		t.Errorf("expected field 'tools', got %q", errs[0].Field)
	}
	if !strings.Contains(errs[0].Message, "FakeTool") {
		t.Errorf("error should mention the invalid tool: %s", errs[0].Message)
	}
	if errs[0].Line == 0 {
		t.Errorf("expected a non-zero line number for tools field")
	}
}

func TestValidateAgent_NotMdExtension(t *testing.T) {
	dir := t.TempDir()
	agentPath := filepath.Join(dir, "agent.yaml")
	writeFile(t, agentPath, `---
name: agent
description: Wrong extension
---
`)

	errs := ValidateAgent(agentPath, filepath.Join(dir, "skills"))
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	if !strings.Contains(errs[0].Message, ".md extension") {
		t.Errorf("error should mention .md extension: %s", errs[0].Message)
	}
}

func TestValidateAgent_FileNotFound(t *testing.T) {
	errs := ValidateAgent("/nonexistent/path/agent.md", "/tmp/skills")
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	if !strings.Contains(errs[0].Message, "not found") {
		t.Errorf("error should mention 'not found': %s", errs[0].Message)
	}
}

func TestValidateAgent_MultipleErrors(t *testing.T) {
	dir := t.TempDir()
	skillsDir := filepath.Join(dir, "skills")
	os.MkdirAll(skillsDir, 0o755)

	agentPath := filepath.Join(dir, "multi.md")
	writeFile(t, agentPath, `---
name: multi
description: Agent with multiple issues
tools: [Read, BadTool1, BadTool2]
skills:
  - ghost-skill
---
`)

	errs := ValidateAgent(agentPath, skillsDir)
	if len(errs) != 3 {
		t.Fatalf("expected 3 errors (2 bad tools + 1 missing skill), got %d: %v", len(errs), errs)
	}
}

func TestValidationError_ErrorFormat(t *testing.T) {
	withLine := ValidationError{
		File:    "agents/backend.md",
		Line:    5,
		Field:   "description",
		Message: "required field is missing or empty",
	}
	want := "agents/backend.md:5: description: required field is missing or empty"
	if got := withLine.Error(); got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}

	withoutLine := ValidationError{
		File:    "skills/broken",
		Field:   "SKILL.md",
		Message: "required file SKILL.md not found",
	}
	want = "skills/broken: SKILL.md: required file SKILL.md not found"
	if got := withoutLine.Error(); got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestValidateAgent_InlineSkillsList(t *testing.T) {
	dir := t.TempDir()
	skillsDir := filepath.Join(dir, "skills")
	createSkillDir(t, skillsDir, "code-reviewer")

	agentPath := filepath.Join(dir, "inline.md")
	writeFile(t, agentPath, `---
name: inline
description: Agent with inline skills
skills: [code-reviewer]
---
`)

	errs := ValidateAgent(agentPath, skillsDir)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %d: %v", len(errs), errs)
	}
}

// --- helpers ---

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func createSkillDir(t *testing.T, skillsDir, name string) {
	t.Helper()
	dir := filepath.Join(skillsDir, name)
	os.MkdirAll(dir, 0o755)
	writeFile(t, filepath.Join(dir, "SKILL.md"), `---
name: `+name+`
description: Test skill `+name+`
---
`)
}
