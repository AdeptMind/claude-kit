package main

import "github.com/spf13/cobra"

// agentCmd is the parent command for single-agent operations.
// Subcommands are registered by their respective files.
var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Manage agents — validate, package",
}
