# DB-BenchMind 任务分解列表

**版本**: 1.0.0
**日期**: 2026-01-27
**状态**: 执行中
**技术组长**: AI Assistant

---

## 📋 文档说明

本文档将 `specs/plan.md` 中的技术方案分解为**原子化、可执行的任务列表**，确保每个任务都可以被 AI 独立完成。

### 任务格式说明

```markdown
### Phase X: [阶段名称]

#### Task X.Y: [任务标题] [P]
**Type**: test/impl
**File**: 文件路径
**Depends**: X.A, X.B（依赖的任务ID）
**Description**: 详细描述
**Acceptance**: 验收标准
**Implementation**: 实现要点
```

- **[P]** 标记表示该任务可与其他标记 `[P]` 的任务并行执行
- **Type**: `test` 表示测试任务，`impl` 表示实现任务
- **TDD 强制**: 测试任务必须在实现任务之前（Red → Green → Refactor）

---

## Phase 1: 项目初始化与基础设施

**目标**: 搭建项目骨架、配置工具链、初始化数据库

---

#### Task 1.1: 创建项目目录结构 [P]
**Type**: impl
**File**: `完整目录树`
**Depends**: 无
**Description**: 根据 plan.md 第3.1节的目录树，创建完整的目录结构
**Acceptance**:
- 所有目录创建成功
- `tree -L 3` 显示结构与 plan.md 一致
**Command**:
```bash
# 创建主目录
mkdir -p cmd/{db-benchmind,cli-test}
mkdir -p internal/{app/usecase,domain/{connection,template,execution,metric}}
mkdir -p internal/{infra/{adapter,database,keyring,report,chart},transport/ui/widgets}
mkdir -p pkg/benchmark
mkdir -p contracts/{templates,schemas,reports}
mkdir -p configs scripts test/{testdata/{connections,outputs,expected},integration}
mkdir -p docs .specify/steering results

# 验证
tree -L 3 -d
```

---

#### Task 1.2: 初始化 go.mod [P]
**Type**: impl
**File**: `go.mod`
**Depends**: 1.1
**Description**: 初始化 Go module，设置 Go 版本为 1.22.2，添加核心依赖
**Acceptance**:
- go.mod 文件创建成功
- `go mod tidy` 无错误
- `go build ./...` 无错误
**Content**:
```go
module github.com/whhaicheng/DB-BenchMind

go 1.22.2

toolchain go1.22.2

require (
    fyne.io/fyne/v2 v2.4.5
    modernc.org/sqlite v1.28.0
    github.com/zalando/go-keyring v0.2.0
    github.com/google/uuid v1.5.0
)
```

---

#### Task 1.3: 创建 Makefile [P]
**Type**: impl
**File**: `Makefile`
**Depends**: 1.1
**Description**: 创建 Makefile，包含 build, test, lint, check, clean, run 目标
**Acceptance**:
- `make build` 成功生成二进制
- `make test` 运行所有测试
- `make lint` 运行 linter
- `make check` 运行所有检查
- `make clean` 清理构建产物
**Content**:
```makefile
.PHONY: all build test lint check clean run

BINARY_NAME=db-benchmind
BUILD_DIR=build
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-X main.Version=$(VERSION)"

all: check

build: clean
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/db-benchmind
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

build-dev:
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/db-benchmind

test:
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage: coverage.html"

test-integration:
	@echo "Running integration tests..."
	@go test -v -tags=integration ./...

lint:
	@echo "Running linter..."
	@golangci-lint run ./...

fmt:
	@echo "Checking format..."
	@test -z $$(gofmt -l .)

vet:
	@echo "Running vet..."
	@go vet ./...

sec:
	@echo "Running security scan..."
	@govulncheck ./...

check: fmt vet lint test

clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html

run: build-dev
	@./$(BUILD_DIR)/$(BINARY_NAME)

install-tools:
	@echo "Installing tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install golang.org/x/vuln/cmd/govulncheck@latest
```

---

#### Task 1.4: 配置 golangci-lint [P]
**Type**: impl
**File**: `.golangci.yml`
**Depends**: 1.1
**Description**: 配置 golangci-lint 规则
**Acceptance**:
- 配置文件创建成功
- `golangci-lint run` 成功执行
- 零错误
**Content**:
```yaml
run:
  timeout: 5m
  tests: true
  build-tags: []
  skip-dirs:
    - vendor
    - testdata
    - results

linters:
  enable:
    - gofmt
    - govet
    - staticcheck
    - unused
    - errcheck
    - gosec
    - ineffassign
    - deadcode
    - varcheck
    - structcheck
    - misspell

linters-settings:
  govet:
    enable-all: true
    disable:
      - shadow
  errcheck:
    check-blank: true
    check-typeAssertions: true
  gosec:
    excludes:
      - G104

issues:
  exclude-use-default: false
  max-issues-per-linter: 0
  max-same-issues: 0
```

---

#### Task 1.5: 创建 .gitignore [P]
**Type**: impl
**File**: `.gitignore`
**Depends**: 1.1
**Description**: 创建 .gitignore 文件，排除所有构建产物和临时文件
**Acceptance**:
- .gitignore 文件创建成功
- `git status` 不显示无关文件
**Content**:
```
# Binaries
db-benchmind
*.exe
*.exe~
*.dll
*.so
*.dylib
/bin/
/build/

# Test binary
*.test

# Output
*.out
coverage.html
coverage.out

# Go workspace
/go.work
/go.work.sum

# IDE
.idea/
.vscode/
*.swp
*.swo
*~

# OS
.DS_Store
Thumbs.db

# Results
results/
*.log

# Temporary
*.tmp
temp/

# Dependencies
vendor/

# Keys
*.key
*.pem

# Compression
*.zip
*.tar.gz
```

---

#### Task 1.6: 创建产品定义文档 [P]
**Type**: impl
**File**: `.specify/steering/product.md`
**Depends**: 1.1
**Description**: 编写产品定义文档
**Acceptance**:
- 文档创建成功
- 包含问题定义、用户与范围边界
**Content**:
```markdown
# 产品定义 (Product)

## 问题定义

数据库工程师和性能测试工程师在日常工作中面临以下问题：

1. **工具分散**: Sysbench、Swingbench、HammerDB 等工具各有特点，但命令行操作复杂
2. **结果难复现**: 缺乏配置和结果的完整保存，难以精确复现测试
3. **对比困难**: 多次测试结果缺乏统一的对比分析工具
4. **报告繁琐**: 手动整理测试报告耗时且容易出错
5. **学习曲线**: 新手难以快速上手专业的压测工具

## 解决方案

DB-BenchMind - 一款桌面压测工作台，提供统一的 GUI 界面来：
- 编排和运行多种压测工具
- 实时监控测试过程
- 自动存储和归档结果
- 生成多格式测试报告
- 对比分析多次运行结果

## 目标用户

### 主要用户
- **数据库工程师**: 需要进行性能测试和调优
- **性能测试工程师**: 专业从事数据库性能测试
- **数据库架构师**: 需要选型和容量规划数据

### 使用场景
- **性能基准测试**: 新系统上线前的性能基准
- **容量规划**: 评估系统承载能力
- **数据库选型**: 对比不同数据库性能
- **优化验证**: 验证调优效果
- **回归测试**: 升级后性能回归检测

## 范围边界

### 包含 (In Scope)
- 支持 4 种数据库：MySQL, Oracle, SQL Server, PostgreSQL
- 支持 3 种压测工具：Sysbench, Swingbench, HammerDB
- 提供桌面 GUI 操作界面
- 连接管理和密码安全存储
- 内置常用压测模板
- 实时监控和指标采集
- 结果存储和历史查询
- 多格式报告导出（MD/HTML/JSON/PDF）
- 结果对比分析

### 不包含 (Out of Scope)
- 分布式压测（多机协作）
- Web UI（仅桌面 GUI）
- 插件系统
- 自定义脚本执行
- 云服务集成
- 实时告警通知

## 成功指标
- 测试配置 100% 可复现
- 报告生成时间 < 10 秒
- 支持 1000+ 条历史记录
- GUI 响应时间 < 100ms
- 内存占用 < 500MB
```

---

#### Task 1.7: 编写数据库 Schema
**Type**: impl
**File**: `internal/infra/database/schema.sql`
**Depends**: 1.1
**Description**: 根据 plan.md 第3.6.1节编写完整的 SQLite Schema
**Acceptance**:
- SQL 语法正确
- 包含所有表：connections, templates, tasks, runs, metric_samples, run_logs, settings, reports
- 包含所有索引
- 包含外键约束
**Content**:
```sql
-- ================================================================
-- DB-BenchMind Database Schema
-- Version: 1.0.0
-- ================================================================

-- 连接表
CREATE TABLE IF NOT EXISTS connections (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    type TEXT NOT NULL,
    config_json TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_connections_type ON connections(type);

-- 模板表
CREATE TABLE IF NOT EXISTS templates (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    tool TEXT NOT NULL,
    database_types TEXT NOT NULL,
    definition_json TEXT NOT NULL,
    is_builtin BOOLEAN NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_templates_tool ON templates(tool);
CREATE INDEX IF NOT EXISTS idx_templates_builtin ON templates(is_builtin);

-- 任务表
CREATE TABLE IF NOT EXISTS tasks (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    connection_id TEXT NOT NULL,
    template_id TEXT NOT NULL,
    parameters_json TEXT NOT NULL,
    options_json TEXT NOT NULL,
    tags TEXT,
    created_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_tasks_connection ON tasks(connection_id);
CREATE INDEX IF NOT EXISTS idx_tasks_template ON tasks(template_id);
CREATE INDEX IF NOT EXISTS idx_tasks_created ON tasks(created_at DESC);

-- 运行表
CREATE TABLE IF NOT EXISTS runs (
    id TEXT PRIMARY KEY,
    task_id TEXT NOT NULL,
    state TEXT NOT NULL,
    created_at TEXT NOT NULL,
    started_at TEXT,
    completed_at TEXT,
    duration_seconds REAL,
    result_json TEXT,
    error_message TEXT,
    work_dir TEXT
);

CREATE INDEX IF NOT EXISTS idx_runs_task ON runs(task_id);
CREATE INDEX IF NOT EXISTS idx_runs_state ON runs(state);
CREATE INDEX IF NOT EXISTS idx_runs_created ON runs(created_at DESC);

-- 时间序列指标表
CREATE TABLE IF NOT EXISTS metric_samples (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    run_id TEXT NOT NULL,
    timestamp TEXT NOT NULL,
    phase TEXT NOT NULL,
    tps REAL,
    qps REAL,
    latency_avg REAL,
    latency_p95 REAL,
    latency_p99 REAL,
    error_rate REAL,
    FOREIGN KEY (run_id) REFERENCES runs(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_metric_samples_run ON metric_samples(run_id);
CREATE INDEX IF NOT EXISTS idx_metric_samples_timestamp ON metric_samples(timestamp);

-- 运行日志表
CREATE TABLE IF NOT EXISTS run_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    run_id TEXT NOT NULL,
    timestamp TEXT NOT NULL,
    stream TEXT NOT NULL,
    content TEXT NOT NULL,
    FOREIGN KEY (run_id) REFERENCES runs(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_run_logs_run ON run_logs(run_id);
CREATE INDEX IF NOT EXISTS idx_run_logs_timestamp ON run_logs(timestamp);

-- 设置表
CREATE TABLE IF NOT EXISTS settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

-- 报告表
CREATE TABLE IF NOT EXISTS reports (
    id TEXT PRIMARY KEY,
    run_id TEXT NOT NULL,
    format TEXT NOT NULL,
    file_path TEXT NOT NULL,
    created_at TEXT NOT NULL,
    FOREIGN KEY (run_id) REFERENCES runs(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_reports_run ON reports(run_id);
```

---

#### Task 1.8: [测试] 测试 SQLite 初始化
**Type**: test
**File**: `internal/infra/database/sqlite_test.go`
**Depends**: 1.7
**Description**: 编写 SQLite 初始化函数的测试（TDD: Red）
**Acceptance**:
- 测试数据库创建成功
- 测试所有表存在
- 测试索引创建成功
- 测试外键约束生效
- 测试当前运行失败（因为实现不存在）
**Content**:
```go
package database

import (
    "database/sql"
    "testing"

    _ "modernc.org/sqlite"
)

func TestOpenSQLite(t *testing.T) {
    t.Run("opens in-memory database", func(t *testing.T) {
        db, err := OpenSQLite(":memory:")
        if err != nil {
            t.Fatalf("OpenSQLite failed: %v", err)
        }
        defer db.Close()

        if db.Ping() != nil {
            t.Error("database is not pingable")
        }
    })
}

func TestInitSchema(t *testing.T) {
    t.Run("creates all tables", func(t *testing.T) {
        db, err := OpenSQLite(":memory:")
        require.NoError(t, err)
        defer db.Close()

        err = InitSchema(db)
        require.NoError(t, err)

        // 验证所有表存在
        tables := []string{
            "connections", "templates", "tasks", "runs",
            "metric_samples", "run_logs", "settings", "reports",
        }

        for _, table := range tables {
            var count int
            err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&count)
            require.NoError(t, err, "table %s should exist", table)
            require.Equal(t, 1, count, "table %s should exist", table)
        }
    })

    t.Run("creates indexes", func(t *testing.T) {
        db, err := OpenSQLite(":memory:")
        require.NoError(t, err)
        defer db.Close()

        err = InitSchema(db)
        require.NoError(t, err)

        // 验证索引存在
        var count int
        err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='index'").Scan(&count)
        require.NoError(t, err)
        require.Greater(t, count, 0, "should have indexes")
    })

    t.Run("enables foreign keys", func(t *testing.T) {
        db, err := OpenSQLite(":memory:")
        require.NoError(t, err)
        defer db.Close()

        err = InitSchema(db)
        require.NoError(t, err)

        // 验证外键开启
        var fkEnabled int
        err = db.QueryRow("PRAGMA foreign_keys").Scan(&fkEnabled)
        require.NoError(t, err)
        require.Equal(t, 1, fkEnabled, "foreign keys should be enabled")
    })
}
```

---

#### Task 1.9: 实现 SQLite 初始化函数
**Type**: impl
**File**: `internal/infra/database/sqlite.go`
**Depends**: 1.8
**Description**: 实现 SQLite 数据库初始化函数，包含连接池配置（TDD: Green）
**Acceptance**:
- 函数签名正确
- 使用 WAL 模式
- 单连接池配置
- 通过 TestInitSchema 所有测试
**Implementation**:
```go
package database

import (
    "database/sql"
    "embed"
    "fmt"
    "time"

    _ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaFS embed.FS

// OpenSQLite 打开 SQLite 数据库连接
func OpenSQLite(path string) (*sql.DB, error) {
    // DSN with WAL mode, normal sync, busy timeout
    dsn := fmt.Sprintf("file:%s?_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)", path)

    db, err := sql.Open("sqlite", dsn)
    if err != nil {
        return nil, fmt.Errorf("open sqlite: %w", err)
    }

    // SQLite 推荐：单连接模式
    db.SetMaxOpenConns(1)
    db.SetMaxIdleConns(1)
    db.SetConnMaxLifetime(time.Hour)

    // 验证连接
    if err := db.Ping(); err != nil {
        return nil, fmt.Errorf("ping database: %w", err)
    }

    return db, nil
}

// InitSchema 初始化数据库 Schema
func InitSchema(db *sql.DB) error {
    schema, err := schemaFS.ReadFile("schema.sql")
    if err != nil {
        return fmt.Errorf("read schema: %w", err)
    }

    _, err = db.Exec(string(schema))
    if err != nil {
        return fmt.Errorf("exec schema: %w", err)
    }

    return nil
}
```

---

#### Task 1.10: 创建 README.md [P]
**Type**: impl
**File**: `README.md`
**Depends**: 1.1
**Description**: 创建项目 README
**Acceptance**:
- README 创建成功
- 包含项目简介、安装说明、快速开始
**Content**:
```markdown
# DB-BenchMind

数据库压测工作台 - 统一的 GUI 界面来编排和运行 Sysbench、Swingbench、HammerDB。

## 功能特性

- 支持 4 种数据库：MySQL, Oracle, SQL Server, PostgreSQL
- 支持 3 种压测工具：Sysbench, Swingbench, HammerDB
- 实时监控测试过程
- 自动存储和归档结果
- 多格式报告导出（MD/HTML/JSON/PDF）
- 结果对比分析

## 快速开始

\`\`\`bash
# 克隆仓库
git clone https://github.com/whhaicheng/DB-BenchMind.git
cd DB-BenchMind

# 安装依赖
go mod download

# 构建
make build

# 运行
./build/db-benchmind
\`\`\`

## 开发

\`\`\`bash
# 运行测试
make test

# 代码检查
make check

# 格式化代码
gofmt -w .
\`\`\`

## 许可证

MIT License
```

---

## Phase 2: 连接管理（4种数据库）

**目标**: 实现连接领域模型和仓储，支持 MySQL/Oracle/SQL Server/PostgreSQL

---

#### Task 2.1: [测试] Connection 接口定义
**Type**: test
**File**: `internal/domain/connection/connection_test.go`
**Depends**: 无
**Description**: 定义 Connection 接口（先写测试用例验证接口设计）
**Acceptance**:
- 测试定义了 Connection 接口的所有方法
- 测试编译通过（但实现不存在）
**Content**:
```go
package connection

import (
    "context"
    "testing"
)

// TestConnectionInterface 确保 Connection 接口包含所有必需方法
func TestConnectionInterface(t *testing.T) {
    type testConn struct{}

    var _ Connection = (*testConn)(nil)
}

// 实现测试用的最小接口
type testConn struct{}

func (t *testConn) GetID() string                                                      { return "" }
func (t *testConn) GetName() string                                                    { return "" }
func (t *testConn) SetName(name string)                                               {}
func (t *testConn) GetType() DatabaseType                                             { return DatabaseTypeMySQL }
func (t *testConn) Validate() error                                                   { return nil }
func (t *testConn) Test(ctx context.Context) (*TestResult, error)                   { return nil, nil }
func (t *testConn) GetDSN() string                                                     { return "" }
func (t *testConn) GetDSNWithPassword() string                                         { return "" }
func (t *testConn) Redact() string                                                   { return "" }
func (t *testConn) ToJSON() ([]byte, error)                                         { return nil, nil }
```

---

#### Task 2.2: 实现 Connection 接口定义
**Type**: impl
**File**: `internal/domain/connection/connection.go`
**Depends**: 2.1
**Description**: 定义 Connection 接口、DatabaseType、TestResult（TDD: Green）
**Acceptance**:
- 接口定义完整
- 包含所有必需方法
- 通过 TestConnectionInterface 测试
**Implementation**:
```go
package connection

import (
    "context"
    "encoding/json"
)

// DatabaseType 数据库类型
type DatabaseType string

const (
    DatabaseTypeMySQL      DatabaseType = "mysql"
    DatabaseTypeOracle     DatabaseType = "oracle"
    DatabaseTypeSQLServer  DatabaseType = "sqlserver"
    DatabaseTypePostgreSQL DatabaseType = "postgresql"
)

// Connection 连接接口
type Connection interface {
    // 基本信息
    GetID() string
    GetName() string
    SetName(name string)
    GetType() DatabaseType

    // 验证与测试
    Validate() error
    Test(ctx context.Context) (*TestResult, error)

    // 连接字符串
    GetDSN() string               // 不含密码
    GetDSNWithPassword() string   // 含密码

    // 脱敏与序列化
    Redact() string               // 脱敏信息
    ToJSON() ([]byte, error)      // JSON序列化（不含密码）
}

// TestResult 连接测试结果
type TestResult struct {
    Success         bool    `json:"success"`
    LatencyMs       int64   `json:"latency_ms"`
    DatabaseVersion string  `json:"database_version,omitempty"`
    Error           string  `json:"error,omitempty"`
}
```

---

#### Task 2.3: [测试] MySQLConnection - Validate 方法
**Type**: test
**File**: `internal/domain/connection/mysql_test.go`
**Depends**: 无
**Description**: 表格驱动测试 MySQLConnection.Validate() 方法（TDD: Red）
**Acceptance**:
- 测试有效连接通过
- 测试无效端口（负数、超大）失败
- 测试缺失必填字段失败
- 测试包含清晰错误消息
- 所有测试失败（实现不存在）
**Content**:
```go
package connection

import (
    "testing"
)

func TestMySQLConnection_Validate(t *testing.T) {
    tests := []struct {
        name    string
        conn    *MySQLConnection
        wantErr bool
        errMsg  string
    }{
        {
            name: "valid connection",
            conn: &MySQLConnection{
                ID:   "test-1",
                Name: "test-conn",
                Host: "localhost",
                Port: 3306,
                Database: "testdb",
                Username: "root",
                Password: "pass",
            },
            wantErr: false,
        },
        {
            name: "invalid port - negative",
            conn: &MySQLConnection{
                Name: "test", Host: "localhost", Port: -1,
                Database: "testdb", Username: "root",
            },
            wantErr: true,
            errMsg:  "port must be between 1 and 65535",
        },
        {
            name: "invalid port - too large",
            conn: &MySQLConnection{
                Name: "test", Host: "localhost", Port: 99999,
                Database: "testdb", Username: "root",
            },
            wantErr: true,
            errMsg:  "port must be between 1 and 65535",
        },
        {
            name: "missing name",
            conn: &MySQLConnection{
                Name: "", Host: "localhost", Port: 3306,
                Database: "testdb", Username: "root",
            },
            wantErr: true,
            errMsg:  "name is required",
        },
        {
            name: "missing host",
            conn: &MySQLConnection{
                Name: "test", Host: "", Port: 3306,
                Database: "testdb", Username: "root",
            },
            wantErr: true,
            errMsg:  "host is required",
        },
        {
            name: "missing database",
            conn: &MySQLConnection{
                Name: "test", Host: "localhost", Port: 3306,
                Database: "", Username: "root",
            },
            wantErr: true,
            errMsg:  "database is required",
        },
        {
            name: "missing username",
            conn: &MySQLConnection{
                Name: "test", Host: "localhost", Port: 3306,
                Database: "testdb", Username: "",
            },
            wantErr: true,
            errMsg:  "username is required",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.conn.Validate()
            if (err != nil) != tt.wantErr {
                t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if tt.wantErr && tt.errMsg != "" && err != nil {
                if err.Error() != tt.errMsg && !containsString(err.Error(), tt.errMsg) {
                    t.Errorf("Validate() error = %v, want contain %v", err.Error(), tt.errMsg)
                }
            }
        })
    }
}

func containsString(s, substr string) bool {
    return len(s) >= len(substr) && (s == substr || len(substr) == 0 || contains(s, substr))
}
```

---

#### Task 2.4: 实现 MySQLConnection 结构和 Validate 方法
**Type**: impl
**File**: `internal/domain/connection/mysql.go`
**Depends**: 2.2, 2.3
**Description**: 实现 MySQLConnection 结构和 Validate 方法（TDD: Green）
**Acceptance**:
- 包含所有字段（ID, Name, Host, Port, Database, Username, Password, SSLMode）
- Password 字段 `json:"-"` 标签
- Validate 方法验证所有规则
- 通过 Task 2.3 所有测试用例
**Implementation**:
```go
package connection

import (
    "context"
    "database/sql"
    "fmt"
    "strings"
    "time"
    _ "modernc.org/sqlite" // 仅用于类型检查
)

// MySQLConnection MySQL 连接配置
type MySQLConnection struct {
    // 基础字段
    ID   string `json:"id"`
    Name string `json:"name"`

    // 连接参数
    Host     string `json:"host"`
    Port     int    `json:"port"`
    Database string `json:"database"`
    Username string `json:"username"`
    Password string `json:"-"` // 不序列化到 JSON

    // SSL 配置
    SSLMode string `json:"ssl_mode"`

    // 元数据
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

// 实现 Connection 接口
func (c *MySQLConnection) GetID() string { return c.ID }
func (c *MySQLConnection) GetName() string { return c.Name }
func (c *MySQLConnection) SetName(name string) { c.Name = name }
func (c *MySQLConnection) GetType() DatabaseType { return DatabaseTypeMySQL }

// Validate 验证连接参数
func (c *MySQLConnection) Validate() error {
    var errs []string

    if c.Name == "" {
        errs = append(errs, "name is required")
    }
    if c.Host == "" {
        errs = append(errs, "host is required")
    }
    if c.Port < 1 || c.Port > 65535 {
        errs = append(errs, fmt.Sprintf("port must be between 1 and 65535, got %d", c.Port))
    }
    if c.Database == "" {
        errs = append(errs, "database is required")
    }
    if c.Username == "" {
        errs = append(errs, "username is required")
    }

    if len(errs) > 0 {
        return fmt.Errorf("validation failed: %s", strings.Join(errs, "; "))
    }
    return nil
}

// GetDSN 生成连接字符串（不含密码）
func (c *MySQLConnection) GetDSN() string {
    return fmt.Sprintf("%s@tcp(%s:%d)/%s", c.Username, c.Host, c.Port, c.Database)
}

// GetDSNWithPassword 生成完整连接字符串（含密码）
func (c *MySQLConnection) GetDSNWithPassword() string {
    return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", c.Username, c.Password, c.Host, c.Port, c.Database)
}

// Redact 返回脱敏后的连接信息
func (c *MySQLConnection) Redact() string {
    return fmt.Sprintf("%s (***@%s:%d/%s)", c.Name, c.Host, c.Port, c.Database)
}

// Test 测试连接（使用 database/sql.Ping）
func (c *MySQLConnection) Test(ctx context.Context) (*TestResult, error) {
    start := time.Now()

    dsn := c.GetDSNWithPassword()
    db, err := sql.Open("mysql", dsn)
    if err != nil {
        return &TestResult{
            Success: false,
            Error:   fmt.Sprintf("failed to open connection: %v", err),
        }, nil
    }
    defer db.Close()

    // 设置超时
    ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
    defer cancel()

    err = db.PingContext(ctx)
    latency := time.Since(start).Milliseconds()

    if err != nil {
        return &TestResult{
            Success:   false,
            LatencyMs: latency,
            Error:     fmt.Sprintf("connection failed: %v", err),
        }, nil
    }

    // 获取数据库版本
    var version string
    err = db.QueryRowContext(ctx, "SELECT VERSION()").Scan(&version)
    if err != nil {
        version = "unknown"
    }

    return &TestResult{
        Success:         true,
        LatencyMs:       latency,
        DatabaseVersion: version,
    }, nil
}

// ToJSON 序列化为 JSON（不含密码）
func (c *MySQLConnection) ToJSON() ([]byte, error) {
    type Alias MySQLConnection
    return json.Marshal(&struct {
        *Alias
    }{
        Alias: (*Alias)(c),
    })
}
```

---

#### Task 2.5: [测试] MySQLConnection - Test 方法
**Type**: test
**File**: `internal/domain/connection/mysql_test.go`
**Depends**: 2.4
**Description**: 测试 MySQLConnection.Test() 方法
**Acceptance**:
- 测试成功场景（需要 fake 或真实 MySQL）
- 测试失败场景
- 测试超时场景
- 测试返回数据库版本
- 测试返回正确延迟

---

#### Task 2.6: 实现 MySQLConnection 其他方法
**Type**: impl
**File**: `internal/domain/connection/mysql.go`
**Depends**: 2.5
**Description**: 完善 MySQLConnection 的 Test 和其他方法
**Acceptance**:
- 通过 Task 2.5 所有测试

---

### Task 2.7-2.18: Oracle/SQL Server/PostgreSQL 连接

**重复 Task 2.1-2.6 模式**，针对其他三种数据库类型：

#### Task 2.7-2.10: OracleConnection
- Task 2.7: [测试] OracleConnection 接口测试
- Task 2.8: 实现 OracleConnection 结构定义
- Task 2.9: [测试] Validate 方法
- Task 2.10: 实现 Validate 和其他方法

#### Task 2.11-2.14: SQLServerConnection
- Task 2.11: [测试] SQLServerConnection 接口测试
- Task 2.12: 实现 SQLServerConnection 结构定义
- Task 2.13: [测试] Validate 方法
- Task 2.14: 实现 Validate 和其他方法

#### Task 2.15-2.18: PostgreSQLConnection
- Task 2.15: [测试] PostgreSQLConnection 接口测试
- Task 2.16: 实现 PostgreSQLConnection 结构定义
- Task 2.17: [测试] Validate 方法
- Task 2.18: 实现 Validate 和其他方法

---

#### Task 2.19: [测试] ConnectionRepository 接口
**Type**: test
**File**: `internal/app/usecase/repository_test.go`
**Depends**: 无
**Description**: 定义并测试 ConnectionRepository 接口需求
**Acceptance**:
- 接口定义包含所有方法
- 测试验证接口可用性

---

#### Task 2.20: 定义 ConnectionRepository 接口
**Type**: impl
**File**: `internal/app/usecase/repository.go`
**Depends**: 2.19
**Description**: 定义 ConnectionRepository 接口
**Acceptance**:
- 接口包含 Save, FindByID, FindAll, Delete, ExistsByName 方法

---

#### Task 2.21: [测试] SQLiteConnectionRepository - Save 方法
**Type**: test
**File**: `internal/infra/database/repository/connection_repo_test.go`
**Depends**: 2.20, 1.9
**Description**: 测试 Save 方法（使用内存 SQLite）
**Acceptance**:
- 测试保存成功
- 测试保存重复连接（唯一约束）
- 测试密码不序列化到数据库
- 所有测试失败（实现不存在）

---

#### Task 2.22: 实现 SQLiteConnectionRepository - Save 方法
**Type**: impl
**File**: `internal/infra/database/repository/connection_repo.go`
**Depends**: 2.21
**Description**: 实现 Save 方法
**Acceptance**:
- 序列化连接配置（不含密码）
- 插入数据库
- 返回适当错误
- 通过 Task 2.21 所有测试

---

#### Task 2.23-2.30: ConnectionRepository 其他方法

按照 TDD 模式实现：
- Task 2.23-2.24: FindByID 方法
- Task 2.25-2.26: FindAll 方法
- Task 2.27-2.28: Delete 方法
- Task 2.29-2.30: ExistsByName 方法

---

#### Task 2.31-2.40: Keyring 密钥管理

按照 TDD 模式实现：
- Task 2.31-2.32: KeyringProvider 接口定义和测试
- Task 2.33-2.34: GoKeyringProvider 实现
- Task 2.35-2.36: EncryptedFileProvider 降级方案
- Task 2.37-2.38: 密码序列化/反序列化
- Task 2.39-2.40: 集成测试

---

#### Task 2.41-2.60: ConnectionUseCase 和 GUI

按照 TDD 模式实现：
- Task 2.41-2.42: ConnectionUseCase 接口定义
- Task 2.43-2.50: ConnectionUseCase 所有方法实现
- Task 2.51-2.60: GUI 连接管理页面

---

## Phase 3: 模板系统与任务配置

**目标**: 实现模板管理、内置模板、任务配置

---

### Task 3.1-3.50: 详细任务列表

（由于篇幅限制，这里提供关键任务的示例）

#### Task 3.1: [测试] Template 结构
**Type**: test
**File**: `internal/domain/template/template_test.go`
**Depends**: 无
**Description**: 测试 Template 结构和 Validate 方法

#### Task 3.2: 实现 Template 结构
**Type**: impl
**File**: `internal/domain/template/template.go`
**Depends**: 3.1
**Description**: 实现 Template 结构

#### Task 3.11-3.18: 7个内置模板
**Type**: impl
**File**: `contracts/templates/*.json`
**Depends**: 3.2
**Description**: 创建 7 个内置模板 JSON 文件

---

## Phase 4: 工具适配器与执行编排

**目标**: 实现三个工具的适配器和执行编排器

---

### Task 4.1-4.130: 详细任务列表

#### Task 4.1-4.10: pkg/benchmark 适配器接口
#### Task 4.11-4.30: Sysbench 适配器
#### Task 4.31-4.50: Swingbench 适配器
#### Task 4.51-4.70: HammerDB 适配器
#### Task 4.71-4.80: 适配器注册表
#### Task 4.81-4.90: RunState 状态机
#### Task 4.91-4.100: Executor 执行编排器
#### Task 4.101-4.110: BenchmarkUseCase
#### Task 4.111-4.120: RunRepository
#### Task 4.121-4.130: GUI 运行监控页面

---

## Phase 5: 结果存储与历史记录

**目标**: 实现结果持久化和历史记录查询

---

### Task 5.1-5.60: 详细任务列表

---

## Phase 6: 报告生成与导出

**目标**: 实现多格式报告导出

---

### Task 6.1-6.40: 详细任务列表

---

## Phase 7: 结果对比功能

**目标**: 实现完整的结果对比系统

---

### Task 7.1-7.50: 详细任务列表

---

## Phase 8: 设置与文档完善

**目标**: 完善设置功能和文档

---

### Task 8.1-8.45: 详细任务列表

---

## 附录

### A. 并行任务索引

Phase 1 可并行执行的任务（所有标记 [P]）：
- Task 1.1: 创建项目目录结构
- Task 1.2: 初始化 go.mod
- Task 1.3: 创建 Makefile
- Task 1.4: 配置 golangci-lint
- Task 1.5: 创建 .gitignore
- Task 1.6: 创建产品定义文档
- Task 1.10: 创建 README.md

Phase 2 可并行执行的任务组：
- Task 2.7-2.10: OracleConnection
- Task 2.11-2.14: SQLServerConnection
- Task 2.15-2.18: PostgreSQLConnection

### B. 关键里程碑

- **M1**: 项目初始化完成（Task 1.9）
- **M2**: 连接领域完成（Task 2.18）
- **M3**: 连接管理完成（Task 2.70）
- **M4**: 模板系统完成（Task 3.50）
- **M5**: 适配器完成（Task 4.80）
- **M6**: 执行编排完成（Task 4.130）
- **M7**: MVP 可用（Task 6.40）
- **M8**: 完整功能（Task 8.45）

### C. TDD 检查清单

每个功能点必须遵循：
1. ✅ 先写测试（Task Type: test）
2. ✅ 确认测试失败
3. ✅ 编写实现（Task Type: impl）
4. ✅ 确认测试通过
5. ✅ 重构优化

### D. 文件命名规范

- 测试文件：`{filename}_test.go`
- 接口文件：`{filename}.go` 或 `{package}.go`
- 实现文件：遵循 Go 命名约定

---

**文档结束**
