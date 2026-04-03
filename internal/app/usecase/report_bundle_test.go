package usecase

import (
	"fmt"
	"testing"
	"time"

	"github.com/whhaicheng/DB-BenchMind/internal/domain/execution"
	"github.com/whhaicheng/DB-BenchMind/internal/domain/report"
)

func TestBundleGenerator_Generate(t *testing.T) {
	gen := NewBundleGenerator()

	tests := []struct {
		name    string
		input   *BundleInput
		wantErr bool
	}{
		{
			name: "nil report returns error",
			input: &BundleInput{
				Report: nil,
			},
			wantErr: true,
		},
		{
			name: "basic report generates bundle",
			input: &BundleInput{
				Report: &report.Report{
					ID:             "rpt-001",
					SuiteID:        "standalone",
					ConnectionName: "test-conn",
					TemplateName:   "oltp",
					DatabaseType:   "mysql",
					StartedAt:      time.Now(),
					EndedAt:        nil,
					DurationMs:     60000,
					TPS:            1000.5,
					TPM:            60030.0,
					LatencyAvgMs:   10.5,
					LatencyP95Ms:   25.3,
					LatencyP99Ms:   45.7,
					ErrorCount:     0,
				},
				AdapterType: "sysbench",
			},
			wantErr: false,
		},
		{
			name: "report with samples and monitoring",
			input: &BundleInput{
				Report: &report.Report{
					ID:             "rpt-002",
					SuiteID:        "standalone",
					ConnectionName: "test-conn",
					TemplateName:   "tpcc",
					DatabaseType:   "postgresql",
					StartedAt:      time.Now(),
					EndedAt:        nil,
					DurationMs:     120000,
					TPS:            500.0,
					TPM:            30000.0,
					LatencyAvgMs:   20.0,
					LatencyP95Ms:   50.0,
					LatencyP99Ms:   80.0,
					ErrorCount:     5,
				},
				AdapterType: "sysbench",
				Samples:     generateTestSamples(100, 500.0, 20.0),
				Monitoring: &report.MonitoringData{
					CPU: &report.MonitoringCPUData{
						UsageAvg: 45.0,
						UsageMax: 78.5,
					},
					Memory: &report.MonitoringMemoryData{
						UsedAvgMB:   4096.0,
						UsedMaxMB:   6144.0,
						UsedPercent: 60.0,
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bundle, err := gen.Generate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if bundle == nil {
				t.Fatal("Generate() returned nil bundle")
			}
			if bundle.SchemaVersion != report.BundleSchemaVersion {
				t.Errorf("SchemaVersion = %v, want %v", bundle.SchemaVersion, report.BundleSchemaVersion)
			}
			if bundle.RunMeta.RunID != tt.input.Report.ID {
				t.Errorf("RunMeta.RunID = %v, want %v", bundle.RunMeta.RunID, tt.input.Report.ID)
			}
		})
	}
}

func TestBundleGenerator_GenerateAndCompress_SizeLimit(t *testing.T) {
	gen := NewBundleGenerator()

	tests := []struct {
		name        string
		sampleCount int
	}{
		{
			name:        "small dataset",
			sampleCount: 50,
		},
		{
			name:        "large dataset triggers truncation",
			sampleCount: 10000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := &BundleInput{
				Report: &report.Report{
					ID:             "rpt-size-test",
					SuiteID:        "standalone",
					ConnectionName: "test-conn",
					TemplateName:   "oltp",
					DatabaseType:   "mysql",
					StartedAt:      time.Now(),
					DurationMs:     300000,
					TPS:            1000.0,
					TPM:            60000.0,
					LatencyAvgMs:   10.0,
					LatencyP95Ms:   25.0,
					LatencyP99Ms:   50.0,
				},
				AdapterType: "sysbench",
				Samples:     generateTestSamples(tt.sampleCount, 1000.0, 10.0),
			}

			compressed, bundle, err := gen.GenerateAndCompress(input)
			if err != nil {
				t.Fatalf("GenerateAndCompress() error = %v", err)
			}

			if len(compressed) > MaxBundleCompressedBytes {
				t.Errorf("Compressed size %d exceeds limit %d", len(compressed), MaxBundleCompressedBytes)
			}

			if bundle == nil {
				t.Fatal("Bundle is nil")
			}

			if bundle.AIMeta.CompressedSizeBytes != int64(len(compressed)) {
				t.Errorf("AIMeta.CompressedSizeBytes = %v, want %v", bundle.AIMeta.CompressedSizeBytes, len(compressed))
			}

			// Verify compressed data can be decompressed
			_, decompressErr := DecompressBundle(compressed)
			if decompressErr != nil {
				t.Errorf("DecompressBundle() error = %v", decompressErr)
			}
		})
	}
}

func TestBundleGenerator_AnomalyDetection(t *testing.T) {
	gen := NewBundleGenerator()

	tests := []struct {
		name             string
		samples          []execution.MetricSample
		expectAnomalies  bool
		anomalyType      string
	}{
		{
			name:            "stable metrics no anomalies",
			samples:         generateTestSamples(100, 1000.0, 10.0),
			expectAnomalies: false,
		},
		{
			name: "TPS drop detected",
			samples: func() []execution.MetricSample {
				samples := generateTestSamples(50, 1000.0, 10.0)
				// Inject TPS drop
				for i := 20; i < 30; i++ {
					samples[i].TPS = 100.0 // 90% drop
				}
				return samples
			}(),
			expectAnomalies: true,
			anomalyType:     "tps_drop",
		},
		{
			name: "latency spike detected",
			samples: func() []execution.MetricSample {
				samples := generateTestSamples(50, 1000.0, 10.0)
				// Inject latency spike
				for i := 25; i < 30; i++ {
					samples[i].LatencyAvg = 500.0 // massive spike
				}
				return samples
			}(),
			expectAnomalies: true,
			anomalyType:     "latency_spike",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			anomalies := gen.detectAnomalies(tt.samples, nil)
			if tt.expectAnomalies {
				if len(anomalies) == 0 {
					t.Error("Expected anomalies but got none")
				}
				found := false
				for _, a := range anomalies {
					if a.Type == tt.anomalyType {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected anomaly type %s but not found", tt.anomalyType)
				}
			}
		})
	}
}

func TestBundleGenerator_RetainedWindows(t *testing.T) {
	gen := NewBundleGenerator()

	tests := []struct {
		name        string
		sampleCount int
		expectCount int
	}{
		{
			name:        "too few samples returns nil",
			sampleCount: 2,
			expectCount: 0,
		},
		{
			name:        "enough samples returns 3 windows",
			sampleCount: 100,
			expectCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			samples := generateTestSamples(tt.sampleCount, 500.0, 10.0)
			windows := gen.buildRetainedWindows(samples)
			if len(windows) != tt.expectCount {
				t.Errorf("Expected %d windows, got %d", tt.expectCount, len(windows))
			}
			if tt.expectCount == 3 {
				names := []string{"front", "middle", "tail"}
				for i, w := range windows {
					if w.Name != names[i] {
						t.Errorf("Window[%d].Name = %v, want %v", i, w.Name, names[i])
					}
				}
			}
		})
	}
}

func TestDecompressBundle(t *testing.T) {
	gen := NewBundleGenerator()

	input := &BundleInput{
		Report: &report.Report{
			ID:             "rpt-decompress",
			SuiteID:        "standalone",
			ConnectionName: "test-conn",
			TemplateName:   "oltp",
			DatabaseType:   "mysql",
			StartedAt:      time.Now(),
			DurationMs:     60000,
			TPS:            1000.0,
		},
		AdapterType: "sysbench",
		Samples:     generateTestSamples(50, 500.0, 10.0),
	}

	compressed, _, err := gen.GenerateAndCompress(input)
	if err != nil {
		t.Fatalf("GenerateAndCompress() error = %v", err)
	}

	// Decompress and verify
	bundle, err := DecompressBundle(compressed)
	if err != nil {
		t.Fatalf("DecompressBundle() error = %v", err)
	}
	if bundle.RunMeta.RunID != "rpt-decompress" {
		t.Errorf("Decompressed RunID = %v, want rpt-decompress", bundle.RunMeta.RunID)
	}
}

func TestDownsample(t *testing.T) {
	tests := []struct {
		name        string
		pointCount  int
		bucketCount int
		wantBuckets int
	}{
		{
			name:        "empty returns nil",
			pointCount:  0,
			bucketCount: 10,
			wantBuckets: 0,
		},
		{
			name:        "fewer points than buckets returns point-count buckets",
			pointCount:  5,
			bucketCount: 10,
			wantBuckets: 5,
		},
		{
			name:        "more points gets downsampled",
			pointCount:  100,
			bucketCount: 10,
			wantBuckets: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			points := make([]tsPoint, tt.pointCount)
			for i := range points {
				points[i] = tsPoint{timestamp: int64(i * 1000), value: float64(i)}
			}
			buckets := downsample(points, tt.bucketCount)
			if len(buckets) != tt.wantBuckets {
				t.Errorf("Expected %d buckets, got %d", tt.wantBuckets, len(buckets))
			}
		})
	}
}

// Helper to generate test samples
func generateTestSamples(count int, baseTPS, baseLatency float64) []execution.MetricSample {
	samples := make([]execution.MetricSample, count)
	baseTime := time.Now().Add(-time.Duration(count) * time.Second)
	for i := 0; i < count; i++ {
		samples[i] = execution.MetricSample{
			Timestamp:  baseTime.Add(time.Duration(i) * time.Second),
			TPS:        baseTPS,
			TPM:        baseTPS * 60,
			LatencyAvg: baseLatency,
			Errors:     0,
			ErrorRate:  0,
		}
	}
	return samples
}

// generateTestSamplesWithAnomalies generates samples with a TPS drop in the specified range.
func generateTestSamplesWithAnomalies(count int, baseTPS, baseLatency float64, anomalyStart, anomalyEnd int, anomalyTPS float64) []execution.MetricSample {
	samples := generateTestSamples(count, baseTPS, baseLatency)
	for i := anomalyStart; i < anomalyEnd && i < count; i++ {
		samples[i].TPS = anomalyTPS
	}
	return samples
}

// ---------------------------------------------------------------------------
// 1. Regression tests (targeting recent simplify fixes)
// ---------------------------------------------------------------------------

func TestBuildPreviewFromBundle(t *testing.T) {
	gen := NewBundleGenerator()
	input := &BundleInput{
		Report: &report.Report{
			ID:             "rpt-preview",
			SuiteID:        "standalone",
			ConnectionName: "test-conn",
			TemplateName:   "oltp",
			DatabaseType:   "mysql",
			StartedAt:      time.Now(),
			DurationMs:     60000,
			TPS:            1000.0,
			TPM:            60000.0,
			LatencyAvgMs:   10.0,
			LatencyP95Ms:   25.0,
			LatencyP99Ms:   50.0,
		},
		AdapterType: "sysbench",
		Samples:     generateTestSamples(100, 1000.0, 10.0),
	}

	compressed, bundle, err := gen.GenerateAndCompress(input)
	if err != nil {
		t.Fatalf("GenerateAndCompress() error = %v", err)
	}

	preview := BuildPreviewFromBundle(bundle, compressed, "rpt-preview")
	if preview == nil {
		t.Fatal("BuildPreviewFromBundle returned nil")
	}

	// BundleFilename
	if want := "report_bundle_rpt-preview.json.gz"; preview.BundleFilename != want {
		t.Errorf("BundleFilename = %q, want %q", preview.BundleFilename, want)
	}

	// CompressedSize
	if preview.CompressedSize != int64(len(compressed)) {
		t.Errorf("CompressedSize = %d, want %d", preview.CompressedSize, len(compressed))
	}

	// SamplingPolicy
	if preview.SamplingPolicy != "front_middle_tail_anomaly" {
		t.Errorf("SamplingPolicy = %q, want %q", preview.SamplingPolicy, "front_middle_tail_anomaly")
	}

	// RetainedWindows: expect 3 (front/middle/tail)
	if len(preview.RetainedWindows) != 3 {
		t.Fatalf("RetainedWindows count = %d, want 3", len(preview.RetainedWindows))
	}
	wantNames := []string{"front", "middle", "tail"}
	for i, w := range preview.RetainedWindows {
		if w.Name != wantNames[i] {
			t.Errorf("RetainedWindows[%d].Name = %q, want %q", i, w.Name, wantNames[i])
		}
		if w.SampleCount <= 0 {
			t.Errorf("RetainedWindows[%d].SampleCount = %d, want > 0", i, w.SampleCount)
		}
	}

	// AnomalyWindows: stable metrics → empty
	if len(preview.AnomalyWindows) != 0 {
		t.Errorf("AnomalyWindows count = %d, want 0 for stable metrics", len(preview.AnomalyWindows))
	}

	// RawSampleCount
	if preview.RawSampleCount <= 0 {
		t.Errorf("RawSampleCount = %d, want > 0", preview.RawSampleCount)
	}
}

func TestBuildPreviewFromBundle_WithAnomalies(t *testing.T) {
	gen := NewBundleGenerator()
	// Inject TPS drop to trigger anomaly detection
	samples := generateTestSamplesWithAnomalies(100, 1000.0, 10.0, 40, 55, 100.0)

	input := &BundleInput{
		Report: &report.Report{
			ID:             "rpt-anomaly",
			SuiteID:        "standalone",
			ConnectionName: "test-conn",
			TemplateName:   "oltp",
			DatabaseType:   "mysql",
			StartedAt:      time.Now(),
			DurationMs:     60000,
			TPS:            1000.0,
			TPM:            60000.0,
			LatencyAvgMs:   10.0,
		},
		AdapterType: "sysbench",
		Samples:     samples,
	}

	compressed, bundle, err := gen.GenerateAndCompress(input)
	if err != nil {
		t.Fatalf("GenerateAndCompress() error = %v", err)
	}

	preview := BuildPreviewFromBundle(bundle, compressed, "rpt-anomaly")
	if preview == nil {
		t.Fatal("BuildPreviewFromBundle returned nil")
	}

	if len(preview.AnomalyWindows) == 0 {
		t.Fatal("AnomalyWindows is empty, expected TPS drop anomalies")
	}

	for i, a := range preview.AnomalyWindows {
		if a.Type == "" {
			t.Errorf("AnomalyWindows[%d].Type is empty", i)
		}
		if a.Severity == "" {
			t.Errorf("AnomalyWindows[%d].Severity is empty", i)
		}
		if a.Summary == "" {
			t.Errorf("AnomalyWindows[%d].Summary is empty", i)
		}
		if a.Value == 0 {
			t.Errorf("AnomalyWindows[%d].Value is zero", i)
		}
	}
}

func TestComputeWindowSummary_NoCPUValues(t *testing.T) {
	samples := []execution.MetricSample{
		{TPS: 100.0, LatencyAvg: 10.0, Timestamp: time.Now()},
		{TPS: 200.0, LatencyAvg: 20.0, Timestamp: time.Now()},
		{TPS: 150.0, LatencyAvg: 15.0, Timestamp: time.Now()},
	}

	summary := computeWindowSummary(samples)

	// TPS stats
	if summary.TPS.Avg != 150.0 {
		t.Errorf("TPS.Avg = %v, want 150.0", summary.TPS.Avg)
	}
	if summary.TPS.Min != 100.0 {
		t.Errorf("TPS.Min = %v, want 100.0", summary.TPS.Min)
	}
	if summary.TPS.Max != 200.0 {
		t.Errorf("TPS.Max = %v, want 200.0", summary.TPS.Max)
	}

	// Latency stats
	if summary.Latency.Avg != 15.0 {
		t.Errorf("Latency.Avg = %v, want 15.0", summary.Latency.Avg)
	}
	if summary.Latency.Min != 10.0 {
		t.Errorf("Latency.Min = %v, want 10.0", summary.Latency.Min)
	}
	if summary.Latency.Max != 20.0 {
		t.Errorf("Latency.Max = %v, want 20.0", summary.Latency.Max)
	}

	// CPU fields should be zero-value (no CPU data in samples)
	if summary.CPU.Avg != 0 {
		t.Errorf("CPU.Avg = %v, want 0", summary.CPU.Avg)
	}
	if summary.CPU.Min != 0 {
		t.Errorf("CPU.Min = %v, want 0", summary.CPU.Min)
	}
	if summary.CPU.Max != 0 {
		t.Errorf("CPU.Max = %v, want 0", summary.CPU.Max)
	}

	// SampleCount
	if summary.SampleCount != 3 {
		t.Errorf("SampleCount = %d, want 3", summary.SampleCount)
	}
}

func TestGetDetailedData_PackageLevelFunction(t *testing.T) {
	// BuildPreviewFromBundle is a package-level function (not a method on BundleGenerator).
	// Verify it works without creating a BundleGenerator instance.
	gen := NewBundleGenerator()
	input := &BundleInput{
		Report: &report.Report{
			ID:             "rpt-pkg-fn",
			SuiteID:        "standalone",
			ConnectionName: "test-conn",
			TemplateName:   "oltp",
			DatabaseType:   "mysql",
			StartedAt:      time.Now(),
			DurationMs:     60000,
			TPS:            500.0,
		},
		AdapterType: "sysbench",
		Samples:     generateTestSamples(50, 500.0, 10.0),
	}

	compressed, bundle, err := gen.GenerateAndCompress(input)
	if err != nil {
		t.Fatalf("GenerateAndCompress() error = %v", err)
	}

	// Call as package-level function (not gen.BuildPreviewFromBundle)
	preview := BuildPreviewFromBundle(bundle, compressed, "rpt-pkg-fn")
	if preview == nil {
		t.Fatal("BuildPreviewFromBundle returned nil")
	}
	if preview.BundleFilename != "report_bundle_rpt-pkg-fn.json.gz" {
		t.Errorf("BundleFilename = %q, want %q", preview.BundleFilename, "report_bundle_rpt-pkg-fn.json.gz")
	}
	if preview.CompressedSize != int64(len(compressed)) {
		t.Errorf("CompressedSize = %d, want %d", preview.CompressedSize, len(compressed))
	}
}

// ---------------------------------------------------------------------------
// 2. Bundle volume tests
// ---------------------------------------------------------------------------

func TestBundleGenerator_VolumeVerification(t *testing.T) {
	gen := NewBundleGenerator()

	tests := []struct {
		name        string
		sampleCount int
	}{
		{
			name:        "10min benchmark 600 samples",
			sampleCount: 600,
		},
		{
			name:        "2h benchmark 7200 samples",
			sampleCount: 7200,
		},
		{
			name:        "6h benchmark 21600 samples",
			sampleCount: 21600,
		},
		{
			name:        "high frequency 10000 samples with anomalies",
			sampleCount: 10000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For the anomaly case, inject TPS drops
			var samples []execution.MetricSample
			if tt.name == "high frequency 10000 samples with anomalies" {
				samples = generateTestSamplesWithAnomalies(tt.sampleCount, 1000.0, 10.0,
					tt.sampleCount/5, tt.sampleCount/5+tt.sampleCount/5, 100.0)
			} else {
				samples = generateTestSamples(tt.sampleCount, 1000.0, 10.0)
			}

			input := &BundleInput{
				Report: &report.Report{
					ID:             "rpt-vol",
					SuiteID:        "standalone",
					ConnectionName: "test-conn",
					TemplateName:   "oltp",
					DatabaseType:   "mysql",
					StartedAt:      time.Now(),
					DurationMs:     int64(tt.sampleCount) * 1000,
					TPS:            1000.0,
					TPM:            60000.0,
					LatencyAvgMs:   10.0,
				},
				AdapterType: "sysbench",
				Samples:     samples,
			}

			compressed, bundle, err := gen.GenerateAndCompress(input)
			if err != nil {
				t.Fatalf("GenerateAndCompress() error = %v", err)
			}

			// 1. Size within limit
			if len(compressed) > MaxBundleCompressedBytes {
				t.Errorf("Compressed size %d exceeds limit %d", len(compressed), MaxBundleCompressedBytes)
			}

			// 2. DecompressBundle can restore
			restored, decompressErr := DecompressBundle(compressed)
			if decompressErr != nil {
				t.Fatalf("DecompressBundle() error = %v", decompressErr)
			}
			if restored.RunMeta.RunID != "rpt-vol" {
				t.Errorf("Restored RunID = %q, want %q", restored.RunMeta.RunID, "rpt-vol")
			}

			// 3. AIMeta.CompressedSizeBytes matches
			if bundle.AIMeta.CompressedSizeBytes != int64(len(compressed)) {
				t.Errorf("AIMeta.CompressedSizeBytes = %d, want %d",
					bundle.AIMeta.CompressedSizeBytes, len(compressed))
			}

			// 4. Verify AIMeta.TruncationApplied is consistent
			if bundle.AIMeta.TruncationApplied && bundle.AIMeta.TruncationReason == "" {
				t.Error("TruncationApplied = true but TruncationReason is empty")
			}

			// 5. Anomaly windows preserved when anomalies present
			if tt.name == "high frequency 10000 samples with anomalies" {
				if len(bundle.AnomalyWindows) == 0 {
					t.Error("Expected anomaly windows to be preserved but got none")
				}
			}
		})
	}
}

func TestBundleGenerator_TruncationLayers(t *testing.T) {
	gen := NewBundleGenerator()

	// Force truncation by inflating the bundle with a large Comparison
	// (directly copied into the bundle without capping) and many samples.
	// Each comparison delta has unique metric names and values, so gzip
	// cannot compress them effectively.
	const totalSamples = 10000
	samples := generateTestSamples(totalSamples, 1000.0, 10.0)

	// Build a Comparison with 40000 unique deltas using long metric names
	// that resist gzip compression (~1.1MB compressed in Go). This exceeds
	// the 1MB limit and triggers all 4 truncation layers.
	// Note: because truncation layers cannot reduce Comparison data, the final
	// size may still exceed 1MB. This test verifies truncation IS applied and
	// structural invariants hold.
	const deltaCount = 40000
	deltas := make([]report.BundleComparisonDelta, deltaCount)
	for i := 0; i < deltaCount; i++ {
		deltas[i] = report.BundleComparisonDelta{
			Metric:    fmt.Sprintf("performance_analysis_throughput_measurement_series_comprehensive_database_benchmark_stress_test_run_%06d", i),
			Current:   float64(i) * 1.123456789,
			Previous:  float64(i) * 0.876543210,
			Delta:     float64(i) * 0.246913579,
			PctChange: float64(i) * 0.135792468,
		}
	}

	input := &BundleInput{
		Report: &report.Report{
			ID:             "rpt-trunc",
			SuiteID:        "standalone",
			ConnectionName: "test-conn",
			TemplateName:   "oltp",
			DatabaseType:   "mysql",
			StartedAt:      time.Now(),
			DurationMs:     50000000,
			TPS:            1000.0,
			TPM:            60000.0,
			LatencyAvgMs:   10.0,
		},
		AdapterType: "sysbench",
		Samples:     samples,
		Comparison: &report.BundleComparison{
			PreviousID: "rpt-prev",
			BaselineID: "rpt-base",
			Deltas:     deltas,
		},
	}

	compressed, bundle, err := gen.GenerateAndCompress(input)
	if err != nil {
		t.Fatalf("GenerateAndCompress() error = %v", err)
	}

	// Verify truncation applied
	if !bundle.AIMeta.TruncationApplied {
		t.Error("TruncationApplied = false, expected true for bundle exceeding 1MB")
	}

	// Verify truncation reason is non-empty
	if bundle.AIMeta.TruncationReason == "" {
		t.Error("TruncationReason is empty, expected non-empty")
	}

	// Verify compressed data can still be decompressed
	_, decompressErr := DecompressBundle(compressed)
	if decompressErr != nil {
		t.Errorf("DecompressBundle() after truncation error = %v", decompressErr)
	}

	// RetainedWindows should still be 3 (truncation preserves structure)
	if len(bundle.RetainedWindows) != 3 {
		t.Errorf("RetainedWindows count = %d, want 3", len(bundle.RetainedWindows))
	}

	// AIMeta.CompressedSizeBytes should match actual size
	if bundle.AIMeta.CompressedSizeBytes != int64(len(compressed)) {
		t.Errorf("AIMeta.CompressedSizeBytes = %d, actual = %d",
			bundle.AIMeta.CompressedSizeBytes, len(compressed))
	}
}

func TestBundleGenerator_PreviewMatchesBundle(t *testing.T) {
	gen := NewBundleGenerator()
	samples := generateTestSamplesWithAnomalies(100, 1000.0, 10.0, 40, 55, 100.0)

	input := &BundleInput{
		Report: &report.Report{
			ID:             "rpt-match",
			SuiteID:        "standalone",
			ConnectionName: "test-conn",
			TemplateName:   "oltp",
			DatabaseType:   "mysql",
			StartedAt:      time.Now(),
			DurationMs:     60000,
			TPS:            1000.0,
		},
		AdapterType: "sysbench",
		Samples:     samples,
	}

	compressed, bundle, err := gen.GenerateAndCompress(input)
	if err != nil {
		t.Fatalf("GenerateAndCompress() error = %v", err)
	}

	preview := BuildPreviewFromBundle(bundle, compressed, "rpt-match")
	if preview == nil {
		t.Fatal("BuildPreviewFromBundle returned nil")
	}

	// RetainedWindows count and names match
	if len(preview.RetainedWindows) != len(bundle.RetainedWindows) {
		t.Errorf("Preview RetainedWindows count = %d, bundle count = %d",
			len(preview.RetainedWindows), len(bundle.RetainedWindows))
	}
	for i, pw := range preview.RetainedWindows {
		if pw.Name != bundle.RetainedWindows[i].Name {
			t.Errorf("Preview.RetainedWindows[%d].Name = %q, bundle = %q",
				i, pw.Name, bundle.RetainedWindows[i].Name)
		}
	}

	// AnomalyWindows count matches
	if len(preview.AnomalyWindows) != len(bundle.AnomalyWindows) {
		t.Errorf("Preview AnomalyWindows count = %d, bundle count = %d",
			len(preview.AnomalyWindows), len(bundle.AnomalyWindows))
	}

	// RawSampleCount matches
	if preview.RawSampleCount != len(bundle.RawSamples) {
		t.Errorf("Preview.RawSampleCount = %d, bundle has %d raw samples",
			preview.RawSampleCount, len(bundle.RawSamples))
	}

	// SamplingPolicy matches
	if preview.SamplingPolicy != bundle.AIMeta.SamplingPolicy {
		t.Errorf("Preview.SamplingPolicy = %q, bundle = %q",
			preview.SamplingPolicy, bundle.AIMeta.SamplingPolicy)
	}

	// CompressedSize matches
	if preview.CompressedSize != int64(len(compressed)) {
		t.Errorf("Preview.CompressedSize = %d, want %d",
			preview.CompressedSize, len(compressed))
	}
}

// ---------------------------------------------------------------------------
// 3. Field consistency test
// ---------------------------------------------------------------------------

func TestDetailedDataPreview_FieldConsistency(t *testing.T) {
	gen := NewBundleGenerator()
	samples := generateTestSamplesWithAnomalies(200, 1000.0, 10.0, 80, 100, 80.0)

	monitoring := &report.MonitoringData{
		CPU: &report.MonitoringCPUData{
			UsageAvg: 55.0,
			UsageMax: 88.0,
		},
		Memory: &report.MonitoringMemoryData{
			UsedAvgMB:   4096.0,
			UsedMaxMB:   6144.0,
			UsedPercent: 60.0,
		},
	}

	input := &BundleInput{
		Report: &report.Report{
			ID:             "rpt-consist",
			SuiteID:        "standalone",
			ConnectionName: "test-conn",
			TemplateName:   "oltp",
			DatabaseType:   "mysql",
			StartedAt:      time.Now(),
			DurationMs:     120000,
			TPS:            1000.0,
			TPM:            60000.0,
			LatencyAvgMs:   10.0,
			LatencyP95Ms:   25.0,
			LatencyP99Ms:   50.0,
		},
		AdapterType: "sysbench",
		Samples:     samples,
		Monitoring:  monitoring,
	}

	compressed, bundle, err := gen.GenerateAndCompress(input)
	if err != nil {
		t.Fatalf("GenerateAndCompress() error = %v", err)
	}

	preview := BuildPreviewFromBundle(bundle, compressed, "rpt-consist")
	if preview == nil {
		t.Fatal("BuildPreviewFromBundle returned nil")
	}

	// Verify retained_windows[i].name == bundle.RetainedWindows[i].Name
	for i, pw := range preview.RetainedWindows {
		if pw.Name != bundle.RetainedWindows[i].Name {
			t.Errorf("retained_windows[%d].name: preview=%q, bundle=%q", i, pw.Name, bundle.RetainedWindows[i].Name)
		}
	}

	// Verify retained_windows[i].sample_count == bundle.RetainedWindows[i].Summary.SampleCount
	for i, pw := range preview.RetainedWindows {
		if pw.SampleCount != bundle.RetainedWindows[i].Summary.SampleCount {
			t.Errorf("retained_windows[%d].sample_count: preview=%d, bundle=%d",
				i, pw.SampleCount, bundle.RetainedWindows[i].Summary.SampleCount)
		}
	}

	// Verify anomaly_windows[i].type == bundle.AnomalyWindows[i].Type
	for i, pa := range preview.AnomalyWindows {
		if pa.Type != bundle.AnomalyWindows[i].Type {
			t.Errorf("anomaly_windows[%d].type: preview=%q, bundle=%q", i, pa.Type, bundle.AnomalyWindows[i].Type)
		}
	}

	// Verify anomaly_windows[i].severity == bundle.AnomalyWindows[i].Severity
	for i, pa := range preview.AnomalyWindows {
		if pa.Severity != bundle.AnomalyWindows[i].Severity {
			t.Errorf("anomaly_windows[%d].severity: preview=%q, bundle=%q", i, pa.Severity, bundle.AnomalyWindows[i].Severity)
		}
	}

	// Verify anomaly_windows[i].summary == bundle.AnomalyWindows[i].Summary
	for i, pa := range preview.AnomalyWindows {
		if pa.Summary != bundle.AnomalyWindows[i].Summary {
			t.Errorf("anomaly_windows[%d].summary: preview=%q, bundle=%q", i, pa.Summary, bundle.AnomalyWindows[i].Summary)
		}
	}

	// Verify raw_sample_count == len(bundle.RawSamples)
	if preview.RawSampleCount != len(bundle.RawSamples) {
		t.Errorf("raw_sample_count: preview=%d, bundle=%d", preview.RawSampleCount, len(bundle.RawSamples))
	}

	// Verify sampling_policy == bundle.AIMeta.SamplingPolicy
	if preview.SamplingPolicy != bundle.AIMeta.SamplingPolicy {
		t.Errorf("sampling_policy: preview=%q, bundle=%q", preview.SamplingPolicy, bundle.AIMeta.SamplingPolicy)
	}

	// Verify compressed_size_bytes == len(compressed)
	if preview.CompressedSize != int64(len(compressed)) {
		t.Errorf("compressed_size_bytes: preview=%d, actual=%d", preview.CompressedSize, len(compressed))
	}
}

