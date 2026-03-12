package improve

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/AdeptMind/infra-tool/claude-cli/internal/claude"
	"github.com/AdeptMind/infra-tool/claude-cli/internal/eval"
)

func mockRunner(response string) PromptRunner {
	return func(_ context.Context, _ string, _ ...claude.Option) (string, error) {
		return response, nil
	}
}

func mockRunnerSequence(responses []string) PromptRunner {
	idx := 0
	return func(_ context.Context, _ string, _ ...claude.Option) (string, error) {
		if idx >= len(responses) {
			return "", fmt.Errorf("no more mock responses")
		}
		r := responses[idx]
		idx++
		return r, nil
	}
}

func mockRunnerErr(err error) PromptRunner {
	return func(_ context.Context, _ string, _ ...claude.Option) (string, error) {
		return "", err
	}
}

// --- buildImprovePrompt tests ---

func TestBuildImprovePrompt_IncludesFailures(t *testing.T) {
	input := ImproveInput{
		CurrentDescription: "A test skill",
		FailedResults: []eval.EvalResult{
			{Query: "deploy the app", ShouldTrigger: true, TriggerRate: 0.20},
			{Query: "write a poem", ShouldTrigger: false, TriggerRate: 0.80},
		},
		MaxDescriptionLen: 1024,
	}

	prompt := buildImprovePrompt(input, false)

	if !strings.Contains(prompt, "deploy the app") {
		t.Error("prompt should contain the failed query 'deploy the app'")
	}
	if !strings.Contains(prompt, "write a poem") {
		t.Error("prompt should contain the failed query 'write a poem'")
	}
	if !strings.Contains(prompt, "should trigger") {
		t.Error("prompt should contain 'should trigger' for positive failures")
	}
	if !strings.Contains(prompt, "should NOT trigger") {
		t.Error("prompt should contain 'should NOT trigger' for negative failures")
	}
	if !strings.Contains(prompt, "trigger_rate=0.20") {
		t.Error("prompt should contain trigger rate for failed queries")
	}
}

func TestBuildImprovePrompt_StripTestScores(t *testing.T) {
	testPassed := 8
	testTotal := 10
	input := ImproveInput{
		CurrentDescription: "A test skill",
		History: []eval.LoopHistory{
			{
				Iteration:   1,
				Description: "first attempt",
				TrainPassed: 5,
				TrainTotal:  10,
				TestPassed:  &testPassed,
				TestTotal:   &testTotal,
			},
		},
		MaxDescriptionLen: 1024,
	}

	prompt := buildImprovePrompt(input, false)

	if strings.Contains(prompt, "test_passed") || strings.Contains(prompt, "test_total") {
		t.Error("prompt should not contain test_passed or test_total fields")
	}
	if strings.Contains(prompt, "8/10 test") {
		t.Error("prompt should not contain test scores in any format")
	}
	if !strings.Contains(prompt, "5/10 train") {
		t.Error("prompt should contain train scores")
	}
}

func TestBuildImprovePrompt_IncludesHistory(t *testing.T) {
	input := ImproveInput{
		CurrentDescription: "A test skill",
		History: []eval.LoopHistory{
			{Iteration: 1, Description: "first try", TrainPassed: 3, TrainTotal: 10},
			{Iteration: 2, Description: "second try", TrainPassed: 7, TrainTotal: 10},
		},
		MaxDescriptionLen: 1024,
	}

	prompt := buildImprovePrompt(input, false)

	if !strings.Contains(prompt, "Iteration 1") {
		t.Error("prompt should contain Iteration 1")
	}
	if !strings.Contains(prompt, "Iteration 2") {
		t.Error("prompt should contain Iteration 2")
	}
	if !strings.Contains(prompt, "first try") {
		t.Error("prompt should contain description from iteration 1")
	}
	if !strings.Contains(prompt, "3/10 train") {
		t.Error("prompt should contain train score from iteration 1")
	}
	if !strings.Contains(prompt, "7/10 train") {
		t.Error("prompt should contain train score from iteration 2")
	}
}

func TestBuildImprovePrompt_StrictMode(t *testing.T) {
	input := ImproveInput{
		CurrentDescription: "A skill",
		MaxDescriptionLen:  512,
	}

	prompt := buildImprovePrompt(input, true)

	if !strings.Contains(prompt, "IMPORTANT") {
		t.Error("strict prompt should contain IMPORTANT warning")
	}
	if !strings.Contains(prompt, "too long") {
		t.Error("strict prompt should mention description was too long")
	}
}

func TestBuildImprovePrompt_IncludesSkillContent(t *testing.T) {
	input := ImproveInput{
		CurrentDescription: "A skill",
		SkillContent:       "# My Skill\nThis skill reviews code.",
		MaxDescriptionLen:  1024,
	}

	prompt := buildImprovePrompt(input, false)

	if !strings.Contains(prompt, "This skill reviews code.") {
		t.Error("prompt should contain skill content")
	}
}

// --- parseDescription tests ---

func TestParseDescription_ValidTags(t *testing.T) {
	response := "Here is the improved description:\n<new_description>A better skill description</new_description>\nDone."

	got, err := parseDescription(response)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "A better skill description" {
		t.Errorf("got %q, want %q", got, "A better skill description")
	}
}

func TestParseDescription_MissingTags(t *testing.T) {
	response := "Here is a description without XML tags."

	_, err := parseDescription(response)
	if err == nil {
		t.Fatal("expected error for missing tags, got nil")
	}
	if !strings.Contains(err.Error(), "no <new_description> tags") {
		t.Errorf("error should mention missing tags, got: %v", err)
	}
}

func TestParseDescription_MultilineContent(t *testing.T) {
	response := `<new_description>
This is a multiline
skill description that
spans several lines.
</new_description>`

	got, err := parseDescription(response)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(got, "multiline") {
		t.Error("should preserve multiline content")
	}
	if !strings.Contains(got, "several lines.") {
		t.Error("should preserve all lines")
	}
}

func TestParseDescription_TrimsWhitespace(t *testing.T) {
	response := "<new_description>  spaced out  </new_description>"

	got, err := parseDescription(response)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "spaced out" {
		t.Errorf("got %q, want %q", got, "spaced out")
	}
}

// --- ImproveDescription integration tests ---

func TestImproveDescription_Success(t *testing.T) {
	runner := mockRunner("<new_description>Improved skill</new_description>")

	out, err := ImproveDescription(context.Background(), runner, ImproveInput{
		CurrentDescription: "Original skill",
		MaxDescriptionLen:  1024,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.NewDescription != "Improved skill" {
		t.Errorf("got %q, want %q", out.NewDescription, "Improved skill")
	}
	if out.Retries != 0 {
		t.Errorf("expected 0 retries, got %d", out.Retries)
	}
}

func TestImproveDescription_RetriesOnTooLong(t *testing.T) {
	longDesc := strings.Repeat("x", 2000)
	shortDesc := "short"

	runner := mockRunnerSequence([]string{
		fmt.Sprintf("<new_description>%s</new_description>", longDesc),
		fmt.Sprintf("<new_description>%s</new_description>", shortDesc),
	})

	out, err := ImproveDescription(context.Background(), runner, ImproveInput{
		CurrentDescription: "Original",
		MaxDescriptionLen:  1024,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.NewDescription != shortDesc {
		t.Errorf("got %q, want %q", out.NewDescription, shortDesc)
	}
	if out.Retries != 1 {
		t.Errorf("expected 1 retry, got %d", out.Retries)
	}
}

func TestImproveDescription_ExhaustsRetries(t *testing.T) {
	longDesc := strings.Repeat("x", 2000)
	runner := mockRunner(fmt.Sprintf("<new_description>%s</new_description>", longDesc))

	_, err := ImproveDescription(context.Background(), runner, ImproveInput{
		CurrentDescription: "Original",
		MaxDescriptionLen:  1024,
	})
	if err == nil {
		t.Fatal("expected error after exhausting retries, got nil")
	}
	if !strings.Contains(err.Error(), "exceeded") {
		t.Errorf("error should mention exceeded retries, got: %v", err)
	}
}

func TestImproveDescription_RunnerError(t *testing.T) {
	runner := mockRunnerErr(fmt.Errorf("connection refused"))

	_, err := ImproveDescription(context.Background(), runner, ImproveInput{
		CurrentDescription: "Original",
	})
	if err == nil {
		t.Fatal("expected error from runner, got nil")
	}
	if !strings.Contains(err.Error(), "connection refused") {
		t.Errorf("error should propagate runner error, got: %v", err)
	}
}

func TestImproveDescription_MissingTags(t *testing.T) {
	runner := mockRunner("No tags here")

	_, err := ImproveDescription(context.Background(), runner, ImproveInput{
		CurrentDescription: "Original",
	})
	if err == nil {
		t.Fatal("expected error for missing tags, got nil")
	}
	if !strings.Contains(err.Error(), "no <new_description> tags") {
		t.Errorf("error should mention missing tags, got: %v", err)
	}
}

func TestImproveDescription_DefaultMaxLen(t *testing.T) {
	desc := strings.Repeat("a", 1025)
	shortDesc := "ok"

	runner := mockRunnerSequence([]string{
		fmt.Sprintf("<new_description>%s</new_description>", desc),
		fmt.Sprintf("<new_description>%s</new_description>", shortDesc),
	})

	out, err := ImproveDescription(context.Background(), runner, ImproveInput{
		CurrentDescription: "Original",
		// MaxDescriptionLen omitted => default 1024
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Retries != 1 {
		t.Errorf("expected 1 retry with default max len, got %d", out.Retries)
	}
}
