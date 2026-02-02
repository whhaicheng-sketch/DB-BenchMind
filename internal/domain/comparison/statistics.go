// Package comparison provides statistical analysis functions.
// This file implements comprehensive statistics calculations for grouped runs.
package comparison

import (
	"fmt"
	"math"
)

// CalculateRunStats calculates comprehensive statistics across N runs.
// This includes throughput, latency, reliability, and query mix statistics.
func CalculateRunStats(runs []*Run) RunStats {
	if len(runs) == 0 {
		return RunStats{}
	}

	n := len(runs)

	// Initialize stats
	stats := RunStats{
		N: n,
		TPS: RunMetricStats{
			N:      n,
			Values: make([]float64, n),
		},
		QPS: RunMetricStats{
			N:      n,
			Values: make([]float64, n),
		},
		LatencyAvg: RunMetricStats{
			N:      n,
			Values: make([]float64, n),
		},
		LatencyP95: RunMetricStats{
			N:      n,
			Values: make([]float64, n),
		},
		LatencyP99: RunMetricStats{
			N:      n,
			Values: make([]float64, n),
		},
	}

	// Collect values from all runs
	for i, run := range runs {
		stats.TPS.Values[i] = run.TPS
		stats.QPS.Values[i] = run.QPS
		stats.LatencyAvg.Values[i] = run.LatencyAvg
		stats.LatencyP95.Values[i] = run.LatencyP95
		stats.LatencyP99.Values[i] = run.LatencyP99

		stats.TotalErrors += run.Errors
		stats.TotalReconnects += run.Reconnects
	}

	// Calculate TPS statistics
	stats.TPS = calculateRunMetricStats(stats.TPS.Values)

	// Calculate QPS statistics
	stats.QPS = calculateRunMetricStats(stats.QPS.Values)

	// Calculate Latency statistics
	stats.LatencyAvg = calculateRunMetricStats(stats.LatencyAvg.Values)
	stats.LatencyP95 = calculateRunMetricStats(stats.LatencyP95.Values)
	stats.LatencyP99 = calculateRunMetricStats(stats.LatencyP99.Values)

	// Find max-of-max for latency
	var maxLatency float64
	for _, run := range runs {
		if run.LatencyMax > maxLatency {
			maxLatency = run.LatencyMax
		}
	}
	stats.LatencyMax = maxLatency

	// Check for errors
	stats.HasErrors = stats.TotalErrors > 0 || stats.TotalReconnects > 0
	stats.AnyNonZero = stats.HasErrors

	// Calculate query mix percentages (averaged across runs)
	var totalRead, totalWrite, totalOther int64
	var totalQueriesPerTx float64
	for _, run := range runs {
		totalRead += run.ReadQueries
		totalWrite += run.WriteQueries
		totalOther += run.OtherQueries
		totalQueriesPerTx += run.QueriesPerTransaction
	}

	totalAllQueries := totalRead + totalWrite + totalOther
	if totalAllQueries > 0 {
		stats.ReadPct = float64(totalRead) / float64(totalAllQueries) * 100
		stats.WritePct = float64(totalWrite) / float64(totalAllQueries) * 100
		stats.OtherPct = float64(totalOther) / float64(totalAllQueries) * 100
	}

	if n > 0 {
		stats.QueriesPerTx = totalQueriesPerTx / float64(n)
	}

	return stats
}

// calculateRunMetricStats calculates statistics for a single metric across N runs.
func calculateRunMetricStats(values []float64) RunMetricStats {
	n := len(values)
	if n == 0 {
		return RunMetricStats{}
	}

	stats := RunMetricStats{
		N:      n,
		Values: values,
	}

	// Calculate min and max
	stats.Min = values[0]
	stats.Max = values[0]
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

	// Calculate standard deviation (sample stddev, n-1)
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

// CalculateCV calculates the Coefficient of Variation (CV%).
// CV = (StdDev / Mean) × 100
// Higher CV indicates more variability/jitter.
func CalculateCV(mean, stddev float64) float64 {
	if mean == 0 {
		return 0
	}
	return (stddev / mean) * 100
}

// CalculateSpeedup calculates speedup vs baseline.
// Speedup = value / baseline
func CalculateSpeedup(value, baseline float64) float64 {
	if baseline == 0 {
		return 0
	}
	return value / baseline
}

// CalculateEfficiency calculates scaling efficiency.
// Efficiency = Speedup / threads
func CalculateEfficiency(speedup float64, threads int) float64 {
	if threads == 0 {
		return 0
	}
	return speedup / float64(threads)
}

// FormatMetricValue formats a metric value with appropriate precision.
func FormatMetricValue(value float64, isInteger bool) string {
	if isInteger {
		return fmt.Sprintf("%.0f", value)
	}
	return fmt.Sprintf("%.2f", value)
}

// FormatMeanStdDev formats mean and stddev as "mean ± stddev".
func FormatMeanStdDev(stats RunMetricStats) string {
	if !stats.IsValid() {
		return "N/A"
	}
	if stats.N == 1 {
		return FormatMetricValue(stats.Mean, false)
	}
	return fmt.Sprintf("%s ± %s",
		FormatMetricValue(stats.Mean, false),
		FormatMetricValue(stats.StdDev, false))
}

// FormatMinMax formats min and max as "(min..max)".
func FormatMinMax(stats RunMetricStats) string {
	if !stats.IsValid() {
		return "N/A"
	}
	if stats.N == 1 {
		return FormatMetricValue(stats.Min, false)
	}
	return fmt.Sprintf("%s .. %s",
		FormatMetricValue(stats.Min, false),
		FormatMetricValue(stats.Max, false))
}

// ValidateMetricStats checks if statistics are mathematically valid.
// Returns true if min ≤ mean ≤ max.
func ValidateMetricStats(stats RunMetricStats) bool {
	if !stats.IsValid() {
		return false
	}
	// Allow small floating point errors
	epsilon := 0.0001
	return stats.Min <= stats.Mean+epsilon &&
		stats.Mean <= stats.Max+epsilon
}

// CalculateDelta calculates the change between two values.
// Returns absolute delta and percentage change.
func CalculateDelta(current, previous float64) (delta float64, pctChange float64) {
	delta = current - previous
	if previous == 0 {
		pctChange = 0
	} else {
		pctChange = (delta / previous) * 100
	}
	return
}

// AggregateStats aggregates multiple RunStats into summary statistics.
// Useful for calculating overall statistics across all config groups.
func AggregateStats(groups []*ConfigGroup) RunStats {
	if len(groups) == 0 {
		return RunStats{}
	}

	// Collect all TPS values across all groups
	var allTPS []float64
	var allQPS []float64
	var allLatencyAvg []float64
	var allLatencyP95 []float64

	for _, group := range groups {
		allTPS = append(allTPS, group.Statistics.TPS.Values...)
		allQPS = append(allQPS, group.Statistics.QPS.Values...)
		allLatencyAvg = append(allLatencyAvg, group.Statistics.LatencyAvg.Values...)
		allLatencyP95 = append(allLatencyP95, group.Statistics.LatencyP95.Values...)
	}

	// Calculate overall statistics
	return RunStats{
		N:          len(allTPS),
		TPS:        calculateRunMetricStats(allTPS),
		QPS:        calculateRunMetricStats(allQPS),
		LatencyAvg: calculateRunMetricStats(allLatencyAvg),
		LatencyP95: calculateRunMetricStats(allLatencyP95),
	}
}

// FindBestTPSConfig finds the config group with the highest mean TPS.
func FindBestTPSConfig(groups []*ConfigGroup) *ConfigGroup {
	if len(groups) == 0 {
		return nil
	}

	best := groups[0]
	for _, group := range groups {
		if group.Statistics.TPS.Mean > best.Statistics.TPS.Mean {
			best = group
		}
	}
	return best
}

// FindWorstLatencyConfig finds the config group with the highest mean p95 latency.
func FindWorstLatencyConfig(groups []*ConfigGroup) *ConfigGroup {
	if len(groups) == 0 {
		return nil
	}

	worst := groups[0]
	for _, group := range groups {
		if group.Statistics.LatencyP95.Mean > worst.Statistics.LatencyP95.Mean {
			worst = group
		}
	}
	return worst
}

// CalculateOverallReliability checks reliability across all groups.
func CalculateOverallReliability(groups []*ConfigGroup) (totalErrors, totalReconnects int64, anyErrors bool) {
	for _, group := range groups {
		totalErrors += group.Statistics.TotalErrors
		totalReconnects += group.Statistics.TotalReconnects
		if group.Statistics.HasErrors {
			anyErrors = true
		}
	}
	return
}

// CalculateConfidenceInterval calculates 95% confidence interval for the mean.
// Returns (lower_bound, upper_bound).
func CalculateConfidenceInterval(stats RunMetricStats) (lower, upper float64) {
	if !stats.IsValid() || stats.N < 2 {
		return stats.Mean, stats.Mean
	}

	// 95% CI = mean ± 1.96 * (stddev / sqrt(n))
	margin := 1.96 * (stats.StdDev / math.Sqrt(float64(stats.N)))
	lower = stats.Mean - margin
	upper = stats.Mean + margin

	return
}

// GetPercentile calculates the percentile of values.
func GetPercentile(values []float64, percentile float64) float64 {
	if len(values) == 0 {
		return 0
	}

	// Make a copy and sort
	sorted := make([]float64, len(values))
	copy(sorted, values)

	// Simple sort (for small datasets)
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	// Calculate percentile index
	index := (percentile / 100.0) * float64(len(sorted)-1)
	lower := int(index)
	upper := lower + 1

	if upper >= len(sorted) {
		return sorted[lower]
	}

	// Linear interpolation
	weight := index - float64(lower)
	return sorted[lower]*(1-weight) + sorted[upper]*weight
}
