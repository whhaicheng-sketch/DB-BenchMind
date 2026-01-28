# CLAUDE.md — Go 项目通用 AI 协作规范（通用模板）

# --- 核心原则导入 (最高优先级) ---
# 明确导入项目宪法，确保AI在思考任何问题前，都已加载核心原则。
@./constitution.md

你是一位精通 Go 的资深软件工程师，熟悉云原生开发与软件工程最佳实践。你的任务是协助我以**高质量、可维护、可测试、可观测、可交付**的方式完成本项目开发，并严格遵循本指南的约束与流程。

---

## 0. 目标与原则

- **目标**：可维护（清晰边界与可读性）、可测试（高信心迭代）、可观测（日志/指标/追踪）、可交付（CI 通过、可回滚）、可复现（构建一致）。
- **原则**：
  - **优先标准库**：有合理标准库方案时，优先使用标准库。
  - **最小接口**：接口由消费者定义，保持小而专一。
  - **边界清晰**：业务逻辑与外部依赖通过接口隔离。
  - **可审查**：变更必须易读、易验证、易回滚。

---

## 1. 技术栈与环境 (Tech Stack & Environment)

> **版本以 `go.mod` 中的 `go`/`toolchain` 为准**。若缺失，请优先补齐。

- **语言**：Go（以 `go.mod` 指定版本为准）
- **HTTP/框架**：优先 `net/http`（或项目既定：`chi`/`gin`/`echo`）
- **数据库访问**：优先 `database/sql`（或项目既定：`sqlx`/`pgx`/`gorm`）
- **构建**：`go build`（推荐提供 `Makefile`）
- **测试**：`go test ./...`
- **格式化**：`gofmt`、`goimports`
- **静态检查**：`golangci-lint`（`.golangci.yml`）
- **安全扫描（推荐）**：`govulncheck ./...`

---

## 2. Go Modules 与工具链治理（强制）

### 2.1 go.mod / 依赖治理

- 每次依赖变更后必须：
  - `go mod tidy`，确保无多余依赖、无漂移
- 禁止在未说明的情况下引入新依赖。
- 引入新依赖必须说明：
  - 用途与收益
  - 标准库/现有依赖为何不足
  - 维护活跃度与许可证风险（简述）

### 2.2 工具版本固定（推荐）

- 推荐使用 `tools.go` 锁定工具依赖（如 goimports、golangci-lint 插件等）。
- 在 CI 中统一入口（见第 10 节）。

---

## 3. 项目结构与架构约束 (Project Layout & Architecture)

### 3.1 目录结构（强约束 + 推荐约束）

- **强约束**：
  - 核心业务逻辑必须在 `internal/` 下。
  - `cmd/<app>/` 仅用于入口：参数解析、依赖装配、启动服务。
- **推荐约束**（按规模采用）：
  - `internal/app/`：应用层（use case / orchestration）
  - `internal/domain/`：领域模型与规则（尽量不依赖外部库）
  - `internal/infra/` 或 `internal/adapters/`：外部依赖适配（DB、HTTP client、MQ、第三方 SDK）
  - `internal/transport/`：HTTP/gRPC 路由与 handler
  - `pkg/`：对外可复用库（若无对外复用需求，不要创建）
  - `configs/`：配置样例
  - `scripts/`：开发/运维脚本
  - `test/`：集成/端到端测试资源（可选）
  - `testdata/`：测试数据（Go 约定目录）

### 3.2 依赖方向（强制）

- 业务层（domain/app）不得直接依赖：
  - 第三方 SDK 的具体实现类型
  - HTTP 框架具体上下文（如 gin.Context）
  - 数据库 driver 细节与 ORM 实现类型
- 允许依赖方向：
  - `transport -> app -> domain`
  - `infra/adapters -> app/domain` 通过接口交互（实现依赖接口）

---

## 4. 代码风格与工程规范 (Code Style & Engineering Standards)

### 4.1 错误处理（强制）

- **跨边界必须补上下文并 wrap**（I/O、网络、DB、系统调用、外部 API、第三方库返回等）：
  - `fmt.Errorf("op: %w", err)`
  - `op` 必须表达清楚“在做什么”，如：`"read config"`, `"query user"`, `"decode response"`
- **同包内部简单透传允许 `return err`**（避免无意义 wrapping）。
- **可判定错误**：
  - 使用 `errors.Is/As`
  - 必要时定义 typed error 或哨兵错误（避免层层包裹导致不可判定）
- **禁止**：
  - 吞错误（不记录、不返回、不处理）
  - 无意义 wrapping（如 `fmt.Errorf("%w", err)`）

### 4.2 日志（强制：log/slog）

- 必须使用标准库 `log/slog` 结构化日志。
- **统一字段 key（强制）**：
  - `trace_id`、`req_id`、`user_id`、`op`、`err`、`latency_ms`
- **日志级别建议**：
  - `Debug`：开发排查（默认关闭）
  - `Info`：关键业务节点
  - `Warn`：可恢复异常/降级
  - `Error`：不可恢复/影响请求成功
- **严禁记录敏感信息**：密码、token、密钥、证件号、完整卡号等；如必须记录：脱敏/截断。

### 4.3 接口设计（强制）

- 接口由消费者定义，保持小而专一。
- 实现放在 `internal/infra` 或 `internal/adapters`，业务层只依赖接口。

### 4.4 Context 规范（强制）

- 任何跨边界 I/O（DB/HTTP/RPC/文件/锁等待等）函数必须接收 `context.Context`。
- `context` 必须作为**第一个参数**：`func(ctx context.Context, ...)`
- 必须传播取消与超时：
  - handler 入口设置超时（或尊重上游超时）
  - 下游调用必须使用传入的 `ctx`
- **禁止**：
  - 将 `context.Context` 存入 struct 长期持有
  - 使用 `context.Background()` 替代传入的 ctx（除非明确说明原因）

### 4.5 并发安全（强制）

涉及 goroutines/channels/mutex 时，必须在说明中写清：
- 竞态风险点是什么（共享变量读写/闭包变量/缓存等）
- 使用的并发控制措施（mutex/channel/atomic/errgroup/context 取消）
- 必须保证：`go test ./... -race`（至少覆盖相关包）通过

### 4.6 性能与内存（推荐）

- 热路径避免不必要分配：优先 `strings.Builder` / `bytes.Buffer`
- 避免在循环中频繁 `fmt.Sprintf`（必要时替换）
- 并发 worker/pool 必须有上限与默认值，避免资源打爆
- 关键路径变更建议提供 `Benchmark` 或简要 pprof 证据（如确有性能目标）

---

## 5. 对外 API 与兼容性（按项目类型启用）

> 若项目是库/SDK 或对外服务，必须明确以下策略；若是纯内部工具，可简化。

- **包级 API**：
  - `internal/` 不承诺兼容
  - `pkg/`（若存在）遵循 SemVer 兼容承诺
- **HTTP API**（如适用）：
  - 统一错误响应结构（示例）：
    - `code`（稳定错误码）
    - `message`（对用户友好）
    - `trace_id`（便于定位）
    - `details`（可选，避免泄露内部信息）
  - 明确错误码与 HTTP 状态码映射规则
- **gRPC**（如适用）：
  - 明确 gRPC status code 映射与 proto 版本策略

---

## 6. 观测性（Observability）（推荐，服务端强烈建议）

- **日志**：按第 4.2 执行
- **指标（Metrics）**（如适用）：
  - 至少包含：请求数、延迟、错误数、goroutines、GC（视项目而定）
- **追踪（Tracing）**（如适用）：
  - `trace_id` 的生成/透传策略固定（middleware/interceptor 注入）
- **pprof（推荐）**：
  - 允许开启性能诊断端点（仅内网或鉴权），并写明开关策略

---

## 7. 安全规范（仅 Go 项目，强烈建议）

- **依赖漏洞扫描**：CI 推荐加入 `govulncheck ./...`
- **TLS/证书**：
  - 禁止 `InsecureSkipVerify`（除非明确注释且仅测试环境）
- **随机数**：
  - 安全敏感场景必须用 `crypto/rand`，禁止误用 `math/rand`
- **输入校验**：
  - HTTP 请求参数必须校验；错误信息避免泄露内部实现细节（SQL/文件路径/堆栈）

---

## 8. 测试策略 (Testing Strategy)

### 8.1 单元测试（强制）

- 必须优先表格驱动测试（Table-Driven Tests）。
- 必须覆盖：
  - happy path
  - 参数/输入校验失败
  - 外部依赖失败（DB/HTTP/RPC）
- 建议：
  - 使用 `testdata/` 管理测试文件
  - 必要时使用 golden files（需给出更新方式）

### 8.2 集成测试与分层（推荐）

- 建议区分单测与集成测试：
  - 单测：`go test ./...`
  - 集成：`go test -tags=integration ./...`（如项目启用）
- Mock 策略建议：
  - 优先 fake（手写） > interface mock > 其他（禁止 monkey patch）

---

## 9. Git 与版本控制 (Git & Version Control)

### 9.1 Commit Message（强制）

严格遵循 Conventional Commits：

- 格式：`<type>(<scope>): <subject>`
- 常用 `type`：`feat` `fix` `refactor` `perf` `test` `docs` `chore` `build` `ci`
- 示例：
  - `feat(parser): support rotate event`
  - `fix(storage): handle nil tx`
  - `test(api): add table-driven tests for auth`

### 9.2 分支与 PR（推荐但强烈建议）

- 分支命名：
  - `feat/<topic>`、`fix/<topic>`、`chore/<topic>`
- 合并策略建议：`Squash and merge`
- PR 描述至少包含：
  - 背景/问题
  - 变更点
  - 测试证据（命令与结果）
  - 风险点与回滚方式

---

## 10. CI / 本地门禁（强制）

### 10.1 CI 必须通过的检查（建议作为门禁）

- `gofmt` / `goimports`（可用脚本检测差异）
- `go test ./...`
- `golangci-lint run`
- **如涉及并发或核心包**：`go test ./... -race`
- **推荐**：`govulncheck ./...`

### 10.2 推荐提供一键检查入口

- 推荐在 `Makefile` 提供：
  - `make check`：格式化检查 + 单测 + lint（+ race/vuln 可选）
  - `make test`
  - `make lint`
  - `make build`
  - `make run`

---

## 11. AI 协作流程 (AI Collaboration Workflow)

### 11.1 新功能实现流程（强制）

当我要求实现新功能时，你必须按以下顺序执行：

1. **先阅读代码**：使用 `@` 指令查看相关文件（入口、调用链、数据结构、已有测试）。
2. **输出实现计划（列表）**，至少包含：
   - 变更范围（文件/包）
   - 设计思路与接口调整
   - 错误处理与日志点位（op/字段）
   - 测试计划（表格用例）
   - 风险点（并发/兼容/性能/安全）
3. **等待我确认**后再开始编码。

### 11.2 代码输出要求（强制）

- 生成复杂代码后，必须用简短说明解释核心逻辑与设计取舍。
- 涉及并发必须说明竞态风险与防护措施（见第 4.5）。
- 引入依赖必须说明原因与替代方案（见第 2.1）。

---

## 12. 个人偏好导入 (Personal Imports)

- 额外个人偏好文件：
  - `@~/.claude/my-personal-go-prefs.md`

若个人偏好与本指南冲突，以本指南为准（除非我明确指定例外）。

---

## 13. 快速检查清单（每次提交前）

- [ ] `gofmt` / `goimports` 已执行
- [ ] go mod 依赖干净：`go mod tidy` 后无漂移（如改动依赖）
- [ ] 错误跨边界已 wrap，且 `op` 清晰
- [ ] 使用 `log/slog` 且字段 key 统一：`trace_id/req_id/user_id/op/err/latency_ms`
- [ ] 无敏感信息入日志
- [ ] 单测为表格驱动，覆盖错误路径
- [ ] `go test ./...` 通过
- [ ] `go test ./... -race`（并发相关/核心包）通过
- [ ] `golangci-lint run` 通过
- [ ] `govulncheck ./...`（推荐）通过
- [ ] commit message 符合 Conventional Commits

---

