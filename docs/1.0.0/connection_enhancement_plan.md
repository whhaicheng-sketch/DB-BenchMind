# Connection Enhancement Plan (2026-03-10)

## 1. 当前实现现状分析

### 1.1 MySQL (已完成)
- 连接表单：完整支持
- 连接测试：正常工作
- SSH Tunnel： 支持
- 列表页显示： 正常

### 1.2 PostgreSQL (需完善)
- 连接表单： 支持
- 连接测试： 支持 TestConnectionDirect
- SSL Mode: 退验证
- **问题**: 列表项显示额外的"postgres" 标签（来自 database 字段）

### 1.3 SQL Server (需完善)
- 连接表单： 支持
- 连接测试： 支持 TestConnectionDirect
- Trust Server Certificate: 支持
- **问题**: WinRM 配置存在，SSH 不支持
- **问题**: 可能存在连接策略问题

### 1.4 Oracle (需完善)
- 连接表单： 支持
- 连接测试： 支持 TestConnectionDirect
- **问题**: 命令行可以连接(sqlplus)，但应用内测试可能存在问题

---

## 2. 需要修改的模块

### 2.1 前端模块

| 文件 | 修改内容 |
|-----|--------|
| `ConnectionsTab.vue` | SSH 标志位置调整，数据库名标签处理 |
| `ConnectionForm.vue` | 无需修改（已支持所有数据库类型） |

### 2.2 后端模块

| 文件 | 修改内容 |
|-----|--------|
| `connection.go` | 错误提示优化 |
| `oracle.go` | 连接测试验证和修复 |
| `sqlserver.go` | 连接测试验证和修复 |
| `postgresql.go` | 连接测试验证和修复 |

---

## 3. Connection 列表页 UI 调整方案

### 3.1 当前布局
```
[Name+Host] [SSH/SSL/DB tags] [Test] [Edit] [More] [Test result]
```

### 3.2 目标布局
```
[Name] [Host] [SSH] [DB]--- [Test] [Edit] [More] [Test result]
```

### 3.3 具体修改

**SSH 标志移到 conn-main 区域**:
```vue
<div class="conn-main">
  <div class="conn-name">{{ conn.name }}</div>
  <div class="conn-host">
    {{ conn.host }}:{{ conn.port }}
    <span v-if="conn.ssh_enabled" class="tag tag-ssh">SSH</span>
  </div>
</div>
```

**数据库名标签规则**:
- 仅当数据库名不为空且不是默认值时显示
- PostgreSQL 默认值: "postgres"
- MySQL 默认值: "" (空)
- Oracle 默认值: "orcl" 或Service Name
- SQL Server 默认值: "" (空)

---

## 4. 各数据库连接测试策略

### 4.1 MySQL
- 使用 go-sql-driver
- SSL Mode: false/preferred/skip-verify/true
- 已验证可用

### 4.2 PostgreSQL
- 使用 pgx 驱动
- SSL Mode: disable/prefer/require/verify-full
- 需要验证

### 4.3 Oracle
- 使用 goracle 驱动
- 连接字符串格式: system/password@host:port/service_name
- 鑗要验证当前实现

### 4.4 SQL Server
- 使用 go-mssqldb 驱动
- TrustServerCertificate: true (默认)
- 需要验证

---

## 5. 错误提示统一方案

### 5.1 当前问题
- 直接返回底层错误信息
- 用户难以理解

### 5.2 优化方案
```go
func formatConnectionError(err error, dbType string) string {
    errMsg := err.Error()

    // 常见错误模式匹配
    switch {
    case strings.Contains(errMsg, "connection refused"):
        return fmt.Sprintf("无法连接到数据库服务器，请检查主机地址和端口是否正确")
    case strings.Contains(errMsg, "authentication failed"):
        return fmt.Sprintf("认证失败，请检查用户名和密码")
    case strings.Contains(errMsg, "no such host"):
        return fmt.Sprintf("无法解析主机名，请检查地址是否正确")
    case strings.Contains(errMsg, "timeout"):
        return fmt.Sprintf("连接超时，请检查网络连接或防火墙设置")
    default:
        return fmt.Sprintf("连接失败: %s", errMsg)
    }
}
```

---

## 6. 鉉实施步骤

1. **Phase 1**: 列表页 UI 调整
   - 移动 SSH 标志位置
   - 处理数据库名标签显示逻辑

2. **Phase 2**: 后端错误提示优化
   - 添加错误格式化函数
   - 更新各数据库类型的 Test 方法

3. **Phase 3**: 饭验证证
   - 测试各数据库类型连接
   - 验证 UI 显示

---

## 7. 风险与回退方案

### 7.1 风险
- Oracle 连接可能需要额外配置
- SQL Server 认证方式可能有问题
- PostgreSQL SSL 配置可能不兼容

### 7.2 回退方案
- 保留原有错误信息作为调试日志
- 提供详细错误模式开关
- 如果出现问题，可以快速回滚到之前版本
