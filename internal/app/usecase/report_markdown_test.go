package usecase

import (
	"strings"
	"testing"

	"github.com/whhaicheng/DB-BenchMind/internal/domain/report"
)

func TestMarkdownGenerator_Generate(t *testing.T) {
	gen := NewMarkdownGenerator()

	tests := []struct {
		name               string
		data               *report.HumanReportData
		wantSections       []string
		wantNACount        int
	}{
		{
			name: "full data generates all sections",
			data: &report.HumanReportData{
				RunMeta: report.RunMetaSection{
					RunID:         "run-001",
					BenchmarkType: "sysbench",
					TemplateName:  "oltp_read_write",
					TargetName:    "mysql-prod",
					StartTime:     "2024-01-15T10:00:00Z",
					EndTime:       "2024-01-15T10:10:00Z",
					DurationMs:    600000,
				},
				OverallStatus:     report.StatusHealthy,
				OneLineConclusion:  "Benchmark completed successfully with 1000 TPS.",
				CoreKPIs: report.CoreKPIsSection{
					TPS:           1000.0,
					TPM:           60000.0,
					AvgLatencyMs:  10.0,
					P95LatencyMs:  25.0,
					P99LatencyMs:  50.0,
					ErrorRate:     0.001,
					CPUPeak:       65.0,
					DiskUtilPeak:  45.0,
				},
				ResourceBottleneck: report.ResourceBottleneckSection{
					CPU:    report.ResourceVerdict{Avg: 40.0, Peak: 65.0, Verdict: "normal"},
					Memory: report.ResourceVerdict{AvgMB: 4096, PeakMB: 6144, UsedPct: 60, Verdict: "normal"},
					Disk:   report.ResourceVerdict{Peak: 45000000, Verdict: "normal"},
					Network: report.ResourceVerdict{RxPeak: 100000000, TxPeak: 50000000, Verdict: "normal"},
				},
				BottleneckJudgment: report.BottleneckJudgment{
					PrimaryBottleneck: report.BottleneckUnknown,
					Confidence:        0.3,
					Evidence:          []string{"No strong bottleneck pattern identified"},
					Recommendations:   []string{"Review full metrics"},
				},
				Recommendation: "No specific recommendations. Benchmark results look normal.",
				DetailedDataPreview: &report.DetailedDataPreview{
					BundleFilename: "report_bundle_run-001.json.gz",
					CompressedSize: 51200,
					SamplingPolicy: "front_middle_tail_anomaly",
					RetainedWindows: []report.WindowPreview{
						{Name: "front", SampleCount: 15},
						{Name: "middle", SampleCount: 10},
						{Name: "tail", SampleCount: 15},
					},
					RawSampleCount: 20,
				},
			},
			wantSections: []string{
				"# Benchmark Report",
				"Overall Status",
				"Core KPIs",
				"Resource Bottleneck Summary",
				"Bottleneck Judgment",
				"Recommendation",
				"Detailed Data",
			},
		},
		{
			name: "missing data shows N/A",
			data: &report.HumanReportData{
				RunMeta: report.RunMetaSection{
					RunID: "run-002",
				},
				OverallStatus:     report.StatusWarning,
				OneLineConclusion:  "",
				CoreKPIs:           report.CoreKPIsSection{},
				ResourceBottleneck: report.ResourceBottleneckSection{},
				BottleneckJudgment: report.BottleneckJudgment{
					PrimaryBottleneck: report.BottleneckUnknown,
					Confidence:        0,
					Evidence:          nil,
				},
				Recommendation:        "",
				DetailedDataPreview: nil,
			},
			wantSections: []string{
				"N/A",
				"Overall Status",
				"Bottleneck Judgment",
			},
		},
		{
			name: "comparison section included when present",
			data: &report.HumanReportData{
				RunMeta: report.RunMetaSection{
					RunID: "run-003",
				},
				OverallStatus: report.StatusHealthy,
				CoreKPIs: report.CoreKPIsSection{
					TPS: 1000,
					TPM: 60000,
				},
				Comparison: &report.ComparisonResult{
					PreviousReportID: "run-002",
					Deltas: []report.ComparisonDelta{
						{Metric: "TPS", Current: 1000, Previous: 800, Delta: 200, PctChange: 25},
					},
					TrendDirection: "improved",
				},
				BottleneckJudgment: report.BottleneckJudgment{
					PrimaryBottleneck: report.BottleneckUnknown,
				},
				Recommendation: "Results look normal.",
			},
			wantSections: []string{
				"Comparison",
				"TPS",
				"25.0%",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gen.Generate(tt.data)
			for _, section := range tt.wantSections {
				if !strings.Contains(result, section) {
					t.Errorf("Markdown missing section: %q", section)
				}
			}
		})
	}
}
