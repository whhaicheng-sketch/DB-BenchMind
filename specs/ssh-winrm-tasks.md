# DB-BenchMind SSH & WinRM 任务分解列表

**版本**: 1.0.0
**日期**: 2026-02-04
**状态**: SSH 已完成，WinRM 进行中

---

## 文档说明

本文档将 SSH & WinRM 实现计划分解为**原子化、可执行的任务列表**，确保每个任务都可以被 AI 独立完成。

### 任务格式说明

```markdown
### Phase X: [阶段名称]

#### Task X.Y: [任务标题] [P]
**Type**: test/impl
**File**: 文件路径
**Depends**: X.A, X.B（依赖的任务ID）
**Description**: 详细描述
**Acceptance**: 验收标准
**Implementation**: 实现要点
```

- **[P]** 标记表示该任务可与其他标记 `[P]` 的任务并行执行
- **Type**: `test` 表示测试任务，`impl` 表示实现任务
- **TDD 强制**: 测试任务必须在实现任务之前（Red → Green → Refactor）

---

## Phase 1: SSH Tunnel 实现 ✅ 已完成

**目标**: 实现 MySQL、PostgreSQL、Oracle 的 SSH Tunnel 支持

**完成日期**: 2026-02-03

---

### Task 1.1: [测试] SSH Tunnel 核心逻辑 ✅

**Type**: test
**File**: `internal/domain/connection/ssh_tunnel_test.go`
**Depends**: 无
**Status**: ✅ 完成

**Description**: 表格驱动测试 SSH Tunnel 核心功能

**Acceptance**:
- 测试 SSH 隧道创建成功
- 测试 SSH 隧道连接失败场景
- 测试本地端口自动分配
- 测试 SSH 隧道关闭

---

### Task 1.2: SSH Tunnel 核心逻辑 ✅

**Type**: impl
**File**: `internal/domain/connection/ssh_tunnel.go`
**Depends**: 1.1
**Status**: ✅ 完成

**Description**: 实现 SSH Tunnel 核心逻辑

**Implementation**:
- 实现 `SSHTunnelConfig` 结构体
- 实现 `SSHTunnel` 结构体
- 实现 `NewSSHTunnel()` 函数
- 实现 `Close()` 方法
- 实现 `GetLocalPort()` 方法
- 实现 `startForwarding()` 方法
- 实现 `forwardConnection()` 方法

---

### Task 1.3: MySQL SSH 支持 ✅

**Type**: impl
**File**: `internal/domain/connection/mysql.go`
**Depends**: 1.2
**Status**: ✅ 完成

**Description**: 在 MySQLConnection 中添加 SSH 支持

**Implementation**:
- 添加 `SSH *SSHTunnelConfig` 字段
- 更新 `Test()` 方法，支持通过 SSH 隧道连接

---

### Task 1.4: PostgreSQL SSH 支持 ✅

**Type**: impl
**File**: `internal/domain/connection/postgresql.go`
**Depends**: 1.2
**Status**: ✅ 完成

**Description**: 在 PostgreSQLConnection 中添加 SSH 支持

**Implementation**:
- 添加 `SSH *SSHTunnelConfig` 字段
- 更新 `Test()` 方法，支持通过 SSH 隧道连接

---

### Task 1.5: Oracle SSH 支持 ✅

**Type**: impl
**File**: `internal/domain/connection/oracle.go`
**Depends**: 1.2
**Status**: ✅ 完成

**Description**: 在 OracleConnection 中添加 SSH 支持

**Implementation**:
- 添加 `SSH *SSHTunnelConfig` 字段
- 更新 `Test()` 方法，支持通过 SSH 隧道连接

---

### Task 1.6: SSH 配置持久化 ✅

**Type**: impl
**File**: `internal/infra/database/repository/connection_repo.go`
**Depends**: 1.3, 1.4, 1.5
**Status**: ✅ 完成

**Description**: 实现 SSH 配置序列化和反序列化

**Implementation**:
- 更新 `serializeConnection()` 序列化 SSH 配置
- 更新 `deserializeConnection()` 反序列化 SSH 配置
- 为 MySQL、PostgreSQL、Oracle 添加 SSH 序列化

---

### Task 1.7: SSH 密码存储 ✅

**Type**: impl
**File**: `internal/app/usecase/connection_usecase.go`
**Depends**: 1.6
**Status**: ✅ 完成

**Description**: 实现 SSH 密码存储到 keyring

**Implementation**:
- 添加 `getSSHPassword()` 函数
- 添加 `setSSHPassword()` 函数
- 更新 `CreateConnection()` 保存 SSH 密码（key: `{conn_id}:ssh`）
- 更新 `UpdateConnection()` 更新 SSH 密码
- 更新 `GetConnectionByID()` 加载 SSH 密码

---

### Task 1.8: SSH UI 组件 ✅

**Type**: impl
**File**: `internal/transport/ui/pages/connection_page.go`
**Depends**: 1.7
**Status**: ✅ 完成

**Description**: 实现 SSH 配置 UI

**Implementation**:
- 添加 SSH 复选框 `sshEnabledCheck`
- 添加 SSH 端口输入框 `sshPortEntry`（默认 22）
- 添加 SSH 用户名输入框 `sshUserEntry`（默认 root）
- 添加 SSH 密码输入框 `sshPassEntry`（密码掩码）
- 添加 Test SSH 按钮 `btnTestSSH`
- 实现显示/隐藏逻辑
- 不显示 SSH Host 字段（使用 Database Host）
- 不显示 Local Port 字段（自动分配）

---

### Task 1.9: Test SSH 按钮逻辑 ✅

**Type**: impl
**File**: `internal/transport/ui/pages/connection_page.go`
**Depends**: 1.8
**Status**: ✅ 完成

**Description**: 实现 Test SSH 按钮测试逻辑

**Implementation**:
- 实现 `onTestSSH()` 函数
- 测试 SSH 隧道连接
- 显示成功/失败对话框

---

### Task 1.10: Test Database 按钮逻辑 ✅

**Type**: impl
**File**: `internal/transport/ui/pages/connection_page.go`
**Depends**: 1.8
**Status**: ✅ 完成

**Description**: 实现 Test Database 按钮测试逻辑（不使用 SSH）

**Implementation**:
- 更新 `onTestInDialog()` 函数
- 创建不包含 SSH 配置的连接对象
- 仅测试直接数据库连接

---

### Task 1.11: Connections 列表 Test 按钮逻辑 ✅

**Type**: impl
**File**: `internal/transport/ui/pages/connection_page.go`
**Depends**: 1.9, 1.10
**Status**: ✅ 完成

**Description**: 实现 Connections 列表 Test 按钮逻辑

**Implementation**:
- 更新 `onTestConnection()` 函数
- 先加载连接（包含密码）
- 先测试 SSH（如启用）
- SSH 成功 → 测试数据库（通过 SSH）
- SSH 失败 → 测试数据库（直接连接）
- 显示综合测试结果

---

### Task 1.12: SSH 状态显示 ✅

**Type**: impl
**File**: `internal/transport/ui/pages/connection_page.go`
**Depends**: 1.11
**Status**: ✅ 完成

**Description**: 在 Connections 列表显示 SSH 状态

**Implementation**:
- 更新连接列表项显示逻辑
- 显示 SSH 状态图标（🔒 SSH）
- 更新 `refreshConnectionList()` 函数

---

### Task 1.13: SSH 配置加载 ✅

**Type**: impl
**File**: `internal/transport/ui/pages/connection_page.go`
**Depends**: 1.8
**Status**: ✅ 完成

**Description**: 实现 Edit 连接时加载 SSH 配置

**Implementation**:
- 在 `loadConnection()` 中加载 SSH 配置
- 设置 SSH 复选框状态
- 填充 SSH 配置字段
- 从 keyring 加载 SSH 密码

---

### Phase 1 验收标准 ✅

- [x] SSH Tunnel 连接正常工作
- [x] SSH 失败时自动测试直接数据库连接
- [x] Connections 列表显示 SSH 状态图标
- [x] Edit 连接时正确加载 SSH 配置
- [x] 所有单元测试通过
- [x] golangci-lint 零错误

---

## Phase 2: WinRM 核心逻辑 🚧 进行中

**目标**: 实现 WinRM 核心逻辑和 SQL Server 支持

---

### Task 2.1: [测试] WinRM 配置结构

**Type**: test
**File**: `internal/domain/connection/winrm_test.go`
**Depends**: 无
**Status**: 🚧 待实现

**Description**: 测试 WinRMConfig 结构体

**Acceptance**:
- 测试 WinRMConfig 字段正确性
- 测试默认值
- 测试验证逻辑

**Content**:
```go
package connection

import (
    "testing"
)

func TestWinRMConfig_Validate(t *testing.T) {
    tests := []struct {
        name    string
        config  *WinRMConfig
        wantErr bool
        errMsg  string
    }{
        {
            name: "valid WinRM config (HTTP)",
            config: &WinRMConfig{
                Enabled:  true,
                Host:     "192.168.1.100",
                Port:     5985,
                Username: "",
                Password: "",
                UseHTTPS: false,
            },
            wantErr: false,
        },
        {
            name: "valid WinRM config (HTTPS)",
            config: &WinRMConfig{
                Enabled:  true,
                Host:     "192.168.1.100",
                Port:     5986,
                Username: "administrator",
                Password: "password",
                UseHTTPS: true,
            },
            wantErr: false,
        },
        {
            name: "invalid port - too low",
            config: &WinRMConfig{
                Enabled: true,
                Host:    "192.168.1.100",
                Port:    0,
                UseHTTPS: false,
            },
            wantErr: true,
            errMsg:  "port must be between 1 and 65535",
        },
        {
            name: "invalid port - too high",
            config: &WinRMConfig{
                Enabled: true,
                Host:    "192.168.1.100",
                Port:    99999,
                UseHTTPS: false,
            },
            wantErr: true,
            errMsg:  "port must be between 1 and 65535",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.config.Validate()
            if (err != nil) != tt.wantErr {
                t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if tt.wantErr && tt.errMsg != "" && err != nil {
                if !containsString(err.Error(), tt.errMsg) {
                    t.Errorf("Validate() error = %v, want contain %v", err.Error(), tt.errMsg)
                }
            }
        })
    }
}
```

---

### Task 2.2: WinRM 核心逻辑

**Type**: impl
**File**: `internal/domain/connection/winrm.go`
**Depends**: 2.1
**Status**: 🚧 待实现

**Description**: 实现 WinRM 核心逻辑

**Implementation**:
```go
// Package connection provides WinRM functionality for Windows Server connections.
package connection

import (
    "context"
    "fmt"
    "log/slog"
    "time"

    "github.com/masterzen/winrm"
)

// WinRMConfig represents WinRM configuration.
// Implements: REQ-WINRM-001 ~ REQ-WINRM-015
type WinRMConfig struct {
    Enabled  bool   `json:"enabled"`    // Whether WinRM is enabled
    Host     string `json:"host"`       // WinRM host (use Database Host)
    Port     int    `json:"port"`       // WinRM port (5985 HTTP, 5986 HTTPS)
    Username string `json:"username"`   // Username (empty = current Windows user)
    Password string `json:"-"`          // Password (stored in keyring)
    UseHTTPS bool   `json:"use_https"`  // Whether to use HTTPS
}

// Validate validates the WinRM configuration.
func (c *WinRMConfig) Validate() error {
    if !c.Enabled {
        return nil
    }

    if c.Host == "" {
        return fmt.Errorf("host is required")
    }

    if c.Port < 1 || c.Port > 65535 {
        return fmt.Errorf("port must be between 1 and 65535, got %d", c.Port)
    }

    // Validate standard ports
    if c.UseHTTPS && c.Port != 5986 {
        return fmt.Errorf("HTTPS requires port 5986, got %d", c.Port)
    }
    if !c.UseHTTPS && c.Port != 5985 {
        return fmt.Errorf("HTTP requires port 5985, got %d", c.Port)
    }

    return nil
}

// WinRMClient manages a WinRM connection.
type WinRMClient struct {
    config *WinRMConfig
    client *winrm.Client
}

// NewWinRMClient creates a new WinRM client.
// Returns an error if the client cannot be created.
func NewWinRMClient(ctx context.Context, config *WinRMConfig) (*WinRMClient, error) {
    if !config.Enabled {
        return nil, fmt.Errorf("WinRM is not enabled")
    }

    slog.Info("WinRM: Creating client",
        "op", "winrm_create",
        "host", config.Host,
        "port", config.Port,
        "https", config.UseHTTPS,
        "username", config.Username)

    // Validate configuration
    if err := config.Validate(); err != nil {
        return nil, fmt.Errorf("invalid WinRM configuration: %w", err)
    }

    // Create WinRM client
    endpoint := winrm.NewEndpoint(
        config.Host,
        config.Port,
        config.UseHTTPS,
        config.Username == "", // Empty username = current user
        config.Username,
        config.Password,
    )

    client, err := winrm.NewClientWithParameters(endpoint, nil, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to create WinRM client: %w", err)
    }

    slog.Info("WinRM: Client created successfully",
        "op", "winrm_created",
        "host", config.Host,
        "port", config.Port)

    return &WinRMClient{
        config: config,
        client: client,
    }, nil
}

// Test tests the WinRM connection.
// Returns TestResult containing success/failure, latency, error.
func (c *WinRMClient) Test(ctx context.Context) (*TestResult, error) {
    start := time.Now()

    // Simple WinRM test: execute "hostname" command
    shell, err := c.client.CreateShell()
    if err != nil {
        latency := time.Since(start).Milliseconds()
        return &TestResult{
            Success:   false,
            LatencyMs: latency,
            Error:     fmt.Sprintf("failed to create shell: %v", err),
        }, nil
    }
    defer shell.Close()

    // Execute hostname command
    ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
    defer cancel()

    _, err = shell.ExecuteWithContext(ctx, "hostname")
    latency := time.Since(start).Milliseconds()

    if err != nil {
        return &TestResult{
            Success:   false,
            LatencyMs: latency,
            Error:     fmt.Sprintf("WinRM command failed: %v", err),
        }, nil
    }

    slog.Info("WinRM: Connection test successful",
        "op", "winrm_test_success",
        "latency_ms", latency)

    return &TestResult{
        Success:         true,
        LatencyMs:       latency,
        DatabaseVersion: "WinRM Connected",
    }, nil
}

// Close closes the WinRM client.
func (c *WinRMClient) Close() error {
    // WinRM client doesn't have explicit close method
    // Resources are cleaned up automatically
    slog.Info("WinRM: Client closed",
        "op", "winrm_close",
        "host", c.config.Host)
    return nil
}
```

**Acceptance**:
- WinRMConfig 结构体正确定义
- Validate 方法正确实现
- WinRMClient 创建逻辑正确
- Test 方法正确实现
- Close 方法正确实现

---

### Task 2.3: SQL Server WinRM 支持

**Type**: impl
**File**: `internal/domain/connection/sqlserver.go`
**Depends**: 2.2
**Status**: 🚧 待实现

**Description**: 在 SQLServerConnection 中添加 WinRM 支持

**Implementation**:

在 `SQLServerConnection` 结构体中添加：
```go
type SQLServerConnection struct {
    BaseConnection

    // ... existing fields ...

    // WinRM configuration (for Windows Server monitoring)
    WinRM *connection.WinRMConfig `json:"winrm,omitempty"` // Added
}
```

添加方法：
```go
// GetWinRMConfig returns the WinRM configuration.
func (c *SQLServerConnection) GetWinRMConfig() *connection.WinRMConfig {
    return c.WinRM
}

// SetWinRMConfig sets the WinRM configuration.
func (c *SQLServerConnection) SetWinRMConfig(config *connection.WinRMConfig) {
    c.WinRM = config
    c.UpdatedAt = time.Now()
}
```

**Acceptance**:
- WinRM 字段正确添加
- GetWinRMConfig 方法正确实现
- SetWinRMConfig 方法正确实现

---

### Task 2.4: WinRM 配置持久化

**Type**: impl
**File**: `internal/infra/database/repository/connection_repo.go`
**Depends**: 2.3
**Status**: 🚧 待实现

**Description**: 实现 WinRM 配置序列化和反序列化

**Implementation**:

在 `serializeConnection()` 中添加：
```go
// WinRM configuration for SQL Server
case *connection.SQLServerConnection:
    // ... existing fields ...
    if c.WinRM != nil {
        data["winrm"] = map[string]interface{}{
            "enabled":   c.WinRM.Enabled,
            "host":      c.WinRM.Host,
            "port":      c.WinRM.Port,
            "username":  c.WinRM.Username,
            "use_https": c.WinRM.UseHTTPS,
        }
        slog.Info("Repository: Serializing SQL Server connection with WinRM",
            "conn_id", conn.GetID(),
            "winrm_enabled", c.WinRM.Enabled)
    }
```

在 `deserializeConnection()` 中添加：
```go
case connection.DatabaseTypeSQLServer:
    conn := &connection.SQLServerConnection{
        // ... existing fields ...
    }

    // Load WinRM configuration if present
    if winrmData, ok := data["winrm"].(map[string]interface{}); ok {
        conn.WinRM = &connection.WinRMConfig{
            Enabled:   getBool(winrmData, "enabled"),
            Host:      getString(winrmData, "host"),
            Port:      getInt(winrmData, "port"),
            Username:  getString(winrmData, "username"),
            UseHTTPS:  getBool(winrmData, "use_https"),
        }
        slog.Info("Repository: Deserialized SQL Server connection with WinRM",
            "conn_id", id,
            "winrm_enabled", conn.WinRM.Enabled)
    }
```

**Acceptance**:
- WinRM 配置正确序列化
- WinRM 配置正确反序列化
- 添加日志记录

---

### Task 2.5: WinRM 密码存储

**Type**: impl
**File**: `internal/app/usecase/connection_usecase.go`
**Depends**: 2.4
**Status**: 🚧 待实现

**Description**: 实现 WinRM 密码存储到 keyring

**Implementation**:

添加辅助函数：
```go
// getWinRMPassword gets WinRM password from a connection.
func getWinRMPassword(conn connection.Connection) string {
    switch c := conn.(type) {
    case *connection.SQLServerConnection:
        if c.WinRM != nil {
            return c.WinRM.Password
        }
    }
    return ""
}

// setWinRMPassword sets WinRM password on a connection.
func setWinRMPassword(conn connection.Connection, password string) {
    switch c := conn.(type) {
    case *connection.SQLServerConnection:
        if c.WinRM != nil {
            c.WinRM.Password = password
        }
    }
}
```

更新 `CreateConnection()`：
```go
// Save WinRM password to keyring if provided
if winrmPwd := getWinRMPassword(conn); winrmPwd != "" {
    winrmKey := conn.GetID() + ":winrm"
    if err := uc.keyring.Set(ctx, winrmKey, winrmPwd); err != nil {
        // Rollback
        _ = uc.keyring.Delete(ctx, conn.GetID())
        _ = uc.keyring.Delete(ctx, conn.GetID()+":ssh")
        return fmt.Errorf("save WinRM password to keyring: %w", err)
    }
}
```

更新 `UpdateConnection()`：
```go
// Update WinRM password in keyring if changed
if winrmPwd := getWinRMPassword(conn); winrmPwd != "" {
    winrmKey := conn.GetID() + ":winrm"
    if err := uc.keyring.Set(ctx, winrmKey, winrmPwd); err != nil {
        return fmt.Errorf("update WinRM password in keyring: %w", err)
    }
}
```

更新 `GetConnectionByID()`：
```go
// Load WinRM password from keyring and set on connection
winrmKey := id + ":winrm"
winrmPassword, err := uc.keyring.Get(ctx, winrmKey)
if err != nil {
    if !keyring.IsNotFound(err) {
        return nil, fmt.Errorf("get WinRM password from keyring: %w", err)
    }
    // WinRM password not in keyring, continue without it
} else {
    setWinRMPassword(conn, winrmPassword)
}
```

更新 `DeleteConnection()`：
```go
// Remove WinRM password from keyring (best effort, ignore if not found)
_ = uc.keyring.Delete(ctx, id+":winrm")
```

**Acceptance**:
- WinRM 密码正确保存到 keyring
- WinRM 密码正确从 keyring 加载
- 使用正确的 key: `{conn_id}:winrm`
- 删除连接时同时删除 WinRM 密码

---

### Task 2.6: [测试] WinRM 连接测试

**Type**: test
**File**: `internal/domain/connection/winrm_test.go`
**Depends**: 2.2
**Status**: 🚧 待实现

**Description**: 测试 WinRM 连接

**Acceptance**:
- 测试连接成功场景（需要真实 WinRM 环境）
- 测试连接失败场景
- 测试超时场景
- 测试认证失败场景

---

### Task 2.7: WinRM 连接测试

**Type**: impl
**File**: `internal/domain/connection/winrm.go`
**Depends**: 2.6
**Status**: 🚧 待实现

**Description**: 实现 WinRM 连接测试逻辑

**Acceptance**:
- Test 方法正确实现
- 成功时返回正确结果
- 失败时返回错误信息
- 超时正确处理

---

## Phase 3: WinRM UI 实现

**目标**: 实现 WinRM 配置界面

---

### Task 3.1: WinRM UI 组件

**Type**: impl
**File**: `internal/transport/ui/pages/connection_page.go`
**Depends**: 2.7
**Status**: 🚧 待实现

**Description**: 实现 WinRM 配置 UI

**Implementation**:

在 `connectionDialog` 结构体中添加：
```go
type connectionDialog struct {
    // ... existing fields ...

    // WinRM components (only for SQL Server)
    winrmEnabledCheck  *widget.Check
    winrmPortEntry     *widget.Entry
    winrmHTTPSCheck    *widget.Check
    winrmUserEntry     *widget.Entry
    winrmPassEntry     *widget.Entry
    btnTestWinRM       *widget.Button
    winrmContainer     *fyne.Container
}
```

在 `buildForm()` 中添加：
```go
// WinRM Configuration (only for SQL Server)
if dbType == "SQL Server" {
    d.winrmEnabledCheck = widget.NewCheck("Enable WinRM", func(checked bool) {
        if checked {
            d.winrmContainer.Show()
        } else {
            d.winrmContainer.Hide()
        }
    })

    d.winrmPortEntry = widget.NewEntry()
    d.winrmPortEntry.SetText("5985")

    d.winrmHTTPSCheck = widget.NewCheck("Use HTTPS", func(checked bool) {
        if checked {
            d.winrmPortEntry.SetText("5986")
        } else {
            d.winrmPortEntry.SetText("5985")
        }
    })

    d.winrmUserEntry = widget.NewEntry()
    d.winrmUserEntry.SetPlaceHolder("(Empty = current Windows user)")

    d.winrmPassEntry = widget.NewEntry()
    d.winrmPassEntry.Password = true

    d.btnTestWinRM = widget.NewButton("Test WinRM", d.onTestWinRM)

    // Build WinRM form
    winrmHeader := container.NewVBox(
        widget.NewLabel("WinRM Configuration:"),
    )

    winrmForm := container.NewVBox(
        container.NewGrid(nil,
            container.NewGridItem(widget.NewLabel("WinRM Port:"), 2, 0, 1.0),
            container.NewGridItem(d.winrmPortEntry, 3, 0, 2.0),
        ),
        widget.NewSeparator(),
        container.NewGrid(nil,
            container.NewGridItem(d.winrmHTTPSCheck, 2, 0, 1.0),
        ),
        widget.NewSeparator(),
        container.NewGrid(nil,
            container.NewGridItem(widget.NewLabel("Username:"), 2, 0, 1.0),
            container.NewGridItem(d.winrmUserEntry, 3, 0, 2.0),
        ),
        container.NewGrid(nil,
            container.NewGridItem(widget.NewLabel("Password:"), 2, 0, 1.0),
            container.NewGridItem(d.winrmPassEntry, 3, 0, 2.0),
        ),
    )

    d.winrmContainer = container.NewVBox(winrmHeader, winrmForm)
    d.winrmContainer.Hide() // Initially hidden
}
```

**Acceptance**:
- WinRM UI 组件正确显示
- 复选框勾选时显示配置
- 取消勾选时隐藏配置
- Use HTTPS 勾选时端口自动更新为 5986
- 取消勾选时端口自动更新为 5985

---

### Task 3.2: Test WinRM 按钮逻辑

**Type**: impl
**File**: `internal/transport/ui/pages/connection_page.go`
**Depends**: 3.1
**Status**: 🚧 待实现

**Description**: 实现 Test WinRM 按钮测试逻辑

**Implementation**:
```go
func (d *connectionDialog) onTestWinRM() {
    // Validate WinRM configuration
    if !d.winrmEnabledCheck.Checked {
        dlg := dialog.NewInformation("WinRM Not Enabled",
            "Please enable WinRM first", d.window)
        dlg.SetConfirmText("OK")
        dlg.Show()
        return
    }

    // Show testing dialog
    progressDialog := dialog.NewCustom("Testing WinRM", "Cancel", nil, d.window)
    statusLabel := widget.NewLabel("Testing WinRM connection...")
    progressDialog.Resize(fyne.NewSize(300, 100))
    progressDialog.SetContent(statusLabel)
    progressDialog.Show()

    go func() {
        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()

        // Create WinRM config
        port, _ := strconv.Atoi(d.winrmPortEntry.Text)
        winrmConfig := &connection.WinRMConfig{
            Enabled:  true,
            Host:     d.hostEntry.Text,
            Port:     port,
            Username: d.winrmUserEntry.Text,
            Password: d.winrmPassEntry.Text,
            UseHTTPS: d.winrmHTTPSCheck.Checked,
        }

        // Test WinRM connection
        client, err := connection.NewWinRMClient(ctx, winrmConfig)
        if err != nil {
            progressDialog.Hide()
            showError(d.window, "WinRM Connection Failed", err.Error())
            return
        }
        defer client.Close()

        result, err := client.Test(ctx)
        progressDialog.Hide()

        if err != nil {
            showError(d.window, "WinRM Test Failed", err.Error())
            return
        }

        if result.Success {
            dialog.NewInformation("WinRM Connection Successful",
                fmt.Sprintf("Latency: %dms\nVersion: %s", result.LatencyMs, result.DatabaseVersion),
                d.window).Show()
        } else {
            showError(d.window, "WinRM Connection Failed", result.Error)
        }
    }()
}
```

**Acceptance**:
- Test WinRM 按钮测试 WinRM 连接
- 成功时显示 "WinRM 连接成功"
- 失败时显示具体错误信息

---

### Task 3.3: WinRM 状态显示

**Type**: impl
**File**: `internal/transport/ui/pages/connection_page.go`
**Depends**: 3.2
**Status**: 🚧 待实现

**Description**: 在 Connections 列表显示 WinRM 状态

**Implementation**:

更新 `refreshConnectionList()`：
```go
// Check WinRM status
var winrmEnabled bool
switch c := conn.(type) {
case *connection.SQLServerConnection:
    winrmEnabled = c.WinRM != nil && c.WinRM.Enabled
}

winrmIndicator := ""
if winrmEnabled {
    winrmIndicator = " | 🖥️ WinRM"
}

infoText := fmt.Sprintf("%s %s  |  %s@%s:%s%s",
    dbIcon, connName, username, host, portStr, winrmIndicator)
```

更新 `onTestConnection()` 添加 WinRM 测试：
```go
// Test WinRM if configured
if winrmConfig != nil && winrmConfig.Enabled {
    client, err := connection.NewWinRMClient(ctx, winrmConfig)
    if err != nil {
        winrmError = err
        winrmSuccess = false
    } else {
        result, err := client.Test(ctx)
        if err != nil || !result.Success {
            winrmError = fmt.Errorf("WinRM test failed: %w", err)
            winrmSuccess = false
        } else {
            client.Close()
            winrmSuccess = true
        }
    }
}

// Build comprehensive message
var message string
if winrmConfig != nil && winrmConfig.Enabled {
    if winrmSuccess {
        message = fmt.Sprintf("✅ WinRM: Connected (%dms)\n✅ Database: Connected via WinRM (%dms)\nVersion: %s",
            winrmLatency, dbLatency, dbVersion)
    } else {
        message = fmt.Sprintf("❌ WinRM: Failed (%dms) - %v\n✅ Database: Connected (Direct, without WinRM) (%dms)\nVersion: %s\n⚠️ WinRM was not used",
            winrmLatency, winrmError, dbLatency, dbVersion)
    }
} else {
    message = fmt.Sprintf("✅ Database: Connected (%dms)\nVersion: %s",
        dbLatency, dbVersion)
}
```

**Acceptance**:
- Connections 列表显示 WinRM 图标（🖥️ WinRM）
- Test 按钮先测试 WinRM（如启用），再测试数据库
- 显示清晰的测试结果

---

### Task 3.4: WinRM 配置加载

**Type**: impl
**File**: `internal/transport/ui/pages/connection_page.go`
**Depends**: 3.1
**Status**: 🚧 待实现

**Description**: 实现 Edit 连接时加载 WinRM 配置

**Implementation**:

更新 `loadConnection()`：
```go
// Load WinRM configuration for SQL Server
if dbType == "SQL Server" {
    sqlServerConn, ok := conn.(*connection.SQLServerConnection)
    if ok && sqlServerConn.WinRM != nil {
        loadedWinRMConfig := sqlServerConn.WinRM

        // Set WinRM enabled checkbox
        d.winrmEnabledCheck.SetChecked(loadedWinRMConfig.Enabled)
        if loadedWinRMConfig.Enabled {
            d.winrmContainer.Show()
        }

        // Set WinRM port
        if loadedWinRMConfig.Port > 0 {
            d.winrmPortEntry.SetText(fmt.Sprintf("%d", loadedWinRMConfig.Port))
        }

        // Set Use HTTPS
        d.winrmHTTPSCheck.SetChecked(loadedWinRMConfig.UseHTTPS)

        // Set username
        d.winrmUserEntry.SetText(loadedWinRMConfig.Username)

        // Try to load WinRM password from keyring for edit mode
        if d.isEditMode && d.conn != nil {
            ctx := context.Background()
            winrmKey := d.conn.GetID() + ":winrm"
            winrmPassword, err := d.connUC.GetKeyring().Get(ctx, winrmKey)
            if err == nil && winrmPassword != "" {
                d.winrmPassEntry.SetText(winrmPassword)
            }
        }
    }
}
```

**Acceptance**:
- Edit 连接时正确加载 WinRM 配置
- WinRM 复选框状态正确
- WinRM 配置字段正确填充
- WinRM 密码正确加载

---

## Phase 4: 测试与文档

**目标**: 完善测试和文档

---

### Task 4.1: 单元测试完善

**Type**: test
**File**: `internal/domain/connection/winrm_test.go`
**Depends**: 2.7
**Status**: 🚧 待实现

**Description**: 完善 WinRM 单元测试

**Acceptance**:
- 测试覆盖率 > 80%
- 所有测试通过

---

### Task 4.2: 集成测试

**Type**: test
**File**: `internal/infra/database/repository/connection_repo_test.go`
**Depends**: 2.4
**Status**: 🚧 待实现

**Description**: WinRM 配置持久化集成测试

**Acceptance**:
- 测试 WinRM 配置序列化
- 测试 WinRM 配置反序列化
- 测试 WinRM 密码存储
- 测试 WinRM 密码加载

---

### Task 4.3: 用户文档

**Type**: impl
**File**: `docs/WINRM_GUIDE.md`
**Depends**: 3.4
**Status**: 🚧 待实现

**Description**: 编写 WinRM 使用指南

**Content**:
```markdown
# WinRM 连接指南

## 概述

DB-BenchMind 支持 SQL Server 通过 WinRM 连接到 Windows 宿主机。

## 配置步骤

1. 新建或编辑 SQL Server 连接
2. 勾选 "Enable WinRM"
3. 配置 WinRM 参数：
   - WinRM Port: 默认 5985（HTTP）或 5986（HTTPS）
   - Use HTTPS: 勾选时自动更新端口为 5986
   - Username: 留空 = 当前 Windows 用户，或输入指定用户名
   - Password: 输入密码（如使用指定用户名）
4. 点击 "Test WinRM" 测试连接
5. 点击 "Save" 保存配置

## 测试 WinRM

点击 "Test WinRM" 按钮测试 WinRM 连接：
- 成功：显示 "WinRM 连接成功" 和延迟
- 失败：显示具体错误信息

## 故障排除

### WinRM 连接失败

1. 检查 WinRM 服务是否已启动
2. 检查防火墙是否允许 5985/5986 端口
3. 检查用户名和密码是否正确
4. 检查 HTTPS 证书是否有效

### WinRM 未启用

在 Windows Server 上启用 WinRM：
```powershell
# 启用 WinRM HTTP
Enable-PSRemoting -Force

# 或启用 WinRM HTTPS
# 需要配置证书
```
```

**Acceptance**:
- 文档完整清晰
- 包含所有必要信息

---

## 附录

### A. 并行任务索引

Phase 2 可并行执行的任务：
- Task 2.1: [测试] WinRM 配置结构
- Task 2.2: WinRM 核心逻辑（依赖 2.1）

### B. 关键里程碑

- **M1**: SSH Tunnel 完成（2026-02-03）✅
- **M2**: WinRM 核心逻辑完成（待定）
- **M3**: WinRM UI 完成（待定）
- **M4**: WinRM 全部完成（待定）

### C. TDD 检查清单

每个功能点必须遵循：
1. ✅ 先写测试（Task Type: test）
2. ✅ 确认测试失败
3. ✅ 编写实现（Task Type: impl）
4. ✅ 确认测试通过
5. ✅ 重构优化

---

**文档结束**
