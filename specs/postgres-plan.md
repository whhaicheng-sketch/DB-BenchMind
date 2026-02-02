# DB-BenchMind PostgreSQL 支持技术实现方案

**版本**: 1.0.0
**日期**: 2026-02-02
**作者**: AI Assistant
**状态**: 待评审

---

## 文档变更历史

| 版本 | 日期 | 作者 | 变更说明 |
|------|------|------|---------|
| 1.0.0 | 2026-02-02 | AI Assistant | 初始版本 |

---

## 目录

1. [技术上下文](#1-技术上下文)
2. [合宪性审查](#2-合宪性审查)
3. [实现架构](#3-实现架构)
4. [技术决策](#4-技术决策)
5. [实现计划](#5-实现计划)
6. [测试策略](#6-测试策略)
7. [风险与缓解](#7-风险与缓解)

---

## 1. 技术上下文

### 1.1 当前状态

**已完成（MySQL 支持）**:
- ✅ MySQL 驱动集成：`github.com/go-sql-driver/mysql`
- ✅ MySQL 连接结构体：`MySQLConnection`
- ✅ MySQL 连接测试：使用 `sql.Open()` + `db.Ping()`
- ✅ MySQL Sysbench 适配器：`--mysql-*` 参数
- ✅ UI 完整支持：表单、列表、编辑、测试

**PostgreSQL 现状**:
- ⚠️ 驱动未导入：缺少 `github.com/lib/pq`
- ⚠️ 连接测试未实现：`PostgreSQLConnection.Test()` 是占位符
- ✅ 数据结构完整：`PostgreSQLConnection` 定义完整
- ✅ Sysbench 适配器已有基础：`--pgsql-*` 参数生成
- ✅ UI 基本支持：但有 SSL Mode 选项不匹配问题

### 1.2 技术目标

**本版本目标**：
1. 添加 PostgreSQL 驱动依赖
2. 实现真实的 PostgreSQL 连接测试
3. 修复 UI SSL Mode 选项不匹配
4. 端到端验证 PostgreSQL 压测流程

**非目标**：
- ❌ 不新增功能，仅对齐 MySQL 支持级别
- ❌ 不实现 PostgreSQL 特有功能（如高级监控）
- ❌ 不重构现有架构

### 1.3 技术栈

| 层级 | 技术 | 版本 | 说明 |
|------|------|------|------|
| **驱动** | github.com/lib/pq | latest | 纯 Go PostgreSQL 驱动 |
| **数据库** | SQLite | modernc.org/sqlite | 存储连接配置 |
| **GUI** | Fyne | v2.7.2 | 连接管理界面 |
| **压测** | Sysbench | 1.0+ | 外部工具 |

---

## 2. 合宪性审查

### 2.1 对照 constitution.md 审查

#### Article I: Library-First Principle ✅

**检查**:
- PostgreSQL 连接测试逻辑在 `domain/connection/postgresql.go`
- 不依赖 GUI，可被 CLI/HTTP 复用
- 使用接口抽象：`Connection` interface

**结论**: 符合

#### Article II: CLI Interface Mandate ✅

**检查**:
- `PostgreSQLConnection.Test()` 方法可独立调用
- 返回结构化的 `TestResult`
- 无需 GUI 即可测试连接

**结论**: 符合

#### Article III: Test-First Imperative ✅

**策略**:
- 先编写单元测试（Red）
- 实现 Test() 方法（Green）
- 重构优化（Refactor）

**结论**: 遵循

#### Article IV: EARS Requirements Format ✅

**检查**:
- `postgres-spec.md` 使用 EARS 格式
- 所有需求有唯一 ID：REQ-PG-XXX-YYY
- 可测试、可验证

**结论**: 符合

#### Article V: Traceability Mandate ✅

**追溯关系**:
```
REQ-PG-CONN-010 → postgresql.go:Test() → connection_postgresql_test.go
REQ-PG-UI-001 → connection_page.go:PostgreSQL case
```

**结论**: 符合

#### Article VI: Project Memory ✅

**本方案将记录到**:
- `.specify/steering/postgresql.md`（新建）
- `specs/postgres-spec.md`（已完成）

**结论**: 符合

#### Article VII: Simplicity Gate ✅

**检查**:
- 不新增可执行入口
- 不新增 module
- 仅对齐现有 MySQL 功能

**结论**: 符合（无例外）

#### Article VIII: Anti-Abstraction Gate ✅

**策略**:
- 直接使用 `database/sql` 标准库
- 直接使用 `github.com/lib/pq` 驱动
- 不创建额外的抽象层

**结论**: 符合

#### Article IX: Integration-First Testing ✅

**测试策略**:
- 单元测试：Mock PostgreSQL 服务器
- 集成测试：需要真实 PostgreSQL 容器/进程
- E2E 测试：通过 CLI 测试完整流程

**结论**: 符合

---

## 3. 实现架构

### 3.1 架构概览

```
┌─────────────────────────────────────────────────────────────┐
│                      GUI Layer (Fyne)                       │
│                  connection_page.go                         │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   UseCase Layer                             │
│              connection_usecase.go                          │
│         TestConnection(id) → TestResult                     │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   Domain Layer                              │
│              connection/postgresql.go                       │
│    PostgreSQLConnection.Test(ctx) → TestResult             │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│              External Dependency                            │
│            github.com/lib/pq                               │
│              sql.Open("postgres", dsn)                     │
└─────────────────────────────────────────────────────────────┘
```

### 3.2 核心数据流

**连接测试流程**：
```
1. User clicks "Test Connection" (GUI)
   ↓
2. ConnectionPage.onTestConnection()
   ↓
3. ConnectionUseCase.TestConnection(ctx, id)
   ↓
4. GetConnectionByID(id) → PostgreSQLConnection
   ↓
5. PostgreSQLConnection.Test(ctx)
   ├─ sql.Open("postgres", dsn)
   ├─ db.PingContext(ctx)
   ├─ db.Query("SELECT version()")
   └─ Return TestResult
   ↓
6. GUI displays TestResult
```

**Sysbench 命令生成流程**：
```
1. User selects PostgreSQL connection + Sysbench template
   ↓
2. SysbenchAdapter.BuildRunCommand()
   ├─ conn.(*PostgreSQLConnection)
   ├─ Generate --pgsql-host={Host}
   ├─ Generate --pgsql-port={Port}
   ├─ Generate --pgsql-user={Username}
   ├─ Generate --pgsql-password={Password}
   └─ Generate --pgsql-db={Database}
   ↓
3. Set PGPASSWORD environment variable
   ↓
4. Execute sysbench command
```

### 3.3 文件结构

```
internal/
├── domain/
│   └── connection/
│       ├── postgresql.go          [MODIFY] 实现 Test() 方法
│       └── postgresql_test.go     [CREATE] 单元测试
├── app/
│   └── usecase/
│       └── connection_usecase.go  [NO CHANGE] 已支持
├── transport/
│   └── ui/
│       └── pages/
│           └── connection_page.go [MODIFY] 修复 SSL Mode 选项
└── infra/
    └── adapter/
        └── sysbench_adapter.go    [NO CHANGE] 已支持

go.mod                            [MODIFY] 添加 lib/pq
```

---

## 4. 技术决策

### 4.1 驱动选择

**决策**: 使用 `github.com/lib/pq`

**理由**:
- ✅ 官方推荐的纯 Go 驱动
- ✅ 广泛使用，社区活跃
- ✅ 支持 `database/sql` 标准接口
- ✅ 完整的 SSL 支持
- ✅ 与 MySQL 驱动 (`go-sql-driver/mysql`) 使用模式一致

**替代方案**:
- `github.com/jackc/pgx` - 性能更好，但 API 不同，增加复杂度

### 4.2 SSL Mode 选项修复

**问题**: UI 当前选项 `{"disabled", "preferred", "required"}` 不匹配 PostgreSQL 规范

**解决方案**: 更新为完整选项
```go
[]string{
    "disable",     // No SSL
    "allow",       // Try SSL, fallback to non-SSL
    "prefer",      // Try SSL first, fallback to non-SSL (default)
    "require",     // Force SSL, no certificate verification
    "verify-ca",   // Force SSL, verify CA certificate
    "verify-full", // Force SSL, verify CA and hostname
}
```

**影响范围**:
- `connection_page.go` line 296
- 不影响数据存储（仅 UI 显示）

### 4.3 连接测试实现策略

**参考 MySQL 实现**：
```go
// MySQL 实现 (mysql.go)
func (c *MySQLConnection) Test(ctx context.Context) (*TestResult, error) {
    start := time.Now()

    dsn := c.GetDSNWithPassword()
    db, err := sql.Open("mysql", dsn)
    if err != nil {
        return nil, err
    }
    defer db.Close()

    // Ping with timeout
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    if err := db.PingContext(ctx); err != nil {
        return &TestResult{
            Success: false,
            LatencyMs: time.Since(start).Milliseconds(),
            Error: err.Error(),
        }, nil
    }

    // Query version
    var version string
    if err := db.QueryRow("SELECT VERSION()").Scan(&version); err != nil {
        version = "Unknown"
    }

    return &TestResult{
        Success: true,
        LatencyMs: time.Since(start).Milliseconds(),
        Version: version,
    }, nil
}
```

**PostgreSQL 实现**：
- 完全相同的模式
- 驱动名称：`"postgres"` 而非 `"mysql"`
- SQL 查询：`SELECT version()` 而非 `SELECT VERSION()`
- DSN 格式：`host=... port=...` 而非 `user:pass@tcp(host:port)/...`

---

## 5. 实现计划

### 5.1 Phase 1: 驱动集成与测试实现

**目标**: 实现核心连接测试功能

#### Task 1.1: 添加 PostgreSQL 驱动依赖
**File**: `go.mod`
**Action**:
```bash
go get github.com/lib/pq
go mod tidy
```
**Acceptance**:
- `go.mod` 包含 `github.com/lib/pq`
- `go build ./...` 成功
- 无依赖冲突

#### Task 1.2: 导入 PostgreSQL 驱动
**File**: `internal/domain/connection/postgresql.go`
**Action**: 添加导入
```go
import (
    "database/sql"
    _ "github.com/lib/pq"  // Register PostgreSQL driver
    "context"
    "fmt"
    "time"
)
```
**Acceptance**:
- 文件编译成功
- 驱动已注册到 `database/sql`

#### Task 1.3: 实现 PostgreSQL 连接测试
**File**: `internal/domain/connection/postgresql.go:107-127`
**Action**: 替换占位符实现
**Implementation**:
```go
func (c *PostgreSQLConnection) Test(ctx context.Context) (*TestResult, error) {
    start := time.Now()

    // Build DSN with password
    dsn := c.GetDSNWithPassword()

    // Open connection
    db, err := sql.Open("postgres", dsn)
    if err != nil {
        return &TestResult{
            Success:   false,
            LatencyMs: time.Since(start).Milliseconds(),
            Error:     fmt.Sprintf("Failed to open connection: %v", err),
        }, nil
    }
    defer db.Close()

    // Set timeout
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    // Test connection
    if err := db.PingContext(ctx); err != nil {
        return &TestResult{
            Success:   false,
            LatencyMs: time.Since(start).Milliseconds(),
            Error:     fmt.Sprintf("Connection failed: %v", err),
        }, nil
    }

    // Query PostgreSQL version
    var version string
    err = db.QueryRow("SELECT version()").Scan(&version)
    if err != nil {
        version = "Unknown"
    }

    latency := time.Since(start).Milliseconds()

    return &TestResult{
        Success:   true,
        LatencyMs: latency,
        Version:   version,
        Error:     "",
    }, nil
}
```
**Acceptance**:
- 方法实现完整
- 成功路径返回 `Success=true`, `Version`, `LatencyMs`
- 失败路径返回 `Success=false`, `Error`
- 超时控制：5 秒

#### Task 1.4: 编写单元测试
**File**: `internal/domain/connection/postgresql_test.go` [NEW]
**Action**: 创建测试文件
**Test Cases**:
1. `TestPostgreSQLConnection_Validate_ValidInput` - 测试参数验证
2. `TestPostgreSQLConnection_Validate_MissingRequiredFields` - 测试必填字段
3. `TestPostgreSQLConnection_Validate_InvalidPort` - 测试无效端口
4. `TestPostgreSQLConnection_Validate_InvalidSSLMode` - 测试无效 SSL Mode
5. `TestPostgreSQLConnection_GetDSN` - 测试 DSN 生成
6. `TestPostgreSQLConnection_Redact` - 测试密码隐藏
7. `TestPostgreSQLConnection_SetPassword_GetPassword` - 测试密码设置

**Acceptance**:
- 所有测试通过
- 覆盖率 > 80%

---

### 5.2 Phase 2: UI 修复

**目标**: 修复 SSL Mode 选项不匹配问题

#### Task 2.1: 修复 SSL Mode 选项
**File**: `internal/transport/ui/pages/connection_page.go:296`
**Action**: 更新选项
```go
d.sslSelect = widget.NewSelect([]string{
    "disable",     // No SSL
    "allow",       // Try SSL, fallback to non-SSL
    "prefer",      // Try SSL first, fallback to non-SSL (default)
    "require",     // Force SSL, no certificate verification
    "verify-ca",   // Force SSL, verify CA certificate
    "verify-full", // Force SSL, verify CA and hostname
}, nil)
d.sslSelect.SetSelected("prefer")  // Default to prefer
```
**Acceptance**:
- 选项与 PostgreSQL 规范一致
- 默认值为 "prefer"
- UI 显示正常

---

### 5.3 Phase 3: 集成测试

**目标**: 端到端验证 PostgreSQL 支持

#### Task 3.1: 端到端连接测试
**Action**:
1. 启动 GUI
2. 进入 Connections 页面
3. 添加 PostgreSQL 连接（使用真实或测试数据库）
4. 点击 "Test Connection"
5. 验证结果显示正确

**Acceptance**:
- 成功连接显示版本号
- 失败连接显示错误信息
- 延迟显示合理（< 5000ms）

#### Task 3.2: Sysbench 厽令生成测试
**Action**:
1. 选择 PostgreSQL 连接
2. 选择 Sysbench 模板
3. 配置参数
4. 启动压测（dry-run 或短时间）
5. 检查生成的命令

**Acceptance**:
- 命令包含 `--pgsql-*` 参数
- `PGPASSWORD` 环境变量已设置
- 命令可以成功执行

#### Task 3.3: 结果解析与存储测试
**Action**:
1. 执行完整的 Sysbench PostgreSQL 压测
2. 等待完成
3. 查看 History 页面

**Acceptance**:
- 结果正确存储
- Database Type = "postgresql"
- 指标正确解析

---

### 5.4 Phase 4: 回归测试

**目标**: 确保 MySQL 功能不受影响

#### Task 4.1: MySQL 连接测试
**Action**:
1. 创建/测试 MySQL 连接
2. 执行 MySQL Sysbench 压测

**Acceptance**:
- 所有 MySQL 功能正常
- 无性能退化

---

## 6. 测试策略

### 6.1 测试金字塔

```
        /\
       /E2E\        10% - 端到端测试
      /------\
     /  集成  \      30% - 集成测试
    /----------\
   /   单元测试  \    60% - 单元测试
  /--------------\
```

### 6.2 单元测试策略

**范围**: `internal/domain/connection/postgresql_test.go`

**Mock 策略**:
- 不 mock `database/sql`（使用标准库）
- 需要 PostgreSQL 服务器运行（可使用 Docker）

**测试覆盖**:
- ✅ 参数验证（所有必填字段）
- ✅ SSL Mode 验证（所有选项）
- ✅ DSN 生成（各种参数组合）
- ✅ 密码隐藏
- ⚠️ 连接测试（需要真实 PostgreSQL，可选）

### 6.3 集成测试策略

**范围**: `test/integration/postgres_integration_test.go` [NEW]

**前提条件**:
- PostgreSQL 服务器运行（或 Docker 容器）
- 测试数据库可访问

**测试场景**:
1. 创建 PostgreSQL 连接
2. 测试连接成功
3. 测试连接失败（错误参数）
4. 读取连接（包含密码）
5. 更新连接
6. 删除连接

### 6.4 E2E 测试策略

**手动测试清单**:

**连接管理**:
- [ ] 创建 PostgreSQL 连接
- [ ] 测试连接成功
- [ ] 测试连接失败（错误主机/端口）
- [ ] 编辑 PostgreSQL 连接
- [ ] 删除 PostgreSQL 连接
- [ ] 列表显示 PostgreSQL 连接

**压测执行**:
- [ ] 配置 Sysbench PostgreSQL 任务
- [ ] 执行 prepare 阶段
- [ ] 执行 run 阶段
- [ ] 执行 cleanup 阶段
- [ ] 查看结果在 History 页面

**SSL 测试**:
- [ ] SSL Mode = disable，连接成功
- [ ] SSL Mode = require，连接成功（需要 SSL 配置）
- [ ] SSL Mode = prefer，连接成功

**UI/UX**:
- [ ] 连接表单字段正确
- [ ] 错误提示友好
- [ ] 密码不以明文显示

---

## 7. 风险与缓解

### 7.1 技术风险

| 风险 | 概率 | 影响 | 缓解措施 |
|------|------|------|---------|
| PostgreSQL 驱动 API 不熟悉 | 中 | 中 | 参考 MySQL 实现，两者模式相同 |
| SSL 证书验证复杂 | 低 | 低 | 先实现基本 SSL，证书验证作为 P1 |
| 测试环境不可用 | 中 | 中 | 使用 Docker PostgreSQL 容器 |
| Sysbench 命令参数错误 | 低 | 低 | 参考官方文档，已有代码基础 |

### 7.2 进度风险

| 风险 | 概率 | 影响 | 缓解措施 |
|------|------|------|---------|
| 预估时间不足 | 中 | 中 | 分阶段交付，优先 P0 功能 |
| 测试环境搭建耗时 | 中 | 低 | 提前准备 Docker Compose |
| 依赖冲突 | 低 | 中 | `go mod tidy` 及时检查 |

### 7.3 质量风险

| 风险 | 概率 | 影响 | 缓解措施 |
|------|------|------|---------|
| 回归 MySQL 功能 | 低 | 高 | 完整回归测试 |
| 代码覆盖率不足 | 中 | 中 | 强制 TDD，覆盖率检查 |
| UI 修复遗漏 | 低 | 中 | 代码审查 + 手动测试 |

---

## 8. 质量门禁

### 8.1 代码质量

- [ ] `go build ./...` 成功
- [ ] `go test ./...` 全部通过
- [ ] `go test ./... -race` 无竞态
- [ ] `gofmt -l .` 无输出
- [ ] `go vet ./...` 无警告
- [ ] `golangci-lint run` 无错误
- [ ] `govulncheck ./...` 无漏洞

### 8.2 功能验收

- [ ] 所有 P0 需求实现
- [ ] 所有单元测试通过
- [ ] 集成测试通过（如有环境）
- [ ] E2E 手动测试通过
- [ ] MySQL 回归测试通过

### 8.3 文档更新

- [ ] `specs/postgres-spec.md` 完成
- [ ] `specs/postgres-plan.md` 完成
- [ ] `specs/postgres-tasks.md` 完成
- [ ] Traceability 更新（需求 → 测试 → 实现）

---

## 9. 时间估算

| Phase | 任务 | 估算时间 |
|-------|------|---------|
| Phase 1 | 驱动集成与测试实现 | 2h |
| Phase 2 | UI 修复 | 0.5h |
| Phase 3 | 集成测试 | 1.5h |
| Phase 4 | 回归测试 | 1h |
| **总计** | | **5h** |

**说明**:
- 不包括测试环境搭建时间
- 不包括文档编写时间
- 假设 PostgreSQL 测试服务器已就绪

---

## 10. 附录

### 10.1 参考文档

- [PostgreSQL 官方文档](https://www.postgresql.org/docs/)
- [lib/pq 驱动文档](https://pkg.go.dev/github.com/lib/pq)
- [Sysbench PostgreSQL 文档](https://github.com/akopytov/sysbench/tree/master/src/drivers/pgsqldriver)
- [MySQL 实现参考](internal/domain/connection/mysql.go)

### 10.2 相关文件

- `specs/postgres-spec.md` - 需求规格
- `specs/postgres-tasks.md` - 任务分解
- `internal/domain/connection/mysql.go` - MySQL 实现参考
- `internal/infra/adapter/sysbench_adapter.go` - Sysbench 适配器
