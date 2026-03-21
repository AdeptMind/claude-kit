package audit

import "fmt"

const (
	maxInputLen    = 200
	defaultTool    = "unknown"
	defaultInput   = "unknown"
	defaultSession = "none"
)

// GenerateHookCommand returns a single-line shell command suitable for a
// Claude PostToolUse hook. It captures CLAUDE_TOOL_NAME and CLAUDE_TOOL_INPUT,
// truncates the input to 200 chars, and appends a JSONL line to auditPath.
func GenerateHookCommand(auditPath string) string {
	// The command:
	// 1. Reads env vars with defaults for missing values
	// 2. Truncates input to 200 chars
	// 3. Escapes double quotes and backslashes in input for valid JSON
	// 4. Writes a JSON line to the audit file
	return fmt.Sprintf(
		`t="${CLAUDE_TOOL_NAME:-unknown}"; i="${CLAUDE_TOOL_INPUT:-unknown}"; s="${CLAUDE_SESSION:-none}"; i="$(printf '%%s' "$i" | cut -c1-200)"; i="$(printf '%%s' "$i" | sed 's/\\/\\\\/g; s/"/\\"/g')"; mkdir -p "$(dirname '%s')" && printf '{"timestamp":"%%s","tool":"%%s","input_summary":"%%s","status":"ok","duration_ms":0,"session_id":"%%s"}\n' "$(date -u +%%Y-%%m-%%dT%%H:%%M:%%SZ)" "$t" "$i" "$s" >> '%s'`,
		auditPath, auditPath,
	)
}

// GenerateHookEntry returns the full hook configuration map suitable for
// inclusion in a Claude settings.json PostToolUse array.
func GenerateHookEntry(auditPath string) map[string]interface{} {
	return map[string]interface{}{
		"type":    "command",
		"command": GenerateHookCommand(auditPath),
	}
}
