package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

// Append serializes event as a single JSON line and appends it to the file at path.
// It creates parent directories and the file if they don't exist.
// An exclusive flock is held during the write for concurrent safety.
func Append(path string, event AuditEvent) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("audit: mkdir: %w", err)
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("audit: open: %w", err)
	}
	defer f.Close()

	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		return fmt.Errorf("audit: flock: %w", err)
	}
	defer syscall.Flock(int(f.Fd()), syscall.LOCK_UN)

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("audit: marshal: %w", err)
	}
	data = append(data, '\n')

	if _, err := f.Write(data); err != nil {
		return fmt.Errorf("audit: write: %w", err)
	}

	return nil
}
