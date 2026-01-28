// Package report provides Markdown report generator implementation.
// Implements: Phase 5 - Report Generation (Markdown)
package report

import (
	"fmt"
	"strings"
	"time"

	"github.com/whhaicheng/DB-BenchMind/internal/domain/report"
)

// MarkdownGenerator generates Markdown format reports.
type MarkdownGenerator struct {
	chartGen *ChartGenerator
}

// NewMarkdownGenerator creates a new Markdown generator.
func NewMarkdownGenerator() *MarkdownGenerator {
	return &MarkdownGenerator{
		chartGen: NewChartGenerator(),
	}
}

// Generate generates a Markdown report.
func (g *MarkdownGenerator) Generate(data *report.GenerateContext) (*report.Report, error) {
	if err := data.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	var sb strings.Builder

	// Title
	g.writeTitle(&sb, data)

	// Summary
	g.writeSummary(&sb, data)

	// Environment
	if data.Config.IncludeParameters {
		g.writeEnvironment(&sb, data)
	}

	// Parameters
	if data.Config.IncludeParameters {
		g.writeParameters(&sb, data)
	}

	// Metrics
	g.writeMetrics(&sb, data)

	// Charts
	if data.Config.IncludeCharts && data.HasSamples() {
		g.writeCharts(&sb, data)
	}

	// Time Series
	if data.Config.IncludeTimeSeries && data.HasSamples() {
		g.writeTimeSeries(&sb, data)
	}

	// Logs
	if data.Config.IncludeLogs && len(data.Logs) > 0 {
		g.writeLogs(&sb, data)
	}

	// Raw Output
	if data.Config.IncludeLogs && data.RawOutput != "" {
		g.writeRawOutput(&sb, data)
	}

	// Footer
	g.writeFooter(&sb)

	return &report.Report{
		Format:      report.FormatMarkdown,
		Content:     []byte(sb.String()),
		GeneratedAt: time.Now(),
		RunID:       data.RunID,
	}, nil
}

// Format returns the format this generator produces.
func (g *MarkdownGenerator) Format() report.ReportFormat {
	return report.FormatMarkdown
}

// writeTitle writes the report title.
func (g *MarkdownGenerator) writeTitle(sb *strings.Builder, data *report.GenerateContext) {
	title := data.Config.Title
	if title == "" {
		title = fmt.Sprintf("Benchmark Report - %s", data.RunID)
	}
	sb.WriteString("# ")
	sb.WriteString(title)
	sb.WriteString("\n\n")
}

// writeSummary writes the summary section.
func (g *MarkdownGenerator) writeSummary(sb *strings.Builder, data *report.GenerateContext) {
	sb.WriteString("## Summary\n\n")

	// Status
	status := "✅ Completed"
	if data.IsFailed() {
		status = "❌ Failed"
	}
	sb.WriteString(fmt.Sprintf("- **Status**: %s\n", status))

	// Tool
	sb.WriteString(fmt.Sprintf("- **Tool**: %s\n", data.Tool))

	// Template
	sb.WriteString(fmt.Sprintf("- **Template**: %s\n", data.TemplateName))

	// Connection
	sb.WriteString(fmt.Sprintf("- **Database**: %s (%s)\n", data.ConnectionName, data.ConnectionType))

	// Duration
	sb.WriteString(fmt.Sprintf("- **Duration**: %s\n", data.GetDuration()))

	// Timestamps
	sb.WriteString(fmt.Sprintf("- **Started**: %s\n", report.GetTimestamp(data.StartedAt)))
	sb.WriteString(fmt.Sprintf("- **Completed**: %s\n", report.GetTimestamp(data.CompletedAt)))

	// Error message if failed
	if data.IsFailed() {
		sb.WriteString(fmt.Sprintf("- **Error**: %s\n", data.ErrorMessage))
	}

	sb.WriteString("\n")
}

// writeEnvironment writes the environment section.
func (g *MarkdownGenerator) writeEnvironment(sb *strings.Builder, data *report.GenerateContext) {
	sb.WriteString("## Environment\n\n")
	sb.WriteString("| Property | Value |\n")
	sb.WriteString("|----------|-------|\n")
	sb.WriteString(fmt.Sprintf("| Run ID | `%s` |\n", data.RunID))
	sb.WriteString(fmt.Sprintf("| Task ID | `%s` |\n", data.TaskID))
	sb.WriteString(fmt.Sprintf("| State | %s |\n", data.State))
	sb.WriteString(fmt.Sprintf("| Created | %s |\n", data.CreatedAt.Format(time.RFC1123)))
	sb.WriteString("\n")
}

// writeParameters writes the parameters section.
func (g *MarkdownGenerator) writeParameters(sb *strings.Builder, data *report.GenerateContext) {
	if len(data.Parameters) == 0 {
		return
	}

	sb.WriteString("## Parameters\n\n")
	for key, value := range data.Parameters {
		sb.WriteString(fmt.Sprintf("- **%s**: %v\n", key, value))
	}
	sb.WriteString("\n")
}

// writeMetrics writes the metrics section.
func (g *MarkdownGenerator) writeMetrics(sb *strings.Builder, data *report.GenerateContext) {
	sb.WriteString("## Metrics\n\n")

	if !data.HasMetrics() {
		sb.WriteString("*No metrics available*\n\n")
		return
	}

	// Key metrics table
	sb.WriteString("| Metric | Value |\n")
	sb.WriteString("|--------|-------|\n")
	sb.WriteString(fmt.Sprintf("| **TPS** | %.2f |\n", data.TPS))
	sb.WriteString(fmt.Sprintf("| **Avg Latency** | %.2f ms |\n", data.LatencyAvg))
	if data.LatencyP95 > 0 {
		sb.WriteString(fmt.Sprintf("| **P95 Latency** | %.2f ms |\n", data.LatencyP95))
	}
	if data.LatencyP99 > 0 {
		sb.WriteString(fmt.Sprintf("| **P99 Latency** | %.2f ms |\n", data.LatencyP99))
	}
	sb.WriteString(fmt.Sprintf("| **Total Transactions** | %d |\n", data.TotalTransactions))
	if data.TotalQueries > 0 {
		sb.WriteString(fmt.Sprintf("| **Total Queries** | %d |\n", data.TotalQueries))
	}
	sb.WriteString(fmt.Sprintf("| **Error Count** | %d |\n", data.ErrorCount))
	if data.ErrorRate > 0 {
		sb.WriteString(fmt.Sprintf("| **Error Rate** | %.2f%% |\n", data.ErrorRate))
	} else {
		sb.WriteString("| **Error Rate** | 0.00% |\n")
	}
	sb.WriteString("\n")
}

// writeCharts writes the charts section.
func (g *MarkdownGenerator) writeCharts(sb *strings.Builder, data *report.GenerateContext) {
	sb.WriteString("## Charts\n\n")

	width := data.Config.ChartWidth
	height := data.Config.ChartHeight

	// TPS sparkline
	if tpsChart := g.chartGen.GenerateTPSSparkline(data.Samples, width, height); tpsChart != "" {
		sb.WriteString("### TPS Over Time\n\n")
		sb.WriteString("```\n")
		sb.WriteString(tpsChart)
		sb.WriteString("\n```\n\n")
	}

	// Latency distribution
	if latencyChart := g.chartGen.GenerateLatencyDistribution(data.Samples, width); latencyChart != "" {
		sb.WriteString("### Latency Distribution\n\n")
		sb.WriteString("```\n")
		sb.WriteString(latencyChart)
		sb.WriteString("\n```\n\n")
	}
}

// writeTimeSeries writes the time series data section.
func (g *MarkdownGenerator) writeTimeSeries(sb *strings.Builder, data *report.GenerateContext) {
	sb.WriteString("## Time Series Data\n\n")
	sb.WriteString("| Timestamp | TPS | Latency (ms) | P95 (ms) | P99 (ms) | Error Rate (%) |\n")
	sb.WriteString("|-----------|-----|--------------|----------|----------|----------------|\n")

	for _, sample := range data.Samples {
		sb.WriteString(fmt.Sprintf("| %s | %.2f | %.2f | %.2f | %.2f | %.2f |\n",
			sample.Timestamp.Format("15:04:05"),
			sample.TPS,
			sample.LatencyAvg,
			sample.LatencyP95,
			sample.LatencyP99,
			sample.ErrorRate,
		))
	}
	sb.WriteString("\n")
}

// writeLogs writes the logs section.
func (g *MarkdownGenerator) writeLogs(sb *strings.Builder, data *report.GenerateContext) {
	sb.WriteString("## Logs\n\n")

	// Group logs by stream
	stdoutLogs := make([]string, 0)
	stderrLogs := make([]string, 0)

	for _, log := range data.Logs {
		if log.Stream == "stderr" {
			stderrLogs = append(stderrLogs, log.Content)
		} else {
			stdoutLogs = append(stdoutLogs, log.Content)
		}
	}

	// Write stderr logs first (errors)
	if len(stderrLogs) > 0 {
		sb.WriteString("### Errors\n\n")
		sb.WriteString("```\n")
		for _, log := range stderrLogs {
			sb.WriteString(log)
			sb.WriteString("\n")
		}
		sb.WriteString("```\n\n")
	}

	// Write stdout logs (limit to last 100 lines)
	if len(stdoutLogs) > 0 {
		sb.WriteString("### Output\n\n")
		sb.WriteString("```\n")
		start := 0
		if len(stdoutLogs) > 100 {
			start = len(stdoutLogs) - 100
			sb.WriteString(fmt.Sprintf("*... showing last 100 of %d lines ...*\n\n", len(stdoutLogs)))
		}
		for _, log := range stdoutLogs[start:] {
			sb.WriteString(log)
			sb.WriteString("\n")
		}
		sb.WriteString("```\n\n")
	}
}

// writeRawOutput writes the raw output section.
func (g *MarkdownGenerator) writeRawOutput(sb *strings.Builder, data *report.GenerateContext) {
	sb.WriteString("## Raw Output\n\n")
	sb.WriteString("```\n")
	// Limit raw output to prevent huge reports
	maxLines := 500
	lines := strings.Split(data.RawOutput, "\n")
	if len(lines) > maxLines {
		sb.WriteString("*... showing first and last 250 lines ...*\n\n")
		for _, line := range lines[:250] {
			sb.WriteString(line)
			sb.WriteString("\n")
		}
		sb.WriteString("\n... [truncated] ...\n\n")
		for _, line := range lines[len(lines)-250:] {
			sb.WriteString(line)
			sb.WriteString("\n")
		}
	} else {
		sb.WriteString(data.RawOutput)
	}
	sb.WriteString("\n```\n\n")
}

// writeFooter writes the report footer.
func (g *MarkdownGenerator) writeFooter(sb *strings.Builder) {
	sb.WriteString("---\n\n")
	sb.WriteString(fmt.Sprintf("*Generated by DB-BenchMind at %s*\n", time.Now().Format(time.RFC1123)))
}
