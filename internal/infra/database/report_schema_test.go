package database

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func TestEnsureReportSchema(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB(t)
	defer db.Close()

	if err := EnsureReportSchema(ctx, db); err != nil {
		t.Fatalf("EnsureReportSchema failed: %v", err)
	}

	// Verify reports table exists
	var tableName string
	err := db.QueryRowContext(ctx,
		"SELECT name FROM sqlite_master WHERE type='table' AND name='reports'").Scan(&tableName)
	if err != nil {
		t.Fatalf("reports table not found: %v", err)
	}
	if tableName != "reports" {
		t.Errorf("expected table name 'reports', got %s", tableName)
	}

	// Verify suites table exists
	err = db.QueryRowContext(ctx,
		"SELECT name FROM sqlite_master WHERE type='table' AND name='suites'").Scan(&tableName)
	if err != nil {
		t.Fatalf("suites table not found: %v", err)
	}
	if tableName != "suites" {
		t.Errorf("expected table name 'suites', got %s", tableName)
	}
}

func TestEnsureReportSchemaIdempotent(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB(t)
	defer db.Close()

	// Run twice to verify idempotency
	if err := EnsureReportSchema(ctx, db); err != nil {
		t.Fatalf("first EnsureReportSchema failed: %v", err)
	}
	if err := EnsureReportSchema(ctx, db); err != nil {
		t.Fatalf("second EnsureReportSchema failed: %v", err)
	}
}

func TestReportsTableIndexes(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB(t)
	defer db.Close()

	if err := EnsureReportSchema(ctx, db); err != nil {
		t.Fatalf("EnsureReportSchema failed: %v", err)
	}

	expectedIndexes := []string{
		"idx_reports_suite_id",
		"idx_reports_source_type",
		"idx_reports_connection_id",
		"idx_reports_status",
		"idx_reports_started_at",
		"idx_reports_suite_item_id",
	}

	for _, idx := range expectedIndexes {
		var name string
		err := db.QueryRowContext(ctx,
			"SELECT name FROM sqlite_master WHERE type='index' AND name=?", idx).Scan(&name)
		if err != nil {
			t.Errorf("index %s not found: %v", idx, err)
		}
	}
}

func TestSuitesTableIndexes(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB(t)
	defer db.Close()

	if err := EnsureReportSchema(ctx, db); err != nil {
		t.Fatalf("EnsureReportSchema failed: %v", err)
	}

	expectedIndexes := []string{
		"idx_suites_status",
		"idx_suites_started_at",
	}

	for _, idx := range expectedIndexes {
		var name string
		err := db.QueryRowContext(ctx,
			"SELECT name FROM sqlite_master WHERE type='index' AND name=?", idx).Scan(&name)
		if err != nil {
			t.Errorf("index %s not found: %v", idx, err)
		}
	}
}

func TestReportsTableColumns(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB(t)
	defer db.Close()

	if err := EnsureReportSchema(ctx, db); err != nil {
		t.Fatalf("EnsureReportSchema failed: %v", err)
	}

	expectedColumns := []string{
		"id", "suite_id", "suite_item_id", "source_type",
		"connection_id", "connection_name", "database_type",
		"template_id", "template_name",
		"started_at", "ended_at", "duration_ms",
		"status", "error_message",
		"tpm", "tps", "qps", "throughput",
		"latency_avg_ms", "latency_p95_ms", "latency_p99_ms", "error_count",
		"metrics_json_path", "monitoring_json_path", "raw_json_path",
		"report_html_path", "summary_json_path",
		"created_at", "updated_at", "tags",
	}

	rows, err := db.QueryContext(ctx, "PRAGMA table_info(reports)")
	if err != nil {
		t.Fatalf("query table info: %v", err)
	}
	defer rows.Close()

	existingColumns := make(map[string]bool)
	for rows.Next() {
		var cid int
		var name, ctype string
		var notNull int
		var dflt sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &ctype, &notNull, &dflt, &pk); err != nil {
			t.Fatalf("scan row: %v", err)
		}
		existingColumns[name] = true
	}

	for _, col := range expectedColumns {
		if !existingColumns[col] {
			t.Errorf("missing column: %s", col)
		}
	}
}

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	dsn := fmt.Sprintf("file:%s?_journal_mode=WAL&_foreign_keys=on", dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	return db
}
