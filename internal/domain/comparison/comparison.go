// Package comparison provides result comparison functionality.
// Implements: Phase 6 - Result comparison and analysis
package comparison

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/whhaicheng/DB-BenchMind/internal/domain/execution"
)

// ComparisonType represents the type of comparison.
type ComparisonType string

const (
	// ComparisonTypeBaseline compares runs against a baseline.
	ComparisonTypeBaseline ComparisonType = "baseline"
	// ComparisonTypeTrend compares runs over time for trend analysis.
	ComparisonTypeTrend ComparisonType = "trend"
	// ComparisonTypeMulti compares multiple runs side by side.
	ComparisonTypeMulti ComparisonType = "multi"
)

// Comparison represents a comparison of multiple benchmark runs.
type Comparison struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Type       ComparisonType    `json:"type"`
	RunIDs     []string          `json:"run_ids"`
	BaselineID string            `json:"baseline_id,omitempty"`
	CreatedAt  time.Time         `json:"created_at"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

// ComparisonResult contains the comparison results.
type ComparisonResult struct {
	Comparison *Comparison      `json:"comparison"`
	Runs       []*execution.Run `json:"runs"`
	Metrics    *MetricDiff      `json:"metrics"`
	Summary    *ComparisonSummary `json:"summary"`
	CreatedAt  time.Time        `json:"created_at"`
}

// MetricDiff represents the difference in metrics between runs.
type MetricDiff struct {
	TPSDiff         []MetricValueDiff `json:"tps_diff"`
	LatencyAvgDiff  []MetricValueDiff `json:"latency_avg_diff"`
	LatencyP95Diff  []MetricValueDiff `json:"latency_p95_diff"`
	ErrorRateDiff   []MetricValueDiff `json:"error_rate_diff"`
	BestTps         *MetricStats      `json:"best_tps,omitempty"`
	WorstTps        *MetricStats      `json:"worst_tps,omitempty"`
	BestLatency     *MetricStats      `json:"best_latency,omitempty"`
	WorstLatency    *MetricStats      `json:"worst_latency,omitempty"`
}

// MetricValueDiff represents a single metric difference.
type MetricValueDiff struct {
	RunID     string  `json:"run_id"`
	RunName   string  `json:"run_name"`
	Value     float64 `json:"value"`
	Diff      float64 `json:"diff,omitempty"`     // Difference from baseline
	DiffPct   float64 `json:"diff_pct,omitempty"` // Percentage difference
	IsBaseline bool    `json:"is_baseline"`
	Timestamp string  `json:"timestamp,omitempty"`
}

// MetricStats contains statistical information about metrics.
type MetricStats struct {
	Min     float64   `json:"min"`
	Max     float64   `json:"max"`
	Avg     float64   `json:"avg"`
	Median  float64   `json:"median"`
	RunID   string    `json:"run_id"`
	RunName string    `json:"run_name"`
}

// ComparisonSummary provides a high-level summary of the comparison.
type ComparisonSummary struct {
	TotalRuns       int     `json:"total_runs"`
	BaselineRunID   string  `json:"baseline_run_id,omitempty"`
	OverallTpsTrend string  `json:"overall_tps_trend"` // "improving", "declining", "stable"
	TpsChangePct    float64 `json:"tps_change_pct"`
	BestRun         *RunHighlight `json:"best_run,omitempty"`
	WorstRun        *RunHighlight `json:"worst_run,omitempty"`
	Insights        []string `json:"insights,omitempty"`
}

// RunHighlight highlights an interesting run.
type RunHighlight struct {
	RunID      string  `json:"run_id"`
	RunName    string  `json:"run_name"`
	TPS        float64 `json:"tps"`
	LatencyAvg float64 `json:"latency_avg_ms"`
	ErrorRate  float64 `json:"error_rate"`
	Reason     string  `json:"reason"` // Why this run is highlighted
}

// Validate validates the comparison.
func (c *Comparison) Validate() error {
	if c.ID == "" {
		return fmt.Errorf("comparison ID is required")
	}
	if c.Name == "" {
		return fmt.Errorf("comparison name is required")
	}
	if len(c.RunIDs) < 2 {
		return fmt.Errorf("at least 2 runs are required for comparison")
	}
	if c.Type == "" {
		return fmt.Errorf("comparison type is required")
	}
	validTypes := map[ComparisonType]bool{
		ComparisonTypeBaseline: true,
		ComparisonTypeTrend:    true,
		ComparisonTypeMulti:    true,
	}
	if !validTypes[c.Type] {
		return fmt.Errorf("invalid comparison type: %s", c.Type)
	}
	if c.Type == ComparisonTypeBaseline && c.BaselineID == "" {
		return fmt.Errorf("baseline ID is required for baseline comparison")
	}
	return nil
}

// ToJSON serializes the comparison to JSON.
func (c *Comparison) ToJSON() ([]byte, error) {
	return json.MarshalIndent(c, "", "  ")
}

// FromJSON deserializes a comparison from JSON.
func (c *Comparison) FromJSON(data []byte) error {
	return json.Unmarshal(data, c)
}

// CompareRuns compares multiple runs and generates a comparison result.
func CompareRuns(runs []*execution.Run, baselineID string, compType ComparisonType) (*ComparisonResult, error) {
	if len(runs) < 2 {
		return nil, fmt.Errorf("at least 2 runs are required for comparison")
	}

	comparison := &Comparison{
		ID:        generateComparisonID(),
		Type:      compType,
		RunIDs:    extractRunIDs(runs),
		BaselineID: baselineID,
		CreatedAt: time.Now(),
	}

	if compType == ComparisonTypeBaseline {
		comparison.BaselineID = baselineID
	}

	// Calculate metric differences
	metrics := calculateMetricDiffs(runs, baselineID)

	// Generate summary
	summary := generateSummary(runs, metrics, baselineID)

	result := &ComparisonResult{
		Comparison: comparison,
		Runs:       runs,
		Metrics:    metrics,
		Summary:    summary,
		CreatedAt:  time.Now(),
	}

	return result, nil
}

// calculateMetricDiffs calculates metric differences between runs.
func calculateMetricDiffs(runs []*execution.Run, baselineID string) *MetricDiff {
	diff := &MetricDiff{}

	// Find baseline run if specified
	var baselineRun *execution.Run
	if baselineID != "" {
		for _, run := range runs {
			if run.ID == baselineID {
				baselineRun = run
				break
			}
		}
	}

	// Extract TPS values
	for _, run := range runs {
		tps := extractTPS(run)
		latencyAvg := extractLatencyAvg(run)
		latencyP95 := extractLatencyP95(run)
		errorRate := extractErrorRate(run)

		valueDiff := MetricValueDiff{
			RunID:     run.ID,
			RunName:   run.TaskID, // Use TaskID as name since GetName doesn't exist
			Value:     tps,
			Timestamp: run.CreatedAt.Format(time.RFC3339),
			IsBaseline: run.ID == baselineID,
		}

		if baselineRun != nil && run.ID != baselineID {
			baselineTPS := extractTPS(baselineRun)
			valueDiff.Diff = tps - baselineTPS
			if baselineTPS != 0 {
				valueDiff.DiffPct = ((tps - baselineTPS) / baselineTPS) * 100
			}
		}

		diff.TPSDiff = append(diff.TPSDiff, valueDiff)
		diff.LatencyAvgDiff = append(diff.LatencyAvgDiff, MetricValueDiff{
			RunID:     run.ID,
			RunName:   run.TaskID,
			Value:     latencyAvg,
			IsBaseline: run.ID == baselineID,
		})
		diff.LatencyP95Diff = append(diff.LatencyP95Diff, MetricValueDiff{
			RunID:     run.ID,
			RunName:   run.TaskID,
			Value:     latencyP95,
			IsBaseline: run.ID == baselineID,
		})
		diff.ErrorRateDiff = append(diff.ErrorRateDiff, MetricValueDiff{
			RunID:     run.ID,
			RunName:   run.TaskID,
			Value:     errorRate,
			IsBaseline: run.ID == baselineID,
		})
	}

	// Calculate best and worst stats
	diff.BestTps = findBestMetric(runs, extractTPS, true)
	diff.WorstTps = findBestMetric(runs, extractTPS, false)
	diff.BestLatency = findBestMetric(runs, extractLatencyAvg, true)
	diff.WorstLatency = findBestMetric(runs, extractLatencyAvg, false)

	return diff
}

// generateSummary generates a comparison summary.
func generateSummary(runs []*execution.Run, metrics *MetricDiff, baselineID string) *ComparisonSummary {
	summary := &ComparisonSummary{
		TotalRuns:     len(runs),
		BaselineRunID: baselineID,
		Insights:      []string{},
	}

	if baselineID != "" {
		summary.BaselineRunID = baselineID
		// Calculate TPS change from baseline
		if len(metrics.TPSDiff) > 0 {
			var totalChange float64
			var count int
			for _, diff := range metrics.TPSDiff {
				if !diff.IsBaseline && diff.DiffPct != 0 {
					totalChange += diff.DiffPct
					count++
				}
			}
			if count > 0 {
				summary.TpsChangePct = totalChange / float64(count)
			}
		}

		// Determine trend
		if summary.TpsChangePct > 5 {
			summary.OverallTpsTrend = "improving"
			summary.Insights = append(summary.Insights,
				fmt.Sprintf("TPS improved by %.1f%% compared to baseline", summary.TpsChangePct))
		} else if summary.TpsChangePct < -5 {
			summary.OverallTpsTrend = "declining"
			summary.Insights = append(summary.Insights,
				fmt.Sprintf("TPS declined by %.1f%% compared to baseline", math.Abs(summary.TpsChangePct)))
		} else {
			summary.OverallTpsTrend = "stable"
			summary.Insights = append(summary.Insights, "TPS remained stable compared to baseline")
		}
	}

	// Find best and worst runs
	if metrics.BestTps != nil {
		summary.BestRun = &RunHighlight{
			RunID:      metrics.BestTps.RunID,
			RunName:    metrics.BestTps.RunName,
			TPS:        metrics.BestTps.Max,
			LatencyAvg: 0, // Would need to extract from run
			ErrorRate:  0,
			Reason:     "Highest TPS",
		}
	}

	if metrics.WorstTps != nil {
		summary.WorstRun = &RunHighlight{
			RunID:      metrics.WorstTps.RunID,
			RunName:    metrics.WorstTps.RunName,
			TPS:        metrics.WorstTps.Min,
			LatencyAvg: 0,
			ErrorRate:  0,
			Reason:     "Lowest TPS",
		}
	}

	return summary
}

// Helper functions
func extractRunIDs(runs []*execution.Run) []string {
	ids := make([]string, len(runs))
	for i, run := range runs {
		ids[i] = run.ID
	}
	return ids
}

func extractTPS(run *execution.Run) float64 {
	if run.Result != nil {
		return run.Result.TPSCalculated
	}
	return 0
}

func extractLatencyAvg(run *execution.Run) float64 {
	if run.Result != nil {
		return run.Result.LatencyAvg
	}
	return 0
}

func extractLatencyP95(run *execution.Run) float64 {
	if run.Result != nil {
		return run.Result.LatencyP95
	}
	return 0
}

func extractErrorRate(run *execution.Run) float64 {
	if run.Result != nil {
		return run.Result.ErrorRate
	}
	return 0
}

func findBestMetric(runs []*execution.Run, extractor func(*execution.Run) float64, findMax bool) *MetricStats {
	if len(runs) == 0 {
		return nil
	}

	values := make([]float64, len(runs))
	for i, run := range runs {
		values[i] = extractor(run)
	}

	sort.Float64s(values)
	min := values[0]
	max := values[len(values)-1]

	// Calculate average
	var sum float64
	for _, v := range values {
		sum += v
	}
	avg := sum / float64(len(values))

	// Calculate median
	var median float64
	if len(values)%2 == 0 {
		median = (values[len(values)/2-1] + values[len(values)/2]) / 2
	} else {
		median = values[len(values)/2]
	}

	// Find the run with the best value
	var bestRun *execution.Run
	var bestVal float64
	if findMax {
		bestVal = max
	} else {
		bestVal = min
	}

	for _, run := range runs {
		if extractor(run) == bestVal {
			bestRun = run
			break
		}
	}

	if bestRun == nil {
		return nil
	}

	return &MetricStats{
		Min:     min,
		Max:     max,
		Avg:     avg,
		Median:  median,
		RunID:   bestRun.ID,
		RunName: bestRun.TaskID, // Use TaskID instead of GetName
	}
}

func generateComparisonID() string {
	return fmt.Sprintf("cmp-%d", time.Now().UnixNano())
}
