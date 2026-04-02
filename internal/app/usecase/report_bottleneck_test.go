package usecase

import (
	"testing"

	"github.com/whhaicheng/DB-BenchMind/internal/domain/report"
)

func TestBottleneckRulesEngine_Judge(t *testing.T) {
	engine := NewBottleneckRulesEngine()

	tests := []struct {
		name               string
		input              *BottleneckInput
		wantBottleneck     report.BottleneckType
		minConfidence      float64
		wantEvidenceCount  int
	}{
		{
			name: "nil input returns unknown",
			input: nil,
			wantBottleneck:    report.BottleneckUnknown,
			minConfidence:     0,
			wantEvidenceCount: 1,
		},
		{
			name: "insufficient samples returns unknown",
			input: &BottleneckInput{
				SampleCount: 2,
			},
			wantBottleneck:    report.BottleneckUnknown,
			minConfidence:     0,
			wantEvidenceCount: 1,
		},
		{
			name: "high CPU peak detects CPU-bound",
			input: &BottleneckInput{
				SampleCount: 100,
				CoreKPIs: report.CoreKPIsSection{
					TPS:          500,
					AvgLatencyMs: 10,
					P95LatencyMs: 25,
				},
				ResourceBottleneck: report.ResourceBottleneckSection{
					CPU: report.ResourceVerdict{
						Avg:     75.0,
						Peak:    95.0,
						Verdict: "critical",
					},
				},
			},
			wantBottleneck:    report.BottleneckCPU,
			minConfidence:     0.5,
			wantEvidenceCount: 2,
		},
		{
			name: "high P99 with moderate CPU detects IO-bound",
			input: &BottleneckInput{
				SampleCount: 100,
				CoreKPIs: report.CoreKPIsSection{
					TPS:          200,
					AvgLatencyMs: 30,
					P95LatencyMs: 100,
					P99LatencyMs: 500,
				},
				ResourceBottleneck: report.ResourceBottleneckSection{
					CPU: report.ResourceVerdict{
						Peak:    50.0,
						Verdict: "normal",
					},
					Disk: report.ResourceVerdict{
						Peak:    90.0,
						Verdict: "critical",
					},
				},
			},
			wantBottleneck:    report.BottleneckIO,
			minConfidence:     0.3,
			wantEvidenceCount: 2,
		},
		{
			name: "high error rate with high P95 and low CPU suggests lock contention",
			input: &BottleneckInput{
				SampleCount: 100,
				ErrorRate:   0.1,
				CoreKPIs: report.CoreKPIsSection{
					P95LatencyMs: 200,
				},
				ResourceBottleneck: report.ResourceBottleneckSection{
					CPU: report.ResourceVerdict{
						Peak:    40.0,
						Verdict: "normal",
					},
				},
			},
			wantBottleneck:    report.BottleneckLockContention,
			minConfidence:     0.3,
			wantEvidenceCount: 2,
		},
		{
			name: "very low TPS with no resource pressure suggests misconfiguration",
			input: &BottleneckInput{
				SampleCount: 100,
				CoreKPIs: report.CoreKPIsSection{
					TPS: 5,
				},
				ResourceBottleneck: report.ResourceBottleneckSection{
					CPU: report.ResourceVerdict{
						Peak:    20.0,
						Verdict: "normal",
					},
					Disk: report.ResourceVerdict{
						Verdict: "normal",
					},
				},
			},
			wantBottleneck:    report.BottleneckMisconfig,
			minConfidence:     0.3,
			wantEvidenceCount: 1,
		},
		{
			name: "no clear bottleneck returns unknown",
			input: &BottleneckInput{
				SampleCount: 100,
				CoreKPIs: report.CoreKPIsSection{
					TPS:          1000,
					AvgLatencyMs: 5,
					P95LatencyMs: 10,
					P99LatencyMs: 15,
				},
				ResourceBottleneck: report.ResourceBottleneckSection{
					CPU: report.ResourceVerdict{
						Peak:    30.0,
						Verdict: "normal",
					},
					Disk: report.ResourceVerdict{
						Verdict: "normal",
					},
				},
			},
			wantBottleneck:    report.BottleneckUnknown,
			minConfidence:     0,
			wantEvidenceCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.Judge(tt.input)
			if result.PrimaryBottleneck != tt.wantBottleneck {
				t.Errorf("PrimaryBottleneck = %v, want %v", result.PrimaryBottleneck, tt.wantBottleneck)
			}
			if result.Confidence < tt.minConfidence {
				t.Errorf("Confidence = %v, want >= %v", result.Confidence, tt.minConfidence)
			}
			if len(result.Evidence) < tt.wantEvidenceCount {
				t.Errorf("Evidence count = %d, want >= %d", len(result.Evidence), tt.wantEvidenceCount)
			}
			if result.PrimaryBottleneck != report.BottleneckUnknown && len(result.Recommendations) == 0 {
				t.Error("Non-unknown bottleneck should have recommendations")
			}
		})
	}
}

func TestDetermineOverallStatus(t *testing.T) {
	tests := []struct {
		name       string
		kpis       report.CoreKPIsSection
		rb         report.ResourceBottleneckSection
		bj         report.BottleneckJudgment
		wantStatus report.OverallStatus
	}{
		{
			name: "healthy when all metrics normal",
			kpis: report.CoreKPIsSection{TPS: 1000, P99LatencyMs: 50},
			rb:   report.ResourceBottleneckSection{},
			bj:   report.BottleneckJudgment{Confidence: 0.2},
			wantStatus: report.StatusHealthy,
		},
		{
			name: "warning when high latency",
			kpis: report.CoreKPIsSection{P99LatencyMs: 300, ErrorRate: 0.02},
			rb:   report.ResourceBottleneckSection{},
			bj:   report.BottleneckJudgment{},
			wantStatus: report.StatusWarning,
		},
		{
			name: "critical when high error rate and CPU",
			kpis: report.CoreKPIsSection{ErrorRate: 0.1, P99LatencyMs: 600},
			rb: report.ResourceBottleneckSection{
				CPU: report.ResourceVerdict{Peak: 95},
			},
			bj: report.BottleneckJudgment{Confidence: 0.8},
			wantStatus: report.StatusCritical,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetermineOverallStatus(tt.kpis, tt.rb, tt.bj)
			if got != tt.wantStatus {
				t.Errorf("DetermineOverallStatus() = %v, want %v", got, tt.wantStatus)
			}
		})
	}
}

func TestCompareReports(t *testing.T) {
	tests := []struct {
		name            string
		current         *report.Report
		previous        *report.Report
		wantNil         bool
		wantDeltas      int
		wantTrend       string
	}{
		{
			name:     "nil previous returns nil",
			current:  &report.Report{TPS: 100},
			previous: nil,
			wantNil:  true,
		},
		{
			name: "improved throughput",
			current:  &report.Report{TPS: 1200, TPM: 72000, LatencyP95Ms: 20, LatencyP99Ms: 40},
			previous: &report.Report{TPS: 1000, TPM: 60000, LatencyP95Ms: 25, LatencyP99Ms: 50},
			wantNil:    false,
			wantDeltas: 4,
			wantTrend:  "improved",
		},
		{
			name: "degraded throughput",
			current:  &report.Report{TPS: 500, TPM: 30000, LatencyP95Ms: 50, LatencyP99Ms: 100},
			previous: &report.Report{TPS: 1000, TPM: 60000, LatencyP95Ms: 20, LatencyP99Ms: 40},
			wantNil:    false,
			wantDeltas: 4,
			wantTrend:  "degraded",
		},
		{
			name: "stable performance",
			current:  &report.Report{TPS: 1005, TPM: 60300, LatencyP95Ms: 25, LatencyP99Ms: 50},
			previous: &report.Report{TPS: 1000, TPM: 60000, LatencyP95Ms: 25, LatencyP99Ms: 50},
			wantNil:    false,
			wantDeltas: 4,
			wantTrend:  "stable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CompareReports(tt.current, tt.previous)
			if tt.wantNil {
				if result != nil {
					t.Error("Expected nil result")
				}
				return
			}
			if result == nil {
				t.Fatal("Expected non-nil result")
			}
			if len(result.Deltas) != tt.wantDeltas {
				t.Errorf("Deltas count = %d, want %d", len(result.Deltas), tt.wantDeltas)
			}
			if result.TrendDirection != tt.wantTrend {
				t.Errorf("TrendDirection = %v, want %v", result.TrendDirection, tt.wantTrend)
			}
		})
	}
}
