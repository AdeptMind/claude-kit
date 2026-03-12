package eval

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/AdeptMind/infra-tool/claude-cli/internal/claude"
)

// withMock replaces runSingleQueryFn for the duration of a test.
func withMock(t *testing.T, fn func(ctx context.Context, query, skillName, description, projectRoot string, opts ...claude.Option) (bool, error)) {
	t.Helper()
	original := runSingleQueryFn
	runSingleQueryFn = fn
	t.Cleanup(func() { runSingleQueryFn = original })
}

func TestRunEval_AllPass(t *testing.T) {
	withMock(t, func(_ context.Context, query, _, _, _ string, _ ...claude.Option) (bool, error) {
		// should_trigger queries always trigger, !should_trigger never trigger
		if query == "trigger me" {
			return true, nil
		}
		return false, nil
	})

	queries := []EvalQuery{
		{Query: "trigger me", ShouldTrigger: true},
		{Query: "ignore me", ShouldTrigger: false},
	}

	report, err := RunEval(context.Background(), queries, RunConfig{
		NumWorkers:       2,
		RunsPerQuery:     3,
		TriggerThreshold: 0.5,
		SkillName:        "test-skill",
		Description:      "test",
		ProjectRoot:      t.TempDir(),
	})
	if err != nil {
		t.Fatalf("RunEval() error: %v", err)
	}

	if report.Summary.Total != 2 {
		t.Errorf("total = %d, want 2", report.Summary.Total)
	}
	if report.Summary.Passed != 2 {
		t.Errorf("passed = %d, want 2", report.Summary.Passed)
	}
	if report.Summary.Failed != 0 {
		t.Errorf("failed = %d, want 0", report.Summary.Failed)
	}

	for _, r := range report.Results {
		if !r.Pass {
			t.Errorf("query %q: expected pass", r.Query)
		}
	}
}

func TestRunEval_SomeFail(t *testing.T) {
	withMock(t, func(_ context.Context, query, _, _, _ string, _ ...claude.Option) (bool, error) {
		// should_trigger query never triggers -> fail
		// !should_trigger query always triggers -> fail
		if query == "should trigger but wont" {
			return false, nil
		}
		return true, nil
	})

	queries := []EvalQuery{
		{Query: "should trigger but wont", ShouldTrigger: true},
		{Query: "should not trigger but will", ShouldTrigger: false},
	}

	report, err := RunEval(context.Background(), queries, RunConfig{
		NumWorkers:       2,
		RunsPerQuery:     3,
		TriggerThreshold: 0.5,
		SkillName:        "test",
		ProjectRoot:      t.TempDir(),
	})
	if err != nil {
		t.Fatalf("RunEval() error: %v", err)
	}

	if report.Summary.Failed != 2 {
		t.Errorf("failed = %d, want 2", report.Summary.Failed)
	}
	if report.Summary.Passed != 0 {
		t.Errorf("passed = %d, want 0", report.Summary.Passed)
	}
}

func TestRunEval_MultipleRunsPerQuery(t *testing.T) {
	var callCount atomic.Int32
	withMock(t, func(_ context.Context, _, _, _, _ string, _ ...claude.Option) (bool, error) {
		callCount.Add(1)
		return true, nil
	})

	queries := []EvalQuery{
		{Query: "q1", ShouldTrigger: true},
		{Query: "q2", ShouldTrigger: true},
	}
	runsPerQuery := 5

	_, err := RunEval(context.Background(), queries, RunConfig{
		NumWorkers:   4,
		RunsPerQuery: runsPerQuery,
		SkillName:    "test",
		ProjectRoot:  t.TempDir(),
	})
	if err != nil {
		t.Fatalf("RunEval() error: %v", err)
	}

	expected := int32(len(queries) * runsPerQuery)
	if got := callCount.Load(); got != expected {
		t.Errorf("call count = %d, want %d", got, expected)
	}
}

func TestRunEval_ConcurrencyLimit(t *testing.T) {
	const maxWorkers = 2
	var concurrent atomic.Int32
	var maxSeen atomic.Int32

	withMock(t, func(_ context.Context, _, _, _, _ string, _ ...claude.Option) (bool, error) {
		cur := concurrent.Add(1)
		// Track the max concurrent goroutines
		for {
			old := maxSeen.Load()
			if cur <= old || maxSeen.CompareAndSwap(old, cur) {
				break
			}
		}
		time.Sleep(10 * time.Millisecond)
		concurrent.Add(-1)
		return true, nil
	})

	queries := make([]EvalQuery, 6)
	for i := range queries {
		queries[i] = EvalQuery{Query: fmt.Sprintf("q%d", i), ShouldTrigger: true}
	}

	_, err := RunEval(context.Background(), queries, RunConfig{
		NumWorkers:   maxWorkers,
		RunsPerQuery: 1,
		SkillName:    "test",
		ProjectRoot:  t.TempDir(),
	})
	if err != nil {
		t.Fatalf("RunEval() error: %v", err)
	}

	if got := maxSeen.Load(); got > int32(maxWorkers) {
		t.Errorf("max concurrent = %d, want <= %d", got, maxWorkers)
	}
}

func TestRunEval_Summary(t *testing.T) {
	var callIdx atomic.Int32
	withMock(t, func(_ context.Context, query, _, _, _ string, _ ...claude.Option) (bool, error) {
		callIdx.Add(1)
		// "pass-trigger" triggers 100%, "fail-trigger" triggers 0%
		if query == "pass-trigger" {
			return true, nil
		}
		return false, nil
	})

	queries := []EvalQuery{
		{Query: "pass-trigger", ShouldTrigger: true},
		{Query: "fail-trigger", ShouldTrigger: true},
		{Query: "pass-no-trigger", ShouldTrigger: false},
	}

	report, err := RunEval(context.Background(), queries, RunConfig{
		NumWorkers:       4,
		RunsPerQuery:     3,
		TriggerThreshold: 0.5,
		SkillName:        "test",
		ProjectRoot:      t.TempDir(),
	})
	if err != nil {
		t.Fatalf("RunEval() error: %v", err)
	}

	if report.Summary.Total != 3 {
		t.Errorf("total = %d, want 3", report.Summary.Total)
	}
	if report.Summary.Passed != 2 {
		t.Errorf("passed = %d, want 2", report.Summary.Passed)
	}
	if report.Summary.Failed != 1 {
		t.Errorf("failed = %d, want 1", report.Summary.Failed)
	}

	// Verify trigger rates
	for _, r := range report.Results {
		switch r.Query {
		case "pass-trigger":
			if r.TriggerRate != 1.0 {
				t.Errorf("pass-trigger rate = %f, want 1.0", r.TriggerRate)
			}
		case "fail-trigger":
			if r.TriggerRate != 0.0 {
				t.Errorf("fail-trigger rate = %f, want 0.0", r.TriggerRate)
			}
		case "pass-no-trigger":
			if r.TriggerRate != 0.0 {
				t.Errorf("pass-no-trigger rate = %f, want 0.0", r.TriggerRate)
			}
		}
	}
}

func TestRunEval_DefaultConfig(t *testing.T) {
	var callCount atomic.Int32
	withMock(t, func(_ context.Context, _, _, _, _ string, _ ...claude.Option) (bool, error) {
		callCount.Add(1)
		return true, nil
	})

	queries := []EvalQuery{
		{Query: "q1", ShouldTrigger: true},
	}

	report, err := RunEval(context.Background(), queries, RunConfig{
		SkillName:   "test",
		ProjectRoot: t.TempDir(),
	})
	if err != nil {
		t.Fatalf("RunEval() error: %v", err)
	}

	// Default RunsPerQuery = 3
	if got := callCount.Load(); got != 3 {
		t.Errorf("call count = %d, want 3 (default RunsPerQuery)", got)
	}

	// Default TriggerThreshold = 0.5, rate = 1.0 -> pass
	if !report.Results[0].Pass {
		t.Error("expected pass with default threshold and 100% trigger rate")
	}

	if report.SkillName != "test" {
		t.Errorf("skill name = %q, want %q", report.SkillName, "test")
	}
}
