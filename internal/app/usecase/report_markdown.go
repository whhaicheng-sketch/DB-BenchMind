// Package usecase provides markdown report generation logic.
package usecase

import (
	"fmt"
	"strings"
	"time"

	"github.com/whhaicheng/DB-BenchMind/internal/domain/report"
)

// MarkdownGenerator generates human-readable markdown reports.
type MarkdownGenerator struct{}

// NewMarkdownGenerator creates a new MarkdownGenerator.
func NewMarkdownGenerator() *MarkdownGenerator {
	return &MarkdownGenerator{}
}

// Generate creates a markdown report from HumanReportData.
func (g *MarkdownGenerator) Generate(data *report.HumanReportData) string {
	var b strings.Builder

	g.writeTitle(&b, data)
	g.writeOverallStatus(&b, data)
	g.writeOneLineConclusion(&b, data)
	g.writeCoreKPIs(&b, data)
	g.writeResourceBottleneck(&b, data)
	g.writeBottleneckJudgment(&b, data)
	g.writeComparison(&b, data)
	g.writeTrendHighlights(&b, data)
	g.writeRecommendation(&b, data)
	g.writeDetailedDataPreview(&b, data)

	return b.String()
}

func (g *MarkdownGenerator) writeTitle(b *strings.Builder, data *report.HumanReportData) {
	meta := data.RunMeta
	b.WriteString(fmt.Sprintf("# Benchmark Report\n\n"))
	b.WriteString(fmt.Sprintf("**Run ID:** %s  \n", naIfEmpty(meta.RunID)))
	b.WriteString(fmt.Sprintf("**Benchmark Type:** %s  \n", naIfEmpty(meta.BenchmarkType)))
	b.WriteString(fmt.Sprintf("**Template:** %s  \n", naIfEmpty(meta.TemplateName)))
	b.WriteString(fmt.Sprintf("**Target:** %s  \n", naIfEmpty(meta.TargetName)))
	b.WriteString(fmt.Sprintf("**Start Time:** %s  \n", naIfEmpty(meta.StartTime)))
	b.WriteString(fmt.Sprintf("**End Time:** %s  \n", naIfEmpty(meta.EndTime)))
	b.WriteString(fmt.Sprintf("**Duration:** %s  \n", formatDurationMs(meta.DurationMs)))
	b.WriteString("\n---\n\n")
}

func (g *MarkdownGenerator) writeOverallStatus(b *strings.Builder, data *report.HumanReportData) {
	status := string(data.OverallStatus)
	emoji := statusEmoji(data.OverallStatus)
	b.WriteString(fmt.Sprintf("## Overall Status: %s %s\n\n", strings.ToUpper(status), emoji))
}

func (g *MarkdownGenerator) writeOneLineConclusion(b *strings.Builder, data *report.HumanReportData) {
	b.WriteString(fmt.Sprintf("> %s\n\n", naIfEmpty(data.OneLineConclusion)))
}

func (g *MarkdownGenerator) writeCoreKPIs(b *strings.Builder, data *report.HumanReportData) {
	kpis := data.CoreKPIs
	b.WriteString("## Core KPIs\n\n")
	b.WriteString("| Metric | Value |\n")
	b.WriteString("|--------|-------|\n")
	b.WriteString(fmt.Sprintf("| TPS | %s |\n", formatFloat(kpis.TPS)))
	b.WriteString(fmt.Sprintf("| TPM | %s |\n", formatFloat(kpis.TPM)))
	b.WriteString(fmt.Sprintf("| Avg Latency | %s ms |\n", formatFloat(kpis.AvgLatencyMs)))
	b.WriteString(fmt.Sprintf("| P95 Latency | %s ms |\n", formatFloat(kpis.P95LatencyMs)))
	b.WriteString(fmt.Sprintf("| P99 Latency | %s ms |\n", formatFloat(kpis.P99LatencyMs)))
	b.WriteString(fmt.Sprintf("| Error Rate | %s |\n", formatFloat(kpis.ErrorRate)))
	b.WriteString(fmt.Sprintf("| CPU Peak | %s %% |\n", formatFloat(kpis.CPUPeak)))
	b.WriteString(fmt.Sprintf("| Disk Util Peak | %s %% |\n", formatFloat(kpis.DiskUtilPeak)))
	b.WriteString("\n")
}

func (g *MarkdownGenerator) writeResourceBottleneck(b *strings.Builder, data *report.HumanReportData) {
	rb := data.ResourceBottleneck
	b.WriteString("## Resource Bottleneck Summary\n\n")
	b.WriteString("| Resource | Avg | Peak | Verdict |\n")
	b.WriteString("|----------|-----|------|--------|\n")
	b.WriteString(fmt.Sprintf("| CPU | %s | %s | %s |\n",
		formatFloat(rb.CPU.Avg), formatFloat(rb.CPU.Peak), rb.CPU.Verdict))
	b.WriteString(fmt.Sprintf("| Memory | %s MB | %s MB (%s%%) | %s |\n",
		formatFloat(rb.Memory.AvgMB), formatFloat(rb.Memory.PeakMB),
		formatFloat(rb.Memory.UsedPct), rb.Memory.Verdict))
	b.WriteString(fmt.Sprintf("| Disk | %s | %s | %s |\n",
		formatFloat(rb.Disk.Avg), formatFloat(rb.Disk.Peak), rb.Disk.Verdict))
	b.WriteString(fmt.Sprintf("| Network | Rx: %s / Tx: %s | - | %s |\n",
		formatFloat(rb.Network.RxPeak), formatFloat(rb.Network.TxPeak),
		rb.Network.Verdict))
	b.WriteString("\n")
}

func (g *MarkdownGenerator) writeBottleneckJudgment(b *strings.Builder, data *report.HumanReportData) {
	j := data.BottleneckJudgment
	b.WriteString("## Bottleneck Judgment\n\n")
	b.WriteString(fmt.Sprintf("**Primary Bottleneck:** %s  \n", j.PrimaryBottleneck))
	b.WriteString(fmt.Sprintf("**Confidence:** %.0f%%  \n\n", j.Confidence*100))

	if len(j.Evidence) > 0 {
		b.WriteString("**Evidence:**\n")
		for _, e := range j.Evidence {
			b.WriteString(fmt.Sprintf("- %s\n", e))
		}
		b.WriteString("\n")
	}
}

func (g *MarkdownGenerator) writeComparison(b *strings.Builder, data *report.HumanReportData) {
	if data.Comparison == nil || len(data.Comparison.Deltas) == 0 {
		return
	}

	comp := data.Comparison
	b.WriteString("## Comparison\n\n")
	if comp.PreviousReportID != "" {
		b.WriteString(fmt.Sprintf("**Compared with previous report:** %s  \n\n", comp.PreviousReportID))
	}
	if comp.BaselineReportID != "" {
		b.WriteString(fmt.Sprintf("**Compared with baseline:** %s  \n\n", comp.BaselineReportID))
	}

	b.WriteString("| Metric | Current | Previous | Delta | Change |\n")
	b.WriteString("|--------|---------|----------|-------|--------|\n")
	for _, d := range comp.Deltas {
		arrow := changeArrow(d.PctChange)
		b.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s %.1f%% |\n",
			d.Metric, formatFloat(d.Current), formatFloat(d.Previous),
			formatFloat(d.Delta), arrow, d.PctChange))
	}
	b.WriteString("\n")
}

func (g *MarkdownGenerator) writeTrendHighlights(b *strings.Builder, data *report.HumanReportData) {
	if len(data.TrendHighlights) == 0 {
		return
	}
	b.WriteString("## Trend Highlights\n\n")
	for _, h := range data.TrendHighlights {
		b.WriteString(fmt.Sprintf("- %s\n", h))
	}
	b.WriteString("\n")
}

func (g *MarkdownGenerator) writeRecommendation(b *strings.Builder, data *report.HumanReportData) {
	b.WriteString("## Recommendation\n\n")
	if data.Recommendation != "" {
		b.WriteString(data.Recommendation)
	} else {
		b.WriteString("N/A")
	}
	b.WriteString("\n\n")
}

func (g *MarkdownGenerator) writeDetailedDataPreview(b *strings.Builder, data *report.HumanReportData) {
	b.WriteString("## Detailed Data\n\n")
	b.WriteString("*This section is collapsed by default in the UI.*\n\n")

	if data.DetailedDataPreview == nil {
		b.WriteString("No detailed data available.\n")
		return
	}

	dp := data.DetailedDataPreview
	b.WriteString(fmt.Sprintf("- **Bundle:** %s\n", naIfEmpty(dp.BundleFilename)))
	b.WriteString(fmt.Sprintf("- **Compressed Size:** %s\n", formatBytes(dp.CompressedSize)))
	b.WriteString(fmt.Sprintf("- **Sampling Policy:** %s\n", dp.SamplingPolicy))

	if len(dp.RetainedWindows) > 0 {
		b.WriteString("- **Retained Windows:**\n")
		for _, w := range dp.RetainedWindows {
			b.WriteString(fmt.Sprintf("  - %s (%d samples)\n", w.Name, w.SampleCount))
		}
	}

	if len(dp.AnomalyWindows) > 0 {
		b.WriteString("- **Anomaly Windows:**\n")
		for _, a := range dp.AnomalyWindows {
			b.WriteString(fmt.Sprintf("  - [%s] %s: %s\n", a.Severity, a.Type, a.Summary))
		}
	}

	b.WriteString(fmt.Sprintf("- **Raw Samples:** %d\n", dp.RawSampleCount))
}

// Helper functions

func naIfEmpty(s string) string {
	if s == "" {
		return "N/A"
	}
	return s
}

func formatFloat(v float64) string {
	if v == 0 {
		return "N/A"
	}
	return fmt.Sprintf("%.2f", v)
}

func formatDurationMs(ms int64) string {
	if ms <= 0 {
		return "N/A"
	}
	d := time.Duration(ms) * time.Millisecond
	if d >= time.Hour {
		return fmt.Sprintf("%dh %dm %ds", int(d.Hours()), int(d.Minutes())%60, int(d.Seconds())%60)
	}
	if d >= time.Minute {
		return fmt.Sprintf("%dm %ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	return fmt.Sprintf("%.1fs", d.Seconds())
}

func formatBytes(b int64) string {
	if b <= 0 {
		return "N/A"
	}
	const kb = 1024
	const mb = kb * 1024
	if b >= mb {
		return fmt.Sprintf("%.1f MB", float64(b)/float64(mb))
	}
	if b >= kb {
		return fmt.Sprintf("%.1f KB", float64(b)/float64(kb))
	}
	return fmt.Sprintf("%d B", b)
}

func statusEmoji(s report.OverallStatus) string {
	switch s {
	case report.StatusHealthy:
		return "✅"
	case report.StatusWarning:
		return "⚠️"
	case report.StatusCritical:
		return "🔴"
	default:
		return "❓"
	}
}

func changeArrow(pctChange float64) string {
	if pctChange > 5 {
		return "📈"
	}
	if pctChange < -5 {
		return "📉"
	}
	return "➡️"
}
