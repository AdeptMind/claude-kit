package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

// Append serializes event as a single JSON line and appends it to the file at path.
// It creates parent directories and the file if they don't exist.
// An exclusive flock is held during the write for concurrent safety.
func Append(path string, event AuditEvent) error {
	return AppendWithRotation(path, event, "")
}

// AppendWithRotation works like Append but rotates the log file when it exceeds maxSize.
// maxSize is a human-readable size string (e.g. "10MB", "500KB", "1GB").
// An empty maxSize disables rotation.
// Rotation keeps at most one backup file (path + ".1"), overwriting any previous backup.
func AppendWithRotation(path string, event AuditEvent, maxSize string) error {
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

	if maxSize != "" {
		limit, err := ParseMaxSize(maxSize)
		if err != nil {
			return fmt.Errorf("audit: %w", err)
		}
		if shouldRotate(f, limit) {
			_ = os.Rename(path, path+".1")
		}
	}

	return nil
}

// ParseMaxSize converts a human-readable size string to bytes.
// Supported suffixes: KB, MB, GB (case-insensitive). Plain numbers are treated as bytes.
func ParseMaxSize(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty size string")
	}

	upper := strings.ToUpper(s)
	var suffix string
	var multiplier int64 = 1

	switch {
	case strings.HasSuffix(upper, "GB"):
		suffix = "GB"
		multiplier = 1024 * 1024 * 1024
	case strings.HasSuffix(upper, "MB"):
		suffix = "MB"
		multiplier = 1024 * 1024
	case strings.HasSuffix(upper, "KB"):
		suffix = "KB"
		multiplier = 1024
	}

	numStr := strings.TrimSuffix(upper, suffix)
	n, err := strconv.ParseInt(strings.TrimSpace(numStr), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid size %q: %w", s, err)
	}
	if n <= 0 {
		return 0, fmt.Errorf("size must be positive: %q", s)
	}

	return n * multiplier, nil
}

// shouldRotate checks if the open file exceeds the given byte limit.
func shouldRotate(f *os.File, maxBytes int64) bool {
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return info.Size() > maxBytes
}
