package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/AdeptMind/infra-tool/claude-cli/internal/validate"
)

var validateSkillCmd = &cobra.Command{
	Use:   "validate <path>",
	Short: "Validate a skill's structure and frontmatter",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		errs := validate.ValidateSkill(args[0])
		if len(errs) == 0 {
			fmt.Println("Valid.")
			return nil
		}
		for _, e := range errs {
			fmt.Fprintln(os.Stderr, e.Error())
		}
		os.Exit(1)
		return nil
	},
}

var validateAgentCmd = &cobra.Command{
	Use:   "validate <path>",
	Short: "Validate an agent file's frontmatter, skill refs, and tools",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		skillsDir := resolveTarget() + "/skills"
		errs := validate.ValidateAgent(args[0], skillsDir)
		if len(errs) == 0 {
			fmt.Println("Valid.")
			return nil
		}
		for _, e := range errs {
			fmt.Fprintln(os.Stderr, e.Error())
		}
		os.Exit(1)
		return nil
	},
}

func init() {
	skillCmd.AddCommand(validateSkillCmd)
	agentCmd.AddCommand(validateAgentCmd)
}
