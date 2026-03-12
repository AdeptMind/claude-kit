package main

import "github.com/spf13/cobra"

// skillCmd is the parent command for all skill subcommands.
// Subcommands are registered by their respective files.
var skillCmd = &cobra.Command{
	Use:   "skill",
	Short: "Manage skills — validate, eval, grade, benchmark, optimize, package, report",
}
