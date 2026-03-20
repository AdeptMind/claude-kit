package policy

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

var (
	validProfiles       = map[string]bool{"permissive": true, "moderate": true, "strict": true}
	validNetworkModes   = map[string]bool{"block_all": true, "allowlist": true, "unrestricted": true}
	validSandboxRuntimes = map[string]bool{"auto": true, "docker": true, "podman": true, "nerdctl": true}
)

// ParseFile reads a YAML policy file and returns a validated PolicySpec.
func ParseFile(path string) (*PolicySpec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("policy: reading %s: %w", path, err)
	}
	return ParseBytes(data)
}

// ParseBytes parses YAML bytes into a validated PolicySpec.
func ParseBytes(data []byte) (*PolicySpec, error) {
	var spec PolicySpec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("policy: invalid YAML: %w", err)
	}
	if err := validate(&spec); err != nil {
		return nil, err
	}
	return &spec, nil
}

func validate(spec *PolicySpec) error {
	if err := validateProfileExtends(spec); err != nil {
		return err
	}
	if spec.Network != nil {
		if err := validateNetworkMode(spec.Network.Mode); err != nil {
			return err
		}
	}
	if spec.Sandbox != nil && spec.Sandbox.Runtime != "" {
		if err := validateSandboxRuntime(spec.Sandbox.Runtime); err != nil {
			return err
		}
	}
	return nil
}

func validateProfileExtends(spec *PolicySpec) error {
	hasProfile := spec.Profile != ""
	hasExtends := spec.Extends != ""

	if !hasProfile && !hasExtends {
		return fmt.Errorf("policy: exactly one of 'profile' or 'extends' is required, but neither is set")
	}
	if hasProfile && hasExtends {
		return fmt.Errorf("policy: 'profile' and 'extends' are mutually exclusive, but both are set")
	}

	value := spec.Profile
	if hasExtends {
		value = spec.Extends
	}
	if !validProfiles[value] {
		field := "profile"
		if hasExtends {
			field = "extends"
		}
		return fmt.Errorf("policy: %s %q is not valid; expected one of: permissive, moderate, strict", field, value)
	}
	return nil
}

func validateNetworkMode(mode string) error {
	if !validNetworkModes[mode] {
		return fmt.Errorf("policy: network.mode %q is not valid; expected one of: block_all, allowlist, unrestricted", mode)
	}
	return nil
}

func validateSandboxRuntime(runtime string) error {
	if !validSandboxRuntimes[runtime] {
		return fmt.Errorf("policy: sandbox.runtime %q is not valid; expected one of: auto, docker, podman, nerdctl", runtime)
	}
	return nil
}
