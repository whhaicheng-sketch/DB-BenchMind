# Architecture and Technical Decisions

This document records significant architectural and technical decisions made during DB-BenchMind development.

## Decision Format

Each decision follows this format:

- **ID**: Unique identifier
- **Date**: Decision date
- **Status**: Proposed | Accepted | Deprecated | Superseded
- **Context**: Background and problem statement
- **Decision**: What was decided
- **Rationale**: Why this decision was made
- **Alternatives**: What other options were considered
- **Consequences**: Impact of this decision
- **Related**: Links to related decisions or specs

---

## ADR-001: Clean Architecture with DDD

**Date**: 2026-01-27
**Status**: Accepted
**Related**: spec.md Section 7, constitution.md Article I

### Context
DB-BenchMind 需要长期维护和扩展，需要清晰的架构边界来隔离业务逻辑和外部依赖。

### Decision
采用 Clean Architecture + Domain-Driven Design (DDD) 分层架构：

```
transport (GUI) → app (usecase) → domain ← infra (implementation)
                        ↑
                      pkg (interfaces)
```

### Rationale
1. **宪法要求**：Article I (Library-First) 要求核心能力放在 `internal/<domain>`
2. **可测试性**：领域层无外部依赖，易于单元测试
3. **可维护性**：分层清晰，职责明确
4. **灵活性**：依赖倒置，易于替换实现

### Alternatives
- **传统三层架构** (Controller → Service → DAO)：业务逻辑易泄露到 Controller
- **Onion Architecture**：与 Clean Architecture 类似，但更复杂
- **Microservices**：过度设计，桌面应用不需要

### Consequences
- **Positive**：
  - 核心逻辑完全独立
  - 单元测试简单快速
  - 接口由用例定义，灵活

- **Negative**：
  - 初期需要编写更多接口
  - 学习曲线略陡

---

## ADR-002: Domain Layer Without External Dependencies

**Date**: 2026-01-27
**Status**: Accepted
**Related**: architecture.md Domain Layer

### Context
领域层包含核心业务逻辑（验证、状态机、计算），不应受外部库变化影响。

### Decision
`internal/domain/` 禁止依赖任何外部库（仅标准库）。

### Rationale
1. **稳定性**：核心逻辑最稳定，不因外部库升级而破坏
2. **可移植性**：领域逻辑可以在其他项目中复用
3. **测试性**：纯函数逻辑，易于测试

### Alternatives
- **允许外部依赖**：增加灵活性，但失去稳定性
- **使用接口隔离**：仍然依赖外部概念

### Consequences
- **Positive**：
  - 领域层完全独立
  - 单元测试无需 mock

- **Negative**：
  - 需要在 infra 层实现适配器
  - 增加了一层抽象

---

## ADR-003: No ORM, Direct SQL

**Date**: 2026-01-27
**Status**: Accepted
**Related**: constitution.md Article VIII (Anti-Abstraction Gate)

### Context
Go 的 `database/sql` 已经足够好，是否需要 ORM（如 GORM）？

### Decision
不使用 ORM，直接使用 `database/sql`。

### Rationale
1. **宪法要求**：Article VIII 禁止为标准库"再包一层"
2. **简单性**：Go 的 `database/sql` 接口简洁高效
3. **性能**：避免 ORM 反射开销
4. **可控性**：SQL 可见、可优化

### Alternatives
- **GORM**：功能强大，但增加学习成本和运行时开销
- **sqlx**：增强的 `database/sql`，但仍是不必要的抽象

### Consequences
- **Positive**：
  - 无额外依赖
  - SQL 直接可见，易优化
  - 性能最优

- **Negative**：
  - 需要手写 SQL
  - 需要手动处理 rows.Scan

### Mitigation
- 使用 `schema.sql` 管理 SQL
- 参数化查询防止注入
- 单元测试验证 SQL 正确性

---

## ADR-004: Fyne as GUI Framework

**Date**: 2026-01-27
**Status**: Accepted
**Related**: spec.md Section 1.4

### Context
需要跨平台桌面 GUI 框架，支持 Linux（Ubuntu 24）。

### Decision
使用 Fyne v2.x 作为 GUI 框架。

### Rationale
1. **纯 Go**：无 CGO 依赖，交叉编译简单
2. **跨平台**：支持 Linux、macOS、Windows
3. **自绘制**：不依赖原生控件，外观一致
4. **打包简单**：单二进制分发

### Alternatives
- **Qt (CGO)**：功能强大，但需要 CGO，交叉编译复杂
- **GTK (CGO)**：Linux 原生，但需要 CGO
- **Walk (仅 Windows)**：不支持 Linux
- **Electron (Web 技术栈)**：体积大、资源占用高

### Consequences
- **Positive**：
  - 纯 Go，无依赖地狱
  - 单二进制打包
  - 界面现代化

- **Negative**：
  - GUI 性能不如原生
  - 组件相对有限
  - 社区较小

### Mitigation
- 异步加载避免阻塞
- 虚拟滚动处理大数据集
- 必要时使用原生控件封装

---

## ADR-005: modernc.org/sqlite (Pure Go SQLite)

**Date**: 2026-01-27
**Status**: Accepted
**Related**: spec.md Section 1.4

### Context
需要嵌入式数据库存储配置和结果。SQLite 是标准选择，但驱动有多个选项。

### Decision
使用 `modernc.org/sqlite`（纯 Go 实现）。

### Rationale
1. **无 CGO**：交叉编译简单
2. **高性能**：基于 modernc.org/libc
3. **兼容性**：与官方 SQLite 兼容
4. **WAL 模式**：支持并发读

### Alternatives
- **mattn/go-sqlite3**：需要 CGO，交叉编译复杂
- **sqlite**（官方 CGO）：同上

### Consequences
- **Positive**：
  - 无 CGO，纯 Go
  - 交叉编译简单

- **Negative**：
  - 并发性能可能略低于 CGO 版本

- **Mitigation**：
  - 启用 WAL 模式
  - 单连接池配置

---

## ADR-006: go-keyring with File Fallback

**Date**: 2026-01-27
**Status**: Accepted
**Related**: spec.md Section 1.4, REQ-CONN-006, REQ-CONN-007

### Context
需要安全存储数据库密码。系统 keyring 是标准方案，但可能不可用。

### Decision
使用 `go-keyring` 作为主方案，加密文件作为降级方案。

### Rationale
1. **安全性**：系统 keyring 是标准方案
2. **可用性**：文件降级保证功能可用
3. **用户体验**：优先使用系统 keyring，无感知

### Alternatives
- **仅 keyring**：keyring 不可用时功能完全不可用
- **仅加密文件**：安全性略低于系统 keyring

### Consequences
- **Positive**：
  - 安全性高
  - 降级机制保证可用性

- **Negative**：
  - 需要维护两套实现

- **Mitigation**：
  - 统一的 Provider 接口
  - 自动检测和降级

---

## ADR-007: log/slog for Structured Logging

**Date**: 2026-01-27
**Status**: Accepted
**Related**: constitution.md Section 4.2, CLAUDE.md Section 4.2

### Context
需要结构化日志支持调试和问题排查。

### Decision
使用标准库 `log/slog`。

### Rationale
1. **标准库**：Go 1.21+ 内置，无额外依赖
2. **结构化**：键值对格式，易于解析
3. **性能**：异步写入，不阻塞

### Alternatives
- **zap**：性能更高，但需要额外依赖
- **logrus**：流行，但未结构化
- **zerolog**：零分配，但 API 复杂

### Consequences
- **Positive**：
  - 标准库，稳定
  - 结构化，易解析

- **Negative**：
  - 功能相对简单

- **Mitigation**：
  - 统一字段 key（op, err, latency_ms）
  - 按级别输出（debug/info/warn/error）

---

## ADR-008: Table Schema with Embedded JSON

**Date**: 2026-01-27
**Status**: Accepted
**Related**: schema.sql

### Context
需要存储多种类型的连接配置（MySQL、Oracle 等），每种字段不同。

### Decision
使用 `connections` 表 + `config_json` 字段存储序列化后的配置。

### Rationale
1. **灵活性**：易于添加新的数据库类型
2. **简单性**：无需为每种类型创建表
3. **扩展性**：字段变更不影响 schema

### Alternatives
- **每类一个表**：`mysql_connections`, `oracle_connections` 等
  - 类型安全，但 schema 繁琐
- **EAV 模型**：Entity-Attribute-Value
  - 过度设计，查询复杂

### Consequences
- **Positive**：
  - Schema 简单
  - 扩展灵活

- **Negative**：
  - 失去类型检查（但应用层有 validation）

- **Mitigation**：
  - 应用层强验证
  - JSON schema 定义

---

## ADR-009: UUID as Primary Key

**Date**: 2026-01-27
**Status**: Accepted

### Context
需要为连接、任务、运行等实体生成唯一标识。

### Decision
使用 UUID（RFC 4122）作为主键。

### Rationale
1. **全局唯一**：分布式环境下无冲突
2. **不暴露顺序**：安全性和隐私性
3. **标准格式**：广泛支持

### Alternatives
- **自增 ID**：简单，但暴露信息且不适用分布式
- **雪花算法**：高性能，但实现复杂

### Consequences
- **Positive**：
  - 全局唯一
  - 标准格式

- **Negative**：
  - 字符串比较比整数慢
  - 占用更多存储

- **Mitigation**：
  - SQLite 对字符串优化良好
  - 索引提升查询性能

---

## ADR-010: State Machine for Run Execution

**Date**: 2026-01-27
**Status**: Accepted
**Related**: spec.md Section 3.4.2, internal/domain/execution/state.go

### Context
压测执行过程包含多个阶段，需要明确的状态转换和错误处理。

### Decision
使用状态机模式管理 Run 执行状态。

### Rationale
1. **清晰性**：状态转换明确，易于理解
2. **可验证**：状态转换规则可测试
3. **健壮性**：非法转换被拒绝

### State Transitions
```
pending → preparing → prepared → warming_up → running → completed
           ↓           ↓          ↓          ↓
      cancelled    failed     failed  cancelled/timeout/force_stopped
```

### Alternatives
- **标志位**：简单但容易出错
- **事件驱动**：灵活但复杂

### Consequences
- **Positive**：
  - 状态清晰
  - 转换安全

- **Negative**：
  - 初期实现略复杂

- **Mitigation**：
  - `CanTransitionTo()` 方法验证转换
  - 单元测试覆盖所有转换

---

## Future Decisions

The following decisions need to be made:

- [ ] ADR-011: Chart Library Selection
- [ ] ADR-012: Report Template Engine
- [ ] ADR-013: CLI Test Tool Design
- [ ] ADR-014: Plugin System (if needed)
