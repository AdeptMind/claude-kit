package claude

import (
	"context"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func mockCmd(stdout string) func(ctx context.Context, name string, args ...string) *exec.Cmd {
	return func(ctx context.Context, name string, args ...string) *exec.Cmd {
		// Use printf to avoid depending on environment variables (buildCmd overwrites Env).
		return exec.CommandContext(ctx, "printf", "%s\n", stdout)
	}
}

func TestRunPrompt_ReturnsStdout(t *testing.T) {
	original := commandBuilder
	defer func() { commandBuilder = original }()

	commandBuilder = mockCmd("hello from claude")

	got, err := RunPrompt(context.Background(), "test prompt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "hello from claude\n"
	if got != expected {
		t.Errorf("got %q, want %q", got, expected)
	}
}

func TestRunQuery_ParsesStreamJSON(t *testing.T) {
	original := commandBuilder
	defer func() { commandBuilder = original }()

	script := `printf '%s\n' '{"type":"assistant","message":{"content":[{"type":"text","text":"parsed ok"}]}}' '{"type":"result","result":{"ok":true}}'`
	commandBuilder = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		return exec.CommandContext(ctx, "sh", "-c", script)
	}

	result, err := RunQuery(context.Background(), "test query")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(result.Events))
	}

	if result.Events[0].Type != "assistant" {
		t.Errorf("expected first event type 'assistant', got %q", result.Events[0].Type)
	}

	if result.TextOut != "parsed ok" {
		t.Errorf("expected TextOut 'parsed ok', got %q", result.TextOut)
	}
}

func TestBuildEnv_StripsClaudeCode(t *testing.T) {
	t.Setenv("CLAUDECODE", "should-be-removed")

	env := buildEnv()
	for _, e := range env {
		if strings.HasPrefix(e, "CLAUDECODE=") {
			t.Fatal("CLAUDECODE should be stripped from environment")
		}
	}
}

func TestRunPrompt_TimeoutKillsProcess(t *testing.T) {
	original := commandBuilder
	defer func() { commandBuilder = original }()

	commandBuilder = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		return exec.CommandContext(ctx, "sleep", "30")
	}

	ctx := context.Background()
	_, err := RunPrompt(ctx, "slow prompt", WithTimeout(50*time.Millisecond))
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}

	if !strings.Contains(err.Error(), "timed out") {
		t.Errorf("expected 'timed out' in error, got: %v", err)
	}
}

func TestRunQuery_TimeoutKillsProcess(t *testing.T) {
	original := commandBuilder
	defer func() { commandBuilder = original }()

	commandBuilder = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		return exec.CommandContext(ctx, "sleep", "30")
	}

	ctx := context.Background()
	_, err := RunQuery(ctx, "slow query", WithTimeout(50*time.Millisecond))
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}

	if !strings.Contains(err.Error(), "timed out") {
		t.Errorf("expected 'timed out' in error, got: %v", err)
	}
}

func TestBuildCmd_SetsModelFlag(t *testing.T) {
	original := commandBuilder
	defer func() { commandBuilder = original }()

	var capturedArgs []string
	commandBuilder = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		capturedArgs = args
		return exec.CommandContext(ctx, "true")
	}

	o := applyOpts([]Option{WithModel("claude-3-opus")})
	buildCmd(context.Background(), "q", "text", o)

	found := false
	for i, a := range capturedArgs {
		if a == "--model" && i+1 < len(capturedArgs) && capturedArgs[i+1] == "claude-3-opus" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected --model claude-3-opus in args, got %v", capturedArgs)
	}
}

func TestBuildCmd_SetsCwd(t *testing.T) {
	original := commandBuilder
	defer func() { commandBuilder = original }()

	commandBuilder = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		return exec.CommandContext(ctx, "true")
	}

	o := applyOpts([]Option{WithCwd("/tmp/project")})
	cmd := buildCmd(context.Background(), "q", "text", o)

	if cmd.Dir != "/tmp/project" {
		t.Errorf("expected Dir '/tmp/project', got %q", cmd.Dir)
	}
}

func TestBuildCmd_StreamJSON_IncludesVerbose(t *testing.T) {
	original := commandBuilder
	defer func() { commandBuilder = original }()

	var capturedArgs []string
	commandBuilder = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		capturedArgs = args
		return exec.CommandContext(ctx, "true")
	}

	o := applyOpts(nil)
	buildCmd(context.Background(), "q", "stream-json", o)

	found := false
	for _, a := range capturedArgs {
		if a == "--verbose" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected --verbose in stream-json args, got %v", capturedArgs)
	}
}
