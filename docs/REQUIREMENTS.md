# 需求文档 (Requirements)

本文档记录 DB-BenchMind 的所有功能需求和变更历史。

---

## 最新需求变更

### REQ-CONN-013: MySQL Database 字段可选

**日期**: 2026-01-28
**状态**: ✅ 已实现
**优先级**: P2 (正常)

#### 需求描述

在连接管理页面的 Add Connection 功能中，当 DatabaseType 为 MySQL 时，Database 字段应该可以为空。

#### 业务理由

1. **灵活性**：用户可以只连接到 MySQL 服务器，而不必指定具体的数据库
2. **数据库管理**：DBA 可能需要先连接到服务器，然后创建或管理数据库
3. **权限配置**：某些用户可能只有服务器级权限，没有特定数据库权限

#### 技术实现

**修改文件**: `internal/domain/connection/mysql.go`

1. **Validate 方法**:
   - 移除对 `database` 字段的必填验证
   - 添加注释说明 database 是可选的

2. **GetDSN 方法**:
   - 处理 `database` 为空的情况
   - 空时返回: `username@tcp(host:port)`
   - 有值时返回: `username@tcp(host:port)/database`

3. **GetDSNWithPassword 方法**:
   - 处理 `database` 为空的情况
   - 空时返回: `username:password@tcp(host:port)`
   - 有值时返回: `username:password@tcp(host:port)/database`

4. **Redact 方法**:
   - 处理 `database` 为空的显示
   - 空时显示: `name (***@host:port)`
   - 有值时显示: `name (***@host:port/database)`

#### 使用示例

```go
// 无数据库的连接（只连接服务器）
conn := &connection.MySQLConnection{
    BaseConnection: connection.BaseConnection{
        ID:   "conn-1",
        Name: "MySQL Server",
    },
    Host:     "localhost",
    Port:     3306,
    Database: "", // 可选
    Username: "root",
    Password: "password",
}
// DSN: root:password@tcp(localhost:3306)

// 有数据库的连接
conn := &connection.MySQLConnection{
    BaseConnection: connection.BaseConnection{
        ID:   "conn-2",
        Name: "MySQL Test DB",
    },
    Host:     "localhost",
    Port:     3306,
    Database: "testdb", // 可选
    Username: "user",
    Password: "password",
}
// DSN: user:password@tcp(localhost:3306)/testdb
```

#### 测试场景

1. ✅ **添加连接 - 无数据库**
   - Database Type: MySQL
   - Database: (留空)
   - 其他字段正常填写
   - 预期: 连接成功保存，可测试连接

2. ✅ **添加连接 - 有数据库**
   - Database Type: MySQL
   - Database: testdb
   - 其他字段正常填写
   - 预期: 连接成功保存，可测试连接

3. ✅ **测试连接 - 无数据库**
   - 使用无数据库的连接配置
   - 预期: 能成功连接到 MySQL 服务器

4. ✅ **DSN 生成正确**
   - 验证 GetDSN() 在 database 为空时不包含 "/"
   - 验证 GetDSNWithPassword() 在 database 为空时不包含 "/"

#### 相关需求

- **REQ-CONN-002**: 支持多种数据库类型连接
- **REQ-CONN-003**: 连接测试功能
- **REQ-CONN-010**: 连接参数验证

#### 影响范围

- **GUI**: 连接管理页面，Database 字段标记为可选
- **CLI**: 连接添加命令，Database 参数可选
- **验证逻辑**: MySQL 连接验证不再强制要求 database
- **文档**: 更新用户手册和 API 文档

---

## 原有需求 (来自 spec.md)

### REQ-CONN-001: 连接列表

系统应能够列出所有已保存的数据库连接。

### REQ-CONN-002: 多数据库支持

系统应支持多种数据库类型：
- MySQL
- PostgreSQL
- Oracle
- SQL Server

### REQ-CONN-003: 连接测试

系统应能够测试数据库连接是否可用，返回：
- 连接状态（成功/失败）
- 连接延迟
- 数据库版本信息

### REQ-CONN-004: 连接验证

系统应验证连接参数的有效性：
- 主机地址格式
- 端口范围（1-65535）
- 必填字段检查

### REQ-CONN-005: 密码安全

系统应安全存储密码，不在配置文件中明文存储。

### REQ-CONN-006: 连接编辑

系统应支持编辑现有连接配置。

### REQ-CONN-007: 连接删除

系统应支持删除不需要的连接。

### REQ-CONN-008: 连接名称唯一性

连接名称必须在所有连接中唯一。

### REQ-CONN-009: 连接导出导入

系统应支持导出和导入连接配置（不包含密码）。

### REQ-CONN-010: SSL 配置

系统应支持配置 SSL 连接选项：
- disabled
- preferred
- required

### REQ-CONN-011: 连接超时

连接测试应支持超时配置（默认 10 秒）。

### REQ-CONN-012: 连接池

系统应使用连接池管理数据库连接。

---

## Bug 修复

### BUG-001: MySQL 驱动缺失

**日期**: 2026-01-28
**状态**: ✅ 已修复
**严重性**: P1 (高)

#### 问题描述

点击 Test Connection 按钮时报错：
```
Error: failed to open connection: sql: unknown driver "mysql" (forgotten import?)
```

#### 根本原因

MySQL 驱动导入被注释掉：
```go
// _ "github.com/go-sql-driver/mysql" // 被注释了
```

#### 修复方案

1. 添加 MySQL 驱动依赖：
   ```bash
   go get github.com/go-sql-driver/mysql
   ```

2. 取消注释驱动导入：
   ```go
   _ "github.com/go-sql-driver/mysql" // MySQL driver
   ```

#### 测试验证

- ✅ MySQL 连接测试不再报 driver 错误
- ✅ 能正常连接到 MySQL 服务器
- ✅ 能正确获取数据库版本信息

---

## 需求变更历史

| 日期 | 需求ID | 描述 | 状态 |
|------|--------|------|------|
| 2026-01-28 | REQ-CONN-013 | MySQL Database 字段可选 | ✅ 已实现 |
| 2026-01-28 | BUG-001 | MySQL 驱动缺失 | ✅ 已修复 |

---

## 参考文档

- [产品需求文档 (spec.md)](../specs/spec.md)
- [技术实现计划 (plan.md)](../specs/plan.md)
- [API 参考文档](./API_REFERENCE.md)
- [用户手册](./USER_GUIDE.md)
