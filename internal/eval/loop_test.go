package eval

import (
	"context"
	"fmt"
	"testing"
)

// withEvalMock replaces runEvalFn for the duration of a test.
func withEvalMock(t *testing.T, fn func(ctx context.Context, queries []EvalQuery, cfg RunConfig) (*EvalReport, error)) {
	t.Helper()
	original := runEvalFn
	runEvalFn = fn
	t.Cleanup(func() { runEvalFn = original })
}

// makeReport builds an EvalReport from pass/fail status per query.
func makeReport(queries []EvalQuery, passes []bool, desc string) *EvalReport {
	results := make([]EvalResult, len(queries))
	summary := EvalSummary{Total: len(queries)}
	for i, q := range queries {
		rate := 0.0
		if passes[i] && q.ShouldTrigger {
			rate = 1.0
		}
		if !passes[i] && !q.ShouldTrigger {
			rate = 1.0
		}
		results[i] = EvalResult{
			Query:         q.Query,
			ShouldTrigger: q.ShouldTrigger,
			TriggerRate:   rate,
			Triggers:      3,
			Runs:          3,
			Pass:          passes[i],
		}
		if passes[i] {
			summary.Passed++
		} else {
			summary.Failed++
		}
	}
	return &EvalReport{
		Description: desc,
		Results:     results,
		Summary:     summary,
	}
}

func allPass(n int) []bool {
	p := make([]bool, n)
	for i := range p {
		p[i] = true
	}
	return p
}

func mockImprover(newDesc string) DescriptionImprover {
	return func(_ context.Context, _ ImproveInput) (*ImproveResult, error) {
		return &ImproveResult{NewDescription: newDesc}, nil
	}
}

// --- StratifiedSplit tests ---

func TestStratifiedSplit_BalancedSplit(t *testing.T) {
	queries := []EvalQuery{
		{Query: "p1", ShouldTrigger: true},
		{Query: "p2", ShouldTrigger: true},
		{Query: "p3", ShouldTrigger: true},
		{Query: "p4", ShouldTrigger: true},
		{Query: "p5", ShouldTrigger: true},
		{Query: "n1", ShouldTrigger: false},
		{Query: "n2", ShouldTrigger: false},
		{Query: "n3", ShouldTrigger: false},
		{Query: "n4", ShouldTrigger: false},
		{Query: "n5", ShouldTrigger: false},
	}

	split := StratifiedSplit(queries, 0.6)

	// 5 positive * 0.6 = 3 train, 2 test
	// 5 negative * 0.6 = 3 train, 2 test
	if len(split.Train) != 6 {
		t.Errorf("train size = %d, want 6", len(split.Train))
	}
	if len(split.Test) != 4 {
		t.Errorf("test size = %d, want 4", len(split.Test))
	}

	// Check both groups are represented in train.
	var trainPos, trainNeg int
	for _, q := range split.Train {
		if q.ShouldTrigger {
			trainPos++
		} else {
			trainNeg++
		}
	}
	if trainPos != 3 {
		t.Errorf("train positive = %d, want 3", trainPos)
	}
	if trainNeg != 3 {
		t.Errorf("train negative = %d, want 3", trainNeg)
	}
}

func TestStratifiedSplit_SmallSet(t *testing.T) {
	queries := []EvalQuery{
		{Query: "p1", ShouldTrigger: true},
		{Query: "n1", ShouldTrigger: false},
	}

	split := StratifiedSplit(queries, 0.7)

	// With 1 per group, each goes fully to train (no split possible).
	if len(split.Train) != 2 {
		t.Errorf("train size = %d, want 2", len(split.Train))
	}
	if len(split.Test) != 0 {
		t.Errorf("test size = %d, want 0", len(split.Test))
	}
}

func TestStratifiedSplit_AllPositive(t *testing.T) {
	queries := []EvalQuery{
		{Query: "p1", ShouldTrigger: true},
		{Query: "p2", ShouldTrigger: true},
		{Query: "p3", ShouldTrigger: true},
		{Query: "p4", ShouldTrigger: true},
	}

	split := StratifiedSplit(queries, 0.7)

	// int(4 * 0.7) = 2, and 2 > 0 and 2 < 4, so 2 train, 2 test.
	if len(split.Train) != 2 {
		t.Errorf("train size = %d, want 2", len(split.Train))
	}
	if len(split.Test) != 2 {
		t.Errorf("test size = %d, want 2", len(split.Test))
	}
}

// --- RunLoop tests ---

func TestRunLoop_StopsOnAllPassed(t *testing.T) {
	withEvalMock(t, func(_ context.Context, queries []EvalQuery, cfg RunConfig) (*EvalReport, error) {
		return makeReport(queries, allPass(len(queries)), cfg.Description), nil
	})

	queries := []EvalQuery{
		{Query: "p1", ShouldTrigger: true},
		{Query: "p2", ShouldTrigger: true},
		{Query: "n1", ShouldTrigger: false},
		{Query: "n2", ShouldTrigger: false},
	}

	result, err := RunLoop(context.Background(), queries, "initial desc",
		mockImprover("unused"), LoopConfig{MaxIterations: 5})
	if err != nil {
		t.Fatalf("RunLoop() error: %v", err)
	}

	if result.Iterations != 1 {
		t.Errorf("iterations = %d, want 1 (should stop on first all-pass)", result.Iterations)
	}
	if result.BestDescription != "initial desc" {
		t.Errorf("best description = %q, want %q", result.BestDescription, "initial desc")
	}
}

func TestRunLoop_ImprovesDescription(t *testing.T) {
	withEvalMock(t, func(_ context.Context, queries []EvalQuery, cfg RunConfig) (*EvalReport, error) {
		// First train eval: fail. Second train eval (after improve): pass. Then test: pass.
		if cfg.Description == "initial desc" {
			passes := make([]bool, len(queries))
			for i, q := range queries {
				passes[i] = !q.ShouldTrigger // only negatives pass
			}
			return makeReport(queries, passes, cfg.Description), nil
		}
		// "improved desc" -> all pass
		return makeReport(queries, allPass(len(queries)), cfg.Description), nil
	})

	queries := []EvalQuery{
		{Query: "p1", ShouldTrigger: true},
		{Query: "p2", ShouldTrigger: true},
		{Query: "p3", ShouldTrigger: true},
		{Query: "n1", ShouldTrigger: false},
		{Query: "n2", ShouldTrigger: false},
		{Query: "n3", ShouldTrigger: false},
	}

	result, err := RunLoop(context.Background(), queries, "initial desc",
		mockImprover("improved desc"), LoopConfig{MaxIterations: 5})
	if err != nil {
		t.Fatalf("RunLoop() error: %v", err)
	}

	if result.Iterations != 2 {
		t.Errorf("iterations = %d, want 2", result.Iterations)
	}
	if result.BestDescription != "improved desc" {
		t.Errorf("best description = %q, want %q", result.BestDescription, "improved desc")
	}
	if len(result.History) != 2 {
		t.Errorf("history length = %d, want 2", len(result.History))
	}
	// First iteration: no test scores.
	if result.History[0].TestPassed != nil {
		t.Error("first iteration should not have test scores")
	}
	// Second iteration: should have test scores.
	if result.History[1].TestPassed == nil {
		t.Error("second iteration should have test scores")
	}
}

func TestRunLoop_MaxIterations(t *testing.T) {
	withEvalMock(t, func(_ context.Context, queries []EvalQuery, cfg RunConfig) (*EvalReport, error) {
		// Always fail one query.
		passes := make([]bool, len(queries))
		for i := range passes {
			passes[i] = i > 0 // first query always fails
		}
		return makeReport(queries, passes, cfg.Description), nil
	})

	iterationCount := 0
	improver := func(_ context.Context, _ ImproveInput) (*ImproveResult, error) {
		iterationCount++
		return &ImproveResult{NewDescription: fmt.Sprintf("attempt %d", iterationCount)}, nil
	}

	queries := []EvalQuery{
		{Query: "p1", ShouldTrigger: true},
		{Query: "p2", ShouldTrigger: true},
		{Query: "p3", ShouldTrigger: true},
		{Query: "n1", ShouldTrigger: false},
		{Query: "n2", ShouldTrigger: false},
		{Query: "n3", ShouldTrigger: false},
	}

	maxIter := 3
	result, err := RunLoop(context.Background(), queries, "initial",
		improver, LoopConfig{MaxIterations: maxIter})
	if err != nil {
		t.Fatalf("RunLoop() error: %v", err)
	}

	if result.Iterations != maxIter {
		t.Errorf("iterations = %d, want %d", result.Iterations, maxIter)
	}
	// No test scores since train never fully passed.
	if result.BestTestScore != 0 {
		t.Errorf("best test score = %f, want 0", result.BestTestScore)
	}
	// Best description is last iteration's.
	last := result.History[len(result.History)-1].Description
	if result.BestDescription != last {
		t.Errorf("best description = %q, want %q (last iteration)", result.BestDescription, last)
	}
}

func TestRunLoop_SelectsBestByTestScore(t *testing.T) {
	withEvalMock(t, func(_ context.Context, queries []EvalQuery, cfg RunConfig) (*EvalReport, error) {
		switch {
		case cfg.Description == "desc-v1":
			// First train: all fail.
			passes := make([]bool, len(queries))
			return makeReport(queries, passes, cfg.Description), nil

		case cfg.Description == "desc-v2":
			// Second train: all pass. Test: all pass.
			return makeReport(queries, allPass(len(queries)), cfg.Description), nil

		default:
			// Final eval: all pass.
			return makeReport(queries, allPass(len(queries)), cfg.Description), nil
		}
	})

	queries := []EvalQuery{
		{Query: "p1", ShouldTrigger: true},
		{Query: "p2", ShouldTrigger: true},
		{Query: "p3", ShouldTrigger: true},
		{Query: "n1", ShouldTrigger: false},
		{Query: "n2", ShouldTrigger: false},
		{Query: "n3", ShouldTrigger: false},
	}

	result, err := RunLoop(context.Background(), queries, "desc-v1",
		mockImprover("desc-v2"), LoopConfig{MaxIterations: 5})
	if err != nil {
		t.Fatalf("RunLoop() error: %v", err)
	}

	if result.BestDescription != "desc-v2" {
		t.Errorf("best description = %q, want %q", result.BestDescription, "desc-v2")
	}
	if result.BestTestScore != 1.0 {
		t.Errorf("best test score = %f, want 1.0", result.BestTestScore)
	}
	if result.FinalReport == nil {
		t.Fatal("final report should not be nil")
	}
}
