package audit

import (
	"os"
	"time"

	"crypto/rand"
	"encoding/hex"
)

// Status constants for audit events.
const (
	StatusOK     = "ok"
	StatusDenied = "denied"
	StatusError  = "error"
)

// AuditEvent represents a single auditable action.
type AuditEvent struct {
	Timestamp    string `json:"timestamp"`
	Tool         string `json:"tool"`
	InputSummary string `json:"input_summary"`
	Status       string `json:"status"`
	DurationMs   int    `json:"duration_ms"`
	SessionID    string `json:"session_id"`
}

// NewEvent creates an AuditEvent with timestamp and session_id populated.
func NewEvent(tool, inputSummary, status string) AuditEvent {
	return AuditEvent{
		Timestamp:    time.Now().UTC().Format(time.RFC3339),
		Tool:         tool,
		InputSummary: inputSummary,
		Status:       status,
		SessionID:    sessionID(),
	}
}

// TruncateInput truncates input to maxLen characters, appending "..." if truncated.
func TruncateInput(input string, maxLen int) string {
	if len(input) <= maxLen {
		return input
	}
	return input[:maxLen] + "..."
}

// sessionID returns CLAUDE_SESSION env var or generates a random hex string.
func sessionID() string {
	if id := os.Getenv("CLAUDE_SESSION"); id != "" {
		return id
	}
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
