package main

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/AdeptMind/infra-tool/claude-cli/internal/registry"
)

var registryUpdateFlag bool

var registryCmd = &cobra.Command{
	Use:   "registry",
	Short: "Generate agent-registry.yaml from all agents",
	RunE: func(cmd *cobra.Command, args []string) error {
		targetDir := resolveTarget()
		agentsDir := filepath.Join(targetDir, "agents")

		agents, err := registry.ScanAgents(agentsDir)
		if err != nil {
			return fmt.Errorf("scanning agents: %w", err)
		}

		outputPath := filepath.Join(targetDir, "agent-registry.yaml")
		if err := registry.WriteRegistry(outputPath, agents); err != nil {
			return fmt.Errorf("writing registry: %w", err)
		}

		fmt.Printf("Generated %s (%d agents)\n", outputPath, len(agents))

		if registryUpdateFlag {
			maps := registry.BuildCollaborationMaps(agents)
			if err := registry.InjectCollaborationMaps(agentsDir, maps); err != nil {
				return fmt.Errorf("injecting collaboration maps: %w", err)
			}
			count := 0
			for _, entries := range maps {
				if len(entries) > 0 {
					count++
				}
			}
			fmt.Printf("Updated collaboration maps for %d agents\n", count)
		}

		return nil
	},
}

func init() {
	registryCmd.Flags().BoolVar(&registryUpdateFlag, "update", false, "Also inject collaboration maps into agent files")
	agentsCmd.AddCommand(registryCmd)
}
