package policy

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

// Settings represents a Claude Code settings.json file.
type Settings struct {
	Permissions *PermissionsBlock      `json:"permissions,omitempty"`
	Hooks       *HooksBlock            `json:"hooks,omitempty"`
	Rest        map[string]interface{} `json:"-"`
}

// PermissionsBlock is the permissions section of settings.json.
type PermissionsBlock struct {
	Deny  []string `json:"deny,omitempty"`
	Allow []string `json:"allow,omitempty"`
}

// HooksBlock is the hooks section of settings.json.
type HooksBlock struct {
	PreToolUse  []HookEntry `json:"PreToolUse,omitempty"`
	PostToolUse []HookEntry `json:"PostToolUse,omitempty"`
}

// HookEntry is a single hook matcher+hooks entry in settings.json.
type HookEntry struct {
	Matcher string        `json:"matcher"`
	Hooks   []HookCommand `json:"hooks"`
}

// HookCommand is a hook command definition.
type HookCommand struct {
	Type    string `json:"type"`
	Command string `json:"command"`
}

// MergeSettings reads the existing settings.json, computes what the policy
// would change, and returns a human-readable diff string. It does NOT write.
func MergeSettings(settingsPath string, spec *PolicySpec) (string, error) {
	current, err := readSettingsRaw(settingsPath)
	if err != nil {
		return "", err
	}

	proposed := applyPolicyToRaw(current, spec)

	currentJSON, _ := json.MarshalIndent(current, "", "  ")
	proposedJSON, _ := json.MarshalIndent(proposed, "", "  ")

	if string(currentJSON) == string(proposedJSON) {
		return "", nil
	}

	return buildDiff(string(currentJSON), string(proposedJSON)), nil
}

// ApplySettings reads settings.json, applies the policy, and writes back.
func ApplySettings(settingsPath string, spec *PolicySpec) error {
	current, err := readSettingsRaw(settingsPath)
	if err != nil {
		return err
	}

	merged := applyPolicyToRaw(current, spec)

	data, err := json.MarshalIndent(merged, "", "  ")
	if err != nil {
		return fmt.Errorf("policy: marshalling settings: %w", err)
	}

	return os.WriteFile(settingsPath, append(data, '\n'), 0o644)
}

func readSettingsRaw(path string) (map[string]interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]interface{}), nil
		}
		return nil, fmt.Errorf("policy: reading settings: %w", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("policy: parsing settings: %w", err)
	}
	return raw, nil
}

func applyPolicyToRaw(raw map[string]interface{}, spec *PolicySpec) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range raw {
		result[k] = v
	}

	if spec.Filesystem != nil {
		perms := make(map[string]interface{})
		if existing, ok := result["permissions"].(map[string]interface{}); ok {
			for k, v := range existing {
				perms[k] = v
			}
		}
		if len(spec.Filesystem.Deny) > 0 {
			perms["deny"] = spec.Filesystem.Deny
		}
		if len(spec.Filesystem.Allow) > 0 {
			perms["allow"] = spec.Filesystem.Allow
		}
		result["permissions"] = perms
	}

	if spec.Hooks != nil {
		hooks := make(map[string]interface{})
		if existing, ok := result["hooks"].(map[string]interface{}); ok {
			for k, v := range existing {
				hooks[k] = v
			}
		}
		if len(spec.Hooks.PreToolUse) > 0 {
			hooks["PreToolUse"] = convertHooks(spec.Hooks.PreToolUse)
		}
		if len(spec.Hooks.PostToolUse) > 0 {
			hooks["PostToolUse"] = convertHooks(spec.Hooks.PostToolUse)
		}
		result["hooks"] = hooks
	}

	return result
}

func convertHooks(defs []HookDef) []map[string]interface{} {
	entries := make([]map[string]interface{}, len(defs))
	for i, d := range defs {
		entries[i] = map[string]interface{}{
			"matcher": d.Matcher,
			"hooks": []map[string]interface{}{
				{"type": "command", "command": d.Command},
			},
		}
	}
	return entries
}

func buildDiff(before, after string) string {
	beforeLines := strings.Split(before, "\n")
	afterLines := strings.Split(after, "\n")

	var sb strings.Builder
	sb.WriteString("--- settings.json (current)\n")
	sb.WriteString("+++ settings.json (after policy apply)\n")

	maxLen := len(beforeLines)
	if len(afterLines) > maxLen {
		maxLen = len(afterLines)
	}

	bSet := make(map[string]int)
	aSet := make(map[string]int)
	for _, l := range beforeLines {
		bSet[l]++
	}
	for _, l := range afterLines {
		aSet[l]++
	}

	bUsed := make(map[string]int)
	aUsed := make(map[string]int)

	// Show removed lines
	for _, l := range beforeLines {
		bUsed[l]++
		if bUsed[l] > aSet[l] {
			sb.WriteString(fmt.Sprintf("- %s\n", l))
		}
	}
	// Show added lines
	for _, l := range afterLines {
		aUsed[l]++
		if aUsed[l] > bSet[l] {
			sb.WriteString(fmt.Sprintf("+ %s\n", l))
		}
	}

	return sb.String()
}

// SortedYAML returns the PolicySpec as deterministically-ordered YAML.
func SortedYAML(spec *PolicySpec) ([]byte, error) {
	// Marshal to JSON first (deterministic field order from struct tags),
	// then convert to a sorted map for YAML output.
	jsonData, err := json.Marshal(spec)
	if err != nil {
		return nil, err
	}
	var m map[string]interface{}
	if err := json.Unmarshal(jsonData, &m); err != nil {
		return nil, err
	}
	return marshalSortedYAML(m, 0), nil
}

func marshalSortedYAML(m map[string]interface{}, indent int) []byte {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	prefix := strings.Repeat("  ", indent)
	var sb strings.Builder
	for _, k := range keys {
		v := m[k]
		switch val := v.(type) {
		case map[string]interface{}:
			sb.WriteString(fmt.Sprintf("%s%s:\n", prefix, k))
			sb.Write(marshalSortedYAML(val, indent+1))
		case []interface{}:
			sb.WriteString(fmt.Sprintf("%s%s:\n", prefix, k))
			for _, item := range val {
				if sub, ok := item.(map[string]interface{}); ok {
					sb.WriteString(fmt.Sprintf("%s  -\n", prefix))
					subYAML := marshalSortedYAML(sub, indent+2)
					sb.Write(subYAML)
				} else {
					sb.WriteString(fmt.Sprintf("%s  - %v\n", prefix, item))
				}
			}
		case bool:
			sb.WriteString(fmt.Sprintf("%s%s: %t\n", prefix, k, val))
		default:
			sb.WriteString(fmt.Sprintf("%s%s: %v\n", prefix, k, val))
		}
	}
	return []byte(sb.String())
}
