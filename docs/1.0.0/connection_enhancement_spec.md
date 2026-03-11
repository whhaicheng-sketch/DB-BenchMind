# Connection Enhancement Specification (2026-03-10)

## 目标

完善 PostgreSQL、SQL Server、Oracle 的连接体验，使其与 MySQL 保持一致，同时优化 Connection 列表页 UI。

## 边界

### 包含
- PostgreSQL/SQL Server/Oracle 连接表单配置
- 连接测试功能完善
- 错误提示优化
- 列表页 SSH 标志位置调整
- PostgreSQL 异常标识排查与处理

### 不包含
- SSH Tunnel 实际连接功能（已实现）
- WinRM 实际连接功能（已实现）
- 新增数据库类型支持

---

## 功能要求

### REQ-CONN-001: SSH 标志位置调整

**描述**:
WHEN 用户查看连接列表时,
THE SYSTEM SHALL 将 SSH 标志显示在连接项左侧信息区域，而非操作按钮区域。

**当前状态**:
- SSH 标志在 `conn-tags` 区域，位于 name/host 之后
- 与 SSL、Database 标签混在一起

**目标状态**:
- SSH 标志移到 `conn-main` 区域，紧跟连接名称
- 作为连接属性标识，而非操作区标识

**验收标准**:
- SSH 标志显示在连接名称旁边
- SSH 标志不与 Test/Edit/Delete 按钮区域混在一起

---

### REQ-CONN-002: PostgreSQL 异常标识处理

**描述**:
WHEN 用户查看 PostgreSQL 连接列表项时,
THE SYSTEM SHALL 不显示无意义的数据库名称标签。

**当前状态**:
- PostgreSQL 列表项显示 "postgres" 小标签（来自 database 字段）
- 这个行为不清晰，且只有 PostgreSQL 有

**分析**:
- 来源：`conn.database` 字段值为"postgres"（PostgreSQL 默认数据库名）
- 当前代码：`<span v-if="conn.database" class="tag tag-db">{{ conn.database }}</span>`

**目标状态**:
- 移除数据库名称标签显示（或统一规则）
- 如果保留，所有数据库类型都应显示，且要有明确产品价值

**验收标准**:
- PostgreSQL 列表项不再显示异常的 "postgres" 标签
- 或者所有数据库类型统一显示数据库名称标签

---

### REQ-CONN-003: PostgreSQL 连接测试完善

**描述**:
WHEN 用户测试 PostgreSQL 连接时,
THE SYSTEM SHALL 提供与 MySQL 一致的测试体验和结果展示。

**当前状态**:
- 连接表单支持 PostgreSQL 配置
- TestConnectionDirect 支持 PostgreSQL
- SSL Mode 需要验证

**验收标准**:
- PostgreSQL 连接可正常配置
- 数据库连接测试可执行
- 结果展示清晰（成功/失败 + 延迟 + 版本）
- 错误提示友好

---

### REQ-CONN-004: SQL Server 连接测试完善

**描述**:
WHEN 用户测试 SQL Server 连接时,
THE SYSTEM SHALL 提供与 MySQL 一致的测试体验和清晰的错误提示。

**当前状态**:
- 连接表单支持 SQL Server 配置
- TestConnectionDirect 支持 SQL Server
- Trust Server Certificate 选项存在
- 可能存在连接策略问题

**验收标准**:
- SQL Server 连接可正常配置
- 数据库连接测试可执行
- 错误提示比当前更清晰、更合理
- 与 MySQL 测试体验一致

---

### REQ-CONN-005: Oracle 连接测试完善

**描述**:
WHEN 用户测试 Oracle 连接时,
THE SYSTEM SHALL 基于当前可用的 Oracle 连接事实实现测试功能。

**当前状态**:
- 命令行可成功连接：`sqlplus system/Qwer1234@192.168.134.129:1521/orcl`
- 连接表单支持 Oracle 配置
- TestConnectionDirect 支持 Oracle
- 需要验证当前实现为何未达预期

**验收标准**:
- Oracle 连接可正常配置
- 数据库连接测试可执行
- 结果展示清晰
- 支持 Service Name / SID 配置

---

## 非功能要求

### NFR-CONN-001: 用户体验一致性
- PostgreSQL/SQL Server/Oracle 的测试按钮、结果提示、保存前验证逻辑与 MySQL 保持一致

### NFR-CONN-002: 错误提示友好
- 失败提示要尽量清晰，不要只抛底层原始错误
- 提供用户可理解的错误原因和建议

### NFR-CONN-003: UI 清晰度
- 属性信息与操作按钮清晰分层
- 避免无意义的调试残留或硬编码标记

---

## 验收清单

### A. 文档
- [ ] 是否检查了现有 spec / plan / tasks
- [ ] 是否新增或更新了相关文档
- [ ] 文档是否与最终实现一致

### B. PostgreSQL
- [ ] 是否可以正常配置连接
- [ ] 是否可以执行数据库连接测试
- [ ] 是否结果展示清晰
- [ ] 是否移除了异常/无意义的 postgres 特殊标识

### C. SQL Server
- [ ] 是否可以正常配置连接
- [ ] 是否可以执行数据库连接测试
- [ ] 是否错误提示比当前更清晰、更合理
- [ ] 是否和 MySQL 的测试体验尽量一致

### D. Oracle
- [ ] 是否基于当前可用的 Oracle 连接事实补齐实现
- [ ] 是否可以正常配置连接
- [ ] 是否可以执行数据库连接测试
- [ ] 是否结果展示清晰

### E. 列表页 UI
- [ ] SSH 标志是否已移到左侧信息区域
- [ ] SSH 标志是否不再出现在 Edit / Test 按钮区域附近
- [ ] 属性信息与操作按钮是否已经清晰分层
