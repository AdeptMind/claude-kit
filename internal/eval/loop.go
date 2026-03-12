package eval

import (
	"context"
	"fmt"
)

const (
	defaultMaxIterations = 10
	defaultTrainRatio    = 0.7
)

// LoopConfig controls the eval-improve loop.
type LoopConfig struct {
	MaxIterations     int
	TrainRatio        float64
	RunConfig         RunConfig
	MaxDescriptionLen int
}

func (c *LoopConfig) applyLoopDefaults() {
	if c.MaxIterations <= 0 {
		c.MaxIterations = defaultMaxIterations
	}
	if c.TrainRatio <= 0 || c.TrainRatio >= 1 {
		c.TrainRatio = defaultTrainRatio
	}
}

// SplitResult holds the train/test partition of eval queries.
type SplitResult struct {
	Train []EvalQuery
	Test  []EvalQuery
}

// StratifiedSplit partitions queries into train/test sets, applying trainRatio
// independently to positive (ShouldTrigger=true) and negative groups.
// Deterministic: first N of each group go to train.
func StratifiedSplit(queries []EvalQuery, trainRatio float64) SplitResult {
	var pos, neg []EvalQuery
	for _, q := range queries {
		if q.ShouldTrigger {
			pos = append(pos, q)
		} else {
			neg = append(neg, q)
		}
	}

	var result SplitResult
	splitGroup(&result, pos, trainRatio)
	splitGroup(&result, neg, trainRatio)
	return result
}

// splitGroup splits a single group (pos or neg) and appends to result.
func splitGroup(result *SplitResult, group []EvalQuery, trainRatio float64) {
	if len(group) == 0 {
		return
	}

	// Single element: cannot split, assign to train.
	if len(group) == 1 {
		result.Train = append(result.Train, group...)
		return
	}

	n := int(float64(len(group)) * trainRatio)

	// Ensure at least 1 in each set when group has >1 queries.
	if n == 0 {
		n = 1
	}
	if n == len(group) {
		n = len(group) - 1
	}

	result.Train = append(result.Train, group[:n]...)
	result.Test = append(result.Test, group[n:]...)
}

// LoopResult holds the outcome of the eval-improve loop.
type LoopResult struct {
	BestDescription string
	BestTestScore   float64
	Iterations      int
	History         []LoopHistory
	FinalReport     *EvalReport
}

// ImproveInput holds the data needed by the improve function.
type ImproveInput struct {
	CurrentDescription string
	SkillContent       string
	FailedResults      []EvalResult
	History            []LoopHistory
	MaxDescriptionLen  int
}

// ImproveResult holds the output of the improve function.
type ImproveResult struct {
	NewDescription string
}

// DescriptionImprover is a function that produces an improved description
// from failed eval results and history. It abstracts improve.ImproveDescription
// to avoid a circular import between eval and improve.
type DescriptionImprover func(ctx context.Context, input ImproveInput) (*ImproveResult, error)

// runEvalFn is the function used to run eval. Replaced in tests.
var runEvalFn = RunEval

// RunLoop executes the eval-improve loop: evaluate on train, improve on failures,
// validate on test when train passes. Returns the best result by test score.
func RunLoop(ctx context.Context, queries []EvalQuery, initialDescription string, improver DescriptionImprover, cfg LoopConfig) (*LoopResult, error) {
	cfg.applyLoopDefaults()

	split := StratifiedSplit(queries, cfg.TrainRatio)
	if len(split.Train) == 0 {
		return nil, fmt.Errorf("eval loop: no train queries after split")
	}

	currentDesc := initialDescription
	var history []LoopHistory
	bestIdx := -1
	bestTestScore := -1.0

	for i := 1; i <= cfg.MaxIterations; i++ {
		// Run eval on train set.
		trainCfg := cfg.RunConfig
		trainCfg.Description = currentDesc
		trainReport, err := runEvalFn(ctx, split.Train, trainCfg)
		if err != nil {
			return nil, fmt.Errorf("eval loop iteration %d train: %w", i, err)
		}

		entry := LoopHistory{
			Iteration:   i,
			Description: currentDesc,
			TrainPassed: trainReport.Summary.Passed,
			TrainTotal:  trainReport.Summary.Total,
		}

		// If all train queries pass, evaluate test set.
		if trainReport.Summary.Failed == 0 {
			if len(split.Test) > 0 {
				testCfg := cfg.RunConfig
				testCfg.Description = currentDesc
				testReport, err := runEvalFn(ctx, split.Test, testCfg)
				if err != nil {
					return nil, fmt.Errorf("eval loop iteration %d test: %w", i, err)
				}
				tp := testReport.Summary.Passed
				tt := testReport.Summary.Total
				entry.TestPassed = &tp
				entry.TestTotal = &tt

				score := float64(tp) / float64(tt)
				if score > bestTestScore {
					bestTestScore = score
					bestIdx = len(history) // index before append
				}
			}

			history = append(history, entry)
			break
		}

		history = append(history, entry)

		// Gather failed results for improvement.
		var failed []EvalResult
		for _, r := range trainReport.Results {
			if !r.Pass {
				failed = append(failed, r)
			}
		}

		// Call improve engine.
		out, err := improver(ctx, ImproveInput{
			CurrentDescription: currentDesc,
			FailedResults:      failed,
			History:            history,
			MaxDescriptionLen:  cfg.MaxDescriptionLen,
		})
		if err != nil {
			return nil, fmt.Errorf("eval loop iteration %d improve: %w", i, err)
		}
		currentDesc = out.NewDescription
	}

	// Select best iteration by test score. Fall back to last iteration.
	result := &LoopResult{
		Iterations: len(history),
		History:    history,
	}

	if bestIdx >= 0 {
		result.BestDescription = history[bestIdx].Description
		result.BestTestScore = bestTestScore
	} else {
		result.BestDescription = history[len(history)-1].Description
		result.BestTestScore = 0
	}

	// Run final eval on test set with best description.
	if len(split.Test) > 0 {
		finalCfg := cfg.RunConfig
		finalCfg.Description = result.BestDescription
		finalReport, err := runEvalFn(ctx, split.Test, finalCfg)
		if err != nil {
			return nil, fmt.Errorf("eval loop final eval: %w", err)
		}
		result.FinalReport = finalReport
	}

	return result, nil
}
