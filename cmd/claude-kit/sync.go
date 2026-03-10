package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/charmbracelet/huh/spinner"
	"github.com/spf13/cobra"

	"github.com/AdeptMind/infra-tool/claude-cli/internal/catalog"
	"github.com/AdeptMind/infra-tool/claude-cli/internal/docsindex"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Update installed components and refresh docs-index",
	Long: `Sync updates installed components from the template catalog.

Only components that are already installed are updated — no new components
are added. After updating, the docs-index is refreshed if stale.`,
	RunE: runSync,
}

func runSync(cmd *cobra.Command, args []string) error {
	tmplDir := resolveTemplateDir()
	targetDir := resolveTarget()

	fmt.Println(banner())

	if _, err := os.Stat(targetDir); err != nil {
		return fmt.Errorf("no .claude directory found at %s — run 'ck init' first", targetDir)
	}

	var updated int
	var syncErr error

	action := func() {
		// Get installed components
		installed, err := catalog.GetInstalled(targetDir)
		if err != nil {
			syncErr = fmt.Errorf("scanning installed: %w", err)
			return
		}

		// Update base files
		if err := catalog.CopyBaseFiles(tmplDir, targetDir); err != nil {
			syncErr = fmt.Errorf("updating base files: %w", err)
			return
		}
		updated++

		// Update each installed component from template
		for _, cat := range installed {
			for _, comp := range cat.Components {
				// Check version before copying
				templatePath := templateComponentPath(tmplDir, cat.Name, comp.Name)
				if templatePath != "" {
					tmplVer := catalog.ExtractVersion(templatePath)
					if tmplVer != "" && comp.Version != "" && semverGT(tmplVer, comp.Version) {
						fmt.Printf("  ↑ %s/%s update available: %s → %s\n", cat.Name, comp.Name, comp.Version, tmplVer)
					}
				}

				err := catalog.CopyComponent(tmplDir, targetDir, cat.Name, comp.Name)
				if err != nil {
					// Skip components not in template (user-created)
					continue
				}
				updated++
			}
		}
	}

	if err := spinner.New().
		Title("Syncing components...").
		Action(action).
		Run(); err != nil {
		return err
	}

	if syncErr != nil {
		return syncErr
	}

	fmt.Println(fmt.Sprintf("  %s %s", checkMark, accentStyle.Render(fmt.Sprintf("Updated %d components", updated))))

	// Refresh docs-index
	projectRoot := filepath.Dir(targetDir)
	if strings.HasSuffix(targetDir, ".claude") {
		stale, reason := docsindex.IsStale(projectRoot)
		if stale {
			fmt.Println(warnStyle.Render(fmt.Sprintf("  %s Docs-index needs refresh: %s", bullet, reason)))

			var techs []string
			var docsErr error

			docsAction := func() {
				techs, docsErr = docsindex.Generate(projectRoot)
			}

			if err := spinner.New().
				Title("Refreshing docs-index...").
				Action(docsAction).
				Run(); err != nil {
				return err
			}

			if docsErr != nil {
				fmt.Fprintln(os.Stderr, errorStyle.Render(fmt.Sprintf("  Docs refresh failed: %v", docsErr)))
			} else {
				fmt.Println(fmt.Sprintf("  %s %s", checkMark, infoStyle.Render(fmt.Sprintf("Docs-index refreshed (stack: %s)", strings.Join(techs, ", ")))))
			}
		} else {
			fmt.Println(fmt.Sprintf("  %s %s", checkMark, dimStyle.Render("Docs-index is up to date")))
		}
	}

	return nil
}

// templateComponentPath returns the file/dir path of a component in the template directory.
// Returns "" if not determinable.
func templateComponentPath(tmplDir, compType, name string) string {
	switch compType {
	case "skills":
		return filepath.Join(tmplDir, "skills", name)
	case "agents", "commands", "rules":
		return filepath.Join(tmplDir, compType, name+".md")
	}
	return ""
}

// semverGT returns true if version a is strictly greater than b.
// Versions must be in the form "X.Y.Z" (semver). Returns false on parse error.
func semverGT(a, b string) bool {
	pa := parseSemver(a)
	pb := parseSemver(b)
	if pa == nil || pb == nil {
		return false
	}
	for i := 0; i < 3; i++ {
		if pa[i] > pb[i] {
			return true
		}
		if pa[i] < pb[i] {
			return false
		}
	}
	return false
}

// parseSemver parses a "X.Y.Z" version string into [3]int. Returns nil on error.
func parseSemver(v string) []int {
	parts := strings.SplitN(v, ".", 3)
	if len(parts) != 3 {
		return nil
	}
	nums := make([]int, 3)
	for i, p := range parts {
		n, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil {
			return nil
		}
		nums[i] = n
	}
	return nums
}
