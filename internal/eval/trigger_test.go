package eval

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AdeptMind/infra-tool/claude-cli/internal/claude"
)

func TestCommandFileContent_HasFrontmatter(t *testing.T) {
	content := commandFileContent("A test skill for reviews")

	if !strings.Contains(content, "---") {
		t.Error("expected YAML frontmatter delimiters")
	}
	if !strings.Contains(content, "description: A test skill for reviews") {
		t.Error("expected description in frontmatter")
	}
}

func TestCreateCommandFile_CreatesAndWritesFile(t *testing.T) {
	dir := t.TempDir()
	path, err := createCommandFile(dir, "review-skill-123", "Review pull requests")
	if err != nil {
		t.Fatalf("createCommandFile() error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading command file: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "description: Review pull requests") {
		t.Errorf("command file missing description, got:\n%s", content)
	}

	if !strings.HasSuffix(path, "review-skill-123.md") {
		t.Errorf("unexpected file path: %s", path)
	}
}

func TestCreateCommandFile_CreatesDirectoryIfMissing(t *testing.T) {
	dir := t.TempDir()
	// .claude/commands/ does not exist yet
	path, err := createCommandFile(dir, "test-skill-456", "test")
	if err != nil {
		t.Fatalf("createCommandFile() error: %v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("command file was not created")
	}

	cmdDir := filepath.Join(dir, ".claude", "commands")
	if _, err := os.Stat(cmdDir); os.IsNotExist(err) {
		t.Error("commands directory was not created")
	}
}

func TestGenerateCommandName_ContainsSkillName(t *testing.T) {
	name := generateCommandName("code-review")
	if !strings.HasPrefix(name, "code-review-skill-") {
		t.Errorf("expected prefix 'code-review-skill-', got %q", name)
	}
}

func TestGenerateCommandName_IsUnique(t *testing.T) {
	a := generateCommandName("test")
	b := generateCommandName("test")
	if a == b {
		t.Error("two consecutive generateCommandName calls returned the same value")
	}
}

// --- detectTrigger tests using io.Reader ---

func TestDetectTrigger_ContentBlockStartWithSkill(t *testing.T) {
	stream := `{"type":"event","event":{"type":"content_block_start","content_block":{"type":"tool_use","name":"Skill","input":"{\"skill\":\"review-skill-abc123\"}"}}}
{"type":"event","event":{"type":"content_block_stop"}}
`
	triggered := detectTrigger(strings.NewReader(stream), "review-skill-abc123")
	if !triggered {
		t.Error("expected trigger detection for Skill tool_use with matching command name")
	}
}

func TestDetectTrigger_ContentBlockDeltaAccumulation(t *testing.T) {
	stream := `{"type":"event","event":{"type":"content_block_start","content_block":{"type":"tool_use","name":"Skill"}}}
{"type":"event","event":{"type":"content_block_delta","delta":{"partial_json":"{\"skill\":\"review"}}}
{"type":"event","event":{"type":"content_block_delta","delta":{"partial_json":"-skill-xyz789\"}"}}}
{"type":"event","event":{"type":"content_block_stop"}}
`
	triggered := detectTrigger(strings.NewReader(stream), "review-skill-xyz789")
	if !triggered {
		t.Error("expected trigger detection from accumulated partial JSON deltas")
	}
}

func TestDetectTrigger_ReadToolTriggersDetection(t *testing.T) {
	stream := `{"type":"event","event":{"type":"content_block_start","content_block":{"type":"tool_use","name":"Read","input":"{\"path\":\"test-skill-abc/.claude/commands/myskill-skill-999.md\"}"}}}
{"type":"event","event":{"type":"content_block_stop"}}
`
	triggered := detectTrigger(strings.NewReader(stream), "myskill-skill-999")
	if !triggered {
		t.Error("expected trigger detection for Read tool referencing command name")
	}
}

func TestDetectTrigger_NoTriggerForUnrelatedToolUse(t *testing.T) {
	stream := `{"type":"event","event":{"type":"content_block_start","content_block":{"type":"tool_use","name":"Bash","input":"{\"command\":\"ls\"}"}}}
{"type":"event","event":{"type":"content_block_stop"}}
`
	triggered := detectTrigger(strings.NewReader(stream), "myskill-skill-123")
	if triggered {
		t.Error("should not detect trigger for non-Skill/Read tool_use")
	}
}

func TestDetectTrigger_NoTriggerForWrongCommandName(t *testing.T) {
	stream := `{"type":"event","event":{"type":"content_block_start","content_block":{"type":"tool_use","name":"Skill","input":"{\"skill\":\"other-skill-456\"}"}}}
{"type":"event","event":{"type":"content_block_stop"}}
`
	triggered := detectTrigger(strings.NewReader(stream), "myskill-skill-123")
	if triggered {
		t.Error("should not detect trigger when command name does not match")
	}
}

func TestDetectTrigger_AssistantMessageFallback(t *testing.T) {
	stream := `{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Skill","input":"{\"skill\":\"review-skill-fallback\"}"}]}}
`
	triggered := detectTrigger(strings.NewReader(stream), "review-skill-fallback")
	if !triggered {
		t.Error("expected trigger detection from assistant message fallback")
	}
}

func TestDetectTrigger_ResultFallback(t *testing.T) {
	stream := `{"type":"result","result":"{\"tool_use\":{\"name\":\"Skill\",\"input\":{\"skill\":\"review-skill-result\"}}}"}
`
	triggered := detectTrigger(strings.NewReader(stream), "review-skill-result")
	if !triggered {
		t.Error("expected trigger detection from result payload")
	}
}

func TestDetectTrigger_EmptyStream(t *testing.T) {
	triggered := detectTrigger(strings.NewReader(""), "myskill-skill-123")
	if triggered {
		t.Error("should not detect trigger on empty stream")
	}
}

func TestDetectTrigger_MalformedJSON(t *testing.T) {
	stream := `not valid json
{"type":"event","event":{"type":"content_block_start","content_block":{"type":"tool_use","name":"Skill","input":"{\"skill\":\"ok-skill-1\"}"}}}
also bad {{{
`
	triggered := detectTrigger(strings.NewReader(stream), "ok-skill-1")
	if !triggered {
		t.Error("should skip malformed lines and still detect trigger")
	}
}

// --- RunSingleQuery integration tests ---

func TestRunSingleQuery_CreatesAndCleansUpCommandFile(t *testing.T) {
	// Override commandBuilder to avoid real subprocess
	original := claude.ExportCommandBuilder()
	defer claude.SetCommandBuilder(original)

	claude.SetCommandBuilder(func(ctx context.Context, name string, args ...string) *exec.Cmd {
		return exec.CommandContext(ctx, "printf", "%s\n", `{"type":"result","result":"{}"}`)
	})

	dir := t.TempDir()
	_, err := RunSingleQuery(context.Background(), "test query", "cleanup-test", "test description", dir)
	if err != nil {
		t.Fatalf("RunSingleQuery() error: %v", err)
	}

	// Command file should have been cleaned up
	cmdDir := filepath.Join(dir, ".claude", "commands")
	entries, err := os.ReadDir(cmdDir)
	if err != nil {
		t.Fatalf("reading commands dir: %v", err)
	}
	for _, e := range entries {
		if strings.Contains(e.Name(), "cleanup-test-skill-") {
			t.Errorf("command file was not cleaned up: %s", e.Name())
		}
	}
}

func TestRunSingleQuery_CleansUpOnError(t *testing.T) {
	original := claude.ExportCommandBuilder()
	defer claude.SetCommandBuilder(original)

	claude.SetCommandBuilder(func(ctx context.Context, name string, args ...string) *exec.Cmd {
		return exec.CommandContext(ctx, "false") // exit code 1
	})

	dir := t.TempDir()
	_, err := RunSingleQuery(context.Background(), "fail query", "error-test", "test description", dir)
	if err == nil {
		t.Fatal("expected error from failing subprocess")
	}

	// Command file should still be cleaned up
	cmdDir := filepath.Join(dir, ".claude", "commands")
	entries, err := os.ReadDir(cmdDir)
	if err != nil {
		t.Fatalf("reading commands dir: %v", err)
	}
	for _, e := range entries {
		if strings.Contains(e.Name(), "error-test-skill-") {
			t.Errorf("command file was not cleaned up after error: %s", e.Name())
		}
	}
}

func TestRunSingleQuery_DetectsTrigger(t *testing.T) {
	original := claude.ExportCommandBuilder()
	defer claude.SetCommandBuilder(original)

	// We need to craft the stream so that the command name appears in events.
	// Since RunSingleQuery generates the name dynamically, we verify that the
	// function returns false for a generic stream without any tool_use.
	claude.SetCommandBuilder(func(ctx context.Context, name string, args ...string) *exec.Cmd {
		// Stream with no tool_use → should not trigger
		return exec.CommandContext(ctx, "printf", "%s\n", `{"type":"assistant","message":{"content":[{"type":"text","text":"hello"}]}}`)
	})

	dir := t.TempDir()
	triggered, err := RunSingleQuery(context.Background(), "hello", "notrigger", "desc", dir)
	if err != nil {
		t.Fatalf("RunSingleQuery() error: %v", err)
	}
	if triggered {
		t.Error("expected no trigger for stream without tool_use")
	}
}
