package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/AdeptMind/infra-tool/claude-cli/internal/benchmark"
	"github.com/AdeptMind/infra-tool/claude-cli/internal/bmadeval"
	"github.com/AdeptMind/infra-tool/claude-cli/internal/claude"
	"github.com/AdeptMind/infra-tool/claude-cli/internal/eval"
	"github.com/AdeptMind/infra-tool/claude-cli/internal/grading"
)

// PhaseResult holds grading results for a single BMAD phase.
type PhaseResult struct {
	Phase   string                `json:"phase"`
	Results []eval.GradingResult  `json:"results"`
	Summary grading.GradingSummary `json:"summary"`
}

var bmadEvalCmd = &cobra.Command{
	Use:   "eval [output-dir]",
	Short: "Evaluate BMAD output against phase assertions",
	Long:  "Grades .claude/output/ artifacts against break, model, and act assertions.",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		targetDir := resolveTarget()
		outputDir := filepath.Join(targetDir, "output")
		if len(args) > 0 {
			outputDir = args[0]
		}

		evalsDir := filepath.Join(targetDir, "evals")
		phase, _ := cmd.Flags().GetString("phase")
		model, _ := cmd.Flags().GetString("model")

		var opts []claude.Option
		if model != "" {
			opts = append(opts, claude.WithModel(model))
		}

		phases := bmadeval.AvailablePhases()
		if phase != "" {
			phases = []bmadeval.Phase{bmadeval.Phase(phase)}
		}

		var results []PhaseResult
		for _, p := range phases {
			assertions, err := bmadeval.LoadAssertions(evalsDir, p)
			if err != nil {
				// Skip phases with missing assertion files
				continue
			}

			output, err := readPhaseOutput(outputDir, p)
			if err != nil {
				continue
			}

			input := grading.GradingInput{
				Output:     output,
				Assertions: assertions,
			}

			report, err := grading.GradeOutput(context.Background(), input, claude.RunPrompt, opts...)
			if err != nil {
				return fmt.Errorf("grading phase %s: %w", p, err)
			}

			results = append(results, PhaseResult{
				Phase:   string(p),
				Results: report.Results,
				Summary: report.Summary,
			})
		}

		data, err := json.MarshalIndent(results, "", "  ")
		if err != nil {
			return fmt.Errorf("marshaling results: %w", err)
		}
		fmt.Println(string(data))

		// Exit 1 if any phase has failures
		for _, r := range results {
			if r.Summary.Failed > 0 {
				os.Exit(1)
			}
		}
		return nil
	},
}

var bmadBenchmarkCmd = &cobra.Command{
	Use:   "benchmark <run1-dir> <run2-dir>",
	Short: "Compare two BMAD eval runs",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		run1Dir := args[0]
		run2Dir := args[1]

		run1Stats, err := loadBmadRunStats(run1Dir)
		if err != nil {
			return fmt.Errorf("loading run 1: %w", err)
		}

		run2Stats, err := loadBmadRunStats(run2Dir)
		if err != nil {
			return fmt.Errorf("loading run 2: %w", err)
		}

		stats := benchmark.ComputeBenchmark(run1Stats, run2Stats)

		data, err := json.MarshalIndent(stats, "", "  ")
		if err != nil {
			return fmt.Errorf("marshaling benchmark: %w", err)
		}
		fmt.Println(string(data))
		return nil
	},
}

// readPhaseOutput reads the relevant output file for a BMAD phase.
func readPhaseOutput(outputDir string, phase bmadeval.Phase) (string, error) {
	candidates := map[bmadeval.Phase][]string{
		bmadeval.PhaseBreak: {"problem.yaml"},
		bmadeval.PhaseModel: {"architecture.yaml", "backlog.yaml"},
		bmadeval.PhaseAct:   {"act-report.md"},
	}

	files, ok := candidates[phase]
	if !ok {
		return "", fmt.Errorf("unknown phase %s", phase)
	}

	var combined string
	for _, f := range files {
		data, err := os.ReadFile(filepath.Join(outputDir, f))
		if err != nil {
			continue
		}
		combined += fmt.Sprintf("## %s\n%s\n\n", f, string(data))
	}

	if combined == "" {
		return "", fmt.Errorf("no output found for phase %s", phase)
	}
	return combined, nil
}

// loadBmadRunStats extracts pass rates from a BMAD eval results file.
func loadBmadRunStats(dir string) ([]benchmark.RunStats, error) {
	data, err := os.ReadFile(filepath.Join(dir, "results.json"))
	if err != nil {
		return nil, err
	}

	var phases []PhaseResult
	if err := json.Unmarshal(data, &phases); err != nil {
		return nil, fmt.Errorf("parsing results: %w", err)
	}

	var stats []benchmark.RunStats
	for _, p := range phases {
		rate := 0.0
		if p.Summary.Total > 0 {
			rate = float64(p.Summary.Passed) / float64(p.Summary.Total)
		}
		stats = append(stats, benchmark.RunStats{PassRate: rate})
	}
	return stats, nil
}

func init() {
	bmadEvalCmd.Flags().String("phase", "", "Evaluate a specific phase only (break, model, act)")
	bmadEvalCmd.Flags().String("model", "", "Model to use for grading")

	// Register under 'bmad' parent command if it exists, otherwise on root
	bmadCmd := findOrCreateBmadCmd()
	bmadCmd.AddCommand(bmadEvalCmd)
	bmadCmd.AddCommand(bmadBenchmarkCmd)
}

func findOrCreateBmadCmd() *cobra.Command {
	for _, c := range rootCmd.Commands() {
		if c.Name() == "bmad" {
			return c
		}
	}
	cmd := &cobra.Command{
		Use:   "bmad",
		Short: "BMAD workflow evaluation and benchmarking",
	}
	rootCmd.AddCommand(cmd)
	return cmd
}
