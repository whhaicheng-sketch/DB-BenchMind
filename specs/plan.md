# DB-BenchMind 技术实现方案 (Plan)

**版本**: 1.0.0
**日期**: 2026-01-27
**作者**: 首席架构师
**状态**: 待评审

---

## 文档变更历史

| 版本 | 日期 | 作者 | 变更说明 |
|------|------|------|---------|
| 1.0.0 | 2026-01-27 | 首席架构师 | 初始版本 |
| 1.1.0 | 2026-01-27 | 首席架构师 | 细化项目结构、核心数据结构、接口设计 |

---

## 目录

1. [技术上下文总结](#1-技术上下文总结)
2. [合宪性审查](#2-合宪性审查)
3. [项目结构细化](#3-项目结构细化)
4. [核心数据结构](#4-核心数据结构)
5. [接口设计](#5-接口设计)
6. [技术决策记录](#6-技术决策记录)
7. [分阶段实施计划](#7-分阶段实施计划)
8. [测试策略](#8-测试策略)
9. [质量门禁](#9-质量门禁)
10. [风险与缓解](#10-风险与缓解)

---

## 1. 技术上下文总结

### 1.1 项目定位

DB-BenchMind 是一款**桌面压测工作台**，面向数据库工程师和性能测试工程师，提供统一的 GUI 界面来编排和运行外部压测工具。

**核心价值**：
- **可复现性**：配置与结果完整保存
- **可对比性**：多次运行结果对比分析
- **可交付性**：多格式报告导出
- **易用性**：降低命令行工具使用门槛

### 1.2 技术选型

#### 1.2.1 核心技术栈

| 技术领域 | 技术选型 | 版本 | 选型理由 |
|---------|---------|------|---------|
| **编程语言** | Go | 1.22.2 | 静态类型、高性能、易于分发 |
| **GUI 框架** | Fyne | v2.4.5 | 纯 Go、跨平台、自绘制引擎、单二进制打包 |
| **数据库** | SQLite | modernc.org/sqlite | 无 CGO、纯 Go、高性能、嵌入式 |
| **密钥管理** | go-keyring | latest | 支持 gnome-keyring、安全可靠 |
| **日志** | log/slog | 标准库 | 结构化日志、标准库、无需额外依赖 |
| **图表** | fynesimplechart | v1.x | Fyne 生态兼容 |

#### 1.2.2 辅助工具

| 工具 | 用途 | 必需性 |
|------|------|--------|
| pandoc | Markdown → PDF 转换 | 可选（降级到 MD/HTML） |
| wkhtmltopdf | HTML → PDF 转换 | 可选 |
| golangci-lint | 代码质量检查 | 必需（开发） |
| govulncheck | 安全漏洞扫描 | 必需（开发） |

#### 1.2.3 压测工具（外部依赖）

| 工具 | 支持数据库 | 优先级 | 备注 |
|------|-----------|--------|------|
| **Sysbench** | MySQL, PostgreSQL | P0 | 开源、文档完善 |
| **Swingbench** | Oracle | P1 | 需要 Java 环境 |
| **HammerDB** | 全部 | P1 | 支持 TCL 脚本 |

### 1.3 运行环境

- **操作系统**: Ubuntu 24（带桌面 GUI）
- **Go 版本**: 1.22.2
- **内存要求**: 最低 2GB，推荐 4GB
- **磁盘空间**: 最低 1GB 可用空间

### 1.4 架构风格

本项目采用 **DDD（领域驱动设计）** + **Clean Architecture（整洁架构）** + **Hexagonal Architecture（六边形架构）** 的混合风格：

- **领域驱动**：核心业务逻辑在 domain 层
- **整洁架构**：依赖方向由外向内
- **六边形架构**：通过适配器隔离外部依赖

---

## 2. 合宪性审查

### 2.1 对照 constitution.md 逐条审查

#### Article I: Library-First Principle（库优先原则）✅

**宪法要求**：
> 任何功能必须先实现为可复用的 Go 包（library），再由上层（CLI/HTTP/Job）调用；禁止把业务逻辑直接写进 main/handler。

**方案体现**：

1. **核心能力在 `internal/<domain>` 或 `pkg/<domain>`**
   ```
   internal/
   ├── domain/              # 领域模型（无外部依赖）
   │   ├── connection/      # 连接领域
   │   ├── template/        # 模板领域
   │   ├── execution/       # 执行领域
   │   └── metric/          # 指标领域
   ├── app/                # 应用层（用例编排）
   │   └── usecase/        # 业务用例
   └── infra/              # 基础设施（适配器实现）
       ├── adapter/        # 工具适配器
       ├── database/       # 数据库
       └── keyring/        # 密钥管理
   ```

2. **`cmd/<app>` 仅做装配与 I/O**
   ```go
   // cmd/db-benchmind/main.go
   func main() {
       // 1. 初始化基础设施
       db := initDatabase()
       keyring := initKeyring()
       logger := initLogger()

       // 2. 初始化仓储
       connRepo := database.NewSQLiteConnectionRepository(db)
       templateRepo := database.NewSQLiteTemplateRepository(db)
       runRepo := database.NewSQLiteRunRepository(db)

       // 3. 初始化用例
       connUC := usecase.NewConnectionUseCase(connRepo, keyring)
       templateUC := usecase.NewTemplateUseCase(templateRepo)
       benchmarkUC := usecase.NewBenchmarkUseCase(runRepo, adapterRegistry, keyring)

       // 4. 启动 GUI
       ui.NewMainWindow(connUC, templateUC, benchmarkUC).ShowAndRun()
   }
   ```

3. **library 必须"可测试、可组合、无隐式副作用"**
   - ✅ 显式依赖注入（struct/func 参数）
   - ✅ 禁止用全局变量传状态
   - ✅ 错误必须携带上下文

**审查结果**: ✅ **符合**

---

#### Article II: CLI Interface Mandate（CLI 接口强制）⚠️

**宪法要求**：
> 每个 library 必须提供 CLI 入口，确保功能可观测、可脚本化、可端到端验证。

**方案体现**：

本项目是 **GUI 桌面应用**，但核心库仍然可以通过 CLI 进行测试和验证。

**调整方案**：
```go
// cmd/cli-test/main.go  (新增测试工具)
package main

func main() {
    app := cli.NewApp()
    app.Name = "db-benchmind-cli"
    app.Usage = "DB-BenchMind CLI test tool"

    app.Commands = []*cli.Command{
        {
            Name:  "connection",
            Usage: "Test database connection",
            Flags: []cli.Flag{
                &cli.StringFlag{Name: "type", Usage: "Database type"},
                &cli.StringFlag{Name: "host", Usage: "Host"},
                &cli.IntFlag{Name: "port", Usage: "Port"},
            },
            Action: testConnection,
        },
        {
            Name:  "benchmark",
            Usage: "Run benchmark",
            Action: runBenchmark,
        },
    }

    app.Run(os.Args)
}
```

**审查结果**: ⚠️ **部分符合**（需要补充 CLI 测试工具）

---

#### Article III: Test-First Imperative（测试先行铁律）✅

**宪法要求**：
> 严格 TDD：先写测试并确认失败（Red），再写实现（Green），最后重构（Refactor）。

**方案体现**：

1. **测试优先表格驱动测试**
   ```go
   // internal/domain/connection/mysql_test.go
   func TestMySQLConnection_Validate(t *testing.T) {
       tests := []struct {
           name    string
           conn    *MySQLConnection
           wantErr bool
           errMsg  string
       }{
           {"valid", validConn, false, ""},
           {"invalid port", invalidPortConn, true, "port must be between"},
           // ... 更多测试用例
       }

       for _, tt := range tests {
           t.Run(tt.name, func(t *testing.T) {
               err := tt.conn.Validate()
               // 断言...
           })
       }
   }
   ```

2. **测试命名体现行为**
   - `Test_<Func>_<Scenario>`
   - 示例：`TestMySQLConnection_Validate_ValidConnection`, `TestMySQLConnection_Validate_InvalidPort`

3. **优先 fake 而非 mock**
   - 数据库：使用内存 SQLite (`:memory:`)
   - HTTP：使用 `httptest.Server`
   - 文件系统：使用临时目录 (`t.TempDir()`)

**审查结果**: ✅ **符合**

---

#### Article IV: EARS Requirements Format（EARS 要求格式）✅

**宪法要求**：
> spec.md 的"需求/验收"必须用 EARS 句式。

**检查 spec.md**：

所有需求都使用 EARS 格式：
- ✅ `WHEN 用户点击"新增连接"，THE SYSTEM SHALL 显示连接配置表单`
- ✅ `WHILE 任务运行，THE SYSTEM SHALL 每秒采集并显示指标`
- ✅ `WHEN 连接测试成功，THE SYSTEM SHALL 显示"成功 (耗时 XXms)"`

**审查结果**: ✅ **符合**

---

#### Article V: Traceability Mandate（全链路可追溯）✅

**宪法要求**：
> 需求→设计→任务→测试→实现必须 100% 可追溯。

**方案体现**：

1. **需求 ID 映射**
   - spec.md 中定义了 `REQ-CONN-001` ~ `REQ-SET-005` 等需求
   - 每个代码文件在头部注释中引用需求：
     ```go
     // internal/domain/connection/mysql.go
     // Implements: REQ-CONN-002, REQ-CONN-003
     package connection
     ```

2. **测试可追溯**
   ```go
   // internal/domain/connection/mysql_test.go
   // Tests for: REQ-CONN-010 (实时输入验证)
   func TestMySQLConnection_Validate_PortOutOfRange(t *testing.T) { ... }
   ```

3. **可追溯性文档**
   - 创建 `docs/traceability.md` 维护映射表

**审查结果**: ✅ **符合**

---

#### Article VI: Project Memory（项目记忆 / Steering）✅

**宪法要求**：
> 项目必须维护"记忆层"，让新 Agent/新成员无需读全仓库也能遵循既定决策。

**方案体现**：

创建 `.specify/steering/` 目录：
```
.specify/steering/
├── product.md          # 问题定义、用户与范围边界
├── architecture.md     # 架构边界、关键依赖、包分层
├── testing.md          # 测试金字塔、环境约束、CI 策略
└── decisions.md        # 重要决策与取舍
```

**审查结果**: ✅ **符合**

---

#### Article VII: Simplicity Gate（简洁之门）✅

**宪法要求**：
> 初始阶段最多 ≤3 个可执行入口；禁止"为未来做预留"的框架化设计。

**方案体现**：

1. **可执行入口控制**
   ```
   cmd/
   ├── db-benchmind/    # 主 GUI 应用
   └── cli-test/        # CLI 测试工具（可选）
   ```
   - 总共 2 个入口（≤3 个）

2. **只实现 spec.md 明确范围内的内容**
   - ✅ 4 种数据库连接
   - ✅ 3 个压测工具
   - ✅ 7 个内置模板
   - ❌ 不做插件系统
   - ❌ 不做分布式执行
   - ❌ 不做 Web UI

**审查结果**: ✅ **符合**

---

#### Article VIII: Anti-Abstraction Gate（反抽象之门）✅

**宪法要求**：
> 不要为框架/标准库"再包一层"；谨慎使用 interface/generics。

**方案体现**：

1. **直接使用标准库**
   ```go
   import "log/slog"

   // ✅ 正确：直接使用
   slog.InfoContext(ctx, "starting execution", "op", "execute")

   // ❌ 错误：不必要的封装
   // type Logger struct { ... }
   // func (l *Logger) Info(msg string) { ... }
   ```

2. **接口由消费者定义**
   ```go
   // internal/app/usecase/connection_usecase.go
   // 用例定义接口，基础设施实现
   type ConnectionRepository interface {
       Save(ctx context.Context, conn connection.Connection) error
       FindByID(ctx context.Context, id string) (connection.Connection, error)
       // ...
   }

   // internal/infra/database/sqlite.go
   type SQLiteConnectionRepository struct { ... }
   func (r *SQLiteConnectionRepository) Save(...) error { ... }
   ```

3. **禁止的封装**
   - ❌ 不为 `database/sql` 封装 ORM
   - ❌ 不为 `log/slog` 封装日志框架
   - ❌ 不为 `context.Context` 封装上下文传递

**审查结果**: ✅ **符合**

---

#### Article IX: Integration-First Testing（集成优先测试）✅

**宪法要求**：
> 优先真实环境/真实交互的测试。

**方案体现**：

1. **Contract Tests（契约测试）**
   - `contracts/` 定义模板 Schema
   - 先写 Schema 验证测试

2. **真实 SQLite 测试**
   ```go
   func TestSQLiteConnectionRepository(t *testing.T) {
       db, _ := sql.Open("sqlite", "file::memory:?mode=memory")
       // 使用真实 SQLite，不是 mock
   }
   ```

3. **真实工具测试（如果可用）**
   ```go
   func TestSysbenchAdapter_Integration(t *testing.T) {
       if !isSysbenchAvailable() {
           t.Skip("sysbench not available")
       }
       // 运行真实的 sysbench 命令
   }
   ```

**审查结果**: ✅ **符合**

---

### 2.2 合宪性审查总结

| 条款 | 状态 | 备注 |
|------|------|------|
| Article I: Library-First | ✅ 符合 | 核心在 internal/ 和 pkg/ |
| Article II: CLI Interface | ⚠️ 部分符合 | 需补充 CLI 测试工具 |
| Article III: Test-First | ✅ 符合 | TDD + 表格驱动测试 |
| Article IV: EARS Format | ✅ 符合 | spec.md 使用 EARS |
| Article V: Traceability | ✅ 符合 | 需求 ID 映射 |
| Article VI: Project Memory | ✅ 符合 | .specify/steering/ |
| Article VII: Simplicity Gate | ✅ 符合 | 2个入口，无过度设计 |
| Article VIII: Anti-Abstraction | ✅ 符合 | 直接使用标准库 |
| Article IX: Integration-First | ✅ 符合 | 真实环境测试 |

**总体评价**: ✅ **符合宪法要求**（需补充 CLI 测试工具）

---

## 3. 项目结构细化

### 3.1 完整目录树

```
DB-BenchMind/
├── cmd/                                    # 可执行入口
│   ├── db-benchmind/                       # 主 GUI 应用
│   │   └── main.go                         # 入口：依赖装配、GUI 启动
│   └── cli-test/                           # CLI 测试工具（可选）
│       └── main.go                         # 入口：CLI 测试命令
│
├── internal/                               # 内部包（不可外部导入）
│   │
│   ├── app/                                # 应用层：用例编排
│   │   └── usecase/                        # 业务用例
│   │       ├── connection_usecase.go       # 连接管理用例
│   │       ├── connection_usecase_test.go  # 单元测试
│   │       ├── template_usecase.go         # 模板管理用例
│   │       ├── template_usecase_test.go
│   │       ├── benchmark_usecase.go        # 压测执行用例
│   │       ├── benchmark_usecase_test.go
│   │       ├── report_usecase.go           # 报告生成用例
│   │       ├── report_usecase_test.go
│   │       ├── comparison_usecase.go       # 结果对比用例
│   │       ├── comparison_usecase_test.go
│   │       ├── settings_usecase.go         # 设置管理用例
│   │       └── settings_usecase_test.go
│   │
│   ├── domain/                             # 领域层：核心业务逻辑
│   │   │
│   │   ├── connection/                     # 连接领域
│   │   │   ├── connection.go               # Connection 接口定义
│   │   │   ├── mysql.go                    # MySQLConnection 实现
│   │   │   ├── mysql_test.go               # 单元测试
│   │   │   ├── oracle.go                   # OracleConnection 实现
│   │   │   ├── oracle_test.go
│   │   │   ├── sqlserver.go                # SQLServerConnection 实现
│   │   │   ├── sqlserver_test.go
│   │   │   ├── postgresql.go               # PostgreSQLConnection 实现
│   │   │   ├── postgresql_test.go
│   │   │   └── testresult.go               # TestResult 结构
│   │   │
│   │   ├── template/                       # 模板领域
│   │   │   ├── template.go                 # Template 结构
│   │   │   ├── template_test.go
│   │   │   ├── parameter.go                # Parameter 定义
│   │   │   └── validator.go                # 模板验证器
│   │   │
│   │   ├── execution/                      # 执行领域
│   │   │   ├── run.go                      # Run 结构
│   │   │   ├── run_test.go
│   │   │   ├── state.go                    # RunState 状态机
│   │   │   ├── task.go                     # BenchmarkTask 结构
│   │   │   ├── task_test.go
│   │   │   └── result.go                   # BenchmarkResult 结构
│   │   │
│   │   └── metric/                         # 指标领域
│   │       ├── sample.go                   # MetricSample 结构
│   │       └── calculator.go               # 统计计算（p95/p99）
│   │
│   ├── infra/                              # 基础设施层：外部依赖实现
│   │   │
│   │   ├── database/                       # 数据库
│   │   │   ├── sqlite.go                   # SQLite 初始化
│   │   │   ├── schema.sql                  # 数据库 Schema
│   │   │   └── repository/                 # 仓储实现
│   │   │       ├── connection_repo.go      # ConnectionRepository 实现
│   │   │       ├── connection_repo_test.go
│   │   │       ├── template_repo.go        # TemplateRepository 实现
│   │   │       ├── template_repo_test.go
│   │   │       ├── run_repo.go             # RunRepository 实现
│   │   │       ├── run_repo_test.go
│   │   │       └── settings_repo.go        # SettingsRepository 实现
│   │   │
│   │   ├── adapter/                        # 工具适配器
│   │   │   ├── sysbench_adapter.go         # Sysbench 适配器
│   │   │   ├── sysbench_adapter_test.go
│   │   │   ├── swingbench_adapter.go       # Swingbench 适配器
│   │   │   ├── swingbench_adapter_test.go
│   │   │   ├── hammerdb_adapter.go         # HammerDB 适配器
│   │   │   ├── hammerdb_adapter_test.go
│   │   │   └── registry.go                 # 适配器注册表
│   │   │
│   │   ├── keyring/                        # 密钥管理
│   │   │   ├── provider.go                 # KeyringProvider 接口
│   │   │   ├── go_keyring.go               # go-keyring 实现
│   │   │   ├── file_fallback.go            # 加密文件降级实现
│   │   │   └── keyring_test.go
│   │   │
│   │   ├── report/                         # 报告生成
│   │   │   ├── generator.go                # ReportGenerator 接口
│   │   │   ├── markdown.go                 # Markdown 生成器
│   │   │   ├── html.go                     # HTML 生成器
│   │   │   ├── json.go                     # JSON 生成器
│   │   │   ├── pdf.go                      # PDF 生成器
│   │   │   └── template/                   # 报告模板
│   │   │       └── benchmark.md.tmpl
│   │   │
│   │   └── chart/                          # 图表生成
│   │       ├── chart.go                    # ChartGenerator 接口
│   │       ├── tps_trend.go                # TPS 趋势图
│   │       ├── latency_dist.go             # 延迟分布图
│   │       └── error_rate.go               # 错误率图
│   │
│   └── transport/                          # 传输层：对外接口
│       └── ui/                             # GUI 界面
│           ├── main_window.go              # 主窗口
│           ├── connection_page.go          # 连接管理页面
│           ├── template_page.go            # 模板管理页面
│           ├── task_page.go                # 任务配置页面
│           ├── monitor_page.go             # 运行监控页面
│           ├── history_page.go             # 历史记录页面
│           ├── comparison_page.go          # 结果对比页面
│           ├── report_page.go              # 报告导出页面
│           ├── settings_page.go            # 设置页面
│           └── widgets/                    # 自定义组件
│               ├── connection_form.go
│               ├── template_selector.go
│               └── log_viewer.go
│
├── pkg/                                    # 对外可复用库
│   └── benchmark/                          # 压测适配器包（可独立使用）
│       ├── adapter.go                      # Adapter 接口定义
│       ├── config.go                       # Config 结构
│       ├── result.go                       # Result 结构
│       └── sample.go                       # Sample 结构
│
├── contracts/                              # 契约定义（不可变资源）
│   ├── templates/                          # 内置模板
│   │   ├── sysbench-oltp-read-write.json
│   │   ├── sysbench-oltp-read-only.json
│   │   ├── sysbench-oltp-write-only.json
│   │   ├── swingbench-soe.json
│   │   ├── swingbench-calling.json
│   │   ├── hammerdb-tpcc.json
│   │   └── hammerdb-tpcb.json
│   ├── schemas/                            # JSON Schema 定义
│   │   └── template-v1.json
│   └── reports/                            # 报告模板
│       └── benchmark-report.md.tmpl
│
├── configs/                                # 配置样例
│   └── settings.example.json
│
├── scripts/                                # 脚本
│   ├── install-tools.sh                    # 工具安装脚本
│   ├── build-appimage.sh                   # AppImage 打包脚本
│   └── migrate-schema.sql                  # 数据库迁移脚本
│
├── test/                                   # 测试资源
│   ├── testdata/                           # 测试数据
│   │   ├── connections/                    # 测试用连接配置
│   │   ├── outputs/                        # 工具输出示例
│   │   │   ├── sysbench_success.txt
│   │   │   └── swingbench.xml
│   │   └── expected/                       # 预期结果
│   └── integration/                        # 集成测试
│       ├── benchmark_test.go
│       └── report_test.go
│
├── docs/                                   # 文档
│   ├── architecture.md                     # 架构文档
│   ├── api.md                              # API 文档
│   └── user-guide.md                       # 用户手册
│
├── .specify/                               # 项目记忆
│   └── steering/
│       ├── product.md                      # 产品定义
│       ├── architecture.md                 # 架构决策
│       ├── testing.md                      # 测试策略
│       └── decisions.md                    # 决策记录
│
├── results/                                # 结果目录（运行时生成）
│
├── go.mod
├── go.sum
├── Makefile
├── .golangci.yml                           # golangci-lint 配置
├── .gitignore
├── CLAUDE.md
├── constitution.md
├── README.md
├── specs/
│   ├── spec.md                             # 产品需求文档
│   └── plan.md                             # 本文档
└── docs/
    └── traceability.md                     # 需求追溯
```

### 3.2 包职责说明

#### 3.2.1 `cmd/` - 可执行入口

| 包 | 职责 | 依赖 |
|---|------|------|
| `cmd/db-benchmind/` | GUI 应用入口：初始化依赖、装配用例、启动 Fyne GUI | `internal/app/usecase`, `internal/transport/ui` |
| `cmd/cli-test/` | CLI 测试工具：提供命令行接口测试核心库 | `internal/app/usecase`, `pkg/benchmark` |

**依赖规则**：
- ✅ 可以依赖 `internal/` 和 `pkg/`
- ✅ 仅做装配和 I/O，不含业务逻辑
- ✅ 必须在 main 函数中完成依赖注入

---

#### 3.2.2 `internal/app/usecase/` - 应用层（用例编排）

| 包 | 职责 | 接口 | 依赖 |
|---|------|------|------|
| `connection_usecase.go` | 连接管理业务逻辑：增删改查、测试连接 | `ConnectionUseCase` | `domain/connection`, `infra/database`, `infra/keyring` |
| `template_usecase.go` | 模板管理业务逻辑：导入导出、验证 | `TemplateUseCase` | `domain/template`, `infra/database` |
| `benchmark_usecase.go` | 压测执行编排：状态机、阶段执行 | `BenchmarkUseCase` | `domain/execution`, `infra/adapter`, `infra/database` |
| `report_usecase.go` | 报告生成业务逻辑：多格式导出 | `ReportUseCase` | `infra/report`, `infra/chart`, `infra/database` |
| `comparison_usecase.go` | 结果对比业务逻辑：A/B对比、趋势分析 | `ComparisonUseCase` | `domain/execution`, `infra/database` |
| `settings_usecase.go` | 设置管理业务逻辑：配置持久化 | `SettingsUseCase` | `infra/database` |

**依赖规则**：
- ✅ 可以依赖 `domain/`、`pkg/`
- ✅ 通过接口依赖 `infra/`（由用例定义接口）
- ❌ 禁止依赖 `transport/`

**示例**：
```go
// internal/app/usecase/connection_usecase.go
package usecase

type ConnectionUseCase struct {
    repo    ConnectionRepository  // 用例定义接口
    keyring KeyringProvider
}

// ConnectionRepository 接口由用例定义
type ConnectionRepository interface {
    Save(ctx context.Context, conn connection.Connection) error
    FindByID(ctx context.Context, id string) (connection.Connection, error)
    // ...
}

// infra/database 实现
type SQLiteConnectionRepository struct { ... }
```

---

#### 3.2.3 `internal/domain/` - 领域层（核心业务逻辑）

| 包 | 职责 | 导出类型 | 依赖 |
|---|------|---------|------|
| `connection/` | 连接领域模型：4种数据库连接、验证、测试 | `Connection` 接口, `MySQLConnection`, `OracleConnection`, ... | **仅标准库** |
| `template/` | 模板领域模型：模板结构、参数验证 | `Template`, `Parameter`, `Validator` | **仅标准库** |
| `execution/` | 执行领域模型：Run、Task、状态机 | `Run`, `BenchmarkTask`, `RunState` | **仅标准库** |
| `metric/` | 指标领域模型：样本、统计计算 | `MetricSample`, `Calculator` | **仅标准库** |

**依赖规则**：
- ❌ **禁止依赖任何外部包**（仅标准库）
- ❌ **禁止依赖** `internal/infra/`
- ✅ 可以被 `app/`、`infra/`、`transport/` 依赖

**示例**：
```go
// internal/domain/connection/mysql.go
package connection

import (
    "context"
    "database/sql"
    "fmt"
    "time"
)

// MySQLConnection MySQL 连接配置
type MySQLConnection struct {
    ID       string    `json:"id"`
    Name     string    `json:"name"`
    Host     string    `json:"host"`
    Port     int       `json:"port"`
    Database string    `json:"database"`
    Username string    `json:"username"`
    Password string    `json:"-"` // 不序列化
    SSLMode  string    `json:"ssl_mode"`
}

// Validate 验证连接参数（纯业务逻辑，无外部依赖）
func (c *MySQLConnection) Validate() error {
    if c.Name == "" {
        return fmt.Errorf("name is required")
    }
    if c.Host == "" {
        return fmt.Errorf("host is required")
    }
    if c.Port < 1 || c.Port > 65535 {
        return fmt.Errorf("port must be between 1 and 65535, got %d", c.Port)
    }
    // ...
}

// Test 测试连接（仅使用标准库 database/sql）
func (c *MySQLConnection) Test(ctx context.Context) (*TestResult, error) {
    dsn := c.GetDSNWithPassword()
    db, err := sql.Open("mysql", dsn)
    if err != nil {
        return &TestResult{Success: false, Error: err.Error()}, nil
    }
    defer db.Close()

    start := time.Now()
    err = db.PingContext(ctx)
    latency := time.Since(start).Milliseconds()

    if err != nil {
        return &TestResult{Success: false, LatencyMs: latency, Error: err.Error()}, nil
    }

    // 获取版本
    var version string
    db.QueryRowContext(ctx, "SELECT VERSION()").Scan(&version)

    return &TestResult{Success: true, LatencyMs: latency, DatabaseVersion: version}, nil
}
```

---

#### 3.2.4 `internal/infra/` - 基础设施层（外部依赖实现）

| 包 | 职责 | 实现接口 | 依赖 |
|---|------|---------|------|
| `database/` | SQLite 仓储实现 | `ConnectionRepository`, `TemplateRepository`, `RunRepository` | `domain/*`, `modernc.org/sqlite` |
| `adapter/` | 工具适配器实现 | `pkg/benchmark.Adapter` | `domain/*`, `pkg/benchmark` |
| `keyring/` | 密钥管理实现 | `KeyringProvider`（用例定义） | `github.com/zalando/go-keyring` |
| `report/` | 报告生成器实现 | `ReportGenerator`（用例定义） | `domain/*` |
| `chart/` | 图表生成器实现 | `ChartGenerator`（用例定义） | `fyne.io/x/fynesimplechart` |

**依赖规则**：
- ✅ 可以依赖 `domain/`、`pkg/`
- ✅ 可以使用外部依赖（SQLite、keyring、Fyne等）
- ✅ 实现 `app/usecase` 定义的接口
- ❌ 禁止依赖 `app/`、`transport/`

**示例**：
```go
// internal/infra/database/repository/connection_repo.go
package repository

import (
    "context"
    "database/sql"
    "github.com/whhaicheng/DB-BenchMind/internal/domain/connection"
)

// SQLiteConnectionRepository 实现 ConnectionRepository 接口
type SQLiteConnectionRepository struct {
    db *sql.DB
}

// Save 保存连接到 SQLite
func (r *SQLiteConnectionRepository) Save(ctx context.Context, conn connection.Connection) error {
    configJSON, err := json.Marshal(conn)
    if err != nil {
        return fmt.Errorf("marshal connection: %w", err)
    }

    _, err = r.db.ExecContext(ctx,
        "INSERT INTO connections (id, name, type, config_json, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
        conn.GetID(), conn.GetName(), conn.GetType(), configJSON, time.Now(), time.Now())

    if err != nil {
        return fmt.Errorf("insert connection: %w", err)
    }

    return nil
}

// FindByID 根据 ID 查找连接
func (r *SQLiteConnectionRepository) FindByID(ctx context.Context, id string) (connection.Connection, error) {
    var name, connType, configJSON string
    err := r.db.QueryRowContext(ctx,
        "SELECT name, type, config_json FROM connections WHERE id = ?", id).
        Scan(&name, &connType, &configJSON)

    if err != nil {
        if err == sql.ErrNoRows {
            return nil, fmt.Errorf("connection not found: %s", id)
        }
        return nil, fmt.Errorf("query connection: %w", err)
    }

    // 根据 type 反序列化为具体类型
    switch connType {
    case "mysql":
        var conn connection.MySQLConnection
        if err := json.Unmarshal([]byte(configJSON), &conn); err != nil {
            return nil, fmt.Errorf("unmarshal mysql connection: %w", err)
        }
        return &conn, nil
    // ... 其他类型
    }

    return nil, fmt.Errorf("unknown connection type: %s", connType)
}
```

---

#### 3.2.5 `internal/transport/ui/` - 传输层（GUI 界面）

| 包 | 职责 | 依赖 |
|---|------|------|
| `main_window.go` | 主窗口：TabLayout 管理 | `fyne`, `app/usecase` |
| `connection_page.go` | 连接管理页面：表单、列表、按钮 | `fyne`, `app/usecase` |
| `template_page.go` | 模板管理页面：列表、详情 | `fyne`, `app/usecase` |
| `task_page.go` | 任务配置页面：动态表单 | `fyne`, `app/usecase` |
| `monitor_page.go` | 运行监控页面：实时更新 | `fyne`, `app/usecase` |
| `history_page.go` | 历史记录页面：列表、详情 | `fyne`, `app/usecase` |
| `comparison_page.go` | 结果对比页面：差异展示 | `fyne`, `app/usecase` |
| `report_page.go` | 报告导出页面：格式选择 | `fyne`, `app/usecase` |
| `settings_page.go` | 设置页面：配置表单 | `fyne`, `app/usecase` |

**依赖规则**：
- ✅ 可以依赖 `app/usecase`、`domain/`、`pkg/`
- ✅ 可以使用 GUI 框架（Fyne）
- ❌ 禁止依赖 `infra/`（必须通过 usecase）
- ❌ 禁止包含业务逻辑（仅 UI 交互）

**示例**：
```go
// internal/transport/ui/connection_page.go
package ui

import (
    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/widget"
    "github.com/whhaicheng/DB-BenchMind/internal/app/usecase"
)

// ConnectionPage 连接管理页面
type ConnectionPage struct {
    usecase *usecase.ConnectionUseCase
    list    *widget.List
}

func NewConnectionPage(u *usecase.ConnectionUseCase) *ConnectionPage {
    return &ConnectionPage{usecase: u}
}

func (p *ConnectionPage) GetContent() fyne.CanvasObject {
    // 连接列表
    p.list = widget.NewList(
        func() int {
            conns, _ := p.usecase.ListConnections(context.Background())
            return len(conns)
        },
        func() fyne.CanvasObject {
            return widget.NewLabel("")
        },
        func(id widget.ListItemID, obj fyne.CanvasObject) {
            conns, _ := p.usecase.ListConnections(context.Background())
            obj.(*widget.Label).SetText(conns[id].GetName())
        },
    )

    // 添加按钮
    addButton := widget.NewButton("Add Connection", func() {
        // 打开添加对话框（调用 usecase）
        p.showAddDialog()
    })

    return container.NewVBox(p.list, addButton)
}

func (p *ConnectionPage) showAddDialog() {
    // 显示表单，收集用户输入
    // 调用 p.usecase.CreateConnection(...)
}
```

---

#### 3.2.6 `pkg/` - 对外可复用库

| 包 | 职责 | 导出类型 | 依赖 |
|---|------|---------|------|
| `benchmark/` | 压测适配器包（可独立使用） | `Adapter` 接口, `Config`, `Result`, `Sample` | 标准库 |

**用途**：
- 其他项目可以单独使用 `pkg/benchmark` 集成压测能力
- 定义了统一的适配器接口规范

**依赖规则**：
- ❌ **禁止依赖** `internal/`
- ✅ 可以被 `internal/` 和外部项目依赖
- ✅ 仅依赖标准库（保持通用性）

---

### 3.3 依赖关系图

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
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │ ConnectionUC │  │ TemplateUC   │  │ BenchmarkUC  │      │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘      │
└─────────┼──────────────────┼──────────────────┼─────────────┘
          │                  │                  │
          ▼                  ▼                  ▼
┌─────────────────────────────────────────────────────────────┐
│                     internal/domain/                        │
│  (核心业务逻辑，无外部依赖)                                   │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │ connection/  │  │  template/   │  │ execution/   │      │
│  │  metric/     │  │              │  │              │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└─────────────────────────────────────────────────────────────┘
          ▲                  ▲                  ▲
          │                  │                  │
┌─────────┼──────────────────┼──────────────────┼─────────────┐
│         │                  │                  │             │
│  ┌──────┴───────┐  ┌──────┴───────┐  ┌──────┴───────┐      │
│  │  internal/  │  │  internal/   │  │   pkg/       │      │
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

**关键依赖规则**：
1. ✅ `cmd/` → `internal/` + `pkg/`
2. ✅ `transport/` → `app/` + `domain/` + `pkg/`
3. ✅ `app/` → `domain/` + `pkg/` + `infra/`(通过接口)
4. ✅ `infra/` → `domain/` + `pkg/`
5. ❌ `domain/` → 无外部依赖
6. ❌ `pkg/` → 无 `internal/` 依赖

---

## 4. 核心数据结构

### 4.1 连接领域（connection）

#### 4.1.1 Connection 接口

```go
// internal/domain/connection/connection.go
package connection

import (
    "context"
    "time"
)

// DatabaseType 数据库类型
type DatabaseType string

const (
    DatabaseTypeMySQL      DatabaseType = "mysql"
    DatabaseTypeOracle     DatabaseType = "oracle"
    DatabaseTypeSQLServer  DatabaseType = "sqlserver"
    DatabaseTypePostgreSQL DatabaseType = "postgresql"
)

// Connection 连接接口（根据 REQ-CONN-001 ~ REQ-CONN-010）
type Connection interface {
    // GetID 获取连接 ID
    GetID() string

    // GetName 获取连接名称
    GetName() string

    // SetName 设置连接名称
    SetName(name string)

    // GetType 获取数据库类型
    GetType() DatabaseType

    // Validate 验证连接参数（REQ-CONN-010）
    Validate() error

    // Test 测试连接可用性（REQ-CONN-003）
    // 返回 TestResult 包含成功/失败、耗时、版本、错误信息
    Test(ctx context.Context) (*TestResult, error)

    // GetDSN 生成连接字符串（不含密码，用于日志）
    GetDSN() string

    // GetDSNWithPassword 生成完整连接字符串（含密码，用于实际连接）
    GetDSNWithPassword() string

    // Redact 返回脱敏后的连接信息（用于日志和报告，REQ-CONN-008）
    Redact() string

    // ToJSON 序列化为 JSON（不含密码）
    ToJSON() ([]byte, error)
}

// TestResult 连接测试结果（REQ-CONN-004, REQ-CONN-005）
type TestResult struct {
    Success         bool    `json:"success"`          // 是否成功
    LatencyMs       int64   `json:"latency_ms"`        // 耗时（毫秒）
    DatabaseVersion string  `json:"database_version"`  // 数据库版本
    Error           string  `json:"error,omitempty"`   // 错误信息
}
```

#### 4.1.2 MySQLConnection 结构

```go
// internal/domain/connection/mysql.go
package connection

// MySQLConnection MySQL 连接配置（完整实现 spec.md 3.2.2）
type MySQLConnection struct {
    // 基础字段
    ID   string `json:"id"`    // UUID
    Name string `json:"name"`  // 连接名称（用户自定义）

    // 连接参数
    Host     string `json:"host"`     // 主机地址
    Port     int    `json:"port"`     // 端口（默认 3306）
    Database string `json:"database"` // 数据库名
    Username string `json:"username"` // 用户名
    Password string `json:"-"`        // 密码（不序列化，存储到 keyring）

    // SSL 配置
    SSLMode string `json:"ssl_mode"` // SSL模式：disabled/preferred/required

    // 元数据
    CreatedAt time.Time  `json:"created_at"`
    UpdatedAt time.Time  `json:"updated_at"`
}

// 实现 Connection 接口的所有方法
func (c *MySQLConnection) GetID() string { return c.ID }
func (c *MySQLConnection) GetName() string { return c.Name }
func (c *MySQLConnection) SetName(name string) { c.Name = name }
func (c *MySQLConnection) GetType() DatabaseType { return DatabaseTypeMySQL }
func (c *MySQLConnection) GetDSN() string {
    return fmt.Sprintf("%s@tcp(%s:%d)/%s", c.Username, c.Host, c.Port, c.Database)
}
func (c *MySQLConnection) GetDSNWithPassword() string {
    return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", c.Username, c.Password, c.Host, c.Port, c.Database)
}
func (c *MySQLConnection) Redact() string {
    return fmt.Sprintf("%s (***@%s:%d/%s)", c.Name, c.Host, c.Port, c.Database)
}

// Validate 验证连接参数（REQ-CONN-010）
func (c *MySQLConnection) Validate() error {
    var errs []error

    if c.Name == "" {
        errs = append(errs, fmt.Errorf("name is required"))
    }
    if c.Host == "" {
        errs = append(errs, fmt.Errorf("host is required"))
    }
    if c.Port < 1 || c.Port > 65535 {
        errs = append(errs, fmt.Errorf("port must be between 1 and 65535, got %d", c.Port))
    }
    if c.Database == "" {
        errs = append(errs, fmt.Errorf("database is required"))
    }
    if c.Username == "" {
        errs = append(errs, fmt.Errorf("username is required"))
    }

    if len(errs) > 0 {
        return fmt.Errorf("validation failed: %w", joinErrors(errs))
    }
    return nil
}

// Test 测试连接（REQ-CONN-003, REQ-CONN-004, REQ-CONN-005）
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

    // 获取数据库版本（REQ-CONN-004）
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
```

#### 4.1.3 OracleConnection 结构

```go
// internal/domain/connection/oracle.go
package connection

// OracleConnection Oracle 连接配置（spec.md 3.2.2）
type OracleConnection struct {
    ID   string `json:"id"`
    Name string `json:"name"`

    // 连接参数
    Host        string `json:"host"`         // 主机地址
    Port        int    `json:"port"`         // 端口（默认 1521）
    ServiceName string `json:"service_name"` // 服务名
    SID         string `json:"sid"`          // SID（与 ServiceName 二选一）
    Username    string `json:"username"`
    Password    string `json:"-"`            // 存储到 keyring

    // 元数据
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

// 实现 Connection 接口
// ... (类似 MySQLConnection)
```

#### 4.1.4 SQLServerConnection 结构

```go
// internal/domain/connection/sqlserver.go
package connection

// SQLServerConnection SQL Server 连接配置（spec.md 3.2.2）
type SQLServerConnection struct {
    ID   string `json:"id"`
    Name string `json:"name"`

    // 连接参数
    Host                   string `json:"host"`
    Port                   int    `json:"port"`                     // 端口（默认 1433）
    Database               string `json:"database"`
    Username               string `json:"username"`
    Password               string `json:"-"`                        // 存储到 keyring
    TrustServerCertificate bool   `json:"trust_server_certificate"` // 信任服务器证书

    // 元数据
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

// 实现 Connection 接口
// ... (类似 MySQLConnection)
```

#### 4.1.5 PostgreSQLConnection 结构

```go
// internal/domain/connection/postgresql.go
package connection

// PostgreSQLConnection PostgreSQL 连接配置（spec.md 3.2.2）
type PostgreSQLConnection struct {
    ID   string `json:"id"`
    Name string `json:"name"`

    // 连接参数
    Host     string `json:"host"`
    Port     int    `json:"port"`     // 端口（默认 5432）
    Database string `json:"database"`
    Username string `json:"username"`
    Password string `json:"-"`        // 存储到 keyring
    SSLMode  string `json:"ssl_mode"` // disable/allow/prefer/require/verify-ca/verify-full

    // 元数据
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

// 实现 Connection 接口
// ... (类似 MySQLConnection)
```

---

### 4.2 模板领域（template）

#### 4.2.1 Template 结构

```go
// internal/domain/template/template.go
package template

import (
    "time"
)

// Template 场景模板（spec.md 3.3.2）
type Template struct {
    // 基本信息
    ID          string    `json:"id"`          // 模板 ID（sysbench-oltp-read-write）
    Name        string    `json:"name"`        // 模板名称
    Description string    `json:"description"` // 描述
    Tool        string    `json:"tool"`        // sysbench/swingbench/hammerdb

    // 数据库类型
    DatabaseTypes []string `json:"database_types"` // ["mysql", "postgresql"]

    // 参数定义
    Parameters map[string]Parameter `json:"parameters"` // 参数名 → 参数定义

    // 命令模板
    CommandTemplate CommandTemplate `json:"command_template"`

    // 输出解析
    OutputParser OutputParser `json:"output_parser"`

    // 元数据
    Version   string    `json:"version"`    // 模板版本
    IsBuiltin bool      `json:"is_builtin"` // 是否为内置模板
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

// Parameter 参数定义（spec.md 3.3.2）
type Parameter struct {
    Type        string      `json:"type"`         // integer/string/boolean/float
    Label       string      `json:"label"`        // 显示标签
    Description string      `json:"description"`  // 参数描述
    Default     interface{} `json:"default"`      // 默认值
    Min         *float64    `json:"min,omitempty"`   // 最小值（数值类型）
    Max         *float64    `json:"max,omitempty"`   // 最大值（数值类型）
    Options     []string    `json:"options,omitempty"` // 选项（枚举类型）
    Required    bool        `json:"required"`     // 是否必需
}

// CommandTemplate 命令模板（spec.md 3.3.2）
type CommandTemplate struct {
    Prepare string `json:"prepare"` // prepare 阶段命令模板
    Run     string `json:"run"`     // run 阶段命令模板
    Cleanup string `json:"cleanup"` // cleanup 阶段命令模板
}

// OutputParser 输出解析器（spec.md 3.3.2）
type OutputParser struct {
    Type     string            `json:"type"`     // regex/json/xml
    Patterns map[string]string `json:"patterns"` // 指标名 → 正则表达式
}

// Validate 验证模板（REQ-TMPL-004）
func (t *Template) Validate() error {
    if t.ID == "" {
        return fmt.Errorf("template id is required")
    }
    if t.Name == "" {
        return fmt.Errorf("template name is required")
    }
    if t.Tool == "" {
        return fmt.Errorf("tool is required")
    }
    if len(t.DatabaseTypes) == 0 {
        return fmt.Errorf("database_types is required")
    }
    if len(t.Parameters) == 0 {
        return fmt.Errorf("parameters is required")
    }
    return nil
}

// SupportsDatabase 检查是否支持指定数据库
func (t *Template) SupportsDatabase(dbType string) bool {
    for _, dt := range t.DatabaseTypes {
        if dt == dbType {
            return true
        }
    }
    return false
}

// GetParameterValue 获取参数值（带默认值）
func (t *Template) GetParameterValue(name string, userValue interface{}) interface{} {
    param, exists := t.Parameters[name]
    if !exists {
        return nil
    }

    // 如果用户提供了值，使用用户值
    if userValue != nil {
        return userValue
    }

    // 否则使用默认值
    return param.Default
}
```

---

### 4.3 执行领域（execution）

#### 4.3.1 RunState 状态机

```go
// internal/domain/execution/state.go
package execution

// RunState 运行状态（spec.md 3.4.2）
type RunState string

const (
    StatePending      RunState = "pending"        // 已创建，待执行
    StatePreparing    RunState = "preparing"      // 准备数据中
    StatePrepared     RunState = "prepared"       // 准备完成
    StateWarmingUp    RunState = "warming_up"     // 预热中
    StateRunning      RunState = "running"        // 正式运行
    StateCompleted    RunState = "completed"      // 正常完成
    StateFailed       RunState = "failed"         // 执行失败
    StateCancelled    RunState = "cancelled"      // 用户取消
    StateTimeout      RunState = "timeout"        // 超时
    StateForceStopped RunState = "force_stopped"  // 强制停止
)

// IsValid 检查状态是否有效
func (s RunState) IsValid() bool {
    switch s {
    case StatePending, StatePreparing, StatePrepared, StateWarmingUp,
         StateRunning, StateCompleted, StateFailed, StateCancelled,
         StateTimeout, StateForceStopped:
        return true
    default:
        return false
    }
}

// IsTerminal 检查是否为终态（REQ-EXEC-008）
func (s RunState) IsTerminal() bool {
    return s == StateCompleted || s == StateFailed ||
           s == StateCancelled || s == StateTimeout || s == StateForceStopped
}

// CanTransitionTo 检查是否可以转换到目标状态（spec.md 3.4.2）
func (s RunState) CanTransitionTo(target RunState) bool {
    // 定义合法的状态转换
    transitions := map[RunState][]RunState{
        StatePending:   {StatePreparing, StateCancelled},
        StatePreparing: {StatePrepared, StateFailed, StateCancelled, StateTimeout},
        StatePrepared:  {StateWarmingUp, StateCancelled},
        StateWarmingUp: {StateRunning, StateFailed, StateCancelled, StateTimeout},
        StateRunning:   {StateCompleted, StateFailed, StateCancelled, StateTimeout, StateForceStopped},
    }

    allowed, ok := transitions[s]
    if !ok {
        return false
    }

    for _, state := range allowed {
        if state == target {
            return true
        }
    }
    return false
}

// String 实现 Stringer 接口
func (s RunState) String() string {
    return string(s)
}
```

#### 4.3.2 BenchmarkTask 结构

```go
// internal/domain/execution/task.go
package execution

import (
    "time"
)

// BenchmarkTask 压测任务（spec.md 3.4.1）
type BenchmarkTask struct {
    // 基本信息
    ID   string    `json:"id"`    // UUID
    Name string    `json:"name"`  // 任务名称

    // 关联
    ConnectionID string `json:"connection_id"` // 连接 ID
    TemplateID   string `json:"template_id"`   // 模板 ID

    // 参数覆盖
    Parameters map[string]interface{} `json:"parameters"` // 参数名 → 参数值

    // 执行选项（spec.md 3.4.1）
    Options TaskOptions `json:"options"`

    // 标签
    Tags []string `json:"tags"`

    // 元数据
    CreatedAt time.Time `json:"created_at"`
}

// TaskOptions 任务选项（spec.md 3.4.1）
type TaskOptions struct {
    // 数据准备
    SkipPrepare    bool          `json:"skip_prepare"`     // 跳过数据准备
    SkipCleanup    bool          `json:"skip_cleanup"`     // 跳过数据清理

    // 预热
    WarmupTime     int           `json:"warmup_time"`      // 预热时长（秒）

    // 采样
    SampleInterval time.Duration `json:"sample_interval"`  // 采样间隔（默认 1s）

    // 执行模式
    DryRun         bool          `json:"dry_run"`          // 仅显示命令不执行（REQ-EXEC-010）

    // 超时
    PrepareTimeout time.Duration `json:"prepare_timeout"`  // 准备阶段超时（默认 30m）
    RunTimeout     time.Duration `json:"run_timeout"`      // 运行阶段超时（默认 24h）
}

// Validate 验证任务配置
func (t *BenchmarkTask) Validate() error {
    if t.ID == "" {
        return fmt.Errorf("task id is required")
    }
    if t.Name == "" {
        return fmt.Errorf("task name is required")
    }
    if t.ConnectionID == "" {
        return fmt.Errorf("connection_id is required")
    }
    if t.TemplateID == "" {
        return fmt.Errorf("template_id is required")
    }

    // 互斥检查：预热时长 < 运行时长（spec.md 3.4.4）
    if t.Options.WarmupTime > 0 && t.Options.RunTimeout > 0 {
        warmupDuration := time.Duration(t.Options.WarmupTime) * time.Second
        if warmupDuration >= t.Options.RunTimeout {
            return fmt.Errorf("warmup_time (%d) must be less than run_timeout (%s)",
                t.Options.WarmupTime, t.Options.RunTimeout)
        }
    }

    return nil
}
```

#### 4.3.3 Run 结构

```go
// internal/domain/execution/run.go
package execution

import (
    "time"
)

// Run 运行记录（spec.md 3.6.1）
type Run struct {
    // 基本信息
    ID    string    `json:"id"`    // UUID
    TaskID string   `json:"task_id"` // 关联任务 ID

    // 状态（spec.md 3.4.2）
    State RunState `json:"state"`

    // 时间戳
    CreatedAt   time.Time     `json:"created_at"`    // 创建时间
    StartedAt   *time.Time    `json:"started_at"`    // 开始时间
    CompletedAt *time.Time    `json:"completed_at"`  // 完成时间
    Duration    *time.Duration `json:"duration"`     // 运行时长

    // 结果
    Result       *BenchmarkResult `json:"result,omitempty"`       // 解析后的结果
    ErrorMessage string           `json:"error_message,omitempty"` // 错误信息

    // 工作目录（用于存储日志等）
    WorkDir string `json:"work_dir,omitempty"`
}

// BenchmarkResult 压测结果（spec.md 3.5.1）
type BenchmarkResult struct {
    // 基本信息
    RunID string `json:"run_id"`

    // 核心指标（spec.md 3.5.2）
    TPSCalculated     float64 `json:"tps_calculated"`      // 计算得到的 TPS
    LatencyAvg        float64 `json:"latency_avg_ms"`      // 平均延迟（ms）
    LatencyP95        float64 `json:"latency_p95_ms"`      // 95分位延迟（ms）
    LatencyP99        float64 `json:"latency_p99_ms"`      // 99分位延迟（ms）
    ErrorCount        int64   `json:"error_count"`         // 错误总数
    ErrorRate         float64 `json:"error_rate_percent"`  // 错误率（%）

    // 统计信息
    Duration          time.Duration `json:"duration"`              // 运行时长
    TotalTransactions int64         `json:"total_transactions"`   // 总事务数
    TotalQueries      int64         `json:"total_queries,omitempty"` // 总查询数

    // 时间序列数据
    TimeSeries []MetricSample `json:"time_series,omitempty"` // 时间序列指标
}

// MetricSample 指标样本（spec.md 3.5.1）
type MetricSample struct {
    Timestamp   time.Time `json:"timestamp"`   // 采样时间
    Phase       string    `json:"phase"`       // 阶段：warmup/run/cooldown
    TPS         float64   `json:"tps"`         // 每秒事务数
    QPS         float64   `json:"qps,omitempty"` // 每秒查询数
    LatencyAvg  float64   `json:"latency_avg_ms"` // 平均延迟（ms）
    LatencyP95  float64   `json:"latency_p95_ms"` // 95分位延迟（ms）
    LatencyP99  float64   `json:"latency_p99_ms"` // 99分位延迟（ms）
    ErrorRate   float64   `json:"error_rate_percent"` // 错误率（%）
}

// IsCompleted 检查运行是否已完成（终态）
func (r *Run) IsCompleted() bool {
    return r.State.IsTerminal()
}

// SetState 设置状态（带状态转换验证）
func (r *Run) SetState(newState RunState) error {
    if !r.State.CanTransitionTo(newState) {
        return fmt.Errorf("invalid state transition: %s -> %s", r.State, newState)
    }
    r.State = newState
    return nil
}

// CalculateDuration 计算运行时长
func (r *Run) CalculateDuration() {
    if r.StartedAt != nil && r.CompletedAt != nil {
        duration := r.CompletedAt.Sub(*r.StartedAt)
        r.Duration = &duration
    }
}
```

---

### 4.4 指标领域（metric）

#### 4.4.1 MetricCalculator

```go
// internal/domain/metric/calculator.go
package metric

import (
    "sort"
)

// Calculator 指标计算器（spec.md 3.5.2）
type Calculator struct {
    samples []float64
}

// NewCalculator 创建计算器
func NewCalculator() *Calculator {
    return &Calculator{
        samples: make([]float64, 0),
    }
}

// AddSample 添加样本
func (c *Calculator) AddSample(value float64) {
    c.samples = append(c.samples, value)
}

// Calculate 计算统计数据
func (c *Calculator) Calculate() *Statistics {
    if len(c.samples) == 0 {
        return &Statistics{}
    }

    // 排序
    sorted := make([]float64, len(c.samples))
    copy(sorted, c.samples)
    sort.Float64s(sorted)

    // 计算
    stats := &Statistics{
        Count: int64(len(sorted)),
        Min:   sorted[0],
        Max:   sorted[len(sorted)-1],
    }

    // Sum 和 Avg
    var sum float64
    for _, v := range sorted {
        sum += v
    }
    stats.Sum = sum
    stats.Avg = sum / float64(len(sorted))

    // P50 (Median)
    stats.P50 = percentile(sorted, 0.50)

    // P95
    stats.P95 = percentile(sorted, 0.95)

    // P99
    stats.P99 = percentile(sorted, 0.99)

    return stats
}

// Statistics 统计数据
type Statistics struct {
    Count int64   `json:"count"` // 样本数
    Min   float64 `json:"min"`   // 最小值
    Max   float64 `json:"max"`   // 最大值
    Sum   float64 `json:"sum"`   // 总和
    Avg   float64 `json:"avg"`   // 平均值
    P50   float64 `json:"p50"`   // 50分位（中位数）
    P95   float64 `json:"p95"`   // 95分位
    P99   float64 `json:"p99"`   // 99分位
}

// percentile 计算百分位
func percentile(sorted []float64, p float64) float64 {
    if len(sorted) == 0 {
        return 0
    }

    index := p * float64(len(sorted)-1)
    lower := int(index)
    upper := lower + 1

    if upper >= len(sorted) {
        return sorted[len(sorted)-1]
    }

    weight := index - float64(lower)
    return sorted[lower]*(1-weight) + sorted[lower]*weight
}
```

---

## 5. 接口设计

### 5.1 UseCase 层接口（由用例定义，基础设施实现）

#### 5.1.1 ConnectionUseCase

```go
// internal/app/usecase/connection_usecase.go
package usecase

import (
    "context"
    "github.com/whhaicheng/DB-BenchMind/internal/domain/connection"
)

// ConnectionUseCase 连接管理用例（REQ-CONN-001 ~ REQ-CONN-010）
type ConnectionUseCase interface {
    // CreateConnection 创建连接（REQ-CONN-001）
    CreateConnection(ctx context.Context, conn connection.Connection) error

    // UpdateConnection 更新连接（REQ-CONN-008）
    UpdateConnection(ctx context.Context, conn connection.Connection) error

    // DeleteConnection 删除连接（REQ-CONN-009）
    DeleteConnection(ctx context.Context, id string) error

    // ListConnections 列出所有连接
    ListConnections(ctx context.Context) ([]connection.Connection, error)

    // GetConnectionByID 根据 ID 获取连接
    GetConnectionByID(ctx context.Context, id string) (connection.Connection, error)

    // TestConnection 测试连接（REQ-CONN-003, REQ-CONN-004, REQ-CONN-005）
    TestConnection(ctx context.Context, id string) (*connection.TestResult, error)

    // SavePassword 保存密码到 keyring（REQ-CONN-006）
    SavePassword(ctx context.Context, connID, password string) error

    // GetPassword 从 keyring 获取密码
    GetPassword(ctx context.Context, connID string) (string, error)

    // DeletePassword 从 keyring 删除密码
    DeletePassword(ctx context.Context, connID string) error
}
```

#### 5.1.2 BenchmarkUseCase

```go
// internal/app/usecase/benchmark_usecase.go
package usecase

import (
    "context"
    "github.com/whhaicheng/DB-BenchMind/internal/domain/execution"
    "github.com/whhaicheng/DB-BenchMind/internal/domain/metric"
)

// BenchmarkUseCase 压测执行用例（REQ-EXEC-001 ~ REQ-EXEC-010）
type BenchmarkUseCase interface {
    // StartTask 启动压测任务（REQ-EXEC-001, REQ-EXEC-002）
    StartTask(ctx context.Context, task *execution.BenchmarkTask) (*execution.Run, error)

    // StopRun 停止运行（REQ-EXEC-006, REQ-EXEC-007）
    StopRun(ctx context.Context, runID string, force bool) error

    // GetRunStatus 获取运行状态（REQ-EXEC-003）
    GetRunStatus(ctx context.Context, runID string) (*execution.Run, error)

    // GetRealtimeMetrics 获取实时指标（REQ-EXEC-004, REQ-METRIC-001）
    GetRealtimeMetrics(ctx context.Context, runID string) (<-chan metric.MetricSample, error)

    // GetRunLogs 获取运行日志（REQ-EXEC-005）
    GetRunLogs(ctx context.Context, runID string) (<-chan LogEntry, error)

    // GetRunResult 获取运行结果（REQ-STORAGE-005）
    GetRunResult(ctx context.Context, runID string) (*execution.BenchmarkResult, error)

    // ListRuns 列出运行记录（REQ-STORAGE-004）
    ListRuns(ctx context.Context, opts ListRunsOptions) ([]*execution.Run, error)

    // DeleteRun 删除运行记录（REQ-STORAGE-006）
    DeleteRun(ctx context.Context, runID string) error
}

// ListRunsOptions 查询选项
type ListRunsOptions struct {
    Limit       int                      `json:"limit"`
    Offset      int                      `json:"offset"`
    StateFilter *execution.RunState      `json:"state_filter,omitempty"`
    TaskID      string                   `json:"task_id,omitempty"`
    SortBy      string                   `json:"sort_by"`      // created_at, started_at, duration
    SortOrder   string                   `json:"sort_order"`   // ASC, DESC
}

// LogEntry 日志条目
type LogEntry struct {
    Timestamp time.Time `json:"timestamp"`
    Stream    string    `json:"stream"`  // stdout/stderr
    Content   string    `json:"content"`
}
```

#### 5.1.3 ReportUseCase

```go
// internal/app/usecase/report_usecase.go
package usecase

import (
    "context"
)

// ReportUseCase 报告生成用例（REQ-RPT-001 ~ REQ-RPT-010）
type ReportUseCase interface {
    // GenerateReport 生成报告（REQ-RPT-002 ~ REQ-RPT-005）
    GenerateReport(ctx context.Context, req GenerateReportRequest) (string, error)

    // GetSupportedFormats 获取支持的格式
    GetSupportedFormats() []string

    // GetReportPath 获取报告路径
    GetReportPath(ctx context.Context, runID, format string) (string, error)
}

// GenerateReportRequest 报告生成请求
type GenerateReportRequest struct {
    RunID      string            `json:"run_id"`       // 运行 ID
    Format     string            `json:"format"`       // md/html/json/pdf
    OutputPath string            `json:"output_path"`  // 输出路径（可选）
    Options    map[string]string `json:"options"`      // 额外选项
}
```

#### 5.1.4 ComparisonUseCase

```go
// internal/app/usecase/comparison_usecase.go
package usecase

import (
    "context"
)

// ComparisonUseCase 结果对比用例（REQ-COMP-001 ~ REQ-COMP-007）
type ComparisonUseCase interface {
    // CompareRuns 对比两次运行（REQ-COMP-002）
    CompareRuns(ctx context.Context, baselineID, compareID string) (*ComparisonResult, error)

    // CompareWithBaseline 对比基线（REQ-COMP-005）
    CompareWithBaseline(ctx context.Context, baselineID string, runIDs []string) (*BaselineComparison, error)

    // AnalyzeTrend 趋势分析（REQ-COMP-006）
    AnalyzeTrend(ctx context.Context, taskID string) (*TrendAnalysis, error)

    // ExportComparisonReport 导出对比报告（REQ-COMP-007）
    ExportComparisonReport(ctx context.Context, comparison *ComparisonResult, format string) (string, error)
}

// ComparisonResult 对比结果
type ComparisonResult struct {
    Baseline    *RunSummary         `json:"baseline"`
    Compare     *RunSummary         `json:"compare"`
    Metrics     []MetricDifference  `json:"metrics"`
    Conclusion  string              `json:"conclusion"`
}

// MetricDifference 指标差异
type MetricDifference struct {
    Name        string  `json:"name"`         // 指标名称
    Baseline    float64 `json:"baseline"`     // 基线值
    Compare     float64 `json:"compare"`      // 对比值
    Change      float64 `json:"change"`       // 绝对变化
    ChangePercent float64 `json:"change_percent"` // 百分比变化
    Trend       string  `json:"trend"`        // ↑↓→
}

// RunSummary 运行摘要
type RunSummary struct {
    ID          string    `json:"id"`
    Name        string    `json:"name"`
    State       string    `json:"state"`
    CompletedAt time.Time `json:"completed_at"`
    TPS         float64   `json:"tps"`
    LatencyP95  float64   `json:"latency_p95_ms"`
    ErrorRate   float64   `json:"error_rate_percent"`
}
```

---

### 5.2 仓储接口（由用例定义，数据库实现）

```go
// internal/app/usecase/repository.go（统一管理所有仓储接口）
package usecase

import (
    "context"
    "github.com/whhaicheng/DB-BenchMind/internal/domain/connection"
    "github.com/whhaicheng/DB-BenchMind/internal/domain/execution"
    "github.com/whhaicheng/DB-BenchMind/internal/domain/template"
)

// ConnectionRepository 连接仓储接口
type ConnectionRepository interface {
    Save(ctx context.Context, conn connection.Connection) error
    FindByID(ctx context.Context, id string) (connection.Connection, error)
    FindAll(ctx context.Context) ([]connection.Connection, error)
    Delete(ctx context.Context, id string) error
    ExistsByName(ctx context.Context, name string) (bool, error)
}

// TemplateRepository 模板仓储接口
type TemplateRepository interface {
    Save(ctx context.Context, tmpl *template.Template) error
    FindByID(ctx context.Context, id string) (*template.Template, error)
    FindAll(ctx context.Context) ([]*template.Template, error)
    FindBuiltin(ctx context.Context) ([]*template.Template, error)
    Delete(ctx context.Context, id string) error
}

// RunRepository 运行仓储接口
type RunRepository interface {
    Save(ctx context.Context, run *execution.Run) error
    FindByID(ctx context.Context, id string) (*execution.Run, error)
    FindAll(ctx context.Context, opts FindOptions) ([]*execution.Run, error)
    UpdateState(ctx context.Context, id string, state execution.RunState) error
    SaveMetricSample(ctx context.Context, runID string, sample execution.MetricSample) error
    SaveLogEntry(ctx context.Context, runID string, entry LogEntry) error
    Delete(ctx context.Context, id string) error
}

// FindOptions 查询选项
type FindOptions struct {
    Limit       int
    Offset      int
    StateFilter *execution.RunState
    TaskID      string
    SortBy      string // created_at, started_at, duration
    SortOrder   string // ASC, DESC
}
```

---

### 5.3 pkg/benchmark.Adapter 接口（对外可复用）

```go
// pkg/benchmark/adapter.go
package benchmark

import (
    "context"
    "io"
    "time"
)

// Adapter 统一工具适配器接口（spec.md 4.1）
type Adapter interface {
    // 基本信息
    ToolName() string
    Version(ctx context.Context) (string, error)
    Available(ctx context.Context) bool

    // 命令构建
    BuildPrepareCommand(ctx context.Context, cfg *Config) (*exec.Cmd, error)
    BuildRunCommand(ctx context.Context, cfg *Config) (*exec.Cmd, error)
    BuildCleanupCommand(ctx context.Context, cfg *Config) (*exec.Cmd, error)

    // 输出解析
    ParseRunOutput(stdout, stderr string) (*Result, error)

    // 实时采集
    StartRealtimeCollection(ctx context.Context, reader io.Reader) (<-chan Sample, error)

    // 配置验证
    ValidateConfig(ctx context.Context, cfg *Config) error
}

// Config 执行配置
type Config struct {
    Connection any       `json:"connection"` // connection.Connection
    Template   *Template `json:"template"`
    Parameters map[string]any `json:"parameters"`
    WorkDir    string    `json:"work_dir"`
    Timeout    time.Duration `json:"timeout"`
}

// Template 模板定义（简化版）
type Template struct {
    ID             string            `json:"id"`
    Name           string            `json:"name"`
    Tool           string            `json:"tool"`
    DatabaseTypes  []string          `json:"database_types"`
    Parameters     map[string]Parameter `json:"parameters"`
    CommandTemplate CommandTemplate   `json:"command_template"`
}

// Parameter 参数定义
type Parameter struct {
    Type    string      `json:"type"`
    Default interface{} `json:"default"`
}

// CommandTemplate 命令模板
type CommandTemplate struct {
    Prepare string `json:"prepare"`
    Run     string `json:"run"`
    Cleanup string `json:"cleanup"`
}

// Result 解析结果
type Result struct {
    Success    bool                   `json:"success"`
    Metrics    map[string]float64     `json:"metrics"`
    RawOutput  string                 `json:"raw_output"`
    TimeSeries []Sample               `json:"time_series,omitempty"`
}

// Sample 实时样本
type Sample struct {
    Timestamp time.Time              `json:"timestamp"`
    Metrics   map[string]float64     `json:"metrics"`
}
```

---

## 6. 技术决策记录

### 6.1 技术选型决策

| ID | 决策 | 理由 | 替代方案 | 风险 | 缓解措施 |
|----|------|------|---------|------|---------|
| ADR-001 | 使用 Fyne 作为 GUI 框架 | 纯 Go、跨平台、单二进制打包 | Qt (CGO), GTK (CGO), Walk (仅 Windows) | GUI 性能可能不如原生 | 异步加载、虚拟滚动 |
| ADR-002 | 使用 modernc.org/sqlite | 无 CGO、交叉编译简单 | mattn/go-sqlite3 (需要 CGO) | 并发性能可能略低 | WAL 模式、单连接 |
| ADR-003 | 不使用 ORM | 遵循反抽象门禁 | GORM, sqlx | 手写 SQL 易错 | 参数化查询、单元测试 |
| ADR-004 | keyring + 加密文件降级 | 系统标准 + 可用性保证 | 仅 keyring、仅加密文件 | keyring 不可用 | 自动降级检测 |
| ADR-005 | 直接使用 log/slog | 标准库、结构化日志 | zap, logrus | 功能相对简单 | 足够使用 |

### 6.2 架构决策

| ID | 决策 | 理由 | 影响 |
|----|------|------|------|
| ADR-101 | 采用 DDD + Clean Architecture | 业务逻辑独立、易于测试 | 需要明确的分层 |
| ADR-102 | 领域层无外部依赖 | 核心逻辑纯、易于测试 | 需要适配器模式 |
| ADR-103 | 用例定义接口，基础设施实现 | 依赖倒置、灵活性高 | 接口定义工作量大 |
| ADR-104 | 单一 SQLite 数据库 | 简单、嵌入式 | 无分布式支持 |

---

## 7. 分阶段实施计划

### Phase 1: 基础设施与连接管理（Week 1-2）

#### 目标
搭建项目骨架，实现四种数据库连接管理

#### 详细任务清单

**1. 项目初始化**
- [ ] 创建完整目录结构
- [ ] 初始化 go.mod
- [ ] 创建 Makefile（build, test, lint, check）
- [ ] 配置 .golangci.yml
- [ ] 创建 .gitignore
- [ ] 创建 README.md

**2. 数据库层**
- [ ] 编写 schema.sql（所有表）
- [ ] 实现 SQLite 初始化函数
- [ ] 定义 ConnectionRepository 接口
- [ ] 实现 MySQLConnection 模型
- [ ] 实现 OracleConnection 模型
- [ ] 实现 SQLServerConnection 模型
- [ ] 实现 PostgreSQLConnection 模型
- [ ] 实现 SQLiteConnectionRepository
- [ ] 单元测试：所有连接模型
- [ ] 单元测试：ConnectionRepository

**3. 密钥管理**
- [ ] 定义 KeyringProvider 接口
- [ ] 实现 GoKeyringProvider
- [ ] 实现加密文件降级方案
- [ ] 单元测试：密码存取

**4. 用例层**
- [ ] 定义 ConnectionUseCase 接口
- [ ] 实现 ConnectionUseCase
- [ ] 单元测试：ConnectionUseCase

**5. GUI - 连接管理页面**
- [ ] 创建主窗口框架
- [ ] 创建连接列表页面
- [ ] 创建连接表单（4种类型）
- [ ] 实现添加功能
- [ ] 实现编辑功能
- [ ] 实现删除功能
- [ ] 实现测试连接功能（异步）
- [ ] 实现密码输入框（掩码）

**6. 文档**
- [ ] 创建 .specify/steering/ 目录
- [ ] 编写 product.md
- [ ] 编写 architecture.md
- [ ] 编写 testing.md
- [ ] 编写 decisions.md

**7. 集成测试**
- [ ] 真实 MySQL 连接测试
- [ ] 真实 PostgreSQL 连接测试

#### 验收标准
- [x] 目录结构完整
- [x] 能添加 4 种数据库连接
- [x] 密码安全存储
- [x] 连接测试显示明确结果
- [x] 所有单元测试通过（覆盖率 > 80%）
- [x] golangci-lint 零错误
- [x] 文档完整

---

### Phase 2: 模板系统与任务配置（Week 3-4）

#### 目标
实现模板管理与任务配置

#### 详细任务清单

**1. 模板领域**
- [ ] 实现 Template 模型
- [ ] 实现 Parameter 模型
- [ ] 实现 CommandTemplate 模型
- [ ] 实现 OutputParser 模型
- [ ] 实现模板验证逻辑
- [ ] 单元测试：Template 模型

**2. 内置模板**
- [ ] 创建 contracts/templates/ 目录
- [ ] 编写 sysbench-oltp-read-write.json
- [ ] 编写 sysbench-oltp-read-only.json
- [ ] 编写 sysbench-oltp-write-only.json
- [ ] 编写 swingbench-soe.json
- [ ] 编写 swingbench-calling.json
- [ ] 编写 hammerdb-tpcc.json
- [ ] 编写 hammerdb-tpcb.json
- [ ] 验证所有模板

**3. 模板仓储**
- [ ] 定义 TemplateRepository 接口
- [ ] 实现 SQLiteTemplateRepository
- [ ] 实现模板加载逻辑
- [ ] 单元测试：TemplateRepository

**4. 模板用例**
- [ ] 定义 TemplateUseCase 接口
- [ ] 实现 TemplateUseCase
- [ ] 实现模板导入功能
- [ ] 实现模板导出功能
- [ ] 单元测试：TemplateUseCase

**5. 任务配置**
- [ ] 实现 BenchmarkTask 模型
- [ ] 实现 TaskOptions 模型
- [ ] 实现配置快照机制
- [ ] 单元测试：Task 配置

**6. GUI - 模板与任务页面**
- [ ] 创建模板列表页面
- [ ] 创建模板详情页面
- [ ] 实现模板导入对话框
- [ ] 实现模板导出对话框
- [ ] 创建任务配置页面
- [ ] 实现连接选择器
- [ ] 实现模板选择器
- [ ] 实现动态参数表单
- [ ] 实现参数验证

#### 验收标准
- [x] 7 个内置模板可见且正确
- [x] 能导入/导出模板
- [x] 任务配置表单动态生成
- [x] 参数验证正确
- [x] 单元测试覆盖率 > 80%

---

### Phase 3: 工具适配器与执行编排（Week 5-7）

#### 目标
实现三个工具的完整适配器和执行编排器

#### 详细任务清单

**1. 适配器接口**
- [ ] 定义 pkg/benchmark/Adapter 接口
- [ ] 定义 Config 结构
- [ ] 定义 Result 结构
- [ ] 定义 Sample 结构

**2. Sysbench 适配器**
- [ ] 实现 SysbenchAdapter
- [ ] 实现 BuildPrepareCommand
- [ ] 实现 BuildRunCommand
- [ ] 实现 BuildCleanupCommand
- [ ] 实现 ParseRunOutput（正则）
- [ ] 实现 StartRealtimeCollection
- [ ] 单元测试：命令构建
- [ ] 单元测试：输出解析

**3. Swingbench 适配器**
- [ ] 实现 SwingbenchAdapter
- [ ] 实现 BuildPrepareCommand
- [ ] 实现 BuildRunCommand
- [ ] 实现 BuildCleanupCommand
- [ ] 实现 ParseRunOutput（XML）
- [ ] 实现 TPM → TPS 转换
- [ ] 单元测试

**4. HammerDB 适配器**
- [ ] 实现 HammerDBAdapter
- [ ] 实现 TCL 脚本生成
- [ ] 实现 BuildPrepareCommand
- [ ] 实现 BuildRunCommand
- [ ] 实现 BuildCleanupCommand
- [ ] 实现 ParseRunOutput（TCL）
- [ ] 单元测试

**5. 适配器注册表**
- [ ] 实现 AdapterRegistry
- [ ] 实现适配器注册逻辑
- [ ] 实现适配器查找逻辑

**6. 执行领域**
- [ ] 实现 RunState 状态机
- [ ] 实现 Run 模型
- [ ] 实现 BenchmarkTask 模型
- [ ] 单元测试：状态转换

**7. 执行编排器**
- [ ] 定义 BenchmarkUseCase 接口
- [ ] 实现 Executor
- [ ] 实现预检查（工具/连接/磁盘/参数）
- [ ] 实现阶段执行（prepare → warmup → run → cleanup）
- [ ] 实现状态管理
- [ ] 实现优雅停止（SIGTERM → SIGKILL）
- [ ] 实现超时处理
- [ ] 单元测试：Executor
- [ ] 集成测试：完整流程

**8. 实时监控**
- [ ] 实现日志采集（stdout/stderr）
- [ ] 实现指标采集（1s 间隔）
- [ ] 实现数据存储（run_logs, metric_samples）
- [ ] 实现通道管理

**9. RunRepository**
- [ ] 定义 RunRepository 接口
- [ ] 实现 SQLiteRunRepository
- [ ] 实现 Save 方法
- [ ] 实现 FindByID 方法
- [ ] 实现 FindAll 方法
- [ ] 实现 UpdateState 方法
- [ ] 实现 SaveMetricSample 方法
- [ ] 实现 SaveLogEntry 方法

**10. GUI - 运行监控页面**
- [ ] 创建监控页面
- [ ] 实现实时指标显示（TPS/延迟/错误率）
- [ ] 实现实时日志滚动
- [ ] 实现启动按钮
- [ ] 实现停止按钮
- [ ] 实现取消按钮
- [ ] 实现状态指示器

#### 验收标准
- [x] 三个工具能跑通完整流程
- [x] 实时监控正常工作
- [x] 优雅停止正确
- [x] 超时处理正确
- [x] 状态转换正确
- [x] 集成测试通过

---

### Phase 4-8

（后续阶段保持原有计划，此处省略以节省篇幅）

---

## 8. 测试策略

### 8.1 测试金字塔

``                    /\
                   /  \
                  / E2E \         10% - 端到端测试
                 /--------\
                /          \
               / Integration\    30% - 集成测试
              /--------------\
             /                \
            /    Unit Tests     \  60% - 单元测试
           /--------------------\
```

### 8.2 测试覆盖率要求

| 层级 | 目标覆盖率 | 必须覆盖 |
|------|-----------|---------|
| domain/ | > 90% | 所有业务逻辑 |
| usecase/ | > 85% | 所有用例 |
| infra/database/ | > 80% | 所有仓储方法 |
| infra/adapter/ | > 75% | 命令构建、输出解析 |
| transport/ui/ | > 40% | 主要逻辑（手动为主） |

### 8.3 测试示例

#### 8.3.1 单元测试（表格驱动）

```go
// internal/domain/connection/mysql_test.go
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
                Name: "test", Host: "localhost", Port: 3306,
                Database: "testdb", Username: "root",
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
            name: "missing required fields",
            conn: &MySQLConnection{
                Name: "", Host: "", Port: 0,
            },
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.conn.Validate()
            if (err != nil) != tt.wantErr {
                t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if tt.wantErr && tt.errMsg != "" && err.Error() != tt.errMsg {
                t.Errorf("Validate() error = %v, want %v", err.Error(), tt.errMsg)
            }
        })
    }
}
```

#### 8.3.2 集成测试（真实数据库）

```go
// internal/infra/database/repository/connection_repo_test.go
func TestSQLiteConnectionRepository(t *testing.T) {
    // 使用临时数据库
    db, err := sql.Open("sqlite", "file::memory:?mode=memory")
    if err != nil {
        t.Fatalf("open sqlite: %v", err)
    }
    defer db.Close()

    // 执行 schema
    schema := readFile("schema.sql")
    _, err = db.Exec(schema)
    if err != nil {
        t.Fatalf("init schema: %v", err)
    }

    repo := NewSQLiteConnectionRepository(db)

    // 测试 Save
    conn := &MySQLConnection{
        ID:   uuid.New().String(),
        Name: "test-conn",
        Host: "localhost",
        Port: 3306,
        // ...
    }

    err = repo.Save(context.Background(), conn)
    if err != nil {
        t.Errorf("Save() error = %v", err)
    }

    // 测试 FindByID
    found, err := repo.FindByID(context.Background(), conn.ID)
    if err != nil {
        t.Errorf("FindByID() error = %v", err)
    }
    if found.GetName() != conn.Name {
        t.Errorf("FindByID() name = %v, want %v", found.GetName(), conn.Name)
    }

    // 测试 FindAll
    all, err := repo.FindAll(context.Background())
    if err != nil {
        t.Errorf("FindAll() error = %v", err)
    }
    if len(all) != 1 {
        t.Errorf("FindAll() count = %v, want 1", len(all))
    }

    // 测试 Delete
    err = repo.Delete(context.Background(), conn.ID)
    if err != nil {
        t.Errorf("Delete() error = %v", err)
    }

    // 验证已删除
    _, err = repo.FindByID(context.Background(), conn.ID)
    if err == nil {
        t.Error("Delete() failed, record still exists")
    }
}
```

#### 8.3.3 E2E 测试（完整流程）

```go
// test/integration/benchmark_test.go
func TestE2E_SysbenchBenchmark(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping E2E test")
    }

    // 启动测试环境
    usecase := setupTestEnvironment(t)

    // 1. 创建连接
    conn := createTestConnection(t, usecase)

    // 2. 配置任务
    task := &BenchmarkTask{
        ID:           uuid.New().String(),
        Name:         "e2e-test-task",
        ConnectionID: conn.ID,
        TemplateID:   "sysbench-oltp-read-write",
        Parameters: map[string]any{
            "threads": 4,
            "time":    60,
        },
    }

    // 3. 启动任务
    run, err := usecase.StartTask(context.Background(), task)
    if err != nil {
        t.Fatalf("StartTask() error = %v", err)
    }

    // 4. 等待完成
    waitForCompletion(t, usecase, run.ID, 5*time.Minute)

    // 5. 验证结果
    result, err := usecase.GetRunResult(context.Background(), run.ID)
    if err != nil {
        t.Fatalf("GetRunResult() error = %v", err)
    }

    if result.TPSCalculated <= 0 {
        t.Errorf("TPS = %v, want > 0", result.TPSCalculated)
    }

    // 6. 导出报告
    reportPath, err := usecase.GenerateReport(context.Background(), GenerateReportRequest{
        RunID:  run.ID,
        Format: "md",
    })
    if err != nil {
        t.Fatalf("GenerateReport() error = %v", err)
    }

    // 验证报告文件存在
    if _, err := os.Stat(reportPath); os.IsNotExist(err) {
        t.Errorf("Report file not found: %s", reportPath)
    }
}
```

---

## 9. 质量门禁

### 9.1 代码质量标准

所有 PR 必须通过：

1. **格式检查**
   ```bash
   gofmt -l . | wc -l  # 必须为 0
   ```

2. **静态检查**
   ```bash
   go vet ./...
   golangci-lint run  # 零错误
   ```

3. **测试覆盖**
   ```bash
   go test -cover ./...
   # 覆盖率 > 80%
   ```

4. **竞态检测**
   ```bash
   go test -race ./...
   # 零竞态
   ```

5. **安全扫描**
   ```bash
   govulncheck ./...
   # 零已知漏洞
   ```

### 9.2 CI/CD 流水线

见后续文档...

---

## 10. 风险与缓解

| 风险 | 影响 | 概率 | 缓解措施 |
|------|------|------|---------|
| Fyne GUI 性能问题 | 高 | 中 | 异步加载、虚拟滚动 |
| 工具输出格式变化 | 高 | 中 | 灵活解析器、保留原始输出 |
| SQLite 并发限制 | 中 | 低 | WAL 模式、单连接 |
| PDF 转换失败 | 中 | 中 | 降级到 MD/HTML |
| Keyring 不可用 | 低 | 中 | 加密文件降级 |

---

**文档结束**
