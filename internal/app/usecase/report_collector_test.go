// Package usecase provides report collector tests.
package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

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

func TestReportCollectorFilePersistence(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	collector := NewDefaultReportCollector(WithReportsDir(tmpDir))

	rptCtx := report.ReportContext{
		SuiteID:        "test-suite",
		SourceType:     report.SourceTypeBenchmark,
		ConnectionID:   "conn-1",
		ConnectionName: "MySQL Test",
		DatabaseType:   "mysql",
		TemplateID:     "tmpl-1",
		TemplateName:   "oltp_read_write",
	}

	runFn := func() (*execution.Run, error) {
		return &execution.Run{
			ID:    "run-1",
			State: execution.StateCompleted,
			Result: &execution.BenchmarkResult{
				TPMCalculated: 15000.5,
				TPSCalculated: 250.5,
				LatencyAvg:    15.8,
				LatencyP95:    25.3,
				LatencyP99:    45.2,
				TotalTime:     60.0,
				TotalEvents:   15000,
				ErrorCount:    0,
				ReadQueries:   12000,
				WriteQueries:  3000,
				TimeSeries: []execution.MetricSample{
					{Timestamp: time.Now(), TPS: 248.5, LatencyAvg: 15.5},
					{Timestamp: time.Now().Add(time.Second), TPS: 251.2, LatencyAvg: 16.1},
				},
			},
		}, nil
	}

	result, err := collector.CollectAndPersist(ctx, runFn, rptCtx)
	if err != nil {
		t.Fatalf("CollectAndPersist failed: %v", err)
	}

	// Verify directory structure
	suiteDir := filepath.Join(tmpDir, "test-suite")
	if _, err := os.Stat(suiteDir); os.IsNotExist(err) {
		t.Error("suite directory not created")
	}

	reportDir := filepath.Join(suiteDir, result.ReportID)
	if _, err := os.Stat(reportDir); os.IsNotExist(err) {
		t.Error("report directory not created")
	}

	// Verify metrics.json exists and has correct schema version
	metricsPath := filepath.Join(reportDir, "metrics.json")
	data, err := os.ReadFile(metricsPath)
	if err != nil {
		t.Fatalf("read metrics.json: %v", err)
	}
	var metrics map[string]interface{}
	if err := json.Unmarshal(data, &metrics); err != nil {
		t.Fatalf("parse metrics.json: %v", err)
	}
	if metrics["schema_version"] != "v1" {
		t.Errorf("expected schema_version v1, got %v", metrics["schema_version"])
	}

	// Verify monitoring.json exists
	monitoringPath := filepath.Join(reportDir, "monitoring.json")
	if _, err := os.Stat(monitoringPath); os.IsNotExist(err) {
		t.Error("monitoring.json not created")
	}

	// Verify raw.json exists
	rawPath := filepath.Join(reportDir, "raw.json")
	if _, err := os.Stat(rawPath); os.IsNotExist(err) {
		t.Error("raw.json not created")
	}

	// Verify summary.json exists
	summaryPath := filepath.Join(reportDir, "summary.json")
	if _, err := os.Stat(summaryPath); os.IsNotExist(err) {
		t.Error("summary.json not created")
	}

	// Verify report.html exists
	reportHTMLPath := filepath.Join(reportDir, "report.html")
	if _, err := os.Stat(reportHTMLPath); os.IsNotExist(err) {
		t.Error("report.html not created")
	}
}
