package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/spf13/cobra"

	"github.com/AdeptMind/infra-tool/claude-cli/internal/eval"
	"github.com/AdeptMind/infra-tool/claude-cli/internal/report"
)

var reportSkillCmd = &cobra.Command{
	Use:   "report <results.json>",
	Short: "Generate an HTML report from eval results",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		resultsPath := args[0]

		data, err := os.ReadFile(resultsPath)
		if err != nil {
			return fmt.Errorf("reading results: %w", err)
		}

		var current eval.EvalReport
		if err := json.Unmarshal(data, &current); err != nil {
			return fmt.Errorf("parsing results: %w", err)
		}

		reportData := report.ReportData{
			SkillName:   current.SkillName,
			GeneratedAt: time.Now().UTC().Format(time.RFC3339),
			Current:     &current,
		}

		// Load previous results for comparison if provided
		previousPath, _ := cmd.Flags().GetString("previous")
		if previousPath != "" {
			prevData, err := os.ReadFile(previousPath)
			if err == nil {
				var prev eval.EvalReport
				if json.Unmarshal(prevData, &prev) == nil {
					reportData.Previous = &prev
				}
			}
		}

		html := report.GenerateHTML(reportData)

		outputPath, _ := cmd.Flags().GetString("output")
		if outputPath != "" {
			if err := os.WriteFile(outputPath, []byte(html), 0o644); err != nil {
				return fmt.Errorf("writing report: %w", err)
			}
			fmt.Printf("Report written to %s\n", outputPath)

			openFlag, _ := cmd.Flags().GetBool("open")
			if openFlag {
				openBrowser(outputPath)
			}
		} else {
			fmt.Print(html)
		}

		return nil
	},
}

// openBrowser opens the given path in the default browser.
func openBrowser(path string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", path)
	case "linux":
		cmd = exec.Command("xdg-open", path)
	default:
		cmd = exec.Command("open", path)
	}
	_ = cmd.Start()
}

func init() {
	reportSkillCmd.Flags().StringP("output", "o", "", "Write HTML to file instead of stdout")
	reportSkillCmd.Flags().String("previous", "", "Path to previous results.json for comparison")
	reportSkillCmd.Flags().Bool("open", false, "Open the report in a browser (requires --output)")
	skillCmd.AddCommand(reportSkillCmd)
}
