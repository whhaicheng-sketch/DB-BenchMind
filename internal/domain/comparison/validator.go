// Package comparison provides data validation functions.
// This file implements sanity checks for comparison report data integrity.
package comparison

import (
	"fmt"
	"math"
	"strings"
)

// ValidateReport performs comprehensive sanity checks on a comparison report.
func ValidateReport(report *ComparisonReport) *SanityCheckResults {
	if report == nil {
		return &SanityCheckResults{
			AllPassed: false,
			Checks: []SanityCheck{
				{Name: "Report exists", Passed: false, Details: "Report is nil"},
			},
		}
	}

	results := &SanityCheckResults{
		Checks: []SanityCheck{},
	}

	// Check 1: Config groups exist
	results.Checks = append(results.Checks,
		SanityCheck{
			Name:   "Config groups exist",
			Passed: len(report.ConfigGroups) > 0,
			Details: fmt.Sprintf("Found %d config groups", len(report.ConfigGroups)),
		},
	)

	// Check each config group
	for _, group := range report.ConfigGroups {
		// Check 2: Latency ordering (min ≤ avg ≤ p95 ≤ p99 ≤ max)
		results.Checks = append(results.Checks,
			validateLatencyOrdering(group)...,
		)

		// Check 3: QPS ≈ TPS × queries_per_transaction
		results.Checks = append(results.Checks,
			validateQPSConsistency(group),
		)

		// Check 4: SQL total = read + write + other
		results.Checks = append(results.Checks,
			validateSQLTotal(group),
		)

		// Check 5: No errors or reconnects
		results.Checks = append(results.Checks,
			validateReliability(group),
		)

		// Check 6: N=1 case (stddev should be 0 or N/A)
		if group.Statistics.N == 1 {
			results.Checks = append(results.Checks,
				SanityCheck{
					Name:   fmt.Sprintf("N=1 stddev check (Group %s)", group.GroupID),
					Passed: group.Statistics.TPS.StdDev == 0,
					Details: fmt.Sprintf("StdDev=%.2f (expected 0 for N=1)", group.Statistics.TPS.StdDev),
				},
			)
		}
	}

	// Check 7: Baseline exists (threads=1)
	hasBaseline := false
	for _, group := range report.ConfigGroups {
		if group.Config.Threads == 1 {
			hasBaseline = true
			break
		}
	}
	results.Checks = append(results.Checks,
		SanityCheck{
			Name:   "Baseline exists",
			Passed: hasBaseline,
			Details: fmt.Sprintf("threads=1 group found: %v", hasBaseline),
		},
	)

	// Calculate overall pass status
	allPassed := true
	for _, check := range results.Checks {
		if !check.Passed {
			allPassed = false
		}
	}
	results.AllPassed = allPassed

	return results
}

// validateLatencyOrdering checks if latency metrics are in correct order.
// Expected: min ≤ avg ≤ p95 ≤ p99 ≤ max
func validateLatencyOrdering(group *ConfigGroup) []SanityCheck {
	checks := []SanityCheck{}

	const epsilon = 0.01 // Allow small floating point errors

	// Check min ≤ avg
	checks = append(checks, SanityCheck{
		Name:   fmt.Sprintf("Latency min ≤ avg (Group %s)", group.GroupID),
		Passed: group.Statistics.LatencyAvg.Min <= group.Statistics.LatencyAvg.Mean+epsilon,
		Details: fmt.Sprintf("min=%.2f, avg=%.2f",
			group.Statistics.LatencyAvg.Min, group.Statistics.LatencyAvg.Mean),
	})

	// Check avg ≤ p95
	checks = append(checks, SanityCheck{
		Name:   fmt.Sprintf("Latency avg ≤ p95 (Group %s)", group.GroupID),
		Passed: group.Statistics.LatencyAvg.Mean <= group.Statistics.LatencyP95.Mean+epsilon,
		Details: fmt.Sprintf("avg=%.2f, p95=%.2f",
			group.Statistics.LatencyAvg.Mean, group.Statistics.LatencyP95.Mean),
	})

	// Check p95 ≤ p99
	checks = append(checks, SanityCheck{
		Name:   fmt.Sprintf("Latency p95 ≤ p99 (Group %s)", group.GroupID),
		Passed: group.Statistics.LatencyP95.Mean <= group.Statistics.LatencyP99.Mean+epsilon,
		Details: fmt.Sprintf("p95=%.2f, p99=%.2f",
			group.Statistics.LatencyP95.Mean, group.Statistics.LatencyP99.Mean),
	})

	// Check p99 ≤ max (using LatencyMax which is max-of-max)
	checks = append(checks, SanityCheck{
		Name:   fmt.Sprintf("Latency p99 ≤ max (Group %s)", group.GroupID),
		Passed: group.Statistics.LatencyP99.Mean <= group.Statistics.LatencyMax+epsilon,
		Details: fmt.Sprintf("p99=%.2f, max=%.2f",
			group.Statistics.LatencyP99.Mean, group.Statistics.LatencyMax),
	})

	return checks
}

// validateQPSConsistency checks if QPS ≈ TPS × queries_per_transaction.
// Allow 5% tolerance for rounding errors.
func validateQPSConsistency(group *ConfigGroup) SanityCheck {
	if len(group.Runs) == 0 {
		return SanityCheck{
			Name:   "QPS consistency (no runs)",
			Passed: false,
			Details: "No runs to validate",
		}
	}

	// Use mean values from statistics
	expectedQPS := group.Statistics.TPS.Mean * group.Statistics.QueriesPerTx
	actualQPS := group.Statistics.QPS.Mean

	// Calculate percentage difference
	diff := math.Abs(actualQPS-expectedQPS) / expectedQPS * 100

	// Allow 5% tolerance
	passed := diff <= 5.0

	return SanityCheck{
		Name:   fmt.Sprintf("QPS ≈ TPS × queries/tx (Group %s)", group.GroupID),
		Passed: passed,
		Details: fmt.Sprintf("Expected=%.2f, Actual=%.2f, Diff=%.2f%%",
			expectedQPS, actualQPS, diff),
	}
}

// validateSQLTotal checks if SQL total = read + write + other.
func validateSQLTotal(group *ConfigGroup) SanityCheck {
	if len(group.Runs) == 0 {
		return SanityCheck{
			Name:   "SQL total validation (no runs)",
			Passed: false,
			Details: "No runs to validate",
		}
	}

	// Use first run for validation (all runs in group should have same structure)
	run := group.Runs[0]

	calculatedTotal := run.ReadQueries + run.WriteQueries + run.OtherQueries
	actualTotal := run.TotalQueries

	passed := calculatedTotal == actualTotal

	details := fmt.Sprintf("Read=%d, Write=%d, Other=%d, Calculated=%d, Actual=%d",
		run.ReadQueries, run.WriteQueries, run.OtherQueries,
		calculatedTotal, actualTotal)

	if !passed {
		details += fmt.Sprintf(" [MISMATCH: diff=%d]", int64(math.Abs(float64(calculatedTotal-actualTotal))))
	}

	return SanityCheck{
		Name:   fmt.Sprintf("SQL total = read + write + other (Group %s)", group.GroupID),
		Passed: passed,
		Details: details,
	}
}

// validateReliability checks if there are no errors or reconnects.
func validateReliability(group *ConfigGroup) SanityCheck {
	hasErrors := group.Statistics.TotalErrors > 0 || group.Statistics.TotalReconnects > 0

	return SanityCheck{
		Name:   fmt.Sprintf("No errors/reconnects (Group %s)", group.GroupID),
		Passed: !hasErrors,
		Details: fmt.Sprintf("Errors=%d, Reconnects=%d",
			group.Statistics.TotalErrors, group.Statistics.TotalReconnects),
	}
}

// ValidateMetricRange checks if a metric value is within acceptable range.
func ValidateMetricRange(value, min, max float64, metricName string) SanityCheck {
	inRange := value >= min && value <= max

	return SanityCheck{
		Name:   fmt.Sprintf("%s in range [%.2f, %.2f]", metricName, min, max),
		Passed: inRange,
		Details: fmt.Sprintf("Value=%.2f", value),
	}
}

// ValidatePositive checks if a metric value is positive.
func ValidatePositive(value float64, metricName string) SanityCheck {
	return SanityCheck{
		Name:   fmt.Sprintf("%s > 0", metricName),
		Passed: value > 0,
		Details: fmt.Sprintf("Value=%.2f", value),
	}
}

// ValidateMonotonicIncrease checks if values increase monotonically.
func ValidateMonotonicIncrease(values []float64, metricName string) SanityCheck {
	if len(values) <= 1 {
		return SanityCheck{
			Name:   fmt.Sprintf("%s monotonic increase", metricName),
			Passed: true,
			Details: "Insufficient data points",
		}
	}

	increasing := true
	for i := 1; i < len(values); i++ {
		if values[i] < values[i-1] {
			increasing = false
			break
		}
	}

	return SanityCheck{
		Name:   fmt.Sprintf("%s monotonic increase", metricName),
		Passed: increasing,
		Details: fmt.Sprintf("Checked %d values", len(values)),
	}
}

// ValidateStdDevReasonable checks if standard deviation is reasonable.
// For stable benchmarks, CV (stddev/mean) should typically be < 10-15%.
func ValidateStdDevReasonable(mean, stddev float64, metricName string) SanityCheck {
	if mean == 0 {
		return SanityCheck{
			Name:   fmt.Sprintf("%s reasonable stddev", metricName),
			Passed: false,
			Details: "Mean is zero, cannot calculate CV",
		}
	}

	cv := CalculateCV(mean, stddev)

	// CV < 15% is considered reasonable for most benchmarks
	// CV < 10% is excellent
	// CV > 20% indicates high variability
	passed := cv < 20.0

	rating := "Excellent"
	if cv > 10 {
		rating = "Good"
	}
	if cv > 15 {
		rating = "Acceptable"
	}
	if cv > 20 {
		rating = "High Variability"
	}

	return SanityCheck{
		Name:   fmt.Sprintf("%s reasonable stddev", metricName),
		Passed: passed,
		Details: fmt.Sprintf("CV=%.2f%% (%s)", cv, rating),
	}
}

// FormatSanityCheckTable formats sanity check results as a table.
func FormatSanityCheckTable(results *SanityCheckResults) string {
	if results == nil {
		return "No sanity check results available"
	}

	var builder strings.Builder

	builder.WriteString("\n## Sanity Checks\n\n")

	if results.AllPassed {
		builder.WriteString("✅ **ALL CHECKS PASSED**\n\n")
	} else {
		builder.WriteString("⚠️  **SOME CHECKS FAILED**\n\n")
	}

	builder.WriteString("│ Check │ Result │ Details │\n")
	builder.WriteString("├───────┼────────┼─────────┤\n")

	for _, check := range results.Checks {
		result := "✅ PASS"
		if !check.Passed {
			result = "❌ FAIL"
		}

		// Truncate details if too long
		details := check.Details
		if len(details) > 50 {
			details = details[:47] + "..."
		}

		builder.WriteString(fmt.Sprintf("│ %-50s │ %-8s │ %-50s │\n",
			check.Name, result, details))
	}

	builder.WriteString("└───────┴────────┴─────────┘\n")

	return builder.String()
}

// GenerateSanityCheckSummary generates a text summary of sanity check results.
func GenerateSanityCheckSummary(results *SanityCheckResults) string {
	if results == nil {
		return "No sanity check results available"
	}

	var builder strings.Builder

	passed := 0
	failed := 0

	for _, check := range results.Checks {
		if check.Passed {
			passed++
		} else {
			failed++
		}
	}

	builder.WriteString(fmt.Sprintf("**Sanity Check Summary:** %d passed, %d failed\n\n",
		passed, failed))

	if failed > 0 {
		builder.WriteString("**Failed Checks:**\n\n")
		for _, check := range results.Checks {
			if !check.Passed {
				builder.WriteString(fmt.Sprintf("• %s: %s\n", check.Name, check.Details))
			}
		}
	}

	return builder.String()
}

// CheckDataQuality checks overall data quality metrics.
func CheckDataQuality(groups []*ConfigGroup) map[string]interface{} {
	quality := make(map[string]interface{})

	// Count total runs
	totalRuns := 0
	for _, group := range groups {
		totalRuns += len(group.Runs)
	}
	quality["total_runs"] = totalRuns

	// Count unique configs
	quality["unique_configs"] = len(groups)

	// Check if all configs have N>1 runs
	allMultiRun := true
	for _, group := range groups {
		if len(group.Runs) <= 1 {
			allMultiRun = false
			break
		}
	}
	quality["all_multi_run"] = allMultiRun

	// Calculate overall reliability
	totalErrors, totalReconnects, anyErrors := CalculateOverallReliability(groups)
	quality["total_errors"] = totalErrors
	quality["total_reconnects"] = totalReconnects
	quality["clean_runs"] = !anyErrors

	// Check thread count range
	if len(groups) > 0 {
		minThreads := groups[0].Config.Threads
		maxThreads := groups[0].Config.Threads
		for _, group := range groups {
			if group.Config.Threads < minThreads {
				minThreads = group.Config.Threads
			}
			if group.Config.Threads > maxThreads {
				maxThreads = group.Config.Threads
			}
		}
		quality["min_threads"] = minThreads
		quality["max_threads"] = maxThreads
		quality["thread_range"] = fmt.Sprintf("%d-%d", minThreads, maxThreads)
	}

	return quality
}
