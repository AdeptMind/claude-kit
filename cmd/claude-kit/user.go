package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/AdeptMind/infra-tool/claude-cli/internal/config"
	"github.com/AdeptMind/infra-tool/claude-cli/internal/roles"
)

var userGlobal bool

var userCmd = &cobra.Command{
	Use:   "user [role]",
	Short: "View or change the current user role",
	Long: `View or change the current user role.

Without arguments, displays the active role and its working method summary.
With a role name, sets CK_USER_ROLE in settings.json.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runUser,
}

func init() {
	userCmd.Flags().BoolVar(&userGlobal, "global", false, "Write to ~/.claude/settings.json instead of project")
}

func runUser(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return showCurrentRole()
	}
	return setRole(args[0])
}

func showCurrentRole() error {
	role, name, err := roles.ActiveRole()
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Printf("  %s Role: %s\n", bullet, accentStyle.Render(name))
	if role.Description != "" {
		fmt.Printf("  %s %s\n", dot, role.Description)
	}
	fmt.Println()
	fmt.Printf("  %s\n", subtitleStyle.Render("Workflow"))
	fmt.Printf("    Inter-round gate:  %s\n", fmtBool(role.Workflow.InterRoundGate))
	fmt.Printf("    Client challenge:  %s\n", fmtBool(role.Workflow.ClientChallenge))
	fmt.Printf("    Readiness check:   %s\n", dimStyle.Render(role.Workflow.ReadinessCheck))
	fmt.Printf("    Done check:        %s\n", dimStyle.Render(role.Workflow.DoneCheck))
	fmt.Println()
	fmt.Printf("  %s\n", subtitleStyle.Render("Reporting"))
	fmt.Printf("    Language: %s\n", dimStyle.Render(role.Reporting.Language))
	fmt.Printf("    Proof:    %s\n", dimStyle.Render(role.Reporting.Proof))
	fmt.Println()

	return nil
}

func setRole(name string) error {
	// Validate against roles.yaml
	path := filepath.Join(config.TemplateDir(), "roles.yaml")
	cfg, err := roles.LoadRoles(path)
	if err != nil {
		return err
	}

	if _, ok := cfg.Roles[name]; !ok {
		names := make([]string, 0, len(cfg.Roles))
		for k := range cfg.Roles {
			names = append(names, k)
		}
		sort.Strings(names)
		return fmt.Errorf("unknown role %q; valid roles: %s", name, strings.Join(names, ", "))
	}

	targetDir := userSettingsDir()

	if err := patchSettingsEnvKey(targetDir, "CK_USER_ROLE", name); err != nil {
		return fmt.Errorf("updating settings.json: %w", err)
	}

	scope := "project"
	if userGlobal {
		scope = "global"
	}
	fmt.Println(successStyle.Render(fmt.Sprintf("  %s Role set to %s (%s)", arrow, name, scope)))
	return nil
}

func userSettingsDir() string {
	if userGlobal {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, ".claude")
	}
	return resolveTarget()
}

// patchSettingsEnvKey reads settings.json from dir, sets env.<key> = value, and writes back.
func patchSettingsEnvKey(dir, key, value string) error {
	path := filepath.Join(dir, "settings.json")

	var settings map[string]interface{}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		settings = make(map[string]interface{})
	} else if err != nil {
		return fmt.Errorf("reading settings.json: %w", err)
	} else {
		if err := json.Unmarshal(data, &settings); err != nil {
			return fmt.Errorf("parsing settings.json: %w", err)
		}
	}

	env, _ := settings["env"].(map[string]interface{})
	if env == nil {
		env = make(map[string]interface{})
	}
	env[key] = value
	settings["env"] = env

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	out, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("marshalling settings.json: %w", err)
	}
	out = append(out, '\n')

	return os.WriteFile(path, out, 0o644)
}

func fmtBool(b bool) string {
	if b {
		return successStyle.Render("on")
	}
	return dimStyle.Render("off")
}
