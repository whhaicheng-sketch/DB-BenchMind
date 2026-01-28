# DB-BenchMind Phase 1 - Headless 执行提示词

**生成时间**: 2026-01-27 13:35
**当前进度**: Phase 1 - 70% 完成 (7/10 任务)
**剩余任务**: 3 个（1 个被阻塞 + 2 个待执行）

---

## 一、当前状态快照

### 已完成（7/10）
- ✅ Task 1.1: 项目目录结构创建
- ✅ Task 1.3: Makefile 创建
- ✅ Task 1.4: golangci-lint 配置
- ✅ Task 1.5: .gitignore 创建
- ✅ Task 1.6: 产品定义文档 (.specify/steering/product.md)
- ✅ Task 1.10: README.md 创建
- ✅ Task 1.7: 数据库 Schema 编写 (443 行，8 个表，7 个内置模板)

### 被阻塞（1/10）
- ⏭️ Task 1.2: go.mod 依赖下载（网络连接问题）

### 待执行（2/10）
- ⏳ Task 1.8: [测试] 测试 SQLite 初始化（TDD Red 阶段）
- ⏳ Task 1.9: 实现 SQLite 初始化函数（TDD Green 阶段）

---

## 二、执行目标

**目标**: 完成剩余 3 个任务，达到 Phase 1 100% 完成度。

**阻塞解除策略**:
1. 如果网络恢复 → 先执行 Task 1.2（依赖下载）
2. 如果网络仍不可用 → 使用 GOPROXY 国内镜像或离线模式

**TDD 执行顺序**（严格遵循）:
1. Task 1.8: **先写测试** → 确认测试失败（Red 阶段）
2. Task 1.9: **后写实现** → 使测试通过（Green 阶段）
3. （可选）重构代码（Refactor 阶段）

---

## 三、Task 1.2: 初始化 go.mod 依赖（解除阻塞）

### 执行步骤

#### 方案 A: 网络恢复后直接下载

```bash
# 1. 进入项目目录
cd /opt/project/DB-BenchMind

# 2. 配置 Go 代理（推荐使用国内镜像）
export GOPROXY=https://goproxy.cn,direct

# 3. 下载所有依赖
go get fyne.io/fyne/v2@latest
go get modernc.org/sqlite@latest
go get github.com/zalando/go-keyring@latest
go get github.com/google/uuid@latest

# 4. 整理依赖
go mod tidy

# 5. 验证 go.mod
cat go.mod
```

#### 方案 B: 离线模式（如果网络完全不可用）

```bash
# 如果有预下载的依赖包，手动复制到 vendor 目录
# 否则跳过此任务，先执行 Task 1.8-1.9（使用标准库 database/sql）
```

### 验收标准

- [ ] `go.mod` 文件包含所有 4 个依赖
- [ ] `go mod tidy` 无错误
- [ ] `go list -m all` 显示所有依赖包

### 注意事项

- **必须优先使用 GOPROXY 国内镜像**（https://goproxy.cn）
- 如果下载超时，尝试增加超时时间：`export GOPROXY=https://goproxy.cn,direct && export GOINSECURE=`

---

## 四、Task 1.8: [测试] 测试 SQLite 初始化（TDD Red 阶段）

**文件路径**: `internal/infra/database/sqlite_test.go`

### 依赖条件

- **前置条件**: Task 1.2 必须完成（需要 modernc.org/sqlite）
- **TDD 原则**: 先写测试，测试必须失败（因为实现不存在）

### 测试用例设计

#### Test 1: 测试数据库初始化成功

```go
func TestInitializeSQLite(t *testing.T) {
    // Arrange
    tmpDir := t.TempDir()
    dbPath := filepath.Join(tmpDir, "test.db")

    // Act
    db, err := InitializeSQLite(context.Background(), dbPath)

    // Assert
    if err != nil {
        t.Fatalf("InitializeSQLite failed: %v", err)
    }
    if db == nil {
        t.Fatal("Expected non-nil database")
    }
    defer db.Close()

    // Verify tables exist
    tables := []string{
        "connections", "templates", "tasks", "runs",
        "metric_samples", "run_logs", "reports", "settings",
    }
    for _, table := range tables {
        var count int
        err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&count)
        if err != nil {
            t.Fatalf("Failed to check table %s: %v", table, err)
        }
        if count != 1 {
            t.Errorf("Table %s not found", table)
        }
    }
}
```

#### Test 2: 测试 WAL 模式启用

```go
func TestInitializeSQLite_WALMode(t *testing.T) {
    tmpDir := t.TempDir()
    dbPath := filepath.Join(tmpDir, "test.db")

    db, err := InitializeSQLite(context.Background(), dbPath)
    if err != nil {
        t.Fatalf("InitializeSQLite failed: %v", err)
    }
    defer db.Close()

    var journalMode string
    err = db.QueryRow("PRAGMA journal_mode").Scan(&journalMode)
    if err != nil {
        t.Fatalf("Failed to query journal_mode: %v", err)
    }
    if journalMode != "wal" {
        t.Errorf("Expected journal_mode='wal', got '%s'", journalMode)
    }
}
```

#### Test 3: 测试外键约束启用

```go
func TestInitializeSQLite_ForeignKeyEnabled(t *testing.T) {
    tmpDir := t.TempDir()
    dbPath := filepath.Join(tmpDir, "test.db")

    db, err := InitializeSQLite(context.Background(), dbPath)
    if err != nil {
        t.Fatalf("InitializeSQLite failed: %v", err)
    }
    defer db.Close()

    var foreignKeys int
    err = db.QueryRow("PRAGMA foreign_keys").Scan(&foreignKeys)
    if err != nil {
        t.Fatalf("Failed to query foreign_keys: %v", err)
    }
    if foreignKeys != 1 {
        t.Errorf("Expected foreign_keys=1, got %d", foreignKeys)
    }
}
```

#### Test 4: 测试内置模板已插入

```go
func TestInitializeSQLite_BuiltinTemplates(t *testing.T) {
    tmpDir := t.TempDir()
    dbPath := filepath.Join(tmpDir, "test.db")

    db, err := InitializeSQLite(context.Background(), dbPath)
    if err != nil {
        t.Fatalf("InitializeSQLite failed: %v", err)
    }
    defer db.Close()

    var count int
    err = db.QueryRow("SELECT COUNT(*) FROM templates WHERE is_builtin=1").Scan(&count)
    if err != nil {
        t.Fatalf("Failed to count builtin templates: %v", err)
    }
    if count != 7 {
        t.Errorf("Expected 7 builtin templates, got %d", count)
    }
}
```

#### Test 5: 测试单连接池配置

```go
func TestInitializeSQLite_SingleConnection(t *testing.T) {
    tmpDir := t.TempDir()
    dbPath := filepath.Join(tmpDir, "test.db")

    db, err := InitializeSQLite(context.Background(), dbPath)
    if err != nil {
        t.Fatalf("InitializeSQLite failed: %v", err)
    }
    defer db.Close()

    stats := db.Stats()
    if stats.OpenConnections != 1 {
        t.Errorf("Expected 1 open connection, got %d", stats.OpenConnections)
    }
}
```

#### Test 6: 测试数据库已存在时重新打开

```go
func TestInitializeSQLite_ReopenExisting(t *testing.T) {
    tmpDir := t.TempDir()
    dbPath := filepath.Join(tmpDir, "test.db")

    // First initialization
    db1, err := InitializeSQLite(context.Background(), dbPath)
    if err != nil {
        t.Fatalf("First InitializeSQLite failed: %v", err)
    }
    db1.Close()

    // Second initialization (reopen)
    db2, err := InitializeSQLite(context.Background(), dbPath)
    if err != nil {
        t.Fatalf("Second InitializeSQLite failed: %v", err)
    }
    defer db2.Close()

    // Verify data persistence
    var count int
    err = db2.QueryRow("SELECT COUNT(*) FROM templates").Scan(&count)
    if err != nil {
        t.Fatalf("Failed to query templates: %v", err)
    }
    if count != 7 {
        t.Errorf("Expected 7 templates after reopen, got %d", count)
    }
}
```

### 完整测试文件结构

```go
package database_test

import (
    "context"
    "database/sql"
    "testing"
    "path/filepath"

    _ "modernc.org/sqlite" // 纯 Go SQLite 驱动
)

// InitializeSQLite 是待测试的函数（在 Task 1.9 实现）
func InitializeSQLite(ctx context.Context, dbPath string) (*sql.DB, error) {
    // Task 1.9 实现此函数
    return nil, nil
}

// 上面定义的所有测试函数...
```

### 验收标准

- [ ] 测试文件创建成功：`internal/infra/database/sqlite_test.go`
- [ ] 包含 6 个测试用例（初始化、WAL、外键、模板、连接池、重新打开）
- [ ] **运行测试失败**（因为 `InitializeSQLite` 函数返回 `nil, nil`）
- [ ] 失败原因清晰："Expected non-nil database" 或类似错误

### 执行命令

```bash
cd /opt/project/DB-BenchMind
go test -v ./internal/infra/database/
```

**预期结果**: 测试失败（Red 阶段成功）

---

## 五、Task 1.9: 实现 SQLite 初始化函数（TDD Green 阶段）

**文件路径**: `internal/infra/database/sqlite.go`

### 依赖条件

- **前置条件**: Task 1.8 完成（测试已编写）
- **TDD 原则**: 实现代码，使 Task 1.8 的所有测试通过

### 函数签名

```go
package database

import (
    "context"
    "database/sql"
    "embed"
    "fmt"
    "os"
    "path/filepath"

    _ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaFS embed.FS

// InitializeSQLite 初始化 SQLite 数据库
// ctx: 上下文（支持取消）
// dbPath: 数据库文件路径（如 "./data/db-benchmind.db"）
// 返回: 数据库连接对象（单连接池）或错误
func InitializeSQLite(ctx context.Context, dbPath string) (*sql.DB, error) {
    // 实现见下方
}
```

### 实现步骤

#### Step 1: 创建目录

```go
// 确保数据库目录存在
dbDir := filepath.Dir(dbPath)
if err := os.MkdirAll(dbDir, 0755); err != nil {
    return nil, fmt.Errorf("create db directory: %w", err)
}
```

#### Step 2: 连接数据库

```go
// SQLite 连接字符串（启用 WAL 模式）
dsn := fmt.Sprintf("file:%s?_journal_mode=WAL&_foreign_keys=on&_cache_size=64000&_synchronous=normal", dbPath)

db, err := sql.Open("sqlite", dsn)
if err != nil {
    return nil, fmt.Errorf("open sqlite: %w", err)
}
```

#### Step 3: 配置单连接池

```go
// SQLite 推荐单连接（避免锁竞争）
db.SetMaxOpenConns(1)
db.SetMaxIdleConns(1)
```

#### Step 4: 执行 Schema

```go
// 读取 schema.sql
schemaBytes, err := schemaFS.ReadFile("schema.sql")
if err != nil {
    return nil, fmt.Errorf("read schema: %w", err)
}

// 执行 Schema（支持多条语句）
_, err = db.ExecContext(ctx, string(schemaBytes))
if err != nil {
    db.Close()
    return nil, fmt.Errorf("execute schema: %w", err)
}
```

#### Step 5: 验证连接

```go
// Ping 确保连接可用
if err := db.PingContext(ctx); err != nil {
    db.Close()
    return nil, fmt.Errorf("ping database: %w", err)
}
```

#### Step 6: 返回数据库对象

```go
return db, nil
```

### 完整实现

```go
package database

import (
    "context"
    "database/sql"
    "embed"
    "fmt"
    "os"
    "path/filepath"

    _ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaFS embed.FS

// InitializeSQLite 初始化 SQLite 数据库
func InitializeSQLite(ctx context.Context, dbPath string) (*sql.DB, error) {
    // 1. 创建目录
    dbDir := filepath.Dir(dbPath)
    if err := os.MkdirAll(dbDir, 0755); err != nil {
        return nil, fmt.Errorf("create db directory: %w", err)
    }

    // 2. 连接数据库（启用 WAL 和外键）
    dsn := fmt.Sprintf("file:%s?_journal_mode=WAL&_foreign_keys=on&_cache_size=64000&_synchronous=normal", dbPath)
    db, err := sql.Open("sqlite", dsn)
    if err != nil {
        return nil, fmt.Errorf("open sqlite: %w", err)
    }

    // 3. 配置单连接池
    db.SetMaxOpenConns(1)
    db.SetMaxIdleConns(1)

    // 4. 执行 Schema
    schemaBytes, err := schemaFS.ReadFile("schema.sql")
    if err != nil {
        db.Close()
        return nil, fmt.Errorf("read schema: %w", err)
    }

    _, err = db.ExecContext(ctx, string(schemaBytes))
    if err != nil {
        db.Close()
        return nil, fmt.Errorf("execute schema: %w", err)
    }

    // 5. 验证连接
    if err := db.PingContext(ctx); err != nil {
        db.Close()
        return nil, fmt.Errorf("ping database: %w", err)
    }

    return db, nil
}
```

### 验收标准

- [ ] `sqlite.go` 文件创建成功
- [ ] 函数签名正确：`InitializeSQLite(ctx context.Context, dbPath string) (*sql.DB, error)`
- [ ] 使用 `embed.FS` 嵌入 `schema.sql`
- [ ] 使用 WAL 模式（连接字符串参数）
- [ ] 单连接池配置（`SetMaxOpenConns(1)`）
- [ ] **通过 Task 1.8 所有测试**（6 个测试全部 PASS）

### 执行命令

```bash
cd /opt/project/DB-BenchMind
go test -v ./internal/infra/database/
```

**预期结果**: 所有测试通过（Green 阶段成功）

---

## 六、执行顺序总结

```
1. 尝试解除 Task 1.2 阻塞
   ├─ 方案 A: 网络恢复 → export GOPROXY=https://goproxy.cn,direct && go get ...
   └─ 方案 B: 网络仍不可用 → 跳过，使用标准库 database/sql（无测试）

2. 执行 Task 1.8（TDD Red）
   ├─ 创建 internal/infra/database/sqlite_test.go
   ├─ 编写 6 个测试用例
   └─ 运行测试确认失败

3. 执行 Task 1.9（TDD Green）
   ├─ 创建 internal/infra/database/sqlite.go
   ├─ 实现 InitializeSQLite 函数
   └─ 运行测试确认通过

4. 更新任务追踪文件
   ├─ 编辑 .tasks/phase1-tasks.md
   └─ 标记所有任务为 ✅ 已完成

5. 验证 Phase 1 完成度
   ├─ 总任务数: 10
   ├─ 已完成: 10 (100%)
   └─ 所有验收标准通过
```

---

## 七、验证清单

### 最终验收标准

- [ ] **Task 1.2**: `go.mod` 包含所有依赖，`go mod tidy` 无错误
- [ ] **Task 1.8**: `sqlite_test.go` 创建，测试失败（Red 阶段）
- [ ] **Task 1.9**: `sqlite.go` 创建，测试通过（Green 阶段）
- [ ] **所有测试通过**: `go test ./internal/infra/database/` 返回 PASS
- [ ] **任务追踪更新**: `.tasks/phase1-tasks.md` 显示 100% 完成

### 构建验证

```bash
cd /opt/project/DB-BenchMind

# 1. 格式化代码
go fmt ./...

# 2. 运行所有测试
go test ./... -v

# 3. 尝试构建
make build  # 或 go build ./...

# 4. 检查依赖
go mod verify
```

---

## 八、注意事项

### 错误处理规范（CLAUDE.md）

- **跨边界必须 wrap 错误**: `fmt.Errorf("operation: %w", err)`
- **op 必须清晰**: "create db directory", "open sqlite", "execute schema" 等
- **禁止吞错误**: 所有错误必须返回或记录

### 测试规范（constitution.md）

- **TDD 强制**: 先写测试（Red），后写实现（Green）
- **表格驱动**: 如果需要多组测试数据，使用表格驱动测试
- **覆盖边界**: 测试正常路径 + 错误路径

### 依赖管理（CLAUDE.md）

- **每次依赖变更后**: 必须执行 `go mod tidy`
- **禁止引入未记录依赖**: 新增依赖必须说明原因

### Git 提交（如果需要）

```bash
git add -A
git commit -m "feat(phase1): complete Phase 1 - project infrastructure

- Task 1.2: Initialize go.mod with dependencies
- Task 1.8: Add SQLite initialization tests (TDD Red)
- Task 1.9: Implement SQLite initialization (TDD Green)

All 10 tasks completed (100%)."
```

---

## 九、参考文件

- **项目宪法**: `/opt/project/DB-BenchMind/constitution.md`
- **开发规范**: `/opt/project/DB-BenchMind/CLAUDE.md`
- **任务追踪**: `/opt/project/DB-BenchMind/.tasks/phase1-tasks.md`
- **Schema 文件**: `/opt/project/DB-BenchMind/internal/infra/database/schema.sql`
- **技术计划**: `/opt/project/DB-BenchMind/specs/plan.md`

---

## 十、联系方式

如果遇到阻塞问题，查看：
- 阻塞问题日志：`.tasks/phase1-tasks.md` 的 "阻塞问题" 章节
- 风险与问题：`.tasks/phase1-tasks.md` 的 "风险与问题" 章节

**祝执行顺利！**

---

**文档版本**: 1.0
**生成工具**: Claude (Sonnet 4.5)
**项目**: DB-BenchMind
**阶段**: Phase 1 - 项目初始化与基础设施
