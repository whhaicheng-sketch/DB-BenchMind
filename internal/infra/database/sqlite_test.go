package database

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite" // 纯 Go SQLite 链动
)

// Test 1: 测试数据库初始化成功
func TestInitializeSQLite(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Act
	db, err := InitializeSQLite(context.Background(), dbPath)

	// Assert
	if err != nil {
		t.Fatalf("InitializeSQLite failed: %v", err)
	}
	if db == nil {
		t.Fatal("Expected non-nil database")
	}
	defer db.Close()

	// Verify tables exist
	tables := []string{
		"connections", "templates", "tasks", "runs",
		"metric_samples", "run_logs", "reports", "settings",
	}
	for _, table := range tables {
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&count)
		if err != nil {
			t.Fatalf("Failed to check table %s: %v", table, err)
		}
		if count != 1 {
			t.Errorf("Table %s not found", table)
		}
	}
}

// Test 2: 测试 WAL 模式启用
func TestInitializeSQLite_WALMode(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := InitializeSQLite(context.Background(), dbPath)
	if err != nil {
		t.Fatalf("InitializeSQLite failed: %v", err)
	}
	defer db.Close()

	var journalMode string
	err = db.QueryRow("PRAGMA journal_mode").Scan(&journalMode)
	if err != nil {
		t.Fatalf("Failed to query journal_mode: %v", err)
	}
	if journalMode != "wal" {
		t.Errorf("Expected journal_mode='wal', got '%s'", journalMode)
	}
}
// Test 3: 测试外键约束启用
func TestInitializeSQLite_ForeignKeyEnabled(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := InitializeSQLite(context.Background(), dbPath)
	if err != nil {
		t.Fatalf("InitializeSQLite failed: %v", err)
	}
	defer db.Close()

	var foreignKeys int
	err = db.QueryRow("PRAGMA foreign_keys").Scan(&foreignKeys)
	if err != nil {
		t.Fatalf("Failed to query foreign_keys: %v", err)
	}
	if foreignKeys != 1 {
		t.Errorf("Expected foreign_keys=1, got %d", foreignKeys)
	}
}
// Test 4: 测试数据库 schema 初始化时不再插入历史模板数据
func TestInitializeSQLite_DoesNotSeedTemplatesInSchema(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := InitializeSQLite(context.Background(), dbPath)
	if err != nil {
		t.Fatalf("InitializeSQLite failed: %v", err)
	}
	defer db.Close()
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM templates").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count templates: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 templates after raw schema init, got %d", count)
	}
}
// Test 5: 测试单连接池配置
func TestInitializeSQLite_SingleConnection(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := InitializeSQLite(context.Background(), dbPath)
	if err != nil {
		t.Fatalf("InitializeSQLite failed: %v", err)
	}
	defer db.Close()
	stats := db.Stats()
	if stats.OpenConnections != 1 {
		t.Errorf("Expected 1 open connection, got %d", stats.OpenConnections)
	}
}
// Test 6: 测试数据库已存在时重新打开
func TestInitializeSQLite_ReopenExisting(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	// First initialization
	db1, err := InitializeSQLite(context.Background(), dbPath)
	if err != nil {
		t.Fatalf("First InitializeSQLite failed: %v", err)
	}
	db1.Close()
	// Second initialization (reopen)
	db2, err := InitializeSQLite(context.Background(), dbPath)
	if err != nil {
		t.Fatalf("Second InitializeSQLite failed: %v", err)
	}
	defer db2.Close()
	// Verify schema init remains idempotent and does not inject legacy templates
	var count int
	err = db2.QueryRow("SELECT COUNT(*) FROM templates").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query templates: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 templates after reopen, got %d", count)
	}
}
// Test 7: 测试报告表模式初始化
func TestInitializeSQLiteWithReportSchema(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := InitializeSQLite(ctx, dbPath)
	if err != nil {
		t.Fatalf("InitializeSQLite failed: %v", err)
	}
	defer db.Close()
	// Verify reports table exists
	var tableName string
	err = db.QueryRowContext(ctx,
		"SELECT name FROM sqlite_master WHERE type='table' AND name='reports'").Scan(&tableName)
	if err != nil {
		t.Fatalf("reports table not found: %v", err)
	}
	// Verify suites table exists
	err = db.QueryRowContext(ctx,
		"SELECT name FROM sqlite_master WHERE type='table' AND name='suites'").Scan(&tableName)
	if err != nil {
		t.Fatalf("suites table not found: %v", err)
	}
}
// Test 8: 测试旧 reports 表迁移(添加缺失的 suite_id 列)
func TestEnsureReportColumns_MigrateOldTable(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	// 1. 直接创建旧版 reports 表（没有 suite_id 列)
	dsn := fmt.Sprintf("file:%s?_journal_mode=WAL&_foreign_keys=on", dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer db.Close()
	oldSchema := `
CREATE TABLE reports (
    id TEXT PRIMARY KEY,
    connection_id TEXT NOT NULL,
    database_type TEXT NOT NULL,
    started_at TEXT NOT NULL,
    status TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
)
	`
	_, err = db.ExecContext(ctx, oldSchema)
	if err != nil {
		t.Fatalf("create old reports table: %v", err)
	}
	// 2. 运行迁移
	if err := ensureReportColumns(ctx, db); err != nil {
		t.Fatalf("ensureReportColumns failed: %v", err)
	}
	// 3. 验证 suite_id 列已添加
	var count int
	err = db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM pragma_table_info(reports) WHERE name='suite_id'").Scan(&count)
	if err != nil {
		t.Fatalf("check suite_id column: %v", err)
	}
	if count == 0 {
		t.Error("reports table should have suite_id column after migration")
	}
	// 4. 鷻加的列应该有默认值
	var defaultVal string
	err = db.QueryRowContext(ctx,
		"SELECT suite_id FROM reports LIMIT 1").Scan(&defaultVal)
	if err != nil {
		t.Fatalf("query suite_id default: %v", err)
	}
	if defaultVal != "standalone" {
		t.Errorf("expected suite_id default='standalone', got '%s'", defaultVal)
	}
}
// Test 9: 测试旧 suites 表迁移
func TestEnsureSuitesColumns_MigrateOldTable(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	dsn := fmt.Sprintf("file:%s?_journal_mode=WAL&_foreign_keys=on", dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer db.Close()
	oldSchema := `
CREATE TABLE suites (
    id TEXT PRIMARY KEY,
    name TEXT,
    status TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
)
	`
	_, err = db.ExecContext(ctx, oldSchema)
	if err != nil {
		t.Fatalf("create old suites table: %v", err)
	}
	// 2. 运行迁移
	if err := ensureSuitesColumns(ctx, db); err != nil {
		t.Fatalf("ensureSuitesColumns failed: %v", err)
	}
	// 3. 验证 suite_manifest_json_path 列已添加
	var count int
	err = db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM pragma_table_info(suites) WHERE name='suite_manifest_json_path'").Scan(&count)
	if err != nil {
		t.Fatalf("check suite_manifest_json_path column: %v", err)
	}
	if count == 0 {
		t.Error("suites table should have suite_manifest_json_path column after migration")
	 }
	db.Close()
}
