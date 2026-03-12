package benchmark

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/AdeptMind/infra-tool/claude-cli/internal/eval"
)

func TestAggregate_MultipleRuns(t *testing.T) {
	runs := []RunStats{
		{PassRate: 0.8},
		{PassRate: 0.6},
		{PassRate: 1.0},
	}
	result := Aggregate(runs)

	expectedMean := 0.8
	if math.Abs(result.Mean-expectedMean) > 0.001 {
		t.Errorf("mean = %f, want %f", result.Mean, expectedMean)
	}
	if result.StdDev < 0.01 {
		t.Errorf("stddev should be > 0 for varying runs, got %f", result.StdDev)
	}
}

func TestAggregate_SingleRun(t *testing.T) {
	runs := []RunStats{{PassRate: 0.75}}
	result := Aggregate(runs)

	if result.Mean != 0.75 {
		t.Errorf("mean = %f, want 0.75", result.Mean)
	}
	if result.StdDev != 0 {
		t.Errorf("stddev = %f, want 0 for single run", result.StdDev)
	}
}

func TestAggregate_Empty(t *testing.T) {
	result := Aggregate(nil)
	if result.Mean != 0 || result.StdDev != 0 {
		t.Errorf("expected zero StatPair, got %+v", result)
	}
}

func TestComputeDelta(t *testing.T) {
	a := eval.StatPair{Mean: 0.9, StdDev: 0.1}
	b := eval.StatPair{Mean: 0.7, StdDev: 0.15}
	delta := ComputeDelta(a, b)

	expectedMean := 0.2
	if math.Abs(delta.Mean-expectedMean) > 0.001 {
		t.Errorf("delta mean = %f, want %f", delta.Mean, expectedMean)
	}
	// Combined stddev via quadrature
	expectedStd := math.Sqrt(0.01 + 0.0225)
	if math.Abs(delta.StdDev-expectedStd) > 0.001 {
		t.Errorf("delta stddev = %f, want %f", delta.StdDev, expectedStd)
	}
}

func TestComputeBenchmark(t *testing.T) {
	with := []RunStats{{PassRate: 0.9}, {PassRate: 0.8}}
	without := []RunStats{{PassRate: 0.5}, {PassRate: 0.6}}

	stats := ComputeBenchmark(with, without)

	if stats.PassRate.Mean < 0.8 {
		t.Errorf("pass_rate mean = %f, want >= 0.8", stats.PassRate.Mean)
	}
	if stats.Delta.Mean <= 0 {
		t.Errorf("delta mean = %f, want > 0 (with > without)", stats.Delta.Mean)
	}
}

func TestLoadRunStats(t *testing.T) {
	dir := t.TempDir()
	data := []RunStats{
		{PassRate: 0.8, TimeSeconds: 1.2, Tokens: 100},
		{PassRate: 0.6, TimeSeconds: 2.0, Tokens: 150},
	}
	bytes, _ := json.Marshal(data)
	path := filepath.Join(dir, "runs.json")
	os.WriteFile(path, bytes, 0o644)

	loaded, err := LoadRunStats(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(loaded) != 2 {
		t.Fatalf("got %d runs, want 2", len(loaded))
	}
	if loaded[0].PassRate != 0.8 {
		t.Errorf("first run pass_rate = %f, want 0.8", loaded[0].PassRate)
	}
}

func TestLoadRunStats_FileNotFound(t *testing.T) {
	_, err := LoadRunStats("/nonexistent/path.json")
	if err == nil {
		t.Error("expected error for missing file")
	}
}
