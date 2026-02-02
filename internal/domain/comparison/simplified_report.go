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
	LatencyMax  GroupMetricStats
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
	latMaxValues := make([]float64, n)

	for i, record := range records {
		tpsValues[i] = record.TPS
		qpsValues[i] = record.QPS
		latAvgValues[i] = record.LatencyAvg
		latP95Values[i] = record.LatencyP95
		latMaxValues[i] = record.LatencyMax
		stats.Errors += record.IgnoredErrors
		stats.Reconnects += record.Reconnects
	}

	// Calculate TPS statistics
	stats.TPS = calculateGroupMetricStats(tpsValues)
	stats.QPS = calculateGroupMetricStats(qpsValues)
	stats.LatencyAvg = calculateGroupMetricStats(latAvgValues)
	stats.LatencyP95 = calculateGroupMetricStats(latP95Values)
	stats.LatencyMax = calculateGroupMetricStats(latMaxValues)

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
	builder.WriteString("# Sysbench Multi-Configuration Comparison Report\n\n")
	builder.WriteString(fmt.Sprintf("* **Generated at:** %s\n", r.GeneratedAt.Format("2006-01-02 15:04:05")))
	builder.WriteString(fmt.Sprintf("* **Report ID:** %s\n", r.ReportID))
	builder.WriteString(fmt.Sprintf("* **Group by:** %s\n", r.GroupBy))
	builder.WriteString(fmt.Sprintf("* **Config Groups:** %d\n", len(r.ConfigGroups)))
	builder.WriteString("\n---\n\n")

	// Section 1: Experiment Metadata
	builder.WriteString("## 1) Experiment Metadata\n\n")
	builder.WriteString("### 1.1 Basic Information\n\n")
	builder.WriteString("| Item | Value |\n")
	builder.WriteString("|------|-------|\n")
	builder.WriteString(fmt.Sprintf("| Report ID | %s |\n", r.ReportID))
	builder.WriteString(fmt.Sprintf("| Generated | %s |\n", r.GeneratedAt.Format("2006-01-02 15:04:05")))
	builder.WriteString(fmt.Sprintf("| Group By | %s |\n", r.GroupBy))
	builder.WriteString(fmt.Sprintf("| Config Groups | %d |\n", len(r.ConfigGroups)))
	builder.WriteString("\n")

	builder.WriteString("### 1.3 Measurement Policy\n\n")
	builder.WriteString("* **Report interval:** 1s\n")
	builder.WriteString("* **Test duration:** Varies by run\n")
	builder.WriteString("* **Runs per config (N):** Varies by config\n")
	builder.WriteString("* **Execution order:** Based on start time\n")
	builder.WriteString("* **Acceptance criteria:** errors=0 && reconnects=0\n")
	builder.WriteString("\n")

	// Section 2: Experiment Matrix
	builder.WriteString("## 2) Experiment Matrix\n\n")
	builder.WriteString("| Config ID | threads | Database | Template | Runs (N) | Tags |\n")
	builder.WriteString("|---------:|-------:|---------|----------|--------:|------|\n")
	for i, group := range r.ConfigGroups {
		cid := fmt.Sprintf("C%d", i+1)
		database := group.Records[0].DatabaseType
		template := group.Records[0].TemplateName
		n := group.Statistics.N

		var tags []string
		if r.Findings != nil && group.Threads == r.Findings.BestTPSThreads {
			tags = append(tags, "best-tps")
		}
		if r.Findings != nil && group.Threads == r.Findings.BestLatencyThreads {
			tags = append(tags, "best-latency")
		}
		if group.Threads == 1 {
			tags = append(tags, "baseline")
		}
		tagStr := strings.Join(tags, " ")

		builder.WriteString(fmt.Sprintf("| %s | %d | %s | %s | %d | %s |\n",
			cid, group.Threads, database, template, n, tagStr))
	}
	builder.WriteString("\n")

	// Section 3: Main Comparison (Run Summary Metrics)
	builder.WriteString("## 3) Main Comparison (Run Summary Metrics)\n\n")
	builder.WriteString("> **Note:** If N=1, StdDev = N/A; Min=Avg=Max=Single value\n")
	builder.WriteString("> Latency unit: milliseconds\n\n")

	builder.WriteString("### 3.1 Throughput & Latency Summary\n\n")
	builder.WriteString("| threads | N | TPS (mean ± sd) | TPS (min..max) | QPS (mean ± sd) | QPS (min..max) | Lat avg ms (mean ± sd) | Lat p95 ms (mean ± sd) | Lat max ms (max-of-max) |\n")
	builder.WriteString("|-------:|:-:|---------------:|--------------:|---------------:|--------------:|----------------------:|----------------------:|-----------------------:|\n")

	for _, group := range r.ConfigGroups {
		// Calculate max latency (max-of-max across all runs in this group)
		maxLat := group.Statistics.LatencyMax.Max

		builder.WriteString(fmt.Sprintf("| %d | %d | %s | %s | %s | %s | %s | %s | %.2f |\n",
			group.Threads,
			group.Statistics.N,
			formatGroupMetric(group.Statistics.TPS),
			formatGroupMetricRange(group.Statistics.TPS),
			formatGroupMetric(group.Statistics.QPS),
			formatGroupMetricRange(group.Statistics.QPS),
			formatGroupMetric(group.Statistics.LatencyAvg),
			formatGroupMetric(group.Statistics.LatencyP95),
			maxLat,
		))
	}
	builder.WriteString("\n")

	builder.WriteString("### 3.2 Reliability\n\n")
	builder.WriteString("| threads | N | Total Errors | Total Reconnects | Any non-zero? |\n")
	builder.WriteString("|-------:|:-:|------------:|---------------:|:-------------|\n")
	for _, group := range r.ConfigGroups {
		anyNonZero := "NO"
		if group.Statistics.Errors > 0 || group.Statistics.Reconnects > 0 {
			anyNonZero = "YES"
		}
		builder.WriteString(fmt.Sprintf("| %d | %d | %d | %d | %s |\n",
			group.Threads, group.Statistics.N,
			group.Statistics.Errors, group.Statistics.Reconnects, anyNonZero))
	}
	builder.WriteString("\n")

	// Calculate query mix from first group (assuming same mix across all)
	if len(r.ConfigGroups) > 0 && len(r.ConfigGroups[0].Records) > 0 {
		record := r.ConfigGroups[0].Records[0]
		totalQ := record.ReadQueries + record.WriteQueries + record.OtherQueries
		if totalQ > 0 {
			builder.WriteString("### 3.3 Actual Query Mix (from SQL statistics)\n\n")
			builder.WriteString("| threads | Read % | Write % | Other % | Queries / Transaction |\n")
			builder.WriteString("|-------:|------:|-------:|-------:|--------------------:|\n")
			for _, group := range r.ConfigGroups {
				if len(group.Records) > 0 {
					r := group.Records[0]
					tot := r.ReadQueries + r.WriteQueries + r.OtherQueries
					if tot > 0 {
						rp := float64(r.ReadQueries) / float64(tot) * 100
						wp := float64(r.WriteQueries) / float64(tot) * 100
						op := float64(r.OtherQueries) / float64(tot) * 100
						qpt := 0.0
						if r.TPS > 0 {
							qpt = float64(r.TotalQueries) / r.TPS
						}
						builder.WriteString(fmt.Sprintf("| %d | %.1f | %.1f | %.1f | %.2f |\n",
							group.Threads, rp, wp, op, qpt))
					}
				}
			}
			builder.WriteString("\n")
		}
	}

	// Section 5: Scaling & Efficiency
	if len(r.ConfigGroups) > 0 && r.ConfigGroups[0].Threads == 1 {
		builder.WriteString("## 5) Scaling & Efficiency (Threads Analysis)\n\n")
		baselineTPS := r.ConfigGroups[0].Statistics.TPS.Mean
		builder.WriteString(fmt.Sprintf("**Baseline:** threads=1 (TPS=%.2f)\n\n", baselineTPS))

		builder.WriteString("| threads | TPS_mean | Speedup | Efficiency (Speedup / threads) | ΔTPS vs prev | Δp95 latency |\n")
		builder.WriteString("|-------:|--------:|-------:|-------------------------------:|------------:|-------------:|\n")

		for i, group := range r.ConfigGroups {
			speedup := group.Statistics.TPS.Mean / baselineTPS
			efficiency := speedup / float64(group.Threads)

			deltaTPS := "—"
			deltaP95 := "—"

			if i > 0 {
				prevGroup := r.ConfigGroups[i-1]
				deltaTPS = fmt.Sprintf("%.2f", group.Statistics.TPS.Mean-prevGroup.Statistics.TPS.Mean)
				deltaP95 = fmt.Sprintf("%.2f", group.Statistics.LatencyP95.Mean-prevGroup.Statistics.LatencyP95.Mean)
			}

			builder.WriteString(fmt.Sprintf("| %d | %.2f | %.2fx | %.2f%% | %s | %s |\n",
				group.Threads,
				group.Statistics.TPS.Mean,
				speedup,
				efficiency*100,
				deltaTPS,
				deltaP95,
			))
		}
		builder.WriteString("\n")
	}

	// Section 6: Visuals
	builder.WriteString("## 6) Visuals (ASCII Charts)\n\n")

	builder.WriteString("### 6.1 TPS vs Threads\n")
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
		builder.WriteString(fmt.Sprintf("threads=%d  |%s%s %.2f\n",
			g.Threads, bar, spaces, tps))
	}
	builder.WriteString("```\n\n")

	builder.WriteString("### 6.2 p95 Latency vs Threads\n")
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
		builder.WriteString(fmt.Sprintf("threads=%d  |%s%s %.2fms\n",
			g.Threads, bar, spaces, p95))
	}
	builder.WriteString("```\n\n")

	// Section 7: Sanity Checks
	builder.WriteString("## 7) Sanity Checks\n\n")

	allPassed := true
	for _, check := range r.SanityChecks {
		if !check.Passed {
			allPassed = false
			break
		}
	}

	if allPassed {
		builder.WriteString("✅ **ALL CHECKS PASSED**\n\n")
	} else {
		builder.WriteString("⚠️  **SOME CHECKS FAILED**\n\n")
	}

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
		builder.WriteString(fmt.Sprintf("| %s | %s | %s |\n", check.Name, result, details))
	}
	builder.WriteString("\n")

	// Section 8: Findings & Recommendations
	builder.WriteString("## 8) Findings & Recommendations\n\n")

	builder.WriteString("### 8.1 Key Findings\n\n")
	if r.Findings != nil {
		builder.WriteString(fmt.Sprintf("* **Best throughput point:** threads=%d (TPS=%.2f, p95=%.2fms)\n",
			r.Findings.BestTPSThreads, r.Findings.BestTPSValue,
			getLatencyForThreads(r.ConfigGroups, r.Findings.BestTPSThreads)))

		if r.Findings.BestLatencyThreads > 0 {
			builder.WriteString(fmt.Sprintf("* **Best latency point:** threads=%d (p95=%.2fms)\n",
				r.Findings.BestLatencyThreads, r.Findings.BestLatencyValue))
		}

		if r.Findings.ScalingKnee > 0 {
			builder.WriteString(fmt.Sprintf("* **Scaling knee:** threads=~%d (efficiency drops significantly)\n",
				r.Findings.ScalingKnee))
		}

		// Check stability
		stable := true
		for _, group := range r.ConfigGroups {
			if group.Statistics.N > 1 {
				cv := (group.Statistics.TPS.StdDev / group.Statistics.TPS.Mean) * 100
				if cv > 10 {
					stable = false
					break
				}
			}
		}
		if stable {
			builder.WriteString("* **Stability:** All configs stable (CV < 10%)\n")
		} else {
			builder.WriteString("* **Stability:** Some configs show high variance (CV > 10%)\n")
		}
	}

	builder.WriteString("\n### 8.2 Recommendation\n\n")
	if r.Findings != nil {
		builder.WriteString(fmt.Sprintf("**Suggested:** threads=%d\n\n", r.Findings.BestTPSThreads))

		// Trade-off statement
		bestGroup := getGroupByThreads(r.ConfigGroups, r.Findings.BestTPSThreads)
		if bestGroup != nil && len(r.ConfigGroups) > 0 && r.ConfigGroups[0].Threads == 1 {
			speedup := bestGroup.Statistics.TPS.Mean / r.ConfigGroups[0].Statistics.TPS.Mean
			efficiency := speedup / float64(bestGroup.Threads)
			builder.WriteString(fmt.Sprintf("**Trade-off:** %.2fx speedup with %.2f%% scaling efficiency at %.2fms p95 latency\n\n",
				speedup, efficiency*100, bestGroup.Statistics.LatencyP95.Mean))
		}

		builder.WriteString("**Next experiment:** Repeat with N=5 runs per config for better statistics\n")
	}

	return builder.String()
}

// getLatencyForThreads returns p95 latency for the given thread count.
func getLatencyForThreads(groups []*ThreadGroup, threads int) float64 {
	for _, g := range groups {
		if g.Threads == threads {
			return g.Statistics.LatencyP95.Mean
		}
	}
	return 0
}

// getGroupByThreads returns the thread group with the given thread count.
func getGroupByThreads(groups []*ThreadGroup, threads int) *ThreadGroup {
	for _, g := range groups {
		if g.Threads == threads {
			return g
		}
	}
	return nil
}

// formatGroupMetric formats mean±stddev for a group metric.
// If N=1 (indicated by StdDev=0 and Min=Max), returns "N/A" for stddev.
func formatGroupMetric(stats GroupMetricStats) string {
	if stats.StdDev == 0 && stats.Min == stats.Max {
		// Single value (N=1)
		return fmt.Sprintf("%.2f", stats.Mean)
	}
	return fmt.Sprintf("%.2f ± %.2f", stats.Mean, stats.StdDev)
}

// formatGroupMetricRange formats min..max for a group metric.
// If N=1, returns the single value.
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
