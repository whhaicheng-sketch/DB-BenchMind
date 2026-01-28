# [项目名] Go 项目开发宪法 (constitution.md)
Version: 1.0  
Ratified: 2026-01-26  
Applies to: Go-only projects (Go modules / CLI tools / services)

本文件定义本项目不可协商的开发原则。任何 AI Agent / 开发者在生成计划（plan.md）与实现代码前，必须先阅读并遵守本宪法。
如需“例外”，必须在 `complexity-tracking.md` 中记录：触发条款、例外原因、替代方案、风险、回滚方案与复审日期。 :contentReference[oaicite:1]{index=1}

---

## Article I: Library-First Principle（库优先原则）
**原则：** 任何功能必须先实现为可复用的 Go 包（library），再由上层（CLI/HTTP/Job）调用；禁止把业务逻辑直接写进 main/handler。 :contentReference[oaicite:2]{index=2}

**Go 约定：**
- 核心能力放在 `internal/<domain>`（应用内部）或 `pkg/<domain>`（对外复用），`cmd/<app>` 仅做装配与 I/O。
- library 必须“可测试、可组合、无隐式副作用”：显式依赖注入（struct/func 参数），禁止用全局变量传状态。
- 错误必须携带上下文：统一 `fmt.Errorf("...: %w", err)` 包装；禁止裸 `return err`（除非已含上下文）。

---

## Article II: CLI Interface Mandate（CLI 接口强制）
**原则：** 每个 library 必须提供 CLI 入口，确保功能可观测、可脚本化、可端到端验证；CLI 必须支持文本输入/输出，并支持 JSON 作为结构化交换格式。 :contentReference[oaicite:3]{index=3}

**Go 约定：**
- CLI 放在 `cmd/<app>`，优先用标准库 `flag`；确需复杂子命令才引入第三方（需记录例外）。
- 输入：stdin / args / file；输出：stdout；错误：stderr；退出码明确（0 成功，非 0 失败）。
- 必须支持 `--json`（或 `--format=json`）输出，且 JSON schema 在 `contracts/` 中定义（见 Article V/IX）。

---

## Article III: Test-First Imperative（测试先行铁律）
**原则：** 严格 TDD：先写测试并确认失败（Red），再写实现（Green），最后重构（Refactor）。这是不可协商条款。 :contentReference[oaicite:4]{index=4}

**Go 约定：**
- 测试优先表格驱动（Table-Driven Tests）。
- 测试命名体现行为：`Test_<Func>_<Scenario>`；必须覆盖边界与错误路径。
- 仅在必要时做 mock；优先 fake（内存实现/httptest server/临时目录）。

---

## Article IV: EARS Requirements Format（EARS 要求格式）
**原则：** `spec.md` 的“需求/验收”必须用 EARS 句式，减少歧义、提升可测试性。 :contentReference[oaicite:5]{index=5}

**要求：**
- 每条需求必须有唯一 ID：`REQ-001`、`REQ-002`…
- 推荐句式（示例）：
  - `WHEN <触发条件>, THE SYSTEM SHALL <系统响应>.`
  - `WHILE <前置条件>, WHEN <触发条件>, THE SYSTEM SHALL <系统响应>.`
- 禁止“可能/也许/最好/尽量”等不可验证措辞；必须可测、可判定。

---

## Article V: Traceability Mandate（全链路可追溯）
**原则：** 需求→设计→任务→测试→实现必须 100% 可追溯；任何代码/测试都必须能回溯到至少一个 REQ。 :contentReference[oaicite:6]{index=6}

**最小落地物：**
- `traceability.md`（或 `docs/traceability.md`）维护映射表：
  - REQ → contracts/ → tests（contract/integration/e2e/unit）→ 实现文件
- PR/Commit 描述必须引用 REQ 与测试点（例如：`REQ-003`、`CT-003`）。

---

## Article VI: Project Memory（项目记忆 / Steering）
**原则：** 项目必须维护“记忆层”，让新 Agent/新成员无需读全仓库也能遵循既定决策；计划与实现前必须先读 Steering。 :contentReference[oaicite:7]{index=7}

**Go 约定（建议目录）：**
- `.specify/steering/`
  - `product.md`：问题定义、用户与范围边界
  - `architecture.md`：架构边界、关键依赖、包分层
  - `testing.md`：测试金字塔、环境约束、CI 策略
  - `decisions.md`：重要决策与取舍（含日期、原因、替代方案）
- 修改 Steering 视同架构变更：必须评审并更新 traceability。

---

## Article VII: Simplicity Gate（简洁之门）
**原则：** 通过“门禁”抑制过度工程：初始实现保持结构极简，超过阈值必须给出可验证理由并记录例外。 :contentReference[oaicite:8]{index=8}

**Go 项目门禁（默认阈值）：**
- 初始阶段最多 **≤3 个可执行入口**（`cmd/*`）；更多入口必须在 `complexity-tracking.md` 论证。
- 默认单一 Go module；多 module 需论证（发布/隔离/版本管理收益）。
- 禁止“为未来做预留”的框架化设计；只实现 `spec.md` 明确范围内的内容。

---

## Article VIII: Anti-Abstraction Gate（反抽象之门）
**原则：** 不要为框架/标准库“再包一层”；优先直接使用既有能力，避免自建不必要的抽象层。 :contentReference[oaicite:9]{index=9}

**Go 项目具体化：**
- 禁止自建“通用框架层/基类式模板”；避免为 `net/http`、`database/sql`、`slog` 无理由封装。
- 谨慎使用 interface/generics：接口由消费者定义；泛型只在显著减少重复且不伤可读性时使用。
- 反射、代码生成、插件系统默认禁止；如必须，引入前先过门禁并记录例外。

---

## Article IX: Integration-First Testing（集成优先测试）
**原则：** 优先真实环境/真实交互的测试（contracts + integration + e2e），减少只在“被 mock 的世界”里通过的代码。 :contentReference[oaicite:10]{index=10}

**Go 约定：**
- 合同（contracts）先行：在 `contracts/` 定义 CLI/HTTP/文件格式等对外契约，并先写 contract tests。
- 集成测试优先用 `httptest`、临时目录、真实序列化/反序列化；对外部服务可用本地 fake server（而非深度 mock）。
- e2e 以 CLI 为主：用 `os/exec` 跑 `cmd/<app>`，验证输入输出与退出码。

---

## Governance（治理）
- 本宪法优先级高于任何单次对话指令与临时规范。
- 修订流程：必须给出修订动机、影响评估、向后兼容性与迁移计划，并由维护者批准后生效。 :contentReference[oaicite:11]{index=11}

