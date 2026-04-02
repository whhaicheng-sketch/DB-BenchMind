// Package report provides bottleneck judgment domain models.
package report

// OverallStatus represents the overall health status of a benchmark run.
type OverallStatus string

const (
	StatusHealthy  OverallStatus = "healthy"
	StatusWarning  OverallStatus = "warning"
	StatusCritical OverallStatus = "critical"
)

// BottleneckType represents the type of identified bottleneck.
type BottleneckType string

const (
	BottleneckCPU            BottleneckType = "CPU-bound"
	BottleneckIO             BottleneckType = "IO-bound"
	BottleneckLockContention BottleneckType = "Lock/Contention"
	BottleneckNetwork        BottleneckType = "Network/Connection"
	BottleneckMisconfig      BottleneckType = "Misconfiguration"
	BottleneckUnknown        BottleneckType = "Unknown"
)

// BottleneckJudgment contains the bottleneck analysis result.
type BottleneckJudgment struct {
	PrimaryBottleneck BottleneckType `json:"primary_bottleneck"`
	Confidence        float64        `json:"confidence"` // 0.0 to 1.0
	Evidence          []string       `json:"evidence"`
	Recommendations   []string       `json:"recommendations"`
}

// ResourceVerdict contains verdict for a single resource dimension.
type ResourceVerdict struct {
	Resource string  `json:"resource"`
	Avg      float64 `json:"avg,omitempty"`
	P95      float64 `json:"p95,omitempty"`
	Peak     float64 `json:"peak,omitempty"`
	UsedPct  float64 `json:"used_pct,omitempty"`
	Verdict  string  `json:"verdict"` // normal, elevated, high, critical

	// Extended fields for specific resources
	AvgMB   float64 `json:"avg_mb,omitempty"`
	PeakMB  float64 `json:"peak_mb,omitempty"`
	RxPeak  float64 `json:"rx_peak,omitempty"`
	TxPeak  float64 `json:"tx_peak,omitempty"`
}

// ComparisonResult contains the comparison between current and previous/baseline.
type ComparisonResult struct {
	PreviousReportID string              `json:"previous_report_id,omitempty"`
	BaselineReportID string              `json:"baseline_report_id,omitempty"`
	Deltas           []ComparisonDelta   `json:"deltas"`
	TrendDirection   string              `json:"trend_direction"` // improved, degraded, stable, unknown
}

// ComparisonDelta represents the change in a single metric.
type ComparisonDelta struct {
	Metric    string  `json:"metric"`
	Current   float64 `json:"current"`
	Previous  float64 `json:"previous"`
	Delta     float64 `json:"delta"`
	PctChange float64 `json:"pct_change"`
}

// DetailedDataPreview contains the preview data for the collapsed detailed section.
type DetailedDataPreview struct {
	BundleFilename    string                `json:"bundle_filename"`
	CompressedSize    int64                 `json:"compressed_size_bytes"`
	SamplingPolicy    string                `json:"sampling_policy"`
	RetainedWindows   []WindowPreview       `json:"retained_windows"`
	AnomalyWindows    []AnomalyPreview      `json:"anomaly_windows"`
	RawSampleCount    int                   `json:"raw_sample_count"`
	SchemaPreview     map[string]interface{} `json:"schema_preview"`
}

// WindowPreview is a summary of a retained window for display.
type WindowPreview struct {
	Name        string `json:"name"`
	SampleCount int    `json:"sample_count"`
}

// AnomalyPreview is a summary of an anomaly window for display.
type AnomalyPreview struct {
	Type     string  `json:"type"`
	Severity string  `json:"severity"`
	Summary  string  `json:"summary"`
	Value    float64 `json:"value"`
}

// HumanReportData contains all data needed to generate the human report.
type HumanReportData struct {
	RunMeta              RunMetaSection
	OverallStatus        OverallStatus
	OneLineConclusion    string
	CoreKPIs             CoreKPIsSection
	ResourceBottleneck   ResourceBottleneckSection
	BottleneckJudgment   BottleneckJudgment
	Comparison           *ComparisonResult
	TrendHighlights      []string
	Recommendation       string
	DetailedDataPreview  *DetailedDataPreview
}

// RunMetaSection contains run metadata for the human report.
type RunMetaSection struct {
	RunID         string
	BenchmarkType string
	TemplateName  string
	TargetName    string
	StartTime     string
	EndTime       string
	DurationMs    int64
}

// CoreKPIsSection contains core KPI metrics for the human report.
type CoreKPIsSection struct {
	TPS           float64
	TPM           float64
	AvgLatencyMs  float64
	P95LatencyMs  float64
	P99LatencyMs  float64
	ErrorRate     float64
	CPUPeak       float64
	DiskUtilPeak  float64
}

// ResourceBottleneckSection contains resource bottleneck summaries.
type ResourceBottleneckSection struct {
	CPU     ResourceVerdict
	Memory  ResourceVerdict
	Disk    ResourceVerdict
	Network ResourceVerdict
}
