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
