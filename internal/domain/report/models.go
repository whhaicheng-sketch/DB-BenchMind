// Package report provides report domain models.
package report

import (
	"encoding/json"
	"fmt"
	"time"
)

// FlexibleTime is a time.Time that can unmarshal from both Unix timestamps (int64/float64)
// and RFC 3339 strings. The write path is unchanged (produces RFC 3339 via time.Time).
type FlexibleTime struct {
	time.Time
}

// UnmarshalJSON implements json.Unmarshaler. It accepts:
//   - JSON number (int64 or float64 Unix seconds)
//   - JSON string (RFC 3339)
func (ft *FlexibleTime) UnmarshalJSON(data []byte) error {
	// Try string (RFC 3339) first — standard time.Time format.
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			return fmt.Errorf("FlexibleTime: parse RFC3339 string: %w", err)
		}
		ft.Time = t
		return nil
	}

	// Try numeric (Unix seconds — written by buildMetricsJSON as s.Timestamp.Unix()).
	var n float64
	if err := json.Unmarshal(data, &n); err == nil {
		ft.Time = time.Unix(int64(n), 0)
		return nil
	}

	return fmt.Errorf("FlexibleTime: unsupported JSON value: %s", string(data))
}

// MarshalJSON delegates to time.Time.MarshalJSON (RFC 3339).
func (ft FlexibleTime) MarshalJSON() ([]byte, error) {
	return ft.Time.MarshalJSON()
}

// ReportStatus represents the status of a report.
type ReportStatus string

const (
	StatusPending   ReportStatus = "pending"
	StatusRunning   ReportStatus = "running"
	StatusCompleted ReportStatus = "completed"
	StatusFailed    ReportStatus = "failed"
	StatusCancelled ReportStatus = "cancelled"
)

// IsTerminal returns true if the status is terminal.
func (s ReportStatus) IsTerminal() bool {
	return s == StatusCompleted || s == StatusFailed || s == StatusCancelled
}

// SourceType represents the source of a report.
type SourceType string

const (
	SourceTypeBenchmark SourceType = "benchmark"
	SourceTypeAutoBench SourceType = "autobench"
)

// SuiteID constants
const (
	StandaloneSuiteID = "standalone"
)

// IsStandalone returns true if the suite_id represents a standalone benchmark.
func IsStandalone(suiteID string) bool {
	return suiteID == StandaloneSuiteID
}

// IsAutoBench returns true if the suite_id represents an AutoBench suite.
func IsAutoBench(suiteID string) bool {
	return suiteID != "" && suiteID != StandaloneSuiteID
}

// Report represents a benchmark report.
type Report struct {
	ID             string       `json:"id"`
	SuiteID        string       `json:"suite_id"`
	SuiteItemID    string       `json:"suite_item_id,omitempty"`
	SourceType     SourceType   `json:"source_type"`
	ConnectionID   string       `json:"connection_id"`
	ConnectionName string       `json:"connection_name,omitempty"`
	DatabaseType   string       `json:"database_type"`
	TemplateID     string       `json:"template_id,omitempty"`
	TemplateName   string       `json:"template_name,omitempty"`
	StartedAt      time.Time    `json:"started_at"`
	EndedAt        *time.Time   `json:"ended_at,omitempty"`
	DurationMs     int64        `json:"duration_ms,omitempty"`
	Status         ReportStatus `json:"status"`
	ErrorMessage   string       `json:"error_message,omitempty"`

	// Core metrics
	TPM          float64 `json:"tpm,omitempty"`
	TPS          float64 `json:"tps,omitempty"`
	QPS          float64 `json:"qps,omitempty"`
	Throughput   float64 `json:"throughput,omitempty"`
	LatencyAvgMs float64 `json:"latency_avg_ms,omitempty"`
	LatencyP95Ms float64 `json:"latency_p95_ms,omitempty"`
	LatencyP99Ms float64 `json:"latency_p99_ms,omitempty"`
	ErrorCount   int64   `json:"error_count,omitempty"`

	// File paths
	MetricsJSONPath    string `json:"metrics_json_path,omitempty"`
	MonitoringJSONPath string `json:"monitoring_json_path,omitempty"`
	RawJSONPath        string `json:"raw_json_path,omitempty"`
	ReportHTMLPath     string `json:"report_html_path,omitempty"`
	SummaryJSONPath    string `json:"summary_json_path,omitempty"`

	// Metadata
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Tags      string    `json:"tags,omitempty"`
}

// IsCompleted returns true if the report is in a terminal state.
func (r *Report) IsCompleted() bool {
	return r.Status.IsTerminal()
}

// SuiteStatus represents the status of a suite.
type SuiteStatus string

const (
	SuiteStatusDraft          SuiteStatus = "draft"
	SuiteStatusRunning        SuiteStatus = "running"
	SuiteStatusPartialSuccess SuiteStatus = "partial_success"
	SuiteStatusSuccess        SuiteStatus = "success"
	SuiteStatusFailed         SuiteStatus = "failed"
	SuiteStatusCancelled      SuiteStatus = "cancelled"
)

// IsTerminal returns true if the status is terminal.
func (s SuiteStatus) IsTerminal() bool {
	return s == SuiteStatusSuccess || s == SuiteStatusFailed ||
		s == SuiteStatusPartialSuccess || s == SuiteStatusCancelled
}

// Suite represents an AutoBench suite report.
type Suite struct {
	ID                    string      `json:"id"`
	Name                  string      `json:"name,omitempty"`
	ExecutionMode         string      `json:"execution_mode,omitempty"`
	FailurePolicy         string      `json:"failure_policy,omitempty"`
	CleanupEnabled        bool        `json:"cleanup_enabled"`
	SuiteManifestJSONPath string      `json:"suite_manifest_json_path,omitempty"`
	Status                SuiteStatus `json:"status"`
	StartedAt             *time.Time  `json:"started_at,omitempty"`
	EndedAt               *time.Time  `json:"ended_at,omitempty"`
	TotalItems            int         `json:"total_items"`
	CompletedItems        int         `json:"completed_items"`
	SuccessItems          int         `json:"success_items"`
	FailedItems           int         `json:"failed_items"`
	SkippedItems          int         `json:"skipped_items"`
	SuiteReportJSONPath   string      `json:"suite_report_json_path,omitempty"`
	SuiteReportHTMLPath   string      `json:"suite_report_html_path,omitempty"`
	CreatedAt             time.Time   `json:"created_at"`
	UpdatedAt             time.Time   `json:"updated_at"`
}

// IsCompleted returns true if the suite is in a terminal state.
func (s *Suite) IsCompleted() bool {
	return s.Status.IsTerminal()
}

// ReportContext provides context for report collection.
type ReportContext struct {
	SuiteID        string
	SuiteItemID    string
	SourceType     SourceType
	ConnectionID   string
	ConnectionName string
	DatabaseType   string
	TemplateID     string
	TemplateName   string
	Tags           string
}

// ReportResult is the result of CollectAndPersist.
type ReportResult struct {
	ReportID     string
	ReportPaths  ReportPaths
	Summary      ReportSummary
	PersistError error
}

// ReportPaths contains paths to report files.
type ReportPaths struct {
	MetricsJSON    string
	MonitoringJSON string
	RawJSON        string
	ReportHTML     string
	SummaryJSON    string
}

// ReportSummary contains key metrics for display.
type ReportSummary struct {
	Status       ReportStatus
	TPM          float64
	TPS          float64
	LatencyAvgMs float64
	LatencyP95Ms float64
	ErrorCount   int64
}

// MetricsData represents the full metrics data loaded from metrics.json.
type MetricsData struct {
	SchemaVersion string                  `json:"schema_version"`
	ReportID      string                  `json:"report_id"`
	SuiteID       string                  `json:"suite_id,omitempty"`
	SuiteItemID   string                  `json:"suite_item_id,omitempty"`
	GeneratedAt   string                  `json:"generated_at,omitempty"`
	Benchmark     *MetricsBenchmark        `json:"benchmark,omitempty"`
	Execution     *MetricsExecution        `json:"execution,omitempty"`
	Summary       *MetricsSummaryData      `json:"summary,omitempty"`
	Percentiles   map[string]float64    `json:"percentiles,omitempty"`
	TimeSeries    []MetricsTimeSeriesItem `json:"time_series,omitempty"`
}

// MetricsBenchmark contains benchmark metadata.
type MetricsBenchmark struct {
	ConnectionID   string `json:"connection_id,omitempty"`
	ConnectionName string `json:"connection_name,omitempty"`
	DatabaseType   string `json:"database_type,omitempty"`
	TemplateID     string `json:"template_id,omitempty"`
	TemplateName   string `json:"template_name,omitempty"`
}

// MetricsExecution contains execution metadata.
type MetricsExecution struct {
	Status     string `json:"status,omitempty"`
	StartedAt  string `json:"started_at,omitempty"`
	DurationMs int64  `json:"duration_ms,omitempty"`
}

// MetricsSummaryData contains summary metrics.
type MetricsSummaryData struct {
	TPM           float64 `json:"tpm,omitempty"`
	TPS           float64 `json:"tps,omitempty"`
	QPS           float64 `json:"qps,omitempty"`
	Throughput    float64 `json:"throughput,omitempty"`
	LatencyAvgMs  float64 `json:"latency_avg_ms,omitempty"`
	LatencyP95Ms  float64 `json:"latency_p95_ms,omitempty"`
	LatencyP99Ms  float64 `json:"latency_p99_ms,omitempty"`
	ErrorCount    int64   `json:"error_count,omitempty"`
}

// MetricsTimeSeriesItem represents a single time series data point.
type MetricsTimeSeriesItem struct {
	Timestamp  FlexibleTime `json:"timestamp"`
	TPS        float64      `json:"tps,omitempty"`
	LatencyAvg float64      `json:"latency_avg,omitempty"`
}

// MonitoringData represents the monitoring metrics data.
type MonitoringData struct {
	SchemaVersion string                 `json:"schema_version"`
	ReportID      string                 `json:"report_id"`
	GeneratedAt   string                 `json:"generated_at,omitempty"`
	CPU           *MonitoringCPUData     `json:"cpu,omitempty"`
	Memory        *MonitoringMemoryData  `json:"memory,omitempty"`
	Disk          *MonitoringDiskData    `json:"disk,omitempty"`
	Network       *MonitoringNetworkData `json:"network,omitempty"`
	TimeSeries    []MonitoringTimeSeries `json:"time_series,omitempty"`
}

// MonitoringCPUData contains CPU monitoring data.
type MonitoringCPUData struct {
	UsageAvg float64 `json:"usage_avg,omitempty"`
	UsageMax float64 `json:"usage_max,omitempty"`
}

// MonitoringMemoryData contains memory monitoring data.
type MonitoringMemoryData struct {
	UsedAvgMB   float64 `json:"used_avg_mb,omitempty"`
	UsedMaxMB   float64 `json:"used_max_mb,omitempty"`
	UsedPercent float64 `json:"used_percent,omitempty"`
}

// MonitoringDiskData contains disk monitoring data.
type MonitoringDiskData struct {
	ReadBytesPerSec  float64 `json:"read_bytes_per_sec,omitempty"`
	WriteBytesPerSec float64 `json:"write_bytes_per_sec,omitempty"`
}

// MonitoringNetworkData contains network monitoring data.
type MonitoringNetworkData struct {
	RxBytesPerSec float64 `json:"rx_bytes_per_sec,omitempty"`
	TxBytesPerSec float64 `json:"tx_bytes_per_sec,omitempty"`
}

// MonitoringTimeSeries represents a single monitoring time series data point.
type MonitoringTimeSeries struct {
	Timestamp    FlexibleTime `json:"timestamp"`
	CPUUsage     float64      `json:"cpu_usage,omitempty"`
	MemoryUsedMB float64      `json:"memory_used_mb,omitempty"`
}

// RawData represents the raw benchmark output data.
type RawData struct {
	SchemaVersion string          `json:"schema_version"`
	ReportID      string          `json:"report_id"`
	GeneratedAt   string          `json:"generated_at,omitempty"`
	BenchmarkTool string          `json:"benchmark_tool,omitempty"`
	Command       string          `json:"command,omitempty"`
	Stdout        string          `json:"stdout,omitempty"`
	Stderr        string          `json:"stderr,omitempty"`
	ExitCode      int             `json:"exit_code,omitempty"`
	Environment   map[string]string `json:"environment,omitempty"`
}
