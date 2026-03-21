package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/AdeptMind/infra-tool/claude-cli/internal/audit"
)

var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "View and export audit logs",
}

var auditViewCmd = &cobra.Command{
	Use:   "view",
	Short: "Display audit log as a formatted table",
	RunE: func(cmd *cobra.Command, args []string) error {
		last, _ := cmd.Flags().GetInt("last")
		filters, _ := cmd.Flags().GetStringArray("filter")
		return runAuditView(last, filters)
	},
}

var auditExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export audit log as CSV or JSON",
	RunE: func(cmd *cobra.Command, args []string) error {
		format, _ := cmd.Flags().GetString("format")
		filters, _ := cmd.Flags().GetStringArray("filter")
		last, _ := cmd.Flags().GetInt("last")
		return runAuditExport(format, last, filters)
	},
}

func init() {
	auditViewCmd.Flags().Int("last", 0, "Show only the last N entries")
	auditViewCmd.Flags().StringArray("filter", nil, "Filter events (key=value, e.g. tool=Bash, status=ok)")

	auditExportCmd.Flags().String("format", "csv", "Export format: csv or json")
	auditExportCmd.Flags().Int("last", 0, "Show only the last N entries")
	auditExportCmd.Flags().StringArray("filter", nil, "Filter events (key=value, e.g. tool=Bash, status=ok)")

	auditCmd.AddCommand(auditViewCmd)
	auditCmd.AddCommand(auditExportCmd)
}

func auditLogPath() string {
	return filepath.Join(resolveProjectRoot(), ".claude", "audit.jsonl")
}

func parseFilters(filters []string) audit.ReadOptions {
	var opts audit.ReadOptions
	for _, f := range filters {
		parts := strings.SplitN(f, "=", 2)
		if len(parts) != 2 {
			continue
		}
		switch parts[0] {
		case "tool":
			opts.ToolFilter = parts[1]
		case "status":
			opts.StatusFilter = parts[1]
		}
	}
	return opts
}

func runAuditView(last int, filters []string) error {
	opts := parseFilters(filters)
	opts.LastN = last

	events, err := audit.Read(auditLogPath(), opts)
	if err != nil {
		return err
	}

	if len(events) == 0 {
		fmt.Println(dimStyle.Render("  No audit events found."))
		return nil
	}

	// Column widths
	const (
		colTimestamp = 20
		colTool      = 16
		colStatus    = 8
		colDuration  = 10
	)
	fixedWidth := colTimestamp + colTool + colStatus + colDuration + 12 // separators
	tw := termWidth()
	colInput := tw - fixedWidth
	if colInput < 20 {
		colInput = 20
	}

	// Header
	header := fmt.Sprintf("%-*s  %-*s  %-*s  %-*s  %-*s",
		colTimestamp, "TIMESTAMP",
		colTool, "TOOL",
		colInput, "INPUT",
		colStatus, "STATUS",
		colDuration, "DURATION",
	)
	fmt.Println(tableHeaderStyle.Render(header))

	// Rows
	rowStyle := lipgloss.NewStyle().Foreground(white)
	for _, ev := range events {
		ts := audit.TruncateInput(ev.Timestamp, colTimestamp)
		tool := audit.TruncateInput(ev.Tool, colTool)
		input := audit.TruncateInput(ev.InputSummary, colInput)
		status := audit.TruncateInput(ev.Status, colStatus)
		dur := fmt.Sprintf("%dms", ev.DurationMs)

		row := fmt.Sprintf("%-*s  %-*s  %-*s  %-*s  %-*s",
			colTimestamp, ts,
			colTool, tool,
			colInput, input,
			colStatus, status,
			colDuration, dur,
		)
		fmt.Println(rowStyle.Render(row))
	}

	return nil
}

func runAuditExport(format string, last int, filters []string) error {
	opts := parseFilters(filters)
	opts.LastN = last

	events, err := audit.Read(auditLogPath(), opts)
	if err != nil {
		return err
	}

	switch format {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(events)
	case "csv":
		w := csv.NewWriter(os.Stdout)
		defer w.Flush()
		if err := w.Write([]string{"timestamp", "tool", "input_summary", "status", "duration_ms"}); err != nil {
			return err
		}
		for _, ev := range events {
			if err := w.Write([]string{
				ev.Timestamp,
				ev.Tool,
				ev.InputSummary,
				ev.Status,
				fmt.Sprintf("%d", ev.DurationMs),
			}); err != nil {
				return err
			}
		}
		return nil
	default:
		return fmt.Errorf("unsupported format: %s (use csv or json)", format)
	}
}
