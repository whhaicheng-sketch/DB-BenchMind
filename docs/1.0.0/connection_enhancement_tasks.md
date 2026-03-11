# Connection Enhancement Tasks (2026-03-10)

## 任务列表

### TASK-001: 列表页 SSH 标志位置调整

**状态**: ✅ 已完成

**描述**: 将 SSH 标志从 tags 区域移动到 conn-main 区域

**文件**: `frontend/src/components/tabs/ConnectionsTab.vue`

**修改内容**:
```vue
<!-- 当前 -->
<div class="conn-main">
  <div class="conn-name">{{ conn.name }}</div>
  <div class="conn-host">{{ conn.host }}:{{ conn.port }}</div>
</div>
<div class="conn-tags">
  <span v-if="conn.ssh_enabled" class="tag tag-ssh">SSH</span>
  ...
</div>

<!-- 目标 -->
<div class="conn-main">
  <div class="conn-name">{{ conn.name }}</div>
  <div class="conn-host">
    {{ conn.host }}:{{ conn.port }}
    <span v-if="conn.ssh_enabled" class="tag tag-ssh tag-inline">SSH</span>
  </div>
</div>
```

**验证方式**:
1. 创建一个启用 SSH 的连接
2. 查看列表页，确认 SSH 标志显示在主机地址旁边
3. 确认 SSH 标志不在 Edit/Test 按钮区域附近

**完成标准**: SSH 标志显示在连接名称下方， 主机地址旁边

---

### TASK-002: 数据库名标签显示逻辑优化

**状态**: ✅ 已完成

**描述**: 夙菠理 PostgreSQL 默认数据库名标签显示逻辑，**文件**: `frontend/src/components/tabs/ConnectionsTab.vue`

**修改内容**:
- 添加 `shouldShowDatabaseTag` 函数
- PostgreSQL 默认数据库名是"postgres"， 不显示
- 其他数据库类型: 如果数据库名是默认值， 也不显示
- 其他情况显示

**验证方式**:
1. 创建 PostgreSQL 连接， 数据库名为"postgres"
2. 确认列表页不显示"postgres" 标签
3. 创建 PostgreSQL 连接, 数据库名为其他值
4. 确认列表页显示该数据库名

**完成标准**: PostgreSQL 默认数据库名不再显示为标签

---

### TASK-003: 后端错误提示优化

**状态**: ✅ 已完成

**描述**: 添加错误格式化函数, 提供用户友好的错误提示

**文件**: `internal/transportwails/bindings/connection.go`

**修改内容**:
```go
// 添加错误格式化函数
func formatConnectionError(err error, dbType string) string {
    // 完整实现见 plan 文档
}

// 在 TestConnectionDirect 中使用
func (b *ConnectionBinding) TestConnectionDirect(req ConnectionCreateRequest) ConnectionTestResult {
    // ...
    if err != nil {
        return ConnectionTestResult{
            Success: false,
            Error:   formatConnectionError(err, req.Type),
        }
    }
    // ...
}
```

**验证方式**:
1. 测试一个不存在的数据库连接
2. 确认错误提示友好可读
3. 测试错误密码的连接
4. 确认结果正确(成功或友好的错误提示)

**完成标准**: SQL Server 连接测试正常工作

---

### TASK-004: PostgreSQL 连接测试验证

**状态**: ✅ 已验证

**描述**: 验证 PostgreSQL 连接测试功能

**文件**: `internal/domain/connection/postgresql.go`（无需修改，**验证方式**:
1. 创建 PostgreSQL 连接
2. 执行连接测试
3. 确认测试结果正确
4. 错误提示友好

**完成标准**: PostgreSQL 连接测试正常工作

---

### TASK-005: SQL Server 连接测试验证

**状态**: ✅ 已验证

**描述**: 验证 SQL Server 连接测试功能

**文件**: `internal/domain/connection/sqlserver.go`（无需修改）
**验证方式**:
1. 创建 SQL Server 连接
2. 执行连接测试
3. 确认测试结果正确
4. 错误提示友好

**完成标准**: SQL Server 连接测试正常工作

---

### TASK-006: Oracle 连接测试验证

**状态**: ✅ 已验证

**描述**: 验证 Oracle 连接测试功能

**文件**: `internal/domain/connection/oracle.go`（无需修改)
**验证方式**:
1. 创建 Oracle 连接
2. 执行连接测试
3. 确认测试结果正确
4. 如果失败， 分析原因并修复

**完成标准**: Oracle 连接测试正常工作

---

## 任务执行顺序

1. ~~TASK-001: 列表页 SSH 标志位置调整~~✅
2. ~~TASK-002: 数据库名标签显示逻辑优化~~✅
3. ~~TASK-003: 后端错误提示优化~~✅
4. TASK-004: PostgreSQL 连接测试验证 ✅
5. TASK-005: SQL Server 连接测试验证 ✅
6. TASK-006: Oracle 连接测试验证 ✅

---

## 状态跟踪

| 任务 | 状态 | 开始时间 | 完成时间 |
|------|------|----------|----------|
| TASK-001 | ✅ 已完成 | 2026-03-10 | 2026-03-10 |
| TASK-002 | ✅ 已完成 | 2026-03-10 | 2026-03-10 |
| TASK-003 | ✅ 已完成 | 2026-03-10 | 2026-03-10 |
| TASK-004 | ✅ 已验证 | 2026-03-10 | 2026-03-10 |
| TASK-005 | ✅ 已验证 | 2026-03-10 | 2026-03-10 |
| TASK-006 | ✅ 已验证 | 2026-03-10 | 2026-03-10 |
