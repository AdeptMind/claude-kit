package report

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AdeptMind/infra-tool/claude-cli/internal/eval"
)

func sampleReport() *eval.EvalReport {
	return &eval.EvalReport{
		SkillName:   "test-skill",
		Description: "A test skill",
		Results: []eval.EvalResult{
			{Query: "do review", ShouldTrigger: true, TriggerRate: 1.0, Triggers: 3, Runs: 3, Pass: true},
			{Query: "write code", ShouldTrigger: false, TriggerRate: 0.0, Triggers: 0, Runs: 3, Pass: true},
			{Query: "check this", ShouldTrigger: true, TriggerRate: 0.33, Triggers: 1, Runs: 3, Pass: false},
		},
		Summary: eval.EvalSummary{Total: 3, Passed: 2, Failed: 1},
	}
}

func TestGenerateHTML_Standalone(t *testing.T) {
	data := NewReportData("code-reviewer", sampleReport())
	html := GenerateHTML(data)

	if !strings.Contains(html, "<!DOCTYPE html>") {
		t.Error("missing DOCTYPE")
	}
	if !strings.Contains(html, "code-reviewer") {
		t.Error("missing skill name")
	}
	if !strings.Contains(html, "</html>") {
		t.Error("missing closing html tag")
	}
	if !strings.Contains(html, "<style>") {
		t.Error("missing embedded styles — not standalone")
	}
}

func TestGenerateHTML_IterationHistory(t *testing.T) {
	tp := 4
	tt := 5
	data := NewReportData("test", sampleReport())
	data.History = []eval.LoopHistory{
		{Iteration: 1, TrainPassed: 3, TrainTotal: 5, TestPassed: &tp, TestTotal: &tt},
		{Iteration: 2, TrainPassed: 5, TrainTotal: 5},
	}
	html := GenerateHTML(data)

	if !strings.Contains(html, "Iteration History") {
		t.Error("missing history section")
	}
	if !strings.Contains(html, "4/5") {
		t.Error("missing test score in history")
	}
}

func TestGenerateHTML_SideBySideComparison(t *testing.T) {
	current := sampleReport()
	previous := &eval.EvalReport{
		Summary: eval.EvalSummary{Total: 3, Passed: 1, Failed: 2},
	}
	data := NewReportData("test", current)
	data.Previous = previous

	html := GenerateHTML(data)

	if !strings.Contains(html, "Current") {
		t.Error("missing current label")
	}
	if !strings.Contains(html, "Previous") {
		t.Error("missing previous label")
	}
}

func TestGenerateHTML_ResultsTable(t *testing.T) {
	data := NewReportData("test", sampleReport())
	html := GenerateHTML(data)

	if !strings.Contains(html, "do review") {
		t.Error("missing query text in table")
	}
	if !strings.Contains(html, "PASS") {
		t.Error("missing PASS badge")
	}
	if !strings.Contains(html, "FAIL") {
		t.Error("missing FAIL badge")
	}
}

func TestGenerateHTML_EscapesHTML(t *testing.T) {
	report := sampleReport()
	report.Results[0].Query = `<script>alert("xss")</script>`
	data := NewReportData("<b>evil</b>", report)
	html := GenerateHTML(data)

	if strings.Contains(html, "<script>") {
		t.Error("XSS: unescaped script tag")
	}
	if strings.Contains(html, "<b>evil</b>") {
		t.Error("XSS: unescaped HTML in title")
	}
}

func TestLoadEvalReport(t *testing.T) {
	dir := t.TempDir()
	report := sampleReport()
	data, _ := json.Marshal(report)
	path := filepath.Join(dir, "results.json")
	os.WriteFile(path, data, 0o644)

	loaded, err := LoadEvalReport(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if loaded.SkillName != "test-skill" {
		t.Errorf("skill_name = %q, want test-skill", loaded.SkillName)
	}
	if len(loaded.Results) != 3 {
		t.Errorf("got %d results, want 3", len(loaded.Results))
	}
}

func TestLoadEvalReport_FileNotFound(t *testing.T) {
	_, err := LoadEvalReport("/nonexistent.json")
	if err == nil {
		t.Error("expected error for missing file")
	}
}
