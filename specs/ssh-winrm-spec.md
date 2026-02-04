# DB-BenchMind SSH & WinRM 需求文档

**版本**: 1.0.0
**日期**: 2026-02-04
**状态**: 需求定稿

---

## 文档变更历史

| 版本 | 日期 | 作者 | 变更说明 |
|------|------|------|---------|
| 1.0.0 | 2026-02-04 | Claude | 初始版本：SSH + WinRM 需求 |

---

## 1. 功能概述

### 1.1 背景

数据库压测场景中，目标数据库通常位于受保护的网络环境中，需要通过跳板机（SSH）或 Windows 宿主机（WinRM）进行访问。本需求定义了 SSH Tunnel 和 WinRM 连接的功能规范。

### 1.2 目标

- **SSH Tunnel**: 支持 MySQL、PostgreSQL、Oracle 通过 SSH 隧道连接
- **WinRM**: 支持 SQL Server 通过 WinRM 连接到 Windows 宿主机（后续性能监控预留接口）
- **统一体验**: 在 Connections 页面中像配置数据库连接一样配置隧道
- **独立测试**: 支持独立测试隧道和数据库连接
- **安全存储**: 隧道密码安全存储在 keyring 中

### 1.3 范围

#### 包含 (In Scope)
- SSH Tunnel 配置（MySQL、PostgreSQL、Oracle）
- WinRM 配置（SQL Server）
- 隧道连接测试
- 隧道配置持久化
- 密码安全存储
- UI 集成（Connections 页面）
- 性能监控预留接口（WinRM，后续 tasks 中实现）

#### 不包含 (Out of Scope)
- SSH 密钥认证（当前仅支持密码认证）
- WinRM 性能数据采集（后续阶段实现）
- 分布式隧道（多跳）
- 隧道性能监控

---

## 2. SSH Tunnel 功能规范

### 2.1 数据库支持

| 数据库类型 | SSH Tunnel 支持 | 优先级 |
|-----------|----------------|--------|
| MySQL     | ✅ 支持 | P0 |
| PostgreSQL| ✅ 支持 | P0 |
| Oracle    | ✅ 支持 | P0 |
| SQL Server| ❌ 不支持 | - |

### 2.2 SSH 配置字段

```go
type SSHTunnelConfig struct {
    Enabled  bool   `json:"enabled"`    // 是否启用 SSH Tunnel
    Host     string `json:"host"`       // SSH 服务器主机（使用 Database Host）
    Port     int    `json:"port"`       // SSH 服务器端口（默认 22）
    Username string `json:"username"`   // SSH 用户名（默认 "root"）
    Password string `json:"-"`          // SSH 密码（存储到 keyring）
    LocalPort int    `json:"local_port"` // 本地端口（0 = 自动分配）
}
```

### 2.3 SSH 功能需求

| 需求 ID | 需求描述 | 优先级 |
|---------|---------|--------|
| REQ-SSH-001 | WHEN 用户勾选 "Enable SSH Tunnel"，THE SYSTEM SHALL 显示 SSH 配置表单 | P0 |
| REQ-SSH-002 | WHEN SSH 配置显示，THE SYSTEM SHALL 使用 Database Host 作为 SSH Host（不显示单独字段） | P0 |
| REQ-SSH-003 | WHEN SSH Port 显示，THE SYSTEM SHALL 默认值为 22 | P0 |
| REQ-SSH-004 | WHEN SSH Username 显示，THE SYSTEM SHALL 默认值为 "root" | P0 |
| REQ-SSH-005 | WHEN SSH Password 显示，THE SYSTEM SHALL 使用密码掩码显示 | P0 |
| REQ-SSH-006 | WHEN Local Port 字段，THE SYSTEM SHALL 不显示（自动分配） | P0 |
| REQ-SSH-007 | WHEN 用户点击 "Test SSH"，THE SYSTEM SHALL 测试 SSH 隧道连接 | P0 |
| REQ-SSH-008 | WHEN SSH 测试成功，THE SYSTEM SHALL 显示 "SSH 连接成功" | P0 |
| REQ-SSH-009 | WHEN SSH 测试失败，THE SYSTEM SHALL 显示具体错误信息 | P0 |
| REQ-SSH-010 | WHEN 用户点击 "Test Database"，THE SYSTEM SHALL 测试数据库连接（不使用 SSH） | P0 |
| REQ-SSH-011 | WHEN 用户保存连接，THE SYSTEM SHALL 将 SSH 配置序列化到数据库 | P0 |
| REQ-SSH-012 | WHEN 用户保存连接，THE SYSTEM SHALL 将 SSH 密码存储到 keyring（key: `{conn_id}:ssh`） | P0 |
| REQ-SSH-013 | WHEN 用户编辑连接，THE SYSTEM SHALL 加载 SSH 配置和密码 | P0 |
| REQ-SSH-014 | WHEN 用户编辑连接，THE SYSTEM SHALL 选中 "Enable SSH Tunnel" 复选框（如已启用） | P0 |
| REQ-SSH-015 | WHEN Connections 列表显示，THE SYSTEM SHALL 显示 SSH 状态图标（🔒 SSH） | P0 |
| REQ-SSH-016 | WHEN Connections 列表 Test 按钮点击，THE SYSTEM SHALL 先测试 SSH，再测试数据库 | P0 |
| REQ-SSH-017 | WHEN SSH 连接失败，THE SYSTEM SHALL 自动测试直接数据库连接 | P0 |
| REQ-SSH-018 | WHEN 测试结果显示，THE SYSTEM SHALL 明确显示是通过 SSH 还是直接连接 | P0 |

### 2.4 SSH 连接流程

```
用户勾选 "Enable SSH Tunnel"
  → 显示 SSH 配置表单
  → 填写 SSH 配置
  → 点击 "Test SSH"
    → 测试 SSH 隧道
    → 成功：显示 "SSH 连接成功"
    → 失败：显示错误信息
  → 点击 "Test Database"
    → 测试数据库连接（不使用 SSH）
  → 点击 Save
    → 保存 SSH 配置到数据库
    → 保存 SSH 密码到 keyring
```

### 2.5 SSH Test 按钮逻辑

| 按钮 | 行为 | 位置 |
|------|------|------|
| Test SSH | 仅测试 SSH 隧道连接 | SSH 配置区域，Test Database 旁边 |
| Test Database | 仅测试数据库连接（不使用 SSH） | Test SSH 旁边 |
| Connections 列表 Test | 先测试 SSH（如启用），再测试数据库 | Connections 列表 |

---

## 3. WinRM 功能规范

### 3.1 数据库支持

| 数据库类型 | WinRM 支持 | 优先级 |
|-----------|------------|--------|
| SQL Server | ✅ 支持 | P0 |
| 其他数据库 | ❌ 不支持 | - |

### 3.2 WinRM 配置字段

```go
type WinRMConfig struct {
    Enabled  bool   `json:"enabled"`    // 是否启用 WinRM
    Host     string `json:"host"`       // WinRM 主机（使用 Database Host）
    Port     int    `json:"port"`       // WinRM 端口（5985 HTTP, 5986 HTTPS）
    Username string `json:"username"`   // 用户名（空 = 当前 Windows 用户）
    Password string `json:"-"`          // 密码（存储到 keyring）
    UseHTTPS bool   `json:"use_https"`  // 是否使用 HTTPS（默认 false）
}
```

### 3.3 WinRM 功能需求（当前阶段）

**当前阶段目标**: 配置和连接测试（不包含性能数据采集）

| 需求 ID | 需求描述 | 优先级 |
|---------|---------|--------|
| REQ-WINRM-001 | WHEN 用户勾选 "Enable WinRM"，THE SYSTEM SHALL 显示 WinRM 配置表单 | P0 |
| REQ-WINRM-002 | WHEN WinRM 配置显示，THE SYSTEM SHALL 使用 Database Host 作为 WinRM Host（不显示单独字段） | P0 |
| REQ-WINRM-003 | WHEN WinRM Port 显示，THE SYSTEM SHALL 默认值为 5985（HTTP） | P0 |
| REQ-WINRM-004 | WHEN Use HTTPS 勾选，THE SYSTEM SHALL 自动将端口改为 5986 | P0 |
| REQ-WINRM-005 | WHEN Username 为空，THE SYSTEM SHALL 使用当前 Windows 用户（集成 Windows 认证） | P0 |
| REQ-WINRM-006 | WHEN Username 不为空，THE SYSTEM SHALL 使用指定用户名和密码 | P0 |
| REQ-WINRM-007 | WHEN 用户点击 "Test WinRM"，THE SYSTEM SHALL 测试 WinRM 连接 | P0 |
| REQ-WINRM-008 | WHEN WinRM 测试成功，THE SYSTEM SHALL 显示 "WinRM 连接成功" | P0 |
| REQ-WINRM-009 | WHEN WinRM 测试失败，THE SYSTEM SHALL 显示具体错误信息 | P0 |
| REQ-WINRM-010 | WHEN 用户保存连接，THE SYSTEM SHALL 将 WinRM 配置序列化到数据库 | P0 |
| REQ-WINRM-011 | WHEN 用户保存连接，THE SYSTEM SHALL 将 WinRM 密码存储到 keyring（key: `{conn_id}:winrm`） | P0 |
| REQ-WINRM-012 | WHEN 用户编辑连接，THE SYSTEM SHALL 加载 WinRM 配置和密码 | P0 |
| REQ-WINRM-013 | WHEN 用户编辑连接，THE SYSTEM SHALL 选中 "Enable WinRM" 复选框（如已启用） | P0 |
| REQ-WINRM-014 | WHEN Connections 列表显示，THE SYSTEM SHALL 显示 WinRM 状态图标（🖥️ WinRM） | P0 |
| REQ-WINRM-015 | WHEN Connections 列表 Test 按钮点击，THE SYSTEM SHALL 先测试 WinRM（如启用），再测试数据库 | P0 |

### 3.4 WinRM 后续功能（预留接口，不在当前阶段）

**后续阶段目标**: 性能数据采集（在 tasks 中实现）

| 需求 ID | 需求描述 | 优先级 | 状态 |
|---------|---------|--------|------|
| REQ-WINRM-101 | WHEN 用户启动压测任务，THE SYSTEM SHALL 通过 WinRM 采集 Windows 宿主机性能指标 | P1 | 预留 |
| REQ-WINRM-102 | WHEN 性能指标采集，THE SYSTEM SHALL 包含 CPU、内存、磁盘、网络 | P1 | 预留 |
| REQ-WINRM-103 | WHEN 性能数据采集，THE SYSTEM SHALL 保存历史数据到数据库 | P1 | 预留 |
| REQ-WINRM-104 | WHEN 用户查看任务，THE SYSTEM SHALL 在 tasks 页面展示性能图表 | P1 | 预留 |

**性能指标定义**（预留）:
```go
type PerformanceMetrics struct {
    Timestamp    time.Time `json:"timestamp"`
    CPUPercent  float64   `json:"cpu_percent"`   // CPU 使用率
    MemoryPercent float64 `json:"memory_percent"` // 内存使用率
    DiskBytes    uint64    `json:"disk_bytes"`    // 磁盘 I/O
    NetworkBytes uint64    `json:"network_bytes"` // 网络 I/O
}
```

---

## 4. UI 设计规范

### 4.1 SSH Tunnel UI（已实现）

```
┌─────────────────────────────────────────────────┐
│ Database Type: [MySQL ▼]                        │
│                                                  │
│ Database Connection:                             │
│   Host: [localhost]                             │
│   Port: [3306]                                  │
│   Database: [testdb]                            │
│   Username: [root]                              │
│   Password: [••••••••]                           │
│                                                  │
│ [Test Database] [Test SSH]                      │
│                                                  │
│ [✓] Enable SSH Tunnel                           │
│                                                  │
│ SSH Configuration:                               │
│   SSH Port: [22]                                │
│   SSH Username: [root]                          │
│   SSH Password: [••••••••]                       │
│                                                  │
│ [Cancel]                        [Save]         │
└─────────────────────────────────────────────────┘
```

**按钮顺序**:
- Test Database（测试数据库，不使用 SSH）
- Test SSH（测试 SSH 隧道）

**SSH Configuration**:
- 不显示 SSH Host（使用 Database Host）
- 不显示 Local Port（自动分配）

### 4.2 WinRM UI（待实现）

```
┌─────────────────────────────────────────────────┐
│ Database Type: [SQL Server ▼]                   │
│                                                  │
│ Database Connection:                             │
│   Host: [192.168.1.100]                         │
│   Port: [1433]                                  │
│   Database: [TestDB]                            │
│   Username: [sa]                                │
│   Password: [••••••••]                           │
│                                                  │
│ [Test Database] [Test WinRM]                    │
│                                                  │
│ [✓] Enable WinRM                                │
│                                                  │
│ WinRM Configuration:                             │
│   WinRM Port: [5985]                            │
│   Use HTTPS: [ ]                                │
│   Username: [] (留空 = 当前 Windows 用户)        │
│   Password: [••••••••]                           │
│                                                  │
│ [Cancel]                        [Save]         │
└─────────────────────────────────────────────────┘
```

**按钮顺序**:
- Test Database（测试数据库，不使用 WinRM）
- Test WinRM（测试 WinRM 连接）

**WinRM Configuration**:
- 不显示 WinRM Host（使用 Database Host）
- Use HTTPS 勾选时自动更新端口为 5986

---

## 5. 连接列表显示规范

### 5.1 连接列表项显示

```
┌─────────────────────────────────────────────────┐
│ Connections                                     │
├─────────────────────────────────────────────────┤
│ 🐭 MySQL-Prod | root@192.168.1.10:3306 | 🔒 SSH │
│ 🐘 Oracle-Test | system@localhost:1521          │
│ 🐧 PostgreSQL-Dev | postgres@db:5432 | 🔒 SSH   │
│ 🔵 SQL Server-Win | sa@192.168.1.100:1433 | 🖥️ WinRM │
└─────────────────────────────────────────────────┘
```

**图标说明**:
- 🐭 MySQL
- 🐘 Oracle
- 🐧 PostgreSQL
- 🔵 SQL Server
- 🔒 SSH（已启用）
- 🖥️ WinRM（已启用）

### 5.2 Test 按钮测试结果

**SSH 测试结果示例**:
```
SSH: ✅ Connected (45ms)
Database: ✅ Connected via SSH (120ms)
Version: MySQL 8.0.35
```

**SSH 失败，直接连接成功**:
```
SSH: ❌ Failed - Authentication failed (15ms)
Database: ✅ Connected (Direct, without SSH) (115ms)
Version: MySQL 8.0.35
⚠️ SSH tunnel was not used
```

**WinRM 测试结果示例**（预留）:
```
WinRM: ✅ Connected (60ms)
Database: ✅ Connected (180ms)
Version: Microsoft SQL Server 2022
```

---

## 6. 数据持久化规范

### 6.1 SSH 配置序列化

```go
// MySQL 连接序列化示例
{
  "id": "uuid-123",
  "name": "MySQL-Prod",
  "type": "mysql",
  "host": "192.168.1.10",
  "port": 3306,
  "database": "testdb",
  "username": "root",
  "ssl_mode": "preferred",
  "ssh": {
    "enabled": true,
    "host": "192.168.1.10",
    "port": 22,
    "username": "root",
    "local_port": 0
  },
  "created_at": "2026-02-04T10:00:00Z",
  "updated_at": "2026-02-04T10:00:00Z"
}
```

### 6.2 WinRM 配置序列化

```go
// SQL Server 连接序列化示例
{
  "id": "uuid-456",
  "name": "SQL Server-Win",
  "type": "sqlserver",
  "host": "192.168.1.100",
  "port": 1433,
  "database": "TestDB",
  "username": "sa",
  "trust_server_certificate": true,
  "winrm": {
    "enabled": true,
    "host": "192.168.1.100",
    "port": 5985,
    "username": "",
    "use_https": false
  },
  "created_at": "2026-02-04T10:00:00Z",
  "updated_at": "2026-02-04T10:00:00Z"
}
```

### 6.3 Keyring 存储规范

| Key | 内容 | 示例 |
|-----|------|------|
| `{conn_id}` | 数据库密码 | `uuid-123` → `db_password` |
| `{conn_id}:ssh` | SSH 密码 | `uuid-123:ssh` → `ssh_password` |
| `{conn_id}:winrm` | WinRM 密码 | `uuid-456:winrm` → `winrm_password` |

---

## 7. 测试场景

### 7.1 SSH 测试场景

| 场景 | 输入 | 预期输出 |
|------|------|---------|
| SSH 连接成功 | 正确的 SSH 配置 | "SSH 连接成功" |
| SSH 认证失败 | 错误的密码 | "SSH 认证失败" |
| SSH 连接超时 | 无法访问的主机 | "SSH 连接超时" |
| SSH 端口错误 | 错误的端口 | "SSH 连接被拒绝" |
| 数据库通过 SSH | 启用 SSH，测试数据库 | 通过 SSH 隧道连接 |
| SSH 失败回退 | SSH 失败，测试数据库 | 自动测试直接连接 |

### 7.2 WinRM 测试场景

| 场景 | 输入 | 预期输出 |
|------|------|---------|
| WinRM HTTP 成功 | 端口 5985 | "WinRM 连接成功" |
| WinRM HTTPS 成功 | 端口 5986，Use HTTPS | "WinRM 连接成功" |
| WinRM 认证失败 | 错误的密码 | "WinRM 认证失败" |
| WinRM 连接超时 | 无法访问的主机 | "WinRM 连接超时" |
| 集成 Windows 认证 | Username 为空 | 使用当前 Windows 用户 |

---

## 8. 可追溯性

### 8.1 需求 → 实现

| 需求 ID | 实现文件 | 状态 |
|---------|---------|------|
| REQ-SSH-001 ~ REQ-SSH-018 | `internal/domain/connection/ssh_tunnel.go` | ✅ 已实现 |
| REQ-SSH-001 ~ REQ-SSH-018 | `internal/transport/ui/pages/connection_page.go` | ✅ 已实现 |
| REQ-WINRM-001 ~ REQ-WINRM-015 | `internal/domain/connection/winrm.go` | 🚧 待实现 |
| REQ-WINRM-001 ~ REQ-WINRM-015 | `internal/transport/ui/pages/connection_page.go` | 🚧 待实现 |

---

## 9. 验收标准

### 9.1 SSH 验收标准

- [x] 支持 MySQL、PostgreSQL、Oracle 的 SSH Tunnel
- [x] SSH 配置字段正确显示
- [x] SSH 密码安全存储到 keyring
- [x] SSH 连接测试正常工作
- [x] SSH 失败时自动测试直接数据库连接
- [x] Connections 列表显示 SSH 状态图标
- [x] Edit 连接时正确加载 SSH 配置

### 9.2 WinRM 验收标准（当前阶段）

- [ ] SQL Server 连接支持 WinRM 配置
- [ ] WinRM 配置字段正确显示
- [ ] WinRM 密码安全存储到 keyring
- [ ] WinRM 连接测试正常工作
- [ ] Connections 列表显示 WinRM 状态图标
- [ ] Edit 连接时正确加载 WinRM 配置

---

**文档结束**
