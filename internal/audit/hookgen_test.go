package audit

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateHookCommand_valid_shell(t *testing.T) {
	cmd := GenerateHookCommand("/tmp/audit.jsonl")
	if cmd == "" {
		t.Fatal("expected non-empty command")
	}
	// Verify it parses as valid shell
	err := exec.Command("bash", "-n", "-c", cmd).Run()
	if err != nil {
		t.Fatalf("generated command is not valid shell: %v", err)
	}
}

func TestGenerateHookCommand_writes_valid_jsonl(t *testing.T) {
	dir := t.TempDir()
	auditPath := filepath.Join(dir, "audit.jsonl")

	cmd := GenerateHookCommand(auditPath)

	// Set env vars and execute the command
	c := exec.Command("bash", "-c", cmd)
	c.Env = append(os.Environ(),
		"CLAUDE_TOOL_NAME=Bash",
		"CLAUDE_TOOL_INPUT=echo hello world",
		"CLAUDE_SESSION=test-session-123",
	)
	out, err := c.CombinedOutput()
	if err != nil {
		t.Fatalf("command failed: %v\noutput: %s", err, out)
	}

	data, err := os.ReadFile(auditPath)
	if err != nil {
		t.Fatalf("failed to read audit file: %v", err)
	}

	line := strings.TrimSpace(string(data))
	var event AuditEvent
	if err := json.Unmarshal([]byte(line), &event); err != nil {
		t.Fatalf("invalid JSON line: %v\nline: %s", err, line)
	}

	if event.Tool != "Bash" {
		t.Errorf("tool = %q, want %q", event.Tool, "Bash")
	}
	if event.InputSummary != "echo hello world" {
		t.Errorf("input_summary = %q, want %q", event.InputSummary, "echo hello world")
	}
	if event.Status != "ok" {
		t.Errorf("status = %q, want %q", event.Status, "ok")
	}
	if event.DurationMs != 0 {
		t.Errorf("duration_ms = %d, want 0", event.DurationMs)
	}
	if event.SessionID != "test-session-123" {
		t.Errorf("session_id = %q, want %q", event.SessionID, "test-session-123")
	}
	if event.Timestamp == "" {
		t.Error("timestamp should not be empty")
	}
}

func TestGenerateHookCommand_missing_env_vars(t *testing.T) {
	dir := t.TempDir()
	auditPath := filepath.Join(dir, "audit.jsonl")

	cmd := GenerateHookCommand(auditPath)

	// Execute without CLAUDE_TOOL_NAME, CLAUDE_TOOL_INPUT, CLAUDE_SESSION
	c := exec.Command("bash", "-c", cmd)
	c.Env = []string{"PATH=" + os.Getenv("PATH")}
	out, err := c.CombinedOutput()
	if err != nil {
		t.Fatalf("command should not crash with missing env vars: %v\noutput: %s", err, out)
	}

	data, err := os.ReadFile(auditPath)
	if err != nil {
		t.Fatalf("failed to read audit file: %v", err)
	}

	line := strings.TrimSpace(string(data))
	var event AuditEvent
	if err := json.Unmarshal([]byte(line), &event); err != nil {
		t.Fatalf("invalid JSON line: %v\nline: %s", err, line)
	}

	if event.Tool != "unknown" {
		t.Errorf("tool = %q, want %q for missing env", event.Tool, "unknown")
	}
	if event.InputSummary != "unknown" {
		t.Errorf("input_summary = %q, want %q for missing env", event.InputSummary, "unknown")
	}
	if event.SessionID != "none" {
		t.Errorf("session_id = %q, want %q for missing env", event.SessionID, "none")
	}
}

func TestGenerateHookCommand_truncates_long_input(t *testing.T) {
	dir := t.TempDir()
	auditPath := filepath.Join(dir, "audit.jsonl")

	cmd := GenerateHookCommand(auditPath)

	longInput := strings.Repeat("x", 300)
	c := exec.Command("bash", "-c", cmd)
	c.Env = append(os.Environ(),
		"CLAUDE_TOOL_NAME=Read",
		"CLAUDE_TOOL_INPUT="+longInput,
	)
	out, err := c.CombinedOutput()
	if err != nil {
		t.Fatalf("command failed: %v\noutput: %s", err, out)
	}

	data, err := os.ReadFile(auditPath)
	if err != nil {
		t.Fatalf("failed to read audit file: %v", err)
	}

	line := strings.TrimSpace(string(data))
	var event AuditEvent
	if err := json.Unmarshal([]byte(line), &event); err != nil {
		t.Fatalf("invalid JSON line: %v\nline: %s", err, line)
	}

	if len(event.InputSummary) > maxInputLen {
		t.Errorf("input_summary length = %d, want <= %d", len(event.InputSummary), maxInputLen)
	}
}

func TestGenerateHookCommand_escapes_quotes_in_input(t *testing.T) {
	dir := t.TempDir()
	auditPath := filepath.Join(dir, "audit.jsonl")

	cmd := GenerateHookCommand(auditPath)

	c := exec.Command("bash", "-c", cmd)
	c.Env = append(os.Environ(),
		"CLAUDE_TOOL_NAME=Bash",
		`CLAUDE_TOOL_INPUT=echo "hello" world`,
	)
	out, err := c.CombinedOutput()
	if err != nil {
		t.Fatalf("command failed: %v\noutput: %s", err, out)
	}

	data, err := os.ReadFile(auditPath)
	if err != nil {
		t.Fatalf("failed to read audit file: %v", err)
	}

	line := strings.TrimSpace(string(data))
	var event AuditEvent
	if err := json.Unmarshal([]byte(line), &event); err != nil {
		t.Fatalf("invalid JSON line (quotes not escaped properly): %v\nline: %s", err, line)
	}

	// After JSON deserialization, escaped quotes become regular quotes
	if event.InputSummary != `echo "hello" world` {
		t.Errorf("input_summary = %q, want %q", event.InputSummary, `echo "hello" world`)
	}
}

func TestGenerateHookEntry_structure(t *testing.T) {
	entry := GenerateHookEntry("/tmp/audit.jsonl")

	typ, ok := entry["type"].(string)
	if !ok || typ != "command" {
		t.Errorf("type = %v, want %q", entry["type"], "command")
	}

	cmd, ok := entry["command"].(string)
	if !ok || cmd == "" {
		t.Error("command should be a non-empty string")
	}

	if !strings.Contains(cmd, "/tmp/audit.jsonl") {
		t.Error("command should contain the audit path")
	}
}
