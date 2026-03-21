package audit

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
)

// ReadOptions controls filtering and limiting when reading audit events.
type ReadOptions struct {
	LastN        int
	ToolFilter   string
	StatusFilter string
}

// Read parses a JSONL audit log at path, applies filters, and returns matching events.
// Malformed lines are skipped with a warning to stderr.
// Returns an error if the file does not exist.
func Read(path string, opts ReadOptions) ([]AuditEvent, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("audit: open: %w", err)
	}
	defer f.Close()

	var events []AuditEvent
	scanner := bufio.NewScanner(f)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Bytes()

		var ev AuditEvent
		if err := json.Unmarshal(line, &ev); err != nil {
			fmt.Fprintf(os.Stderr, "audit: skipping malformed line %d: %v\n", lineNum, err)
			continue
		}

		if opts.ToolFilter != "" && ev.Tool != opts.ToolFilter {
			continue
		}
		if opts.StatusFilter != "" && ev.Status != opts.StatusFilter {
			continue
		}

		events = append(events, ev)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("audit: scan: %w", err)
	}

	if opts.LastN > 0 && len(events) > opts.LastN {
		events = events[len(events)-opts.LastN:]
	}

	return events, nil
}
