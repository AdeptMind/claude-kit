package improve

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/AdeptMind/infra-tool/claude-cli/internal/claude"
	"github.com/AdeptMind/infra-tool/claude-cli/internal/eval"
)

const (
	defaultMaxDescriptionLen = 1024
	maxRetries               = 3
)

var descriptionTagRe = regexp.MustCompile(`(?s)<new_description>(.*?)</new_description>`)

// PromptRunner abstracts the call to claude so it can be replaced in tests.
type PromptRunner func(ctx context.Context, prompt string, opts ...claude.Option) (string, error)

// ImproveInput holds the data needed to build an improvement prompt.
type ImproveInput struct {
	CurrentDescription string
	SkillContent       string
	FailedResults      []eval.EvalResult
	History            []eval.LoopHistory
	MaxDescriptionLen  int
}

// ImproveOutput contains the result of the improvement attempt.
type ImproveOutput struct {
	NewDescription string
	Retries        int
}

// ImproveDescription asks an LLM to produce a better skill description based on
// the current description, failed eval results, and iteration history. It retries
// up to maxRetries times if the returned description exceeds MaxDescriptionLen.
func ImproveDescription(ctx context.Context, runner PromptRunner, input ImproveInput, opts ...claude.Option) (*ImproveOutput, error) {
	if input.MaxDescriptionLen <= 0 {
		input.MaxDescriptionLen = defaultMaxDescriptionLen
	}

	var retries int
	for range maxRetries + 1 {
		prompt := buildImprovePrompt(input, retries > 0)
		response, err := runner(ctx, prompt, opts...)
		if err != nil {
			return nil, fmt.Errorf("improve description: %w", err)
		}

		desc, err := parseDescription(response)
		if err != nil {
			return nil, fmt.Errorf("improve description: %w", err)
		}

		if len(desc) <= input.MaxDescriptionLen {
			return &ImproveOutput{NewDescription: desc, Retries: retries}, nil
		}
		retries++
	}

	return nil, fmt.Errorf("improve description: exceeded %d retries, description still too long", maxRetries)
}

// buildImprovePrompt assembles the full prompt sent to the LLM.
func buildImprovePrompt(input ImproveInput, strict bool) string {
	var b strings.Builder

	b.WriteString("You are improving a skill description used for intent matching.\n\n")

	b.WriteString("## Current Description\n")
	b.WriteString(input.CurrentDescription)
	b.WriteString("\n\n")

	if input.SkillContent != "" {
		b.WriteString("## Skill Content\n")
		b.WriteString(input.SkillContent)
		b.WriteString("\n\n")
	}

	if len(input.FailedResults) > 0 {
		b.WriteString("## Failed Queries\n")
		for _, r := range input.FailedResults {
			expected := "should NOT trigger"
			if r.ShouldTrigger {
				expected = "should trigger"
			}
			fmt.Fprintf(&b, "- Query: %q (%s, trigger_rate=%.2f)\n", r.Query, expected, r.TriggerRate)
		}
		b.WriteString("\n")
	}

	if len(input.History) > 0 {
		b.WriteString("## Iteration History (train scores only)\n")
		for _, h := range input.History {
			fmt.Fprintf(&b, "- Iteration %d: %d/%d train passed | Description: %s\n",
				h.Iteration, h.TrainPassed, h.TrainTotal, h.Description)
		}
		b.WriteString("\n")
	}

	maxLen := input.MaxDescriptionLen
	if maxLen <= 0 {
		maxLen = defaultMaxDescriptionLen
	}

	b.WriteString("## Instructions\n")
	if strict {
		fmt.Fprintf(&b, "IMPORTANT: Your previous description was too long. Keep it under %d characters.\n", maxLen)
	}
	fmt.Fprintf(&b, "Write an improved description that better matches the intended queries and rejects unintended ones. Maximum %d characters.\n", maxLen)
	b.WriteString("Return your answer inside <new_description> tags.\n")

	return b.String()
}

// parseDescription extracts the text between <new_description> tags.
func parseDescription(response string) (string, error) {
	matches := descriptionTagRe.FindStringSubmatch(response)
	if len(matches) < 2 {
		return "", fmt.Errorf("no <new_description> tags found in response")
	}
	return strings.TrimSpace(matches[1]), nil
}
