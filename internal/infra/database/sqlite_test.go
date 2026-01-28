package database

import (
	"context"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite" // 纯 Go SQLite 驱动
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

// Test 4: 测试内置模板已插入
func TestInitializeSQLite_BuiltinTemplates(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := InitializeSQLite(context.Background(), dbPath)
	if err != nil {
		t.Fatalf("InitializeSQLite failed: %v", err)
	}
	defer db.Close()

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM templates WHERE is_builtin=1").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count builtin templates: %v", err)
	}
	if count != 7 {
		t.Errorf("Expected 7 builtin templates, got %d", count)
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

	// Verify data persistence
	var count int
	err = db2.QueryRow("SELECT COUNT(*) FROM templates").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query templates: %v", err)
	}
	if count != 7 {
		t.Errorf("Expected 7 templates after reopen, got %d", count)
	}
}
