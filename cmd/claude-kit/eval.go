package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/AdeptMind/infra-tool/claude-cli/internal/eval"
)

var evalSkillCmd = &cobra.Command{
	Use:   "eval <skill-dir>",
	Short: "Run trigger evals for a skill using its evals.json",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		skillDir := args[0]

		queries, err := eval.ParseEvalSet(filepath.Join(skillDir, "evals.json"))
		if err != nil {
			return err
		}

		name, description := readSkillMeta(filepath.Join(skillDir, "SKILL.md"))

		model, _ := cmd.Flags().GetString("model")
		workers, _ := cmd.Flags().GetInt("workers")
		runs, _ := cmd.Flags().GetInt("runs")
		threshold, _ := cmd.Flags().GetFloat64("threshold")

		cfg := eval.RunConfig{
			NumWorkers:       workers,
			RunsPerQuery:     runs,
			TriggerThreshold: threshold,
			SkillName:        name,
			Description:      description,
			ProjectRoot:      resolveTarget(),
			Model:            model,
		}

		report, err := eval.RunEval(context.Background(), queries, cfg)
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

// readSkillMeta extracts name and description from a SKILL.md frontmatter.
// Returns the directory basename as fallback name if frontmatter is missing.
func readSkillMeta(path string) (name, description string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return filepath.Base(filepath.Dir(path)), ""
	}

	content := string(data)
	if !strings.HasPrefix(strings.TrimSpace(content), "---") {
		return filepath.Base(filepath.Dir(path)), ""
	}

	fields := make(map[string]string)
	inFrontmatter := false
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "---" {
			if inFrontmatter {
				break
			}
			inFrontmatter = true
			continue
		}
		if !inFrontmatter {
			continue
		}
		if len(line) > 0 && (line[0] == ' ' || line[0] == '\t' || line[0] == '-') {
			continue
		}
		idx := strings.Index(line, ":")
		if idx < 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])
		if len(val) >= 2 && ((val[0] == '"' && val[len(val)-1] == '"') || (val[0] == '\'' && val[len(val)-1] == '\'')) {
			val = val[1 : len(val)-1]
		}
		fields[key] = val
	}

	name = fields["name"]
	if name == "" {
		name = filepath.Base(filepath.Dir(path))
	}
	description = fields["description"]
	return name, description
}

func init() {
	evalSkillCmd.Flags().String("model", "", "Model to use for eval runs")
	evalSkillCmd.Flags().Int("workers", 10, "Number of parallel workers")
	evalSkillCmd.Flags().Int("runs", 3, "Number of runs per query")
	evalSkillCmd.Flags().Float64("threshold", 0.5, "Trigger threshold (0.0-1.0)")
	skillCmd.AddCommand(evalSkillCmd)
}
