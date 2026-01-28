// Package report provides HTML report generator implementation.
// Implements: Phase 5 - Report Generation (HTML)
package report

import (
	"fmt"
	"strings"
	"time"

	"github.com/whhaicheng/DB-BenchMind/internal/domain/report"
)

// HTMLGenerator generates HTML format reports.
type HTMLGenerator struct {
	chartGen *ChartGenerator
}

// NewHTMLGenerator creates a new HTML generator.
func NewHTMLGenerator() *HTMLGenerator {
	return &HTMLGenerator{
		chartGen: NewChartGenerator(),
	}
}

// Generate generates an HTML report.
func (g *HTMLGenerator) Generate(data *report.GenerateContext) (*report.Report, error) {
	if err := data.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	var sb strings.Builder

	// HTML header
	g.writeHeader(&sb, data)

	// Body start
	sb.WriteString(`<body>`)

	// Container
	sb.WriteString(`<div class="container">`)

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

	// Container end
	sb.WriteString(`</div>`)

	// Footer
	g.writeFooter(&sb)

	// Body end
	sb.WriteString(`</body>`)

	// HTML end
	sb.WriteString(`</html>`)

	return &report.Report{
		Format:      report.FormatHTML,
		Content:     []byte(sb.String()),
		GeneratedAt: time.Now(),
		RunID:       data.RunID,
	}, nil
}

// Format returns the format this generator produces.
func (g *HTMLGenerator) Format() report.ReportFormat {
	return report.FormatHTML
}

// writeHeader writes the HTML header with embedded CSS.
func (g *HTMLGenerator) writeHeader(sb *strings.Builder, data *report.GenerateContext) {
	title := data.Config.Title
	if title == "" {
		title = fmt.Sprintf("Benchmark Report - %s", data.RunID)
	}

	sb.WriteString(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>`)
	sb.WriteString(title)
	sb.WriteString(`</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
            line-height: 1.6;
            color: #333;
            background: #f5f5f5;
            padding: 20px;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
            background: white;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
            padding: 40px;
        }
        h1 {
            color: #2c3e50;
            margin-bottom: 30px;
            border-bottom: 3px solid #3498db;
            padding-bottom: 10px;
        }
        h2 {
            color: #34495e;
            margin-top: 30px;
            margin-bottom: 15px;
            font-size: 1.5em;
        }
        .summary {
            background: #ecf0f1;
            padding: 20px;
            border-radius: 5px;
            margin-bottom: 20px;
        }
        .summary-item {
            display: inline-block;
            margin: 5px 15px 5px 0;
        }
        .summary-label {
            font-weight: bold;
            color: #7f8c8d;
        }
        .status-success {
            color: #27ae60;
            font-weight: bold;
        }
        .status-failed {
            color: #e74c3c;
            font-weight: bold;
        }
        table {
            width: 100%;
            border-collapse: collapse;
            margin: 20px 0;
        }
        th, td {
            padding: 12px;
            text-align: left;
            border-bottom: 1px solid #ddd;
        }
        th {
            background-color: #3498db;
            color: white;
            font-weight: 600;
        }
        tr:hover {
            background-color: #f5f5f5;
        }
        .metric-card {
            display: inline-block;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 20px;
            margin: 10px;
            border-radius: 8px;
            min-width: 200px;
            box-shadow: 0 4px 6px rgba(0,0,0,0.1);
        }
        .metric-label {
            font-size: 0.9em;
            opacity: 0.9;
        }
        .metric-value {
            font-size: 2em;
            font-weight: bold;
            margin-top: 5px;
        }
        .chart-container {
            background: #f8f9fa;
            padding: 20px;
            border-radius: 5px;
            margin: 20px 0;
            overflow-x: auto;
        }
        pre {
            background: #2c3e50;
            color: #ecf0f1;
            padding: 15px;
            border-radius: 5px;
            overflow-x: auto;
            font-family: "Monaco", "Menlo", "Ubuntu Mono", monospace;
            font-size: 0.9em;
        }
        .log-entry {
            padding: 5px 0;
            border-bottom: 1px solid #34495e;
        }
        .log-error {
            color: #e74c3c;
        }
        .log-info {
            color: #3498db;
        }
        .footer {
            margin-top: 40px;
            padding-top: 20px;
            border-top: 1px solid #ddd;
            text-align: center;
            color: #7f8c8d;
            font-size: 0.9em;
        }
    </style>
</head>
`)
}

// writeTitle writes the report title.
func (g *HTMLGenerator) writeTitle(sb *strings.Builder, data *report.GenerateContext) {
	title := data.Config.Title
	if title == "" {
		title = fmt.Sprintf("Benchmark Report - %s", data.RunID)
	}
	sb.WriteString(fmt.Sprintf("<h1>%s</h1>\n", title))
}

// writeSummary writes the summary section.
func (g *HTMLGenerator) writeSummary(sb *strings.Builder, data *report.GenerateContext) {
	status := `<span class="status-success">✅ Completed</span>`
	if data.IsFailed() {
		status = `<span class="status-failed">❌ Failed</span>`
	}

	sb.WriteString(`<div class="summary">`)
	sb.WriteString(fmt.Sprintf(`<div class="summary-item"><span class="summary-label">Status:</span> %s</div>`, status))
	sb.WriteString(fmt.Sprintf(`<div class="summary-item"><span class="summary-label">Tool:</span> %s</div>`, data.Tool))
	sb.WriteString(fmt.Sprintf(`<div class="summary-item"><span class="summary-label">Template:</span> %s</div>`, data.TemplateName))
	sb.WriteString(fmt.Sprintf(`<div class="summary-item"><span class="summary-label">Database:</span> %s (%s)</div>`, data.ConnectionName, data.ConnectionType))
	sb.WriteString(fmt.Sprintf(`<div class="summary-item"><span class="summary-label">Duration:</span> %s</div>`, data.GetDuration()))
	sb.WriteString(fmt.Sprintf(`<div class="summary-item"><span class="summary-label">Started:</span> %s</div>`, report.GetTimestamp(data.StartedAt)))
	sb.WriteString(fmt.Sprintf(`<div class="summary-item"><span class="summary-label">Completed:</span> %s</div>`, report.GetTimestamp(data.CompletedAt)))

	if data.IsFailed() {
		sb.WriteString(fmt.Sprintf(`<div class="summary-item"><span class="summary-label">Error:</span> %s</div>`, data.ErrorMessage))
	}

	sb.WriteString(`</div>`)
}

// writeEnvironment writes the environment section.
func (g *HTMLGenerator) writeEnvironment(sb *strings.Builder, data *report.GenerateContext) {
	sb.WriteString(`<h2>Environment</h2>`)
	sb.WriteString(`<table>`)
	sb.WriteString(`<tr><th>Property</th><th>Value</th></tr>`)
	sb.WriteString(fmt.Sprintf(`<tr><td>Run ID</td><td><code>%s</code></td></tr>`, data.RunID))
	sb.WriteString(fmt.Sprintf(`<tr><td>Task ID</td><td><code>%s</code></td></tr>`, data.TaskID))
	sb.WriteString(fmt.Sprintf(`<tr><td>State</td><td>%s</td></tr>`, data.State))
	sb.WriteString(fmt.Sprintf(`<tr><td>Created</td><td>%s</td></tr>`, data.CreatedAt.Format(time.RFC1123)))
	sb.WriteString(`</table>`)
}

// writeParameters writes the parameters section.
func (g *HTMLGenerator) writeParameters(sb *strings.Builder, data *report.GenerateContext) {
	if len(data.Parameters) == 0 {
		return
	}

	sb.WriteString(`<h2>Parameters</h2>`)
	sb.WriteString(`<table>`)
	sb.WriteString(`<tr><th>Parameter</th><th>Value</th></tr>`)
	for key, value := range data.Parameters {
		sb.WriteString(fmt.Sprintf(`<tr><td><strong>%s</strong></td><td>%v</td></tr>`, key, value))
	}
	sb.WriteString(`</table>`)
}

// writeMetrics writes the metrics section.
func (g *HTMLGenerator) writeMetrics(sb *strings.Builder, data *report.GenerateContext) {
	sb.WriteString(`<h2>Metrics</h2>`)

	if !data.HasMetrics() {
		sb.WriteString(`<p><em>No metrics available</em></p>`)
		return
	}

	sb.WriteString(`<div class="metrics">`)
	sb.WriteString(fmt.Sprintf(`<div class="metric-card"><div class="metric-label">TPS</div><div class="metric-value">%.2f</div></div>`, data.TPS))
	sb.WriteString(fmt.Sprintf(`<div class="metric-card"><div class="metric-label">Avg Latency</div><div class="metric-value">%.2f ms</div></div>`, data.LatencyAvg))
	if data.LatencyP95 > 0 {
		sb.WriteString(fmt.Sprintf(`<div class="metric-card"><div class="metric-label">P95 Latency</div><div class="metric-value">%.2f ms</div></div>`, data.LatencyP95))
	}
	if data.LatencyP99 > 0 {
		sb.WriteString(fmt.Sprintf(`<div class="metric-card"><div class="metric-label">P99 Latency</div><div class="metric-value">%.2f ms</div></div>`, data.LatencyP99))
	}
	sb.WriteString(fmt.Sprintf(`<div class="metric-card"><div class="metric-label">Transactions</div><div class="metric-value">%d</div></div>`, data.TotalTransactions))
	if data.TotalQueries > 0 {
		sb.WriteString(fmt.Sprintf(`<div class="metric-card"><div class="metric-label">Queries</div><div class="metric-value">%d</div></div>`, data.TotalQueries))
	}
	sb.WriteString(fmt.Sprintf(`<div class="metric-card"><div class="metric-label">Error Rate</div><div class="metric-value">%.2f%%</div></div>`, data.ErrorRate))
	sb.WriteString(`</div>`)
}

// writeCharts writes the charts section.
func (g *HTMLGenerator) writeCharts(sb *strings.Builder, data *report.GenerateContext) {
	sb.WriteString(`<h2>Charts</h2>`)

	width := data.Config.ChartWidth
	height := data.Config.ChartHeight

	if tpsChart := g.chartGen.GenerateTPSSparkline(data.Samples, width, height); tpsChart != "" {
		sb.WriteString(`<h3>TPS Over Time</h3>`)
		sb.WriteString(`<div class="chart-container"><pre>`)
		sb.WriteString(tpsChart)
		sb.WriteString(`</pre></div>`)
	}

	if latencyChart := g.chartGen.GenerateLatencyDistribution(data.Samples, width); latencyChart != "" {
		sb.WriteString(`<h3>Latency Distribution</h3>`)
		sb.WriteString(`<div class="chart-container"><pre>`)
		sb.WriteString(latencyChart)
		sb.WriteString(`</pre></div>`)
	}
}

// writeTimeSeries writes the time series data section.
func (g *HTMLGenerator) writeTimeSeries(sb *strings.Builder, data *report.GenerateContext) {
	sb.WriteString(`<h2>Time Series Data</h2>`)
	sb.WriteString(`<table>`)
	sb.WriteString(`<tr><th>Timestamp</th><th>TPS</th><th>Latency (ms)</th><th>P95 (ms)</th><th>P99 (ms)</th><th>Error Rate (%)</th></tr>`)

	for _, sample := range data.Samples {
		sb.WriteString(fmt.Sprintf(`<tr><td>%s</td><td>%.2f</td><td>%.2f</td><td>%.2f</td><td>%.2f</td><td>%.2f</td></tr>`,
			sample.Timestamp.Format("15:04:05"),
			sample.TPS,
			sample.LatencyAvg,
			sample.LatencyP95,
			sample.LatencyP99,
			sample.ErrorRate,
		))
	}
	sb.WriteString(`</table>`)
}

// writeLogs writes the logs section.
func (g *HTMLGenerator) writeLogs(sb *strings.Builder, data *report.GenerateContext) {
	sb.WriteString(`<h2>Logs</h2>`)

	// Group by stream
	stderrLogs := make([]string, 0)
	stdoutLogs := make([]string, 0)

	for _, log := range data.Logs {
		if log.Stream == "stderr" {
			stderrLogs = append(stderrLogs, log.Content)
		} else {
			stdoutLogs = append(stdoutLogs, log.Content)
		}
	}

	// Write stderr first
	if len(stderrLogs) > 0 {
		sb.WriteString(`<h3>Errors</h3>`)
		sb.WriteString(`<pre>`)
		for _, log := range stderrLogs {
			sb.WriteString(fmt.Sprintf(`<div class="log-entry log-error">%s</div>`, log))
		}
		sb.WriteString(`</pre>`)
	}

	// Write stdout (limit to 100 lines)
	if len(stdoutLogs) > 0 {
		sb.WriteString(`<h3>Output</h3>`)
		sb.WriteString(`<pre>`)
		start := 0
		if len(stdoutLogs) > 100 {
			start = len(stdoutLogs) - 100
			sb.WriteString(fmt.Sprintf(`<div class="log-entry">... showing last 100 of %d lines ...</div>`, len(stdoutLogs)))
		}
		for _, log := range stdoutLogs[start:] {
			sb.WriteString(fmt.Sprintf(`<div class="log-entry log-info">%s</div>`, log))
		}
		sb.WriteString(`</pre>`)
	}
}

// writeFooter writes the report footer.
func (g *HTMLGenerator) writeFooter(sb *strings.Builder) {
	sb.WriteString(`<div class="footer">`)
	sb.WriteString(fmt.Sprintf("<p>Generated by DB-BenchMind at %s</p>", time.Now().Format(time.RFC1123)))
	sb.WriteString(`</div>`)
}
