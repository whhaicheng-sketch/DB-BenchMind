# Connections 页面密码/SSH/测试修复实施计划 (v2)

## 1. 代码定位结果

### 1.1 关键文件

| 文件 | 作用 |
|------|------|
| `internal/transportwails/bindings/connection.go` | Wails binding，DTO 定义，接口实现 |
| `frontend/src/components/connection/ConnectionForm.vue` | 表单组件，测试逻辑 |
| `frontend/src/components/pages/ConnectionsPage.vue` | 连接列表，显示 SSH 状态 |
| `frontend/src/stores/connection.js` | Pinia store，API 调用封装 |

---

## 2. 实现方案

### 2.1 后端：新增 TestSSHConnection 接口

```go
// SSHTestRequest represents a request to test SSH connection.
type SSHTestRequest struct {
    Host     string `json:"host"`
    Port     int    `json:"port"`
    Username string `json:"username"`
    Password string `json:"password"`
}

// TestSSHConnection tests SSH connection.
func (b *ConnectionBinding) TestSSHConnection(req SSHTestRequest) ConnectionTestResult
```

实现逻辑：
1. 验证必填字段（host, port, username）
2. 建立 SSH 连接
3. 返回测试结果

### 2.2 前端：测试按钮分离

#### 按钮1：Test Database
- 始终显示
- 调用 `TestConnectionDirect`
- 不受 SSH 配置影响

#### 按钮2：Test SSH
- 仅在 `ssh_enabled` 为 true 时显示
- 调用 `TestSSHConnection`
- 显示独立状态

### 2.3 前端：状态分别展示

```javascript
// 两个独立状态
testDatabaseStatus: 'idle' | 'testing' | 'success' | 'failed'
testSSHStatus: 'idle' | 'testing' | 'success' | 'failed'
```

### 2.4 连接列表显示

在 ConnectionsPage 中添加 SSH 状态显示：
- Tag 显示：SSH
- 或明确文字：SSH: Enabled

---

## 3. 实现步骤

### 步骤1：后端添加 SSH 测试接口

1. 添加 `SSHTestRequest` 结构体
2. 添加 `TestSSHConnection` 方法
3. 验证构建

### 步骤2：更新前端 store

1. 添加 `testSSHConnection` action
2. 添加对应 binding

### 步骤3：修改 ConnectionForm 测试区域

1. 移除 "via SSH" 相关文案
2. 添加独立 Test SSH 按钮
3. 修改测试结果展示
4. 修复 UI 样式

### 步骤4：更新连接列表

1. 修改 `getConnectionTags` 添加 SSH 标签
2. 验证显示

### 步骤5：验证

1. 前端构建
2. 后端构建
3. 手动测试场景

---

## 4. 回归验证策略

1. **新建流程**：创建新连接（含 SSH），验证保存成功
2. **编辑流程**：编辑已有连接，验证密码和 SSH 正确
3. **测试流程**：
   - 未启用 SSH：只显示 Test Database
   - 启用 SSH：显示两个按钮
   - Test Database 不走 SSH
   - Test SSH 真实测试
4. **连接列表**：显示 SSH 状态
