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

	// 5. 验证连接
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return db, nil
}
