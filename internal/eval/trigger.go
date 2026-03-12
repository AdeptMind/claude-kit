package eval

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/AdeptMind/infra-tool/claude-cli/internal/claude"
)

// commandFileContent builds the markdown content for a temporary skill command file.
func commandFileContent(description string) string {
	return fmt.Sprintf("---\ndescription: %s\n---\n\nFollow the skill instructions.\n", description)
}

// generateCommandName returns a unique command name like "myskill-skill-a1b2c3d4".
func generateCommandName(skillName string) string {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		// Fallback: use zero bytes (should never happen)
		b = []byte{0, 0, 0, 0}
	}
	return fmt.Sprintf("%s-skill-%s", skillName, hex.EncodeToString(b))
}

// createCommandFile writes a temporary command file under projectRoot/.claude/commands/
// and returns the full path. The caller is responsible for cleanup.
func createCommandFile(projectRoot, commandName, description string) (string, error) {
	dir := filepath.Join(projectRoot, ".claude", "commands")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("creating commands dir: %w", err)
	}

	path := filepath.Join(dir, commandName+".md")
	content := commandFileContent(description)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return "", fmt.Errorf("writing command file: %w", err)
	}
	return path, nil
}

// streamEvent mirrors the subset of claude stream-json fields needed for trigger detection.
type streamEvent struct {
	Type    string          `json:"type"`
	Event   *eventPayload   `json:"event,omitempty"`
	Message *messagePayload `json:"message,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
}

type eventPayload struct {
	Type         string `json:"type"`
	ContentBlock *struct {
		Type  string `json:"type"`
		Name  string `json:"name,omitempty"`
		Input json.RawMessage `json:"input,omitempty"`
	} `json:"content_block,omitempty"`
	Delta *struct {
		Type           string `json:"type,omitempty"`
		PartialJSON    string `json:"partial_json,omitempty"`
		InputJSONDelta json.RawMessage `json:"input_json_delta,omitempty"`
	} `json:"delta,omitempty"`
}

type messagePayload struct {
	Content []contentBlock `json:"content"`
}

type contentBlock struct {
	Type  string          `json:"type"`
	Name  string          `json:"name,omitempty"`
	Input json.RawMessage `json:"input,omitempty"`
}

// detectTrigger reads stream-json lines from r and returns true if the skill
// identified by commandName was triggered (tool_use of "Skill" or "Read"
// referencing commandName).
func detectTrigger(r io.Reader, commandName string) bool {
	scanner := bufio.NewScanner(r)

	var pendingTool string
	var accumulatedJSON strings.Builder

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var ev streamEvent
		if err := json.Unmarshal([]byte(line), &ev); err != nil {
			continue
		}

		// Check content_block_start for tool_use
		if ev.Event != nil && ev.Event.Type == "content_block_start" {
			if cb := ev.Event.ContentBlock; cb != nil && cb.Type == "tool_use" {
				if cb.Name == "Skill" || cb.Name == "Read" {
					pendingTool = cb.Name
					accumulatedJSON.Reset()
					if len(cb.Input) > 0 {
						accumulatedJSON.Write(cb.Input)
						if strings.Contains(accumulatedJSON.String(), commandName) {
							return true
						}
					}
				} else {
					// Wrong tool — not our skill
					pendingTool = ""
					accumulatedJSON.Reset()
				}
			}
		}

		// Accumulate partial JSON from deltas
		if ev.Event != nil && ev.Event.Type == "content_block_delta" && pendingTool != "" {
			if d := ev.Event.Delta; d != nil {
				if d.PartialJSON != "" {
					accumulatedJSON.WriteString(d.PartialJSON)
				}
				if len(d.InputJSONDelta) > 0 {
					accumulatedJSON.Write(d.InputJSONDelta)
				}
				if strings.Contains(accumulatedJSON.String(), commandName) {
					return true
				}
			}
		}

		// On block stop, check accumulated
		if ev.Event != nil && (ev.Event.Type == "content_block_stop" || ev.Event.Type == "message_stop") {
			if pendingTool != "" && strings.Contains(accumulatedJSON.String(), commandName) {
				return true
			}
			pendingTool = ""
			accumulatedJSON.Reset()
		}

		// Fallback: check assistant message content for tool_use blocks
		if ev.Type == "assistant" && ev.Message != nil {
			for _, block := range ev.Message.Content {
				if block.Type == "tool_use" && (block.Name == "Skill" || block.Name == "Read") {
					if len(block.Input) > 0 && strings.Contains(string(block.Input), commandName) {
						return true
					}
				}
			}
		}

		// Fallback: check result payload
		if ev.Type == "result" && len(ev.Result) > 0 {
			if strings.Contains(string(ev.Result), commandName) {
				return true
			}
		}
	}

	return false
}

// RunSingleQuery creates a temporary command file for the skill, runs a single
// claude query, and detects whether the skill was triggered.
// It cleans up the command file on all exit paths.
func RunSingleQuery(ctx context.Context, query, skillName, description, projectRoot string, opts ...claude.Option) (bool, error) {
	commandName := generateCommandName(skillName)
	cmdFilePath, err := createCommandFile(projectRoot, commandName, description)
	if err != nil {
		return false, fmt.Errorf("setup command file: %w", err)
	}
	defer os.Remove(cmdFilePath)

	result, err := claude.RunQuery(ctx, query, opts...)
	if err != nil {
		return false, fmt.Errorf("claude query: %w", err)
	}

	// Re-serialize events to JSON lines for trigger detection
	triggered := detectTriggerFromEvents(result.Events, commandName)
	return triggered, nil
}

// detectTriggerFromEvents checks parsed stream events for skill trigger.
func detectTriggerFromEvents(events []claude.StreamEvent, commandName string) bool {
	for _, ev := range events {
		// Check assistant message content blocks
		if ev.Type == "assistant" && ev.Message != nil {
			for _, block := range ev.Message.Content {
				if block.Type == "tool_use" && (block.Name == "Skill" || block.Name == "Read") {
					if strings.Contains(block.ID, commandName) || strings.Contains(block.Text, commandName) {
						return true
					}
				}
			}
		}

		// Check raw event JSON for commandName reference
		if len(ev.Event) > 0 && strings.Contains(string(ev.Event), commandName) {
			return true
		}

		// Check result payload
		if len(ev.Result) > 0 && strings.Contains(string(ev.Result), commandName) {
			return true
		}
	}
	return false
}
