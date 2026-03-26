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
