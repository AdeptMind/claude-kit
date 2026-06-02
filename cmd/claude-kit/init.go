package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"github.com/AdeptMind/infra-tool/claude-cli/internal/catalog"
	"github.com/AdeptMind/infra-tool/claude-cli/internal/policy"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Interactive setup — pick agents, everything else is automatic",
	Long: `Initialize a project with an interactive agent picker.

Pick the agents you need and Claude Kit automatically installs
all related skills, commands, and rules.

Only agents not yet installed are shown. Use 'ck remove' to
remove installed agents and their dependencies.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runInteractiveInit()
	},
}

func runInteractiveInit() error {
	tmplDir := resolveTemplateDir()
	targetDir := resolveTarget()

	fmt.Println(banner())

	// Scan available components
	categories, err := catalog.ScanTemplate(tmplDir)
	if err != nil {
		return fmt.Errorf("scanning templates: %w", err)
	}

	if len(categories) == 0 {
		return fmt.Errorf("no components found in template directory: %s", tmplDir)
	}

	// Find the agents category
	var agentComps []catalog.Component
	for _, cat := range categories {
		if cat.Name == "agents" {
			agentComps = cat.Components
			break
		}
	}

	if len(agentComps) == 0 {
		return fmt.Errorf("no agents found in template directory")
	}

	// Check for existing installation
	isExisting := false
	installedAgents := make(map[string]bool)
	if _, err := os.Stat(targetDir); err == nil {
		isExisting = true
		installed, _ := catalog.GetInstalled(targetDir)
		for _, cat := range installed {
			if cat.Name == "agents" {
				for _, c := range cat.Components {
					installedAgents[c.Name] = true
				}
			}
		}
	}

	if isExisting {
		fmt.Println(subtitleStyle.Render("  Project Setup (existing .claude/ detected)"))
		if len(installedAgents) > 0 {
			fmt.Println(dimStyle.Render(fmt.Sprintf("  %d agents already installed", len(installedAgents))))
		}
	} else {
		fmt.Println(subtitleStyle.Render("  Project Setup"))
	}
	fmt.Println(dimStyle.Render(fmt.Sprintf("  Template: %s", tmplDir)))
	fmt.Println(dimStyle.Render(fmt.Sprintf("  Target:   %s", targetDir)))
	fmt.Println()

	// Step 1: Ask which methodology to install (skip if BMAD core agents already installed)
	methodology := "skip" // "bmad" | "hybrid" | "skip"
	bmadAlreadyInstalled := installedAgents["product-owner"] && installedAgents["architect"] && installedAgents["tech-lead"]
	if !bmadAlreadyInstalled {
		methodForm := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Install methodology?").
					Description("Pre-selects core agents (product-owner, architect, tech-lead) + workflow commands.").
					Options(
						huh.NewOption("BMAD+ (custom) -- Break/Model/Act/Deliver + Ralph teammates", "bmad"),
						huh.NewOption("Hybrid SDD/BMAD -- BMAD+ phases as ramp to a declarative spec; drift check after Act", "hybrid"),
						huh.NewOption("Skip -- pick agents manually, no workflow commands", "skip"),
					).
					Value(&methodology),
			),
		).WithTheme(ckTheme())
		if err := methodForm.Run(); err != nil {
			return err
		}
	}
	useBmad := methodology == "bmad" || methodology == "hybrid"

	// Step 2: Agent-only picker — only show agents not yet installed
	// BMAD agents that aren't installed yet get pre-selected
	bmadAgents := map[string]bool{"product-owner": true, "architect": true, "tech-lead": true}

	const allSentinel = "__all__"
	var preselected []string
	agentOptions := make([]huh.Option[string], 0, len(agentComps))
	for _, c := range agentComps {
		if installedAgents[c.Name] {
			continue // skip already installed
		}
		label := agentLabel(c.Name, c.Description)
		agentOptions = append(agentOptions, huh.NewOption(label, c.Name))
		if useBmad && bmadAgents[c.Name] {
			preselected = append(preselected, c.Name)
		}
	}

	if len(agentOptions) == 0 {
		fmt.Println(successStyle.Render(fmt.Sprintf("  %s All agents already installed!", arrow)))
		fmt.Println(dimStyle.Render("  Use 'ck remove' to remove agents."))
		return nil
	}

	// Prepend "All agents" option
	options := make([]huh.Option[string], 0, len(agentOptions)+1)
	options = append(options, huh.NewOption("All agents", allSentinel))
	options = append(options, agentOptions...)

	selectedAgents := preselected
	agentForm := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Select agents to add (skills, commands & rules are automatic)").
				Options(options...).
				Value(&selectedAgents),
		),
	).WithTheme(ckTheme())

	if err := agentForm.Run(); err != nil {
		return err
	}

	// Expand "all" sentinel to all available agents
	for _, s := range selectedAgents {
		if s == allSentinel {
			selectedAgents = make([]string, 0, len(agentOptions))
			for _, opt := range agentOptions {
				selectedAgents = append(selectedAgents, opt.Value)
			}
			break
		}
	}

	if len(selectedAgents) == 0 {
		fmt.Println("No agents selected.")
		return nil
	}

	// Step 2b: Ask for teammate mode if 2+ agents selected
	teammateMode := "auto"
	if len(selectedAgents) >= 2 {
		teammateModeForm := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Teammate display mode").
					Options(
						huh.NewOption("auto -- split panes in tmux, otherwise in-process", "auto"),
						huh.NewOption("in-process -- all teammates in one terminal", "in-process"),
						huh.NewOption("tmux -- each teammate in its own pane (requires tmux/iTerm2)", "tmux"),
					).
					Value(&teammateMode),
			),
		).WithTheme(ckTheme())
		if err := teammateModeForm.Run(); err != nil {
			return err
		}
	}

	// Step 2c: Ask about Ralph worktree isolation (only relevant with BMAD + 2+ agents)
	useWorktrees := false
	if useBmad && len(selectedAgents) >= 2 {
		worktreesForm := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title("Enable Ralph worktree isolation?").
					Description("Each teammate gets an isolated git worktree. Adds a merge phase per round. Default: off.").
					Value(&useWorktrees),
			),
		).WithTheme(ckTheme())
		if err := worktreesForm.Run(); err != nil {
			return err
		}
	}

	// Step 2d: Security policy profile
	policyProfile := "none"
	policyForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Security policy profile").
				Options(
					huh.NewOption("none -- skip policy (no restrictions)", "none"),
					huh.NewOption("permissive -- basic secret deny list, no hooks, no sandbox", "permissive"),
					huh.NewOption("moderate -- broader deny list + git commit/push hook, no sandbox", "moderate"),
					huh.NewOption("strict -- strict deny list + git/network-block hooks + sandbox ready for 'ck sandbox'", "strict"),
				).
				Value(&policyProfile),
		),
	).WithTheme(ckTheme())
	if err := policyForm.Run(); err != nil {
		return err
	}

	// Step 2e: Permission mode
	permissionMode := "default"
	permModeForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Permission mode").
				Options(
					huh.NewOption("default -- ask before every action", "default"),
					huh.NewOption("acceptEdits -- auto-edit files, ask for bash commands", "acceptEdits"),
					huh.NewOption("plan -- read-only exploration, propose plans without editing", "plan"),
					huh.NewOption("bypassPermissions -- no checks (isolated environments only)", "bypassPermissions"),
				).
				Value(&permissionMode),
		),
	).WithTheme(ckTheme())
	if err := permModeForm.Run(); err != nil {
		return err
	}

	// Step 3: Auto-compute all defaults from selected agents
	selectedSet := make(map[string]bool)
	for _, name := range selectedAgents {
		selectedSet[name] = true
	}

	// Compute skills, commands, rules from all selected agents
	skillSet := make(map[string]bool)
	commandSet := make(map[string]bool)
	ruleSet := make(map[string]bool)

	for _, name := range selectedAgents {
		// Skills from frontmatter
		agentPath := filepath.Join(tmplDir, "agents", name+".md")
		for _, dep := range catalog.ExtractSkillDeps(agentPath) {
			skillSet[dep] = true
		}

		// Extra skills
		for _, skill := range guessExtraSkillsForAgent(name) {
			skillSet[skill] = true
		}

		// Rules
		for _, rule := range guessRulesForAgent(name) {
			ruleSet[rule] = true
		}

		// Commands
		for _, cmd := range guessCommandsForAgent(name) {
			commandSet[cmd] = true
		}
	}

	// If BMAD accepted, add BMAD workflow commands
	if useBmad {
		for _, cmd := range []string{
			"bmad-run", "bmad-break", "bmad-model", "bmad-act", "bmad-deliver",
			"principles", "clarify", "analyze", "checklist",
			"ralph", "ralph-loop", "ralph-cancel",
			"r", "p", "c", "g",
			"gsd-prep",
			"bmad-help", "party", "quick-spec", "quick-dev", "shard", "brainstorm", "ux-spec", "create-component",
			"h", "qs", "qd",
		} {
			commandSet[cmd] = true
		}
		// Base rules for BMAD
		for _, rule := range []string{"code-style", "testing", "security", "documentation"} {
			ruleSet[rule] = true
		}
		// Hybrid mode adds the spec-driven discipline rule on top of BMAD
		if methodology == "hybrid" {
			ruleSet["spec-driven"] = true
		}
	}

	// Always add ck-sync command
	commandSet["ck-sync"] = true

	// Convert sets to sorted slices for display
	skills := sortedKeys(skillSet)
	commands := sortedKeys(commandSet)
	rules := sortedKeys(ruleSet)

	// Step 4: Show summary
	fmt.Println()
	fmt.Println(subtitleStyle.Render("  Will install:"))
	fmt.Println(fmt.Sprintf("    %s %s: %s",
		bullet,
		accentStyle.Render(fmt.Sprintf("%d agents", len(selectedAgents))),
		dimStyle.Render(strings.Join(selectedAgents, ", ")),
	))
	fmt.Println(fmt.Sprintf("    %s %s: %s",
		bullet,
		accentStyle.Render(fmt.Sprintf("%d skills", len(skills))),
		dimStyle.Render(strings.Join(skills, ", ")),
	))
	fmt.Println(fmt.Sprintf("    %s %s: %s",
		bullet,
		accentStyle.Render(fmt.Sprintf("%d commands", len(commands))),
		dimStyle.Render(strings.Join(commands, ", ")),
	))
	fmt.Println(fmt.Sprintf("    %s %s: %s",
		bullet,
		accentStyle.Render(fmt.Sprintf("%d rules", len(rules))),
		dimStyle.Render(strings.Join(rules, ", ")),
	))
	if useBmad {
		methodLabel := "BMAD+"
		if methodology == "hybrid" {
			methodLabel = "Hybrid SDD/BMAD (+ spec-driven rule)"
		}
		fmt.Println(fmt.Sprintf("    %s %s: %s",
			bullet,
			accentStyle.Render("methodology"),
			dimStyle.Render(methodLabel),
		))
	}
	if len(selectedAgents) >= 2 {
		fmt.Println(fmt.Sprintf("    %s %s: %s",
			bullet,
			accentStyle.Render("teammate mode"),
			dimStyle.Render(teammateMode),
		))
	}
	if useBmad && len(selectedAgents) >= 2 {
		worktreesLabel := "off (default)"
		if useWorktrees {
			worktreesLabel = "on — worktree per teammate, merge phase per round"
		}
		fmt.Println(fmt.Sprintf("    %s %s: %s",
			bullet,
			accentStyle.Render("ralph worktrees"),
			dimStyle.Render(worktreesLabel),
		))
	}
	if policyProfile != "none" {
		fmt.Println(fmt.Sprintf("    %s %s: %s",
			bullet,
			accentStyle.Render("security policy"),
			dimStyle.Render(policyProfile),
		))
	}
	if permissionMode != "default" {
		fmt.Println(fmt.Sprintf("    %s %s: %s",
			bullet,
			accentStyle.Render("permission mode"),
			dimStyle.Render(permissionMode),
		))
	}

	fmt.Println()

	// Step 5: Confirm
	var confirm bool
	confirmForm := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Apply changes?").
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

	// Install base files
	if err := catalog.CopyBaseFiles(tmplDir, targetDir); err != nil {
		return fmt.Errorf("copying base files: %w", err)
	}
	if err := catalog.PatchSettingsTeammateMode(targetDir, teammateMode); err != nil {
		return fmt.Errorf("patching teammate mode: %w", err)
	}
	if permissionMode != "default" {
		if err := catalog.PatchSettingsPermissionMode(targetDir, permissionMode); err != nil {
			return fmt.Errorf("patching permission mode: %w", err)
		}
	}
	if !isExisting {
		fmt.Println(fmt.Sprintf("  %s %s", checkMark, accentStyle.Render("Installed CLAUDE.md + settings.json")))
	}

	// Install agents
	fmt.Println(sectionHeader("Agents"))
	for _, name := range selectedAgents {
		if err := catalog.CopyComponent(tmplDir, targetDir, "agents", name); err != nil {
			fmt.Fprintln(os.Stderr, errorStyle.Render(fmt.Sprintf("  Agent %s: %v", name, err)))
			continue
		}
		fmt.Println(fmt.Sprintf("  %s %s", checkMark, accentStyle.Render(fmt.Sprintf("agent: %s", name))))
	}

	// Install skills
	fmt.Println(sectionHeader("Skills"))
	for _, name := range skills {
		if err := catalog.CopyComponent(tmplDir, targetDir, "skills", name); err != nil {
			continue
		}
		fmt.Println(fmt.Sprintf("  %s %s", checkMark, infoStyle.Render(fmt.Sprintf("skill: %s", name))))
	}

	// Install commands
	fmt.Println(sectionHeader("Commands"))
	for _, name := range commands {
		if err := catalog.CopyComponent(tmplDir, targetDir, "commands", name); err != nil {
			continue
		}
		fmt.Println(fmt.Sprintf("  %s %s", checkMark, infoStyle.Render(fmt.Sprintf("command: %s", name))))
	}

	// Install rules
	fmt.Println(sectionHeader("Rules"))
	for _, name := range rules {
		if err := catalog.CopyComponent(tmplDir, targetDir, "rules", name); err != nil {
			continue
		}
		fmt.Println(fmt.Sprintf("  %s %s", checkMark, infoStyle.Render(fmt.Sprintf("rule: %s", name))))
	}

	// Patch ralph.md for worktree isolation if requested
	if useWorktrees {
		if err := catalog.PatchRalphWorktrees(targetDir); err != nil {
			fmt.Fprintln(os.Stderr, errorStyle.Render(fmt.Sprintf("  ralph worktrees: %v", err)))
		} else {
			fmt.Println(fmt.Sprintf("  %s %s", checkMark, infoStyle.Render("command: ralph (worktree isolation enabled)")))
		}
	}

	// Install agent-teams rule if more than one agent selected
	if len(selectedAgents) > 1 {
		if err := catalog.CopyComponent(tmplDir, targetDir, "rules", "agent-teams"); err == nil {
			fmt.Println(fmt.Sprintf("  %s %s", checkMark, infoStyle.Render("rule: agent-teams")))
		}
	}

	// Apply security policy profile
	if policyProfile != "none" {
		if err := applyPolicyProfile(tmplDir, targetDir, policyProfile); err != nil {
			fmt.Fprintln(os.Stderr, errorStyle.Render(fmt.Sprintf("  policy: %v", err)))
		}
	}

	// Ensure global rules in ~/.claude/rules/
	fmt.Println(sectionHeader("Global Rules"))
	EnsureGlobalRules(tmplDir)

	// Check recommended dependencies (rtk, claude-mem, etc.)
	CheckDependencies()

	fmt.Println()
	fmt.Println(successStyle.Render(fmt.Sprintf("  %s Setup complete!", arrow)))
	fmt.Println(dimStyle.Render("  Run 'ck add' for more agents, 'ck remove' to remove components."))

	return nil
}

// applyPolicyProfile copies the selected profile as policy.yaml and applies it to settings.json.
func applyPolicyProfile(tmplDir, targetDir, profile string) error {
	profilePath := filepath.Join(tmplDir, "policies", profile+".yaml")
	content, err := os.ReadFile(profilePath)
	if err != nil {
		return fmt.Errorf("reading profile %s: %w", profile, err)
	}

	// Write policy.yaml
	policyPath := filepath.Join(targetDir, "policy.yaml")
	if err := os.WriteFile(policyPath, content, 0o644); err != nil {
		return fmt.Errorf("writing policy.yaml: %w", err)
	}
	fmt.Println(sectionHeader("Policy"))
	fmt.Println(fmt.Sprintf("  %s %s %s",
		checkMark,
		accentStyle.Render("policy.yaml"),
		dimStyle.Render(fmt.Sprintf("(%s profile)", profile)),
	))

	// Parse, resolve, and apply to settings.json
	spec, err := policy.ParseBytes(content)
	if err != nil {
		return fmt.Errorf("parsing policy: %w", err)
	}

	profilesDir := filepath.Join(tmplDir, "policies")
	resolved, err := policy.Resolve(spec, profilesDir)
	if err != nil {
		return fmt.Errorf("resolving policy: %w", err)
	}

	settingsPath := filepath.Join(targetDir, "settings.json")
	if err := policy.ApplySettings(settingsPath, resolved); err != nil {
		return fmt.Errorf("applying policy to settings.json: %w", err)
	}
	fmt.Println(fmt.Sprintf("  %s %s",
		checkMark,
		infoStyle.Render("settings.json updated with policy"),
	))

	return nil
}

// sortedKeys returns the keys of a map sorted alphabetically.
func sortedKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

