# Connections Page Refactor Specification

## 概述

本文档定义 Wails 版本 Connections 页面的功能规格，目标是与原 Fyne 版本功能对齐。

## 当前状态

Wails 版本完成度约 40%，缺失以下核心功能：
- SSH Tunnel 配置
- WinRM 配置（SQL Server）
- 分组显示和 UI 优化
- 数据库特定字段处理

---

## 功能需求

### REQ-001: SSH Tunnel 配置

**优先级**: P0 (核心功能)

**描述**: 
WHEN 用户创建或编辑 MySQL/PostgreSQL/Oracle 连接时,
THE SYSTEM SHALL 提供 SSH Tunnel 配置选项。

**字段**:
| 字段 | 类型 | 默认值 | 说明 |
|-----|------|-------|------|
| Enable SSH | boolean | false | 启用 SSH 隧道 |
| SSH Host | string | (使用 Database Host) | SSH 服务器地址 |
| SSH Port | number | 22 | SSH 端口 |
| SSH Username | string | "root" | SSH 用户名 |
| SSH Password | password | - | SSH 密码（存储到 keyring） |

**验证规则**:
- SSH Host 使用 Database Host 的值，无需单独输入
- SSH Password 必须安全存储到 keyring

---

### REQ-002: WinRM 配置

**优先级**: P0 (核心功能)

**描述**:
WHEN 用户创建或编辑 SQL Server 连接时,
THE SYSTEM SHALL 提供 WinRM (Windows Remote Management) 配置选项。

**字段**:
| 字段 | 类型 | 默认值 | 说明 |
|-----|------|-------|------|
| Enable WinRM | boolean | false | 启用 WinRM |
| WinRM Port | number | 5985 (HTTP) / 5986 (HTTPS) | WinRM 端口 |
| Use HTTPS | boolean | false | 使用 HTTPS 连接 |
| WinRM Username | string | - | WinRM 用户名（空 = Windows 集成认证） |
| WinRM Password | password | - | WinRM 密码（存储到 keyring） |

**交互行为**:
- WHEN 用户勾选 "Use HTTPS", THE SYSTEM SHALL 自动将端口更新为 5986
- WHEN 用户取消 "Use HTTPS", THE SYSTEM SHALL 自动将端口更新为 5985

---

### REQ-003: Trust Server Certificate (SQL Server)

**优先级**: P1 (重要功能)

**描述**:
WHEN 用户创建或编辑 SQL Server 连接时,
THE SYSTEM SHALL 提供 "Trust Server Certificate" 选项。

**字段**:
| 字段 | 类型 | 默认值 | 说明 |
|-----|------|-------|------|
| Trust Server Certificate | boolean | true | 信任服务器证书 |

---

### REQ-004: 连接列表分组显示

**优先级**: P1 (重要功能)

**描述**:
WHEN 用户查看连接列表时,
THE SYSTEM SHALL 按数据库类型分组显示连接。

**分组顺序**:
1. MySQL
2. PostgreSQL
3. Oracle
4. SQL Server

**分组特性**:
- 每个分组显示数据库名称和连接数量
- 分组可折叠/展开
- 每个分组使用对应的数据库图标

---

### REQ-005: 数据库图标

**优先级**: P2 (UI 增强)

**描述**:
WHEN 用户查看连接列表时,
THE SYSTEM SHALL 为每种数据库类型显示对应图标。

**图标映射**:
| 数据库类型 | 图标 |
|-----------|------|
| MySQL | 🐬 |
| PostgreSQL | 🐘 |
| Oracle | 🔴 |
| SQL Server | 🔷 |

---

### REQ-006: SSH/WinRM 状态指示器

**优先级**: P2 (UI 增强)

**描述**:
WHEN 连接配置了 SSH Tunnel 或 WinRM 时,
THE SYSTEM SHALL 在连接列表中显示状态指示器。

**显示规则**:
- SSH 启用时显示: `🔒 SSH`
- WinRM 启用时显示: `🖥️ WinRM`

---

### REQ-007: 动态标签 (Database/SID)

**优先级**: P1 (重要功能)

**描述**:
WHEN 用户选择不同数据库类型时,
THE SYSTEM SHALL 动态更新表单标签。

**标签映射**:
| 数据库类型 | 标签 |
|-----------|------|
| MySQL | Database |
| PostgreSQL | Database |
| Oracle | SID |
| SQL Server | Database |

---

### REQ-008: 数据库特定默认值

**优先级**: P2 (UI 增强)

**描述**:
WHEN 用户创建新连接时,
THE SYSTEM SHALL 根据数据库类型设置默认值。

**默认值**:
| 数据库类型 | 字段 | 默认值 |
|-----------|------|-------|
| PostgreSQL | Database | "postgres" |
| Oracle | SID | "orcl" |

---

### REQ-009: SSH 连接测试

**优先级**: P0 (核心功能)

**描述**:
WHEN 用户点击 "Test SSH" 按钮时,
THE SYSTEM SHALL 测试 SSH Tunnel 连接并显示结果。

**测试内容**:
- SSH 连接是否成功
- SSH 连接延迟

**显示结果**:
```
SSH TUNNEL
  Status: ✓ Connected / ✗ Failed
  Host: <ssh_host>
  Port: <ssh_port>
  User: <ssh_user>
  Latency: <latency_ms>ms
  Error: <error_message> (如果失败)
```

---

### REQ-010: WinRM 连接测试

**优先级**: P0 (核心功能)

**描述**:
WHEN 用户点击 "Test WinRM" 按钮时,
THE SYSTEM SHALL 测试 WinRM 连接并显示结果。

**测试内容**:
- WinRM 连接是否成功
- WinRM 连接延迟

---

### REQ-011: 智能连接测试

**优先级**: P1 (重要功能)

**描述**:
WHEN 用户测试配置了 SSH 的连接时,
THE SYSTEM SHALL 执行智能测试流程。

**测试流程**:
1. 首先测试 SSH Tunnel
2. IF SSH 成功, THEN 通过 SSH 测试数据库连接
3. IF SSH 失败, THEN 尝试直接连接数据库
4. 显示详细测试结果，包含 SSH 和数据库两部分

**显示结果示例**:
```
Connection Test Results: <connection_name>

━━━━━━━━━━━━━━━━━━━━━━━━━━━
📡 SSH TUNNEL
  Status: ✓ Connected
  Host: 192.168.1.100
  Port: 22
  User: root
  Latency: 15ms

━━━━━━━━━━━━━━━━━━━━━━━━━━━
💾 DATABASE
  Status: ✓ Connected
  Version: 8.0.33
  Latency: 45ms
```

---

### REQ-012: 密码持久化

**优先级**: P0 (核心功能)

**描述**:
WHEN 用户保存连接时,
THE SYSTEM SHALL 将密码安全存储到系统 keyring。

**存储内容**:
- 数据库密码: `<connection_id>`
- SSH 密码: `<connection_id>:ssh`
- WinRM 密码: `<connection_id>:winrm`

**WHEN 用户编辑连接时, THE SYSTEM SHALL 从 keyring 加载已保存的密码。**

---

## 非功能需求

### NFR-001: 安全性
- 所有密码必须存储到系统 keyring，禁止明文存储
- 密码传输必须加密

### NFR-002: 用户体验
- 表单字段变化时即时反馈
- 测试操作显示进度指示
- 错误信息清晰易懂

### NFR-003: 兼容性
- 保持与现有 Fyne 版本数据格式兼容
- 支持从现有数据迁移

---

## 追溯矩阵

| 需求 ID | 前端组件 | 后端 Binding | 测试 |
|--------|---------|-------------|------|
| REQ-001 | ConnectionForm.vue | ConnectionBinding | UT-001 |
| REQ-002 | ConnectionForm.vue | ConnectionBinding | UT-002 |
| REQ-003 | ConnectionForm.vue | ConnectionBinding | UT-003 |
| REQ-004 | ConnectionsPage.vue | - | UT-004 |
| REQ-005 | ConnectionsPage.vue | - | UT-005 |
| REQ-006 | ConnectionsPage.vue | - | UT-006 |
| REQ-007 | ConnectionForm.vue | - | UT-007 |
| REQ-008 | ConnectionForm.vue | - | UT-008 |
| REQ-009 | ConnectionForm.vue | ConnectionBinding | UT-009 |
| REQ-010 | ConnectionForm.vue | ConnectionBinding | UT-010 |
| REQ-011 | ConnectionsPage.vue | ConnectionBinding | UT-011 |
| REQ-012 | - | ConnectionBinding | UT-012 |
