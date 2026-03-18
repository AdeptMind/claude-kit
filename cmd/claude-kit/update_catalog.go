package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh/spinner"
	"github.com/spf13/cobra"

	"github.com/AdeptMind/infra-tool/claude-cli/internal/config"
)

var updateCatalogCmd = &cobra.Command{
	Use:   "update-catalog",
	Short: "Sync the local template catalog from the current repo",
	Long: `Copy the project-template/.claude/ directory into ~/.bmad/templates/.

This updates the local catalog so that 'ck init' and 'ck add' see all
available agents, skills, commands, and rules from this repository.

Run this after pulling new agents or merging upstream changes.`,
	RunE: runUpdateCatalog,
}

func init() {
	rootCmd.AddCommand(updateCatalogCmd)
}

func runUpdateCatalog(cmd *cobra.Command, args []string) error {
	// Resolve source: use --template-dir if set, otherwise find project-template/.claude/ from cwd
	srcDir := findRepoTemplate()
	if srcDir == "" {
		return fmt.Errorf("cannot find project-template/.claude/ — run this from the claude-kit repo root")
	}

	dstDir := config.InstalledTemplatesDir()

	fmt.Println(banner())
	fmt.Println(subtitleStyle.Render("  Update local catalog"))
	fmt.Println(dimStyle.Render(fmt.Sprintf("  Source: %s", srcDir)))
	fmt.Println(dimStyle.Render(fmt.Sprintf("  Target: %s", dstDir)))
	fmt.Println()

	var copied int
	var syncErr error

	action := func() {
		copied, syncErr = copyTree(srcDir, dstDir)
	}

	if err := spinner.New().
		Title("Copying templates...").
		Action(action).
		Run(); err != nil {
		return err
	}

	if syncErr != nil {
		return syncErr
	}

	fmt.Println(fmt.Sprintf("  %s %s", checkMark, accentStyle.Render(fmt.Sprintf("Catalog updated — %d files copied", copied))))
	return nil
}

// findRepoTemplate locates project-template/.claude/ relative to cwd or the binary.
func findRepoTemplate() string {
	// If --template-dir was explicitly set and looks like a repo template, use it
	if templateDir != "" {
		if _, err := os.Stat(filepath.Join(templateDir, "agents")); err == nil {
			return templateDir
		}
	}

	// Walk up from cwd looking for project-template/.claude/
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	for dir := cwd; ; dir = filepath.Dir(dir) {
		candidate := filepath.Join(dir, "project-template", config.ClaudeDirName)
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
	}

	// Try relative to binary (dev mode)
	if exe, err := os.Executable(); err == nil {
		candidate := filepath.Join(filepath.Dir(exe), "project-template", config.ClaudeDirName)
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
	}

	return ""
}

// copyTree recursively copies src into dst, returning the number of files copied.
func copyTree(src, dst string) (int, error) {
	count := 0

	if err := os.MkdirAll(dst, 0o755); err != nil {
		return 0, err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return 0, err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			n, err := copyTree(srcPath, dstPath)
			if err != nil {
				return count, err
			}
			count += n
		} else {
			data, err := os.ReadFile(srcPath)
			if err != nil {
				return count, err
			}
			if err := os.WriteFile(dstPath, data, 0o644); err != nil {
				return count, err
			}
			count++
		}
	}

	return count, nil
}
