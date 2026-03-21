package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/AdeptMind/infra-tool/claude-cli/internal/policy"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var policyCmd = &cobra.Command{
	Use:   "policy",
	Short: "Manage Claude Code policy configuration",
}

var policyInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize policy.yaml with an interactive profile selector",
	RunE: func(cmd *cobra.Command, args []string) error {
		force, _ := cmd.Flags().GetBool("force")
		return runPolicyInit(force)
	},
}

var policyApplyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply policy.yaml to settings.json (shows diff, asks confirmation)",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPolicyApply()
	},
}

var policyShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show resolved policy as YAML",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPolicyShow()
	},
}

func init() {
	policyInitCmd.Flags().Bool("force", false, "Overwrite existing policy.yaml")
	policyCmd.AddCommand(policyInitCmd)
	policyCmd.AddCommand(policyApplyCmd)
	policyCmd.AddCommand(policyShowCmd)
}

func runPolicyInit(force bool) error {
	targetDir := resolveTarget()
	policyPath := filepath.Join(targetDir, "policy.yaml")

	// Check if policy.yaml already exists
	if _, err := os.Stat(policyPath); err == nil && !force {
		return fmt.Errorf("policy.yaml already exists at %s — use --force to overwrite", policyPath)
	}

	fmt.Println()
	fmt.Println(subtitleStyle.Render("  Policy Setup"))
	fmt.Println(dimStyle.Render(fmt.Sprintf("  Target: %s", policyPath)))
	fmt.Println()

	// Profile selection
	profiles := []struct {
		Name string
		Desc string
	}{
		{"permissive", "Minimal restrictions — open network, no sandbox, no audit"},
		{"moderate", "Balanced — git hook guard, audit logging, broader file deny list"},
		{"strict", "Locked down — network blocked, sandbox enabled, full audit"},
	}

	options := make([]huh.Option[string], len(profiles))
	for i, p := range profiles {
		options[i] = huh.NewOption(fmt.Sprintf("%s -- %s", p.Name, p.Desc), p.Name)
	}

	var selected string
	selectForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select a policy profile").
				Options(options...).
				Value(&selected),
		),
	).WithTheme(ckTheme())

	if err := selectForm.Run(); err != nil {
		return err
	}

	// Read the selected profile YAML
	profilePath := filepath.Join(resolveTemplateDir(), "policies", selected+".yaml")
	content, err := os.ReadFile(profilePath)
	if err != nil {
		return fmt.Errorf("reading profile %s: %w", selected, err)
	}

	// Display preview
	fmt.Println()
	fmt.Println(subtitleStyle.Render(fmt.Sprintf("  Preview: %s", selected)))
	fmt.Println()

	previewBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(rose).
		Padding(0, 2).
		MarginLeft(2)
	fmt.Println(previewBox.Render(string(content)))
	fmt.Println()

	// Confirm
	var confirm bool
	confirmForm := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Write this policy?").
				Value(&confirm),
		),
	).WithTheme(ckTheme())

	if err := confirmForm.Run(); err != nil {
		return err
	}
	if !confirm {
		fmt.Println("Aborted.")
		return nil
	}

	// Ensure target directory exists
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return fmt.Errorf("creating target directory: %w", err)
	}

	// Write policy.yaml
	if err := os.WriteFile(policyPath, content, 0o644); err != nil {
		return fmt.Errorf("writing policy.yaml: %w", err)
	}

	fmt.Println(fmt.Sprintf("  %s %s %s",
		checkMark,
		accentStyle.Render("policy.yaml"),
		dimStyle.Render(fmt.Sprintf("(%s profile)", selected)),
	))
	fmt.Println()
	fmt.Println(successStyle.Render(fmt.Sprintf("  %s Policy initialized!", arrow)))

	return nil
}

func loadResolvedPolicy() (*policy.PolicySpec, string, error) {
	targetDir := resolveTarget()
	policyPath := filepath.Join(targetDir, "policy.yaml")

	if _, err := os.Stat(policyPath); os.IsNotExist(err) {
		return nil, "", fmt.Errorf("policy.yaml not found at %s — run 'ck policy init' first", policyPath)
	}

	spec, err := policy.ParseFile(policyPath)
	if err != nil {
		return nil, "", err
	}

	profilesDir := filepath.Join(resolveTemplateDir(), "policies")
	resolved, err := policy.Resolve(spec, profilesDir)
	if err != nil {
		return nil, "", err
	}

	return resolved, targetDir, nil
}

func runPolicyApply() error {
	resolved, targetDir, err := loadResolvedPolicy()
	if err != nil {
		return err
	}

	settingsPath := filepath.Join(targetDir, "settings.json")

	diff, err := policy.MergeSettings(settingsPath, resolved)
	if err != nil {
		return err
	}

	if diff == "" {
		fmt.Println()
		fmt.Println(successStyle.Render(fmt.Sprintf("  %s settings.json is already up to date", checkMark)))
		return nil
	}

	fmt.Println()
	fmt.Println(subtitleStyle.Render("  Policy Apply — Diff"))
	fmt.Println()

	diffBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(rose).
		Padding(0, 2).
		MarginLeft(2)
	fmt.Println(diffBox.Render(diff))
	fmt.Println()

	var confirm bool
	confirmForm := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Apply these changes to settings.json?").
				Value(&confirm),
		),
	).WithTheme(ckTheme())

	if err := confirmForm.Run(); err != nil {
		return err
	}
	if !confirm {
		fmt.Println("Aborted.")
		return nil
	}

	if err := policy.ApplySettings(settingsPath, resolved); err != nil {
		return err
	}

	fmt.Println(fmt.Sprintf("  %s %s",
		checkMark,
		accentStyle.Render("settings.json updated"),
	))
	fmt.Println()
	fmt.Println(successStyle.Render(fmt.Sprintf("  %s Policy applied!", arrow)))

	return nil
}

func runPolicyShow() error {
	resolved, _, err := loadResolvedPolicy()
	if err != nil {
		return err
	}

	yamlBytes, err := policy.SortedYAML(resolved)
	if err != nil {
		return fmt.Errorf("marshalling resolved policy: %w", err)
	}

	fmt.Print(string(yamlBytes))
	return nil
}
