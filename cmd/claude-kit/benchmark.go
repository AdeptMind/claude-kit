package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/AdeptMind/infra-tool/claude-cli/internal/benchmark"
	"github.com/AdeptMind/infra-tool/claude-cli/internal/eval"
)

var benchmarkSkillCmd = &cobra.Command{
	Use:   "benchmark <results-dir>",
	Short: "Compare eval results with and without a skill",
	Long:  "Reads with_skill/ and without_skill/ subdirectories containing run stats, computes benchmark.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		resultsDir := args[0]

		withDir := filepath.Join(resultsDir, "with_skill")
		withoutDir := filepath.Join(resultsDir, "without_skill")

		withRuns, err := loadRunsFromDir(withDir)
		if err != nil {
			return fmt.Errorf("loading with_skill runs: %w", err)
		}

		withoutRuns, err := loadRunsFromDir(withoutDir)
		if err != nil {
			// Single-dir mode: only with_skill is required
			withoutRuns = nil
		}

		stats := benchmark.ComputeBenchmark(withRuns, withoutRuns)

		outputPath, _ := cmd.Flags().GetString("output")
		data, err := json.MarshalIndent(stats, "", "  ")
		if err != nil {
			return fmt.Errorf("marshaling benchmark: %w", err)
		}

		if outputPath != "" {
			if err := os.WriteFile(outputPath, data, 0o644); err != nil {
				return fmt.Errorf("writing benchmark: %w", err)
			}
			fmt.Printf("Benchmark written to %s\n", outputPath)
		} else {
			fmt.Println(string(data))
		}

		return nil
	},
}

// loadRunsFromDir reads all .json files in a directory as RunStats arrays.
func loadRunsFromDir(dir string) ([]benchmark.RunStats, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var allRuns []benchmark.RunStats
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		runs, err := benchmark.LoadRunStats(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, err
		}
		allRuns = append(allRuns, runs...)
	}

	if len(allRuns) == 0 {
		// Try loading a single results.json from eval output format
		resultPath := filepath.Join(filepath.Dir(dir), "results.json")
		if data, err := os.ReadFile(resultPath); err == nil {
			var report eval.EvalReport
			if err := json.Unmarshal(data, &report); err == nil {
				passRate := 0.0
				if report.Summary.Total > 0 {
					passRate = float64(report.Summary.Passed) / float64(report.Summary.Total)
				}
				allRuns = append(allRuns, benchmark.RunStats{PassRate: passRate})
			}
		}
	}

	return allRuns, nil
}

func init() {
	benchmarkSkillCmd.Flags().String("output", "", "Write benchmark.json to this path instead of stdout")
	skillCmd.AddCommand(benchmarkSkillCmd)
}
