# Architecture Decisions

## System Architecture

DB-BenchMind 采用 **DDD + Clean Architecture + Hexagonal Architecture** 混合风格：

```
┌─────────────────────────────────────────────────────────────┐
│                        cmd/                                 │
│  ┌────────────────┐          ┌────────────────┐            │
│  │ db-benchmind/  │          │   cli-test/    │            │
│  └────────┬───────┘          └────────┬───────┘            │
└───────────┼──────────────────────────┼────────────────────┘
            │                          │
            ▼                          ▼
┌─────────────────────────────────────────────────────────────┐
│                   internal/transport/ui/                    │
│  (GUI 界面，仅 I/O，无业务逻辑)                              │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                    internal/app/usecase/                    │
│  (用例编排，定义接口，协调领域和基础设施)                     │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                     internal/domain/                        │
│  (核心业务逻辑，无外部依赖)                                   │
└─────────────────────────────────────────────────────────────┘
          ▲                  ▲                  ▲
          │                  │                  │
┌─────────┼──────────────────┼──────────────────┼─────────────┐
│         │                  │                  │             │
│  ┌──────┴───────┐  ┌──────┴───────┐  ┌──────┴───────┐      │
│  │  internal/  │  │              │  │   pkg/       │      │
│  │   infra/    │  │              │  │              │      │
│  │ ┌─────────┐ │  │              │  │ ┌─────────┐   │      │
│  │ │database │ │  │              │  │ │benchmark│   │      │
│  │ │adapter  │ │  │              │  │ │Adapter  │   │      │
│  │ │keyring  │ │  │              │  │ └─────────┘   │      │
│  │ │report   │ │  │              │  │               │      │
│  │ │chart    │ │  │              │  │               │      │
│  │ └─────────┘ │  │              │  │               │      │
│  └─────────────┘  │              │  │               │      │
│                   │              │  │               │      │
│  实现用例定义的接口 │              │  │ 定义适配器接口  │      │
└───────────────────┴──────────────┘  └───────────────┘      │
└─────────────────────────────────────────────────────────────┘
```

## Layer Responsibilities

### 1. Domain Layer (`internal/domain/`)
**职责**：核心业务逻辑，无外部依赖

- `connection/`：数据库连接模型和验证
- `template/`：压测模板定义和验证
- `execution/`：任务执行、状态机、结果
- `metric/`：指标计算（p95/p99）

**约束**：
- ❌ 禁止依赖 `internal/infra/`
- ❌ 禁止依赖外部库（仅标准库）

### 2. Application Layer (`internal/app/usecase/`)
**职责**：用例编排，定义接口

- `connection_usecase.go`：连接管理业务逻辑
- `template_usecase.go`：模板管理
- `benchmark_usecase.go`：压测执行编排
- `report_usecase.go`：报告生成
- `comparison_usecase.go`：结果对比

**约束**：
- ✅ 可以依赖 `domain/`、`pkg/`
- ✅ 通过接口依赖 `infra/`（依赖倒置）
- ❌ 禁止依赖 `transport/`

### 3. Infrastructure Layer (`internal/infra/`)
**职责**：外部依赖实现

- `database/`：SQLite 仓储实现
- `adapter/`：工具适配器（Sysbench/Swingbench/HammerDB）
- `keyring/`：密钥管理
- `report/`：报告生成器
- `chart/`：图表生成器

**约束**：
- ✅ 可以依赖 `domain/`、`pkg/`
- ✅ 可以使用外部依赖
- ✅ 实现 `app/` 定义的接口
- ❌ 禁止依赖 `app/`、`transport/`

### 4. Transport Layer (`internal/transport/ui/`)
**职责**：GUI 界面，仅 I/O

- `main_window.go`：主窗口
- `connection_page.go`：连接管理页面
- 其他页面：模板、任务、监控等

**约束**：
- ✅ 可以依赖 `app/usecase`、`domain/`、`pkg/`
- ✅ 可以使用 GUI 框架（Fyne）
- ❌ 禁止包含业务逻辑
- ❌ 禁止依赖 `infra/`

## Key Architectural Decisions

### ADR-001: Clean Architecture with DDD
**决策**：采用 Clean Architecture + DDD 分层

**理由**：
- 业务逻辑独立于外部依赖
- 易于测试和维护
- 符合宪法要求（Article I: Library-First）

**影响**：
- 需要明确的分层和接口定义
- 开发初期会有更多的抽象层

### ADR-002: Domain Layer Without External Dependencies
**决策**：领域层禁止依赖外部库

**理由**：
- 核心业务逻辑最稳定
- 易于单元测试
- 避免外部依赖变化影响核心逻辑

**影响**：
- 所有数据库操作通过接口
- 工具适配器在 infra 层实现

### ADR-003: Use Case Layer Defines Interfaces
**决策**：用例层定义接口，基础设施层实现

**理由**：
- 依赖倒置原则（Clean Architecture）
- 用例决定需要什么能力，而不是基础设施提供什么
- 易于替换实现（测试、mock）

**示例**：
```go
// internal/app/usecase/repository.go (用例定义)
type ConnectionRepository interface {
    Save(ctx context.Context, conn connection.Connection) error
    FindByID(ctx context.Context, id string) (connection.Connection, error)
    // ...
}

// internal/infra/database/repository/connection_repo.go (基础设施实现)
type SQLiteConnectionRepository struct { ... }
func (r *SQLiteConnectionRepository) Save(...) error { ... }
```

### ADR-004: No ORM, Direct SQL
**决策**：不使用 ORM，直接使用 `database/sql`

**理由**：
- 遵循宪法 Article VIII（反抽象门禁）
- Go 的 `database/sql` 已经足够好
- 避免学习 ORM 的复杂查询语言

**影响**：
- 需要手写 SQL（但有 schema.sql）
- 参数化查询防止注入
- 单元测试使用内存 SQLite

### ADR-005: Single SQLite Database
**决策**：使用单一 SQLite 数据库存储所有数据

**理由**：
- 简单、嵌入式、无服务依赖
- 满足桌面应用需求
- 易于备份和迁移

**约束**：
- 启用 WAL 模式提升并发
- 单连接池避免锁问题

## Data Flow

### Connection Management Flow

```
User (GUI)
  │
  ▼
┌─────────────────────┐
│ ConnectionPage      │
│ (transport/ui)      │
└─────────┬───────────┘
          │
          ▼
┌─────────────────────┐
│ ConnectionUseCase    │
│ (app/usecase)       │
│  - Validate         │
│  - Check Exists    │
└─────────┬───────────┘
          │
    ┌─────┴──────┐
    ▼             ▼
┌─────────┐  ┌──────────┐
│ Repo    │  │ Keyring  │
│ (infra) │  │ (infra)  │
└─────────┘  └──────────┘
```

## Technology Choices

| 技术 | 版本 | 理由 |
|------|------|------|
| Go | 1.22.2 | 静态类型、高性能、易于分发 |
| Fyne | v2.4.5 | 纯 Go、跨平台、自绘制引擎 |
| SQLite | modernc.org/sqlite | 无 CGO、纯 Go |
| go-keyring | latest | 支持 gnome-keyring |
| log/slog | 标准库 | 结构化日志 |

## Non-Goals

以下功能**明确不在当前范围**：

- ❌ 插件系统
- ❌ 分布式执行
- ❌ Web UI
- ❌ 实时报警
- ❌ 自动化测试调度
- ❌ 集群压测

如需添加，必须更新此文档并获得批准。
