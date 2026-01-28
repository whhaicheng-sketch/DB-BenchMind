// Package usecase provides comparison business logic.
// Implements: Phase 6 - Result comparison and analysis
package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/whhaicheng/DB-BenchMind/internal/domain/comparison"
	"github.com/whhaicheng/DB-BenchMind/internal/domain/execution"
)

// ComparisonUseCase provides comparison business operations.
type ComparisonUseCase struct {
	runRepo RunRepository
}

// NewComparisonUseCase creates a new comparison use case.
func NewComparisonUseCase(runRepo RunRepository) *ComparisonUseCase {
	return &ComparisonUseCase{
		runRepo: runRepo,
	}
}

// CompareRuns compares multiple runs by their IDs.
func (uc *ComparisonUseCase) CompareRuns(ctx context.Context, runIDs []string, baselineID string, compType comparison.ComparisonType) (*comparison.ComparisonResult, error) {
	if len(runIDs) < 2 {
		return nil, fmt.Errorf("at least 2 runs are required for comparison")
	}

	// Fetch all runs
	runs := make([]*execution.Run, 0, len(runIDs))
	for _, runID := range runIDs {
		run, err := uc.runRepo.FindByID(ctx, runID)
		if err != nil {
			return nil, fmt.Errorf("get run %s: %w", runID, err)
		}
		runs = append(runs, run)
	}

	// Perform comparison
	result, err := comparison.CompareRuns(runs, baselineID, compType)
	if err != nil {
		return nil, fmt.Errorf("compare runs: %w", err)
	}

	return result, nil
}

// CompareRecentRuns compares the most recent N runs.
func (uc *ComparisonUseCase) CompareRecentRuns(ctx context.Context, count int, baselineID string) (*comparison.ComparisonResult, error) {
	if count < 2 {
		return nil, fmt.Errorf("at least 2 runs are required for comparison")
	}

	// Fetch recent runs
	allRuns, err := uc.runRepo.FindAll(ctx, FindOptions{
		Limit:    count,
		SortBy:   "created_at",
		SortOrder: "DESC",
	})
	if err != nil {
		return nil, fmt.Errorf("fetch runs: %w", err)
	}

	if len(allRuns) < 2 {
		return nil, fmt.Errorf("not enough runs to compare (found %d)", len(allRuns))
	}

	// If no baseline specified, use the oldest (last in list)
	if baselineID == "" {
		baselineID = allRuns[len(allRuns)-1].ID
	}

	return uc.CompareRuns(ctx, extractRunIDs(allRuns), baselineID, comparison.ComparisonTypeTrend)
}

// CreateComparison creates a saved comparison.
func (uc *ComparisonUseCase) CreateComparison(ctx context.Context, name string, runIDs []string, baselineID string, compType comparison.ComparisonType) (*comparison.Comparison, error) {
	comp := &comparison.Comparison{
		ID:         fmt.Sprintf("cmp-%s", uuid.New().String()),
		Name:       name,
		Type:       compType,
		RunIDs:     runIDs,
		BaselineID: baselineID,
		CreatedAt:  time.Now(),
		Metadata:   make(map[string]string),
	}

	if err := comp.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Note: In a real implementation, you'd save this to a repository
	return comp, nil
}

// GetTrendAnalysis analyzes the trend of a specific metric across runs.
func (uc *ComparisonUseCase) GetTrendAnalysis(ctx context.Context, runIDs []string, metric string) (*TrendAnalysis, error) {
	if len(runIDs) < 2 {
		return nil, fmt.Errorf("at least 2 runs are required for trend analysis")
	}

	runs := make([]*execution.Run, 0, len(runIDs))
	for _, runID := range runIDs {
		run, err := uc.runRepo.FindByID(ctx, runID)
		if err != nil {
			return nil, fmt.Errorf("get run %s: %w", runID, err)
		}
		runs = append(runs, run)
	}

	analysis := &TrendAnalysis{
		Metric: metric,
		Values: make([]TrendValue, len(runs)),
	}

	// Extract values
	for i, run := range runs {
		var value float64
		switch metric {
		case "tps":
			value = extractTPS(run)
		case "latency_avg":
			value = extractLatencyAvg(run)
		case "latency_p95":
			value = extractLatencyP95(run)
		case "error_rate":
			value = extractErrorRate(run)
		default:
			return nil, fmt.Errorf("unknown metric: %s", metric)
		}

		analysis.Values[i] = TrendValue{
			RunID:    run.ID,
			RunName:  run.TaskID, // Use TaskID instead of GetName
			Value:    value,
			Datetime: run.CreatedAt,
		}
	}

	// Calculate trend
	analysis.calculateTrend()

	return analysis, nil
}

// TrendAnalysis represents a trend analysis for a metric.
type TrendAnalysis struct {
	Metric    string       `json:"metric"`
	Values    []TrendValue `json:"values"`
	Trend     string       `json:"trend"`     // "increasing", "decreasing", "stable"
	ChangePct float64      `json:"change_pct"`
	MinValue  float64      `json:"min_value"`
	MaxValue  float64      `json:"max_value"`
	AvgValue  float64      `json:"avg_value"`
}

// TrendValue represents a single value in the trend.
type TrendValue struct {
	RunID    string    `json:"run_id"`
	RunName  string    `json:"run_name"`
	Value    float64   `json:"value"`
	Datetime time.Time `json:"datetime"`
}

// calculateTrend calculates the overall trend.
func (ta *TrendAnalysis) calculateTrend() {
	if len(ta.Values) < 2 {
		return
	}

	first := ta.Values[0].Value
	last := ta.Values[len(ta.Values)-1].Value

	// Calculate percentage change
	if first != 0 {
		ta.ChangePct = ((last - first) / first) * 100
	}

	// Determine trend direction
	if ta.ChangePct > 5 {
		ta.Trend = "increasing"
	} else if ta.ChangePct < -5 {
		ta.Trend = "decreasing"
	} else {
		ta.Trend = "stable"
	}

	// Calculate min/max/avg
	min := ta.Values[0].Value
	max := ta.Values[0].Value
	var sum float64

	for _, v := range ta.Values {
		if v.Value < min {
			min = v.Value
		}
		if v.Value > max {
			max = v.Value
		}
		sum += v.Value
	}

	ta.MinValue = min
	ta.MaxValue = max
	ta.AvgValue = sum / float64(len(ta.Values))
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
