package audit

import (
	"os"
	"path/filepath"
	"testing"
)

func writeJSONL(t *testing.T, path string, lines []string) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create file: %v", err)
	}
	defer f.Close()
	for _, l := range lines {
		f.WriteString(l + "\n")
	}
}

func sampleLines() []string {
	return []string{
		`{"timestamp":"2025-01-01T00:00:00Z","tool":"bash","input_summary":"ls","status":"ok","duration_ms":10,"session_id":"s1"}`,
		`{"timestamp":"2025-01-01T00:01:00Z","tool":"edit","input_summary":"fix","status":"denied","duration_ms":20,"session_id":"s1"}`,
		`{"timestamp":"2025-01-01T00:02:00Z","tool":"bash","input_summary":"pwd","status":"error","duration_ms":5,"session_id":"s1"}`,
		`{"timestamp":"2025-01-01T00:03:00Z","tool":"read","input_summary":"cat","status":"ok","duration_ms":15,"session_id":"s1"}`,
	}
}

func TestRead_AllEvents(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit.jsonl")
	writeJSONL(t, path, sampleLines())

	events, err := Read(path, ReadOptions{})
	if err != nil {
		t.Fatalf("Read() error: %v", err)
	}
	if len(events) != 4 {
		t.Fatalf("got %d events, want 4", len(events))
	}
	if events[0].Tool != "bash" {
		t.Errorf("events[0].Tool = %q, want %q", events[0].Tool, "bash")
	}
}

func TestRead_FilterByTool(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit.jsonl")
	writeJSONL(t, path, sampleLines())

	events, err := Read(path, ReadOptions{ToolFilter: "bash"})
	if err != nil {
		t.Fatalf("Read() error: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("got %d events, want 2", len(events))
	}
	for _, ev := range events {
		if ev.Tool != "bash" {
			t.Errorf("unexpected tool %q", ev.Tool)
		}
	}
}

func TestRead_FilterByStatus(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit.jsonl")
	writeJSONL(t, path, sampleLines())

	events, err := Read(path, ReadOptions{StatusFilter: "ok"})
	if err != nil {
		t.Fatalf("Read() error: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("got %d events, want 2", len(events))
	}
	for _, ev := range events {
		if ev.Status != "ok" {
			t.Errorf("unexpected status %q", ev.Status)
		}
	}
}

func TestRead_CombinedFilters(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit.jsonl")
	writeJSONL(t, path, sampleLines())

	events, err := Read(path, ReadOptions{ToolFilter: "bash", StatusFilter: "ok"})
	if err != nil {
		t.Fatalf("Read() error: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("got %d events, want 1", len(events))
	}
	if events[0].InputSummary != "ls" {
		t.Errorf("input_summary = %q, want %q", events[0].InputSummary, "ls")
	}
}

func TestRead_LastN(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit.jsonl")
	writeJSONL(t, path, sampleLines())

	events, err := Read(path, ReadOptions{LastN: 2})
	if err != nil {
		t.Fatalf("Read() error: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("got %d events, want 2", len(events))
	}
	if events[0].Tool != "bash" || events[0].InputSummary != "pwd" {
		t.Errorf("events[0] = %+v, want bash/pwd", events[0])
	}
	if events[1].Tool != "read" {
		t.Errorf("events[1].Tool = %q, want %q", events[1].Tool, "read")
	}
}

func TestRead_MalformedLines(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit.jsonl")
	lines := []string{
		`{"timestamp":"2025-01-01T00:00:00Z","tool":"bash","input_summary":"ls","status":"ok","duration_ms":10,"session_id":"s1"}`,
		`not valid json`,
		`{"timestamp":"2025-01-01T00:02:00Z","tool":"edit","input_summary":"fix","status":"ok","duration_ms":5,"session_id":"s1"}`,
		`{broken`,
	}
	writeJSONL(t, path, lines)

	events, err := Read(path, ReadOptions{})
	if err != nil {
		t.Fatalf("Read() error: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("got %d events, want 2 (malformed lines skipped)", len(events))
	}
	if events[0].Tool != "bash" {
		t.Errorf("events[0].Tool = %q, want %q", events[0].Tool, "bash")
	}
	if events[1].Tool != "edit" {
		t.Errorf("events[1].Tool = %q, want %q", events[1].Tool, "edit")
	}
}

func TestRead_EmptyFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit.jsonl")
	os.WriteFile(path, []byte{}, 0o644)

	events, err := Read(path, ReadOptions{})
	if err != nil {
		t.Fatalf("Read() error: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("got %d events, want 0", len(events))
	}
}

func TestRead_MissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nonexistent.jsonl")

	_, err := Read(path, ReadOptions{})
	if err == nil {
		t.Fatal("Read() expected error for missing file, got nil")
	}
}
