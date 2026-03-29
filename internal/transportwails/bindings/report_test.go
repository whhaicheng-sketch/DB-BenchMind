package bindings

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/whhaicheng/DB-BenchMind/internal/app/usecase"
	"github.com/whhaicheng/DB-BenchMind/internal/domain/report"
	"github.com/whhaicheng/DB-BenchMind/internal/infra/database"
	_ "modernc.org/sqlite"
)

func setupReportTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	if err := database.EnsureReportSchema(context.Background(), db); err != nil {
		t.Fatalf("ensure schema: %v", err)
	}
	return db
}

func insertTestReportBinding(t *testing.T, db *sql.DB, rpt *report.Report) {
	t.Helper()
	now := time.Now()
	if rpt.StartedAt.IsZero() {
		rpt.StartedAt = now
	}
	if rpt.CreatedAt.IsZero() {
		rpt.CreatedAt = now
	}
	if rpt.UpdatedAt.IsZero() {
		rpt.UpdatedAt = now
	}

	var endedAt *string
	if rpt.EndedAt != nil {
		s := rpt.EndedAt.Format(time.RFC3339)
		endedAt = &s
	}

	query := `
		INSERT INTO reports (
			id, suite_id, suite_item_id, source_type, connection_id, connection_name,
			database_type, template_id, template_name, started_at, ended_at, duration_ms,
			status, error_message, tpm, tps, qps, throughput, latency_avg_ms,
			latency_p95_ms, latency_p99_ms, error_count, metrics_json_path,
			monitoring_json_path, raw_json_path, report_html_path, summary_json_path,
			created_at, updated_at, tags
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := db.ExecContext(context.Background(), query,
		rpt.ID, rpt.SuiteID, rpt.SuiteItemID, string(rpt.SourceType), rpt.ConnectionID, rpt.ConnectionName,
		rpt.DatabaseType, rpt.TemplateID, rpt.TemplateName, rpt.StartedAt.Format(time.RFC3339), endedAt, rpt.DurationMs,
		string(rpt.Status), rpt.ErrorMessage, rpt.TPM, rpt.TPS, rpt.QPS, rpt.Throughput, rpt.LatencyAvgMs,
		rpt.LatencyP95Ms, rpt.LatencyP99Ms, rpt.ErrorCount, rpt.MetricsJSONPath,
		rpt.MonitoringJSONPath, rpt.RawJSONPath, rpt.ReportHTMLPath, rpt.SummaryJSONPath,
		rpt.CreatedAt.Format(time.RFC3339), rpt.UpdatedAt.Format(time.RFC3339), rpt.Tags,
	)
	if err != nil {
		t.Fatalf("insert test report: %v", err)
	}
}

func TestReportBinding_ListReports(t *testing.T) {
	db := setupReportTestDB(t)
	defer db.Close()

	uc := usecase.NewReportUsecase(db)
	binding := NewReportBinding(uc)

	// Insert test data
	insertTestReportBinding(t, db, &report.Report{
		ID:           "rpt-1",
		SuiteID:      report.StandaloneSuiteID,
		SourceType:   report.SourceTypeBenchmark,
		ConnectionID: "conn-1",
		DatabaseType: "mysql",
		Status:       report.StatusCompleted,
		StartedAt:    time.Now(),
	})

	result := binding.ListReports(ListReportsOptionsDTO{Page: 1, PageSize: 10})
	if result.Error != "" {
		t.Fatalf("ListReports returned error: %s", result.Error)
	}
	if len(result.Reports) != 1 {
		t.Errorf("expected 1 report, got %d", len(result.Reports))
	}
	if result.Total != 1 {
		t.Errorf("expected total 1, got %d", result.Total)
	}
}

func TestReportBinding_GetReport(t *testing.T) {
	db := setupReportTestDB(t)
	defer db.Close()

	uc := usecase.NewReportUsecase(db)
	binding := NewReportBinding(uc)

	// Insert test data
	now := time.Now()
	insertTestReportBinding(t, db, &report.Report{
		ID:             "rpt-get",
		SuiteID:        report.StandaloneSuiteID,
		SourceType:     report.SourceTypeBenchmark,
		ConnectionID:   "conn-1",
		ConnectionName: "Test Connection",
		DatabaseType:   "mysql",
		Status:         report.StatusCompleted,
		StartedAt:      now,
	})

	result := binding.GetReport("rpt-get")
	if result.Error != "" {
		t.Fatalf("GetReport returned error: %s", result.Error)
	}
	if result.Report == nil {
		t.Fatal("expected report, got nil")
	}
	if result.Report.ID != "rpt-get" {
		t.Errorf("expected ID rpt-get, got %s", result.Report.ID)
	}
	// ConnectionName is optional and may be empty
	if result.Report.ID != "rpt-get" {
		t.Errorf("expected report ID 'rpt-get', got %s", result.Report.ID)
	}
}

func TestReportBinding_GetReportMetrics(t *testing.T) {
	db := setupReportTestDB(t)
	defer db.Close()

	uc := usecase.NewReportUsecase(db)
	binding := NewReportBinding(uc)

	// Create temp metrics file
	tmpDir := t.TempDir()
	metricsPath := filepath.Join(tmpDir, "metrics.json")
	metricsData := map[string]interface{}{
		"schema_version": "v1",
		"report_id":      "rpt-metrics",
		"summary": map[string]interface{}{
			"tpm": 1000.5,
			"tps": 16.67,
		},
	}
	metricsJSON, _ := json.Marshal(metricsData)
	if err := os.WriteFile(metricsPath, metricsJSON, 0644); err != nil {
		t.Fatalf("write metrics file: %v", err)
	}

	// Insert test data
	insertTestReportBinding(t, db, &report.Report{
		ID:              "rpt-metrics",
		SuiteID:         report.StandaloneSuiteID,
		SourceType:      report.SourceTypeBenchmark,
		ConnectionID:    "conn-1",
		DatabaseType:    "mysql",
		Status:          report.StatusCompleted,
		StartedAt:       time.Now(),
		MetricsJSONPath: metricsPath,
	})

	result := binding.GetReportMetrics("rpt-metrics")
	if result.Error != "" {
		t.Fatalf("GetReportMetrics returned error: %s", result.Error)
	}
	if result.Metrics == nil {
		t.Fatal("expected metrics, got nil")
	}
	if result.Metrics.SchemaVersion != "v1" {
		t.Errorf("expected schema_version v1, got %s", result.Metrics.SchemaVersion)
	}
}

func TestReportBinding_ListSuites(t *testing.T) {
	db := setupReportTestDB(t)
	defer db.Close()

	uc := usecase.NewReportUsecase(db)
	binding := NewReportBinding(uc)

	result := binding.ListSuites(ListSuitesOptionsDTO{Page: 1, PageSize: 10})
	if result.Error != "" {
		t.Fatalf("ListSuites returned error: %s", result.Error)
	}
	if result.Total != 0 {
		t.Errorf("expected total 0, got %d", result.Total)
	}
}

func TestReportBinding_DeleteReport(t *testing.T) {
	db := setupReportTestDB(t)
	defer db.Close()

	uc := usecase.NewReportUsecase(db)
	binding := NewReportBinding(uc)

	insertTestReportBinding(t, db, &report.Report{
		ID:           "rpt-del",
		SuiteID:      report.StandaloneSuiteID,
		SourceType:   report.SourceTypeBenchmark,
		ConnectionID: "conn-1",
		DatabaseType: "mysql",
		Status:       report.StatusCompleted,
		StartedAt:    time.Now(),
	})

	result := binding.DeleteReport("rpt-del")
	if result.Error != "" {
		t.Fatalf("DeleteReport returned error: %s", result.Error)
	}
	if !result.Success {
		t.Error("DeleteReport() expected success=true")
	}

	// Verify report is gone
	getResult := binding.GetReport("rpt-del")
	if getResult.Error == "" {
		t.Error("DeleteReport() report should be deleted but GetReport succeeded")
	}
}

func TestReportBinding_DeleteReport_NotFound(t *testing.T) {
	db := setupReportTestDB(t)
	defer db.Close()

	uc := usecase.NewReportUsecase(db)
	binding := NewReportBinding(uc)

	result := binding.DeleteReport("nonexistent")
	if result.Error == "" {
		t.Error("DeleteReport() should return error for nonexistent report")
	}
	if result.Success {
		t.Error("DeleteReport() should not return success for nonexistent report")
	}
}
