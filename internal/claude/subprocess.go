package claude

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// ContentBlock represents a single content block in an assistant message.
type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
	Name string `json:"name,omitempty"`
	ID   string `json:"id,omitempty"`
}

// AssistantMessage is the message payload inside a stream event of type "assistant".
type AssistantMessage struct {
	Content []ContentBlock `json:"content"`
}

// StreamEvent represents a single JSON line from claude's stream-json output.
type StreamEvent struct {
	Type    string           `json:"type"`
	Event   json.RawMessage  `json:"event,omitempty"`
	Message *AssistantMessage `json:"message,omitempty"`
	Result  json.RawMessage  `json:"result,omitempty"`
}

// QueryResult holds the parsed output from a stream-json claude invocation.
type QueryResult struct {
	Events  []StreamEvent
	TextOut string // concatenated text from assistant content blocks
}

// options holds the configurable parameters for a claude invocation.
type options struct {
	model   string
	cwd     string
	timeout time.Duration
}

// Option configures a claude subprocess invocation.
type Option func(*options)

// WithModel sets the --model flag.
func WithModel(m string) Option {
	return func(o *options) { o.model = m }
}

// WithCwd sets the working directory for the subprocess.
func WithCwd(dir string) Option {
	return func(o *options) { o.cwd = dir }
}

// WithTimeout sets the maximum execution time before the process is killed.
func WithTimeout(d time.Duration) Option {
	return func(o *options) { o.timeout = d }
}

// commandBuilder is the function used to create exec.Cmd. Overridable for tests.
var commandBuilder = defaultCommandBuilder

func defaultCommandBuilder(ctx context.Context, name string, args ...string) *exec.Cmd {
	return exec.CommandContext(ctx, name, args...)
}

// RunQuery spawns `claude -p` with --output-format stream-json, parses JSON
// lines, and returns the collected events plus concatenated text output.
func RunQuery(ctx context.Context, query string, opts ...Option) (*QueryResult, error) {
	o := applyOpts(opts)

	if o.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, o.timeout)
		defer cancel()
	}

	cmd := buildCmd(ctx, query, "stream-json", o)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("claude query timed out: %w", ctx.Err())
		}
		return nil, fmt.Errorf("claude query failed: %w", err)
	}

	result := &QueryResult{}
	scanner := bufio.NewScanner(&stdout)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var ev StreamEvent
		if err := json.Unmarshal([]byte(line), &ev); err != nil {
			continue // skip malformed lines
		}
		result.Events = append(result.Events, ev)

		if ev.Type == "assistant" && ev.Message != nil {
			for _, block := range ev.Message.Content {
				if block.Type == "text" {
					result.TextOut += block.Text
				}
			}
		}
	}

	return result, nil
}

// RunPrompt spawns `claude -p` with --output-format text and returns stdout.
func RunPrompt(ctx context.Context, prompt string, opts ...Option) (string, error) {
	o := applyOpts(opts)

	if o.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, o.timeout)
		defer cancel()
	}

	cmd := buildCmd(ctx, prompt, "text", o)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("claude prompt timed out: %w", ctx.Err())
		}
		return "", fmt.Errorf("claude prompt failed: %w", err)
	}

	return stdout.String(), nil
}

func applyOpts(opts []Option) options {
	var o options
	for _, fn := range opts {
		fn(&o)
	}
	return o
}

func buildCmd(ctx context.Context, query, format string, o options) *exec.Cmd {
	args := []string{"-p", query, "--output-format", format}
	if format == "stream-json" {
		args = append(args, "--verbose")
	}
	if o.model != "" {
		args = append(args, "--model", o.model)
	}

	cmd := commandBuilder(ctx, "claude", args...)
	cmd.Env = buildEnv()
	if o.cwd != "" {
		cmd.Dir = o.cwd
	}
	return cmd
}

// buildEnv returns a copy of the current environment with CLAUDECODE stripped.
func buildEnv() []string {
	var env []string
	for _, e := range os.Environ() {
		if strings.HasPrefix(e, "CLAUDECODE=") {
			continue
		}
		env = append(env, e)
	}
	return env
}
