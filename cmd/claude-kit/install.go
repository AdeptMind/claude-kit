package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/AdeptMind/infra-tool/claude-cli/internal/packaging"
)

var installForce bool

var installCmd = &cobra.Command{
	Use:   "install <archive-path>",
	Short: "Install a .skill, .agent, or .claude-kit archive into .claude/",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		targetDir := resolveTarget()

		result, err := packaging.Install(args[0], targetDir, installForce)
		if err != nil {
			var ce *packaging.ConflictError
			if errors.As(err, &ce) {
				fmt.Fprintln(os.Stderr, "Conflicts detected (use --force to overwrite):")
				for _, c := range ce.Conflicts {
					fmt.Fprintf(os.Stderr, "  %s\n", c)
				}
				os.Exit(1)
			}
			return err
		}

		fmt.Printf("Installed %s %q (v%s)\n", result.Type, result.Name, result.Version)
		for _, f := range result.FilesCreated {
			fmt.Printf("  %s\n", f)
		}
		return nil
	},
}

func init() {
	installCmd.Flags().BoolVar(&installForce, "force", false, "Overwrite existing files")
	rootCmd.AddCommand(installCmd)
}
