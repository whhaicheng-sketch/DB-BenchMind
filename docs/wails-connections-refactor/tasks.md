# Connections Page Refactor Tasks

## 任务清单

使用方法：
- [ ] 未开始
- [~] 进行中
- [x] 已完成
- [!] 阻塞

---

## Phase 1: 核心功能 (P0) ✅

### TASK-001: 扩展后端 Connection DTO [DONE]
- [x] 在 `internal/transportwails/bindings/connection.go` 添加 SSH 字段
  - `SSHEnabled bool`
  - `SSHPort int`
  - `SSHUsername string`
  - `SSHPassword string`
- [x] 添加 WinRM 字段
  - `WinRMEnabled bool`
  - `WinRMPort int`
  - `WinRMUseHTTPS bool`
  - `WinRMUsername string`
  - `WinRMPassword string`
- [x] 添加 SQL Server 字段
  - `TrustServerCertificate bool`

**关联需求**: REQ-001, REQ-002, REQ-003  
**预估工时**: 1h

---

### TASK-002: 实现 SSH 配置保存逻辑 [DONE]
- [x] 修改 `CreateConnection` 方法处理 SSH 字段
- [x] 修改 `UpdateConnection` 方法处理 SSH 字段
- [x] SSH 密码存储到 keyring (`<conn_id>:ssh`)
- [x] 编辑时从 keyring 加载 SSH 密码

**关联需求**: REQ-001, REQ-012  
**预估工时**: 1.5h

---

### TASK-003: 实现 WinRM 配置保存逻辑 [DONE]
- [x] 修改 `CreateConnection` 方法处理 WinRM 字段
- [x] 修改 `UpdateConnection` 方法处理 WinRM 字段
- [x] WinRM 密码存储到 keyring (`<conn_id>:winrm`)
- [x] 编辑时从 keyring 加载 WinRM 密码

**关联需求**: REQ-002, REQ-012  
**预估工时**: 1.5h

---

### TASK-004: 前端 ConnectionForm SSH 配置 UI [DONE]
- [x] 在 ConnectionForm.vue 添加 "Enable SSH Tunnel" 复选框
- [x] 添加 SSH 配置表单区域（Port, Username, Password）
- [x] 实现显示/隐藏逻辑
- [x] SQL Server 类型时隐藏 SSH 选项
- [x] 添加帮助文本 "💡 SSH Host uses Database Host"

**关联需求**: REQ-001  
**预估工时**: 1.5h

---

### TASK-005: 前端 ConnectionForm WinRM 配置 UI [DONE]
- [x] 在 ConnectionForm.vue 添加 "Enable WinRM" 复选框
- [x] 添加 WinRM 配置表单区域
- [x] 添加 "Use HTTPS" 复选框（自动切换端口 5985/5986）
- [x] 仅 SQL Server 类型时显示 WinRM 选项
- [x] 添加 "Trust Server Certificate" 选项

**关联需求**: REQ-002, REQ-003  
**预估工时**: 1.5h

---

### TASK-006: 前端 Store 更新 [DONE]
- [x] 更新 `connection.js` store 的 createConnection action
- [x] 更新 updateConnection action
- [x] 添加 SSH/WinRM 字段到请求

**关联需求**: REQ-001, REQ-002  
**预估工时**: 0.5h

---

## Phase 2: 测试功能 (P0/P1)

### TASK-007: 后端 SSH 测试方法 [DONE]
- [x] 在 ConnectionBinding 添加 `TestSSHConnection` 方法
- [x] 实现 SSH 连接测试逻辑
- [x] 返回测试结果（成功/失败、延迟、错误信息）

**关联需求**: REQ-009
**预估工时**: 1h

---

### TASK-008: 后端 WinRM 测试方法 [DONE]
- [x] 在 ConnectionBinding 添加 `TestWinRMConnection` 方法
- [x] 实现 WinRM 连接测试逻辑
- [x] 返回测试结果

**关联需求**: REQ-010
**预估工时**: 1h

---

### TASK-009: 前端 SSH/WinRM 测试按钮 [DONE]
- [x] 在 ConnectionForm 添加 "Test SSH" 按钮
- [x] 在 ConnectionForm 添加 "Test WinRM" 按钮
- [x] 根据配置状态动态显示/隐藏测试按钮
- [x] 显示测试结果

**关联需求**: REQ-009, REQ-010
**预估工时**: 1h

---

### TASK-010: 智能连接测试 [DONE]
- [x] 增强后端 TestConnection 方法
- [x] 实现智能测试流程（SSH → 数据库 → 直连 fallback）
- [x] 返回详细测试结果（SSH 状态 + 数据库 状态）
- [x] 前端显示详细测试结果

**关联需求**: REQ-011
**预估工时**: 1.5h

---

## Phase 3: UI 优化 (P1/P2)

### TASK-011: 连接列表分组显示 [DONE]
- [x] 修改 ConnectionsPage.vue 实现分组布局
- [x] 按数据库类型分组（MySQL, PostgreSQL, Oracle, SQL Server）
- [x] 每组显示连接数量
- [x] 实现分组折叠/展开功能

**关联需求**: REQ-004
**预估工时**: 1.5h

---

### TASK-012: 数据库图标 [DONE]
- [x] 为 MySQL 添加 🐬 图标
- [x] 为 PostgreSQL 添加 🐘 图标
- [x] 为 Oracle 添加 🔴 图标
- [x] 为 SQL Server 添加 🔷 图标

**关联需求**: REQ-005
**预估工时**: 0.5h

---

### TASK-013: SSH/WinRM 状态指示器 [DONE]
- [x] 连接启用 SSH 时显示 🔒 SSH 指示器
- [x] 连接启用 WinRM 时显示 🖥️ WinRM 指示器
- [x] 在连接列表和详情中显示

**关联需求**: REQ-006
**预估工时**: 0.5h

---

### TASK-014: 动态标签 Database/SID [DONE]
- [x] 监听数据库类型变化
- [x] MySQL/PostgreSQL/SQL Server 显示 "Database" 标签
- [x] Oracle 显示 "SID" 标签

**关联需求**: REQ-007
**预估工时**: 0.5h

---

### TASK-015: 数据库特定默认值 [DONE]
- [x] PostgreSQL 新建时 Database 默认 "postgres"
- [x] Oracle 新建时 SID 默认 "orcl"
- [x] SSH Username 默认 "root"

**关联需求**: REQ-008
**预估工时**: 0.5h

---

## Phase 4: 完善和测试

### TASK-016: 密码持久化验证
- [ ] 验证数据库密码正确存储和加载
- [ ] 验证 SSH 密码正确存储和加载
- [ ] 验证 WinRM 密码正确存储和加载
- [ ] 验证编辑连接时密码正确回填

**关联需求**: REQ-012  
**预估工时**: 1h

---

### TASK-016: 密码持久化验证 [IN PROGRESS]
- [~] 验证数据库密码正确存储和加载
- [~] 验证 SSH 密码正确存储和加载
- [~] 验证 WinRM 密码正确存储和加载
- [~] 验证编辑连接时密码正确回填

**预估工时**: 1.5h

---

### TASK-018: 前端组件测试
- [ ] ConnectionForm 组件测试（SSH 配置）
- [ ] ConnectionForm 组件测试（WinRM 配置）
- [ ] ConnectionForm 组件测试（动态标签）
- [ ] ConnectionsPage 组件测试（分组显示）

**预估工时**: 1h

---

### TASK-019: 集成测试
- [ ] E2E: 创建 MySQL 连接 + SSH
- [ ] E2E: 测试 SSH 连接
- [ ] E2E: 测试数据库连接
- [ ] E2E: 编辑连接修改 SSH
- [ ] E2E: 创建 SQL Server 连接 + WinRM

**预估工时**: 1h

---

## 任务依赖关系

```
TASK-001 ─┬─► TASK-002 ─┬─► TASK-007 ─► TASK-009
          │             │
          │             └─► TASK-010
          │
          ├─► TASK-003 ───► TASK-008 ─► TASK-009
          │
          ├─► TASK-004 ───► TASK-006
          │
          └─► TASK-005 ───► TASK-006

TASK-011 ─► TASK-012 ─► TASK-013

TASK-014 ─► TASK-015

TASK-016 ─► TASK-017 ─► TASK-018 ─► TASK-019
```

---

## 进度追踪

| 阶段 | 总任务 | 已完成 | 进行中 | 未开始 | 完成率 |
|-----|-------|-------|-------|-------|-------|
| Phase 1 | 6 | 6 | 0 | 0 | 100% |
| Phase 2 | 4 | 4 | 0 | 0 | 100% |
| Phase 3 | 5 | 5 | 0 | 0 | 100% |
| Phase 4 | 4 | 1 | 1 | 2 | 25% |
| **总计** | **19** | **16** | **1** | **2** | **84%** |


---

## 开始日期: 2026-03-07
## 目标完成日期: TBD

---

## 变更记录

> **重要**: 实现过程中如有功能变动、需求变更或设计调整，必须在此记录并同步更新 spec.md。

| 日期 | 任务 | 变更内容 | 更新文件 |
|-----|------|---------|---------|
| 2026-03-07 | 初始化 | 创建任务清单 | tasks.md, spec.md, plan.md |
| 2026-03-07 | TASK-001 | 完成 DTO 扩展：SSH/WinRM/TrustServerCertificate 字段 | connection.go |
| 2026-03-07 | TASK-002 | SSH 配置保存逻辑 (usecase 层已实现) | connection_usecase.go, connection.go |
| 2026-03-07 | TASK-003 | WinRM 配置保存逻辑 (usecase 层已实现) | connection_usecase.go, connection.go |
| 2026-03-07 | TASK-004 | ConnectionForm SSH 配置 UI 完成 | ConnectionForm.vue |
| 2026-03-07 | TASK-005 | ConnectionForm WinRM 配置 UI 完成 | ConnectionForm.vue |
| 2026-03-07 | TASK-006 | Store 更新支持 SSH/WinRM 字段 | connection.js |
| 2026-03-07 | TASK-007 | 后端 SSH 测试方法完成 | connection.go, ssh_tunnel.go |
| 2026-03-07 | TASK-008 | 后端 WinRM 测试方法完成 | connection.go, winrm.go |
| 2026-03-07 | TASK-009 | 前端 SSH/WinRM 测试按钮完成 | ConnectionForm.vue, connection.js |
| 2026-03-07 | TASK-010 | 智能连接测试完成 (含 SSH/WinRM 详细结果) | connection.go, ConnectionsPage.vue |
| 2026-03-07 | TASK-011~015 | Phase 3 UI 优化全部完成 | ConnectionsPage.vue, ConnectionForm.vue |
| 2026-03-07 | TASK-016 | 密码持久化验证完成 | connection_usecase.go, keyring/provider.go |
| 2026-03-07 | TASK-010 | 智能连接测试完成（含 SSH/WinRM 详细结果） | connection.go, ConnectionsPage.vue |

**Phase 2 完成 (100%)**

---

## 开发规范

1. **进度更新**: 每完成一个子任务，立即更新 tasks.md 中的勾选状态
2. **变更记录**: 任何功能变动必须记录到"变更记录"表格
3. **同步更新**: 如果变更影响需求，同步更新 spec.md
