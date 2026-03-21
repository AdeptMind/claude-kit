package roles

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/AdeptMind/infra-tool/claude-cli/internal/config"
)

const (
	rolesFileName  = "roles.yaml"
	envRoleKey     = "CK_USER_ROLE"
	defaultRole    = "dev"
)

// LoadRoles parses a roles.yaml file at the given path.
func LoadRoles(path string) (*RolesConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("roles.yaml: %w", err)
	}

	var cfg RolesConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("roles.yaml: invalid YAML: %w", err)
	}

	if len(cfg.Roles) == 0 {
		return nil, fmt.Errorf("roles.yaml: no roles defined")
	}

	return &cfg, nil
}

// ActiveRole reads CK_USER_ROLE from the environment (defaults to "dev"),
// loads roles.yaml from the template directory, and returns the role definition
// and its name.
func ActiveRole() (*RoleDefinition, string, error) {
	name := os.Getenv(envRoleKey)
	if name == "" {
		name = defaultRole
	}

	path := filepath.Join(config.TemplateDir(), rolesFileName)
	cfg, err := LoadRoles(path)
	if err != nil {
		return nil, "", err
	}

	role, ok := cfg.Roles[name]
	if !ok {
		return nil, "", unknownRoleError(name, cfg)
	}

	return &role, name, nil
}

// IsGateActive checks whether a given gate/check is active for the current role.
// Accepted gate names: "inter_round_gate", "client_challenge", "readiness_check", "done_check".
func IsGateActive(gate string) bool {
	role, _, err := ActiveRole()
	if err != nil {
		return false
	}

	switch gate {
	case "inter_round_gate":
		return role.Workflow.InterRoundGate
	case "client_challenge":
		return role.Workflow.ClientChallenge
	case "readiness_check":
		return role.Workflow.ReadinessCheck != "none"
	case "done_check":
		// "technical" is the base level (not active as enhanced); anything else is active.
		return role.Workflow.DoneCheck != "technical"
	default:
		return false
	}
}

// GetPhaseRoles returns the role names assigned to a phase in the "all" role.
func GetPhaseRoles(phase string) ([]string, error) {
	path := filepath.Join(config.TemplateDir(), rolesFileName)
	cfg, err := LoadRoles(path)
	if err != nil {
		return nil, err
	}

	allRole, ok := cfg.Roles["all"]
	if !ok {
		return nil, fmt.Errorf("roles.yaml: 'all' role not defined")
	}

	switch phase {
	case "break":
		return allRole.Phases.Break, nil
	case "model":
		return allRole.Phases.Model, nil
	case "analyze":
		return allRole.Phases.Analyze, nil
	case "act":
		return allRole.Phases.Act, nil
	case "deliver":
		return allRole.Phases.Deliver, nil
	default:
		return nil, fmt.Errorf("unknown phase %q; valid phases: break, model, analyze, act, deliver", phase)
	}
}

func unknownRoleError(name string, cfg *RolesConfig) error {
	names := make([]string, 0, len(cfg.Roles))
	for k := range cfg.Roles {
		names = append(names, k)
	}
	sort.Strings(names)
	return fmt.Errorf("unknown role %q; valid roles: %s", name, strings.Join(names, ", "))
}
