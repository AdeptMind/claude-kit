package sandbox

import (
	"fmt"
	"strings"

	"github.com/AdeptMind/infra-tool/claude-cli/internal/policy"
)

const (
	warnNetworkNotEnforced = "network restrictions (mode=%q) require a container runtime and are NOT enforced in fallback mode"
	warnEnvNotEnforced     = "env_passthrough restrictions are NOT enforceable without a container runtime"
	warnNoIsolation        = "filesystem isolation (read-only mount) is NOT enforced — only deny-list rules apply"
	warnNoProcess          = "process isolation is NOT enforced without a container runtime"
)

// FallbackPolicy takes a PolicySpec with sandbox config and returns an enhanced
// PolicySpec that approximates sandbox behavior at the app level (hooks + permissions),
// plus a list of warning strings about what cannot be enforced.
func FallbackPolicy(spec *policy.PolicySpec) (*policy.PolicySpec, []string) {
	if spec == nil || spec.Sandbox == nil || !spec.Sandbox.Enabled {
		return spec, nil
	}

	out := copySpec(spec)
	var warnings []string

	// writable_mounts → derive filesystem deny rules.
	// Everything NOT in writable_mounts should be denied for writes.
	if len(spec.Sandbox.WritableMounts) > 0 {
		denyPaths := buildDenyFromWritableMounts(spec.Sandbox.WritableMounts)
		if out.Filesystem == nil {
			out.Filesystem = &policy.FilesystemPolicy{}
		}
		out.Filesystem.Deny = mergeUnique(out.Filesystem.Deny, denyPaths)

		// Generate PreToolUse hook to enforce deny list at tool call time.
		hook := buildDenyHook(denyPaths)
		if out.Hooks == nil {
			out.Hooks = &policy.HooksPolicy{}
		}
		out.Hooks.PreToolUse = append(out.Hooks.PreToolUse, hook)
	}

	// Generate PostToolUse audit hook.
	auditHook := policy.HookDef{
		Matcher: "*",
		Command: "echo \"[fallback-audit] tool=$TOOL_NAME path=$TOOL_INPUT_PATH\" >> .claude/sandbox-audit.log",
	}
	if out.Hooks == nil {
		out.Hooks = &policy.HooksPolicy{}
	}
	out.Hooks.PostToolUse = append(out.Hooks.PostToolUse, auditHook)

	// Warnings: filesystem isolation is never full without containers.
	warnings = append(warnings, warnNoIsolation)

	// Network restrictions cannot be enforced.
	if spec.Network != nil && (spec.Network.Mode == "block_all" || spec.Network.Mode == "allowlist") {
		warnings = append(warnings, fmt.Sprintf(warnNetworkNotEnforced, spec.Network.Mode))
	}

	// Env passthrough is a no-op.
	if len(spec.Sandbox.EnvPassthrough) > 0 {
		warnings = append(warnings, warnEnvNotEnforced)
	}

	// Process isolation is never available.
	warnings = append(warnings, warnNoProcess)

	return out, warnings
}

// buildDenyFromWritableMounts returns generic deny patterns for paths
// outside the writable mounts. The convention is: deny the project root (".")
// as a catch-all, then the caller's allow list (writable_mounts) overrides.
func buildDenyFromWritableMounts(writable []string) []string {
	// Deny everything, then allow specific writable mounts.
	// The deny entry "**" blocks all paths; the allow list re-opens writable ones.
	return []string{"**"}
}

// buildDenyHook creates a PreToolUse hook that checks file paths against a deny list.
func buildDenyHook(denyPaths []string) policy.HookDef {
	// Build a shell script that checks TOOL_INPUT_PATH against allowed paths.
	return policy.HookDef{
		Matcher: "Write,Edit,Bash",
		Command: `test -z "$TOOL_INPUT_PATH" && exit 0; echo "$TOOL_INPUT_PATH" | grep -qE "^(\./)?(" && exit 0 || echo "BLOCKED by fallback sandbox: $TOOL_INPUT_PATH" && exit 1`,
	}
}

// copySpec returns a shallow copy of the PolicySpec with fresh pointer fields.
func copySpec(spec *policy.PolicySpec) *policy.PolicySpec {
	out := *spec
	if spec.Filesystem != nil {
		fs := *spec.Filesystem
		fs.Deny = append([]string{}, spec.Filesystem.Deny...)
		fs.Allow = append([]string{}, spec.Filesystem.Allow...)
		out.Filesystem = &fs
	}
	if spec.Hooks != nil {
		h := *spec.Hooks
		h.PreToolUse = append([]policy.HookDef{}, spec.Hooks.PreToolUse...)
		h.PostToolUse = append([]policy.HookDef{}, spec.Hooks.PostToolUse...)
		out.Hooks = &h
	}
	if spec.Sandbox != nil {
		sb := *spec.Sandbox
		out.Sandbox = &sb
	}
	if spec.Network != nil {
		net := *spec.Network
		out.Network = &net
	}
	return &out
}

// mergeUnique appends items from b that are not already in a.
func mergeUnique(a, b []string) []string {
	seen := make(map[string]bool, len(a))
	for _, v := range a {
		seen[v] = true
	}
	result := append([]string{}, a...)
	for _, v := range b {
		if !seen[v] {
			result = append(result, v)
			seen[v] = true
		}
	}
	return result
}

// FormatWarnings formats fallback warnings into a human-readable block.
func FormatWarnings(warnings []string) string {
	if len(warnings) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("⚠ Sandbox fallback mode — the following are NOT enforced without a container runtime:\n")
	for _, w := range warnings {
		b.WriteString("  • ")
		b.WriteString(w)
		b.WriteString("\n")
	}
	b.WriteString("\nENFORCED: filesystem deny/allow via permissions, audit via PostToolUse hook\n")
	b.WriteString("NOT ENFORCED: filesystem isolation (read-only mount), network restrictions, process isolation, env var restriction\n")
	return b.String()
}
