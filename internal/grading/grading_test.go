package grading

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/AdeptMind/infra-tool/claude-cli/internal/claude"
	"github.com/AdeptMind/infra-tool/claude-cli/internal/eval"
)

func TestBuildGradingPrompt_IncludesOutputAndAssertions(t *testing.T) {
	input := GradingInput{
		Output: "The server returns a 200 OK with JSON body.",
		Assertions: []eval.GradingAssertion{
			{Description: "mentions HTTP status", Type: "contains", Expected: "200"},
			{Description: "mentions JSON", Type: "contains", Expected: "JSON"},
		},
	}

	prompt, err := BuildGradingPrompt(input)
	if err != nil {
		t.Fatalf("BuildGradingPrompt() error: %v", err)
	}

	if !strings.Contains(prompt, input.Output) {
		t.Error("prompt does not contain the output text")
	}
	if !strings.Contains(prompt, "mentions HTTP status") {
		t.Error("prompt does not contain assertion description")
	}
	if !strings.Contains(prompt, "mentions JSON") {
		t.Error("prompt does not contain second assertion description")
	}
	if !strings.Contains(prompt, "You are a grading agent") {
		t.Error("prompt does not contain grading agent preamble")
	}
}

func TestParseGradingResponse_Valid(t *testing.T) {
	response := `[
		{"assertion":"mentions 200","pass":true,"evidence":"found 200 OK","claims":["status code present"]},
		{"assertion":"mentions JSON","pass":false,"evidence":"no JSON reference","claims":[]}
	]`

	results, err := ParseGradingResponse(response)
	if err != nil {
		t.Fatalf("ParseGradingResponse() error: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if !results[0].Pass {
		t.Error("expected first result to pass")
	}
	if results[1].Pass {
		t.Error("expected second result to fail")
	}
	if results[0].Assertion != "mentions 200" {
		t.Errorf("expected assertion 'mentions 200', got %q", results[0].Assertion)
	}
}

func TestParseGradingResponse_InvalidJSON(t *testing.T) {
	_, err := ParseGradingResponse("{not valid json")
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
	if !strings.Contains(err.Error(), "parsing grading response") {
		t.Errorf("expected 'parsing grading response' in error, got: %v", err)
	}
}

func TestParseGradingFile_Valid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "grading.json")

	assertions := []eval.GradingAssertion{
		{Description: "output mentions security", Type: "contains", Expected: "security"},
		{Description: "output is valid JSON", Type: "format"},
	}
	data, err := json.Marshal(assertions)
	if err != nil {
		t.Fatalf("marshalling test data: %v", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("writing test file: %v", err)
	}

	got, err := ParseGradingFile(path)
	if err != nil {
		t.Fatalf("ParseGradingFile() error: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 assertions, got %d", len(got))
	}
	if got[0].Description != "output mentions security" {
		t.Errorf("unexpected description: %q", got[0].Description)
	}
}

func TestParseGradingFile_MissingDescription(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "grading.json")

	assertions := []eval.GradingAssertion{
		{Description: "valid", Type: "contains"},
		{Description: "", Type: "contains"},
	}
	data, err := json.Marshal(assertions)
	if err != nil {
		t.Fatalf("marshalling test data: %v", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("writing test file: %v", err)
	}

	_, err = ParseGradingFile(path)
	if err == nil {
		t.Fatal("expected error for empty description, got nil")
	}

	want := "grading.json: entry 1: description is empty"
	if err.Error() != want {
		t.Errorf("error = %q, want %q", err.Error(), want)
	}
}

func TestParseGradingFile_MissingType(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "grading.json")

	assertions := []eval.GradingAssertion{
		{Description: "valid", Type: ""},
	}
	data, err := json.Marshal(assertions)
	if err != nil {
		t.Fatalf("marshalling test data: %v", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("writing test file: %v", err)
	}

	_, err = ParseGradingFile(path)
	if err == nil {
		t.Fatal("expected error for empty type, got nil")
	}

	want := "grading.json: entry 0: type is empty"
	if err.Error() != want {
		t.Errorf("error = %q, want %q", err.Error(), want)
	}
}

func TestParseGradingFile_FileNotFound(t *testing.T) {
	_, err := ParseGradingFile("/nonexistent/grading.json")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestParseGradingFile_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "grading.json")
	if err := os.WriteFile(path, []byte(`{bad json`), 0o644); err != nil {
		t.Fatalf("writing test file: %v", err)
	}

	_, err := ParseGradingFile(path)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
	if !strings.Contains(err.Error(), "invalid JSON") {
		t.Errorf("expected 'invalid JSON' in error, got: %v", err)
	}
}

func TestComputeSummary(t *testing.T) {
	results := []eval.GradingResult{
		{Assertion: "a", Pass: true},
		{Assertion: "b", Pass: false},
		{Assertion: "c", Pass: true},
		{Assertion: "d", Pass: true},
	}

	summary := ComputeSummary(results)

	if summary.Total != 4 {
		t.Errorf("Total = %d, want 4", summary.Total)
	}
	if summary.Passed != 3 {
		t.Errorf("Passed = %d, want 3", summary.Passed)
	}
	if summary.Failed != 1 {
		t.Errorf("Failed = %d, want 1", summary.Failed)
	}
}

func TestComputeSummary_Empty(t *testing.T) {
	summary := ComputeSummary(nil)
	if summary.Total != 0 || summary.Passed != 0 || summary.Failed != 0 {
		t.Errorf("expected all zeros, got %+v", summary)
	}
}

func TestGradeOutput_WithMockRunner(t *testing.T) {
	mockResponse := `[
		{"assertion":"checks format","pass":true,"evidence":"valid format","claims":["well structured"]},
		{"assertion":"checks content","pass":false,"evidence":"missing content","claims":[]}
	]`

	runner := func(_ context.Context, _ string, _ ...claude.Option) (string, error) {
		return mockResponse, nil
	}

	input := GradingInput{
		Output: "some output",
		Assertions: []eval.GradingAssertion{
			{Description: "checks format", Type: "format"},
			{Description: "checks content", Type: "contains", Expected: "keyword"},
		},
	}

	report, err := GradeOutput(context.Background(), input, runner)
	if err != nil {
		t.Fatalf("GradeOutput() error: %v", err)
	}

	if len(report.Results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(report.Results))
	}
	if report.Summary.Total != 2 {
		t.Errorf("Summary.Total = %d, want 2", report.Summary.Total)
	}
	if report.Summary.Passed != 1 {
		t.Errorf("Summary.Passed = %d, want 1", report.Summary.Passed)
	}
	if report.Summary.Failed != 1 {
		t.Errorf("Summary.Failed = %d, want 1", report.Summary.Failed)
	}
}

func TestGradeParallel_Concurrency(t *testing.T) {
	var callCount atomic.Int32

	runner := func(_ context.Context, _ string, _ ...claude.Option) (string, error) {
		callCount.Add(1)
		return `[{"assertion":"a","pass":true,"evidence":"ok","claims":[]}]`, nil
	}

	inputs := make([]GradingInput, 5)
	for i := range inputs {
		inputs[i] = GradingInput{
			Output: "output",
			Assertions: []eval.GradingAssertion{
				{Description: "assertion", Type: "contains"},
			},
		}
	}

	reports, err := GradeParallel(context.Background(), inputs, 3, runner)
	if err != nil {
		t.Fatalf("GradeParallel() error: %v", err)
	}

	if len(reports) != 5 {
		t.Fatalf("expected 5 reports, got %d", len(reports))
	}

	if int(callCount.Load()) != 5 {
		t.Errorf("expected 5 runner calls, got %d", callCount.Load())
	}

	for i, r := range reports {
		if r == nil {
			t.Errorf("report %d is nil", i)
			continue
		}
		if r.Summary.Total != 1 {
			t.Errorf("report %d: Summary.Total = %d, want 1", i, r.Summary.Total)
		}
		if !r.Results[0].Pass {
			t.Errorf("report %d: expected pass", i)
		}
	}
}

func TestGradeParallel_PropagatesError(t *testing.T) {
	runner := func(_ context.Context, prompt string, _ ...claude.Option) (string, error) {
		return "not json", nil
	}

	inputs := []GradingInput{
		{Output: "output", Assertions: []eval.GradingAssertion{{Description: "a", Type: "t"}}},
	}

	_, err := GradeParallel(context.Background(), inputs, 2, runner)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGradeParallel_DefaultWorkers(t *testing.T) {
	runner := func(_ context.Context, _ string, _ ...claude.Option) (string, error) {
		return `[{"assertion":"a","pass":true,"evidence":"ok","claims":[]}]`, nil
	}

	inputs := []GradingInput{
		{Output: "o", Assertions: []eval.GradingAssertion{{Description: "a", Type: "t"}}},
	}

	reports, err := GradeParallel(context.Background(), inputs, 0, runner)
	if err != nil {
		t.Fatalf("GradeParallel() error: %v", err)
	}
	if len(reports) != 1 {
		t.Fatalf("expected 1 report, got %d", len(reports))
	}
}
