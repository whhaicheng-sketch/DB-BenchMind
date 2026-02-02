# DB-BenchMind PostgreSQL 支持规格文档

**版本**: 1.0.0
**日期**: 2026-02-02
**状态**: 待评审

---

## 文档变更历史

| 版本 | 日期 | 作者 | 变更说明 |
|------|------|------|---------|
| 1.0.0 | 2026-02-02 | Claude | 初始版本 - PostgreSQL 连接与压测支持 |

---

## 1. 功能概述

### 1.1 目标

为 DB-BenchMind 添加完整的 **PostgreSQL** 数据库支持，使系统能够：
- 管理 PostgreSQL 数据库连接
- 对 PostgreSQL 执行 Sysbench 压测
- 存储和对比 PostgreSQL 压测结果

### 1.2 范围

**包含**:
- PostgreSQL 连接管理（CRUD + 测试）
- PostgreSQL 驱动集成
- Sysbench PostgreSQL 适配器
- SSL 连接支持
- 连接参数验证

**不包含**:
- Swingbench PostgreSQL 支持（Swingbench 仅支持 Oracle）
- HammerDB PostgreSQL 支持（已有基础，本版本不测试）
- PostgreSQL 特有监控指标（延迟分析等）

### 1.3 背景

当前系统已完成 **MySQL** 的完整支持，包括：
- ✅ MySQL 连接管理
- ✅ MySQL 驱动集成 (`github.com/go-sql-driver/mysql`)
- ✅ MySQL 连接测试实现
- ✅ Sysbench MySQL 适配器
- ✅ 完整的 UI 支持

本版本将同等级别的支持扩展到 PostgreSQL。

---

## 2. 核心概念定义

| 概念 | PostgreSQL 特性 | 示例 |
|------|-----------------|------|
| **连接字符串** | `host=host port=port database=db user=user password=pass sslmode=mode` | `host=localhost port=5432 database=testdb user=postgres` |
| **SSL Mode** | disable, allow, prefer, require, verify-ca, verify-full | `prefer`（默认） |
| **默认端口** | 5432 | localhost:5432 |
| **系统数据库** | postgres | 用于连接测试 |
| **驱动** | github.com/lib/pq | 纯 Go PostgreSQL 驱动 |

---

## 3. 功能需求

### 3.1 连接管理（Connections 页面）

#### 3.1.1 创建 PostgreSQL 连接

| 需求 ID | 需求描述 | 优先级 |
|---------|---------|--------|
| REQ-PG-CONN-001 | WHEN 用户在 Connections 页面选择 "PostgreSQL" 数据类型，THE SYSTEM SHALL 显示 PostgreSQL 连接表单 | P0 |
| REQ-PG-CONN-002 | WHEN PostgreSQL 连接表单加载，THE SYSTEM SHALL 设置默认端口为 5432 | P0 |
| REQ-PG-CONN-003 | WHEN 用户填写 PostgreSQL 连接表单，THE SYSTEM SHALL 验证以下字段：Name, Host, Port, Username | P0 |
| REQ-PG-CONN-004 | WHEN Database 字段为空，THE SYSTEM SHALL 允许保存（PostgreSQL 可连接不指定数据库） | P0 |
| REQ-PG-CONN-005 | WHEN SSL Mode 字段，THE SYSTEM SHALL 提供选项：disable, allow, prefer, require, verify-ca, verify-full | P0 |
| REQ-PG-CONN-006 | WHEN SSL Mode 为空，THE SYSTEM SHALL 使用默认值 "prefer" | P0 |
| REQ-PG-CONN-007 | WHEN 用户保存 PostgreSQL 连接，THE SYSTEM SHALL 存储连接信息到 SQLite 数据库 | P0 |
| REQ-PG-CONN-008 | WHEN 用户保存 PostgreSQL 连接，THE SYSTEM SHALL 使用 keyring 存储密码 | P0 |

#### 3.1.2 测试 PostgreSQL 连接

| 需求 ID | 需求描述 | 优先级 |
|---------|---------|--------|
| REQ-PG-CONN-010 | WHEN 用户点击 "Test Connection" 按钮，THE SYSTEM SHALL 使用 `github.com/lib/pq` 驱动建立连接 | P0 |
| REQ-PG-CONN-011 | WHEN 连接测试成功，THE SYSTEM SHALL 返回成功状态、延迟（ms）、PostgreSQL 版本号 | P0 |
| REQ-PG-CONN-012 | WHEN 连接测试失败，THE SYSTEM SHALL 返回失败状态和具体错误信息（如 "connection refused", "authentication failed"） | P0 |
| REQ-PG-CONN-013 | WHEN 测试 PostgreSQL 连接，THE SYSTEM SHALL 使用用户指定的 Database 字段，如果为空则使用 "postgres" | P0 |
| REQ-PG-CONN-014 | WHEN SSL Mode 为 require/verify-ca/verify-full，THE SYSTEM SHALL 强制使用 SSL 连接 | P0 |

#### 3.1.3 编辑与删除 PostgreSQL 连接

| 需求 ID | 需求描述 | 优先级 |
|---------|---------|--------|
| REQ-PG-CONN-020 | WHEN 用户编辑 PostgreSQL 连接，THE SYSTEM SHALL 预填充所有字段（Name, Host, Port, Database, Username, SSL Mode） | P0 |
| REQ-PG-CONN-021 | WHEN 用户编辑 PostgreSQL 连接，THE SYSTEM SHALL 从 keyring 加载密码但不显示在 UI | P0 |
| REQ-PG-CONN-022 | WHEN 用户删除 PostgreSQL 连接，THE SYSTEM SHALL 从数据库删除连接信息 | P0 |
| REQ-PG-CONN-023 | WHEN 用户删除 PostgreSQL 连接，THE SYSTEM SHALL 从 keyring 删除密码 | P0 |

---

### 3.2 压测执行（Sysbench 适配器）

#### 3.2.1 Sysbench PostgreSQL 命令生成

| 需求 ID | 需求描述 | 优先级 |
|---------|---------|--------|
| REQ-PG-SYS-001 | WHEN 用户选择 PostgreSQL 连接和 Sysbench 模板，THE SYSTEM SHALL 生成 `--pgsql-*` 参数的命令 | P0 |
| REQ-PG-SYS-002 | WHEN 生成 Sysbench PostgreSQL 命令，THE SYSTEM SHALL 包含：--pgsql-host, --pgsql-port, --pgsql-user, --pgsql-password, --pgsql-db | P0 |
| REQ-PG-SYS-003 | WHEN Sysbench PostgreSQL 命令执行，THE SYSTEM SHALL 设置 `PGPASSWORD` 环境变量传递密码 | P0 |
| REQ-PG-SYS-004 | WHEN PostgreSQL 连接的 Database 字段为空，THE SYSTEM SHALL 使用 "postgres" 作为数据库名 | P0 |
| REQ-PG-SYS-005 | WHEN 用户执行 Sysbench prepare，THE SYSTEM SHALL 使用 `psql` 命令创建数据库 | P0 |
| REQ-PG-SYS-006 | WHEN Sysbench prepare 执行 psql 命令，THE SYSTEM SHALL 格式：`psql -h host -p port -U user -c "CREATE DATABASE db;"` | P0 |

#### 3.2.2 结果解析与存储

| 需求 ID | 需求描述 | 优先级 |
|---------|---------|--------|
| REQ-PG-SYS-010 | WHEN Sysbench PostgreSQL 执行完成，THE SYSTEM SHALL 解析输出获取 TPS, QPS, Latency 等指标 | P0 |
| REQ-PG-SYS-011 | WHEN 存储 PostgreSQL 压测结果，THE SYSTEM SHALL 记录 Database Type = "postgresql" | P0 |
| REQ-PG-SYS-012 | WHEN PostgreSQL 结果显示在 History 页面，THE SYSTEM SHALL 显示连接名称、数据库类型、线程数等 | P0 |

---

### 3.3 UI/UX 需求

#### 3.3.1 连接表单布局

| 需求 ID | 需求描述 | 优先级 |
|---------|---------|--------|
| REQ-PG-UI-001 | WHEN PostgreSQL 连接表单显示，THE SYSTEM SHALL 包含字段：Name, Host, Port, Database, Username, Password, SSL Mode | P0 |
| REQ-PG-UI-002 | WHEN 用户切换数据库类型到 PostgreSQL，THE SYSTEM SHALL 自动更新默认端口为 5432 | P0 |
| REQ-PG-UI-003 | WHEN PostgreSQL 连接显示在列表，THE SYSTEM SHALL 格式：`ConnectionName (Username@Host:Port/Database)` | P0 |
| REQ-PG-UI-004 | WHEN PostgreSQL 连接列表为空，THE SYSTEM SHALL 在 "PostgreSQL" 分组显示 "No connections" 提示 | P0 |

#### 3.3.2 错误提示与反馈

| 需求 ID | 需求描述 | 优先级 |
|---------|---------|--------|
| REQ-PG-UI-010 | WHEN PostgreSQL 连接测试失败，THE SYSTEM SHALL 显示友好的错误消息（如 "无法连接到数据库服务器"） | P0 |
| REQ-PG-UI-011 | WHEN PostgreSQL 连接测试超时（>10s），THE SYSTEM SHALL 显示 "连接超时，请检查网络和防火墙" | P0 |
| REQ-PG-UI-012 | WHEN PostgreSQL 驱动未安装，THE SYSTEM SHALL 显示 "PostgreSQL 驱动不可用，请运行 go get github.com/lib/pq" | P0 |

---

## 4. 非功能需求

### 4.1 性能需求

| 需求 ID | 需求描述 | 优先级 |
|---------|---------|--------|
| REQ-PG-NFR-001 | PostgreSQL 连接测试应在 5 秒内完成（成功或超时） | P0 |
| REQ-PG-NFR-002 | Sysbench PostgreSQL 命令生成应在 100ms 内完成 | P0 |

### 4.2 安全需求

| 需求 ID | 需求描述 | 优先级 |
|---------|---------|--------|
| REQ-PG-NFR-010 | PostgreSQL 密码不得以明文形式存储在 SQLite 数据库 | P0 |
| REQ-PG-NFR-011 | PostgreSQL 密码不得以明文形式显示在 UI 或日志中 | P0 |
| REQ-PG-NFR-012 | WHEN SSL Mode = verify-ca 或 verify-full，THE SYSTEM SHALL 验证服务器证书 | P0 |

### 4.3 兼容性需求

| 需求 ID | 需求描述 | 优先级 |
|---------|---------|--------|
| REQ-PG-NFR-020 | 支持 PostgreSQL 9.6+ | P0 |
| REQ-PG-NFR-021 | 支持连接到远程 PostgreSQL 服务器 | P0 |
| REQ-PG-NFR-022 | 支持通过 SSH 隧道的本地端口转发连接 | P1 |

---

## 5. 数据模型

### 5.1 PostgreSQL 连接 Schema

```go
type PostgreSQLConnection struct {
    BaseConnection

    // Connection parameters
    Host     string `json:"host"`     // e.g., "localhost"
    Port     int    `json:"port"`     // Default: 5432
    Database string `json:"database"` // Optional, can be empty
    Username string `json:"username"` // e.g., "postgres"
    Password string `json:"-"`        // Stored in keyring
    SSLMode  string `json:"ssl_mode"` // disable, allow, prefer, require, verify-ca, verify-full
}
```

### 5.2 连接测试结果 Schema

```go
type TestResult struct {
    Success   bool    `json:"success"`
    LatencyMs int64   `json:"latency_ms"`
    Version   string  `json:"version"`   // e.g., "PostgreSQL 13.14"
    Error     string  `json:"error"`     // Error message if failed
}
```

---

## 6. 验收标准

### 6.1 基本功能验收

**场景 1: 创建 PostgreSQL 连接**
1. 打开 DB-BenchMind，进入 Connections 页面
2. 点击 "Add Connection"
3. 选择 Database Type = "PostgreSQL"
4. 填写：Name=Test PG, Host=localhost, Port=5432, Database=testdb, Username=postgres, Password=***
5. 选择 SSL Mode = "prefer"
6. 点击 "Test Connection"
   **预期**: 显示成功，延迟 <1000ms，版本 = "PostgreSQL xx.x"
7. 点击 "Save"
   **预期**: 连接保存到数据库，密码存储到 keyring

**场景 2: 执行 PostgreSQL 压测**
1. 进入 Tasks 页面
2. 选择：Connection=Test PG, Template=Sysbench OLTP Read-Write
3. 配置：Threads=4, Time=60
4. 点击 "Start"
   **预期**: Sysbench 命令包含 `--pgsql-host=localhost --pgsql-port=5432 --pgsql-user=postgres --pgsql-db=testdb`
5. 等待完成
   **预期**: History 页面显示新记录，Database Type = "postgresql"

**场景 3: 错误处理**
1. 创建连接：Host=invalid-host, Port=9999
2. 点击 "Test Connection"
   **预期**: 显示错误 "connection refused" 或 "no such host"

### 6.2 回归测试

**场景**: MySQL 功能不受影响
1. 创建/测试/编辑 MySQL 连接
   **预期**: 所有功能正常工作
2. 执行 MySQL Sysbench 压测
   **预期**: 压测正常执行，结果正确解析

---

## 7. 依赖关系

### 7.1 外部依赖

| 依赖 | 版本 | 用途 |
|------|------|------|
| github.com/lib/pq | latest | PostgreSQL 驱动 |
| PostgreSQL 服务器 | 9.6+ | 测试目标数据库 |
| Sysbench | 1.0+ | 压测工具 |

### 7.2 内部依赖

| 模块 | 依赖项 |
|------|--------|
| PostgreSQLConnection | connection.BaseConnection |
| SysbenchAdapter | connection.PostgreSQLConnection |
| ConnectionPage | connection_usecase.ConnectionUseCase |

---

## 8. 风险与限制

### 8.1 技术风险

| 风险 | 影响 | 缓解措施 |
|------|------|---------|
| PostgreSQL 驱动 API 不熟悉 | 中 | 参考官方文档和 MySQL 实现作为模板 |
| SSL 证书验证复杂性 | 中 | 先实现基本 SSL，证书验证作为 P1 |

### 8.2 功能限制

1. **本版本不实现**:
   - PostgreSQL 连接池配置
   - PostgreSQL 特有的性能指标（如 vacuum 统计）
   - PostgreSQL 高可用特性（主从、流复制）

2. **已知限制**:
   - SSL 证书验证需要额外配置（CA 证书路径）
   - SSH 隧道连接需要外部工具

---

## 9. 附录

### 9.1 PostgreSQL 连接字符串示例

```
# 基本连接
host=localhost port=5432 database=testdb user=postgres

# 带密码
host=localhost port=5432 database=testdb user=postgres password=secret

# 带 SSL
host=localhost port=5432 database=testdb user=postgres sslmode=require

# 完整示例
host=192.168.1.100 port=5432 database=production user=app_user password=*** sslmode=verify-full
```

### 9.2 Sysbench PostgreSQL 命令示例

```bash
# Prepare
sysbench /usr/share/sysbench/oltp_read_write.lua \
  --pgsql-host=localhost \
  --pgsql-port=5432 \
  --pgsql-user=postgres \
  --pgsql-password=secret \
  --pgsql-db=testdb \
  --tables=10 \
  --table-size=10000 \
  prepare

# Run
sysbench /usr/share/sysbench/oltp_read_write.lua \
  --pgsql-host=localhost \
  --pgsql-port=5432 \
  --pgsql-user=postgres \
  --pgsql-password=secret \
  --pgsql-db=testdb \
  --threads=8 \
  --time=60 \
  run

# Cleanup
sysbench /usr/share/sysbench/oltp_read_write.lua \
  --pgsql-host=localhost \
  --pgsql-port=5432 \
  --pgsql-user=postgres \
  --pgsql-password=secret \
  --pgsql-db=testdb \
  cleanup
```
