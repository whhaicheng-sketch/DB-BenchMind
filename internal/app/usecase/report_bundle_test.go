package usecase

import (
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

