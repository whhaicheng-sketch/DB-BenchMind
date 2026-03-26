// Package usecase provides report collector tests.
package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/whhaicheng/DB-BenchMind/internal/domain/execution"
	"github.com/whhaicheng/DB-BenchMind/internal/domain/report"
)

func TestReportCollectorInterface(t *testing.T) {
	// Verifies the interface exists and can be implemented
	var _ ReportCollector = (*DefaultReportCollector)(nil)
}

func TestReportCollectorCollectAndPersist(t *testing.T) {
	tests := []struct {
		name       string
		rptCtx     report.ReportContext
		runFn      func() (*execution.Run, error)
		wantErr    bool
		wantStatus report.ReportStatus
		wantTPM    float64
	}{
		{
			name: "successful collection",
			rptCtx: report.ReportContext{
				SuiteID:        report.StandaloneSuiteID,
				SourceType:     report.SourceTypeBenchmark,
				ConnectionID:   "conn-1",
				ConnectionName: "Test Conn",
				DatabaseType:   "mysql",
			},
			runFn: func() (*execution.Run, error) {
				return &execution.Run{
					ID:    "run-1",
					State: execution.StateCompleted,
					Result: &execution.BenchmarkResult{
						TPMCalculated: 15000.5,
						TPSCalculated: 250.5,
						LatencyAvg:    15.8,
						LatencyP95:    25.3,
					},
				}, nil
			},
			wantErr:    false,
			wantStatus: report.StatusCompleted,
			wantTPM:    15000.5,
		},
		{
			name: "run error",
			rptCtx: report.ReportContext{
				SuiteID:        report.StandaloneSuiteID,
				SourceType:     report.SourceTypeBenchmark,
				ConnectionID:   "conn-1",
				ConnectionName: "Test Conn",
				DatabaseType:   "mysql",
			},
			runFn: func() (*execution.Run, error) {
				return nil, errors.New("benchmark failed")
			},
			wantErr: true,
		},
		{
			name: "failed state",
			rptCtx: report.ReportContext{
				SuiteID:        report.StandaloneSuiteID,
				SourceType:     report.SourceTypeBenchmark,
				ConnectionID:   "conn-1",
				ConnectionName: "Test Conn",
				DatabaseType:   "mysql",
			},
			runFn: func() (*execution.Run, error) {
				return &execution.Run{
					ID:           "run-2",
					State:        execution.StateFailed,
					ErrorMessage: "connection timeout",
				}, nil
			},
			wantErr:    false,
			wantStatus: report.StatusFailed,
		},
		{
			name: "cancelled state",
			rptCtx: report.ReportContext{
				SuiteID:        report.StandaloneSuiteID,
				SourceType:     report.SourceTypeBenchmark,
				ConnectionID:   "conn-1",
				ConnectionName: "Test Conn",
				DatabaseType:   "mysql",
			},
			runFn: func() (*execution.Run, error) {
				return &execution.Run{
					ID:    "run-3",
					State: execution.StateCancelled,
				}, nil
			},
			wantErr:    false,
			wantStatus: report.StatusCancelled,
		},
		{
			name: "force stopped state",
			rptCtx: report.ReportContext{
				SuiteID:        report.StandaloneSuiteID,
				SourceType:     report.SourceTypeBenchmark,
				ConnectionID:   "conn-1",
				ConnectionName: "Test Conn",
				DatabaseType:   "mysql",
			},
			runFn: func() (*execution.Run, error) {
				return &execution.Run{
					ID:    "run-4",
					State: execution.StateForceStopped,
				}, nil
			},
			wantErr:    false,
			wantStatus: report.StatusCancelled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			collector := NewDefaultReportCollector(
				WithReportsDir(t.TempDir()),
			)

			result, err := collector.CollectAndPersist(ctx, tt.runFn, tt.rptCtx)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("CollectAndPersist failed: %v", err)
			}

			if result.ReportID == "" {
				t.Error("expected non-empty ReportID")
			}
			if result.Summary.Status != tt.wantStatus {
				t.Errorf("expected status %v, got %v", tt.wantStatus, result.Summary.Status)
			}
			if tt.wantTPM > 0 && result.Summary.TPM != tt.wantTPM {
				t.Errorf("expected TPM %f, got %f", tt.wantTPM, result.Summary.TPM)
			}
		})
	}
}
