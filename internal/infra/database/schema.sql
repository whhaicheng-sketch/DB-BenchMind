-- DB-BenchMind Database Schema
-- SQLite Database (modernc.org/sqlite - no CGO)
-- WAL mode enabled for better concurrency

-- Enable WAL mode and optimize for performance
PRAGMA journal_mode = WAL;
PRAGMA synchronous = NORMAL;
PRAGMA foreign_keys = ON;
PRAGMA cache_size = -64000;  -- 64MB cache
PRAGMA temp_store = MEMORY;
PRAGMA mmap_size = 30000000000;

-- =============================================================================
-- Table 1: connections
-- 数据库连接配置表
-- =============================================================================
CREATE TABLE IF NOT EXISTS connections (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    db_type TEXT NOT NULL,  -- 'mysql', 'oracle', 'sqlserver', 'postgresql'
    config_json TEXT NOT NULL,  -- 连接配置（JSON 格式，不含密码）
    created_at TEXT NOT NULL,  -- ISO 8601 format
    updated_at TEXT NOT NULL   -- ISO 8601 format
);

-- Index for connections
CREATE INDEX IF NOT EXISTS idx_connections_db_type ON connections(db_type);
CREATE INDEX IF NOT EXISTS idx_connections_created_at ON connections(created_at);

-- =============================================================================
-- Table 2: templates
-- 场景模板表
-- =============================================================================
CREATE TABLE IF NOT EXISTS templates (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    tool TEXT NOT NULL,  -- 'sysbench', 'swingbench', 'hammerdb'
    database_types TEXT NOT NULL,  -- JSON array: ["mysql", "postgresql"]
    version TEXT NOT NULL,
    parameters_json TEXT NOT NULL,  -- 参数定义（JSON Schema）
    command_template_json TEXT NOT NULL,  -- 命令模板（prepare/run/cleanup）
    output_parser_json TEXT NOT NULL,  -- 输出解析规则（JSON）
    config_json TEXT NOT NULL DEFAULT '',
    is_builtin BOOLEAN NOT NULL DEFAULT 0,  -- 是否为内置模板
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

-- Index for templates
CREATE INDEX IF NOT EXISTS idx_templates_tool ON templates(tool);
CREATE INDEX IF NOT EXISTS idx_templates_is_builtin ON templates(is_builtin);

-- =============================================================================
-- Table 3: tasks
-- 压测任务配置表
-- =============================================================================
CREATE TABLE IF NOT EXISTS tasks (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    connection_id TEXT NOT NULL,
    template_id TEXT NOT NULL,
    parameters_json TEXT NOT NULL,  -- 参数覆盖（JSON）
    options_json TEXT,  -- 执行选项（JSON）
    tags TEXT,  -- 标签（逗号分隔）
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY (connection_id) REFERENCES connections(id) ON DELETE CASCADE,
    FOREIGN KEY (template_id) REFERENCES templates(id) ON DELETE CASCADE
);

-- Index for tasks
CREATE INDEX IF NOT EXISTS idx_tasks_connection_id ON tasks(connection_id);
CREATE INDEX IF NOT EXISTS idx_tasks_template_id ON tasks(template_id);
CREATE INDEX IF NOT EXISTS idx_tasks_created_at ON tasks(created_at DESC);

-- =============================================================================
-- Table 4: runs
-- 运行记录表（核心表）
-- =============================================================================
CREATE TABLE IF NOT EXISTS runs (
    id TEXT PRIMARY KEY,
    task_id TEXT NOT NULL,
    state TEXT NOT NULL,  -- pending, preparing, prepared, warming_up, running,
                          -- completed, failed, cancelled, timeout, force_stopped
    created_at TEXT NOT NULL,
    started_at TEXT,
    completed_at TEXT,
    duration_seconds REAL,
    result_summary_json TEXT,  -- 结果摘要（JSON）
    result_detail_json TEXT,  -- 结果详情（JSON）
    error_message TEXT,
    config_snapshot_path TEXT,  -- 配置快照目录路径
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
);

-- Index for runs
CREATE INDEX IF NOT EXISTS idx_runs_task_id ON runs(task_id);
CREATE INDEX IF NOT EXISTS idx_runs_state ON runs(state);
CREATE INDEX IF NOT EXISTS idx_runs_created_at ON runs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_runs_completed_at ON runs(completed_at DESC);

-- =============================================================================
-- Table 5: metric_samples
-- 时间序列指标表
-- =============================================================================
CREATE TABLE IF NOT EXISTS metric_samples (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    run_id TEXT NOT NULL,
    timestamp TEXT NOT NULL,  -- ISO 8601 format
    phase TEXT NOT NULL,  -- 'warmup', 'run', 'cooldown'
    tps REAL,  -- Transactions Per Second
    qps REAL,  -- Queries Per Second
    latency_avg REAL,  -- Average Latency (ms)
    latency_p95 REAL,  -- 95th Percentile Latency (ms)
    latency_p99 REAL,  -- 99th Percentile Latency (ms)
    error_rate REAL,  -- Error Rate (%)
    FOREIGN KEY (run_id) REFERENCES runs(id) ON DELETE CASCADE
);

-- Index for metric_samples
CREATE INDEX IF NOT EXISTS idx_metric_samples_run_id ON metric_samples(run_id);
CREATE INDEX IF NOT EXISTS idx_metric_samples_timestamp ON metric_samples(timestamp);
CREATE INDEX IF NOT EXISTS idx_metric_samples_phase ON metric_samples(phase);

-- =============================================================================
-- Table 6: run_logs
-- 运行日志表
-- =============================================================================
CREATE TABLE IF NOT EXISTS run_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    run_id TEXT NOT NULL,
    timestamp TEXT NOT NULL,  -- ISO 8601 format
    stream TEXT NOT NULL,  -- 'stdout' or 'stderr'
    content TEXT NOT NULL,
    FOREIGN KEY (run_id) REFERENCES runs(id) ON DELETE CASCADE
);

-- Index for run_logs
CREATE INDEX IF NOT EXISTS idx_run_logs_run_id ON run_logs(run_id);
CREATE INDEX IF NOT EXISTS idx_run_logs_timestamp ON run_logs(timestamp);
CREATE INDEX IF NOT EXISTS idx_run_logs_stream ON run_logs(stream);

-- =============================================================================
-- Table 6.5: history_records
-- 历史记录表（保存成功的运行记录）
-- =============================================================================
CREATE TABLE IF NOT EXISTS history_records (
    id TEXT PRIMARY KEY,  -- Run ID
    created_at TEXT NOT NULL,  -- When the record was created (when saved to history)
    connection_name TEXT NOT NULL,  -- Connection name
    template_name TEXT NOT NULL,  -- Template name
    database_type TEXT NOT NULL,  -- Database type (MySQL/PostgreSQL)
    threads INTEGER NOT NULL,  -- Thread count
    start_time TEXT NOT NULL,  -- Benchmark start time
    duration_seconds REAL NOT NULL,  -- Run duration in seconds
    tps REAL NOT NULL,  -- Transactions per second
    record_json TEXT NOT NULL  -- Full record JSON with all statistics
);

-- Index for history_records
CREATE INDEX IF NOT EXISTS idx_history_records_connection_name ON history_records(connection_name);
CREATE INDEX IF NOT EXISTS idx_history_records_template_name ON history_records(template_name);
CREATE INDEX IF NOT EXISTS idx_history_records_database_type ON history_records(database_type);
CREATE INDEX IF NOT EXISTS idx_history_records_start_time ON history_records(start_time DESC);
CREATE INDEX IF NOT EXISTS idx_history_records_tps ON history_records(tps DESC);

-- =============================================================================
-- Table 7: reports
-- 报告记录表 (AutoBench reports)
-- =============================================================================
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
);

-- Indexes for reports
CREATE INDEX IF NOT EXISTS idx_reports_suite_id ON reports(suite_id);
CREATE INDEX IF NOT EXISTS idx_reports_source_type ON reports(source_type);
CREATE INDEX IF NOT EXISTS idx_reports_connection_id ON reports(connection_id);
CREATE INDEX IF NOT EXISTS idx_reports_status ON reports(status);
CREATE INDEX IF NOT EXISTS idx_reports_started_at ON reports(started_at DESC);
CREATE INDEX IF NOT EXISTS idx_reports_suite_item_id ON reports(suite_item_id);

-- =============================================================================
-- Table 7.5: suites
-- AutoBench 测试套件表
-- =============================================================================
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
);

-- Indexes for suites
CREATE INDEX IF NOT EXISTS idx_suites_status ON suites(status);
CREATE INDEX IF NOT EXISTS idx_suites_started_at ON suites(started_at DESC);

-- =============================================================================
-- Table 8: settings
-- 全局设置表
-- =============================================================================
CREATE TABLE IF NOT EXISTS settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    value_type TEXT NOT NULL,  -- 'string', 'int', 'duration', 'bool'
    description TEXT,
    updated_at TEXT NOT NULL
);

-- Index for settings
CREATE INDEX IF NOT EXISTS idx_settings_key ON settings(key);

-- =============================================================================
-- Templates are synchronized at application startup via the template use case.
-- Do not seed template rows from raw schema initialization.
-- =============================================================================

-- =============================================================================
-- Initial Data: Default Settings
-- 默认设置初始数据
-- =============================================================================

INSERT OR IGNORE INTO settings (key, value, value_type, description, updated_at) VALUES
('result_directory', './results', 'string', '结果文件存储目录', datetime('now')),
('log_level', 'info', 'string', '日志级别：debug/info/warn/error', datetime('now')),
('sample_interval', '1s', 'duration', '默认采样间隔', datetime('now')),
('max_history_runs', '1000', 'int', '最大保留历史记录数', datetime('now')),
('auto_cleanup_days', '30', 'int', '自动清理天数（0=禁用）', datetime('now')),
('default_timeout_prepare', '30m', 'duration', '准备阶段默认超时', datetime('now')),
('default_timeout_run', '24h', 'duration', '运行阶段默认超时', datetime('now')),
('graceful_stop_timeout', '30s', 'duration', '优雅停止超时', datetime('now')),
('enable_keyring', 'true', 'bool', '是否启用 keyring', datetime('now'));

-- =============================================================================
-- Schema Version
-- Schema 版本记录
-- =============================================================================

CREATE TABLE IF NOT EXISTS schema_migrations (
    version INTEGER PRIMARY KEY,
    applied_at TEXT NOT NULL
);

INSERT OR IGNORE INTO schema_migrations (version, applied_at) VALUES (1, datetime('now'));
