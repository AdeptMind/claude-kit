package eval

import (
	"encoding/json"
	"fmt"
	"os"
)

// EvalQuery represents a single eval query from evals.json.
type EvalQuery struct {
	Query         string `json:"query"`
	ShouldTrigger bool   `json:"should_trigger"`
}

// EvalResult represents the result of running one eval query.
type EvalResult struct {
	Query         string  `json:"query"`
	ShouldTrigger bool    `json:"should_trigger"`
	TriggerRate   float64 `json:"trigger_rate"`
	Triggers      int     `json:"triggers"`
	Runs          int     `json:"runs"`
	Pass          bool    `json:"pass"`
}

// EvalSummary aggregates eval pass/fail counts.
type EvalSummary struct {
	Total  int `json:"total"`
	Passed int `json:"passed"`
	Failed int `json:"failed"`
}

// EvalReport is the full output of an eval run.
type EvalReport struct {
	SkillName   string       `json:"skill_name"`
	Description string       `json:"description"`
	Results     []EvalResult `json:"results"`
	Summary     EvalSummary  `json:"summary"`
}

// GradingAssertion defines a quality assertion to grade against.
type GradingAssertion struct {
	Description string `json:"description"`
	Type        string `json:"type"`
	Expected    string `json:"expected,omitempty"`
}

// GradingResult is the grader's verdict on one assertion.
type GradingResult struct {
	Assertion string   `json:"assertion"`
	Pass      bool     `json:"pass"`
	Evidence  string   `json:"evidence"`
	Claims    []string `json:"claims,omitempty"`
}

// StatPair holds mean and standard deviation.
type StatPair struct {
	Mean   float64 `json:"mean"`
	StdDev float64 `json:"stddev"`
}

// BenchmarkStats aggregates grading results into stats.
type BenchmarkStats struct {
	PassRate    StatPair `json:"pass_rate"`
	TimeSeconds StatPair `json:"time_seconds"`
	Tokens      StatPair `json:"tokens"`
	Delta       StatPair `json:"delta"`
}

// LoopHistory tracks one iteration of the eval-improve loop.
type LoopHistory struct {
	Iteration   int    `json:"iteration"`
	Description string `json:"description"`
	TrainPassed int    `json:"train_passed"`
	TrainTotal  int    `json:"train_total"`
	TestPassed  *int   `json:"test_passed,omitempty"`
	TestTotal   *int   `json:"test_total,omitempty"`
}

// ParseEvalSet reads and validates an evals.json file at the given path.
// It returns a slice of EvalQuery or a descriptive error if the file is
// missing, contains invalid JSON, or has entries with empty queries.
func ParseEvalSet(path string) ([]EvalQuery, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("evals.json: %w", err)
	}

	var queries []EvalQuery
	if err := json.Unmarshal(data, &queries); err != nil {
		return nil, fmt.Errorf("evals.json: invalid JSON: %w", err)
	}

	for i, q := range queries {
		if q.Query == "" {
			return nil, fmt.Errorf("evals.json: entry %d: query is empty", i)
		}
	}

	return queries, nil
}
