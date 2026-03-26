package database

import (
	"context"
	"database/sql"
	"fmt"
)

// EnsureReportSchema creates reports and suites tables if not exist.
func EnsureReportSchema(ctx context.Context, db *sql.DB) error {
	// Create reports table
	reportsDDL := `
	CREATE TABLE IF NOT EXISTS reports (
		id              TEXT PRIMARY KEY,
		suite_id        TEXT NOT NULL,
		suite_item_id   TEXT,
		source_type     TEXT NOT NULL,
		connection_id   TEXT NOT NULL,
		connection_name TEXT,
		database_type   TEXT NOT NULL,
		template_id     TEXT,
		template_name   TEXT,
		started_at      TEXT NOT NULL,
		ended_at        TEXT,
		duration_ms     INTEGER,
		status          TEXT NOT NULL,
		error_message   TEXT,
		tpm             REAL,
		tps             REAL,
		qps             REAL,
		throughput      REAL,
		latency_avg_ms  REAL,
		latency_p95_ms  REAL,
		latency_p99_ms  REAL,
		error_count     INTEGER,
		metrics_json_path    TEXT,
		monitoring_json_path TEXT,
		raw_json_path        TEXT,
		report_html_path     TEXT,
		summary_json_path    TEXT,
		created_at      TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at      TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		tags            TEXT
	);`
	if _, err := db.ExecContext(ctx, reportsDDL); err != nil {
		return fmt.Errorf("create reports table: %w", err)
	}

	// Create reports indexes
	reportIndexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_reports_suite_id ON reports(suite_id)`,
		`CREATE INDEX IF NOT EXISTS idx_reports_source_type ON reports(source_type)`,
		`CREATE INDEX IF NOT EXISTS idx_reports_connection_id ON reports(connection_id)`,
		`CREATE INDEX IF NOT EXISTS idx_reports_status ON reports(status)`,
		`CREATE INDEX IF NOT EXISTS idx_reports_started_at ON reports(started_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_reports_suite_item_id ON reports(suite_item_id)`,
	}
	for _, idx := range reportIndexes {
		if _, err := db.ExecContext(ctx, idx); err != nil {
			return fmt.Errorf("create reports indexes: %w", err)
		}
	}

	// Create suites table
	suitesDDL := `
	CREATE TABLE IF NOT EXISTS suites (
		id              TEXT PRIMARY KEY,
		name            TEXT,
		execution_mode         TEXT DEFAULT 'serial',
		failure_policy         TEXT DEFAULT 'continue_by_connection',
		cleanup_enabled        INTEGER DEFAULT 1,
		suite_manifest_json_path TEXT,
		status          TEXT NOT NULL,
		started_at      TEXT,
		ended_at        TEXT,
		total_items         INTEGER DEFAULT 0,
		completed_items     INTEGER DEFAULT 0,
		success_items       INTEGER DEFAULT 0,
		failed_items        INTEGER DEFAULT 0,
		skipped_items       INTEGER DEFAULT 0,
		suite_report_json_path  TEXT,
		suite_report_html_path  TEXT,
		created_at      TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at      TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
	);`
	if _, err := db.ExecContext(ctx, suitesDDL); err != nil {
		return fmt.Errorf("create suites table: %w", err)
	}

	// Create suites indexes
	suiteIndexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_suites_status ON suites(status)`,
		`CREATE INDEX IF NOT EXISTS idx_suites_started_at ON suites(started_at DESC)`,
	}
	for _, idx := range suiteIndexes {
		if _, err := db.ExecContext(ctx, idx); err != nil {
			return fmt.Errorf("create suites indexes: %w", err)
		}
	}

	return nil
}
