// Package comparison provides scaling analysis functions.
// This file implements scaling efficiency analysis for multi-threaded benchmarks.
package comparison

import (
	"fmt"
	"strings"
)

// AnalyzeScaling performs comprehensive scaling analysis on config groups.
// It calculates speedup, efficiency, and identifies key scaling characteristics.
func AnalyzeScaling(groups []*ConfigGroup, baselineGroup *ConfigGroup) *ScalingAnalysis {
	if len(groups) == 0 {
		return nil
	}

	analysis := &ScalingAnalysis{
		ByGroup: make(map[string]*ScalingMetrics),
	}

	// Set baseline
	if baselineGroup != nil {
		analysis.BaselineGroup = baselineGroup
		analysis.BaselineTPS = baselineGroup.Statistics.TPS.Mean
	} else if len(groups) > 0 && groups[0].Config.Threads == 1 {
		// Auto-detect baseline (threads=1)
		analysis.BaselineGroup = groups[0]
		analysis.BaselineTPS = groups[0].Statistics.TPS.Mean
	}

	// Calculate metrics for each group
	var previousGroup *ConfigGroup
	for _, group := range groups {
		metrics := &ScalingMetrics{}

		// Calculate speedup vs baseline
		if analysis.BaselineTPS > 0 {
			metrics.Speedup = CalculateSpeedup(group.Statistics.TPS.Mean, analysis.BaselineTPS)
		}

		// Calculate efficiency
		if group.Config.Threads > 0 {
			metrics.Efficiency = CalculateEfficiency(metrics.Speedup, group.Config.Threads)
		}

		// Calculate delta vs previous config
		if previousGroup != nil {
			metrics.DeltaTPS = group.Statistics.TPS.Mean - previousGroup.Statistics.TPS.Mean
			metrics.DeltaP95 = group.Statistics.LatencyP95.Mean - previousGroup.Statistics.LatencyP95.Mean
		}

		analysis.ByGroup[group.GroupID] = metrics
		previousGroup = group
	}

	// Find best throughput config
	analysis.BestTPSConfig = FindBestTPSConfig(groups)

	// Find worst latency config
	analysis.WorstLatencyConfig = FindWorstLatencyConfig(groups)

	// Identify scaling knee (diminishing returns point)
	analysis.ScalingKnee, analysis.ScalingKneeThread = identifyScalingKnee(groups, analysis.ByGroup)

	return analysis
}

// identifyScalingKnee identifies the point where scaling efficiency drops significantly.
// The "knee" is where we get diminishing returns from adding more threads.
//
// Algorithm:
// 1. Calculate efficiency drop between consecutive thread counts
// 2. Find where efficiency drops below 70% (configurable threshold)
// 3. Or find where delta TPS gain is minimal
//
// Returns: the config group at the scaling knee, and the thread count
func identifyScalingKnee(groups []*ConfigGroup, metrics map[string]*ScalingMetrics) (*ConfigGroup, int) {
	if len(groups) < 2 {
		return nil, 0
	}

	// Threshold for considering efficiency as "diminishing returns"
	const efficiencyThreshold = 0.70 // 70%

	// Find first group where efficiency drops below threshold
	for _, group := range groups {
		groupMetrics := metrics[group.GroupID]

		// Skip baseline
		if group.Config.Threads == 1 {
			continue
		}

		// Check if efficiency is below threshold
		if groupMetrics.Efficiency < efficiencyThreshold {
			return group, group.Config.Threads
		}

		// Also check if TPS gain is minimal (< 10% improvement)
		if groupMetrics.DeltaTPS > 0 {
			improvementPct := (groupMetrics.DeltaTPS / group.Statistics.TPS.Mean) * 100
			if improvementPct < 10 {
				return group, group.Config.Threads
			}
		}
	}

	// No clear knee found - return highest thread count
	return groups[len(groups)-1], groups[len(groups)-1].Config.Threads
}

// FormatScalingTable formats a scaling analysis table for reports.
// Returns a formatted string showing speedup and efficiency for each config.
func FormatScalingTable(analysis *ScalingAnalysis, groups []*ConfigGroup) string {
	if analysis == nil || len(groups) == 0 {
		return "No scaling data available"
	}

	var builder strings.Builder

	builder.WriteString("\n## Scaling & Efficiency Analysis\n\n")
	builder.WriteString("Baseline: threads=1 (TPS=")
	builder.WriteString(fmt.Sprintf("%.2f", analysis.BaselineTPS))
	builder.WriteString(")\n\n")

	builder.WriteString("│ threads │ TPS_mean │ Speedup │ Efficiency │ ΔTPS vs prev │ Δp95 latency │\n")
	builder.WriteString("├─────────┼──────────┼─────────┼────────────┼──────────────┼───────────────┤\n")

	for _, group := range groups {
		metrics := analysis.ByGroup[group.GroupID]

		builder.WriteString(fmt.Sprintf("│ %7d │ %8.2f │ %7.2f │ %10.2f │ ",
			group.Config.Threads,
			group.Statistics.TPS.Mean,
			metrics.Speedup,
			metrics.Efficiency))

		if group.Config.Threads == 1 {
			builder.WriteString("         — │")
		} else {
			builder.WriteString(fmt.Sprintf("%12.2f │", metrics.DeltaTPS))
		}

		if group.Config.Threads == 1 {
			builder.WriteString("             — │\n")
		} else {
			builder.WriteString(fmt.Sprintf(" %13.2f │\n", metrics.DeltaP95))
		}
	}

	builder.WriteString("└─────────┴──────────┴─────────┴────────────┴──────────────┴───────────────┘\n")

	return builder.String()
}

// GenerateScalingFindings generates text findings from scaling analysis.
func GenerateScalingFindings(analysis *ScalingAnalysis, groups []*ConfigGroup) string {
	if analysis == nil || len(groups) == 0 {
		return "Insufficient data for scaling analysis"
	}

	var findings []string

	// Best throughput point
	if analysis.BestTPSConfig != nil {
		bestTPS := analysis.BestTPSConfig.Statistics.TPS.Mean
		bestP95 := analysis.BestTPSConfig.Statistics.LatencyP95.Mean
		findings = append(findings,
			fmt.Sprintf("• Best throughput: threads=%d (TPS=%.2f, p95=%.2fms)",
				analysis.BestTPSConfig.Config.Threads, bestTPS, bestP95))
	}

	// Scaling knee
	if analysis.ScalingKnee != nil {
		kneeEfficiency := analysis.ByGroup[analysis.ScalingKnee.GroupID].Efficiency
		findings = append(findings,
			fmt.Sprintf("• Scaling knee: threads=%d (efficiency=%.2f%%)",
				analysis.ScalingKneeThread, kneeEfficiency*100))
	}

	// Latency risk
	if analysis.WorstLatencyConfig != nil {
		worstP95 := analysis.WorstLatencyConfig.Statistics.LatencyP95.Mean
		findings = append(findings,
			fmt.Sprintf("• Latency risk: threads=%d (p95=%.2fms)",
				analysis.WorstLatencyConfig.Config.Threads, worstP95))
	}

	// Stability concerns (high CV groups)
	for _, group := range groups {
		cv := CalculateCV(group.Statistics.TPS.Mean, group.Statistics.TPS.StdDev)
		if cv > 10 { // CV > 10% indicates instability
			findings = append(findings,
				fmt.Sprintf("• Stability concern: threads=%d (CV=%.2f%% - high variability)",
					group.Config.Threads, cv))
		}
	}

	result := "Key Findings:\n\n"
	for _, finding := range findings {
		result += finding + "\n"
	}

	return result
}

// CalculateOptimalThreadCount recommends the optimal thread count based on scaling analysis.
// This balances throughput gain vs latency cost.
func CalculateOptimalThreadCount(analysis *ScalingAnalysis, groups []*ConfigGroup) int {
	if analysis == nil || len(groups) == 0 {
		return 1
	}

	// If we have a clear scaling knee, recommend just before it
	if analysis.ScalingKnee != nil && analysis.ScalingKneeThread > 1 {
		// Recommend the config just before the knee
		for i, group := range groups {
			if group.GroupID == analysis.ScalingKnee.GroupID && i > 0 {
				return groups[i-1].Config.Threads
			}
		}
	}

	// Otherwise, find the point with best efficiency score
	// Score = TPS / (1 + latency_penalty)
	bestScore := 0.0
	bestThreads := 1

	for _, group := range groups {
		if group.Config.Threads == 1 {
			continue
		}

		metrics := analysis.ByGroup[group.GroupID]

		// Score formula: reward speedup, penalize high latency
		latencyPenalty := group.Statistics.LatencyP95.Mean / 10.0 // 10ms per unit penalty
		score := metrics.Speedup / (1.0 + latencyPenalty)

		if score > bestScore {
			bestScore = score
			bestThreads = group.Config.Threads
		}
	}

	return bestThreads
}

// GenerateRecommendation generates a text recommendation based on scaling analysis.
func GenerateRecommendation(analysis *ScalingAnalysis, groups []*ConfigGroup) string {
	if analysis == nil || len(groups) == 0 {
		return "Insufficient data for recommendation"
	}

	optimalThreads := CalculateOptimalThreadCount(analysis, groups)

	var rec strings.Builder

	rec.WriteString(fmt.Sprintf("**Recommended: threads=%d**\n\n", optimalThreads))

	// Find the optimal group
	var optimalGroup *ConfigGroup
	for _, group := range groups {
		if group.Config.Threads == optimalThreads {
			optimalGroup = group
			break
		}
	}

	if optimalGroup != nil {
		metrics := analysis.ByGroup[optimalGroup.GroupID]

		rec.WriteString("**Trade-off Analysis:**\n")
		rec.WriteString(fmt.Sprintf("• Throughput: %.2f TPS (%.2fx speedup vs baseline)\n",
			optimalGroup.Statistics.TPS.Mean, metrics.Speedup))
		rec.WriteString(fmt.Sprintf("• Efficiency: %.2f%% (%.2f%% of ideal linear scaling)\n",
			metrics.Efficiency*100, metrics.Efficiency*100))
		rec.WriteString(fmt.Sprintf("• Latency: p95=%.2fms, p99=%.2fms\n",
			optimalGroup.Statistics.LatencyP95.Mean, optimalGroup.Statistics.LatencyP99.Mean))

		// Check stability
		cv := CalculateCV(optimalGroup.Statistics.TPS.Mean, optimalGroup.Statistics.TPS.StdDev)
		rec.WriteString(fmt.Sprintf("• Stability: CV=%.2f%% (%s)\n",
			cv, getStabilityRating(cv)))

		rec.WriteString("\n**Rationale:**\n")

		// Explain why this config is recommended
		if metrics.Efficiency > 0.8 {
			rec.WriteString("• Excellent scaling efficiency (>80%)\n")
		} else if metrics.Efficiency > 0.6 {
			rec.WriteString("• Good scaling efficiency (>60%)\n")
		}

		if optimalGroup.Statistics.LatencyP95.Mean < 50 {
			rec.WriteString("• Acceptable latency (<50ms p95)\n")
		} else if optimalGroup.Statistics.LatencyP95.Mean < 100 {
			rec.WriteString("• Moderate latency (<100ms p95)\n")
		}

		if cv < 5 {
			rec.WriteString("• Low variability (stable performance)\n")
		} else if cv < 10 {
			rec.WriteString("• Moderate variability\n")
		}
	}

	return rec.String()
}

// getStabilityRating returns a textual stability rating based on CV.
func getStabilityRating(cv float64) string {
	switch {
	case cv < 3:
		return "Very Stable"
	case cv < 5:
		return "Stable"
	case cv < 10:
		return "Moderate"
	case cv < 15:
		return "Variable"
	default:
		return "Highly Variable"
	}
}

// CalculateLinearScalingEfficiency calculates how close the actual scaling is to ideal linear scaling.
// Returns a percentage (0-100) where 100% = perfect linear scaling.
func CalculateLinearScalingEfficiency(actualSpeedup float64, threads int) float64 {
	if threads <= 1 {
		return 100.0
	}

	idealSpeedup := float64(threads)
	efficiency := (actualSpeedup / idealSpeedup) * 100

	if efficiency > 100 {
		efficiency = 100 // Super-linear scaling is rare, cap at 100%
	}

	return efficiency
}

// PredictScaling predicts TPS for a target thread count based on current data.
// Uses simple linear extrapolation (not highly accurate but useful for rough estimates).
func PredictScaling(groups []*ConfigGroup, targetThreads int) float64 {
	if len(groups) == 0 {
		return 0
	}

	// If we have data at or near target threads, use it
	for _, group := range groups {
		if group.Config.Threads == targetThreads {
			return group.Statistics.TPS.Mean
		}
	}

	// Otherwise, extrapolate from the two closest data points
	// Find the closest groups below and above target
	var belowGroup, aboveGroup *ConfigGroup

	for _, group := range groups {
		if group.Config.Threads < targetThreads {
			if belowGroup == nil || group.Config.Threads > belowGroup.Config.Threads {
				belowGroup = group
			}
		}
		if group.Config.Threads > targetThreads {
			if aboveGroup == nil || group.Config.Threads < aboveGroup.Config.Threads {
				aboveGroup = group
			}
		}
	}

	// Linear interpolation/extrapolation
	if belowGroup != nil && aboveGroup != nil {
		// Interpolate between two points
		t1, t2 := belowGroup.Config.Threads, aboveGroup.Config.Threads
		tps1, tps2 := belowGroup.Statistics.TPS.Mean, aboveGroup.Statistics.TPS.Mean

		// Linear interpolation: tps = tps1 + (target-t1) * (tps2-tps1) / (t2-t1)
		slope := (tps2 - tps1) / float64(t2-t1)
		predictedTPS := tps1 + float64(targetThreads-t1)*slope
		return predictedTPS

	} else if belowGroup != nil {
		// Extrapolate from below point
		// Assume diminishing returns (same speedup as last step)
		metrics := CalculateSpeedup(belowGroup.Statistics.TPS.Mean, 1.0) // Rough estimate
		return belowGroup.Statistics.TPS.Mean * metrics
	}

	// Fallback: return highest measured TPS
	maxTPS := 0.0
	for _, group := range groups {
		if group.Statistics.TPS.Mean > maxTPS {
			maxTPS = group.Statistics.TPS.Mean
		}
	}
	return maxTPS
}
