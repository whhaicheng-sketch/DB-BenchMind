# DB-BenchMind API 参考文档

**版本**: 1.0.0
**更新日期**: 2026-01-28

---

## 目录

1. [概述](#概述)
2. [Domain 层](#domain-层)
3. [Application 层](#application-层)
4. [Infrastructure 层](#infrastructure-层)
5. [Transport 层](#transport-层)
6. [类型定义](#类型定义)

---

## 概述

DB-BenchMind 采用 Clean Architecture + DDD 设计，分层清晰：

```
domain/          (核心业务逻辑，无外部依赖)
app/usecase/     (用例编排，定义接口)
infra/           (外部依赖实现)
transport/ui/    (GUI 界面)
```

---

## Domain 层

### connection.Connection

数据库连接接口。

```go
package connection

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
```

#### MySQLConnection

```go
type MySQLConnection struct {
    ID        string    `json:"id"`
    Name      string    `json:"name"`
    Host      string    `json:"host"`
    Port      int       `json:"port"`
    Database  string    `json:"database"`
    Username  string    `json:"username"`
    Password  string    `json:"-"`           // 不序列化
    SSLMode   string    `json:"ssl_mode"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

**Validate() 验证规则**:
- Name: 非空
- Host: 非空
- Port: 1-65535
- Database: 非空
- Username: 非空

#### OracleConnection

```go
type OracleConnection struct {
    ID          string    `json:"id"`
    Name        string    `json:"name"`
    Host        string    `json:"host"`
    Port        int       `json:"port"`
    SID         string    `json:"sid"`
    ServiceName string    `json:"service_name"`
    Username    string    `json:"username"`
    Password    string    `json:"-"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}
```

**Validate() 验证规则**:
- Name: 非空
- Host: 非空
- Port: 1-65535
- SID 或 ServiceName: 至少一个非空
- Username: 非空

#### SQLServerConnection

```go
type SQLServerConnection struct {
    ID        string    `json:"id"`
    Name      string    `json:"name"`
    Host      string    `json:"host"`
    Port      int       `json:"port"`
    Database  string    `json:"database"`
    Username  string    `json:"username"`
    Password  string    `json:"-"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

#### PostgreSQLConnection

```go
type PostgreSQLConnection struct {
    ID         string    `json:"id"`
    Name       string    `json:"name"`
    Host       string    `json:"host"`
    Port       int       `json:"port"`
    Database   string    `json:"database"`
    Username   string    `json:"username"`
    Password   string    `json:"-"`
    SSLMode    string    `json:"ssl_mode"`
    CreatedAt  time.Time `json:"created_at"`
    UpdatedAt  time.Time `json:"updated_at"`
}
```

---

### template.Template

基准测试模板定义。

```go
package template

type Template struct {
    ID            string                 `json:"id"`
    Name          string                 `json:"name"`
    Description   string                 `json:"description"`
    Tool          string                 `json:"tool"`           // sysbench, swingbench, hammerdb
    DatabaseTypes []DatabaseType         `json:"database_types"`
    BenchmarkType string                 `json:"benchmark_type"`
    Parameters    map[string]interface{} `json:"parameters"`
    Options       map[string]interface{} `json:"options"`
    IsBuiltin     bool                   `json:"is_builtin"`
    CreatedAt     time.Time              `json:"created_at"`
    UpdatedAt     time.Time              `json:"updated_at"`
}

func (t *Template) Validate() error
func (t *Template) SupportsDatabase(dbType DatabaseType) bool
func (t *Template) GetParameter(key string) (interface{}, bool)
func (t *Template) MergeParameters(params map[string]interface{}) map[string]interface{}
```

**Validate() 验证规则**:
- Name: 非空
- Tool: 必须是 sysbench/swingbench/hammerdb 之一
- DatabaseTypes: 非空
- BenchmarkType: 非空

---

### execution.Task

基准测试任务定义。

```go
package execution

type Task struct {
    ID           string                 `json:"id"`
    Name         string                 `json:"name"`
    ConnectionID string                 `json:"connection_id"`
    TemplateID   string                 `json:"template_id"`
    Parameters   map[string]interface{} `json:"parameters"`
    Options      map[string]interface{} `json:"options"`
    Tags         []string               `json:"tags"`
    CreatedAt    time.Time              `json:"created_at"`
}

func (t *Task) Validate() error
```

---

### execution.Run

基准测试运行实例。

```go
package execution

type RunState string

const (
    StatePending   RunState = "pending"
    StateRunning   RunState = "running"
    StateCompleted RunState = "completed"
    StateFailed    RunState = "failed"
    StateCancelled RunState = "cancelled"
)

type Run struct {
    ID            string          `json:"id"`
    TaskID        string          `json:"task_id"`
    State         RunState        `json:"state"`
    CreatedAt     time.Time       `json:"created_at"`
    StartedAt     *time.Time      `json:"started_at,omitempty"`
    CompletedAt   *time.Time      `json:"completed_at,omitempty"`
    DurationSeconds float64        `json:"duration_seconds,omitempty"`
    Result        *BenchmarkResult `json:"result,omitempty"`
    ErrorMessage  string          `json:"error_message,omitempty"`
    WorkDir       string          `json:"work_dir,omitempty"`
}
```

---

### execution.BenchmarkResult

基准测试结果。

```go
type BenchmarkResult struct {
    // 基本指标
    TPSCalculated float64 `json:"tps_calculated"`
    QPS           float64 `json:"qps,omitempty"`

    // 延迟指标（毫秒）
    LatencyAvg float64 `json:"latency_avg_ms"`
    LatencyMin float64 `json:"latency_min_ms"`
    LatencyMax float64 `json:"latency_max_ms"`
    LatencyP95 float64 `json:"latency_p95_ms"`
    LatencyP99 float64 `json:"latency_p99_ms"`

    // 事务统计
    TotalTransactions    int64   `json:"total_transactions"`
    FailedTransactions   int64   `json:"failed_transactions"`
    IgnoredTransactions  int64   `json:"ignored_transactions,omitempty"`
    TotalQueries         int64   `json:"total_queries,omitempty"`
    FailedQueries        int64   `json:"failed_queries,omitempty"`

    // 错误率
    ErrorRate float64 `json:"error_rate"`

    // 其他指标
    TotalBytesRead   int64   `json:"total_bytes_read,omitempty"`
    TotalBytesWrite  int64   `json:"total_bytes_written,omitempty"`

    // 工具特定指标
    RawMetrics map[string]interface{} `json:"raw_metrics,omitempty"`

    // 采样数据
    Samples []MetricSample `json:"samples,omitempty"`
}

type MetricSample struct {
    Timestamp time.Time `json:"timestamp"`
    Phase     string    `json:"phase"`
    TPS       float64   `json:"tps"`
    LatencyAvg float64  `json:"latency_avg_ms"`
}
```

---

### comparison.Comparison

结果对比定义。

```go
package comparison

type ComparisonType string

const (
    ComparisonTypeBaseline ComparisonType = "baseline"
    ComparisonTypeTrend    ComparisonType = "trend"
    ComparisonTypeMulti    ComparisonType = "multi"
)

type Comparison struct {
    ID         string            `json:"id"`
    Name       string            `json:"name"`
    Type       ComparisonType    `json:"type"`
    RunIDs     []string          `json:"run_ids"`
    BaselineID string            `json:"baseline_id,omitempty"`
    CreatedAt  time.Time         `json:"created_at"`
    Metadata   map[string]string `json:"metadata,omitempty"`
}

type ComparisonResult struct {
    Comparison *Comparison      `json:"comparison"`
    Runs       []*execution.Run `json:"runs"`
    Metrics    *MetricDiff      `json:"metrics"`
    Summary    *ComparisonSummary `json:"summary"`
    CreatedAt  time.Time        `json:"created_at"`
}

type MetricDiff struct {
    TPSDiff        []MetricValueDiff `json:"tps_diff"`
    LatencyAvgDiff []MetricValueDiff `json:"latency_avg_diff"`
    LatencyP95Diff []MetricValueDiff `json:"latency_p95_diff"`
    ErrorRateDiff  []MetricValueDiff `json:"error_rate_diff"`
    BestTps        *MetricStats      `json:"best_tps,omitempty"`
    WorstTps       *MetricStats      `json:"worst_tps,omitempty"`
}
```

---

## Application 层

### usecase.ConnectionUseCase

连接管理业务逻辑。

```go
package usecase

type ConnectionUseCase struct {
    // 私有字段
}

func NewConnectionUseCase(
    repo ConnectionRepository,
    keyring KeyringProvider,
) *ConnectionUseCase

// 创建连接
func (uc *ConnectionUseCase) CreateConnection(
    ctx context.Context,
    conn connection.Connection,
) error

// 更新连接
func (uc *ConnectionUseCase) UpdateConnection(
    ctx context.Context,
    conn connection.Connection,
) error

// 删除连接
func (uc *ConnectionUseCase) DeleteConnection(
    ctx context.Context,
    id string,
) error

// 获取连接
func (uc *ConnectionUseCase) GetConnection(
    ctx context.Context,
    id string,
) (connection.Connection, error)

// 列出所有连接
func (uc *ConnectionUseCase) ListConnections(
    ctx context.Context,
) ([]connection.Connection, error)

// 测试连接
func (uc *ConnectionUseCase) TestConnection(
    ctx context.Context,
    id string,
) (*connection.TestResult, error)

// 检查名称是否存在
func (uc *ConnectionUseCase) ExistsByName(
    ctx context.Context,
    name string,
) (bool, error)
```

**错误类型**:
- `ErrConnectionNotFound`: 连接不存在
- `ErrConnectionAlreadyExists`: 连接名称已存在
- `ErrInvalidConnection`: 连接配置无效

---

### usecase.TemplateUseCase

模板管理业务逻辑。

```go
package usecase

type TemplateUseCase struct {
    // 私有字段
}

func NewTemplateUseCase(repo TemplateRepository) *TemplateUseCase

// 创建模板
func (uc *TemplateUseCase) CreateTemplate(
    ctx context.Context,
    tmpl *template.Template,
) error

// 更新模板
func (uc *TemplateUseCase) UpdateTemplate(
    ctx context.Context,
    tmpl *template.Template,
) error

// 删除模板
func (uc *TemplateUseCase) DeleteTemplate(
    ctx context.Context,
    id string,
) error

// 获取模板
func (uc *TemplateUseCase) GetTemplate(
    ctx context.Context,
    id string,
) (*template.Template, error)

// 列出所有模板
func (uc *TemplateUseCase) ListTemplates(
    ctx context.Context,
) ([]*template.Template, error)

// 列出内置模板
func (uc *TemplateUseCase) ListBuiltinTemplates(
    ctx context.Context,
) ([]*template.Template, error)

// 列出自定义模板
func (uc *TemplateUseCase) ListCustomTemplates(
    ctx context.Context,
) ([]*template.Template, error)
```

---

### usecase.BenchmarkUseCase

基准测试执行编排。

```go
package usecase

type BenchmarkUseCase struct {
    // 私有字段
}

func NewBenchmarkUseCase(
    runRepo RunRepository,
    registry AdapterRegistry,
    keyring KeyringProvider,
) *BenchmarkUseCase

// 执行任务
func (uc *BenchmarkUseCase) ExecuteTask(
    ctx context.Context,
    task *execution.Task,
) (*execution.Run, error)

// 取消运行
func (uc *BenchmarkUseCase) CancelRun(
    ctx context.Context,
    runID string,
) error

// 获取运行状态
func (uc *BenchmarkUseCase) GetRunStatus(
    ctx context.Context,
    runID string,
) (*RunStatus, error)

// 获取运行结果
func (uc *BenchmarkUseCase) GetRunResult(
    ctx context.Context,
    runID string,
) (*execution.BenchmarkResult, error)

// 获取运行日志
func (uc *BenchmarkUseCase) GetRunLogs(
    ctx context.Context,
    runID string,
    offset int,
    limit int,
) ([]LogEntry, error)

// 获取实时指标
func (uc *BenchmarkUseCase) GetRunMetrics(
    ctx context.Context,
    runID string,
) ([]execution.MetricSample, error)
```

**RunStatus 结构**:
```go
type RunStatus struct {
    RunID     string        `json:"run_id"`
    State     execution.RunState `json:"state"`
    Progress  float64       `json:"progress"`
    StartedAt *time.Time    `json:"started_at,omitempty"`
    ETA       *time.Time    `json:"eta,omitempty"`
}
```

**LogEntry 结构**:
```go
type LogEntry struct {
    Timestamp time.Time `json:"timestamp"`
    Stream    string    `json:"stream"`    // stdout or stderr
    Content   string    `json:"content"`
}
```

---

### usecase.ReportUseCase

报告生成。

```go
package usecase

type ReportUseCase struct {
    // 私有字段
}

func NewReportUseCase(
    repo ReportRepository,
    runRepo RunRepository,
    generators map[string]ReportGenerator,
) *ReportUseCase

// 生成报告
func (uc *ReportUseCase) GenerateReport(
    ctx context.Context,
    runID string,
    format string,
) (string, error)

// 批量生成报告
func (uc *ReportUseCase) GenerateReports(
    ctx context.Context,
    runIDs []string,
    format string,
) ([]string, error)

// 获取报告
func (uc *ReportUseCase) GetReport(
    ctx context.Context,
    runID string,
    format string,
) ([]byte, error)

// 列出报告
func (uc *ReportUseCase) ListReports(
    ctx context.Context,
    runID string,
) ([]*report.Report, error)
```

**支持的格式**:
- `markdown`: .md 文件
- `html`: .html 文件
- `json`: .json 文件
- `pdf`: .pdf 文件（需要 pandoc）

---

### usecase.ComparisonUseCase

结果对比分析。

```go
package usecase

type ComparisonUseCase struct {
    // 私有字段
}

func NewComparisonUseCase(runRepo RunRepository) *ComparisonUseCase

// 对比多次运行
func (uc *ComparisonUseCase) CompareRuns(
    ctx context.Context,
    runIDs []string,
    baselineID string,
    compType comparison.ComparisonType,
) (*comparison.ComparisonResult, error)

// 对比最近的 N 次运行
func (uc *ComparisonUseCase) CompareRecentRuns(
    ctx context.Context,
    count int,
    baselineID string,
) (*comparison.ComparisonResult, error)

// 创建保存的对比
func (uc *ComparisonUseCase) CreateComparison(
    ctx context.Context,
    name string,
    runIDs []string,
    baselineID string,
    compType comparison.ComparisonType,
) (*comparison.Comparison, error)

// 获取趋势分析
func (uc *ComparisonUseCase) GetTrendAnalysis(
    ctx context.Context,
    runIDs []string,
    metric string,
) (*TrendAnalysis, error)
```

**TrendAnalysis 结构**:
```go
type TrendAnalysis struct {
    Metric    string       `json:"metric"`
    Values    []TrendValue `json:"values"`
    Trend     string       `json:"trend"`     // "increasing", "decreasing", "stable"
    ChangePct float64      `json:"change_pct"`
    MinValue  float64      `json:"min_value"`
    MaxValue  float64      `json:"max_value"`
    AvgValue  float64      `json:"avg_value"`
}
```

---

### usecase.SettingsUseCase

系统设置管理。

```go
package usecase

type SettingsUseCase struct {
    // 私有字段
}

func NewSettingsUseCase(
    repo SettingsRepository,
    detector ToolDetector,
) *SettingsUseCase

// 检测工具
func (uc *SettingsUseCase) DetectTools(
    ctx context.Context,
) map[string]ToolInfo

// 获取设置
func (uc *SettingsUseCase) GetSetting(
    ctx context.Context,
    key string,
) (string, error)

// 设置值
func (uc *SettingsUseCase) SetSetting(
    ctx context.Context,
    key string,
    value string,
) error

// 获取所有设置
func (uc *SettingsUseCase) GetAllSettings(
    ctx context.Context,
) (map[string]string, error)
```

**ToolInfo 结构**:
```go
type ToolInfo struct {
    Found   bool   `json:"found"`
    Path    string `json:"path,omitempty"`
    Version string `json:"version,omitempty"`
}
```

---

## Infrastructure 层

### database.InitializeSQLite

初始化 SQLite 数据库。

```go
package database

func InitializeSQLite(
    ctx context.Context,
    path string,
) (*sql.DB, error)
```

**返回**:
- `*sql.DB`: 数据库连接
- `error`: 初始化错误

**配置**:
- WAL 模式
- 外键约束启用
- 单连接池

---

### repository.SQLiteConnectionRepository

连接仓储实现。

```go
package repository

func NewSQLiteConnectionRepository(db *sql.DB) *SQLiteConnectionRepository

func (r *SQLiteConnectionRepository) Save(
    ctx context.Context,
    conn connection.Connection,
) error

func (r *SQLiteConnectionRepository) FindByID(
    ctx context.Context,
    id string,
) (connection.Connection, error)

func (r *SQLiteConnectionRepository) FindAll(
    ctx context.Context,
    opts FindOptions,
) ([]connection.Connection, error)

func (r *SQLiteConnectionRepository) Delete(
    ctx context.Context,
    id string,
) error

func (r *SQLiteConnectionRepository) ExistsByName(
    ctx context.Context,
    name string,
) (bool, error)
```

---

### repository.SQLiteRunRepository

运行仓储实现。

```go
package repository

func NewSQLiteRunRepository(db *sql.DB) *SQLiteRunRepository

func (r *SQLiteRunRepository) Save(
    ctx context.Context,
    run *execution.Run,
) error

func (r *SQLiteRunRepository) FindByID(
    ctx context.Context,
    id string,
) (*execution.Run, error)

func (r *SQLiteRunRepository) FindAll(
    ctx context.Context,
    opts FindOptions,
) ([]*execution.Run, error)

func (r *SQLiteRunRepository) FindByConnection(
    ctx context.Context,
    connID string,
    opts FindOptions,
) ([]*execution.Run, error)

func (r *SQLiteRunRepository) FindByState(
    ctx context.Context,
    state execution.RunState,
    opts FindOptions,
) ([]*execution.Run, error)

func (r *SQLiteRunRepository) Update(
    ctx context.Context,
    run *execution.Run,
) error
```

---

### adapter.SysbenchAdapter

Sysbench 工具适配器。

```go
package adapter

type SysbenchAdapter struct {
    SysbenchPath string
}

func NewSysbenchAdapter(path string) *SysbenchAdapter

func (a *SysbenchAdapter) Detect(ctx context.Context) (*ToolInfo, error)

func (a *SysbenchAdapter) BuildRunCommand(
    ctx context.Context,
    conn connection.Connection,
    config *Config,
) (*Command, error)

func (a *SysbenchAdapter) ParseOutput(
    ctx context.Context,
    stdout string,
    stderr string,
) (*execution.BenchmarkResult, error)
```

**Config 结构**:
```go
type Config struct {
    BenchmarkType string                 // oltp_read_write, etc.
    Parameters    map[string]interface{} // threads, time, etc.
    Options       map[string]interface{} // report_interval, etc.
}
```

**Command 结构**:
```go
type Command struct {
    Args        []string
    Env         []string
    WorkDir     string
    Stdin       string
}
```

---

### adapter.SwingbenchAdapter

Swingbench 工具适配器。

```go
package adapter

type SwingbenchAdapter struct {
    SwingbenchPath string
    JavaPath       string
}

func NewSwingbenchAdapter(swingbenchPath string, javaPath string) *SwingbenchAdapter

func (a *SwingbenchAdapter) Detect(ctx context.Context) (*ToolInfo, error)

func (a *SwingbenchAdapter) BuildRunCommand(
    ctx context.Context,
    conn connection.Connection,
    config *Config,
) (*Command, error)

func (a *SwingbenchAdapter) ParseOutput(
    ctx context.Context,
    stdout string,
    stderr string,
) (*execution.BenchmarkResult, error)
```

**注意**: Swingbench 只支持 Oracle 数据库。

---

### adapter.HammerDBAdapter

HammerDB 工具适配器。

```go
package adapter

type HammerDBAdapter struct {
    HammerDBPath string
    TCLPath       string
}

func NewHammerDBAdapter(hammerdbPath string) *HammerDBAdapter

func (a *HammerDBAdapter) Detect(ctx context.Context) (*ToolInfo, error)

func (a *HammerDBAdapter) BuildRunCommand(
    ctx context.Context,
    conn connection.Connection,
    config *Config,
) (*Command, error)

func (a *HammerDBAdapter) ParseOutput(
    ctx context.Context,
    stdout string,
    stderr string,
) (*execution.BenchmarkResult, error)
```

**支持数据库**: MySQL, Oracle, SQL Server, PostgreSQL

---

### keyring.FileFallback

文件降级密钥管理。

```go
package keyring

func NewFileFallback(keyDir string, serviceName string) (KeyringProvider, error)

type KeyringProvider interface {
    Set(password string) error
    Get() (string, error)
    Delete() error
}
```

**实现**: 使用 AES-256-GCM 加密，密钥派生自系统信息。

---

## Transport 层

### CLI 命令

```bash
# 查看版本
./build/db-benchmind-cli version

# 查看帮助
./build/db-benchmind-cli help

# 列出连接
./build/db-benchmind-cli list

# 检测工具
./build/db-benchmind-cli detect
```

---

## 类型定义

### DatabaseType

```go
type DatabaseType string

const (
    DatabaseTypeMySQL      DatabaseType = "mysql"
    DatabaseTypeOracle     DatabaseType = "oracle"
    DatabaseTypeSQLServer  DatabaseType = "sqlserver"
    DatabaseTypePostgreSQL DatabaseType = "postgresql"
)
```

---

### FindOptions

```go
type FindOptions struct {
    Limit     int    // 限制返回数量
    Offset    int    // 跳过数量
    SortBy    string // 排序字段
    SortOrder string // 排序方向: ASC, DESC
}
```

---

### ToolInfo

```go
type ToolInfo struct {
    Found   bool   `json:"found"`
    Path    string `json:"path,omitempty"`
    Version string `json:"version,omitempty"`
}
```

---

### ReportFormat

```go
type ReportFormat string

const (
    FormatMarkdown ReportFormat = "markdown"
    FormatHTML     ReportFormat = "html"
    FormatJSON     ReportFormat = "json"
    FormatPDF      ReportFormat = "pdf"
)
```

---

## 错误处理

### 通用错误

```go
var (
    ErrNotFound      = errors.New("not found")
    ErrAlreadyExists = errors.New("already exists")
    ErrInvalidInput  = errors.New("invalid input")
)
```

### 连接错误

```go
var (
    ErrConnectionNotFound      = errors.New("connection not found")
    ErrConnectionAlreadyExists = errors.New("connection already exists")
    ErrInvalidConnection       = errors.New("invalid connection configuration")
)
```

### 运行错误

```go
var (
    ErrRunNotFound      = errors.New("run not found")
    ErrRunNotCancelable = errors.New("run cannot be cancelled")
    ErrInvalidRunState  = errors.New("invalid run state transition")
)
```

### 工具错误

```go
var (
    ErrToolNotDetected  = errors.New("benchmark tool not detected")
    ErrToolFailed       = errors.New("benchmark tool execution failed")
    ErrOutputParseError = errors.New("failed to parse tool output")
)
```

---

## 使用示例

### 完整工作流示例

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/whhaicheng/DB-BenchMind/internal/app/usecase"
    "github.com/whhaicheng/DB-BenchMind/internal/domain/connection"
    "github.com/whhaicheng/DB-BenchMind/internal/domain/execution"
    "github.com/whhaicheng/DB-BenchMind/internal/infra/adapter"
    "github.com/whhaicheng/DB-BenchMind/internal/infra/database"
    "github.com/whhaicheng/DB-BenchMind/internal/infra/database/repository"
    "github.com/whhaicheng/DB-BenchMind/internal/infra/keyring"
    "github.com/whhaicheng/DB-BenchMind/internal/infra/report"
)

func main() {
    ctx := context.Background()

    // 1. 初始化数据库
    db, err := database.InitializeSQLite(ctx, "./data/db-benchmind.db")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // 2. 初始化 repositories
    connRepo := repository.NewSQLiteConnectionRepository(db)
    runRepo := repository.NewSQLiteRunRepository(db)
    templateRepo := repository.NewTemplateRepository(db)

    // 3. 初始化 keyring
    keyringProvider, err := keyring.NewFileFallback("./data", "")
    if err != nil {
        log.Fatal(err)
    }

    // 4. 初始化 adapters
    adapterRegistry := adapter.NewRegistry()
    sysbenchAdapter := adapter.NewSysbenchAdapter("/usr/bin/sysbench")
    adapterRegistry.Register("sysbench", sysbenchAdapter)

    // 5. 初始化 use cases
    connUC := usecase.NewConnectionUseCase(connRepo, keyringProvider)
    templateUC := usecase.NewTemplateUseCase(templateRepo)
    benchmarkUC := usecase.NewBenchmarkUseCase(runRepo, adapterRegistry, keyringProvider)

    // 6. 创建连接
    mysqlConn := &connection.MySQLConnection{
        ID:       "test-conn-1",
        Name:     "Test MySQL",
        Host:     "localhost",
        Port:     3306,
        Database: "sbtest",
        Username: "root",
        Password: "password",
    }

    err = connUC.CreateConnection(ctx, mysqlConn)
    if err != nil {
        log.Fatal(err)
    }

    // 7. 测试连接
    result, err := connUC.TestConnection(ctx, mysqlConn.GetID())
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("Connection test: %+v\n", result)

    // 8. 获取模板
    templates, err := templateUC.ListBuiltinTemplates(ctx)
    if err != nil {
        log.Fatal(err)
    }

    // 9. 创建任务
    task := &execution.Task{
        ID:           "task-1",
        Name:         "Quick Test",
        ConnectionID: mysqlConn.GetID(),
        TemplateID:   templates[0].ID,
        Parameters: map[string]interface{}{
            "threads": 4,
            "time":    60,
        },
    }

    // 10. 执行测试
    run, err := benchmarkUC.ExecuteTask(ctx, task)
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("Started run: %s\n", run.ID)

    // 11. 等待完成
    for {
        status, err := benchmarkUC.GetRunStatus(ctx, run.ID)
        if err != nil {
            log.Fatal(err)
        }

        log.Printf("Status: %s, Progress: %.1f%%\n", status.State, status.Progress)

        if status.State == execution.StateCompleted || status.State == execution.StateFailed {
            break
        }

        time.Sleep(5 * time.Second)
    }

    // 12. 获取结果
    runResult, err := benchmarkUC.GetRunResult(ctx, run.ID)
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("TPS: %.2f\n", runResult.TPSCalculated)
    log.Printf("Avg Latency: %.2f ms\n", runResult.LatencyAvg)
    log.Printf("P95 Latency: %.2f ms\n", runResult.LatencyP95)
    log.Printf("Error Rate: %.2f%%\n", runResult.ErrorRate)
}
```

---

**版本 1.0.0 - 完**
