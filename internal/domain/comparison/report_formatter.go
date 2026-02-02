// Package comparison provides report formatting functions.
// This file implements Markdown and TXT report generators following the
// professional comparison report template.
package comparison

import (
	"fmt"
	"strings"
)

// FormatMarkdown generates a comprehensive Markdown comparison report.
// This implements the professional template provided by the user.
func (r *ComparisonReport) FormatMarkdown() string {
	if r == nil {
		return "# Comparison Report\n\nNo data available"
	}

	var builder strings.Builder

	// Header
	builder.WriteString("# Sysbench Multi-Configuration Comparison Report\n\n")
	builder.WriteString(fmt.Sprintf("* **Generated at:** %s\n", r.GeneratedAt.Format("2006-01-02 15:04:05")))
	builder.WriteString(fmt.Sprintf("* **Group by:** %s\n", r.GroupBy))
	builder.WriteString(fmt.Sprintf("* **Report ID:** %s\n", r.ReportID))
	builder.WriteString("\n---\n\n")

	// 1) Experiment Metadata
	builder.WriteString(r.formatMetadataMarkdown())

	// 2) Experiment Matrix
	builder.WriteString(r.formatExperimentMatrixMarkdown())

	// 3) Main Comparison (Run Summary Metrics)
	builder.WriteString(r.formatThroughputLatencyMarkdown())
	builder.WriteString(r.formatReliabilityMarkdown())
	builder.WriteString(r.formatQueryMixMarkdown())

	// 4) Steady-state Comparison (if time series data available)
	// TODO: Add when time series data is implemented

	// 5) Scaling & Efficiency Analysis
	builder.WriteString(r.formatScalingAnalysisMarkdown())

	// 6) Visuals (ASCII Charts)
	builder.WriteString(r.formatVisualsMarkdown())

	// 7) Sanity Checks
	if r.SanityChecks != nil {
		builder.WriteString(FormatSanityCheckTable(r.SanityChecks))
	}

	// 8) Findings & Recommendations
	builder.WriteString(r.formatFindingsMarkdown())

	return builder.String()
}

// formatMetadataMarkdown formats the experiment metadata section.
func (r *ComparisonReport) formatMetadataMarkdown() string {
	var builder strings.Builder

	builder.WriteString("## 1) Experiment Metadata\n\n")

	// Basic info
	builder.WriteString("### 1.1 Basic Information\n\n")
	builder.WriteString("| Item | Value |\n")
	builder.WriteString("|------|-------|\n")
	builder.WriteString(fmt.Sprintf("| Report ID | %s |\n", r.ReportID))
	builder.WriteString(fmt.Sprintf("| Generated | %s |\n", r.GeneratedAt.Format("2006-01-02 15:04:05")))
	builder.WriteString(fmt.Sprintf("| Group By | %s |\n", r.GroupBy))
	builder.WriteString(fmt.Sprintf("| Config Groups | %d |\n", len(r.ConfigGroups)))
	if r.SimilarityConfig != nil {
		builder.WriteString(fmt.Sprintf("| Time Window | %s |\n", r.SimilarityConfig.TimeWindow))
	}
	builder.WriteString("\n")

	// Measurement policy
	builder.WriteString("### 1.3 Measurement Policy\n\n")
	builder.WriteString("* **Report interval:** 1s\n")
	builder.WriteString("* **Test duration:** Varies by run\n")
	builder.WriteString("* **Runs per config (N):** Varies by config\n")
	builder.WriteString("* **Execution order:** Based on start time\n")

	// Check for errors across all groups
	totalErrors, totalReconnects, _ := CalculateOverallReliability(r.ConfigGroups)
	acceptanceCriteria := "errors=0 && reconnects=0"
	if totalErrors > 0 || totalReconnects > 0 {
		acceptanceCriteria = "FAILED - Some runs have errors"
	}
	builder.WriteString(fmt.Sprintf("* **Acceptance criteria:** %s\n", acceptanceCriteria))
	builder.WriteString("\n")

	return builder.String()
}

// formatExperimentMatrixMarkdown formats the experiment matrix table.
func (r *ComparisonReport) formatExperimentMatrixMarkdown() string {
	var builder strings.Builder

	builder.WriteString("## 2) Experiment Matrix\n\n")
	builder.WriteString("| Config ID | threads | Database | Template | Runs (N) | Tags |\n")
	builder.WriteString("|---------:|-------:|---------|----------|--------:|------|\n")

	for _, group := range r.ConfigGroups {
		tags := ""
		if r.ScalingAnalysis != nil {
			if r.ScalingAnalysis.BestTPSConfig == group {
				tags += "best-tps "
			}
			if r.ScalingAnalysis.WorstLatencyConfig == group {
				tags += "worst-latency "
			}
		}

		// Check if baseline
		if group.Config.Threads == 1 {
			if tags != "" {
				tags += " "
			}
			tags += "baseline"
		}

		builder.WriteString(fmt.Sprintf("| %s | %d | %s | %s | %d | %s |\n",
			group.GroupID,
			group.Config.Threads,
			group.Config.DatabaseType,
			group.Config.TemplateName,
			len(group.Runs),
			tags))
	}

	builder.WriteString("\n")
	builder.WriteString("> **Definitions:**\n")
	builder.WriteString("> * **Run Summary Metrics** = sysbench summary statistics (SQL statistics, Latency, etc.)\n")
	builder.WriteString("> * **Config ID** = Unique identifier for each configuration group (C1, C2, C3...)\n")
	builder.WriteString("> * **Runs (N)** = Number of benchmark executions for this configuration\n")
	builder.WriteString("\n")

	return builder.String()
}

// formatThroughputLatencyMarkdown formats the throughput and latency summary table.
func (r *ComparisonReport) formatThroughputLatencyMarkdown() string {
	var builder strings.Builder

	builder.WriteString("## 3) Main Comparison (Run Summary Metrics)\n\n")
	builder.WriteString("> **Note:** If N=1, StdDev = N/A; Min=Avg=Max=Single value\n")
	builder.WriteString("> Latency unit: milliseconds\n\n")

	builder.WriteString("### 3.1 Throughput & Latency Summary\n\n")
	builder.WriteString("| threads | N | TPS (mean ± sd) | TPS (min..max) | QPS (mean ± sd) | QPS (min..max) | Lat avg ms (mean ± sd) | Lat p95 ms (mean ± sd) | Lat max ms (max-of-max) |\n")
	builder.WriteString("|-------:|:-:|---------------:|--------------:|---------------:|--------------:|----------------------:|----------------------:|-----------------------:|\n")

	for _, group := range r.ConfigGroups {
		builder.WriteString(fmt.Sprintf("| %d | %d | %s | %s | %s | %s | %s | %s | %.2f |\n",
			group.Config.Threads,
			group.Statistics.N,
			group.Statistics.TPS.FormatMeanStdDev(),
			group.Statistics.TPS.FormatMinMax(),
			group.Statistics.QPS.FormatMeanStdDev(),
			group.Statistics.QPS.FormatMinMax(),
			group.Statistics.LatencyAvg.FormatMeanStdDev(),
			group.Statistics.LatencyP95.FormatMeanStdDev(),
			group.Statistics.LatencyMax,
		))
	}

	builder.WriteString("\n")

	return builder.String()
}

// formatReliabilityMarkdown formats the reliability table.
func (r *ComparisonReport) formatReliabilityMarkdown() string {
	var builder strings.Builder

	builder.WriteString("### 3.2 Reliability\n\n")
	builder.WriteString("| threads | N | Total Errors | Total Reconnects | Any non-zero? |\n")
	builder.WriteString("|-------:|:-:|------------:|---------------:|:-------------|\n")

	for _, group := range r.ConfigGroups {
		anyNonZero := "NO"
		if group.Statistics.HasErrors {
			anyNonZero = "YES"
		}

		builder.WriteString(fmt.Sprintf("| %d | %d | %d | %d | %s |\n",
			group.Config.Threads,
			group.Statistics.N,
			group.Statistics.TotalErrors,
			group.Statistics.TotalReconnects,
			anyNonZero))
	}

	builder.WriteString("\n")

	return builder.String()
}

// formatQueryMixMarkdown formats the query mix table.
func (r *ComparisonReport) formatQueryMixMarkdown() string {
	var builder strings.Builder

	builder.WriteString("### 3.3 Actual Query Mix (from SQL statistics)\n\n")
	builder.WriteString("| threads | Read %% | Write %% | Other %% | Queries / Transaction |\n")
	builder.WriteString("|-------:|------:|-------:|-------:|--------------------:|\n")

	for _, group := range r.ConfigGroups {
		builder.WriteString(fmt.Sprintf("| %d | %.1f | %.1f | %.1f | %.2f |\n",
			group.Config.Threads,
			group.Statistics.ReadPct,
			group.Statistics.WritePct,
			group.Statistics.OtherPct,
			group.Statistics.QueriesPerTx))
	}

	builder.WriteString("\n")

	return builder.String()
}

// formatScalingAnalysisMarkdown formats the scaling analysis section.
func (r *ComparisonReport) formatScalingAnalysisMarkdown() string {
	if r.ScalingAnalysis == nil {
		return ""
	}

	var builder strings.Builder

	builder.WriteString("## 5) Scaling & Efficiency (Threads Analysis)\n\n")

	if r.ScalingAnalysis.BaselineTPS > 0 {
		builder.WriteString(fmt.Sprintf("**Baseline:** threads=1 (TPS=%.2f)\n\n", r.ScalingAnalysis.BaselineTPS))
	}

	builder.WriteString("| threads | TPS_mean | Speedup | Efficiency (Speedup / threads) | ΔTPS vs prev | Δp95 latency |\n")
	builder.WriteString("|-------:|--------:|-------:|-------------------------------:|------------:|-------------:|\n")

	for _, group := range r.ConfigGroups {
		metrics, exists := r.ScalingAnalysis.ByGroup[group.GroupID]
		if !exists {
			continue
		}

		deltaTPS := "—"
		deltaP95 := "—"

		if group.Config.Threads > 1 {
			deltaTPS = fmt.Sprintf("%.2f", metrics.DeltaTPS)
			deltaP95 = fmt.Sprintf("%.2f", metrics.DeltaP95)
		}

		builder.WriteString(fmt.Sprintf("| %d | %.2f | %.2f | %.2f | %s | %s |\n",
			group.Config.Threads,
			group.Statistics.TPS.Mean,
			metrics.Speedup,
			metrics.Efficiency,
			deltaTPS,
			deltaP95))
	}

	builder.WriteString("\n")

	// Add interpretation
	if r.ScalingAnalysis.ScalingKnee != nil {
		builder.WriteString("**Analysis:**\n")
		builder.WriteString(fmt.Sprintf("- **Best throughput:** threads=%d (TPS=%.2f)\n",
			r.ScalingAnalysis.BestTPSConfig.Config.Threads,
			r.ScalingAnalysis.BestTPSConfig.Statistics.TPS.Mean))
		builder.WriteString(fmt.Sprintf("- **Scaling knee:** threads=~%d (efficiency drops significantly)\n",
			r.ScalingAnalysis.ScalingKneeThread))
		builder.WriteString("\n")
	}

	return builder.String()
}

// formatVisualsMarkdown formats ASCII visual charts.
func (r *ComparisonReport) formatVisualsMarkdown() string {
	var builder strings.Builder

	builder.WriteString("## 6) Visuals (ASCII Charts)\n\n")

	// TPS vs Threads chart
	builder.WriteString("### 6.1 TPS vs Threads\n")
	builder.WriteString("```\n")
	maxTPS := 0.0
	for _, group := range r.ConfigGroups {
		if group.Statistics.TPS.Mean > maxTPS {
			maxTPS = group.Statistics.TPS.Mean
		}
	}

	for _, group := range r.ConfigGroups {
		tps := group.Statistics.TPS.Mean
		barWidth := 50
		barLength := int((tps / maxTPS) * float64(barWidth))
		if barLength < 1 {
			barLength = 1
		}
		bar := strings.Repeat("█", barLength)
		spaces := strings.Repeat(" ", barWidth-barLength)

		builder.WriteString(fmt.Sprintf("threads=%-2d |%s%s %.2f\n",
			group.Config.Threads, bar, spaces, tps))
	}
	builder.WriteString("```\n\n")

	// p95 Latency vs Threads chart
	builder.WriteString("### 6.2 p95 Latency vs Threads\n")
	builder.WriteString("```\n")
	maxP95 := 0.0
	for _, group := range r.ConfigGroups {
		if group.Statistics.LatencyP95.Mean > maxP95 {
			maxP95 = group.Statistics.LatencyP95.Mean
		}
	}

	for _, group := range r.ConfigGroups {
		p95 := group.Statistics.LatencyP95.Mean
		barWidth := 50
		barLength := int((p95 / maxP95) * float64(barWidth))
		if barLength < 1 {
			barLength = 1
		}
		bar := strings.Repeat("█", barLength)
		spaces := strings.Repeat(" ", barWidth-barLength)

		builder.WriteString(fmt.Sprintf("threads=%-2d |%s%s %.2fms\n",
			group.Config.Threads, bar, spaces, p95))
	}
	builder.WriteString("```\n\n")

	return builder.String()
}

// formatFindingsMarkdown formats the findings and recommendations section.
func (r *ComparisonReport) formatFindingsMarkdown() string {
	var builder strings.Builder

	builder.WriteString("## 8) Findings & Recommendations\n\n")

	if r.Findings != nil {
		builder.WriteString("### 8.1 Key Findings\n\n")

		if r.Findings.BestThroughput != "" {
			builder.WriteString(fmt.Sprintf("* **Best throughput point:** %s\n", r.Findings.BestThroughput))
		}
		if r.Findings.ScalingKnee != "" {
			builder.WriteString(fmt.Sprintf("* **Scaling knee:** %s\n", r.Findings.ScalingKnee))
		}
		if r.Findings.LatencyRisk != "" {
			builder.WriteString(fmt.Sprintf("* **Latency risk point:** %s\n", r.Findings.LatencyRisk))
		}
		if r.Findings.StabilityConcerns != "" {
			builder.WriteString(fmt.Sprintf("* **Stability:** %s\n", r.Findings.StabilityConcerns))
		}

		builder.WriteString("\n")

		builder.WriteString("### 8.2 Recommendation\n\n")
		builder.WriteString(fmt.Sprintf("**Suggested:** %s\n\n", r.Findings.Recommendation))

		if r.Findings.TradeoffStatement != "" {
			builder.WriteString(fmt.Sprintf("**Trade-off:** %s\n\n", r.Findings.TradeoffStatement))
		}

		if r.Findings.NextExperiment != "" {
			builder.WriteString(fmt.Sprintf("**Next experiment:** %s\n\n", r.Findings.NextExperiment))
		}
	}

	return builder.String()
}

// FormatTXT generates a plain text version of the comparison report.
func (r *ComparisonReport) FormatTXT() string {
	if r == nil {
		return "COMPARISON REPORT\n\nNo data available"
	}

	var builder strings.Builder

	builder.WriteString("╔══════════════════════════════════════════════════════════════════╗\n")
	builder.WriteString("║        SYSBENCH MULTI-CONFIGURATION COMPARISON REPORT            ║\n")
	builder.WriteString("╚══════════════════════════════════════════════════════════════════╝\n\n")

	builder.WriteString(fmt.Sprintf("Generated: %s\n", r.GeneratedAt.Format("2006-01-02 15:04:05")))
	builder.WriteString(fmt.Sprintf("Report ID: %s\n", r.ReportID))
	builder.WriteString(fmt.Sprintf("Group By: %s\n", r.GroupBy))
	builder.WriteString(fmt.Sprintf("Config Groups: %d\n\n", len(r.ConfigGroups)))

	// Experiment Matrix
	builder.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	builder.WriteString("2) EXPERIMENT MATRIX\n")
	builder.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	for _, group := range r.ConfigGroups {
		builder.WriteString(fmt.Sprintf("Config %s: threads=%d, database=%s, runs=%d\n",
			group.GroupID, group.Config.Threads, group.Config.DatabaseType, len(group.Runs)))
	}
	builder.WriteString("\n")

	// Throughput & Latency
	builder.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	builder.WriteString("3) THROUGHPUT & LATENCY SUMMARY\n")
	builder.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	for _, group := range r.ConfigGroups {
		builder.WriteString(fmt.Sprintf("threads=%d (N=%d):\n", group.Config.Threads, group.Statistics.N))
		builder.WriteString(fmt.Sprintf("  TPS:  %s\n", group.Statistics.TPS.FormatMeanStdDev()))
		builder.WriteString(fmt.Sprintf("  QPS:  %s\n", group.Statistics.QPS.FormatMeanStdDev()))
		builder.WriteString(fmt.Sprintf("  Lat:  avg=%s, p95=%s, max=%.2f\n\n",
			group.Statistics.LatencyAvg.FormatMeanStdDev(),
			group.Statistics.LatencyP95.FormatMeanStdDev(),
			group.Statistics.LatencyMax))
	}

	// Scaling Analysis
	if r.ScalingAnalysis != nil {
		builder.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		builder.WriteString("4) SCALING ANALYSIS\n")
		builder.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

		for _, group := range r.ConfigGroups {
			metrics := r.ScalingAnalysis.ByGroup[group.GroupID]
			builder.WriteString(fmt.Sprintf("threads=%d: speedup=%.2fx, efficiency=%.2f%%\n",
				group.Config.Threads, metrics.Speedup, metrics.Efficiency*100))
		}
		builder.WriteString("\n")
	}

	// Sanity Checks
	if r.SanityChecks != nil {
		builder.WriteString(FormatSanityCheckTable(r.SanityChecks))
	}

	// Findings
	if r.Findings != nil && r.Findings.Recommendation != "" {
		builder.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		builder.WriteString("5) RECOMMENDATION\n")
		builder.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
		builder.WriteString(r.Findings.Recommendation)
		builder.WriteString("\n")
	}

	builder.WriteString("═══════════════════════════════════════════════════════════════════\n")

	return builder.String()
}

// GenerateReportFindings auto-generates findings from the analysis results.
func GenerateReportFindings(report *ComparisonReport) *ReportFindings {
	if report == nil || len(report.ConfigGroups) == 0 {
		return nil
	}

	findings := &ReportFindings{}

	// Best throughput
	if report.ScalingAnalysis != nil && report.ScalingAnalysis.BestTPSConfig != nil {
		best := report.ScalingAnalysis.BestTPSConfig
		findings.BestThroughput = fmt.Sprintf("threads=%d (TPS=%.2f, p95=%.2fms)",
			best.Config.Threads, best.Statistics.TPS.Mean, best.Statistics.LatencyP95.Mean)
	}

	// Scaling knee
	if report.ScalingAnalysis != nil && report.ScalingAnalysis.ScalingKnee != nil {
		findings.ScalingKnee = fmt.Sprintf("threads=~%d (efficiency drops significantly)",
			report.ScalingAnalysis.ScalingKneeThread)
	}

	// Latency risk
	if report.ScalingAnalysis != nil && report.ScalingAnalysis.WorstLatencyConfig != nil {
		worst := report.ScalingAnalysis.WorstLatencyConfig
		findings.LatencyRisk = fmt.Sprintf("threads=%d (p95=%.2fms - highest latency)",
			worst.Config.Threads, worst.Statistics.LatencyP95.Mean)
	}

	// Stability concerns
	var unstableConfigs []string
	for _, group := range report.ConfigGroups {
		cv := CalculateCV(group.Statistics.TPS.Mean, group.Statistics.TPS.StdDev)
		if cv > 10 {
			unstableConfigs = append(unstableConfigs,
				fmt.Sprintf("threads=%d (CV=%.2f%%)", group.Config.Threads, cv))
		}
	}
	if len(unstableConfigs) > 0 {
		findings.StabilityConcerns = strings.Join(unstableConfigs, "; ")
	} else {
		findings.StabilityConcerns = "All configs stable (CV < 10%)"
	}

	// Recommendation
	if report.ScalingAnalysis != nil {
		optimalThreads := CalculateOptimalThreadCount(report.ScalingAnalysis, report.ConfigGroups)
		findings.Recommendation = fmt.Sprintf("threads=%d", optimalThreads)

		// Find optimal group
		var optimalGroup *ConfigGroup
		for _, group := range report.ConfigGroups {
			if group.Config.Threads == optimalThreads {
				optimalGroup = group
				break
			}
		}

		if optimalGroup != nil {
			metrics := report.ScalingAnalysis.ByGroup[optimalGroup.GroupID]
			findings.TradeoffStatement = fmt.Sprintf(
				"%.2fx speedup with %.2f%% scaling efficiency at %.2fms p95 latency",
				metrics.Speedup, metrics.Efficiency*100, optimalGroup.Statistics.LatencyP95.Mean)
		}
	}

	// Next experiment
	findings.NextExperiment = "Repeat with N=5 runs per config for better statistics"

	return findings
}
