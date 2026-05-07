package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/AdeptMind/infra-tool/claude-cli/internal/catalog"
)

// globalRules lists rules that should always be present in ~/.claude/rules/.
// These are safety guardrails that apply globally, not per-project.
var globalRules = []string{
	"shared-systems",
}

// EnsureGlobalRules checks that all global rules exist in ~/.claude/rules/
// and copies them from the template if missing. Called by both init and sync.
func EnsureGlobalRules(tmplDir string) {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}

	globalRulesDir := filepath.Join(home, ".claude", "rules")
	if err := os.MkdirAll(globalRulesDir, 0o755); err != nil {
		return
	}

	installed := 0
	for _, rule := range globalRules {
		target := filepath.Join(globalRulesDir, rule+".md")
		if _, err := os.Stat(target); err == nil {
			continue // already exists
		}

		src := filepath.Join(tmplDir, "rules", rule+".md")
		content, err := os.ReadFile(src)
		if err != nil {
			fmt.Fprintln(os.Stderr, warnStyle.Render(fmt.Sprintf("  ⚠ global rule %s: template not found", rule)))
			continue
		}

		if err := os.WriteFile(target, content, 0o644); err != nil {
			fmt.Fprintln(os.Stderr, warnStyle.Render(fmt.Sprintf("  ⚠ global rule %s: %v", rule, err)))
			continue
		}
		installed++
		fmt.Println(fmt.Sprintf("  %s %s", checkMark, infoStyle.Render(fmt.Sprintf("global rule: %s → ~/.claude/rules/", rule))))
	}

	if installed == 0 {
		// Check for updates on existing rules
		for _, rule := range globalRules {
			target := filepath.Join(globalRulesDir, rule+".md")
			src := filepath.Join(tmplDir, "rules", rule+".md")

			tmplVer := catalog.ExtractVersion(src)
			installedVer := catalog.ExtractVersion(target)
			if tmplVer != "" && installedVer != "" && semverGT(tmplVer, installedVer) {
				// Update to newer version
				content, err := os.ReadFile(src)
				if err != nil {
					continue
				}
				if err := os.WriteFile(target, content, 0o644); err != nil {
					continue
				}
				fmt.Println(fmt.Sprintf("  %s %s", checkMark, infoStyle.Render(fmt.Sprintf("global rule: %s updated %s → %s", rule, installedVer, tmplVer))))
			}
		}
	}
}
