package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/AdeptMind/infra-tool/claude-cli/internal/claude"
	"github.com/AdeptMind/infra-tool/claude-cli/internal/eval"
	"github.com/AdeptMind/infra-tool/claude-cli/internal/improve"
	"github.com/AdeptMind/infra-tool/claude-cli/internal/report"
)

var optimizeSkillCmd = &cobra.Command{
	Use:   "optimize <skill-dir>",
	Short: "Run eval-improve loop to optimize a skill description",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		skillDir := args[0]

		queries, err := eval.ParseEvalSet(filepath.Join(skillDir, "evals.json"))
		if err != nil {
			return err
		}

		skillPath := filepath.Join(skillDir, "SKILL.md")
		name, description := readSkillMeta(skillPath)

		skillContent, err := os.ReadFile(skillPath)
		if err != nil {
			return fmt.Errorf("reading SKILL.md: %w", err)
		}

		model, _ := cmd.Flags().GetString("model")
		workers, _ := cmd.Flags().GetInt("workers")
		runs, _ := cmd.Flags().GetInt("runs")
		threshold, _ := cmd.Flags().GetFloat64("threshold")
		maxIterations, _ := cmd.Flags().GetInt("max-iterations")
		trainRatio, _ := cmd.Flags().GetFloat64("train-ratio")
		reportPath, _ := cmd.Flags().GetString("report")

		runCfg := eval.RunConfig{
			NumWorkers:       workers,
			RunsPerQuery:     runs,
			TriggerThreshold: threshold,
			SkillName:        name,
			Description:      description,
			ProjectRoot:      resolveTarget(),
			Model:            model,
		}

		loopCfg := eval.LoopConfig{
			MaxIterations: maxIterations,
			TrainRatio:    trainRatio,
			RunConfig:     runCfg,
		}

		content := string(skillContent)
		improver := newImprover(model, content)

		result, err := eval.RunLoop(context.Background(), queries, description, improver, loopCfg)
		if err != nil {
			return err
		}

		out, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return fmt.Errorf("marshaling result: %w", err)
		}
		fmt.Println(string(out))

		if reportPath != "" {
			data := report.ReportData{
				SkillName:   name,
				GeneratedAt: time.Now().UTC().Format(time.RFC3339),
				Current:     result.FinalReport,
				History:     result.History,
			}
			html := report.GenerateHTML(data)
			if err := os.WriteFile(reportPath, []byte(html), 0644); err != nil {
				return fmt.Errorf("writing report: %w", err)
			}
		}

		return nil
	},
}

// newImprover creates a DescriptionImprover that bridges eval.ImproveInput
// to improve.ImproveDescription.
func newImprover(model, skillContent string) eval.DescriptionImprover {
	return func(ctx context.Context, input eval.ImproveInput) (*eval.ImproveResult, error) {
		var opts []claude.Option
		if model != "" {
			opts = append(opts, claude.WithModel(model))
		}

		impInput := improve.ImproveInput{
			CurrentDescription: input.CurrentDescription,
			SkillContent:       skillContent,
			FailedResults:      input.FailedResults,
			History:            input.History,
			MaxDescriptionLen:  input.MaxDescriptionLen,
		}

		out, err := improve.ImproveDescription(ctx, claude.RunPrompt, impInput, opts...)
		if err != nil {
			return nil, err
		}
		return &eval.ImproveResult{NewDescription: out.NewDescription}, nil
	}
}

func init() {
	optimizeSkillCmd.Flags().String("model", "", "Model to use for eval and improve runs")
	optimizeSkillCmd.Flags().Int("workers", 10, "Number of parallel workers")
	optimizeSkillCmd.Flags().Int("runs", 3, "Number of runs per query")
	optimizeSkillCmd.Flags().Float64("threshold", 0.5, "Trigger threshold (0.0-1.0)")
	optimizeSkillCmd.Flags().Int("max-iterations", 10, "Maximum eval-improve iterations")
	optimizeSkillCmd.Flags().Float64("train-ratio", 0.7, "Train/test split ratio (0.0-1.0)")
	optimizeSkillCmd.Flags().String("report", "", "Output path for HTML report")
	skillCmd.AddCommand(optimizeSkillCmd)
}
