package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaFS embed.FS

// InitializeSQLite 初始化 SQLite 数据库
// ctx: 上下文（支持取消）
// dbPath: 数据库文件路径（如 "./data/db-benchmind.db"）
// 返回: 数据库连接对象（单连接池）或错误
func InitializeSQLite(ctx context.Context, dbPath string) (*sql.DB, error) {
	// 1. 创建目录
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("create db directory: %w", err)
	}

	// 2. 连接数据库（启用 WAL 和外键）
	dsn := fmt.Sprintf("file:%s?_journal_mode=WAL&_foreign_keys=on&_cache_size=64000&_synchronous=normal", dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	// 3. 配置单连接池
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	// 3.5 迁移旧表结构（在执行 schema.sql 之前）
	if err := ensureReportColumns(ctx, db); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate report schema: %w", err)
	}

	if err := ensureSuitesColumns(ctx, db); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate suites schema: %w", err)
	}

	// 4. 执行 Schema
	schemaBytes, err := schemaFS.ReadFile("schema.sql")
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("read schema: %w", err)
	}

	_, err = db.ExecContext(ctx, string(schemaBytes))
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("execute schema: %w", err)
	}

	if err := ensureTemplateColumns(ctx, db); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate template schema: %w", err)
	}

	if err := EnsureReportSchema(ctx, db); err != nil {
		db.Close()
		return nil, fmt.Errorf("ensure report schema: %w", err)
	}

	// 5. 验证连接
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return db, nil
}

func ensureTemplateColumns(ctx context.Context, db *sql.DB) error {
	columns := map[string]string{
		"config_json": "ALTER TABLE templates ADD COLUMN config_json TEXT NOT NULL DEFAULT ''",
		"is_builtin":  "ALTER TABLE templates ADD COLUMN is_builtin BOOLEAN NOT NULL DEFAULT 0",
	}

	existing := map[string]bool{}
	rows, err := db.QueryContext(ctx, "PRAGMA table_info(templates)")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name, ctype string
		var notNull int
		var dflt sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &ctype, &notNull, &dflt, &pk); err != nil {
			return err
		}
		existing[name] = true
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for name, ddl := range columns {
		if existing[name] {
			continue
		}
		if _, err := db.ExecContext(ctx, ddl); err != nil {
			return err
		}
	}
	return nil
}

// ensureReportColumns migrates the reports table to add new columns if missing.
// This must run BEFORE schema.sql to prevent index creation failures.
func ensureReportColumns(ctx context.Context, db *sql.DB) error {
	// Check if reports table exists
	var tableExists bool
	row := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='reports'")
	var count int
	if err := row.Scan(&count); err != nil {
		return fmt.Errorf("check reports table existence: %w", err)
	}
	if count == 0 {
		// Table doesn't exist, schema.sql will create it
		return nil
	}
	_ = tableExists

	// Get existing columns
	existing := map[string]bool{}
	rows, err := db.QueryContext(ctx, "PRAGMA table_info(reports)")
	if err != nil {
		return fmt.Errorf("get reports table info: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name, ctype string
		var notNull int
		var dflt sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &ctype, &notNull, &dflt, &pk); err != nil {
			return fmt.Errorf("scan reports column info: %w", err)
		}
		existing[name] = true
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate reports columns: %w", err)
	}

	// Define columns to add (only those that might be missing in old versions)
	columns := []struct {
		name string
		ddl  string
	}{
		{"suite_id", "ALTER TABLE reports ADD COLUMN suite_id TEXT NOT NULL DEFAULT 'standalone'"},
		{"suite_item_id", "ALTER TABLE reports ADD COLUMN suite_item_id TEXT"},
		{"source_type", "ALTER TABLE reports ADD COLUMN source_type TEXT NOT NULL DEFAULT 'benchmark'"},
		{"metrics_json_path", "ALTER TABLE reports ADD COLUMN metrics_json_path TEXT"},
		{"monitoring_json_path", "ALTER TABLE reports ADD COLUMN monitoring_json_path TEXT"},
		{"raw_json_path", "ALTER TABLE reports ADD COLUMN raw_json_path TEXT"},
		{"report_html_path", "ALTER TABLE reports ADD COLUMN report_html_path TEXT"},
		{"summary_json_path", "ALTER TABLE reports ADD COLUMN summary_json_path TEXT"},
	}

	for _, col := range columns {
		if existing[col.name] {
			continue
		}
		if _, err := db.ExecContext(ctx, col.ddl); err != nil {
			return fmt.Errorf("add column %s to reports: %w", col.name, err)
		}
	}

	return nil
}

// ensureSuitesColumns migrates the suites table to add new columns if missing.
// This must run BEFORE schema.sql to prevent index creation failures.
func ensureSuitesColumns(ctx context.Context, db *sql.DB) error {
	// Check if suites table exists
	row := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='suites'")
	var count int
	if err := row.Scan(&count); err != nil {
		return fmt.Errorf("check suites table existence: %w", err)
	}
	if count == 0 {
		// Table doesn't exist, schema.sql will create it
		return nil
	}

	// Get existing columns
	existing := map[string]bool{}
	rows, err := db.QueryContext(ctx, "PRAGMA table_info(suites)")
	if err != nil {
		return fmt.Errorf("get suites table info: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name, ctype string
		var notNull int
		var dflt sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &ctype, &notNull, &dflt, &pk); err != nil {
			return fmt.Errorf("scan suites column info: %w", err)
		}
		existing[name] = true
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate suites columns: %w", err)
	}

	// Define columns to add
	columns := []struct {
		name string
		ddl  string
	}{
		{"suite_manifest_json_path", "ALTER TABLE suites ADD COLUMN suite_manifest_json_path TEXT"},
		{"suite_report_json_path", "ALTER TABLE suites ADD COLUMN suite_report_json_path TEXT"},
		{"suite_report_html_path", "ALTER TABLE suites ADD COLUMN suite_report_html_path TEXT"},
	}

	for _, col := range columns {
		if existing[col.name] {
			continue
		}
		if _, err := db.ExecContext(ctx, col.ddl); err != nil {
			return fmt.Errorf("add column %s to suites: %w", col.name, err)
		}
	}

	return nil
}
