// Implements: ReportUsecase tests
// Uses table-driven tests following constitution.md requirements
package usecase

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/whhaicheng/DB-BenchMind/internal/domain/report"
	"github.com/whhaicheng/DB-BenchMind/internal/infra/database"
	_ "modernc.org/sqlite"
)

// setupReportTestDB creates an in-memory SQLite database for testing.
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

// insertTestReport inserts a test report into the database.
func insertTestReport(t *testing.T, db *sql.DB, rpt *report.Report) {
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

// insertTestSuite inserts a test suite into the database.
func insertTestSuite(t *testing.T, db *sql.DB, suite *report.Suite) {
	t.Helper()
	now := time.Now()
	if suite.CreatedAt.IsZero() {
		suite.CreatedAt = now
	}
	if suite.UpdatedAt.IsZero() {
		suite.UpdatedAt = now
	}

	var startedAt, endedAt *string
	if suite.StartedAt != nil {
		s := suite.StartedAt.Format(time.RFC3339)
		startedAt = &s
	}
	if suite.EndedAt != nil {
		s := suite.EndedAt.Format(time.RFC3339)
		endedAt = &s
	}

	query := `
        INSERT INTO suites (
            id, name, execution_mode, failure_policy, cleanup_enabled,
            suite_manifest_json_path, status, started_at, ended_at,
            total_items, completed_items, success_items, failed_items, skipped_items,
            suite_report_json_path, suite_report_html_path, created_at, updated_at
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `

	_, err := db.ExecContext(context.Background(), query,
		suite.ID, suite.Name, suite.ExecutionMode, suite.FailurePolicy, suite.CleanupEnabled,
		suite.SuiteManifestJSONPath, string(suite.Status), startedAt, endedAt,
		suite.TotalItems, suite.CompletedItems, suite.SuccessItems, suite.FailedItems, suite.SkippedItems,
		suite.SuiteReportJSONPath, suite.SuiteReportHTMLPath,
		suite.CreatedAt.Format(time.RFC3339), suite.UpdatedAt.Format(time.RFC3339),
	)
	if err != nil {
		t.Fatalf("insert test suite: %v", err)
	}
}

// =============================================================================
// ListReports Tests
// =============================================================================

func TestReportUsecase_ListReports_Empty(t *testing.T) {
	db := setupReportTestDB(t)
	defer db.Close()

	uc := NewReportUsecase(db)
	ctx := context.Background()

	reports, total, err := uc.ListReports(ctx, ListReportsOptions{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("ListReports() error = %v", err)
	}
	if total != 0 {
		t.Errorf("ListReports() total = %d, want 0", total)
	}
	if len(reports) != 0 {
		t.Errorf("ListReports() returned %d reports, want 0", len(reports))
	}
}

func TestReportUsecase_ListReports_Pagination(t *testing.T) {
	db := setupReportTestDB(t)
	defer db.Close()

	uc := NewReportUsecase(db)
	ctx := context.Background()

	// Insert 15 reports with staggered times for consistent ordering
	for i := 1; i <= 15; i++ {
		rpt := &report.Report{
			ID:           string(rune('a' + i)),
			SuiteID:      report.StandaloneSuiteID,
			SourceType:   report.SourceTypeBenchmark,
			ConnectionID: "conn-1",
			DatabaseType: "mysql",
			Status:       report.StatusCompleted,
			StartedAt:    time.Now().Add(-time.Duration(i) * time.Hour),
		}
		insertTestReport(t, db, rpt)
	}

	tests := []struct {
		name      string
		opts      ListReportsOptions
		wantCount int
		wantTotal int
	}{
		{
			name:      "first page",
			opts:      ListReportsOptions{Page: 1, PageSize: 10},
			wantCount: 10,
			wantTotal: 15,
		},
		{
			name:      "second page",
			opts:      ListReportsOptions{Page: 2, PageSize: 10},
			wantCount: 5,
			wantTotal: 15,
		},
		{
			name:      "third page empty",
			opts:      ListReportsOptions{Page: 3, PageSize: 10},
			wantCount: 0,
			wantTotal: 15,
		},
		{
			name:      "page size 5",
			opts:      ListReportsOptions{Page: 1, PageSize: 5},
			wantCount: 5,
			wantTotal: 15,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reports, total, err := uc.ListReports(ctx, tt.opts)
			if err != nil {
				t.Fatalf("ListReports() error = %v", err)
			}
			if total != tt.wantTotal {
				t.Errorf("ListReports() total = %d, want %d", total, tt.wantTotal)
			}
			if len(reports) != tt.wantCount {
				t.Errorf("ListReports() count = %d, want %d", len(reports), tt.wantCount)
			}
		})
	}
}

func TestReportUsecase_ListReports_Filters(t *testing.T) {
	db := setupReportTestDB(t)
	defer db.Close()

	uc := NewReportUsecase(db)
	ctx := context.Background()

	// Insert reports with different attributes
	reports := []*report.Report{
		{
			ID:           "rpt-1",
			SuiteID:      "suite-1",
			SourceType:   report.SourceTypeAutoBench,
			ConnectionID: "conn-1",
			DatabaseType: "mysql",
			Status:       report.StatusCompleted,
		},
		{
			ID:           "rpt-2",
			SuiteID:      "suite-1",
			SourceType:   report.SourceTypeAutoBench,
			ConnectionID: "conn-2",
			DatabaseType: "postgresql",
			Status:       report.StatusFailed,
		},
		{
			ID:           "rpt-3",
			SuiteID:      report.StandaloneSuiteID,
			SourceType:   report.SourceTypeBenchmark,
			ConnectionID: "conn-1",
			DatabaseType: "mysql",
			Status:       report.StatusCompleted,
		},
	}
	for _, rpt := range reports {
		insertTestReport(t, db, rpt)
	}

	tests := []struct {
		name      string
		opts      ListReportsOptions
		wantIDs   []string
		wantTotal int
	}{
		{
			name:      "filter by suite_id",
			opts:      ListReportsOptions{Page: 1, PageSize: 10, SuiteID: "suite-1"},
			wantIDs:   []string{"rpt-1", "rpt-2"},
			wantTotal: 2,
		},
		{
			name:      "filter by status",
			opts:      ListReportsOptions{Page: 1, PageSize: 10, Status: "failed"},
			wantIDs:   []string{"rpt-2"},
			wantTotal: 1,
		},
		{
			name:      "filter by connection_id",
			opts:      ListReportsOptions{Page: 1, PageSize: 10, ConnectionID: "conn-1"},
			wantIDs:   []string{"rpt-1", "rpt-3"},
			wantTotal: 2,
		},
		{
			name:      "filter by multiple criteria",
			opts:      ListReportsOptions{Page: 1, PageSize: 10, SuiteID: "suite-1", Status: "completed"},
			wantIDs:   []string{"rpt-1"},
			wantTotal: 1,
		},
		{
			name:      "no matching results",
			opts:      ListReportsOptions{Page: 1, PageSize: 10, Status: "cancelled"},
			wantIDs:   []string{},
			wantTotal: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reports, total, err := uc.ListReports(ctx, tt.opts)
			if err != nil {
				t.Fatalf("ListReports() error = %v", err)
			}
			if total != tt.wantTotal {
				t.Errorf("ListReports() total = %d, want %d", total, tt.wantTotal)
			}
			if len(reports) != len(tt.wantIDs) {
				t.Errorf("ListReports() count = %d, want %d", len(reports), len(tt.wantIDs))
				return
			}
			for i, r := range reports {
				if i < len(tt.wantIDs) && r.ID != tt.wantIDs[i] {
					t.Errorf("ListReports()[%d].ID = %s, want %s", i, r.ID, tt.wantIDs[i])
				}
			}
		})
	}
}

// =============================================================================
// GetReport Tests
// =============================================================================

func TestReportUsecase_GetReport_Found(t *testing.T) {
	db := setupReportTestDB(t)
	defer db.Close()

	uc := NewReportUsecase(db)
	ctx := context.Background()

	now := time.Now()
	endedAt := now.Add(5 * time.Minute)
	rpt := &report.Report{
		ID:             "rpt-found",
		SuiteID:        "suite-1",
		SuiteItemID:    "item-1",
		SourceType:     report.SourceTypeAutoBench,
		ConnectionID:   "conn-1",
		ConnectionName: "Test Connection",
		DatabaseType:   "mysql",
		TemplateID:     "tmpl-1",
		TemplateName:   "Test Template",
		StartedAt:      now,
		EndedAt:        &endedAt,
		DurationMs:     300000,
		Status:         report.StatusCompleted,
		TPM:            1000.5,
		TPS:            16.67,
		LatencyAvgMs:   50.2,
		LatencyP95Ms:   120.5,
		LatencyP99Ms:   200.3,
		ErrorCount:     5,
	}
	insertTestReport(t, db, rpt)

	result, err := uc.GetReport(ctx, "rpt-found")
	if err != nil {
		t.Fatalf("GetReport() error = %v", err)
	}

	if result.ID != rpt.ID {
		t.Errorf("GetReport().ID = %s, want %s", result.ID, rpt.ID)
	}
	if result.SuiteID != rpt.SuiteID {
		t.Errorf("GetReport().SuiteID = %s, want %s", result.SuiteID, rpt.SuiteID)
	}
	if result.ConnectionName != rpt.ConnectionName {
		t.Errorf("GetReport().ConnectionName = %s, want %s", result.ConnectionName, rpt.ConnectionName)
	}
	if result.TPM != rpt.TPM {
		t.Errorf("GetReport().TPM = %f, want %f", result.TPM, rpt.TPM)
	}
}

func TestReportUsecase_GetReport_NotFound(t *testing.T) {
	db := setupReportTestDB(t)
	defer db.Close()

	uc := NewReportUsecase(db)
	ctx := context.Background()

	_, err := uc.GetReport(ctx, "nonexistent")
	if err == nil {
		t.Error("GetReport() should return error for nonexistent ID")
	}
}

// =============================================================================
// GetReportMetrics Tests
// =============================================================================

func TestReportUsecase_GetReportMetrics_Success(t *testing.T) {
	db := setupReportTestDB(t)
	defer db.Close()

	uc := NewReportUsecase(db)
	ctx := context.Background()

	// Create temp directory for metrics file
	tmpDir := t.TempDir()
	metricsPath := filepath.Join(tmpDir, "metrics.json")
	metricsData := map[string]interface{}{
		"schema_version": "v1",
		"report_id":      "rpt-metrics",
		"benchmark": map[string]interface{}{
			"database_type": "mysql",
		},
		"execution": map[string]interface{}{
			"status":      "completed",
			"duration_ms": 300000,
		},
		"summary": map[string]interface{}{
			"tpm": 1000.5,
			"tps": 16.67,
		},
	}
	metricsJSON, _ := json.Marshal(metricsData)
	if err := os.WriteFile(metricsPath, metricsJSON, 0644); err != nil {
		t.Fatalf("write metrics file: %v", err)
	}

	rpt := &report.Report{
		ID:              "rpt-metrics",
		SuiteID:         report.StandaloneSuiteID,
		SourceType:      report.SourceTypeBenchmark,
		ConnectionID:    "conn-1",
		DatabaseType:    "mysql",
		Status:          report.StatusCompleted,
		MetricsJSONPath: metricsPath,
	}
	insertTestReport(t, db, rpt)

	result, err := uc.GetReportMetrics(ctx, "rpt-metrics")
	if err != nil {
		t.Fatalf("GetReportMetrics() error = %v", err)
	}

	if result.SchemaVersion != "v1" {
		t.Errorf("GetReportMetrics().SchemaVersion = %s, want v1", result.SchemaVersion)
	}
	if result.ReportID != "rpt-metrics" {
		t.Errorf("GetReportMetrics().ReportID = %s, want rpt-metrics", result.ReportID)
	}
	if result.Benchmark == nil {
		t.Error("GetReportMetrics().Benchmark should not be nil")
	}
}

func TestReportUsecase_GetReportMetrics_NoPath(t *testing.T) {
	db := setupReportTestDB(t)
	defer db.Close()

	uc := NewReportUsecase(db)
	ctx := context.Background()

	rpt := &report.Report{
		ID:              "rpt-no-metrics",
		SuiteID:         report.StandaloneSuiteID,
		SourceType:      report.SourceTypeBenchmark,
		ConnectionID:    "conn-1",
		DatabaseType:    "mysql",
		Status:          report.StatusCompleted,
		MetricsJSONPath: "", // No path
	}
	insertTestReport(t, db, rpt)

	_, err := uc.GetReportMetrics(ctx, "rpt-no-metrics")
	if err == nil {
		t.Error("GetReportMetrics() should return error when no path")
	}
}

func TestReportUsecase_GetReportMetrics_ReportNotFound(t *testing.T) {
	db := setupReportTestDB(t)
	defer db.Close()

	uc := NewReportUsecase(db)
	ctx := context.Background()

	_, err := uc.GetReportMetrics(ctx, "nonexistent")
	if err == nil {
		t.Error("GetReportMetrics() should return error for nonexistent report")
	}
}

// =============================================================================
// ListSuites Tests
// =============================================================================

func TestReportUsecase_ListSuites_Empty(t *testing.T) {
	db := setupReportTestDB(t)
	defer db.Close()

	uc := NewReportUsecase(db)
	ctx := context.Background()

	suites, total, err := uc.ListSuites(ctx, ListSuitesOptions{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("ListSuites() error = %v", err)
	}
	if total != 0 {
		t.Errorf("ListSuites() total = %d, want 0", total)
	}
	if len(suites) != 0 {
		t.Errorf("ListSuites() returned %d suites, want 0", len(suites))
	}
}

func TestReportUsecase_ListSuites_Pagination(t *testing.T) {
	db := setupReportTestDB(t)
	defer db.Close()

	uc := NewReportUsecase(db)
	ctx := context.Background()

	// Insert 8 suites
	for i := 1; i <= 8; i++ {
		suite := &report.Suite{
			ID:         string(rune('A' + i)),
			Name:       string(rune('A' + i)),
			Status:     report.SuiteStatusSuccess,
			TotalItems: 5,
		}
		insertTestSuite(t, db, suite)
	}

	tests := []struct {
		name      string
		opts      ListSuitesOptions
		wantCount int
		wantTotal int
	}{
		{
			name:      "first page",
			opts:      ListSuitesOptions{Page: 1, PageSize: 5},
			wantCount: 5,
			wantTotal: 8,
		},
		{
			name:      "second page",
			opts:      ListSuitesOptions{Page: 2, PageSize: 5},
			wantCount: 3,
			wantTotal: 8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		suites, total, err := uc.ListSuites(ctx, tt.opts)
			if err != nil {
				t.Fatalf("ListSuites() error = %v", err)
			}
			if total != tt.wantTotal {
				t.Errorf("ListSuites() total = %d, want %d", total, tt.wantTotal)
			}
			if len(suites) != tt.wantCount {
				t.Errorf("ListSuites() count = %d, want %d", len(suites), tt.wantCount)
			}
		})
	}
}

func TestReportUsecase_ListSuites_FilterByStatus(t *testing.T) {
	db := setupReportTestDB(t)
	defer db.Close()

	uc := NewReportUsecase(db)
	ctx := context.Background()

	suites := []*report.Suite{
		{ID: "suite-1", Name: "Success Suite", Status: report.SuiteStatusSuccess, TotalItems: 5},
		{ID: "suite-2", Name: "Failed Suite", Status: report.SuiteStatusFailed, TotalItems: 5},
		{ID: "suite-3", Name: "Running Suite", Status: report.SuiteStatusRunning, TotalItems: 5},
	}
	for _, s := range suites {
		insertTestSuite(t, db, s)
	}

	tests := []struct {
		name      string
		opts      ListSuitesOptions
		wantIDs   []string
		wantTotal int
	}{
		{
			name:      "filter by success",
			opts:      ListSuitesOptions{Page: 1, PageSize: 10, Status: "success"},
			wantIDs:   []string{"suite-1"},
			wantTotal: 1,
		},
		{
			name:      "filter by running",
			opts:      ListSuitesOptions{Page: 1, PageSize: 10, Status: "running"},
			wantIDs:   []string{"suite-3"},
			wantTotal: 1,
		},
		{
			name:      "no filter returns all",
			opts:      ListSuitesOptions{Page: 1, PageSize: 10},
			wantIDs:   []string{"suite-1", "suite-2", "suite-3"},
			wantTotal: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		suites, total, err := uc.ListSuites(ctx, tt.opts)
		if err != nil {
			 t.Fatalf("ListSuites() error = %v", err)
            }
            if total != tt.wantTotal {
                t.Errorf("ListSuites() total = %d, want %d", total, tt.wantTotal)
            }
            if len(suites) != len(tt.wantIDs) {
                t.Errorf("ListSuites() count = %d, want %d", len(suites), len(tt.wantIDs))
            }
        })
    }
}

// =============================================================================
// GetSuite Tests
// =============================================================================

func TestReportUsecase_GetSuite_Found(t *testing.T) {
	db := setupReportTestDB(t)
	defer db.Close()

	uc := NewReportUsecase(db)
	ctx := context.Background()

	now := time.Now()
	startedAt := now.Add(-1 * time.Hour)
	 endedAt := now
	suite := &report.Suite{
		ID:             "suite-get",
		Name:           "Test Suite",
		ExecutionMode:  "serial",
		FailurePolicy:  "continue_by_connection",
		CleanupEnabled:   true,
		Status:         report.SuiteStatusSuccess,
		StartedAt:      &startedAt,
		EndedAt:        &endedAt,
		TotalItems:     10,
		CompletedItems: 10,
		SuccessItems:   8,
		FailedItems:    2,
		SkippedItems:   0,
	}
	insertTestSuite(t, db, suite)

	result, err := uc.GetSuite(ctx, "suite-get")
	if err != nil {
		t.Fatalf("GetSuite() error = %v", err)
	}

	if result.ID != suite.ID {
		t.Errorf("GetSuite().ID = %s, want %s", result.ID, suite.ID)
	}
	if result.Name != suite.Name {
		t.Errorf("GetSuite().Name = %s, want %s", result.Name, suite.Name)
	}
	if result.TotalItems != suite.TotalItems {
		t.Errorf("GetSuite().TotalItems = %d, want %d", result.TotalItems, suite.TotalItems)
	}
	if result.SuccessItems != suite.SuccessItems {
		t.Errorf("GetSuite().SuccessItems = %d, want %d", result.SuccessItems, suite.SuccessItems)
	}
}

func TestReportUsecase_GetSuite_NotFound(t *testing.T) {
	db := setupReportTestDB(t)
	defer db.Close()

	uc := NewReportUsecase(db)
	ctx := context.Background()

	_, err := uc.GetSuite(ctx, "nonexistent")
	if err == nil {
		t.Error("GetSuite() should return error for nonexistent ID")
	}
}

// =============================================================================
// Default Values Tests
// =============================================================================

func TestReportUsecase_ListReports_DefaultPageSize(t *testing.T) {
	db := setupReportTestDB(t)
	defer db.Close()

	uc := NewReportUsecase(db)
	ctx := context.Background()

	// Insert 5 reports
	for i := 0; i < 5; i++ {
		rpt := &report.Report{
			ID:           string(rune('a' + i)),
			SuiteID:      report.StandaloneSuiteID,
			SourceType:   report.SourceTypeBenchmark,
			ConnectionID: "conn-1",
			DatabaseType: "mysql",
			Status:       report.StatusCompleted,
		}
		insertTestReport(t, db, rpt)
	}

	// Test with page 0 and pageSize 0 (should use defaults)
	reports, total, err := uc.ListReports(ctx, ListReportsOptions{Page: 0, PageSize: 0})
	if err != nil {
		t.Fatalf("ListReports() error = %v", err)
	}
	if total != 5 {
		t.Errorf("ListReports() total = %d, want 5", total)
	}
	// Default page size should be applied
	if len(reports) > 100 {
		t.Errorf("ListReports() should use default page size, got %d", len(reports))
	 }
}

func TestReportUsecase_ListReports_MaxPageSize(t *testing.T) {
	db := setupReportTestDB(t)
	defer db.Close()

	uc := NewReportUsecase(db)
	ctx := context.Background()

	// Insert 5 reports
	for i := 0; i < 5; i++ {
		rpt := &report.Report{
			ID:           string(rune('a' + i)),
			SuiteID:      report.StandaloneSuiteID,
			SourceType:   report.SourceTypeBenchmark,
			ConnectionID: "conn-1",
			DatabaseType: "mysql",
			Status:       report.StatusCompleted,
		}
		insertTestReport(t, db, rpt)
	}

	// Test with page size > 100 (should be capped)
	reports, total, err := uc.ListReports(ctx, ListReportsOptions{Page: 1, PageSize: 200})
	if err != nil {
		t.Fatalf("ListReports() error = %v", err)
	}
	if total != 5 {
		t.Errorf("ListReports() total = %d, want 5", total)
	}
	// Page size should be capped to 100, but since we only have 5, we should get 5
	if len(reports) != 5 {
		t.Errorf("ListReports() count = %d, want 5", len(reports))
	}
}

// =============================================================================
// DeleteReport Tests
// =============================================================================

func TestReportUsecase_DeleteReport(t *testing.T) {
	db := setupReportTestDB(t)
	defer db.Close()

	uc := NewReportUsecase(db)
	ctx := context.Background()

	insertTestReport(t, db, &report.Report{
		ID:           "rpt-del",
		SuiteID:      report.StandaloneSuiteID,
		SourceType:   report.SourceTypeBenchmark,
		ConnectionID: "conn-1",
		DatabaseType: "mysql",
		Status:       report.StatusCompleted,
	})

	err := uc.DeleteReport(ctx, "rpt-del")
	if err != nil {
		t.Fatalf("DeleteReport() error = %v", err)
	}

	// Verify report is gone
	_, err = uc.GetReport(ctx, "rpt-del")
	if err == nil {
		t.Error("DeleteReport() report should be deleted but still found")
	}
}

func TestReportUsecase_DeleteReport_NotFound(t *testing.T) {
	db := setupReportTestDB(t)
	defer db.Close()

	uc := NewReportUsecase(db)
	ctx := context.Background()

	err := uc.DeleteReport(ctx, "nonexistent")
	if err == nil {
		t.Error("DeleteReport() should return error for nonexistent report")
	}
}

func TestReportUsecase_DeleteReport_CleansUpFiles(t *testing.T) {
	db := setupReportTestDB(t)
	defer db.Close()

	uc := NewReportUsecase(db)
	ctx := context.Background()

	tmpDir := t.TempDir()
	metricsPath := filepath.Join(tmpDir, "metrics.json")
	htmlPath := filepath.Join(tmpDir, "report.html")
	os.WriteFile(metricsPath, []byte(`{"test":true}`), 0644)
	os.WriteFile(htmlPath, []byte(`<html></html>`), 0644)

	insertTestReport(t, db, &report.Report{
		ID:              "rpt-files",
		SuiteID:         report.StandaloneSuiteID,
		SourceType:      report.SourceTypeBenchmark,
		ConnectionID:    "conn-1",
		DatabaseType:    "mysql",
		Status:          report.StatusCompleted,
		MetricsJSONPath: metricsPath,
		ReportHTMLPath:  htmlPath,
	})

	err := uc.DeleteReport(ctx, "rpt-files")
	if err != nil {
		t.Fatalf("DeleteReport() error = %v", err)
	}

	// Verify files were cleaned up
	if _, err := os.Stat(metricsPath); !os.IsNotExist(err) {
		t.Error("DeleteReport() should have removed metrics file")
	}
	if _, err := os.Stat(htmlPath); !os.IsNotExist(err) {
		t.Error("DeleteReport() should have removed HTML file")
	}
}

// =============================================================================
// GetReportBySuiteItemID Tests
// =============================================================================

func TestReportUsecase_GetReportBySuiteItemID_Found(t *testing.T) {
	db := setupReportTestDB(t)
	defer db.Close()

	uc := NewReportUsecase(db)
	ctx := context.Background()

	insertTestReport(t, db, &report.Report{
		ID:           "rpt-suite-item",
		SuiteID:      "suite-1",
		SuiteItemID:  "item-1",
		SourceType:   report.SourceTypeAutoBench,
		ConnectionID: "conn-1",
		DatabaseType: "mysql",
		Status:       report.StatusCompleted,
	})

	tests := []struct {
		name          string
		suiteItemID   string
		wantID        string
		wantErr       bool
	}{
		{
			name:        "found by suite_item_id",
			suiteItemID: "item-1",
			wantID:      "rpt-suite-item",
			wantErr:     false,
		},
		{
			name:        "not found for nonexistent suite_item_id",
			suiteItemID: "item-nonexistent",
			wantID:      "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rpt, err := uc.GetReportBySuiteItemID(ctx, tt.suiteItemID)
			if tt.wantErr {
				if err == nil {
					t.Fatal("GetReportBySuiteItemID() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("GetReportBySuiteItemID() error = %v", err)
			}
			if rpt.ID != tt.wantID {
				t.Errorf("GetReportBySuiteItemID() ID = %q, want %q", rpt.ID, tt.wantID)
			}
		})
	}
}

func TestReportUsecase_GetReportBySuiteItemID_ReturnsCorrectFields(t *testing.T) {
	db := setupReportTestDB(t)
	defer db.Close()

	uc := NewReportUsecase(db)
	ctx := context.Background()

	insertTestReport(t, db, &report.Report{
		ID:           "rpt-fields",
		SuiteID:      "suite-1",
		SuiteItemID:  "item-fields",
		SourceType:   report.SourceTypeAutoBench,
		ConnectionID: "conn-fields",
		DatabaseType: "mysql",
		TemplateID:   "tmpl-1",
		TemplateName: "MySQL OLTP",
		Status:       report.StatusCompleted,
		TPM:          1234.56,
		TPS:          789.01,
	})

	rpt, err := uc.GetReportBySuiteItemID(ctx, "item-fields")
	if err != nil {
		t.Fatalf("GetReportBySuiteItemID() error = %v", err)
	}

	if rpt.SuiteItemID != "item-fields" {
		t.Errorf("SuiteItemID = %q, want %q", rpt.SuiteItemID, "item-fields")
	}
	if rpt.ConnectionID != "conn-fields" {
		t.Errorf("ConnectionID = %q, want %q", rpt.ConnectionID, "conn-fields")
	}
	if rpt.DatabaseType != "mysql" {
		t.Errorf("DatabaseType = %q, want %q", rpt.DatabaseType, "mysql")
	}
	if rpt.TPM != 1234.56 {
		t.Errorf("TPM = %v, want 1234.56", rpt.TPM)
	}
	if rpt.TPS != 789.01 {
		t.Errorf("TPS = %v, want 789.01", rpt.TPS)
	}
}

// =============================================================================
// UpdateReportStatusBySuiteItemID Tests
// =============================================================================

func TestReportUsecase_UpdateReportStatusBySuiteItemID_Failed(t *testing.T) {
	db := setupReportTestDB(t)
	defer db.Close()

	uc := NewReportUsecase(db)
	ctx := context.Background()

	// Insert a running report
	insertTestReport(t, db, &report.Report{
		ID:           "rpt-sync-fail",
		SuiteID:      "suite-sync",
		SuiteItemID:  "item-sync-1",
		SourceType:   report.SourceTypeAutoBench,
		ConnectionID: "conn-1",
		DatabaseType: "mysql",
		Status:       report.StatusRunning,
	})

	err := uc.UpdateReportStatusBySuiteItemID(ctx, "item-sync-1", report.StatusFailed, "connection refused")
	if err != nil {
		t.Fatalf("UpdateReportStatusBySuiteItemID() error = %v", err)
	}

	rpt, err := uc.GetReportBySuiteItemID(ctx, "item-sync-1")
	if err != nil {
		t.Fatalf("GetReportBySuiteItemID() error = %v", err)
	}
	if rpt.Status != report.StatusFailed {
		t.Errorf("Status = %q, want %q", rpt.Status, report.StatusFailed)
	}
	if rpt.ErrorMessage != "connection refused" {
		t.Errorf("ErrorMessage = %q, want %q", rpt.ErrorMessage, "connection refused")
	}
	if rpt.EndedAt == nil {
		t.Error("EndedAt should be set for terminal status")
	}
}

func TestReportUsecase_UpdateReportStatusBySuiteItemID_Completed(t *testing.T) {
	db := setupReportTestDB(t)
	defer db.Close()

	uc := NewReportUsecase(db)
	ctx := context.Background()

	insertTestReport(t, db, &report.Report{
		ID:           "rpt-sync-ok",
		SuiteID:      "suite-sync",
		SuiteItemID:  "item-sync-2",
		SourceType:   report.SourceTypeAutoBench,
		ConnectionID: "conn-1",
		DatabaseType: "mysql",
		Status:       report.StatusRunning,
	})

	err := uc.UpdateReportStatusBySuiteItemID(ctx, "item-sync-2", report.StatusCompleted, "")
	if err != nil {
		t.Fatalf("UpdateReportStatusBySuiteItemID() error = %v", err)
	}

	rpt, err := uc.GetReportBySuiteItemID(ctx, "item-sync-2")
	if err != nil {
		t.Fatalf("GetReportBySuiteItemID() error = %v", err)
	}
	if rpt.Status != report.StatusCompleted {
		t.Errorf("Status = %q, want %q", rpt.Status, report.StatusCompleted)
	}
	if rpt.ErrorMessage != "" {
		t.Errorf("ErrorMessage = %q, want empty", rpt.ErrorMessage)
	}
}

func TestReportUsecase_UpdateReportStatusBySuiteItemID_Nonexistent(t *testing.T) {
	db := setupReportTestDB(t)
	defer db.Close()

	uc := NewReportUsecase(db)
	ctx := context.Background()

	// Should not error on nonexistent suite_item_id (no rows affected)
	err := uc.UpdateReportStatusBySuiteItemID(ctx, "item-nonexistent", report.StatusFailed, "err")
	if err != nil {
		t.Fatalf("UpdateReportStatusBySuiteItemID() error = %v", err)
	}
}
