package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/AdeptMind/infra-tool/claude-cli/internal/packaging"
)

var (
	packageOutput         string
	packageSkipValidation bool
)

var packageSkillCmd = &cobra.Command{
	Use:   "package <skill-dir>",
	Short: "Package a skill directory into a .skill archive",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := packaging.PackageOptions{SkipValidation: packageSkipValidation}
		if err := packaging.PackageSkill(args[0], packageOutput, opts); err != nil {
			return err
		}
		fmt.Printf("Packaged skill to %s\n", packageOutput)
		return nil
	},
}

var packageAgentCmd = &cobra.Command{
	Use:   "package <agent-file>",
	Short: "Package an agent file into a .agent archive",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := packaging.PackageOptions{SkipValidation: packageSkipValidation}
		if err := packaging.PackageAgent(args[0], packageOutput, opts); err != nil {
			return err
		}
		fmt.Printf("Packaged agent to %s\n", packageOutput)
		return nil
	},
}

var packageTemplateCmd = &cobra.Command{
	Use:   "package <template-dir>",
	Short: "Package a template directory into a .claude-kit archive",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := packaging.PackageOptions{}
		if err := packaging.PackageTemplate(args[0], packageOutput, opts); err != nil {
			return err
		}
		fmt.Printf("Packaged template to %s\n", packageOutput)
		return nil
	},
}

func init() {
	packageSkillCmd.Flags().StringVarP(&packageOutput, "output", "o", ".", "Output directory for the archive")
	packageSkillCmd.Flags().BoolVar(&packageSkipValidation, "skip-validation", false, "Skip validation before packaging")
	skillCmd.AddCommand(packageSkillCmd)

	packageAgentCmd.Flags().StringVarP(&packageOutput, "output", "o", ".", "Output directory for the archive")
	packageAgentCmd.Flags().BoolVar(&packageSkipValidation, "skip-validation", false, "Skip validation before packaging")
	agentCmd.AddCommand(packageAgentCmd)

	packageTemplateCmd.Flags().StringVarP(&packageOutput, "output", "o", ".", "Output directory for the archive")
	rootCmd.AddCommand(packageTemplateCmd)
}
