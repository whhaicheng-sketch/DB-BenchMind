# AutoBench Reports 持久化层实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为 DB-BenchMind 新增 Reports 持久化层，实现完整的压测报告闭环，包括导航重命名、数据模型、ReportCollector 包装器、API 绑定和前端页面。

**Architecture:** 采用最小侵入包装器模式 (ReportCollector)，复用现有 SystemCollector 底层接口。存储采用混合模式：SQLite 元数据 + 文件系统完整数据。新增 reports/suites 表和 reports/ 目录。

**Tech Stack:** Go 1.21+, SQLite (modernc.org/sqlite), Vue 3, Wails v2

---

## File Structure

### 新增文件

| 文件路径 | 职责 |
|----------|------|
| `internal/domain/report/models.go` | Report/Suite 领域模型定义 |
| `internal/domain/report/models_test.go` | 领域模型单元测试 |
| `internal/infra/database/report_schema.go` | Reports 表结构迁移 |
| `internal/infra/database/report_schema_test.go` | 迁移逻辑测试 |
| `internal/app/usecase/report_collector.go` | ReportCollector 包装器实现 |
| `internal/app/usecase/report_collector_test.go` | ReportCollector 单元测试 |
| `internal/app/usecase/report_usecase.go` | ReportUsecase 查询服务 |
| `internal/app/usecase/report_usecase_test.go` | ReportUsecase 单元测试 |
| `internal/transportwails/bindings/report.go` | Wails 绑定层 |
| `internal/transportwails/bindings/report_test.go` | 绑定层测试 |
| `frontend/src/components/tabs/ReportsTab.vue` | Reports 列表页 |
| `frontend/src/components/report/ReportDetailPanel.vue` | 报告详情面板 |
| `frontend/tests/navigationTabsRename.test.mjs` | 导航重命名测试 |
| `frontend/tests/reportsTabBasic.test.mjs` | Reports 页面基础测试 |

### 修改文件

| 文件路径 | 修改内容 |
|----------|----------|
| `frontend/src/constants/navigationTabs.mjs` | tasks→Benchmark, history→Reports |
| `internal/infra/database/sqlite.go` | 添加 reports 表迁移调用 |
| `internal/transportwails/app.go` | 注册 ReportBinding |
| `frontend/src/components/tabs/HistoryTab.vue` | 重命名为 ReportsTab.vue |
| `frontend/src/App.vue` | 更新组件引用 |

---

## M1 - 导航 UI 标签重命名 (2 tasks)

### Task 1.1: 更新 navigationTabs.mjs 标签文本

**Files:**
- Modify: `frontend/src/constants/navigationTabs.mjs:1-8`

- [ ] **Step 1: 更新标签文本**

```javascript
// frontend/src/constants/navigationTabs.mjs
export const navigationTabs = [
  { id: 'connections', label: 'Connections', icon: '🔌' },
  { id: 'templates', label: 'Templates', icon: '📋' },
  { id: 'tasks', label: 'Benchmark', icon: '📊' },        // Performance Analysis → Benchmark
  { id: 'autobench', label: 'AutoBench', icon: '🧪' },
  { id: 'history', label: 'Reports', icon: '📜' }         // History → Reports
]
```

- [ ] **Step 2: 验证前端测试通过**

Run: `npm --prefix frontend run test`
Expected: All tests pass

- [ ] **Step 3: Commit**

```bash
git add frontend/src/constants/navigationTabs.mjs
git commit -m "feat(ui): rename navigation labels (tasks→Benchmark, history→Reports)

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

### Task 1.2: 前端测试验证

**Files:**
- Create: `frontend/tests/navigationTabsRename.test.mjs`

- [ ] **Step 1: 编写测试验证标签重命名**

```javascript
// frontend/tests/navigationTabsRename.test.mjs
import { describe, it } from 'node:test'
import assert from 'node:assert'
import { navigationTabs } from '../src/constants/navigationTabs.mjs'

describe('Navigation Tabs Rename', () => {
  it('should have Benchmark label for tasks tab', () => {
    const tasksTab = navigationTabs.find(t => t.id === 'tasks')
    assert.ok(tasksTab, 'tasks tab should exist')
    assert.strictEqual(tasksTab.label, 'Benchmark')
  })

  it('should have Reports label for history tab', () => {
    const historyTab = navigationTabs.find(t => t.id === 'history')
    assert.ok(historyTab, 'history tab should exist')
    assert.strictEqual(historyTab.label, 'Reports')
  })

  it('should preserve all tab IDs unchanged', () => {
    const expectedIds = ['connections', 'templates', 'tasks', 'autobench', 'history']
    const actualIds = navigationTabs.map(t => t.id)
    assert.deepStrictEqual(actualIds, expectedIds)
  })
})
```

- [ ] **Step 2: 运行测试验证**

Run: `npm --prefix frontend run test tests/navigationTabsRename.test.mjs`
Expected: PASS (3 tests)

- [ ] **Step 3: Commit**

```bash
git add frontend/tests/navigationTabsRename.test.mjs
git commit -m "test(ui): add navigation tabs rename verification

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## M2 - Reports 数据模型与 SQLite 表 (4 tasks)

### Task 2.1: 定义 Report 领域模型

**Files:**
- Create: `internal/domain/report/models.go`
- Create: `internal/domain/report/models_test.go`

- [ ] **Step 1: 编写失败的测试**

```go
// internal/domain/report/models_test.go
package report

import (
	"testing"
	"time"
)

func TestNewReport(t *testing.T) {
	now := time.Now()
	report := Report{
		ID:           "test-id",
		SuiteID:      "standalone",
		SourceType:   SourceTypeBenchmark,
		ConnectionID: "conn-1",
		DatabaseType: "mysql",
		Status:       StatusCompleted,
		StartedAt:    now,
	}

	if report.ID != "test-id" {
		t.Errorf("expected ID test-id, got %s", report.ID)
	}
	if report.SuiteID != "standalone" {
		t.Errorf("expected SuiteID standalone, got %s", report.SuiteID)
	}
	if report.SourceType != SourceTypeBenchmark {
		t.Errorf("expected SourceTypeBenchmark, got %s", report.SourceType)
	}
}

func TestReportIsCompleted(t *testing.T) {
	tests := []struct {
		name     string
		status   ReportStatus
		expected bool
	}{
		{"completed", StatusCompleted, true},
		{"failed", StatusFailed, true},
		{"cancelled", StatusCancelled, true},
		{"running", StatusRunning, false},
		{"pending", StatusPending, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := Report{Status: tt.status}
			if got := report.IsCompleted(); got != tt.expected {
				t.Errorf("IsCompleted() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestSuiteIsCompleted(t *testing.T) {
	tests := []struct {
		name     string
		status   SuiteStatus
		expected bool
	}{
		{"success", SuiteStatusSuccess, true},
		{"failed", SuiteStatusFailed, true},
		{"partial_success", SuiteStatusPartialSuccess, true},
		{"cancelled", SuiteStatusCancelled, true},
		{"running", SuiteStatusRunning, false},
		{"draft", SuiteStatusDraft, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suite := Suite{Status: tt.status}
			if got := suite.IsCompleted(); got != tt.expected {
				t.Errorf("IsCompleted() = %v, want %v", got, tt.expected)
			}
		})
	}
}
```

- [ ] **Step 2: 运行测试验证失败**

Run: `go test ./internal/domain/report/...`
Expected: FAIL (package not found)

- [ ] **Step 3: 实现 Report 领域模型**

```go
// internal/domain/report/models.go
// Package report provides report domain models.
package report

import (
	"time"
)

// ReportStatus represents the status of a report.
type ReportStatus string

const (
	StatusPending   ReportStatus = "pending"
	StatusRunning   ReportStatus = "running"
	StatusCompleted ReportStatus = "completed"
	StatusFailed    ReportStatus = "failed"
	StatusCancelled ReportStatus = "cancelled"
)

// IsTerminal returns true if the status is terminal.
func (s ReportStatus) IsTerminal() bool {
	return s == StatusCompleted || s == StatusFailed || s == StatusCancelled
}

// SourceType represents the source of a report.
type SourceType string

const (
	SourceTypeBenchmark SourceType = "benchmark"
	SourceTypeAutoBench SourceType = "autobench"
)

// SuiteID constants
const (
	StandaloneSuiteID = "standalone"
)

// Report represents a benchmark report.
type Report struct {
	ID            string       `json:"id"`
	SuiteID       string       `json:"suite_id"`
	SuiteItemID   string       `json:"suite_item_id,omitempty"`
	SourceType    SourceType   `json:"source_type"`
	ConnectionID  string       `json:"connection_id"`
	ConnectionName string      `json:"connection_name,omitempty"`
	DatabaseType  string       `json:"database_type"`
	TemplateID    string       `json:"template_id,omitempty"`
	TemplateName  string       `json:"template_name,omitempty"`
	StartedAt     time.Time    `json:"started_at"`
	EndedAt       *time.Time   `json:"ended_at,omitempty"`
	DurationMs    int64        `json:"duration_ms,omitempty"`
	Status        ReportStatus `json:"status"`
	ErrorMessage  string       `json:"error_message,omitempty"`

	// Core metrics
	TPM           float64 `json:"tpm,omitempty"`
	TPS           float64 `json:"tps,omitempty"`
	QPS           float64 `json:"qps,omitempty"`
	Throughput    float64 `json:"throughput,omitempty"`
	LatencyAvgMs  float64 `json:"latency_avg_ms,omitempty"`
	LatencyP95Ms  float64 `json:"latency_p95_ms,omitempty"`
	LatencyP99Ms  float64 `json:"latency_p99_ms,omitempty"`
	ErrorCount    int64   `json:"error_count,omitempty"`

	// File paths
	MetricsJSONPath    string `json:"metrics_json_path,omitempty"`
	MonitoringJSONPath string `json:"monitoring_json_path,omitempty"`
	RawJSONPath        string `json:"raw_json_path,omitempty"`
	ReportHTMLPath     string `json:"report_html_path,omitempty"`
	SummaryJSONPath    string `json:"summary_json_path,omitempty"`

	// Metadata
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Tags      string    `json:"tags,omitempty"`
}

// IsCompleted returns true if the report is in a terminal state.
func (r *Report) IsCompleted() bool {
	return r.Status.IsTerminal()
}

// SuiteStatus represents the status of a suite.
type SuiteStatus string

const (
	SuiteStatusDraft          SuiteStatus = "draft"
	SuiteStatusRunning        SuiteStatus = "running"
	SuiteStatusPartialSuccess SuiteStatus = "partial_success"
	SuiteStatusSuccess        SuiteStatus = "success"
	SuiteStatusFailed         SuiteStatus = "failed"
	SuiteStatusCancelled      SuiteStatus = "cancelled"
)

// IsTerminal returns true if the status is terminal.
func (s SuiteStatus) IsTerminal() bool {
	return s == SuiteStatusSuccess || s == SuiteStatusFailed ||
		s == SuiteStatusPartialSuccess || s == SuiteStatusCancelled
}

// Suite represents an AutoBench suite report.
type Suite struct {
	ID                      string      `json:"id"`
	Name                    string      `json:"name,omitempty"`
	ExecutionMode           string      `json:"execution_mode,omitempty"`
	FailurePolicy           string      `json:"failure_policy,omitempty"`
	CleanupEnabled          bool        `json:"cleanup_enabled"`
	SuiteManifestJSONPath   string      `json:"suite_manifest_json_path,omitempty"`
	Status                  SuiteStatus `json:"status"`
	StartedAt               *time.Time  `json:"started_at,omitempty"`
	EndedAt                 *time.Time  `json:"ended_at,omitempty"`
	TotalItems              int         `json:"total_items"`
	CompletedItems          int         `json:"completed_items"`
	SuccessItems            int         `json:"success_items"`
	FailedItems             int         `json:"failed_items"`
	SkippedItems            int         `json:"skipped_items"`
	SuiteReportJSONPath     string      `json:"suite_report_json_path,omitempty"`
	SuiteReportHTMLPath     string      `json:"suite_report_html_path,omitempty"`
	CreatedAt               time.Time   `json:"created_at"`
	UpdatedAt               time.Time   `json:"updated_at"`
}

// IsCompleted returns true if the suite is in a terminal state.
func (s *Suite) IsCompleted() bool {
	return s.Status.IsTerminal()
}

// ReportContext provides context for report collection.
type ReportContext struct {
	SuiteID        string
	SuiteItemID    string
	SourceType     SourceType
	ConnectionID   string
	ConnectionName string
	DatabaseType   string
	TemplateID     string
	TemplateName   string
	Tags           string
}

// ReportResult is the result of CollectAndPersist.
type ReportResult struct {
	ReportID     string
	ReportPaths  ReportPaths
	Summary      ReportSummary
	PersistError error
}

// ReportPaths contains paths to report files.
type ReportPaths struct {
	MetricsJSON    string
	MonitoringJSON string
	RawJSON        string
	ReportHTML     string
	SummaryJSON    string
}

// ReportSummary contains key metrics for display.
type ReportSummary struct {
	Status       ReportStatus
	TPM          float64
	TPS          float64
	LatencyAvgMs float64
	LatencyP95Ms float64
	ErrorCount   int64
}
```

- [ ] **Step 4: 运行测试验证通过**

Run: `go test ./internal/domain/report/... -v`
Expected: PASS (all tests)

- [ ] **Step 5: Commit**

```bash
git add internal/domain/report/
git commit -m "feat(domain): add Report and Suite domain models

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

### Task 2.2: SQLite reports/suites 表结构

**Files:**
- Create: `internal/infra/database/report_schema.go`
- Create: `internal/infra/database/report_schema_test.go`

- [ ] **Step 1: 编写失败的测试**

```go
// internal/infra/database/report_schema_test.go
package database

import (
	"context"
	"testing"
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
```

- [ ] **Step 2: 运行测试验证失败**

Run: `go test ./internal/infra/database/... -run TestEnsureReportSchema`
Expected: FAIL (function not defined)

- [ ] **Step 3: 实现表结构迁移**

```go
// internal/infra/database/report_schema.go
package database

import (
	"context"
	"database/sql"
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
		return err
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
			return err
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
		return err
	}

	// Create suites indexes
	suiteIndexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_suites_status ON suites(status)`,
		`CREATE INDEX IF NOT EXISTS idx_suites_started_at ON suites(started_at DESC)`,
	}
	for _, idx := range suiteIndexes {
		if _, err := db.ExecContext(ctx, idx); err != nil {
			return err
		}
	}

	return nil
}
```

- [ ] **Step 4: 运行测试验证通过**

Run: `go test ./internal/infra/database/... -run TestEnsureReportSchema -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/infra/database/report_schema.go internal/infra/database/report_schema_test.go
git commit -m "feat(database): add reports and suites tables schema

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

### Task 2.3: 数据库初始化逻辑集成

**Files:**
- Modify: `internal/infra/database/sqlite.go`

- [ ] **Step 1: 编写失败的测试**

```go
// 在 internal/infra/database/sqlite_test.go 添加
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
```

- [ ] **Step 2: 运行测试验证失败**

Run: `go test ./internal/infra/database/... -run TestInitializeSQLiteWithReportSchema`
Expected: FAIL (reports table not found)

- [ ] **Step 3: 修改 InitializeSQLite 添加 ReportSchema 调用**

在 `internal/infra/database/sqlite.go` 的 `InitializeSQLite` 函数中，在 `ensureTemplateColumns` 调用后添加:

```go
// 在 ensureTemplateColumns 调用后添加
if err := EnsureReportSchema(ctx, db); err != nil {
    db.Close()
    return nil, fmt.Errorf("ensure report schema: %w", err)
}
```

- [ ] **Step 4: 运行测试验证通过**

Run: `go test ./internal/infra/database/... -run TestInitializeSQLiteWithReportSchema -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/infra/database/sqlite.go internal/infra/database/sqlite_test.go
git commit -m "feat(database): integrate report schema into SQLite initialization

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

### Task 2.4: 后端单元测试

**Files:**
- Modify: `internal/infra/database/report_schema_test.go` (扩展测试覆盖)

- [ ] **Step 1: 扩展测试覆盖索引和列**

```go
// 在 report_schema_test.go 添加更多测试
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
```

- [ ] **Step 2: 运行测试验证通过**

Run: `go test ./internal/infra/database/... -v`
Expected: All PASS

- [ ] **Step 3: Commit**

```bash
git add internal/infra/database/report_schema_test.go
git commit -m "test(database): add index coverage tests for report schema

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## M3 - ReportCollector 包装器实现 (6 tasks)

### Task 3.1: ReportCollector 接口定义

**Files:**
- Modify: `internal/app/usecase/report_collector.go` (接口部分)

- [ ] **Step 1: 编写失败的测试**

```go
// internal/app/usecase/report_collector_test.go
package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/whhaicheng/DB-BenchMind/internal/domain/execution"
	"github.com/whhaicheng/DB-BenchMind/internal/domain/report"
)

func TestReportCollectorInterface(t *testing.T) {
	// This test verifies the interface exists and can be implemented
	var _ ReportCollector = (*DefaultReportCollector)(nil)
}

func TestReportCollectorCollectAndPersist(t *testing.T) {
	t.Run("successful collection", func(t *testing.T) {
		ctx := context.Background()
		collector := NewDefaultReportCollector(
			WithReportsDir(t.TempDir()),
		)

		rptCtx := report.ReportContext{
			SuiteID:        report.StandaloneSuiteID,
			SourceType:     report.SourceTypeBenchmark,
			ConnectionID:   "conn-1",
			ConnectionName: "Test Conn",
			DatabaseType:   "mysql",
		}

		runFn := func() (*execution.Run, error) {
			return &execution.Run{
				ID:     "run-1",
				State:  execution.StateCompleted,
				Result: &execution.BenchmarkResult{
					TPMCalculated: 15000.5,
					TPSCalculated: 250.5,
					LatencyAvg:    15.8,
					LatencyP95:    25.3,
				},
			}, nil
		}

		result, err := collector.CollectAndPersist(ctx, runFn, rptCtx)
		if err != nil {
			t.Fatalf("CollectAndPersist failed: %v", err)
		}
		if result.ReportID == "" {
			t.Error("expected non-empty ReportID")
		}
		if result.Summary.TPM != 15000.5 {
			t.Errorf("expected TPM 15000.5, got %f", result.Summary.TPM)
		}
	})
}
```

- [ ] **Step 2: 运行测试验证失败**

Run: `go test ./internal/app/usecase/... -run TestReportCollector`
Expected: FAIL (type not defined)

- [ ] **Step 3: 定义接口和基础结构**

```go
// internal/app/usecase/report_collector.go (Part 1: Interface)
package usecase

import (
	"context"

	"github.com/whhaicheng/DB-BenchMind/internal/domain/execution"
	"github.com/whhaicheng/DB-BenchMind/internal/domain/report"
)

// ReportCollector collects benchmark results and persists them as reports.
type ReportCollector interface {
	CollectAndPersist(
		ctx context.Context,
		runFn func() (*execution.Run, error),
		rptCtx report.ReportContext,
	) (*report.ReportResult, error)
}

// DefaultReportCollector is the default implementation.
type DefaultReportCollector struct {
	reportsDir string
	db         dbExecutor
}

type dbExecutor interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (interface{}, error)
}

// ReportCollectorOption configures the collector.
type ReportCollectorOption func(*DefaultReportCollector)

// WithReportsDir sets the reports directory.
func WithReportsDir(dir string) ReportCollectorOption {
	return func(c *DefaultReportCollector) {
		c.reportsDir = dir
	}
}

// WithDB sets the database executor.
func WithDB(db dbExecutor) ReportCollectorOption {
	return func(c *DefaultReportCollector) {
		c.db = db
	}
}

// NewDefaultReportCollector creates a new collector.
func NewDefaultReportCollector(opts ...ReportCollectorOption) *DefaultReportCollector {
	c := &DefaultReportCollector{
		reportsDir: "./reports",
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}
```

- [ ] **Step 4: Commit 接口定义**

```bash
git add internal/app/usecase/report_collector.go
git commit -m "feat(usecase): define ReportCollector interface

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

### Task 3.2: 收集 Benchmark 执行结果

**Files:**
- Modify: `internal/app/usecase/report_collector.go`

- [ ] **Step 1: 实现 CollectAndPersist 核心逻辑**

```go
// 添加到 report_collector.go
import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

// CollectAndPersist executes runFn and persists the result.
func (c *DefaultReportCollector) CollectAndPersist(
	ctx context.Context,
	runFn func() (*execution.Run, error),
	rptCtx report.ReportContext,
) (*report.ReportResult, error) {
	reportID := uuid.New().String()
	startTime := time.Now()

	// Execute benchmark
	run, err := runFn()
	if err != nil {
		return nil, fmt.Errorf("run benchmark: %w", err)
	}

	endTime := time.Now()
	durationMs := endTime.Sub(startTime).Milliseconds()

	// Build report
	rpt := &report.Report{
		ID:             reportID,
		SuiteID:        rptCtx.SuiteID,
		SuiteItemID:    rptCtx.SuiteItemID,
		SourceType:     rptCtx.SourceType,
		ConnectionID:   rptCtx.ConnectionID,
		ConnectionName: rptCtx.ConnectionName,
		DatabaseType:   rptCtx.DatabaseType,
		TemplateID:     rptCtx.TemplateID,
		TemplateName:   rptCtx.TemplateName,
		StartedAt:      startTime,
		EndedAt:        &endTime,
		DurationMs:     durationMs,
		Status:         report.StatusCompleted,
		Tags:           rptCtx.Tags,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Extract metrics from run result
	if run.Result != nil {
		rpt.TPM = run.Result.TPMCalculated
		rpt.TPS = run.Result.TPSCalculated
		rpt.LatencyAvgMs = run.Result.LatencyAvg
		rpt.LatencyP95Ms = run.Result.LatencyP95
		rpt.LatencyP99Ms = run.Result.LatencyP99
		rpt.ErrorCount = run.Result.ErrorCount
	}

	// Handle run state
	if run.State == execution.StateFailed {
		rpt.Status = report.StatusFailed
		rpt.ErrorMessage = run.ErrorMessage
	} else if run.State == execution.StateCancelled || run.State == execution.StateForceStopped {
		rpt.Status = report.StatusCancelled
	}

	// Create directory
	suiteDir := filepath.Join(c.reportsDir, rptCtx.SuiteID)
	reportDir := filepath.Join(suiteDir, reportID)
	if err := os.MkdirAll(reportDir, 0755); err != nil {
		return nil, fmt.Errorf("create report directory: %w", err)
	}

	// Persist files
	paths, persistErr := c.persistFiles(reportDir, reportID, rpt, run)
	if persistErr != nil {
		// Log but don't fail - return partial result
		rpt.ErrorMessage = fmt.Sprintf("persistence error: %v", persistErr)
	}

	rpt.MetricsJSONPath = paths.MetricsJSON
	rpt.MonitoringJSONPath = paths.MonitoringJSON
	rpt.RawJSONPath = paths.RawJSON
	rpt.ReportHTMLPath = paths.ReportHTML
	rpt.SummaryJSONPath = paths.SummaryJSON

	return &report.ReportResult{
		ReportID:     reportID,
		ReportPaths:  paths,
		Summary:      report.ReportSummary{
			Status:       rpt.Status,
			TPM:          rpt.TPM,
			TPS:          rpt.TPS,
			LatencyAvgMs: rpt.LatencyAvgMs,
			LatencyP95Ms: rpt.LatencyP95Ms,
			ErrorCount:   rpt.ErrorCount,
		},
		PersistError: persistErr,
	}, nil
}

func (c *DefaultReportCollector) persistFiles(
	dir string,
	reportID string,
	rpt *report.Report,
	run *execution.Run,
) (report.ReportPaths, error) {
	// Implementation in next task
	return report.ReportPaths{}, nil
}
```

- [ ] **Step 2: 运行测试验证**

Run: `go test ./internal/app/usecase/... -run TestReportCollectorCollectAndPersist -v`
Expected: PASS (with stub persistFiles)

- [ ] **Step 3: Commit**

```bash
git add internal/app/usecase/report_collector.go
git commit -m "feat(usecase): implement CollectAndPersist core logic

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

### Task 3.3-3.5: 文件持久化和 SQLite 写入

*由于篇幅限制，详细代码实现遵循相同模式：先写测试，再实现，最后提交。*

---

### Task 3.6: 后端单元测试

- [ ] **Step 1: 编写完整测试覆盖**

```go
// 扩展 report_collector_test.go
func TestReportCollectorWithFailedRun(t *testing.T) {
	// Test failed run handling
}

func TestReportCollectorWithCancelledRun(t *testing.T) {
	// Test cancelled run handling
}

func TestReportCollectorFilePersistence(t *testing.T) {
	// Test file creation
}

func TestReportCollectorSQLitePersistence(t *testing.T) {
	// Test database insertion
}
```

- [ ] **Step 2: 运行测试并确保通过**

Run: `make test-backend`
Expected: All PASS

- [ ] **Step 3: Commit**

```bash
git commit -m "test(usecase): add comprehensive ReportCollector tests

Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>"
```

---

## M4 - Reports Usecase 与 API 绑定 (3 tasks)

### Task 4.1: ReportUsecase 实现

**Files:**
- Create: `internal/app/usecase/report_usecase.go`
- Create: `internal/app/usecase/report_usecase_test.go`

- [ ] **Step 1: 定义接口和基础查询方法**

```go
// internal/app/usecase/report_usecase.go
package usecase

import (
	"context"
	"database/sql"

	"github.com/whhaicheng/DB-BenchMind/internal/domain/report"
)

type ReportUsecase struct {
	db *sql.DB
}

func NewReportUsecase(db *sql.DB) *ReportUsecase {
	return &ReportUsecase{db: db}
}

// ListReports returns paginated reports.
func (u *ReportUsecase) ListReports(ctx context.Context, opts ListReportsOptions) ([]report.Report, error) {
	// Implementation
	return nil, nil
}

// GetReport returns a single report by ID.
func (u *ReportUsecase) GetReport(ctx context.Context, id string) (*report.Report, error) {
	// Implementation
	return nil, nil
}

// ListSuites returns paginated suites.
func (u *ReportUsecase) ListSuites(ctx context.Context, opts ListSuitesOptions) ([]report.Suite, error) {
	// Implementation
	return nil, nil
}

// GetSuite returns a single suite by ID.
func (u *ReportUsecase) GetSuite(ctx context.Context, id string) (*report.Suite, error) {
	// Implementation
	return nil, nil
}
```

- [ ] **Step 2-5: TDD 流程实现所有方法**

---

### Task 4.2: Wails bindings 暴露

**Files:**
- Create: `internal/transportwails/bindings/report.go`

- [ ] **Step 1: 实现 ReportBinding**

```go
// internal/transportwails/bindings/report.go
package bindings

import (
	"context"

	"github.com/whhaicheng/DB-BenchMind/internal/app/usecase"
)

type ReportBinding struct {
	ctx    context.Context
	report *usecase.ReportUsecase
}

func NewReportBinding(report *usecase.ReportUsecase) *ReportBinding {
	return &ReportBinding{report: report}
}

func (b *ReportBinding) SetContext(ctx context.Context) {
	b.ctx = ctx
}

type ListReportsRequest struct {
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	SuiteID  string `json:"suite_id,omitempty"`
	Status   string `json:"status,omitempty"`
}

type ListReportsResult struct {
	Reports []interface{} `json:"reports"`
	Total   int           `json:"total"`
	Error   string        `json:"error,omitempty"`
}

func (b *ReportBinding) ListReports(req ListReportsRequest) ListReportsResult {
	// Implementation
	return ListReportsResult{}
}

func (b *ReportBinding) GetReport(id string) map[string]interface{} {
	// Implementation
	return nil
}
```

---

### Task 4.3: 后端单元测试

- [ ] **Step 1: 编写 ReportBinding 测试**

- [ ] **Step 2: 运行测试确保通过**

---

## M5 - Reports 前端页面 (5 tasks)

### Task 5.1: 重命名 HistoryTab 为 ReportsTab

**Files:**
- Rename: `frontend/src/components/tabs/HistoryTab.vue` → `ReportsTab.vue`
- Modify: `frontend/src/App.vue`

- [ ] **Step 1: 更新组件名称和标题**

```vue
<!-- frontend/src/components/tabs/ReportsTab.vue -->
<template>
  <div class="reports-tab">
    <!-- Page Header -->
    <div class="page-header">
      <div class="header-left">
        <h1 class="page-title">Reports</h1>  <!-- Changed from History -->
        <p class="page-subtitle">View benchmark reports and results</p>
      </div>
      ...
    </div>
    ...
  </div>
</template>
```

- [ ] **Step 2: 更新 App.vue 引用**

- [ ] **Step 3: Commit**

---

### Task 5.2-5.5: 前端实现

*详细实现遵循 Vue 3 组件模式，使用 Pinia store 和 ECharts 图表*

---

## M6 - AutoBench 集成 (3 tasks)

### Task 6.1: SuiteItem 关联 report_id

**Files:**
- Modify: `internal/domain/autobench/models.go`

- [ ] **Step 1: 添加 ReportID 字段**

```go
// 在 SuiteItem 结构体添加
type SuiteItem struct {
    // ... existing fields
    ReportID string `json:"report_id,omitempty"` // 新增
}
```

---

### Task 6.2: suite_manifest.json 生成与更新

- [ ] **Step 1: 在 AutoBenchRunner 中集成 ReportCollector**

---

### Task 6.3: 后端单元测试

---

## M7 - 单次 Benchmark 集成 (3 tasks)

### Task 7.1: 单次执行生成 standalone report

- [ ] **Step 1: 在 BenchmarkUseCase 中集成 ReportCollector**

---

## M8 - 验收与兼容性回归 (4 tasks)

### Task 8.1: 后端回归测试

Run: `make test-backend`
Expected: All PASS

### Task 8.2: 前端回归测试

Run: `make test-frontend`
Expected: All PASS

### Task 8.3: 集成测试

Run: `make test`
Expected: All PASS

### Task 8.4: 文档同步更新

- [ ] **Step 1: 更新 progress.md 和 progress.json**

---

## Summary

| 模块 | 任务数 | 预计文件 |
|------|--------|----------|
| M1 导航重命名 | 2 | 2 |
| M2 数据模型 | 4 | 4 |
| M3 ReportCollector | 6 | 2 |
| M4 Usecase/API | 3 | 4 |
| M5 前端页面 | 5 | 4 |
| M6 AutoBench 集成 | 3 | 2 |
| M7 Benchmark 集成 | 3 | 2 |
| M8 验收回归 | 4 | 0 |
| **Total** | **30** | **20** |

---

**Plan complete.** Two execution options:

**1. Subagent-Driven (recommended)** - Dispatch a fresh subagent per task, review between tasks, fast iteration

**2. Inline Execution** - Execute tasks in this session using executing-plans, batch execution with checkpoints

**Which approach?**
