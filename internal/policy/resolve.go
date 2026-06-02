package policy

import (
	"fmt"
	"path/filepath"
)

// Resolve applies profile inheritance. If spec uses "extends", the base profile
// is loaded and custom overrides are deep-merged on top of it.
func Resolve(spec *PolicySpec, profilesDir string) (*PolicySpec, error) {
	if spec.Extends == "" {
		return spec, nil
	}

	basePath := filepath.Join(profilesDir, spec.Extends+".yaml")
	base, err := ParseFile(basePath)
	if err != nil {
		return nil, fmt.Errorf("policy: resolving extends %q: %w", spec.Extends, err)
	}

	merged := *base
	merged.Profile = ""
	merged.Extends = ""

	mergeFilesystem(&merged, spec)
	mergeNetwork(&merged, spec)
	mergeHooks(&merged, spec)
	mergeSandbox(&merged, spec)

	return &merged, nil
}

// mergeNetwork: custom overrides base entirely if specified.
func mergeNetwork(base *PolicySpec, custom *PolicySpec) {
	if custom.Network != nil {
		base.Network = custom.Network
	}
}

// mergeSandbox: custom overrides base entirely if specified.
func mergeSandbox(base *PolicySpec, custom *PolicySpec) {
	if custom.Sandbox != nil {
		base.Sandbox = custom.Sandbox
	}
}

// mergeFilesystem: deny is additive (union, deduplicated), allow is replaced if custom specifies it.
func mergeFilesystem(base *PolicySpec, custom *PolicySpec) {
	if custom.Filesystem == nil {
		return
	}
	if base.Filesystem == nil {
		base.Filesystem = &FilesystemPolicy{}
	}

	base.Filesystem.Deny = dedup(append(base.Filesystem.Deny, custom.Filesystem.Deny...))

	if len(custom.Filesystem.Allow) > 0 {
		base.Filesystem.Allow = custom.Filesystem.Allow
	}
}

// mergeHooks: hooks are appended (base + custom).
func mergeHooks(base *PolicySpec, custom *PolicySpec) {
	if custom.Hooks == nil {
		return
	}
	if base.Hooks == nil {
		base.Hooks = &HooksPolicy{}
	}
	base.Hooks.PreToolUse = append(base.Hooks.PreToolUse, custom.Hooks.PreToolUse...)
	base.Hooks.PostToolUse = append(base.Hooks.PostToolUse, custom.Hooks.PostToolUse...)
}

func dedup(items []string) []string {
	seen := make(map[string]struct{}, len(items))
	result := make([]string, 0, len(items))
	for _, item := range items {
		if _, ok := seen[item]; !ok {
			seen[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}
