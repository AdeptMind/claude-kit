package audit

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestAppend_CreatesFileAndWritesEvent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", "audit.jsonl")

	ev := NewEvent("bash", "echo hello", StatusOK)
	if err := Append(path, ev); err != nil {
		t.Fatalf("Append() error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error: %v", err)
	}

	var got AuditEvent
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal() error: %v", err)
	}
	if got.Tool != "bash" {
		t.Errorf("tool = %q, want %q", got.Tool, "bash")
	}
	if got.Status != StatusOK {
		t.Errorf("status = %q, want %q", got.Status, StatusOK)
	}
}

func TestAppend_AppendsMultipleEvents(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit.jsonl")

	for i := range 5 {
		ev := NewEvent("tool", "input", StatusOK)
		ev.DurationMs = i
		if err := Append(path, ev); err != nil {
			t.Fatalf("Append() iteration %d error: %v", i, err)
		}
	}

	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	count := 0
	for scanner.Scan() {
		count++
	}
	if count != 5 {
		t.Errorf("line count = %d, want 5", count)
	}
}

func TestAppend_ConcurrentWrites(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit.jsonl")
	const n = 10

	var wg sync.WaitGroup
	wg.Add(n)
	for i := range n {
		go func(idx int) {
			defer wg.Done()
			ev := NewEvent("concurrent", "input", StatusOK)
			ev.DurationMs = idx
			if err := Append(path, ev); err != nil {
				t.Errorf("goroutine %d: Append() error: %v", idx, err)
			}
		}(i)
	}
	wg.Wait()

	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	count := 0
	for scanner.Scan() {
		line := scanner.Text()
		var ev AuditEvent
		if err := json.Unmarshal([]byte(line), &ev); err != nil {
			t.Errorf("line %d: invalid JSON: %v", count, err)
		}
		count++
	}
	if count != n {
		t.Errorf("line count = %d, want %d", count, n)
	}
}

func TestNewEvent_SetsTimestamp(t *testing.T) {
	before := time.Now().UTC()
	ev := NewEvent("test", "input", StatusDenied)
	after := time.Now().UTC()

	ts, err := time.Parse(time.RFC3339, ev.Timestamp)
	if err != nil {
		t.Fatalf("invalid timestamp %q: %v", ev.Timestamp, err)
	}
	if ts.Before(before.Truncate(time.Second)) || ts.After(after.Add(time.Second)) {
		t.Errorf("timestamp %v not between %v and %v", ts, before, after)
	}
	if ev.SessionID == "" {
		t.Error("session_id is empty")
	}
}

func TestNewEvent_UsesEnvSessionID(t *testing.T) {
	t.Setenv("CLAUDE_SESSION", "test-session-42")
	ev := NewEvent("test", "input", StatusOK)
	if ev.SessionID != "test-session-42" {
		t.Errorf("session_id = %q, want %q", ev.SessionID, "test-session-42")
	}
}

func TestTruncateInput_LongInput(t *testing.T) {
	long := string(make([]byte, 300))
	got := TruncateInput(long, 200)
	if len(got) != 203 { // 200 + "..."
		t.Errorf("len = %d, want 203", len(got))
	}
	if got[200:] != "..." {
		t.Errorf("suffix = %q, want %q", got[200:], "...")
	}
}

func TestTruncateInput_ShortInput(t *testing.T) {
	short := "hello"
	got := TruncateInput(short, 200)
	if got != short {
		t.Errorf("got = %q, want %q", got, short)
	}
}

func TestParseMaxSize_Various(t *testing.T) {
	tests := []struct {
		input string
		want  int64
	}{
		{"500KB", 500 * 1024},
		{"10MB", 10 * 1024 * 1024},
		{"1GB", 1024 * 1024 * 1024},
		{"10mb", 10 * 1024 * 1024},
		{"100", 100},
	}
	for _, tt := range tests {
		got, err := ParseMaxSize(tt.input)
		if err != nil {
			t.Errorf("ParseMaxSize(%q) error: %v", tt.input, err)
			continue
		}
		if got != tt.want {
			t.Errorf("ParseMaxSize(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestParseMaxSize_Invalid(t *testing.T) {
	for _, input := range []string{"", "abc", "-5MB", "0KB"} {
		if _, err := ParseMaxSize(input); err == nil {
			t.Errorf("ParseMaxSize(%q) expected error, got nil", input)
		}
	}
}

func TestRotation_TriggersAtMaxSize(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.jsonl")

	// Write events until we exceed a small threshold.
	for range 50 {
		ev := NewEvent("bash", "echo hello world this is a long input string", StatusOK)
		if err := AppendWithRotation(path, ev, "1KB"); err != nil {
			t.Fatalf("AppendWithRotation() error: %v", err)
		}
	}

	// Backup file must exist after rotation.
	if _, err := os.Stat(path + ".1"); err != nil {
		t.Fatalf("expected backup file %s.1 to exist: %v", path, err)
	}

	// Current file should be fresh (smaller than backup or newly created).
	cur, err := os.Stat(path)
	if err != nil {
		// After rotation the current file is gone until next write; that's OK.
		// But we wrote 50 events, so at least the last write after rotation should exist
		// only if there was a write after the rename. The rename happens after write,
		// so the current path may not exist. Let's check that backup is there.
		return
	}
	bak, _ := os.Stat(path + ".1")
	if cur.Size() > bak.Size() {
		t.Errorf("current file (%d) should be smaller than backup (%d)", cur.Size(), bak.Size())
	}
}

func TestRotation_OverwritesPreviousBackup(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.jsonl")

	// Create a sentinel backup file.
	if err := os.WriteFile(path+".1", []byte("old-backup\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Write enough to trigger rotation.
	for range 50 {
		ev := NewEvent("bash", "padding data to exceed threshold quickly", StatusOK)
		if err := AppendWithRotation(path, ev, "1KB"); err != nil {
			t.Fatalf("AppendWithRotation() error: %v", err)
		}
	}

	// Backup should be overwritten (no longer contains sentinel).
	data, err := os.ReadFile(path + ".1")
	if err != nil {
		t.Fatalf("ReadFile() error: %v", err)
	}
	if string(data) == "old-backup\n" {
		t.Error("backup file was not overwritten by rotation")
	}
}

func TestRotation_AtomicNoLostEvents(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.jsonl")

	// Use a large enough threshold so only one rotation occurs.
	// Each event is ~120 bytes; 20 events ~ 2.4KB; threshold at 2KB triggers once.
	const total = 20
	for i := range total {
		ev := NewEvent("tool", "input", StatusOK)
		ev.DurationMs = i
		if err := AppendWithRotation(path, ev, "2KB"); err != nil {
			t.Fatalf("AppendWithRotation() iteration %d error: %v", i, err)
		}
	}

	// With exactly one rotation, all events must be in current + backup.
	count := countLines(t, path) + countLines(t, path+".1")
	if count != total {
		t.Errorf("total events = %d, want %d (current + backup)", count, total)
	}

	// Verify every line is valid JSON.
	for _, p := range []string{path, path + ".1"} {
		f, err := os.Open(p)
		if err != nil {
			continue
		}
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			var ev AuditEvent
			if err := json.Unmarshal(scanner.Bytes(), &ev); err != nil {
				t.Errorf("invalid JSON in %s: %v", p, err)
			}
		}
		f.Close()
	}
}

func countLines(t *testing.T, path string) int {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0
		}
		t.Fatalf("Open(%s) error: %v", path, err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	n := 0
	for scanner.Scan() {
		n++
	}
	return n
}
