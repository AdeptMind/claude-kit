package bmadeval

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/AdeptMind/infra-tool/claude-cli/internal/eval"
)

func writeAssertions(t *testing.T, dir, filename string, assertions []eval.GradingAssertion) {
	t.Helper()
	data, err := json.Marshal(assertions)
	if err != nil {
		t.Fatalf("marshalling test data: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, filename), data, 0o644); err != nil {
		t.Fatalf("writing test file: %v", err)
	}
}

func TestLoadAssertions_Break(t *testing.T) {
	dir := t.TempDir()
	want := []eval.GradingAssertion{
		{Description: "Problem statement is specific", Type: "quality", Expected: "Clear scope"},
		{Description: "User stories defined", Type: "presence", Expected: "Numbered items"},
	}
	writeAssertions(t, dir, "break-assertions.json", want)

	got, err := LoadAssertions(dir, PhaseBreak)
	if err != nil {
		t.Fatalf("LoadAssertions(break) unexpected error: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("LoadAssertions(break) got %d assertions, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("assertion[%d] = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestLoadAssertions_Model(t *testing.T) {
	dir := t.TempDir()
	want := []eval.GradingAssertion{
		{Description: "Architecture defines components", Type: "presence", Expected: "3+ entries"},
		{Description: "ADRs document decisions", Type: "presence", Expected: "Decision records"},
	}
	writeAssertions(t, dir, "model-assertions.json", want)

	got, err := LoadAssertions(dir, PhaseModel)
	if err != nil {
		t.Fatalf("LoadAssertions(model) unexpected error: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("LoadAssertions(model) got %d assertions, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("assertion[%d] = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestLoadAssertions_Act(t *testing.T) {
	dir := t.TempDir()
	want := []eval.GradingAssertion{
		{Description: "All tasks addressed", Type: "completeness", Expected: "Code changes"},
		{Description: "Tests exist", Type: "presence", Expected: "Test files present"},
	}
	writeAssertions(t, dir, "act-assertions.json", want)

	got, err := LoadAssertions(dir, PhaseAct)
	if err != nil {
		t.Fatalf("LoadAssertions(act) unexpected error: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("LoadAssertions(act) got %d assertions, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("assertion[%d] = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestLoadAssertions_InvalidPhase(t *testing.T) {
	dir := t.TempDir()
	_, err := LoadAssertions(dir, Phase("unknown"))
	if err == nil {
		t.Fatal("LoadAssertions(unknown) expected error, got nil")
	}
	want := `bmadeval: unknown phase "unknown"`
	if err.Error() != want {
		t.Errorf("error = %q, want %q", err.Error(), want)
	}
}

func TestAssertionSchema(t *testing.T) {
	dir := t.TempDir()
	assertions := []eval.GradingAssertion{
		{Description: "Valid assertion", Type: "quality", Expected: "Expected value"},
		{Description: "", Type: "presence", Expected: "Something"},
	}
	writeAssertions(t, dir, "break-assertions.json", assertions)

	_, err := LoadAssertions(dir, PhaseBreak)
	if err == nil {
		t.Fatal("expected error for empty description, got nil")
	}

	assertions = []eval.GradingAssertion{
		{Description: "Valid assertion", Type: "", Expected: "Something"},
	}
	writeAssertions(t, dir, "break-assertions.json", assertions)

	_, err = LoadAssertions(dir, PhaseBreak)
	if err == nil {
		t.Fatal("expected error for empty type, got nil")
	}
}

func TestAvailablePhases(t *testing.T) {
	phases := AvailablePhases()
	if len(phases) != 3 {
		t.Fatalf("AvailablePhases() returned %d phases, want 3", len(phases))
	}
	expected := []Phase{PhaseBreak, PhaseModel, PhaseAct}
	for i, p := range phases {
		if p != expected[i] {
			t.Errorf("phase[%d] = %q, want %q", i, p, expected[i])
		}
	}
}
