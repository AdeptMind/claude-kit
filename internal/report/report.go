package report

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/AdeptMind/infra-tool/claude-cli/internal/eval"
)

// ReportData holds all data needed to render the HTML report.
type ReportData struct {
	SkillName   string             `json:"skill_name"`
	GeneratedAt string             `json:"generated_at"`
	Current     *eval.EvalReport   `json:"current"`
	Previous    *eval.EvalReport   `json:"previous,omitempty"`
	History     []eval.LoopHistory `json:"history,omitempty"`
}

// GenerateHTML produces a standalone HTML report from eval results.
func GenerateHTML(data ReportData) string {
	var b strings.Builder

	b.WriteString(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Eval Report — ` + escape(data.SkillName) + `</title>
<style>
  * { margin: 0; padding: 0; box-sizing: border-box; }
  body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; line-height: 1.6; color: #1a1a2e; background: #f0f0f5; padding: 2rem; }
  .container { max-width: 960px; margin: 0 auto; }
  h1 { font-size: 1.5rem; margin-bottom: 0.5rem; }
  .meta { color: #666; font-size: 0.85rem; margin-bottom: 1.5rem; }
  .summary { display: flex; gap: 1rem; margin-bottom: 2rem; }
  .card { background: #fff; border-radius: 8px; padding: 1.25rem; box-shadow: 0 1px 3px rgba(0,0,0,0.1); flex: 1; }
  .card h2 { font-size: 0.9rem; color: #666; margin-bottom: 0.5rem; }
  .card .value { font-size: 2rem; font-weight: 700; }
  .pass { color: #16a34a; }
  .fail { color: #dc2626; }
  table { width: 100%; border-collapse: collapse; background: #fff; border-radius: 8px; overflow: hidden; box-shadow: 0 1px 3px rgba(0,0,0,0.1); margin-bottom: 2rem; }
  th, td { padding: 0.75rem 1rem; text-align: left; border-bottom: 1px solid #eee; }
  th { background: #f8f8fc; font-size: 0.85rem; color: #666; font-weight: 600; }
  .badge { display: inline-block; padding: 0.15rem 0.5rem; border-radius: 4px; font-size: 0.8rem; font-weight: 600; }
  .badge-pass { background: #dcfce7; color: #16a34a; }
  .badge-fail { background: #fee2e2; color: #dc2626; }
  .history { margin-top: 2rem; }
  .history h2 { font-size: 1.1rem; margin-bottom: 1rem; }
  .comparison { display: grid; grid-template-columns: 1fr 1fr; gap: 1rem; }
  @media (max-width: 640px) { .summary { flex-direction: column; } .comparison { grid-template-columns: 1fr; } }
</style>
</head>
<body>
<div class="container">
`)

	b.WriteString(fmt.Sprintf(`<h1>%s</h1>`, escape(data.SkillName)))
	b.WriteString(fmt.Sprintf(`<div class="meta">Generated %s</div>`, escape(data.GeneratedAt)))

	// Summary cards
	if data.Current != nil {
		s := data.Current.Summary
		rate := float64(0)
		if s.Total > 0 {
			rate = float64(s.Passed) / float64(s.Total) * 100
		}
		rateClass := "pass"
		if rate < 100 {
			rateClass = "fail"
		}
		b.WriteString(`<div class="summary">`)
		b.WriteString(fmt.Sprintf(`<div class="card"><h2>Pass Rate</h2><div class="value %s">%.0f%%</div></div>`, rateClass, rate))
		b.WriteString(fmt.Sprintf(`<div class="card"><h2>Passed</h2><div class="value pass">%d</div></div>`, s.Passed))
		b.WriteString(fmt.Sprintf(`<div class="card"><h2>Failed</h2><div class="value fail">%d</div></div>`, s.Failed))
		b.WriteString(fmt.Sprintf(`<div class="card"><h2>Total</h2><div class="value">%d</div></div>`, s.Total))
		b.WriteString(`</div>`)

		// Results table
		b.WriteString(`<table><thead><tr><th>Query</th><th>Expected</th><th>Rate</th><th>Result</th></tr></thead><tbody>`)
		for _, r := range data.Current.Results {
			expected := "trigger"
			if !r.ShouldTrigger {
				expected = "no trigger"
			}
			badge := `<span class="badge badge-pass">PASS</span>`
			if !r.Pass {
				badge = `<span class="badge badge-fail">FAIL</span>`
			}
			b.WriteString(fmt.Sprintf(`<tr><td>%s</td><td>%s</td><td>%.0f%% (%d/%d)</td><td>%s</td></tr>`,
				escape(r.Query), expected, r.TriggerRate*100, r.Triggers, r.Runs, badge))
		}
		b.WriteString(`</tbody></table>`)
	}

	// Comparison with previous
	if data.Previous != nil {
		b.WriteString(`<div class="comparison">`)
		b.WriteString(fmt.Sprintf(`<div class="card"><h2>Current</h2><div class="value">%d/%d</div></div>`,
			data.Current.Summary.Passed, data.Current.Summary.Total))
		b.WriteString(fmt.Sprintf(`<div class="card"><h2>Previous</h2><div class="value">%d/%d</div></div>`,
			data.Previous.Summary.Passed, data.Previous.Summary.Total))
		b.WriteString(`</div>`)
	}

	// History
	if len(data.History) > 0 {
		b.WriteString(`<div class="history"><h2>Iteration History</h2>`)
		b.WriteString(`<table><thead><tr><th>#</th><th>Train</th><th>Test</th></tr></thead><tbody>`)
		for _, h := range data.History {
			testStr := "—"
			if h.TestPassed != nil && h.TestTotal != nil {
				testStr = fmt.Sprintf("%d/%d", *h.TestPassed, *h.TestTotal)
			}
			b.WriteString(fmt.Sprintf(`<tr><td>%d</td><td>%d/%d</td><td>%s</td></tr>`,
				h.Iteration, h.TrainPassed, h.TrainTotal, testStr))
		}
		b.WriteString(`</tbody></table></div>`)
	}

	b.WriteString(`</div></body></html>`)
	return b.String()
}

// LoadEvalReport reads an eval report JSON file.
func LoadEvalReport(path string) (*eval.EvalReport, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	var report eval.EvalReport
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	return &report, nil
}

// NewReportData creates ReportData with current timestamp.
func NewReportData(skillName string, current *eval.EvalReport) ReportData {
	return ReportData{
		SkillName:   skillName,
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Current:     current,
	}
}

func escape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	return s
}
