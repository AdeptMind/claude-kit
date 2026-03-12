package eval

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestParseEvalSet_ValidFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "evals.json")

	queries := []EvalQuery{
		{Query: "Review this pull request for issues", ShouldTrigger: true},
		{Query: "Write a function to parse CSV files", ShouldTrigger: false},
	}
	data, err := json.Marshal(queries)
	if err != nil {
		t.Fatalf("marshalling test data: %v", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("writing test file: %v", err)
	}

	got, err := ParseEvalSet(path)
	if err != nil {
		t.Fatalf("ParseEvalSet() unexpected error: %v", err)
	}

	if !reflect.DeepEqual(got, queries) {
		t.Errorf("ParseEvalSet() = %+v, want %+v", got, queries)
	}
}

func TestParseEvalSet_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "evals.json")

	if err := os.WriteFile(path, []byte(`{not valid json`), 0o644); err != nil {
		t.Fatalf("writing test file: %v", err)
	}

	_, err := ParseEvalSet(path)
	if err == nil {
		t.Fatal("ParseEvalSet() expected error for invalid JSON, got nil")
	}
}

func TestParseEvalSet_MissingQuery(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "evals.json")

	queries := []EvalQuery{
		{Query: "valid query", ShouldTrigger: true},
		{Query: "", ShouldTrigger: false},
	}
	data, err := json.Marshal(queries)
	if err != nil {
		t.Fatalf("marshalling test data: %v", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("writing test file: %v", err)
	}

	_, err = ParseEvalSet(path)
	if err == nil {
		t.Fatal("ParseEvalSet() expected error for empty query, got nil")
	}

	want := "evals.json: entry 1: query is empty"
	if err.Error() != want {
		t.Errorf("ParseEvalSet() error = %q, want %q", err.Error(), want)
	}
}

func TestParseEvalSet_FileNotFound(t *testing.T) {
	_, err := ParseEvalSet("/nonexistent/path/evals.json")
	if err == nil {
		t.Fatal("ParseEvalSet() expected error for missing file, got nil")
	}
}

func TestEvalReport_JSONRoundTrip(t *testing.T) {
	report := EvalReport{
		SkillName:   "code-reviewer",
		Description: "Code review skill evaluation",
		Results: []EvalResult{
			{
				Query:         "Review this PR",
				ShouldTrigger: true,
				TriggerRate:   0.95,
				Triggers:      19,
				Runs:          20,
				Pass:          true,
			},
			{
				Query:         "Write a CSV parser",
				ShouldTrigger: false,
				TriggerRate:   0.05,
				Triggers:      1,
				Runs:          20,
				Pass:          true,
			},
		},
		Summary: EvalSummary{
			Total:  2,
			Passed: 2,
			Failed: 0,
		},
	}

	data, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("Marshal() unexpected error: %v", err)
	}

	var got EvalReport
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal() unexpected error: %v", err)
	}

	if !reflect.DeepEqual(got, report) {
		t.Errorf("round-trip mismatch:\ngot  %+v\nwant %+v", got, report)
	}
}

func TestGradingResult_JSONRoundTrip(t *testing.T) {
	result := GradingResult{
		Assertion: "Response mentions security headers",
		Pass:      true,
		Evidence:  "Found reference to Content-Security-Policy",
		Claims:    []string{"CSP header mentioned", "HSTS recommended"},
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Marshal() unexpected error: %v", err)
	}

	var got GradingResult
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal() unexpected error: %v", err)
	}

	if !reflect.DeepEqual(got, result) {
		t.Errorf("round-trip mismatch:\ngot  %+v\nwant %+v", got, result)
	}
}

func TestBenchmarkStats_JSONRoundTrip(t *testing.T) {
	stats := BenchmarkStats{
		PassRate:    StatPair{Mean: 0.92, StdDev: 0.03},
		TimeSeconds: StatPair{Mean: 12.5, StdDev: 2.1},
		Tokens:      StatPair{Mean: 1500, StdDev: 300},
		Delta:       StatPair{Mean: 0.05, StdDev: 0.01},
	}

	data, err := json.Marshal(stats)
	if err != nil {
		t.Fatalf("Marshal() unexpected error: %v", err)
	}

	var got BenchmarkStats
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal() unexpected error: %v", err)
	}

	if !reflect.DeepEqual(got, stats) {
		t.Errorf("round-trip mismatch:\ngot  %+v\nwant %+v", got, stats)
	}
}
