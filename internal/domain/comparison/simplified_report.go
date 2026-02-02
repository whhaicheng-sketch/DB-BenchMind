// Package comparison provides simplified comparison report implementation.
// This file implements a simplified version of the comprehensive report template
// using existing history_records data without Template Variant complexity.
package comparison

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

// SimplifiedReportFindings contains findings for simplified report.
type SimplifiedReportFindings struct {
	BestTPSThreads     int
	BestTPSValue       float64
	BestLatencyThreads int
	BestLatencyValue   float64
	ScalingKnee        int
	Recommendation     string
}

// SimplifiedReport represents a simplified comparison report.
type SimplifiedReport struct {
	GeneratedAt      time.Time
	ReportID         string
	SelectedRecords  int
	GroupBy          GroupByField
	Records          []*RecordRef
	ConfigGroups     []*ThreadGroup
	SanityChecks     []SanityCheckResult
	Findings         *SimplifiedReportFindings
	Notes            string
}

// ThreadGroup groups records by thread count for analysis.
type ThreadGroup struct {
	Threads      int
	Records     []*RecordRef
	Statistics  ThreadGroupStats
}

// ThreadGroupStats contains statistics for a thread group.
type ThreadGroupStats struct {
	N           int
	TPS         GroupMetricStats
	QPS         GroupMetricStats
	LatencyAvg  GroupMetricStats
	LatencyP95  GroupMetricStats
	Errors      int64
	Reconnects  int64
}

// GroupMetricStats contains statistics across N runs.
type GroupMetricStats struct {
	Mean     float64
	StdDev   float64
	Min      float64
	Max      float64
}

// SanityCheckResult represents a single sanity check result.
type SanityCheckResult struct {
	Name    string
	Passed  bool
	Details string
}

// GenerateSimplifiedReport generates a simplified comparison report from history records.
func GenerateSimplifiedReport(records []*RecordRef, groupBy GroupByField) *SimplifiedReport {
	if len(records) == 0 {
		return nil
	}

	report := &SimplifiedReport{
		GeneratedAt:     time.Now(),
		ReportID:        fmt.Sprintf("report-%s", time.Now().Format("20060102_150405")),
		SelectedRecords: len(records),
		GroupBy:         groupBy,
		Records:         records,
		Notes:           "Simplified report (no Template Variant, no time series)",
	}

	// Group by threads
	report.ConfigGroups = groupByThreads(records)

	// Perform sanity checks
	report.SanityChecks = performSimplifiedChecks(report.ConfigGroups)

	// Generate findings
	report.Findings = generateSimplifiedFindings(report.ConfigGroups)

	return report
}

// groupByThreads groups records by thread count.
func groupByThreads(records []*RecordRef) []*ThreadGroup {
	groups := make(map[int]*ThreadGroup)

	for _, record := range records {
		threads := record.Threads
		if groups[threads] == nil {
			groups[threads] = &ThreadGroup{
				Threads:  threads,
				Records: []*RecordRef{},
			}
		}
		groups[threads].Records = append(groups[threads].Records, record)
	}

	// Convert map to slice and sort by threads
	var groupList []*ThreadGroup
	for _, group := range groups {
		// Calculate statistics
		group.Statistics = calculateThreadStats(group.Records)
		groupList = append(groupList, group)
	}

	sort.Slice(groupList, func(i, j int) bool {
		return groupList[i].Threads < groupList[j].Threads
	})

	return groupList
}

// calculateThreadStats calculates statistics for a thread group.
func calculateThreadStats(records []*RecordRef) ThreadGroupStats {
	n := len(records)
	stats := ThreadGroupStats{N: n}

	// Collect TPS values
	tpsValues := make([]float64, n)
	qpsValues := make([]float64, n)
	latAvgValues := make([]float64, n)
	latP95Values := make([]float64, n)

	for i, record := range records {
		tpsValues[i] = record.TPS
		qpsValues[i] = record.QPS
		latAvgValues[i] = record.LatencyAvg
		latP95Values[i] = record.LatencyP95
		stats.Errors += record.IgnoredErrors
		stats.Reconnects += record.Reconnects
	}

	// Calculate TPS statistics
	stats.TPS = calculateGroupMetricStats(tpsValues)
	stats.QPS = calculateGroupMetricStats(qpsValues)
	stats.LatencyAvg = calculateGroupMetricStats(latAvgValues)
	stats.LatencyP95 = calculateGroupMetricStats(latP95Values)

	return stats
}

// calculateGroupMetricStats calculates statistics for a metric.
func calculateGroupMetricStats(values []float64) GroupMetricStats {
	n := len(values)
	if n == 0 {
		return GroupMetricStats{}
	}

	stats := GroupMetricStats{
		Min: values[0],
		Max: values[0],
	}

	// Calculate min and max
	for _, v := range values {
		if v < stats.Min {
			stats.Min = v
		}
		if v > stats.Max {
			stats.Max = v
		}
	}

	// Calculate mean
	var sum float64
	for _, v := range values {
		sum += v
	}
	stats.Mean = sum / float64(n)

	// Calculate stddev
	if n > 1 {
		var varianceSum float64
		for _, v := range values {
			diff := v - stats.Mean
			varianceSum += diff * diff
		}
		stats.StdDev = math.Sqrt(varianceSum / float64(n-1))
	}

	return stats
}

// performSimplifiedChecks performs sanity checks on grouped data.
func performSimplifiedChecks(groups []*ThreadGroup) []SanityCheckResult {
	var checks []SanityCheckResult

	// Check 1: SQL total = read + write + other
	// (Use first record from each group)
	sqlPassed := true
	sqlDetails := ""
	for _, group := range groups {
		if len(group.Records) > 0 {
			record := group.Records[0]
			total := record.ReadQueries + record.WriteQueries + record.OtherQueries
			if total != record.TotalQueries {
				sqlPassed = false
				sqlDetails += fmt.Sprintf("Group %d: total=%d vs calc=%d",
					group.Threads, record.TotalQueries, total)
			}
		}
	}
	checks = append(checks, SanityCheckResult{
		Name:    "SQL total = read + write + other",
		Passed:  sqlPassed,
			Details: sqlDetails,
	})

	// Check 2: QPS ≈ TPS × 20
	qpsPassed := true
	qpsDetails := ""
	for _, group := range groups {
		expectedQPS := group.Statistics.TPS.Mean * 20
		actualQPS := group.Statistics.QPS.Mean
		diff := math.Abs(expectedQPS - actualQPS)
		if expectedQPS > 0 && (diff/expectedQPS) > 0.05 { // 5% tolerance
			qpsPassed = false
			qpsDetails += fmt.Sprintf("Group %d: expected=%.2f, actual=%.2f",
				group.Threads, expectedQPS, actualQPS)
		}
	}
	checks = append(checks, SanityCheckResult{
		Name:    "QPS ≈ TPS × 20",
		Passed:  qpsPassed,
		Details: qpsDetails,
	})

	// Check 3: Latency ordering (min ≤ avg ≤ p95)
	latencyPassed := true
	latencyDetails := ""
	for _, group := range groups {
		if group.Statistics.LatencyAvg.Min > group.Statistics.LatencyAvg.Mean ||
			group.Statistics.LatencyAvg.Mean > group.Statistics.LatencyP95.Mean {
			latencyPassed = false
			latencyDetails += fmt.Sprintf("Group %d: min=%.2f, avg=%.2f, p95=%.2f",
				group.Threads, group.Statistics.LatencyAvg.Min,
				group.Statistics.LatencyAvg.Mean, group.Statistics.LatencyP95.Mean)
		}
	}
	checks = append(checks, SanityCheckResult{
		Name:    "Latency min ≤ avg ≤ p95",
		Passed:  latencyPassed,
		Details: latencyDetails,
	})

	// Check 4: No errors
	errorsPassed := true
	errorsDetails := ""
	for _, group := range groups {
		if group.Statistics.Errors > 0 || group.Statistics.Reconnects > 0 {
			errorsPassed = false
			errorsDetails += fmt.Sprintf("Group %d: errors=%d, reconnects=%d",
				group.Threads, group.Statistics.Errors, group.Statistics.Reconnects)
		}
	}
	checks = append(checks, SanityCheckResult{
		Name:    "errors=0 & reconnects=0",
		Passed:  errorsPassed,
		Details: errorsDetails,
	})

	return checks
}

// generateSimplifiedFindings generates findings from grouped data.
func generateSimplifiedFindings(groups []*ThreadGroup) *SimplifiedReportFindings {
	findings := &SimplifiedReportFindings{}

	// Find best TPS
	var bestTPSGroup *ThreadGroup
	for _, group := range groups {
		if bestTPSGroup == nil || group.Statistics.TPS.Mean > bestTPSGroup.Statistics.TPS.Mean {
			bestTPSGroup = group
		}
	}
	if bestTPSGroup != nil {
		findings.BestTPSThreads = bestTPSGroup.Threads
		findings.BestTPSValue = bestTPSGroup.Statistics.TPS.Mean
	}

	// Find best latency
	var bestLatencyGroup *ThreadGroup
	for _, group := range groups {
		if bestLatencyGroup == nil || group.Statistics.LatencyP95.Mean < bestLatencyGroup.Statistics.LatencyP95.Mean {
			bestLatencyGroup = group
		}
	}
	if bestLatencyGroup != nil {
		findings.BestLatencyThreads = bestLatencyGroup.Threads
		findings.BestLatencyValue = bestLatencyGroup.Statistics.LatencyP95.Mean
	}

	// Identify scaling knee
	if len(groups) > 1 {
		// Find where efficiency drops below 70%
		for i := 1; i < len(groups); i++ {
			group := groups[i]

			speedup := group.Statistics.TPS.Mean / groups[0].Statistics.TPS.Mean
			efficiency := speedup / float64(group.Threads)

			if efficiency < 0.70 {
				findings.ScalingKnee = group.Threads
				break
			}
		}
	}

	// Generate recommendation
	if bestTPSGroup != nil {
		findings.Recommendation = fmt.Sprintf("threads=%d (TPS=%.2f, p95=%.2fms)",
			bestTPSGroup.Threads,
			bestTPSGroup.Statistics.TPS.Mean,
			bestTPSGroup.Statistics.LatencyP95.Mean)
	}

	return findings
}

// FormatMarkdown formats the simplified report as Markdown.
func (r *SimplifiedReport) FormatMarkdown() string {
	if r == nil {
		return ""
	}

	var builder strings.Builder

	// Header
	builder.WriteString("# Sysbench Comparison Report\n\n")
	builder.WriteString(fmt.Sprintf("- **Generated at:** %s\n", r.GeneratedAt.Format("2006-01-02 15:04:05")))
	builder.WriteString(fmt.Sprintf("- **Report ID:** %s\n", r.ReportID))
	builder.WriteString(fmt.Sprintf("- **Group By:** %s\n", r.GroupBy))
	builder.WriteString(fmt.Sprintf("- **Selected Records:** %d\n", r.SelectedRecords))
	builder.WriteString(fmt.Sprintf("- **Notes:** %s\n\n", r.Notes))
	builder.WriteString("---\n\n")

	// Section 0: Record Selection Summary
	builder.WriteString("## 0) Record Selection Summary\n\n")
	builder.WriteString("### 0.1 Filters (UI Inputs)\n")
	builder.WriteString("- Search Query: (none)\n")
	builder.WriteString("- Selected Templates: All templates\n")
	builder.WriteString("- Threads in Selection: ")
	threads := make(map[int]bool)
	for _, group := range r.ConfigGroups {
		threads[group.Threads] = true
	}
	threadList := make([]int, 0, len(threads))
	for t := range threads {
		threadList = append(threadList, t)
	}
	sort.Ints(threadList)
	for i, t := range threadList {
		if i > 0 {
			builder.WriteString(", ")
		}
		builder.WriteString(fmt.Sprintf("%d", t))
	}
	builder.WriteString("\n")

	// Section 1: Parsing & Sanity Checks
	builder.WriteString("## 1) Parsing & Sanity Checks (Global)\n\n")
	builder.WriteString("### 1.1 Global Parse Summary\n")
	builder.WriteString("| Item | Value |\n")
	builder.WriteString("|------|-------|\n")
	builder.WriteString(fmt.Sprintf("| Runs parsed successfully | %d |\n", r.SelectedRecords))
	builder.WriteString("| Runs failed to parse | 0 |\n")
	builder.WriteString("| Unknown/Unparsed lines (count) | 0 |\n\n")

	builder.WriteString("### 1.2 Sanity Checks (Must Pass)\n")
	builder.WriteString("| Check | Result | Details |\n")
	builder.WriteString("|------|--------|----------|\n")
	for _, check := range r.SanityChecks {
		result := "✅ PASS"
		if !check.Passed {
			result = "❌ FAIL"
		}
		details := check.Details
		if len(details) > 50 {
			details = details[:47] + "..."
		}
		builder.WriteString(fmt.Sprintf("| %s | %s | %s |\n",
			check.Name, result, details))
	}
	builder.WriteString("\n")

	// Section 2: Executive Summary
	builder.WriteString("## 2) Executive Summary (Across All Included Runs)\n\n")
	builder.WriteString("### 2.1 Best Points (by common criteria)\n")
	if r.Findings != nil {
		builder.WriteString(fmt.Sprintf("- Highest TPS (run-summary mean): **threads=%d** → TPS=%.2f\n",
			r.Findings.BestTPSThreads, r.Findings.BestTPSValue))
		if r.Findings.BestLatencyThreads > 0 {
			builder.WriteString(fmt.Sprintf("- Lowest p95 latency (run-summary mean): **threads=%d** → p95=%.2f\n",
				r.Findings.BestLatencyThreads, r.Findings.BestLatencyValue))
		}
	}
	builder.WriteString("\n")

	// Section 4: Comparison Sections
	builder.WriteString("## 4) Comparison Sections (Per Thread Count)\n\n")

	for i, group := range r.ConfigGroups {
		builder.WriteString(fmt.Sprintf("## 4.%d Thread Group: threads=%d\n\n", i+1, group.Threads))

		// Experiment Matrix
		builder.WriteString(fmt.Sprintf("### 4.%d.1 Experiment Matrix\n\n", i+1))
		builder.WriteString("| threads | N | date_span | tags |\n")
		builder.WriteString("|-------:|-:|----------|------|\n")
		// Calculate date span
		var earliest, latest time.Time
		for _, record := range group.Records {
			if earliest.IsZero() || record.StartTime.Before(earliest) {
				earliest = record.StartTime
			}
			if latest.IsZero() || record.StartTime.After(latest) {
				latest = record.StartTime
			}
		}
		dateSpan := "N/A"
		if !earliest.IsZero() && !latest.IsZero() {
			dateSpan = fmt.Sprintf("%s to %s",
				earliest.Format("2006-01-02 15:04"),
				latest.Format("2006-01-02 15:04"))
		}

		var tags []string
		if r.Findings != nil && group.Threads == r.Findings.BestTPSThreads {
			tags = append(tags, "best-tps")
		}
		if r.Findings != nil && group.Threads == r.Findings.BestLatencyThreads {
			tags = append(tags, "best-latency")
		}
		tagStr := strings.Join(tags, ", ")
		if tagStr == "" {
			tagStr = "-"
		}

		builder.WriteString(fmt.Sprintf("| %d | %d | %s | %s |\n\n",
			group.Threads, group.Statistics.N, dateSpan, tagStr))

		// Main Comparison
		builder.WriteString(fmt.Sprintf("### 4.%d.2 Main Comparison (Run Summary Metrics)\n\n", i+1))
		builder.WriteString("> Source: sysbench tail summary (SQL statistics / General statistics / Latency)\n\n")
		builder.WriteString("|                          threads |  N | TPS mean±sd | TPS min..max | QPS mean±sd | Lat avg mean±sd | Lat p95 mean±sd | Errors | Reconnects |\n")
		builder.WriteString("| -------------------------------: | -: | ----------: | -----------: | ----------: | --------------: | --------------: | -----: | ---------: |\n")
		builder.WriteString(fmt.Sprintf("| %30s | %d | %s | %s | %s | %s | %s | %d | %d |\n",
			"-",
			group.Statistics.N,
			formatGroupMetric(group.Statistics.TPS),
			formatGroupMetricRange(group.Statistics.TPS),
			formatGroupMetric(group.Statistics.QPS),
			formatGroupMetric(group.Statistics.LatencyAvg),
			formatGroupMetric(group.Statistics.LatencyP95),
			group.Statistics.Errors,
			group.Statistics.Reconnects,
		))
		builder.WriteString("\n")

		// Scaling & Efficiency
		if i > 0 {
			builder.WriteString("### 4.%d.5 Scaling & Efficiency (Threads Analysis)\n\n")
			builder.WriteString("|                  threads | TPS_mean | Speedup | Efficiency | ΔTPS vs prev | Δp95 vs prev |\n")
			builder.WriteString("| -----------------------: | -------: | ------: | ---------: | -----------: | -----------: |\n")

			for j := 0; j <= i; j++ {
				group := r.ConfigGroups[j]
				metrics := calculateScalingMetrics(group, r.ConfigGroups)

				deltaTPS := "—"
				deltaP95 := "—"

				if j > 0 {
					prevGroup := r.ConfigGroups[j-1]
					deltaTPS = fmt.Sprintf("%.2f", group.Statistics.TPS.Mean-prevGroup.Statistics.TPS.Mean)
					deltaP95 = fmt.Sprintf("%.2f", group.Statistics.LatencyP95.Mean-prevGroup.Statistics.LatencyP95.Mean)
				}

				builder.WriteString(fmt.Sprintf("| %25s | %8.2f | %7.2fx | %9s | %11s | %12s |\n",
					fmt.Sprintf("threads=%d", group.Threads),
					group.Statistics.TPS.Mean,
					metrics.Speedup,
					formatPercentage(metrics.Efficiency),
					deltaTPS,
					deltaP95,
				))
			}
			builder.WriteString("\n")
		}

		// Visuals
		builder.WriteString(fmt.Sprintf("### 4.%d.7 Visuals (ASCII)\n\n", i+1))

		// TPS Bar Chart
		builder.WriteString("#### TPS vs Threads (run summary mean)\n")
		builder.WriteString("```text\n")
		maxTPS := 0.0
		for _, g := range r.ConfigGroups {
			if g.Statistics.TPS.Max > maxTPS {
				maxTPS = g.Statistics.TPS.Max
			}
		}
		for _, g := range r.ConfigGroups {
			tps := g.Statistics.TPS.Mean
			barWidth := 50
			barLength := int((tps / maxTPS) * float64(barWidth))
			if barLength < 1 {
				barLength = 1
			}
			if barLength > barWidth {
				barLength = barWidth
			}
			bar := strings.Repeat("█", barLength)
			spaces := strings.Repeat(" ", barWidth-barLength)
			builder.WriteString(fmt.Sprintf("threads=%-2d |%s%s %.2f\n",
				g.Threads, bar, spaces, tps))
		}
		builder.WriteString("```\n\n")

		// p95 Bar Chart
		builder.WriteString("#### p95 Latency vs Threads (run summary mean)\n")
		builder.WriteString("```text\n")
		maxP95 := 0.0
		for _, g := range r.ConfigGroups {
			if g.Statistics.LatencyP95.Max > maxP95 {
				maxP95 = g.Statistics.LatencyP95.Max
			}
		}
		for _, g := range r.ConfigGroups {
			p95 := g.Statistics.LatencyP95.Mean
			barWidth := 50
			barLength := int((p95 / maxP95) * float64(barWidth))
			if barLength < 1 {
				barLength = 1
			}
			if barLength > barWidth {
				barLength = barWidth
			}
			bar := strings.Repeat("█", barLength)
			spaces := strings.Repeat(" ", barWidth-barLength)
			builder.WriteString(fmt.Sprintf("threads=%-2d |%s%s %.2fms\n",
				g.Threads, bar, spaces, p95))
		}
		builder.WriteString("```\n\n")

		// Findings
		if r.Findings != nil {
			builder.WriteString(fmt.Sprintf("### 4.%d.8 Findings & Recommendation\n\n", i+1))
			builder.WriteString(fmt.Sprintf("* Best throughput threads: %d (TPS=%.2f)\n",
				r.Findings.BestTPSThreads, r.Findings.BestTPSValue))
			if r.Findings.BestLatencyThreads > 0 {
				builder.WriteString(fmt.Sprintf("* Best latency threads: %d (p95=%.2fms)\n",
					r.Findings.BestLatencyThreads, r.Findings.BestLatencyValue))
			}
			if r.Findings.ScalingKnee > 0 {
				builder.WriteString(fmt.Sprintf("* Knee point: threads=%d\n",
					r.Findings.ScalingKnee))
			}
			builder.WriteString(fmt.Sprintf("* Recommendation: %s\n",
				r.Findings.Recommendation))
			builder.WriteString("\n")
		}
	}

	return builder.String()
}

// formatGroupMetric formats mean±stddev for a group metric.
func formatGroupMetric(stats GroupMetricStats) string {
	if stats.StdDev == 0 {
		return fmt.Sprintf("%.2f", stats.Mean)
	}
	return fmt.Sprintf("%.2f ± %.2f", stats.Mean, stats.StdDev)
}

// formatGroupMetricRange formats min..max for a group metric.
func formatGroupMetricRange(stats GroupMetricStats) string {
	if stats.Min == stats.Max {
		return fmt.Sprintf("%.2f", stats.Min)
	}
	return fmt.Sprintf("%.2f .. %.2f", stats.Min, stats.Max)
}

// ScalingMetrics represents scaling analysis metrics.
type SimplifiedScalingMetrics struct {
	Speedup    float64
	Efficiency float64
}

// calculateScalingMetrics calculates scaling metrics for a thread group.
func calculateScalingMetrics(group *ThreadGroup, allGroups []*ThreadGroup) SimplifiedScalingMetrics {
	metrics := SimplifiedScalingMetrics{}

	if len(allGroups) == 0 || allGroups[0].Threads == group.Threads {
		return metrics
	}

	baselineTPS := allGroups[0].Statistics.TPS.Mean
	if baselineTPS == 0 {
		return metrics
	}

	metrics.Speedup = group.Statistics.TPS.Mean / baselineTPS
	metrics.Efficiency = metrics.Speedup / float64(group.Threads)

	return metrics
}

// formatPercentage formats a float as a percentage.
func formatPercentage(val float64) string {
	return fmt.Sprintf("%.2f%%", val*100)
}

// FormatTXT formats the simplified report as plain text.
func (r *SimplifiedReport) FormatTXT() string {
	if r == nil {
		return ""
	}

	var builder strings.Builder

	builder.WriteString("╔════════════════════════════════════════════════════════════════╗\n")
	builder.WriteString("║              Sysbench Comparison Report (Simplified)                        ║\n")
	builder.WriteString("╚════════════════════════════════════════════════════════════════╝\n\n")

	builder.WriteString(fmt.Sprintf("Generated: %s\n", r.GeneratedAt.Format("2006-01-02 15:04:05")))
	builder.WriteString(fmt.Sprintf("Report ID: %s\n", r.ReportID))
	builder.WriteString(fmt.Sprintf("Records: %d\n\n", r.SelectedRecords))

	// Config groups
	builder.WriteString("Configuration Groups:\n")
	for _, group := range r.ConfigGroups {
		builder.WriteString(fmt.Sprintf("  threads=%d: %d run(s), TPS=%.2f\n",
			group.Threads, group.Statistics.N, group.Statistics.TPS.Mean))
	}
	builder.WriteString("\n")

	// Sanity checks
	builder.WriteString("Sanity Checks:\n")
	passed := 0
	for _, check := range r.SanityChecks {
		if check.Passed {
			passed++
		} else {
			builder.WriteString(fmt.Sprintf("  ❌ %s\n", check.Name))
		}
	}
	builder.WriteString(fmt.Sprintf("\nTotal: %d/%d passed\n\n", passed, len(r.SanityChecks)))

	// Findings
	if r.Findings != nil {
		builder.WriteString("Findings:\n")
		builder.WriteString(fmt.Sprintf("  Best TPS: threads=%d (TPS=%.2f)\n",
			r.Findings.BestTPSThreads, r.Findings.BestTPSValue))
		if r.Findings.BestLatencyThreads > 0 {
			builder.WriteString(fmt.Sprintf("  Best Latency: threads=%d (p95=%.2fms)\n",
				r.Findings.BestLatencyThreads, r.Findings.BestLatencyValue))
		}
		builder.WriteString(fmt.Sprintf("  Recommendation: %s\n", r.Findings.Recommendation))
	}

	return builder.String()
}
