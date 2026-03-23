package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AdeptMind/infra-tool/claude-cli/internal/catalog"
	"github.com/AdeptMind/infra-tool/claude-cli/internal/config"
	"github.com/AdeptMind/infra-tool/claude-cli/internal/policy"
)

type AvailableAgent struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"` // management, dev, special
	Installed   bool   `json:"installed"`
}

type InstallSummary struct {
	Agents   int `json:"agents"`
	Skills   int `json:"skills"`
	Commands int `json:"commands"`
	Rules    int `json:"rules"`
}

type InitService struct{}

// ListAvailable returns all agents from the template dir, with installed status.
func (s *InitService) ListAvailable(projectPath string) ([]AvailableAgent, error) {
	tmplDir := config.TemplateDir()
	categories, err := catalog.ScanTemplate(tmplDir)
	if err != nil {
		return nil, fmt.Errorf("scanning templates: %w", err)
	}

	var agentComps []catalog.Component
	for _, cat := range categories {
		if cat.Name == "agents" {
			agentComps = cat.Components
			break
		}
	}

	agents := make([]AvailableAgent, 0, len(agentComps))
	for _, c := range agentComps {
		installed := false
		if projectPath != "" {
			installed = catalog.IsInstalled(filepath.Join(projectPath, ".claude"), "agents", c.Name)
		}
		agents = append(agents, AvailableAgent{
			Name:        c.Name,
			Description: c.Description,
			Category:    categorize(c.Name),
			Installed:   installed,
		})
	}
	return agents, nil
}

// PreviewInstall returns counts of what would be installed.
func (s *InitService) PreviewInstall(agentNames []string) InstallSummary {
	tmplDir := config.TemplateDir()
	skillSet := map[string]bool{}
	cmdSet := map[string]bool{}
	ruleSet := map[string]bool{}

	for _, name := range agentNames {
		// Skill deps from frontmatter
		agentPath := filepath.Join(tmplDir, "agents", name+".md")
		for _, dep := range catalog.ExtractSkillDeps(agentPath) {
			skillSet[dep] = true
		}
		// Extra skills
		for _, sk := range initGuessExtraSkills(name) {
			skillSet[sk] = true
		}
		// Commands
		for _, cmd := range initGuessCommands(name) {
			cmdSet[cmd] = true
		}
		// Rules
		for _, rule := range initGuessRules(name) {
			ruleSet[rule] = true
		}
	}

	return InstallSummary{
		Agents:   len(agentNames),
		Skills:   len(skillSet),
		Commands: len(cmdSet),
		Rules:    len(ruleSet),
	}
}

// Install installs selected agents + their dependencies into the project.
func (s *InitService) Install(projectPath string, agentNames []string, policyProfile string) error {
	tmplDir := config.TemplateDir()
	targetDir := filepath.Join(projectPath, ".claude")

	// Ensure base files
	claudeMd := filepath.Join(targetDir, "CLAUDE.md")
	if _, err := os.Stat(claudeMd); os.IsNotExist(err) {
		_ = catalog.CopyBaseFiles(tmplDir, targetDir)
	}

	for _, name := range agentNames {
		// Install agent
		if err := catalog.CopyComponent(tmplDir, targetDir, "agents", name); err != nil {
			continue
		}

		// Skill deps from frontmatter
		agentPath := filepath.Join(tmplDir, "agents", name+".md")
		for _, dep := range catalog.ExtractSkillDeps(agentPath) {
			if !catalog.IsInstalled(targetDir, "skills", dep) {
				_ = catalog.CopyComponent(tmplDir, targetDir, "skills", dep)
			}
		}

		// Extra skills
		for _, sk := range initGuessExtraSkills(name) {
			if !catalog.IsInstalled(targetDir, "skills", sk) {
				_ = catalog.CopyComponent(tmplDir, targetDir, "skills", sk)
			}
		}

		// Rules
		for _, rule := range initGuessRules(name) {
			if !catalog.IsInstalled(targetDir, "rules", rule) {
				_ = catalog.CopyComponent(tmplDir, targetDir, "rules", rule)
			}
		}

		// Commands
		for _, cmd := range initGuessCommands(name) {
			if !catalog.IsInstalled(targetDir, "commands", cmd) {
				_ = catalog.CopyComponent(tmplDir, targetDir, "commands", cmd)
			}
		}
	}

	// Apply policy profile
	if policyProfile != "" && policyProfile != "none" {
		if err := applyPolicy(tmplDir, targetDir, policyProfile); err != nil {
			return fmt.Errorf("applying policy %s: %w", policyProfile, err)
		}
	}

	return nil
}

func applyPolicy(tmplDir, targetDir, profile string) error {
	profilePath := filepath.Join(tmplDir, "policies", profile+".yaml")
	content, err := os.ReadFile(profilePath)
	if err != nil {
		return err
	}

	policyPath := filepath.Join(targetDir, "policy.yaml")
	if err := os.WriteFile(policyPath, content, 0o644); err != nil {
		return err
	}

	spec, err := policy.ParseBytes(content)
	if err != nil {
		return err
	}

	profilesDir := filepath.Join(tmplDir, "policies")
	resolved, err := policy.Resolve(spec, profilesDir)
	if err != nil {
		return err
	}

	settingsPath := filepath.Join(targetDir, "settings.json")
	return policy.ApplySettings(settingsPath, resolved)
}

// --- agent dependency maps (duplicated from cmd/claude-kit/add.go, package main) ---

func initGuessRules(name string) []string {
	switch strings.ToLower(name) {
	case "backend", "tech-lead":
		return []string{"code-style", "testing", "security", "api"}
	case "frontend":
		return []string{"code-style", "testing", "security", "frontend"}
	case "mobile-react-native", "mobile-flutter", "mobile-ios", "mobile-android":
		return []string{"code-style", "testing", "security"}
	case "ui-designer", "ux-designer":
		return []string{"frontend"}
	case "architect":
		return []string{"code-style", "security", "api", "infrastructure"}
	case "product-owner":
		return []string{"documentation"}
	case "devops":
		return []string{"infrastructure", "security", "documentation"}
	case "security", "pentester":
		return []string{"security"}
	case "finops":
		return []string{"finops", "infrastructure"}
	case "eval-grader", "eval-comparator", "eval-analyzer":
		return nil
	default:
		return []string{"code-style", "security"}
	}
}

func initGuessCommands(name string) []string {
	devCmds := []string{"commit-msg", "review", "test-gen", "pr-review", "code-only"}
	roleCmd := "role-" + name

	switch strings.ToLower(name) {
	case "backend", "frontend":
		return append(devCmds, roleCmd)
	case "mobile-react-native", "mobile-flutter", "mobile-ios", "mobile-android":
		return append(devCmds, roleCmd)
	case "tech-lead":
		return append(devCmds, "security-check", roleCmd)
	case "devops":
		return []string{"commit-msg", roleCmd}
	case "finops":
		return []string{"cost-review", roleCmd}
	case "security":
		return []string{"security-check", roleCmd}
	case "pentester":
		return []string{"pentest", roleCmd}
	case "eval-grader", "eval-comparator", "eval-analyzer":
		return nil
	case "architect":
		return []string{"docs-gen", roleCmd}
	case "product-owner":
		return []string{roleCmd}
	case "ui-designer", "ux-designer":
		return []string{roleCmd}
	default:
		return []string{roleCmd}
	}
}

func initGuessExtraSkills(name string) []string {
	switch strings.ToLower(name) {
	case "backend", "frontend", "mobile-react-native", "mobile-flutter",
		"mobile-ios", "mobile-android", "tech-lead", "devops":
		return []string{"git-commit-helper"}
	case "finops":
		return []string{"finops"}
	case "security":
		return []string{"security"}
	case "architect":
		return []string{"terraform-review"}
	default:
		return nil
	}
}
