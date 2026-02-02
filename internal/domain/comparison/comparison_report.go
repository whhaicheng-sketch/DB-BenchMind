// Package comparison provides comprehensive result comparison functionality.
// This file implements the new domain models for professional-grade
// multi-configuration comparison reports with grouped runs.
package comparison

import (
	"fmt"
	"time"
)

// Additional GroupByField constants for enhanced grouping.
const (
	// GroupByConnection groups results by connection (different server instances).
	GroupByConnection GroupByField = "connection"
	// GroupByCustom allows user-defined custom grouping.
	GroupByCustom GroupByField = "custom"
)

// ConfigSpec defines the configuration dimensions that determine "same configuration".
// Two runs with matching ConfigSpec are considered the same configuration.
type ConfigSpec struct {
	// Core dimensions (must match for same config)
	Threads        int    `json:"threads"`
	DatabaseType   string `json:"database_type"`
	TemplateName   string `json:"template_name"`

	// Optional dimensions (user can choose whether to consider)
	ConnectionName string `json:"connection_name,omitempty"`

	// Future extensions (optional)
	// BufferPoolSize string `json:"buffer_pool_size,omitempty"`
	// DBVersion      string `json:"db_version,omitempty"`
}

// Equals checks if two ConfigSpec are equal.
func (c *ConfigSpec) Equals(other *ConfigSpec) bool {
	if c == nil || other == nil {
		return c == other
	}
	return c.Threads == other.Threads &&
		c.DatabaseType == other.DatabaseType &&
		c.TemplateName == other.TemplateName &&
		c.ConnectionName == other.ConnectionName
}

// String returns a string representation of the config.
func (c *ConfigSpec) String() string {
	if c == nil {
		return "unknown"
	}
	conn := ""
	if c.ConnectionName != "" {
		conn = "@" + c.ConnectionName
	}
	return c.DatabaseType + conn + " (" + c.TemplateName + ", " + string(rune(c.Threads)) + " threads)"
}

// Run represents a single benchmark run with all its metrics.
// This is extracted from a history.Record for statistical analysis.
type Run struct {
	// Identification
	RunID      string    `json:"run_id"`
	StartTime  time.Time `json:"start_time"`
	Duration   time.Duration `json:"duration"`

	// Throughput metrics
	TPS        float64 `json:"tps"`         // Transactions per second
	QPS        float64 `json:"qps"`         // Queries per second

	// Latency metrics (milliseconds)
	LatencyAvg float64 `json:"latency_avg_ms"`
	LatencyMin float64 `json:"latency_min_ms"`
	LatencyMax float64 `json:"latency_max_ms"`
	LatencyP95 float64 `json:"latency_p95_ms"`
	LatencyP99 float64 `json:"latency_p99_ms"`

	// Query distribution
	ReadQueries  int64 `json:"read_queries"`
	WriteQueries int64 `json:"write_queries"`
	OtherQueries int64 `json:"other_queries"`
	TotalQueries int64 `json:"total_queries"`

	// Transaction metrics
	TotalTransactions int64 `json:"total_transactions"`

	// Reliability metrics
	Errors      int64 `json:"errors"`
	Reconnects  int64 `json:"reconnects"`

	// Other metrics
	TotalTime   float64 `json:"total_time_seconds"`
	TotalEvents int64  `json:"total_events"`

	// Query mix
	QueriesPerTransaction float64 `json:"queries_per_transaction"`
}

// RunMetricStats represents statistical analysis of a single metric across N runs.
// This is used for grouped runs where we calculate statistics across multiple executions.
type RunMetricStats struct {
	N         int      `json:"n"`          // Number of runs
	Mean      float64  `json:"mean"`       // Mean (average)
	StdDev    float64  `json:"std_dev"`    // Standard deviation
	Min       float64  `json:"min"`        // Minimum value
	Max       float64  `json:"max"`        // Maximum value
	Values    []float64 `json:"values,omitempty"` // Individual values (for debugging)
}

// IsValid checks if the stats are valid (N > 0).
func (m *RunMetricStats) IsValid() bool {
	return m != nil && m.N > 0
}

// FormatMeanStdDev returns formatted string "mean ± stddev".
func (m *RunMetricStats) FormatMeanStdDev() string {
	if !m.IsValid() {
		return "N/A"
	}
	if m.N == 1 {
		return formatFloat(m.Mean)
	}
	return formatFloat(m.Mean) + " ± " + formatFloat(m.StdDev)
}

// FormatMinMax returns formatted string "(min..max)".
func (m *RunMetricStats) FormatMinMax() string {
	if !m.IsValid() {
		return "N/A"
	}
	if m.N == 1 {
		return formatFloat(m.Min)
	}
	return formatFloat(m.Min) + " .. " + formatFloat(m.Max)
}

// RunStats contains aggregated statistics across N runs of the same configuration.
type RunStats struct {
	// Number of runs
	N int `json:"n"`

	// Throughput statistics
	TPS RunMetricStats `json:"tps"`
	QPS RunMetricStats `json:"qps"`

	// Latency statistics
	LatencyAvg RunMetricStats `json:"latency_avg_ms"`
	LatencyP95 RunMetricStats `json:"latency_p95_ms"`
	LatencyP99 RunMetricStats `json:"latency_p99_ms"`
	LatencyMax float64      `json:"latency_max_ms"` // Max-of-max across all runs

	// Reliability statistics
	TotalErrors    int64 `json:"total_errors"`
	TotalReconnects int64 `json:"total_reconnects"`
	HasErrors      bool  `json:"has_errors"`
	AnyNonZero     bool  `json:"any_non_zero"` // Any run has errors or reconnects

	// Query mix (averaged across runs)
	ReadPct      float64 `json:"read_pct"`
	WritePct     float64 `json:"write_pct"`
	OtherPct     float64 `json:"other_pct"`
	QueriesPerTx float64 `json:"queries_per_transaction"`
}

// ConfigGroup represents a group of runs with the same configuration.
// This is used for analyzing N runs of a configuration (e.g., threads=4).
type ConfigGroup struct {
	// Group identification
	GroupID   string    `json:"group_id"` // e.g., "C1", "C2", "C3"
	Config    ConfigSpec `json:"config"`

	// Runs in this group
	Runs      []*Run    `json:"runs"`

	// Aggregated statistics across all runs
	Statistics RunStats  `json:"statistics"`

	// Additional metadata
	Tags      []string  `json:"tags,omitempty"` // e.g., "baseline", "best"
}

// GetThreadCount returns the thread count for this config group.
func (g *ConfigGroup) GetThreadCount() int {
	return g.Config.Threads
}

// GetDatabaseType returns the database type for this config group.
func (g *ConfigGroup) GetDatabaseType() string {
	return g.Config.DatabaseType
}

// ComparisonReport is a comprehensive multi-configuration comparison report.
// This is the main structure for professional-grade benchmark analysis.
type ComparisonReport struct {
	// Metadata
	GeneratedAt    time.Time     `json:"generated_at"`
	ReportID       string        `json:"report_id"`
	GroupBy        GroupByField  `json:"group_by"`

	// Experiment data
	ConfigGroups   []*ConfigGroup `json:"config_groups"`

	// Analysis results
	ScalingAnalysis *ScalingAnalysis `json:"scaling_analysis,omitempty"`
	SanityChecks    *SanityCheckResults `json:"sanity_checks,omitempty"`
	Findings        *ReportFindings `json:"findings,omitempty"`

	// Report settings
	SimilarityConfig *SimilarityConfig `json:"similarity_config,omitempty"`
}

// SimilarityConfig defines how to detect and group similar runs.
type SimilarityConfig struct {
	// Time window for detecting similar runs
	TimeWindow time.Duration `json:"time_window"` // e.g., 5 minutes

	// Whether to require exact match on all config dimensions
	RequireExactMatch bool `json:"require_exact_match"`

	// Primary grouping dimension
	GroupBy GroupByField `json:"group_by"`

	// Optional: consider connection name in grouping
	ConsiderConnection bool `json:"consider_connection"`
}

// DefaultSimilarityConfig returns default similarity detection settings.
func DefaultSimilarityConfig() *SimilarityConfig {
	return &SimilarityConfig{
		TimeWindow:         5 * time.Minute,
		RequireExactMatch:  true,
		GroupBy:            GroupByThreads,
		ConsiderConnection: false,
	}
}

// ScalingAnalysis contains scaling efficiency analysis.
type ScalingAnalysis struct {
	// Baseline
	BaselineTPS      float64 `json:"baseline_tps"`       // TPS for threads=1
	BaselineGroup    *ConfigGroup `json:"baseline_group"` // Reference to baseline group

	// Speedup and efficiency for each config group
	ByGroup         map[string]*ScalingMetrics `json:"by_group"` // group_id -> metrics

	// Key findings
	BestTPSConfig   *ConfigGroup `json:"best_tps_config"`   // Config with highest TPS
	WorstLatencyConfig *ConfigGroup `json:"worst_latency_config"` // Config with worst p95

	// Scaling knee (diminishing returns point)
	ScalingKnee     *ConfigGroup `json:"scaling_knee,omitempty"` // Where efficiency drops
	ScalingKneeThread int       `json:"scaling_knee_thread,omitempty"` // Thread count
}

// ScalingMetrics represents scaling metrics for a single config.
type ScalingMetrics struct {
	Speedup     float64 `json:"speedup"`     // Speedup vs baseline (TPS / baselineTPS)
	Efficiency  float64 `json:"efficiency"`  // Efficiency = Speedup / threads
	DeltaTPS    float64 `json:"delta_tps"`   // TPS change vs previous config
	DeltaP95    float64 `json:"delta_p95"`   // p95 latency change vs previous
}

// SanityCheckResults contains validation check results.
type SanityCheckResults struct {
	AllPassed bool           `json:"all_passed"`
	Checks    []SanityCheck  `json:"checks"`
}

// SanityCheck represents a single validation check.
type SanityCheck struct {
	Name    string `json:"name"`
	Passed  bool   `json:"passed"`
	Details string `json:"details"`
}

// ReportFindings contains auto-generated findings and recommendations.
type ReportFindings struct {
	BestThroughput     string `json:"best_throughput"`
	ScalingKnee        string `json:"scaling_knee"`
	LatencyRisk        string `json:"latency_risk"`
	StabilityConcerns  string `json:"stability_concerns"`
	Recommendation     string `json:"recommendation"`
	TradeoffStatement  string `json:"tradeoff_statement"`
	NextExperiment     string `json:"next_experiment"`
}

// FormatReportID generates a unique report ID.
func FormatReportID() string {
	return "report-" + time.Now().Format("20060102-150405")
}

// formatFloat formats a float with 2 decimal places.
func formatFloat(f float64) string {
	return fmt.Sprintf("%.2f", f)
}
