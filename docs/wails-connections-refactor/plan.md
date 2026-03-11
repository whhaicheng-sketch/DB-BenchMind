# Connections Page Refactor Implementation Plan

## 概述

本文档描述 Wails 版本 Connections 页面的实现计划，分为 4 个阶段。

---

## 阶段划分

```
Phase 1: 核心功能 (P0)
├── 后端 SSH/WinRM Binding
├── 前端 SSH 配置表单
└── 前端 WinRM 配置表单

Phase 2: 测试功能 (P0/P1)
├── SSH 连接测试
├── WinRM 连接测试
└── 智能连接测试

Phase 3: UI 优化 (P1/P2)
├── 分组显示
├── 数据库图标
├── 状态指示器
└── 动态标签和默认值

Phase 4: 完善和测试
├── 密码持久化完善
├── 单元测试
└── 集成测试
```

---

## Phase 1: 核心功能

### 1.1 后端 SSH/WinRM Binding

**目标**: 扩展 ConnectionBinding 支持 SSH 和 WinRM 配置

**变更文件**:
```
internal/transportwails/bindings/connection.go  # 扩展 DTO 和方法
internal/domain/connection/                     # 确认数据结构支持
```

**DTO 扩展**:
```go
// ConnectionCreateRequest 扩展
type ConnectionCreateRequest struct {
    // ... existing fields ...
    
    // SSH Configuration
    SSHEnabled   bool   `json:"ssh_enabled"`
    SSHPort      int    `json:"ssh_port"`
    SSHUsername  string `json:"ssh_username"`
    SSHPassword  string `json:"ssh_password"`
    
    // WinRM Configuration (SQL Server only)
    WinRMEnabled   bool   `json:"winrm_enabled"`
    WinRMPort      int    `json:"winrm_port"`
    WinRMUseHTTPS  bool   `json:"winrm_use_https"`
    WinRMUsername  string `json:"winrm_username"`
    WinRMPassword  string `json:"winrm_password"`
    
    // SQL Server specific
    TrustServerCertificate bool `json:"trust_server_certificate"`
}
```

**依赖**: 无

**风险**: 
- 数据结构变更需要考虑向后兼容

---

### 1.2 前端 SSH 配置表单

**目标**: 在 ConnectionForm.vue 添加 SSH 配置字段

**变更文件**:
```
frontend/src/components/connection/ConnectionForm.vue
frontend/src/stores/connection.js
```

**UI 布局**:
```
┌─────────────────────────────────────┐
│ Database Type: [MySQL          ▼]   │
│ Name:         [_________________]   │
│ Host:         [_________________]   │
│ Port:         [3306            ]   │
│ Database:     [_________________]   │
│ Username:     [_________________]   │
│ Password:     [_________________]   │
│ SSL Mode:     [Preferred       ▼]   │
├─────────────────────────────────────┤
│ ☑ Enable SSH Tunnel                 │
│ ┌─────────────────────────────────┐ │
│ │ SSH Configuration               │ │
│ │ SSH Port:    [22             ]  │ │
│ │ SSH Username:[root           ]  │ │
│ │ SSH Password:[______________]  │ │
│ │ 💡 SSH Host uses Database Host  │ │
│ └─────────────────────────────────┘ │
├─────────────────────────────────────┤
│ [Test SSH] [Test Database]          │
│ [Cancel]  [Save]                    │
└─────────────────────────────────────┘
```

**交互逻辑**:
- SSH 配置区域默认隐藏
- 勾选 "Enable SSH Tunnel" 后显示
- SQL Server 类型时隐藏 SSH 选项

---

### 1.3 前端 WinRM 配置表单

**目标**: 在 ConnectionForm.vue 添加 WinRM 配置字段（仅 SQL Server）

**UI 布局**:
```
┌─────────────────────────────────────┐
│ ... (其他字段)                       │
│ ☑ Trust Server Certificate          │
├─────────────────────────────────────┤
│ ☑ Enable WinRM (Windows Remote Mgmt)│
│ ┌─────────────────────────────────┐ │
│ │ WinRM Configuration             │ │
│ │ WinRM Port:  [5985           ]  │ │
│ │ ☑ Use HTTPS                    │ │
│ │ Username:    [________________]  │ │
│ │ Password:    [________________]  │ │
│ │ 💡 WinRM Host uses Database Host│ │
│ └─────────────────────────────────┘ │
├─────────────────────────────────────┤
│ [Test WinRM] [Test Database]        │
│ [Cancel]  [Save]                    │
└─────────────────────────────────────┘
```

**交互逻辑**:
- WinRM 配置区域默认隐藏
- 仅当 Database Type = SQL Server 时显示
- Use HTTPS 勾选时自动更新端口为 5986

---

## Phase 2: 测试功能

### 2.1 SSH 连接测试

**目标**: 实现 TestSSH 后端方法和前端调用

**变更文件**:
```
internal/transportwails/bindings/connection.go
frontend/src/stores/connection.js
frontend/src/components/connection/ConnectionForm.vue
```

**后端方法**:
```go
func (b *ConnectionBinding) TestSSHConnection(ctx context.Context, req SSHTTestRequest) (*SSHTTestResult, error)

type SSHTTestResult struct {
    Success  bool   `json:"success"`
    Host     string `json:"host"`
    Port     int    `json:"port"`
    Username string `json:"username"`
    Latency  int64  `json:"latency_ms"`
    Error    string `json:"error,omitempty"`
}
```

---

### 2.2 WinRM 连接测试

**目标**: 实现 TestWinRM 后端方法和前端调用

**后端方法**:
```go
func (b *ConnectionBinding) TestWinRMConnection(ctx context.Context, req WinRMTestRequest) (*WinRMTestResult, error)
```

---

### 2.3 智能连接测试

**目标**: 增强现有 TestConnection 方法，支持智能测试流程

**测试流程**:
1. 检查是否启用 SSH/WinRM
2. 先测试隧道连接
3. 再测试数据库连接
4. 返回详细结果

**结果 DTO**:
```go
type ConnectionTestResult struct {
    Success         bool   `json:"success"`
    LatencyMs       int64  `json:"latency_ms"`
    DatabaseVersion string `json:"database_version"`
    Error           string `json:"error,omitempty"`
    
    // 新增字段
    SSHTested       bool           `json:"ssh_tested"`
    SSHResult       *SSHTTestResult `json:"ssh_result,omitempty"`
    DirectConnect   bool           `json:"direct_connect"` // 是否直连（SSH失败时）
}
```

---

## Phase 3: UI 优化

### 3.1 分组显示

**目标**: 按数据库类型分组显示连接列表

**变更文件**:
```
frontend/src/components/pages/ConnectionsPage.vue
```

**布局**:
```
┌─────────────────────────────────────┐
│ [+ New Connection]    [Search...]   │
├─────────────────────────────────────┤
│ 📊 MySQL (2)                    [▼] │
│   🐬 MySQL5.7 | root@192.168.1.100:3306 🔒 SSH
│       [Test] [Edit] [Delete]
│   🐬 MySQL8.0 | root@192.168.1.101:3306
│       [Test] [Edit] [Delete]
├─────────────────────────────────────┤
│ 📊 PostgreSQL (1)               [▼] │
│   🐘 PostgreSQL13 | postgres@... 🔒 SSH
│       [Test] [Edit] [Delete]
├─────────────────────────────────────┤
│ 📊 Oracle (1)                   [▼] │
│   🔴 Oracle11g | system@... 🔒 SSH
│       [Test] [Edit] [Delete]
├─────────────────────────────────────┤
│ 📊 SQL Server (1)               [▼] │
│   🔷 SQLServer2019 | sa@... 🖥️ WinRM
│       [Test] [Edit] [Delete]
└─────────────────────────────────────┘
```

---

### 3.2 数据库图标和状态指示器

**目标**: 添加数据库图标和 SSH/WinRM 状态指示

**实现方式**:
- 使用 CSS 类或 emoji 图标
- 根据 connection.type 和 ssh_enabled/winrm_enabled 状态显示

---

### 3.3 动态标签和默认值

**目标**: 根据数据库类型动态更新表单

**实现**:
- 监听 formData.type 变化
- 更新标签文本
- 设置默认值

---

## Phase 4: 完善和测试

### 4.1 密码持久化完善

**目标**: 确保 SSH 和 WinRM 密码正确存储和加载

**验证点**:
- 创建连接时保存密码到 keyring
- 编辑连接时从 keyring 加载密码
- 更新连接时更新 keyring 中的密码

---

### 4.2 单元测试

**目标**: 为新增功能编写单元测试

**测试覆盖**:
- ConnectionBinding 方法测试
- ConnectionForm 组件测试
- Store actions 测试

---

### 4.3 集成测试

**目标**: 端到端功能验证

**测试场景**:
1. 创建带 SSH 的 MySQL 连接
2. 测试 SSH 连接
3. 测试数据库连接
4. 编辑连接，修改 SSH 配置
5. 删除连接

---

## 风险评估

| 风险 | 概率 | 影响 | 缓解措施 |
|-----|------|------|---------|
| 数据结构变更导致不兼容 | 中 | 高 | 保持向后兼容，添加迁移逻辑 |
| Keyring 在某些环境不可用 | 低 | 高 | 提供降级方案（加密文件存储） |
| SSH/WinRM 连接超时 | 中 | 中 | 设置合理超时，提供重试选项 |

---

## 时间估算

| 阶段 | 预估工时 |
|-----|---------|
| Phase 1: 核心功能 | 4-6 小时 |
| Phase 2: 测试功能 | 3-4 小时 |
| Phase 3: UI 优化 | 2-3 小时 |
| Phase 4: 完善和测试 | 2-3 小时 |
| **总计** | **11-16 小时** |
