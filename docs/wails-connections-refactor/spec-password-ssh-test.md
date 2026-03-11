# Connections 页面密码/SSH/测试修复规格说明书 (v2)

## 1. 核心需求定义

### 需求1：Database Test 与 SSH Test 必须彻底分离

**目标行为**：
- 无论是否勾选 SSH，都必须存在独立的"Test Database"按钮
- 当勾选 SSH 后，额外出现独立的"Test SSH"按钮
- 这两个按钮是两个完全独立的测试动作，禁止复用同一逻辑

### 需求2：Database Test 行为定义

**目标行为**：
- 点击"Test Database"时，只测试数据库本身直连能力
- 禁止自动走 SSH 隧道
- 勾选 SSH 仅表示该连接保存了 SSH 配置，不影响 Database Test 行为

### 需求3：必须实现独立 SSH 测试

**目标行为**：
- 后端必须新增独立 SSH 测试接口 `TestSSHConnection`
- SSH 测试必须真实使用 host / port / username / password 建立 SSH 连接
- 如果 SSH 密码错误，测试必须失败

### 需求4：测试结果分别展示

**目标行为**：
- 点击 Test Database：只返回数据库测试结果
- 点击 Test SSH：只返回 SSH 测试结果
- 页面同时展示两种测试状态：
  - Database: Success / Failed
  - SSH: Success / Failed / Not tested

### 需求5：编辑时正确回显

**目标行为**：
- Password 字段编辑时必须正确回显已有值
- SSH 开关状态必须正确回显
- SSH Host / Port / Username / Password 必须正确回显

### 需求6：连接列表显示 SSH 状态

**目标行为**：
- 连接列表中明确显示该连接是否配置了 SSH
- 使用 Tag 或明确信息标识

### 需求7：UI 样式修复

**目标行为**：
- Database Name 输入框对齐正常
- Database Type 为深色背景

---

## 2. 前后端接口定义

### 2.1 后端接口

#### 新增：TestSSHConnection

```go
// SSHTestRequest represents a request to test SSH connection.
type SSHTestRequest struct {
    Host     string `json:"host"`
    Port     int    `json:"port"`
    Username string `json:"username"`
    Password string `json:"password"`
}

// ConnectionBinding 增加方法
func (b *ConnectionBinding) TestSSHConnection(req SSHTestRequest) ConnectionTestResult
```

#### 修改：TestConnectionDirect

- 无论 SSH 是否启用，都只测试数据库直连
- 不再自动走 SSH 隧道

### 2.2 前端交互

#### 按钮布局

| 场景 | 显示按钮 |
|------|----------|
| 未启用 SSH | Test Database |
| 启用 SSH | Test Database + Test SSH |

#### 状态展示

| 按钮 | 状态 |
|------|------|
| Test Database | Idle / Testing / Success / Failed |
| Test SSH | Not tested / Testing / Success / Failed |

---

## 3. 数据流设计

### 3.1 DTO 扩展（已完成）

```go
type SSHTunnelDTO struct {
    Enabled  bool   `json:"enabled"`
    Host     string `json:"host,omitempty"`
    Port     int    `json:"port"`
    Username string `json:"username,omitempty"`
}

type ConnectionDTO struct {
    // ... 现有字段
    SSH *SSHTunnelDTO `json:"ssh,omitempty"`
    // ... 其他字段
}
```

### 3.2 前端表单状态

```javascript
// 测试状态
testDatabaseStatus: 'idle' | 'testing' | 'success' | 'failed'
testDatabaseResult: { success, latency_ms, error }

testSSHStatus: 'idle' | 'testing' | 'success' | 'failed'
testSSHResult: { success, latency_ms, error }
```

---

## 4. 验收标准

### 4.1 密码回显

- [ ] Edit 打开时 password 字段可见
- [ ] 不修改密码直接保存，原密码不丢失
- [ ] 修改密码后保存，新密码生效
- [ ] 无 "(leave empty to keep current)" 提示

### 4.2 SSH 回显

- [ ] Edit 打开时 SSH 开关正确回显
- [ ] SSH Host/Port/Username/Password 正确回显
- [ ] 保存后再次进入 Edit，SSH 配置一致

### 4.3 测试分离

- [ ] 未启用 SSH：只显示 Test Database 按钮
- [ ] 启用 SSH：同时显示 Test Database 和 Test SSH 按钮
- [ ] Test Database 只测数据库直连
- [ ] Test SSH 真实测试 SSH 连接
- [ ] 错误 SSH 密码时 Test SSH 失败

### 4.4 连接列表

- [ ] 明确显示 SSH: Enabled / Disabled

### 4.5 UI

- [ ] Database Name 输入框对齐正常
- [ ] Database Type 为深色背景

---

## 5. 实现步骤

### 步骤1：后端新增 SSH 测试接口

文件：`internal/transportwails/bindings/connection.go`
- 添加 `SSHTestRequest` 结构体
- 添加 `TestSSHConnection` 方法

### 步骤2：修改前端测试区域

文件：`frontend/src/components/connection/ConnectionForm.vue`
- 添加独立 Test SSH 按钮
- 修改 Test Database 行为（不再自动走 SSH）
- 分别展示测试状态

### 步骤3：更新连接列表显示

文件：`frontend/src/components/pages/ConnectionsPage.vue`
- 添加 SSH 状态显示

### 步骤4：UI 样式修复

文件：`frontend/src/components/connection/ConnectionForm.vue`
- 修复 Database Name 对齐
- 修复 Database Type 背景颜色
