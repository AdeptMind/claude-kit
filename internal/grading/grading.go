package grading

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/AdeptMind/infra-tool/claude-cli/internal/claude"
	"github.com/AdeptMind/infra-tool/claude-cli/internal/eval"
)

// GradingInput holds the output text and assertions to grade against.
type GradingInput struct {
	Output     string                 `json:"output"`
	Assertions []eval.GradingAssertion `json:"assertions"`
}

// GradingReport holds the grading results and summary.
type GradingReport struct {
	Results []eval.GradingResult `json:"results"`
	Summary GradingSummary       `json:"summary"`
}

// GradingSummary aggregates pass/fail counts.
type GradingSummary struct {
	Total  int `json:"total"`
	Passed int `json:"passed"`
	Failed int `json:"failed"`
}

// PromptRunner abstracts the claude prompt invocation for testability.
type PromptRunner func(ctx context.Context, prompt string, opts ...claude.Option) (string, error)

// ParseGradingFile reads and validates a grading.json file.
func ParseGradingFile(path string) ([]eval.GradingAssertion, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("grading.json: %w", err)
	}

	var assertions []eval.GradingAssertion
	if err := json.Unmarshal(data, &assertions); err != nil {
		return nil, fmt.Errorf("grading.json: invalid JSON: %w", err)
	}

	for i, a := range assertions {
		if a.Description == "" {
			return nil, fmt.Errorf("grading.json: entry %d: description is empty", i)
		}
		if a.Type == "" {
			return nil, fmt.Errorf("grading.json: entry %d: type is empty", i)
		}
	}

	return assertions, nil
}

// BuildGradingPrompt constructs the grader prompt from the output and assertions.
func BuildGradingPrompt(input GradingInput) (string, error) {
	assertionsJSON, err := json.Marshal(input.Assertions)
	if err != nil {
		return "", fmt.Errorf("marshalling assertions: %w", err)
	}

	prompt := fmt.Sprintf(`You are a grading agent. Grade the following output against each assertion.

## Output to grade
%s

## Assertions
%s

## Instructions
For each assertion, return a JSON object with: assertion (string), pass (bool), evidence (string), claims ([]string).
Return ONLY a JSON array, no other text.`, input.Output, string(assertionsJSON))

	return prompt, nil
}

// ParseGradingResponse parses the JSON response from the grader into results.
func ParseGradingResponse(response string) ([]eval.GradingResult, error) {
	var results []eval.GradingResult
	if err := json.Unmarshal([]byte(response), &results); err != nil {
		return nil, fmt.Errorf("parsing grading response: %w", err)
	}
	return results, nil
}

// ComputeSummary calculates pass/fail counts from grading results.
func ComputeSummary(results []eval.GradingResult) GradingSummary {
	s := GradingSummary{Total: len(results)}
	for _, r := range results {
		if r.Pass {
			s.Passed++
		} else {
			s.Failed++
		}
	}
	return s
}

// GradeOutput grades the given input by invoking the prompt runner.
func GradeOutput(ctx context.Context, input GradingInput, runner PromptRunner, opts ...claude.Option) (*GradingReport, error) {
	prompt, err := BuildGradingPrompt(input)
	if err != nil {
		return nil, err
	}

	response, err := runner(ctx, prompt, opts...)
	if err != nil {
		return nil, fmt.Errorf("grading prompt failed: %w", err)
	}

	results, err := ParseGradingResponse(response)
	if err != nil {
		return nil, err
	}

	return &GradingReport{
		Results: results,
		Summary: ComputeSummary(results),
	}, nil
}

// GradeParallel grades multiple inputs concurrently with a configurable worker count.
func GradeParallel(ctx context.Context, inputs []GradingInput, workers int, runner PromptRunner, opts ...claude.Option) ([]*GradingReport, error) {
	if workers < 1 {
		workers = 1
	}

	reports := make([]*GradingReport, len(inputs))
	errs := make([]error, len(inputs))

	sem := make(chan struct{}, workers)
	var wg sync.WaitGroup

	for i, input := range inputs {
		wg.Add(1)
		go func(idx int, inp GradingInput) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			report, err := GradeOutput(ctx, inp, runner, opts...)
			if err != nil {
				errs[idx] = err
				return
			}
			reports[idx] = report
		}(i, input)
	}

	wg.Wait()

	for i, err := range errs {
		if err != nil {
			return nil, fmt.Errorf("grading input %d: %w", i, err)
		}
	}

	return reports, nil
}
