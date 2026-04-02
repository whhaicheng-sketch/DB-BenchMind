// Package usecase provides report collector tests.
package usecase

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"math"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/whhaicheng/DB-BenchMind/internal/domain/execution"
	"github.com/whhaicheng/DB-BenchMind/internal/domain/report"
	"github.com/whhaicheng/DB-BenchMind/internal/infra/database"
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

// helperOpenTestDB creates an in-memory SQLite with the reports schema.
func helperOpenTestDB(t *testing.T) *sql.DB {
	t.Helper()
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := database.InitializeSQLite(ctx, dbPath)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	return db
}

func TestReportCollectorDBPersistence(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	db := helperOpenTestDB(t)
	defer db.Close()

	collector := NewDefaultReportCollector(
		WithReportsDir(tmpDir),
		WithDB(db),
	)

	rptCtx := report.ReportContext{
		SuiteID:        "suite-db-test",
		SuiteItemID:    "item-1",
		SourceType:     report.SourceTypeAutoBench,
		ConnectionID:   "conn-1",
		ConnectionName: "MySQL Test",
		DatabaseType:   "mysql",
		TemplateID:     "tmpl-1",
		TemplateName:   "oltp_read_write",
	}

	runFn := func() (*execution.Run, error) {
		return &execution.Run{
			ID:    "run-db-1",
			State: execution.StateCompleted,
			Result: &execution.BenchmarkResult{
				TPMCalculated: 12000.0,
				TPSCalculated: 200.0,
				LatencyAvg:    12.5,
				LatencyP95:    22.1,
				LatencyP99:    35.8,
				ErrorCount:    2,
			},
		}, nil
	}

	result, err := collector.CollectAndPersist(ctx, runFn, rptCtx)
	if err != nil {
		t.Fatalf("CollectAndPersist failed: %v", err)
	}

	// Verify the report was inserted into the database
	reportUC := NewReportUsecase(db)
	rpt, err := reportUC.GetReport(ctx, result.ReportID)
	if err != nil {
		t.Fatalf("GetReport failed: %v", err)
	}

	if rpt.ID != result.ReportID {
		t.Errorf("expected report ID %s, got %s", result.ReportID, rpt.ID)
	}
	if rpt.SuiteID != "suite-db-test" {
		t.Errorf("expected suite_id suite-db-test, got %s", rpt.SuiteID)
	}
	if rpt.SuiteItemID != "item-1" {
		t.Errorf("expected suite_item_id item-1, got %s", rpt.SuiteItemID)
	}
	if rpt.SourceType != report.SourceTypeAutoBench {
		t.Errorf("expected source_type autobench, got %s", rpt.SourceType)
	}
	if rpt.ConnectionID != "conn-1" {
		t.Errorf("expected connection_id conn-1, got %s", rpt.ConnectionID)
	}
	if rpt.ConnectionName != "MySQL Test" {
		t.Errorf("expected connection_name MySQL Test, got %s", rpt.ConnectionName)
	}
	if rpt.DatabaseType != "mysql" {
		t.Errorf("expected database_type mysql, got %s", rpt.DatabaseType)
	}
	if rpt.Status != report.StatusCompleted {
		t.Errorf("expected status completed, got %s", rpt.Status)
	}
	if rpt.TPM != 12000.0 {
		t.Errorf("expected TPM 12000.0, got %f", rpt.TPM)
	}
	if rpt.TPS != 200.0 {
		t.Errorf("expected TPS 200.0, got %f", rpt.TPS)
	}
	if rpt.LatencyAvgMs != 12.5 {
		t.Errorf("expected latency_avg_ms 12.5, got %f", rpt.LatencyAvgMs)
	}
	if rpt.ErrorCount != 2 {
		t.Errorf("expected error_count 2, got %d", rpt.ErrorCount)
	}

	// Verify file paths were persisted
	if rpt.MetricsJSONPath == "" {
		t.Error("expected non-empty metrics_json_path")
	}
	if rpt.MonitoringJSONPath == "" {
		t.Error("expected non-empty monitoring_json_path")
	}
	if rpt.RawJSONPath == "" {
		t.Error("expected non-empty raw_json_path")
	}
	if rpt.ReportHTMLPath == "" {
		t.Error("expected non-empty report_html_path")
	}
	if rpt.SummaryJSONPath == "" {
		t.Error("expected non-empty summary_json_path")
	}

	// Verify ListReports returns the report
	reports, total, err := reportUC.ListReports(ctx, ListReportsOptions{})
	if err != nil {
		t.Fatalf("ListReports failed: %v", err)
	}
	if total != 1 {
		t.Errorf("expected 1 report, got %d", total)
	}
	if len(reports) != 1 {
		t.Fatalf("expected 1 report in slice, got %d", len(reports))
	}
	if reports[0].ID != result.ReportID {
		t.Errorf("expected report ID %s, got %s", result.ReportID, reports[0].ID)
	}
}

func TestReportCollectorDBPersistence_FailedRun(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	db := helperOpenTestDB(t)
	defer db.Close()

	collector := NewDefaultReportCollector(
		WithReportsDir(tmpDir),
		WithDB(db),
	)

	rptCtx := report.ReportContext{
		SuiteID:        "suite-fail-test",
		SourceType:     report.SourceTypeBenchmark,
		ConnectionID:   "conn-2",
		ConnectionName: "PG Test",
		DatabaseType:   "postgresql",
	}

	runFn := func() (*execution.Run, error) {
		return &execution.Run{
			ID:           "run-fail-1",
			State:        execution.StateFailed,
			ErrorMessage: "connection refused",
		}, nil
	}

	result, err := collector.CollectAndPersist(ctx, runFn, rptCtx)
	if err != nil {
		t.Fatalf("CollectAndPersist failed: %v", err)
	}

	reportUC := NewReportUsecase(db)
	rpt, err := reportUC.GetReport(ctx, result.ReportID)
	if err != nil {
		t.Fatalf("GetReport failed: %v", err)
	}

	if rpt.Status != report.StatusFailed {
		t.Errorf("expected status failed, got %s", rpt.Status)
	}
	if rpt.ErrorMessage != "connection refused" {
		t.Errorf("expected error_message 'connection refused', got %s", rpt.ErrorMessage)
	}
}

func TestReportCollectorWithoutDB(t *testing.T) {
	// Verify collector works without DB (no panic, no error)
	ctx := context.Background()
	tmpDir := t.TempDir()

	collector := NewDefaultReportCollector(WithReportsDir(tmpDir))

	rptCtx := report.ReportContext{
		SuiteID:        "standalone",
		SourceType:     report.SourceTypeBenchmark,
		ConnectionID:   "conn-1",
		ConnectionName: "Test",
		DatabaseType:   "mysql",
	}

	runFn := func() (*execution.Run, error) {
		return &execution.Run{
			ID:    "run-no-db",
			State: execution.StateCompleted,
			Result: &execution.BenchmarkResult{
				TPMCalculated: 5000.0,
			},
		}, nil
	}

	result, err := collector.CollectAndPersist(ctx, runFn, rptCtx)
	if err != nil {
		t.Fatalf("CollectAndPersist failed: %v", err)
	}
	if result.ReportID == "" {
		t.Error("expected non-empty ReportID")
	}
	if result.PersistError != nil {
		t.Errorf("expected no PersistError without DB, got: %v", result.PersistError)
	}
}

func TestComputeStatsFromSamples(t *testing.T) {
	tests := []struct {
		name                string
		samples             []execution.MetricSample
		wantValid           bool
		wantTPSAvg          float64
		wantTPSMax          float64
		wantTPSMin          float64
		wantLatP50LessP90   bool
		wantLatP90LessP95   bool
		wantLatP95LessP99   bool
	}{
		{
			name:      "empty samples returns invalid",
			samples:   []execution.MetricSample{},
			wantValid: false,
		},
		{
			name: "single sample",
			samples: []execution.MetricSample{
				{TPS: 100, LatencyAvg: 10},
			},
			wantValid:  true,
			wantTPSAvg: 100,
			wantTPSMax: 100,
			wantTPSMin: 100,
		},
		{
			name: "multi-sample with distinct values",
			samples: []execution.MetricSample{
				{TPS: 100, TPM: 6000, LatencyAvg: 5.0},
				{TPS: 200, TPM: 12000, LatencyAvg: 10.0},
				{TPS: 300, TPM: 18000, LatencyAvg: 15.0},
				{TPS: 400, TPM: 24000, LatencyAvg: 20.0},
				{TPS: 500, TPM: 30000, LatencyAvg: 25.0},
			},
			wantValid:         true,
			wantTPSAvg:        300,
			wantTPSMax:        500,
			wantTPSMin:        100,
			wantLatP50LessP90: true,
			wantLatP90LessP95: true,
			wantLatP95LessP99: true,
		},
		{
			name: "samples with zero latency are filtered from latency stats",
			samples: []execution.MetricSample{
				{TPS: 100, LatencyAvg: 0},
				{TPS: 200, LatencyAvg: 10.0},
				{TPS: 300, LatencyAvg: 20.0},
			},
			wantValid:  true,
			wantTPSAvg: 200,
			wantTPSMax: 300,
			wantTPSMin: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := computeStatsFromSamples(tt.samples)
			if stats.Valid != tt.wantValid {
				t.Errorf("Valid = %v, want %v", stats.Valid, tt.wantValid)
			}
			if !stats.Valid {
				return
			}
			if stats.TPSAvg != tt.wantTPSAvg {
				t.Errorf("TPSAvg = %v, want %v", stats.TPSAvg, tt.wantTPSAvg)
			}
			if stats.TPSMax != tt.wantTPSMax {
				t.Errorf("TPSMax = %v, want %v", stats.TPSMax, tt.wantTPSMax)
			}
			if stats.TPSMin != tt.wantTPSMin {
				t.Errorf("TPSMin = %v, want %v", stats.TPSMin, tt.wantTPSMin)
			}
			if tt.wantLatP50LessP90 && !(stats.LatP50 <= stats.LatP90) {
				t.Errorf("P50 (%v) should be <= P90 (%v)", stats.LatP50, stats.LatP90)
			}
			if tt.wantLatP90LessP95 && !(stats.LatP90 <= stats.LatP95) {
				t.Errorf("P90 (%v) should be <= P95 (%v)", stats.LatP90, stats.LatP95)
			}
			if tt.wantLatP95LessP99 && !(stats.LatP95 <= stats.LatP99) {
				t.Errorf("P95 (%v) should be <= P99 (%v)", stats.LatP95, stats.LatP99)
			}
		})
	}
}

func TestPercentile(t *testing.T) {
	tests := []struct {
		name   string
		sorted []float64
		p      float64
		want   float64
	}{
		{"empty", []float64{}, 50, 0},
		{"single element", []float64{10}, 50, 10},
		{"two elements p50", []float64{10, 20}, 50, 15},
		{"two elements p0", []float64{10, 20}, 0, 10},
		{"two elements p100", []float64{10, 20}, 100, 20},
		{"five elements p50", []float64{1, 2, 3, 4, 5}, 50, 3},
		{"five elements p90", []float64{1, 2, 3, 4, 5}, 90, 4.6},
		{"five elements p99", []float64{1, 2, 3, 4, 5}, 99, 4.96},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := percentile(tt.sorted, tt.p)
			if math.Abs(got-tt.want) > 0.001 {
				t.Errorf("percentile(%v, %v) = %v, want %v", tt.sorted, tt.p, got, tt.want)
			}
		})
	}
}

func TestReportCollectorWithSamples(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	collector := NewDefaultReportCollector(WithReportsDir(tmpDir))

	samples := []execution.MetricSample{
		{Timestamp: time.Now().Add(-5 * time.Second), TPS: 100, LatencyAvg: 5.0},
		{Timestamp: time.Now().Add(-4 * time.Second), TPS: 200, LatencyAvg: 10.0},
		{Timestamp: time.Now().Add(-3 * time.Second), TPS: 300, LatencyAvg: 15.0},
		{Timestamp: time.Now().Add(-2 * time.Second), TPS: 250, LatencyAvg: 12.5},
		{Timestamp: time.Now().Add(-1 * time.Second), TPS: 280, LatencyAvg: 13.0},
	}

	rptCtx := report.ReportContext{
		SuiteID:        "standalone",
		SourceType:     report.SourceTypeBenchmark,
		ConnectionID:   "conn-1",
		ConnectionName: "Test",
		DatabaseType:   "mysql",
	}

	runFn := func() (*execution.Run, error) {
		return &execution.Run{
			ID:    "run-samples-1",
			State: execution.StateCompleted,
			Result: &execution.BenchmarkResult{
				TPSCalculated: 280,
				TPMCalculated: 16800,
				LatencyAvg:    13.0,
				LatencyP95:    15.0,
				LatencyP99:    15.0,
			},
		}, nil
	}

	result, err := collector.CollectAndPersist(ctx, runFn, rptCtx,
		WithSamples(samples),
		WithAdapterType("sysbench"),
	)
	if err != nil {
		t.Fatalf("CollectAndPersist failed: %v", err)
	}

	// Verify metrics.json has time_series
	reportDir := filepath.Join(tmpDir, "standalone", result.ReportID)
	metricsData, err := os.ReadFile(filepath.Join(reportDir, "metrics.json"))
	if err != nil {
		t.Fatalf("read metrics.json: %v", err)
	}
	var metrics map[string]interface{}
	if err := json.Unmarshal(metricsData, &metrics); err != nil {
		t.Fatalf("parse metrics.json: %v", err)
	}

	// Check time_series exists
	ts, ok := metrics["time_series"].([]interface{})
	if !ok {
		t.Fatal("expected time_series to be an array")
	}
	if len(ts) != 5 {
		t.Errorf("expected 5 time series points, got %d", len(ts))
	}

	// Check percentiles are real (not fake approximations)
	percentiles, ok := metrics["percentiles"].(map[string]interface{})
	if !ok {
		t.Fatal("expected percentiles to be a map")
	}
	// With 5 latency samples (5, 10, 15, 12.5, 13.0), sorted: [5, 10, 12.5, 13, 15]
	// P50 should be ~12.5, not 13 (the LatencyAvg from tool)
	p50, _ := percentiles["p50"].(float64)
	if p50 == 13.0 {
		t.Errorf("p50 should be computed from samples, not equal to LatencyAvg (13.0), got %v", p50)
	}

	// Check raw.json has correct adapter_type
	rawData, err := os.ReadFile(filepath.Join(reportDir, "raw.json"))
	if err != nil {
		t.Fatalf("read raw.json: %v", err)
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(rawData, &raw); err != nil {
		t.Fatalf("parse raw.json: %v", err)
	}
	if raw["adapter_type"] != "sysbench" {
		t.Errorf("expected adapter_type sysbench, got %v", raw["adapter_type"])
	}
}
