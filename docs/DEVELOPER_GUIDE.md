# DB-BenchMind 开发者指南

**版本**: 1.0.0
**更新日期**: 2026-01-28

---

## 目录

1. [开发环境设置](#开发环境设置)
2. [项目结构](#项目结构)
3. [架构原则](#架构原则)
4. [开发工作流](#开发工作流)
5. [测试策略](#测试策略)
6. [代码规范](#代码规范)
7. [调试技巧](#调试技巧)
8. [贡献指南](#贡献指南)

---

## 开发环境设置

### 前置要求

- **Go**: 1.22.2 或更高版本
- **Git**: 用于版本控制
- **Make**: 用于构建自动化（可选）
- **golangci-lint**: 用于代码检查

### 安装 Go 工具链

#### Linux (Ubuntu/Debian)

```bash
# 安装 Go
wget https://go.dev/dl/go1.22.2.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.22.2.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# 验证安装
go version
```

#### macOS

```bash
# 使用 Homebrew
brew install go@1.22

# 验证安装
go version
```

### 克隆仓库

```bash
git clone https://github.com/whhaicheng/DB-BenchMind.git
cd DB-BenchMind
```

### 安装依赖

```bash
# 下载 Go 模块依赖
go mod download

# 验证依赖
go mod verify
```

### 安装开发工具

```bash
# 安装 golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# 安装 goimports
go install golang.org/x/tools/cmd/goimports@latest

# 安装 govulncheck
go install golang.org/x/vuln/cmd/govulncheck@latest
```

### 配置环境变量（可选）

```bash
# ~/.bashrc 或 ~/.zshrc
export PATH=$PATH:$(go env GOPATH)/bin
export GO111MODULE=on
export GOPROXY=https://goproxy.cn,direct  # 中国用户
```

---

## 项目结构

### 目录布局

```
DB-BenchMind/
├── cmd/                          # 应用入口
│   ├── db-benchmind/            # GUI 主程序
│   └── db-benchmind-cli/        # CLI 工具
├── internal/                     # 私有代码
│   ├── app/                     # 应用层
│   │   └── usecase/             # 用例编排
│   ├── domain/                  # 领域层
│   │   ├── connection/          # 连接模型
│   │   ├── template/            # 模板模型
│   │   ├── execution/           # 执行模型
│   │   ├── comparison/          # 对比模型
│   │   └── report/              # 报告模型
│   ├── infra/                   # 基础设施层
│   │   ├── adapter/             # 工具适配器
│   │   ├── database/            # 数据库
│   │   │   └── repository/      # 仓储实现
│   │   ├── keyring/             # 密钥管理
│   │   ├── report/              # 报告生成器
│   │   └── tool/                # 工具检测
│   └── transport/               # 传输层
│       └── ui/                  # GUI 界面
├── pkg/                         # 公共库
│   └── benchmark/               # 基准测试接口
├── contracts/                   # 契约定义
│   ├── templates/               # 内置模板 JSON
│   └── schemas/                 # JSON Schema
├── test/                        # 测试
│   ├── integration/             # 集成测试
│   └── testdata/                # 测试数据
├── docs/                        # 文档
├── .specify/                    # 项目治理
│   └── steering/                # 架构决策
├── specs/                       # 规格文档
├── configs/                     # 配置样例
├── scripts/                     # 脚本
├── go.mod                       # Go 模块定义
├── go.sum                       # 依赖锁定
├── Makefile                     # 构建自动化
├── .golangci.yml                # Linter 配置
└── README.md                    # 项目说明
```

### 依赖规则

```
domain/      ← 无外部依赖（仅标准库）
app/usecase/ ← 依赖 domain/, pkg/
infra/       ← 依赖 domain/, pkg/（实现 app/usecase/ 定义的接口）
transport/   ← 依赖 app/usecase/, domain/, pkg/
```

**禁止依赖**:
- ❌ `domain/` → `infra/`
- ❌ `app/usecase/` → `infra/`（通过接口）
- ❌ `transport/` → `infra/`

---

## 架构原则

### Clean Architecture + DDD

项目采用分层架构，每层有明确的职责：

#### 1. Domain Layer (`internal/domain/`)

**职责**: 核心业务逻辑，无外部依赖

**示例**:
```go
// internal/domain/connection/mysql.go
package connection

type MySQLConnection struct {
    ID       string
    Name     string
    Host     string
    Port     int
    Database string
    Username string
    Password string // json:"-"
}

func (c *MySQLConnection) Validate() error {
    // 纯业务逻辑，无外部依赖
    if c.Name == "" {
        return errors.New("name is required")
    }
    // ...
}
```

**约束**:
- ✅ 可以使用标准库
- ❌ 禁止依赖外部包
- ❌ 禁止依赖 infra/, app/, transport/

#### 2. Application Layer (`internal/app/usecase/`)

**职责**: 用例编排，定义接口

**示例**:
```go
// internal/app/usecase/connection_usecase.go
package usecase

// 定义接口（依赖倒置）
type ConnectionRepository interface {
    Save(ctx context.Context, conn connection.Connection) error
    FindByID(ctx context.Context, id string) (connection.Connection, error)
    // ...
}

type ConnectionUseCase struct {
    repo    ConnectionRepository  // 接口，不是具体实现
    keyring KeyringProvider
}

func (uc *ConnectionUseCase) CreateConnection(ctx context.Context, conn connection.Connection) error {
    // 业务编排
    if err := conn.Validate(); err != nil {
        return fmt.Errorf("validation failed: %w", err)
    }

    exists, err := uc.repo.ExistsByName(ctx, conn.GetName())
    if err != nil {
        return fmt.Errorf("check exists: %w", err)
    }

    if exists {
        return ErrConnectionAlreadyExists
    }

    return uc.repo.Save(ctx, conn)
}
```

**约束**:
- ✅ 可以依赖 domain/, pkg/
- ✅ 定义接口供 infra/ 实现
- ❌ 禁止依赖 infra/, transport/

#### 3. Infrastructure Layer (`internal/infra/`)

**职责**: 外部依赖实现

**示例**:
```go
// internal/infra/database/repository/connection_repo.go
package repository

// 实现 usecase 定义的接口
type SQLiteConnectionRepository struct {
    db      *sql.DB
    keyring KeyringProvider
}

func (r *SQLiteConnectionRepository) Save(ctx context.Context, conn connection.Connection) error {
    // 具体实现
    query := `INSERT INTO connections (id, name, type, config_json, created_at, updated_at)
              VALUES (?, ?, ?, ?, ?, ?)`

    configJSON, err := conn.ToJSON()
    if err != nil {
        return fmt.Errorf("serialize connection: %w", err)
    }

    _, err = r.db.ExecContext(ctx, query,
        conn.GetID(),
        conn.GetName(),
        string(conn.GetType()),
        string(configJSON),
        time.Now(),
        time.Now(),
    )

    return err
}
```

**约束**:
- ✅ 可以依赖 domain/, pkg/
- ✅ 实现 usecase/ 定义的接口
- ✅ 可以使用外部库
- ❌ 禁止依赖 app/, transport/

#### 4. Transport Layer (`internal/transport/ui/`)

**职责**: GUI 界面，仅 I/O

**示例**:
```go
// internal/transport/ui/connection_page.go
package ui

type ConnectionPage struct {
    connUC *usecase.ConnectionUseCase  // 依赖用例，不是仓储
}

func (p *ConnectionPage) OnAddConnection() {
    // 仅 I/O，业务逻辑委托给 use case
    conn := p.getConnectionFromUI()
    err := p.connUC.CreateConnection(ctx, conn)
    p.showError(err)
}
```

**约束**:
- ✅ 可以依赖 app/usecase/, domain/, pkg/
- ✅ 可以使用 GUI 框架
- ❌ 禁止包含业务逻辑
- ❌ 禁止依赖 infra/

---

## 开发工作流

### 1. 分支策略

```bash
# 主分支
main/master          # 稳定版本

# 功能分支
feat/feature-name    # 新功能
fix/bug-name         # Bug 修复
refactor/scope       # 重构
```

### 2. 开发流程

#### Step 1: 创建功能分支

```bash
git checkout -b feat/add-hammerdb-support
```

#### Step 2: 编写代码

遵循以下原则：
- **TDD**: 先写测试，再写实现
- **小步提交**: 频繁提交，每次提交一个逻辑单元
- **原子提交**: 每次提交应该能独立运行

#### Step 3: 运行检查

```bash
# 格式化代码
goimports -w .
gofmt -w .

# 运行测试
go test ./...

# 运行 linter
golangci-lint run

# 安全扫描
govulncheck ./...
```

#### Step 4: 提交代码

```bash
git add .
git commit -m "feat(adapter): add HammerDB adapter

- Implement HammerDBAdapter for MySQL, Oracle, SQL Server, PostgreSQL
- Add TCL script generation for tpcc and tpch benchmarks
- Parse NOPM/TPM metrics from HammerDB output
- Add 10 unit tests with 100% coverage

Closes #123"
```

**Commit Message 规范** (Conventional Commits):

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Type**:
- `feat`: 新功能
- `fix`: Bug 修复
- `refactor`: 重构
- `perf`: 性能优化
- `test`: 测试
- `docs`: 文档
- `chore`: 构建/工具
- `ci`: CI 配置

#### Step 5: 推送和 PR

```bash
git push origin feat/add-hammerdb-support
# 创建 Pull Request
```

---

## 测试策略

### 测试金字塔

```
        /\
       /  \
      / E2E \        少量端到端测试
     /--------\
    /  集成测试  \     适量集成测试
   /--------------\
  /    单元测试      \   大量单元测试
 /--------------------\
```

### 1. 单元测试

**位置**: 与源码同目录，`*_test.go`

**原则**:
- 快速运行（毫秒级）
- 无外部依赖
- 表格驱动测试

**示例**:

```go
// internal/domain/connection/mysql_test.go
package connection

import "testing"

func TestMySQLConnection_Validate(t *testing.T) {
    tests := []struct {
        name    string
        conn    *MySQLConnection
        wantErr bool
        errMsg  string
    }{
        {
            name: "valid connection",
            conn: &MySQLConnection{
                Name:     "test",
                Host:     "localhost",
                Port:     3306,
                Database: "testdb",
                Username: "root",
            },
            wantErr: false,
        },
        {
            name: "invalid port - negative",
            conn: &MySQLConnection{
                Name: "test",
                Host: "localhost",
                Port: -1,
                Database: "testdb",
                Username: "root",
            },
            wantErr: true,
            errMsg:  "port must be between 1 and 65535",
        },
        {
            name: "missing name",
            conn: &MySQLConnection{
                Name: "",
                Host: "localhost",
                Port: 3306,
                Database: "testdb",
                Username: "root",
            },
            wantErr: true,
            errMsg:  "name is required",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.conn.Validate()
            if (err != nil) != tt.wantErr {
                t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if tt.wantErr && tt.errMsg != "" && err != nil {
                if !strings.Contains(err.Error(), tt.errMsg) {
                    t.Errorf("Validate() error = %v, want contain %v", err.Error(), tt.errMsg)
                }
            }
        })
    }
}
```

**运行单元测试**:

```bash
# 运行所有单元测试
go test ./...

# 运行特定包
go test ./internal/domain/connection

# 详细输出
go test -v ./...

# 覆盖率
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 2. 集成测试

**位置**: `test/integration/`

**原则**:
- 测试组件间交互
- 使用真实依赖（数据库）
- 可以较慢

**示例**:

```go
// test/integration/connection_test.go
package integration

import (
    "context"
    "testing"
    "github.com/whhaicheng/DB-BenchMind/internal/domain/connection"
    "github.com/whhaicheng/DB-BenchMind/internal/infra/database"
    "github.com/whhaicheng/DB-BenchMind/internal/infra/database/repository"
    "github.com/whhaicheng/DB-BenchMind/internal/infra/keyring"
    "github.com/whhaicheng/DB-BenchMind/internal/app/usecase"
)

func TestIntegration_ConnectionWorkflow(t *testing.T) {
    // 使用内存数据库
    db, err := database.InitializeSQLite(context.Background(), ":memory:")
    require.NoError(t, err)
    defer db.Close()

    // 初始化
    connRepo := repository.NewSQLiteConnectionRepository(db)
    keyringProvider, err := keyring.NewFileFallback("", "test")
    require.NoError(t, err)
    connUC := usecase.NewConnectionUseCase(connRepo, keyringProvider)

    // 创建连接
    conn := &connection.MySQLConnection{
        ID:       "test-1",
        Name:     "Test Connection",
        Host:     "localhost",
        Port:     3306,
        Database: "testdb",
        Username: "root",
        Password: "password",
    }

    err = connUC.CreateConnection(context.Background(), conn)
    require.NoError(t, err)

    // 查询连接
    retrieved, err := connUC.GetConnection(context.Background(), conn.GetID())
    require.NoError(t, err)
    assert.Equal(t, conn.GetName(), retrieved.GetName())

    // 删除连接
    err = connUC.DeleteConnection(context.Background(), conn.GetID())
    require.NoError(t, err)

    // 验证删除
    _, err = connUC.GetConnection(context.Background(), conn.GetID())
    assert.Error(t, err)
}
```

**运行集成测试**:

```bash
# 运行所有测试（包括集成测试）
go test ./... -tags=integration

# 仅运行集成测试
go test ./test/integration/...
```

### 3. E2E 测试

**位置**: `test/e2e/`

**原则**:
- 测试完整用户场景
- 可以使用真实工具
- 最慢，最少

---

## 代码规范

### Go 代码规范

#### 1. 包命名

- **全小写**: `connection`, `usecase`
- **单数**: `repository` (不是 `repositories`)
- **简短**: `conn` 可以，但 `connection` 更清晰

#### 2. 文件命名

- **全小写**: `connection.go`, `mysql.go`
- **测试文件**: `connection_test.go`
- **避免下划线**: 例外：测试文件

#### 3. 接口命名

```go
// ✅ 好：动词 + 名词
type ConnectionRepository interface {}
type KeyringProvider interface {}

// ❌ 差：IInterface 风格（Java 风格）
type IConnection interface {}
```

#### 4. 错误处理

```go
// ✅ 好：添加上下文
return fmt.Errorf("failed to save connection: %w", err)

// ❌ 差：丢弃上下文
return err

// ❌ 差：无意义 wrapping
return fmt.Errorf("%w", err)
```

**错误类型定义**:

```go
package connection

import "errors"

var (
    ErrConnectionNotFound  = errors.New("connection not found")
    ErrInvalidConnection   = errors.New("invalid connection")
)
```

#### 5. Context 传递

```go
// ✅ 好：context 作为第一个参数
func (uc *ConnectionUseCase) CreateConnection(ctx context.Context, conn connection.Connection) error

// ❌ 差：context 在后面
func (uc *ConnectionUseCase) CreateConnection(conn connection.Connection, ctx context.Context) error
```

#### 6. 结构体初始化

```go
// ✅ 好：命名字段
conn := &connection.MySQLConnection{
    Host:     "localhost",
    Port:     3306,
    Database: "testdb",
}

// ❌ 差：位置依赖
conn := &connection.MySQLConnection{"", "localhost", 3306, "testdb", "root", "pass", "", time.Time{}, time.Time{}}
```

---

### 注释规范

#### 1. 包注释

```go
// Package connection provides database connection models and validation.
//
// The connection package supports four database types:
//   - MySQL
//   - Oracle
//   - SQL Server
//   - PostgreSQL
//
// Each connection type implements the Connection interface.
package connection
```

#### 2. 导出函数注释

```go
// CreateConnection creates a new database connection after validation.
//
// It validates the connection configuration, checks if a connection
// with the same name already exists, and saves the connection.
//
// Returns an error if validation fails or a connection with the
// same name already exists.
func (uc *ConnectionUseCase) CreateConnection(ctx context.Context, conn connection.Connection) error
```

#### 3. 关键逻辑注释

```go
// Use AES-256-GCM for encryption (best practice for file-based keyring)
block, err := aes.NewCipher(key)
if err != nil {
    return fmt.Errorf("create cipher: %w", err)
}

gcm, err := cipher.NewGCM(block)
if err != nil {
    return fmt.Errorf("create gcm: %w", err)
}
```

---

## 调试技巧

### 1. 使用 Delve 调试器

```bash
# 安装 Delve
go install github.com/go-delve/delve/cmd/dlv@latest

# 调试测试
dlv test ./internal/domain/connection

# 调试主程序
dlv debug ./cmd/db-benchmind-cli

# Delve 命令
(dlv) break connection.go:42    # 设置断点
(dlv) continue                  # 继续执行
(dlv) next                      # 单步执行
(dlv) print conn.Name           # 打印变量
(dlv) locals                    # 打印局部变量
```

### 2. 日志调试

```go
import "log/slog"

// 启用 debug 日志
opts := &slog.HandlerOptions{
    Level: slog.LevelDebug,
}

logger := slog.New(slog.NewTextHandler(os.Stdout, opts))
slog.SetDefault(logger)

// 使用
slog.Debug("creating connection", "id", conn.ID, "name", conn.Name)
slog.Info("connection saved successfully", "id", conn.ID)
slog.Error("failed to save connection", "error", err, "id", conn.ID)
```

### 3. 性能分析

```bash
# CPU 分析
go test -cpuprofile=cpu.prof ./internal/domain/connection
go tool pprof cpu.prof

# 内存分析
go test -memprofile=mem.prof ./internal/domain/connection
go tool pprof mem.prof

# HTTP pprof（需要在代码中启用）
import _ "net/http/pprof"
# 然后访问 http://localhost:6060/debug/pprof/
```

### 4. 竞态检测

```bash
# 运行测试时启用竞态检测
go test -race ./...
```

---

## 贡献指南

### 报告 Bug

1. 在 GitHub Issues 中创建问题
2. 使用 Bug 模板
3. 提供重现步骤
4. 包含日志和错误信息

### 提交功能

1. 先讨论：在 Issue 中讨论大功能
2. 分支：从 `main` 创建功能分支
3. 测试：确保所有测试通过
4. 文档：更新相关文档
5. PR：创建 Pull Request，引用 Issue

### 代码审查

审查要点：
- ✅ 代码清晰易懂
- ✅ 遵循项目规范
- ✅ 有充分测试
- ✅ 有必要文档
- ✅ 无安全漏洞

---

## 构建和发布

### 本地构建

```bash
# 构建所有平台
make build

# 手动构建
go build -o build/db-benchmind-cli ./cmd/db-benchmind-cli
go build -o build/db-benchmind ./cmd/db-benchmind
```

### 交叉编译

```bash
# Linux amd64
GOOS=linux GOARCH=amd64 go build -o build/db-benchmind-cli-linux-amd64 ./cmd/db-benchmind-cli

# macOS amd64
GOOS=darwin GOARCH=amd64 go build -o build/db-benchmind-cli-darwin-amd64 ./cmd/db-benchmind-cli

# Windows amd64
GOOS=windows GOARCH=amd64 go build -o build/db-benchmind-cli-windows-amd64.exe ./cmd/db-benchmind-cli
```

### 发布检查清单

- [ ] 所有测试通过
- [ ] Linter 无错误
- [ ] 安全扫描通过
- [ ] 文档已更新
- [ ] CHANGELOG 已更新
- [ ] 版本号已更新

---

## 常见问题

### Q: 如何添加新的基准测试工具？

1. 在 `pkg/benchmark/` 中定义接口（如果需要）
2. 在 `internal/infra/adapter/` 中实现适配器
3. 实现 `BenchmarkAdapter` 接口
4. 在 `AdapterRegistry` 中注册
5. 添加单元测试
6. 更新文档

### Q: 如何添加新的数据库类型？

1. 在 `internal/domain/connection/` 中定义连接结构
2. 实现 `Connection` 接口
3. 添加 `Validate()` 方法
4. 在适配器中添加支持
5. 添加测试

### Q: 依赖冲突怎么办？

```bash
# 查看依赖图
go mod graph

# 清理依赖
go mod tidy

# 更新依赖
go get -u ./...

# 验证依赖
go mod verify
```

---

**版本 1.0.0 - 完**
