# DB-BenchMind SSH & WinRM 实现计划 (Plan)

**版本**: 1.0.0
**日期**: 2026-02-04
**状态**: SSH 已完成，WinRM 进行中

---

## 文档变更历史

| 版本 | 日期 | 作者 | 变更说明 |
|------|------|------|---------|
| 1.0.0 | 2026-02-04 | Claude | 初始版本：SSH + WinRM 实现计划 |

---

## 目录

1. [技术上下文总结](#1-技术上下文总结)
2. [合宪性审查](#2-合宪性审查)
3. [项目结构](#3-项目结构)
4. [核心数据结构](#4-核心数据结构)
5. [接口设计](#5-接口设计)
6. [技术决策记录](#6-技术决策记录)
7. [分阶段实施计划](#7-分阶段实施计划)
8. [测试策略](#8-测试策略)
9. [质量门禁](#9-质量门禁)

---

## 1. 技术上下文总结

### 1.1 项目定位

DB-BenchMind 连接隧道功能，为数据库压测提供安全的远程访问能力：
- **SSH Tunnel**: 通过 SSH 跳板机连接 MySQL、PostgreSQL、Oracle
- **WinRM**: 通过 WinRM 连接 Windows 宿主机（为 SQL Server 性能监控预留接口）

### 1.2 技术选型

| 技术领域 | 技术选型 | 版本 | 选型理由 |
|---------|---------|------|---------|
| **SSH 客户端** | golang.org/x/crypto/ssh | v0.47.0 | 标准库扩展，稳定可靠 |
| **WinRM 客户端** | github.com/masterzen/winrm | latest | 功能完整，社区活跃 |
| **GUI 框架** | Fyne | v2.7.2 | 跨平台、纯 Go |
| **存储** | SQLite | modernc.org/sqlite | 无 CGO、纯 Go |
| **密钥管理** | go-keyring | latest | 支持 gnome-keyring |

### 1.3 架构风格

**遵循 DDD + Clean Architecture**:
- **domain 层**: SSH Tunnel 和 WinRM 核心逻辑（无外部依赖）
- **usecase 层**: 连接管理业务逻辑
- **transport 层**: GUI 界面
- **infra 层**: 数据库持久化

---

## 2. 合宪性审查

### 2.1 Library-First Principle ✅

- SSH Tunnel 核心逻辑在 `internal/domain/connection/ssh_tunnel.go`
- WinRM 核心逻辑在 `internal/domain/connection/winrm.go`
- 可独立测试和复用

### 2.2 CLI Interface Mandate ⚠️

- 部分符合：GUI 应用，但核心库可通过 CLI 测试

### 2.3 Test-First Imperative ✅

- SSH: 已实现表格驱动测试
- WinRM: TDD 模式（先测试，后实现）

### 2.4 EARS Requirements Format ✅

- 所有需求使用 EARS 格式定义

### 2.5 Traceability Mandate ✅

- 需求 ID → 设计 → 任务 → 测试 → 实现
- 完整可追溯

### 2.6 Project Memory ✅

- 文档存储在 `specs/` 目录

### 2.7 Simplicity Gate ✅

- 单一功能：连接隧道
- 无过度设计

### 2.8 Anti-Abstraction Gate ✅

- 直接使用标准库 `golang.org/x/crypto/ssh`
- 无不必要的封装

### 2.9 Integration-First Testing ✅

- 真实 SSH/WinRM 连接测试

---

## 3. 项目结构

### 3.1 SSH Tunnel 相关文件

```
internal/
├── domain/connection/
│   ├── ssh_tunnel.go          # SSH Tunnel 核心逻辑 ✅ 已实现
│   ├── mysql.go               # MySQL SSH 支持 ✅ 已实现
│   ├── postgresql.go          # PostgreSQL SSH 支持 ✅ 已实现
│   └── oracle.go              # Oracle SSH 支持 ✅ 已实现
├── app/usecase/
│   └── connection_usecase.go  # 连接管理用例 ✅ 已实现
├── infra/database/
│   └── repository/
│       └── connection_repo.go # SSH 配置持久化 ✅ 已实现
└── transport/ui/pages/
    └── connection_page.go     # SSH UI ✅ 已实现
```

### 3.2 WinRM 相关文件（待实现）

```
internal/
├── domain/connection/
│   ├── winrm.go               # WinRM 核心逻辑 🚧 待实现
│   └── sqlserver.go           # SQL Server WinRM 支持 🚧 待实现
├── app/usecase/
│   └── connection_usecase.go  # 连接管理用例 🚧 待更新
├── infra/database/
│   └── repository/
│       └── connection_repo.go # WinRM 配置持久化 🚧 待更新
└── transport/ui/pages/
    └── connection_page.go     # WinRM UI 🚧 待实现
```

---

## 4. 核心数据结构

### 4.1 SSH Tunnel（已实现）

```go
// internal/domain/connection/ssh_tunnel.go

// SSHTunnelConfig SSH 隧道配置
type SSHTunnelConfig struct {
    Enabled  bool   `json:"enabled"`    // 是否启用 SSH Tunnel
    Host     string `json:"host"`       // SSH 服务器主机
    Port     int    `json:"port"`       // SSH 服务器端口（默认 22）
    Username string `json:"username"`   // SSH 用户名（默认 "root"）
    Password string `json:"-"`          // SSH 密码（存储到 keyring）
    KeyPath  string `json:"key_path"`   // SSH 私钥路径（预留）
    LocalPort int    `json:"local_port"` // 本地端口（0 = 自动分配）
}

// SSHTunnel SSH 隧道连接
type SSHTunnel struct {
    config    *SSHTunnelConfig
    client    *ssh.Client
    listener  net.Listener
    localPort int
    cancel    context.CancelFunc
    mu        sync.Mutex
    closed    bool
}

// NewSSHTunnel 创建 SSH 隧道
func NewSSHTunnel(ctx context.Context, config *SSHTunnelConfig, remoteHost string, remotePort int) (*SSHTunnel, error)

// Close 关闭 SSH 隧道
func (t *SSHTunnel) Close() error

// GetLocalPort 获取本地端口
func (t *SSHTunnel) GetLocalPort() int
```

### 4.2 WinRM（待实现）

```go
// internal/domain/connection/winrm.go

// WinRMConfig WinRM 配置
type WinRMConfig struct {
    Enabled  bool   `json:"enabled"`    // 是否启用 WinRM
    Host     string `json:"host"`       // WinRM 主机（使用 Database Host）
    Port     int    `json:"port"`       // WinRM 端口（5985 HTTP, 5986 HTTPS）
    Username string `json:"username"`   // 用户名（空 = 当前 Windows 用户）
    Password string `json:"-"`          // 密码（存储到 keyring）
    UseHTTPS bool   `json:"use_https"`  // 是否使用 HTTPS
}

// WinRMClient WinRM 客户端
type WinRMClient struct {
    config *WinRMConfig
    client *winrm.Client
}

// NewWinRMClient 创建 WinRM 客户端
func NewWinRMClient(ctx context.Context, config *WinRMConfig) (*WinRMClient, error)

// Test 测试 WinRM 连接
func (c *WinRMClient) Test(ctx context.Context) (*TestResult, error)

// Close 关闭 WinRM 连接
func (c *WinRMClient) Close() error
```

---

## 5. 接口设计

### 5.1 SSH Tunnel 接口（已实现）

```go
// Connection 接口扩展
type Connection interface {
    // ... 现有方法

    // GetSSHConfig 获取 SSH 配置（如支持）
    GetSSHConfig() *SSHTunnelConfig
    // SetSSHConfig 设置 SSH 配置
    SetSSHConfig(config *SSHTunnelConfig)
}
```

### 5.2 WinRM 接口（待实现）

```go
// Connection 接口扩展
type Connection interface {
    // ... 现有方法

    // GetWinRMConfig 获取 WinRM 配置（如支持）
    GetWinRMConfig() *WinRMConfig
    // SetWinRMConfig 设置 WinRM 配置
    SetWinRMConfig(config *WinRMConfig)
}
```

---

## 6. 技术决策记录

| ID | 决策 | 理由 | 替代方案 |
|----|------|------|---------|
| ADR-SSH-001 | 使用 golang.org/x/crypto/ssh | 标准库扩展，稳定可靠 | third-party SSH 库 |
| ADR-SSH-002 | 仅支持密码认证 | 简化实现，降低复杂度 | 密钥认证 |
| ADR-SSH-003 | Local Port 自动分配 | 避免端口冲突 | 用户指定端口 |
| ADR-WINRM-001 | 使用 masterzen/winrm | 功能完整，社区活跃 | 自己实现 WinRM 协议 |
| ADR-WINRM-002 | 当前阶段仅连接测试 | 分阶段实现，降低风险 | 一次性实现所有功能 |
| ADR-WINRM-003 | 性能监控预留接口 | 后续 tasks 中实现 | 当前阶段实现 |

---

## 7. 分阶段实施计划

### Phase 1: SSH Tunnel 实现 ✅ 已完成

**目标**: 实现 MySQL、PostgreSQL、Oracle 的 SSH Tunnel 支持

**交付物**:
- [x] SSH Tunnel 核心逻辑
- [x] SSH 配置持久化
- [x] SSH 密码存储到 keyring
- [x] SSH 连接测试
- [x] SSH UI 集成
- [x] 单元测试 + 集成测试

**验收标准**:
- [x] SSH Tunnel 连接正常工作
- [x] SSH 失败时自动测试直接数据库连接
- [x] Connections 列表显示 SSH 状态
- [x] Edit 连接时正确加载 SSH 配置

---

### Phase 2: WinRM 基础实现 🚧 当前阶段

**目标**: 实现 SQL Server 的 WinRM 连接配置和测试

**范围**:
- ✅ WinRM 配置 UI
- ✅ WinRM 连接测试
- ✅ WinRM 配置持久化
- ✅ WinRM 密码存储到 keyring
- ❌ 性能数据采集（后续阶段）

**交付物**:
- [ ] WinRM 核心逻辑
- [ ] WinRM 配置持久化
- [ ] WinRM 密码存储到 keyring
- [ ] WinRM 连接测试
- [ ] WinRM UI 集成
- [ ] 单元测试 + 集成测试

**验收标准**:
- [ ] WinRM 连接测试正常工作
- [ ] Connections 列表显示 WinRM 状态
- [ ] Edit 连接时正确加载 WinRM 配置

**详细任务**:

#### Task 2.1: WinRM 核心逻辑

**Type**: impl
**File**: `internal/domain/connection/winrm.go`

**Description**:
- 实现 `WinRMConfig` 结构体
- 实现 `WinRMClient` 结构体
- 实现 `NewWinRMClient` 函数
- 实现 `Test` 方法
- 实现 `Close` 方法

**Acceptance**:
- WinRM 配置结构体正确定义
- WinRM 客户端能够连接
- 测试连接成功返回正确结果
- 测试连接失败返回错误信息

#### Task 2.2: SQL Server WinRM 支持

**Type**: impl
**File**: `internal/domain/connection/sqlserver.go`

**Description**:
- 在 `SQLServerConnection` 中添加 `WinRM *WinRMConfig` 字段
- 实现 `GetWinRMConfig()` 方法
- 实现 `SetWinRMConfig()` 方法
- 更新 `Validate()` 方法验证 WinRM 配置

**Acceptance**:
- WinRM 字段正确添加
- Get/Set 方法正确实现
- 验证逻辑正确

#### Task 2.3: WinRM 配置持久化

**Type**: impl
**File**: `internal/infra/database/repository/connection_repo.go`

**Description**:
- 更新 `serializeConnection()` 序列化 WinRM 配置
- 更新 `deserializeConnection()` 反序列化 WinRM 配置
- 确保 WinRM 配置正确保存到数据库

**Acceptance**:
- WinRM 配置正确序列化
- WinRM 配置正确反序列化
- 数据库中能正确保存和加载

#### Task 2.4: WinRM 密码存储

**Type**: impl
**File**: `internal/app/usecase/connection_usecase.go`

**Description**:
- 添加 `getWinRMPassword()` 函数
- 添加 `setWinRMPassword()` 函数
- 更新 `CreateConnection()` 保存 WinRM 密码
- 更新 `UpdateConnection()` 更新 WinRM 密码
- 更新 `GetConnectionByID()` 加载 WinRM 密码
- 使用 key: `{conn_id}:winrm`

**Acceptance**:
- WinRM 密码正确保存到 keyring
- WinRM 密码正确从 keyring 加载
- 密码存储使用正确的 key

#### Task 2.5: WinRM 连接测试

**Type**: test + impl
**File**: `internal/domain/connection/winrm_test.go`, `internal/domain/connection/winrm.go`

**Description**:
- 编写 WinRM 连接测试（TDD）
- 实现连接测试逻辑
- 测试成功场景
- 测试失败场景

**Acceptance**:
- 测试覆盖成功场景
- 测试覆盖失败场景
- 测试覆盖超时场景

---

### Phase 3: WinRM UI 实现

**目标**: 实现 WinRM 配置界面

**详细任务**:

#### Task 3.1: WinRM UI 组件

**Type**: impl
**File**: `internal/transport/ui/pages/connection_page.go`

**Description**:
- 添加 WinRM 复选框 `winrmEnabledCheck`
- 添加 WinRM 端口输入框 `winrmPortEntry`
- 添加 HTTPS 复选框 `winrmHTTPSCheck`
- 添加用户名输入框 `winrmUserEntry`
- 添加密码输入框 `winrmPassEntry`
- 添加 Test WinRM 按钮 `btnTestWinRM`
- 实现显示/隐藏逻辑
- 实现 Use HTTPS 勾选时自动更新端口

**Acceptance**:
- WinRM UI 组件正确显示
- 复选框勾选时显示配置
- 取消勾选时隐藏配置
- Use HTTPS 勾选时端口自动更新为 5986
- 取消勾选时端口自动更新为 5985

#### Task 3.2: Test WinRM 按钮逻辑

**Type**: impl
**File**: `internal/transport/ui/pages/connection_page.go`

**Description**:
- 实现 `onTestWinRM()` 函数
- 测试 WinRM 连接
- 显示成功/失败对话框
- 显示详细错误信息

**Acceptance**:
- Test WinRM 按钮测试 WinRM 连接
- 成功时显示 "WinRM 连接成功"
- 失败时显示具体错误信息

#### Task 3.3: WinRM 状态显示

**Type**: impl
**File**: `internal/transport/ui/pages/connection_page.go`

**Description**:
- 在 Connections 列表显示 WinRM 状态图标（🖥️ WinRM）
- 更新连接列表项显示逻辑
- 更新 Test 按钮测试逻辑（先测试 WinRM，再测试数据库）

**Acceptance**:
- Connections 列表显示 WinRM 图标
- Test 按钮先测试 WinRM，再测试数据库
- 显示清晰的测试结果

#### Task 3.4: WinRM 配置加载

**Type**: impl
**File**: `internal/transport/ui/pages/connection_page.go`

**Description**:
- 在 `loadConnection()` 中加载 WinRM 配置
- 设置 WinRM 复选框状态
- 填充 WinRM 配置字段
- 从 keyring 加载 WinRM 密码

**Acceptance**:
- Edit 连接时正确加载 WinRM 配置
- WinRM 复选框状态正确
- WinRM 配置字段正确填充
- WinRM 密码正确加载

---

### Phase 4: 测试与文档

**目标**: 完善测试和文档

**详细任务**:

#### Task 4.1: 单元测试

**Type**: test
**File**: `internal/domain/connection/winrm_test.go`

**Description**:
- 测试 WinRMConfig 结构体
- 测试 WinRMClient 创建
- 测试连接成功场景
- 测试连接失败场景
- 测试超时场景

**Acceptance**:
- 测试覆盖率 > 80%
- 所有测试通过

#### Task 4.2: 集成测试

**Type**: test
**File**: `internal/infra/database/repository/connection_repo_test.go`

**Description**:
- 测试 WinRM 配置序列化
- 测试 WinRM 配置反序列化
- 测试 WinRM 密码存储
- 测试 WinRM 密码加载

**Acceptance**:
- 集成测试通过
- 真实 SQLite 数据库测试

#### Task 4.3: 用户文档

**Type**: impl
**File**: `docs/WINRM_GUIDE.md`

**Description**:
- 编写 WinRM 使用指南
- 包含配置说明
- 包含测试步骤
- 包含故障排除

**Acceptance**:
- 文档完整清晰
- 包含所有必要信息

---

## 8. 测试策略

### 8.1 测试金字塔

```
                    /\
                   /  \
                  / E2E \         5% - 端到端测试
                 /--------\
                /          \
               / Integration\    25% - 集成测试
              /--------------\
             /                \
            /    Unit Tests     \  70% - 单元测试
           /--------------------\
```

### 8.2 测试覆盖率要求

| 层级 | 目标覆盖率 | 必须覆盖 |
|------|-----------|---------|
| domain/connection/ssh_tunnel.go | > 90% | 所有 SSH 逻辑 |
| domain/connection/winrm.go | > 90% | 所有 WinRM 逻辑 |
| usecase/connection_usecase.go | > 85% | SSH/WinRM 密码管理 |
| infra/database/repository/ | > 80% | 配置持久化 |
| transport/ui/pages/ | > 40% | UI 逻辑（手动为主） |

---

## 9. 质量门禁

### 9.1 代码质量标准

所有 PR 必须通过：

1. **格式检查**
   ```bash
   gofmt -l . | wc -l  # 必须为 0
   ```

2. **静态检查**
   ```bash
   go vet ./...
   golangci-lint run  # 零错误
   ```

3. **测试覆盖**
   ```bash
   go test -cover ./...
   # 覆盖率 > 80%
   ```

4. **竞态检测**
   ```bash
   go test -race ./...
   # 零竞态
   ```

5. **安全扫描**
   ```bash
   govulncheck ./...
   # 零已知漏洞
   ```

---

## 10. 风险与缓解

| 风险 | 影响 | 概率 | 缓解措施 |
|------|------|------|---------|
| WinRM 连接不稳定 | 高 | 中 | 增加重试机制、超时控制 |
| WinRM 库不兼容 | 中 | 低 | 选择成熟的库（masterzen/winrm） |
| Windows 认证复杂 | 中 | 中 | 支持集成 Windows 认证和用户名密码 |
| UI 状态管理复杂 | 中 | 低 | 清晰的状态管理逻辑 |

---

## 11. 验收标准

### 11.1 SSH 验收标准 ✅

- [x] 支持 MySQL、PostgreSQL、Oracle 的 SSH Tunnel
- [x] SSH 配置字段正确显示
- [x] SSH 密码安全存储到 keyring
- [x] SSH 连接测试正常工作
- [x] SSH 失败时自动测试直接数据库连接
- [x] Connections 列表显示 SSH 状态图标
- [x] Edit 连接时正确加载 SSH 配置
- [x] 所有单元测试通过
- [x] golangci-lint 零错误

### 11.2 WinRM 验收标准（当前阶段）

- [ ] SQL Server 连接支持 WinRM 配置
- [ ] WinRM 配置字段正确显示
- [ ] WinRM 密码安全存储到 keyring
- [ ] WinRM 连接测试正常工作
- [ ] Connections 列表显示 WinRM 状态图标
- [ ] Edit 连接时正确加载 WinRM 配置
- [ ] 所有单元测试通过（覆盖率 > 80%）
- [ ] golangci-lint 零错误

---

**文档结束**
