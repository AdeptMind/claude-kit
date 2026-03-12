package main

import "github.com/spf13/cobra"

// agentsCmd is the parent command for multi-agent operations (registry, etc).
// Subcommands are registered by their respective files.
var agentsCmd = &cobra.Command{
	Use:   "agents",
	Short: "Multi-agent operations — registry, collaboration maps",
}
