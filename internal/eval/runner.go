package eval

import (
	"context"
	"sync"
	"time"

	"github.com/AdeptMind/infra-tool/claude-cli/internal/claude"
)

const (
	defaultNumWorkers       = 10
	defaultRunsPerQuery     = 3
	defaultTriggerThreshold = 0.5
)

// RunConfig controls parallel eval execution.
type RunConfig struct {
	NumWorkers       int
	RunsPerQuery     int
	TriggerThreshold float64
	SkillName        string
	Description      string
	ProjectRoot      string
	Model            string
	Timeout          time.Duration
}

// applyDefaults fills zero-value fields with sensible defaults.
func (c *RunConfig) applyDefaults() {
	if c.NumWorkers <= 0 {
		c.NumWorkers = defaultNumWorkers
	}
	if c.RunsPerQuery <= 0 {
		c.RunsPerQuery = defaultRunsPerQuery
	}
	if c.TriggerThreshold <= 0 {
		c.TriggerThreshold = defaultTriggerThreshold
	}
}

// runSingleQueryFn is the function used to run a single query.
// It can be replaced in tests to inject a mock.
var runSingleQueryFn = RunSingleQuery

// queryResult collects trigger counts for a single EvalQuery across multiple runs.
type queryResult struct {
	index    int
	triggers int
	runs     int
	err      error
}

// RunEval executes all queries concurrently with a semaphore-based worker pool.
// Each query is run RunsPerQuery times. Pass/fail is determined by TriggerThreshold.
func RunEval(ctx context.Context, queries []EvalQuery, cfg RunConfig) (*EvalReport, error) {
	cfg.applyDefaults()

	var opts []claude.Option
	if cfg.Model != "" {
		opts = append(opts, claude.WithModel(cfg.Model))
	}

	type runResult struct {
		queryIdx  int
		triggered bool
		err       error
	}

	totalRuns := len(queries) * cfg.RunsPerQuery
	results := make(chan runResult, totalRuns)

	sem := make(chan struct{}, cfg.NumWorkers)
	var wg sync.WaitGroup

	for qi, q := range queries {
		for r := 0; r < cfg.RunsPerQuery; r++ {
			wg.Add(1)
			go func(idx int, query string) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()

				triggered, err := runSingleQueryFn(ctx, query, cfg.SkillName, cfg.Description, cfg.ProjectRoot, opts...)
				results <- runResult{queryIdx: idx, triggered: triggered, err: err}
			}(qi, q.Query)
		}
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	// Aggregate per-query
	type aggregate struct {
		triggers int
		runs     int
	}
	agg := make([]aggregate, len(queries))
	for rr := range results {
		if rr.err != nil {
			// Count failed runs but don't abort; the run just didn't trigger.
			agg[rr.queryIdx].runs++
			continue
		}
		agg[rr.queryIdx].runs++
		if rr.triggered {
			agg[rr.queryIdx].triggers++
		}
	}

	// Build results
	evalResults := make([]EvalResult, len(queries))
	summary := EvalSummary{Total: len(queries)}

	for i, q := range queries {
		rate := 0.0
		if agg[i].runs > 0 {
			rate = float64(agg[i].triggers) / float64(agg[i].runs)
		}

		pass := false
		if q.ShouldTrigger {
			pass = rate >= cfg.TriggerThreshold
		} else {
			pass = rate < cfg.TriggerThreshold
		}

		evalResults[i] = EvalResult{
			Query:         q.Query,
			ShouldTrigger: q.ShouldTrigger,
			TriggerRate:   rate,
			Triggers:      agg[i].triggers,
			Runs:          agg[i].runs,
			Pass:          pass,
		}

		if pass {
			summary.Passed++
		} else {
			summary.Failed++
		}
	}

	return &EvalReport{
		SkillName:   cfg.SkillName,
		Description: cfg.Description,
		Results:     evalResults,
		Summary:     summary,
	}, nil
}
