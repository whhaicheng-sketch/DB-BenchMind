// Package report provides AI Bundle domain models.
package report

// BundleSchemaVersion is the current schema version for AI bundles.
const BundleSchemaVersion = "v1"

// Bundle represents the AI analysis data bundle.
type Bundle struct {
	SchemaVersion        string               `json:"schema_version"`
	RunMeta              BundleRunMeta        `json:"run_meta"`
	BenchmarkSummary     BundleBenchmarkSummary `json:"benchmark_summary"`
	ResourceSummary      BundleResourceSummary  `json:"resource_summary"`
	TimeseriesDownsampled BundleTimeseries       `json:"timeseries_downsampled"`
	PhaseBreakdown       BundlePhaseBreakdown   `json:"phase_breakdown,omitempty"`
	RetainedWindows      []BundleWindow         `json:"retained_windows"`
	AnomalyWindows       []BundleAnomalyWindow  `json:"anomaly_windows"`
	Comparison           *BundleComparison      `json:"comparison,omitempty"`
	RawSamples           []BundleRawSample      `json:"raw_samples"`
	AIMeta               BundleAIMeta           `json:"ai_meta"`
}

// BundleRunMeta contains run metadata for the bundle.
type BundleRunMeta struct {
	RunID         string `json:"run_id"`
	BenchmarkType string `json:"benchmark_type"`
	TemplateName  string `json:"template_name"`
	TargetName    string `json:"target_name"`
	DatabaseType  string `json:"database_type"`
	StartTime     string `json:"start_time"`
	EndTime       string `json:"end_time,omitempty"`
	DurationMs    int64  `json:"duration_ms"`
}

// BundleBenchmarkSummary contains normalized benchmark metrics.
type BundleBenchmarkSummary struct {
	Throughput   BundleThroughputData    `json:"throughput"`
	Latency      BundleLatencyData      `json:"latency"`
	Errors       BundleErrorData        `json:"errors"`
	ToolSpecific map[string]interface{} `json:"tool_specific,omitempty"`
}

// BundleThroughputData contains throughput metrics.
type BundleThroughputData struct {
	TPS float64 `json:"tps"`
	TPM float64 `json:"tpm"`
	QPS float64 `json:"qps,omitempty"`
}

// BundleLatencyData contains latency metrics.
type BundleLatencyData struct {
	AvgMs float64 `json:"avg_ms"`
	P95Ms float64 `json:"p95_ms"`
	P99Ms float64 `json:"p99_ms"`
	MinMs float64 `json:"min_ms,omitempty"`
	MaxMs float64 `json:"max_ms,omitempty"`
}

// BundleErrorData contains error metrics.
type BundleErrorData struct {
	ErrorCount int64   `json:"error_count"`
	ErrorRate  float64 `json:"error_rate,omitempty"`
}

// BundleResourceSummary contains resource metrics.
type BundleResourceSummary struct {
	CPU     BundleCPUData     `json:"cpu"`
	Memory  BundleMemoryData  `json:"memory"`
	Disk    BundleDiskData    `json:"disk"`
	Network BundleNetworkData `json:"network"`
}

// BundleCPUData contains CPU metrics.
type BundleCPUData struct {
	Avg       float64 `json:"avg,omitempty"`
	P95       float64 `json:"p95,omitempty"`
	Peak      float64 `json:"peak,omitempty"`
	Verdict   string  `json:"verdict"`
}

// BundleMemoryData contains memory metrics.
type BundleMemoryData struct {
	AvgMB     float64 `json:"avg_mb,omitempty"`
	PeakMB    float64 `json:"peak_mb,omitempty"`
	UsedPct   float64 `json:"used_pct,omitempty"`
	Verdict   string  `json:"verdict"`
}

// BundleDiskData contains disk metrics.
type BundleDiskData struct {
	ReadIOPSPeak    float64 `json:"read_iops_peak,omitempty"`
	WriteIOPSPeak   float64 `json:"write_iops_peak,omitempty"`
	UtilPeak        float64 `json:"util_peak,omitempty"`
	AwaitPeak       float64 `json:"await_peak,omitempty"`
	Verdict         string  `json:"verdict"`
}

// BundleNetworkData contains network metrics.
type BundleNetworkData struct {
	RxPeak      float64 `json:"rx_peak,omitempty"`
	TxPeak      float64 `json:"tx_peak,omitempty"`
	Retransmits int64   `json:"retransmits,omitempty"`
	Errors      int64   `json:"errors,omitempty"`
	Verdict     string  `json:"verdict"`
}

// BundleTimeseries contains downsampled time series data.
type BundleTimeseries struct {
	Metrics []BundleTimeseriesMetric `json:"metrics"`
}

// BundleTimeseriesMetric contains a single downsampled metric series.
type BundleTimeseriesMetric struct {
	Name      string                   `json:"name"`
	Buckets   []BundleTimeseriesBucket `json:"buckets"`
	PointCount int                     `json:"point_count"`
}

// BundleTimeseriesBucket is a single bucket in a downsampled series.
type BundleTimeseriesBucket struct {
	Timestamp int64   `json:"timestamp"`
	Avg       float64 `json:"avg"`
	Min       float64 `json:"min"`
	Max       float64 `json:"max"`
	P95       float64 `json:"p95,omitempty"`
}

// BundlePhaseBreakdown contains phase timing data.
type BundlePhaseBreakdown struct {
	Phases []BundlePhase `json:"phases"`
}

// BundlePhase is a single benchmark phase.
type BundlePhase struct {
	Name       string `json:"name"`
	DurationMs int64  `json:"duration_ms"`
	Status     string `json:"status"`
}

// BundleWindow represents a retained time window.
type BundleWindow struct {
	Name       string  `json:"name"`       // front, middle, tail
	StartPct   float64 `json:"start_pct"`
	EndPct     float64 `json:"end_pct"`
	Summary    BundleWindowSummary `json:"summary"`
}

// BundleWindowSummary contains summary statistics for a window.
type BundleWindowSummary struct {
	TPS       WindowStat `json:"tps"`
	Latency   WindowStat `json:"latency"`
	CPU       WindowStat `json:"cpu"`
	SampleCount int      `json:"sample_count"`
}

// WindowStat contains statistical summary for a metric in a window.
type WindowStat struct {
	Avg float64 `json:"avg"`
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

// BundleAnomalyWindow represents an anomaly detected during benchmark.
type BundleAnomalyWindow struct {
	Type      string  `json:"type"`      // latency_spike, error_burst, tps_drop, cpu_pressure, disk_pressure
	StartIdx  int     `json:"start_idx"`
	EndIdx    int     `json:"end_idx"`
	Severity  string  `json:"severity"`  // low, medium, high, critical
	Value     float64 `json:"value"`
	Threshold float64 `json:"threshold"`
	Summary   string  `json:"summary"`
}

// BundleComparison contains comparison data with previous/baseline.
type BundleComparison struct {
	PreviousID string                  `json:"previous_id,omitempty"`
	BaselineID string                  `json:"baseline_id,omitempty"`
	Deltas     []BundleComparisonDelta `json:"deltas"`
}

// BundleComparisonDelta represents a single metric comparison.
type BundleComparisonDelta struct {
	Metric    string  `json:"metric"`
	Current   float64 `json:"current"`
	Previous  float64 `json:"previous"`
	Delta     float64 `json:"delta"`
	PctChange float64 `json:"pct_change"`
}

// BundleRawSample is a high-value raw sample kept for AI analysis.
type BundleRawSample struct {
	Index     int                    `json:"index"`
	Timestamp int64                  `json:"timestamp"`
	Type      string                 `json:"type"` // normal, anomaly, boundary
	Data      map[string]interface{} `json:"data"`
}

// BundleAIMeta contains metadata about the bundle generation.
type BundleAIMeta struct {
	CompressedSizeBytes int64  `json:"compressed_size_bytes"`
	TruncationApplied   bool   `json:"truncation_applied"`
	TruncationReason    string `json:"truncation_reason,omitempty"`
	SamplingPolicy      string `json:"sampling_policy"`
	GeneratedAt         string `json:"generated_at"`
}
