package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/AdeptMind/infra-tool/claude-cli/internal/claude"
	"github.com/AdeptMind/infra-tool/claude-cli/internal/grading"
)

var gradeSkillCmd = &cobra.Command{
	Use:   "grade <skill-dir>",
	Short: "Grade skill output against grading.json assertions",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		skillDir := args[0]

		assertions, err := grading.ParseGradingFile(filepath.Join(skillDir, "grading.json"))
		if err != nil {
			return err
		}

		outputPath, _ := cmd.Flags().GetString("output-file")
		outputText, err := os.ReadFile(outputPath)
		if err != nil {
			return fmt.Errorf("reading output file: %w", err)
		}

		model, _ := cmd.Flags().GetString("model")
		var opts []claude.Option
		if model != "" {
			opts = append(opts, claude.WithModel(model))
		}

		input := grading.GradingInput{
			Output:     string(outputText),
			Assertions: assertions,
		}

		report, err := grading.GradeOutput(context.Background(), input, claude.RunPrompt, opts...)
		if err != nil {
			return err
		}

		out, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return fmt.Errorf("marshaling report: %w", err)
		}
		fmt.Println(string(out))

		if report.Summary.Failed > 0 {
			os.Exit(1)
		}
		return nil
	},
}

func init() {
	gradeSkillCmd.Flags().String("output-file", "", "Path to the output file to grade (required)")
	gradeSkillCmd.Flags().String("model", "", "Model to use for grading")
	_ = gradeSkillCmd.MarkFlagRequired("output-file")
	skillCmd.AddCommand(gradeSkillCmd)
}
