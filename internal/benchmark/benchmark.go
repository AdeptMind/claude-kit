package benchmark

import (
	"encoding/json"
	"fmt"
	"math"
	"os"

	"github.com/AdeptMind/infra-tool/claude-cli/internal/eval"
)

// RunStats holds pass rates from a single eval run.
type RunStats struct {
	PassRate    float64 `json:"pass_rate"`
	TimeSeconds float64 `json:"time_seconds,omitempty"`
	Tokens      int     `json:"tokens,omitempty"`
}

// Aggregate computes mean and stddev for a set of runs.
func Aggregate(runs []RunStats) eval.StatPair {
	if len(runs) == 0 {
		return eval.StatPair{}
	}
	if len(runs) == 1 {
		return eval.StatPair{Mean: runs[0].PassRate, StdDev: 0}
	}

	var sum float64
	for _, r := range runs {
		sum += r.PassRate
	}
	mean := sum / float64(len(runs))

	var variance float64
	for _, r := range runs {
		diff := r.PassRate - mean
		variance += diff * diff
	}
	variance /= float64(len(runs))
	stddev := math.Sqrt(variance)

	return eval.StatPair{Mean: mean, StdDev: stddev}
}

// ComputeDelta returns the delta between two StatPairs (a - b).
func ComputeDelta(a, b eval.StatPair) eval.StatPair {
	return eval.StatPair{
		Mean:   a.Mean - b.Mean,
		StdDev: math.Sqrt(a.StdDev*a.StdDev + b.StdDev*b.StdDev),
	}
}

// ComputeBenchmark produces BenchmarkStats from with-skill and without-skill runs.
func ComputeBenchmark(withSkill, withoutSkill []RunStats) eval.BenchmarkStats {
	withAgg := Aggregate(withSkill)
	withoutAgg := Aggregate(withoutSkill)
	delta := ComputeDelta(withAgg, withoutAgg)

	return eval.BenchmarkStats{
		PassRate: withAgg,
		Delta:    delta,
	}
}

// LoadRunStats reads a JSON file containing an array of RunStats.
func LoadRunStats(path string) ([]RunStats, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	var runs []RunStats
	if err := json.Unmarshal(data, &runs); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	return runs, nil
}
