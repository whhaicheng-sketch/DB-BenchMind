# DB-BenchMind 测试文档

**版本**: 1.0.0
**更新日期**: 2026-01-28

---

## 目录

1. [测试策略](#测试策略)
2. [单元测试](#单元测试)
3. [集成测试](#集成测试)
4. [E2E 测试](#e2e-测试)
5. [性能测试](#性能测试)
6. [测试覆盖率](#测试覆盖率)
7. [CI/CD 集成](#cicd-集成)

---

## 测试策略

### 测试金字塔

```
        /\
       /  \
      / E2E \        5% - 端到端测试
     /--------\
    /  集成测试  \     15% - 集成测试
   /--------------\
  /    单元测试      \   80% - 单元测试
 /--------------------\
```

### 测试原则

1. **TDD**: 测试先行（Red → Green → Refactor）
2. **快速**: 单元测试应该在毫秒级完成
3. **独立**: 测试之间不依赖顺序
4. **可重复**: 测试结果应该一致
5. **清晰**: 测试名称和断言应该清楚表达意图

---

## 单元测试

### 目标

- 验证单个函数/方法的行为
- 无外部依赖（数据库、网络、文件系统）
- 快速执行（毫秒级）

### 示例：Domain 层测试

```go
// internal/domain/connection/mysql_test.go
package connection

import (
    "testing"
)

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
                ID:       "test-1",
                Name:     "test-conn",
                Host:     "localhost",
                Port:     3306,
                Database: "testdb",
                Username: "root",
                Password: "pass",
            },
            wantErr: false,
        },
        {
            name: "invalid port - negative",
            conn: &MySQLConnection{
                Name:     "test",
                Host:     "localhost",
                Port:     -1,
                Database: "testdb",
                Username: "root",
            },
            wantErr: true,
            errMsg:  "port must be between 1 and 65535",
        },
        {
            name: "invalid port - too large",
            conn: &MySQLConnection{
                Name:     "test",
                Host:     "localhost",
                Port:     99999,
                Database: "testdb",
                Username: "root",
            },
            wantErr: true,
            errMsg:  "port must be between 1 and 65535",
        },
        {
            name: "missing name",
            conn: &MySQLConnection{
                Name:     "",
                Host:     "localhost",
                Port:     3306,
                Database: "testdb",
                Username: "root",
            },
            wantErr: true,
            errMsg:  "name is required",
        },
        {
            name: "missing host",
            conn: &MySQLConnection{
                Name:     "test",
                Host:     "",
                Port:     3306,
                Database: "testdb",
                Username: "root",
            },
            wantErr: true,
            errMsg:  "host is required",
        },
        {
            name: "missing database",
            conn: &MySQLConnection{
                Name:     "test",
                Host:     "localhost",
                Port:     3306,
                Database: "",
                Username: "root",
            },
            wantErr: true,
            errMsg:  "database is required",
        },
        {
            name: "missing username",
            conn: &MySQLConnection{
                Name:     "test",
                Host:     "localhost",
                Port:     3306,
                Database: "testdb",
                Username: "",
            },
            wantErr: true,
            errMsg:  "username is required",
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
                if !containsString(err.Error(), tt.errMsg) {
                    t.Errorf("Validate() error = %v, want contain %v", err.Error(), tt.errMsg)
                }
            }
        })
    }
}

func containsString(s, substr string) bool {
    return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
           func() bool {
               for i := 0; i <= len(s)-len(substr); i++ {
                   if s[i:i+len(substr)] == substr {
                       return true
                   }
               }
               return false
           }())
}
```

### 示例：UseCase 层测试（使用 Mock）

```go
// internal/app/usecase/connection_usecase_test.go
package usecase

import (
    "context"
    "errors"
    "testing"

    "github.com/whhaicheng/DB-BenchMind/internal/domain/connection"
)

// Mock 实现
type mockConnectionRepository struct {
    connections map[string]connection.Connection
    saveErr     error
}

func (m *mockConnectionRepository) Save(ctx context.Context, conn connection.Connection) error {
    if m.saveErr != nil {
        return m.saveErr
    }
    m.connections[conn.GetID()] = conn
    return nil
}

func (m *mockConnectionRepository) FindByID(ctx context.Context, id string) (connection.Connection, error) {
    conn, ok := m.connections[id]
    if !ok {
        return nil, ErrConnectionNotFound
    }
    return conn, nil
}

func (m *mockConnectionRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
    for _, conn := range m.connections {
        if conn.GetName() == name {
            return true, nil
        }
    }
    return false, nil
}

// ... 其他方法

func TestConnectionUseCase_CreateConnection(t *testing.T) {
    tests := []struct {
        name        string
        conn        connection.Connection
        setupMock   func(*mockConnectionRepository)
        wantErr     bool
        expectedErr error
    }{
        {
            name: "successful creation",
            conn: &connection.MySQLConnection{
                ID:       "test-1",
                Name:     "Test Connection",
                Host:     "localhost",
                Port:     3306,
                Database: "testdb",
                Username: "root",
            },
            setupMock: func(m *mockConnectionRepository) {
                m.connections = make(map[string]connection.Connection)
            },
            wantErr: false,
        },
        {
            name: "duplicate name",
            conn: &connection.MySQLConnection{
                ID:       "test-2",
                Name:     "Duplicate Name",
                Host:     "localhost",
                Port:     3306,
                Database: "testdb",
                Username: "root",
            },
            setupMock: func(m *mockConnectionRepository) {
                m.connections = map[string]connection.Connection{
                    "test-1": &connection.MySQLConnection{Name: "Duplicate Name"},
                }
            },
            wantErr:     true,
            expectedErr: ErrConnectionAlreadyExists,
        },
        {
            name: "invalid connection",
            conn: &connection.MySQLConnection{
                Name: "",
                Host: "localhost",
                Port: 3306,
            },
            setupMock: func(m *mockConnectionRepository) {
                m.connections = make(map[string]connection.Connection)
            },
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mock := &mockConnectionRepository{}
            tt.setupMock(mock)

            uc := NewConnectionUseCase(mock, nil)
            err := uc.CreateConnection(context.Background(), tt.conn)

            if (err != nil) != tt.wantErr {
                t.Errorf("CreateConnection() error = %v, wantErr %v", err, tt.wantErr)
                return
            }

            if tt.wantErr && tt.expectedErr != nil {
                if !errors.Is(err, tt.expectedErr) {
                    t.Errorf("CreateConnection() error = %v, want %v", err, tt.expectedErr)
                }
            }
        })
    }
}
```

### 运行单元测试

```bash
# 运行所有单元测试
go test ./...

# 运行特定包
go test ./internal/domain/connection

# 详细输出
go test -v ./internal/domain/connection

# 覆盖率
go test -cover ./internal/domain/connection

# 生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

---

## 集成测试

### 目标

- 验证组件间协作
- 使用真实数据库（SQLite in-memory）
- 测试完整的工作流

### 示例：连接管理集成测试

```go
// test/integration/connection_workflow_test.go
package integration

import (
    "context"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"

    "github.com/whhaicheng/DB-BenchMind/internal/app/usecase"
    "github.com/whhaicheng/DB-BenchMind/internal/domain/connection"
    "github.com/whhaicheng/DB-BenchMind/internal/infra/database"
    "github.com/whhaicheng/DB-BenchMind/internal/infra/database/repository"
    "github.com/whhaicheng/DB-BenchMind/internal/infra/keyring"
)

func TestIntegration_ConnectionWorkflow(t *testing.T) {
    // 使用内存数据库
    ctx := context.Background()
    db, err := database.InitializeSQLite(ctx, ":memory:")
    require.NoError(t, err)
    defer db.Close()

    // 初始化
    connRepo := repository.NewSQLiteConnectionRepository(db)
    keyringProvider, err := keyring.NewFileFallback("", "test")
    require.NoError(t, err)
    connUC := usecase.NewConnectionUseCase(connRepo, keyringProvider)

    t.Run("create and retrieve connection", func(t *testing.T) {
        // 创建连接
        conn := &connection.MySQLConnection{
            ID:       "test-1",
            Name:     "Test Connection",
            Host:     "localhost",
            Port:     3306,
            Database: "testdb",
            Username: "root",
            Password: "password",
            CreatedAt: time.Now(),
            UpdatedAt: time.Now(),
        }

        err := connUC.CreateConnection(ctx, conn)
        require.NoError(t, err)

        // 查询连接
        retrieved, err := connUC.GetConnection(ctx, conn.GetID())
        require.NoError(t, err)

        assert.Equal(t, conn.GetID(), retrieved.GetID())
        assert.Equal(t, conn.GetName(), retrieved.GetName())
        assert.Equal(t, conn.GetType(), retrieved.GetType())
    })

    t.Run("list connections", func(t *testing.T) {
        // 创建多个连接
        connections := []connection.Connection{
            &connection.MySQLConnection{
                ID:       "mysql-1",
                Name:     "MySQL 1",
                Host:     "localhost",
                Port:     3306,
                Database: "db1",
                Username: "root",
                CreatedAt: time.Now(),
                UpdatedAt: time.Now(),
            },
            &connection.PostgreSQLConnection{
                ID:       "pg-1",
                Name:     "PostgreSQL 1",
                Host:     "localhost",
                Port:     5432,
                Database: "db2",
                Username: "postgres",
                CreatedAt: time.Now(),
                UpdatedAt: time.Now(),
            },
        }

        for _, conn := range connections {
            err := connUC.CreateConnection(ctx, conn)
            require.NoError(t, err)
        }

        // 列出所有连接
        all, err := connUC.ListConnections(ctx)
        require.NoError(t, err)

        assert.GreaterOrEqual(t, len(all), 2)
    })

    t.Run("update connection", func(t *testing.T) {
        conn := &connection.MySQLConnection{
            ID:       "test-update",
            Name:     "Original Name",
            Host:     "localhost",
            Port:     3306,
            Database: "testdb",
            Username: "root",
            CreatedAt: time.Now(),
            UpdatedAt: time.Now(),
        }

        err := connUC.CreateConnection(ctx, conn)
        require.NoError(t, err)

        // 更新连接
        conn.SetName("Updated Name")
        err = connUC.UpdateConnection(ctx, conn)
        require.NoError(t, err)

        // 验证更新
        updated, err := connUC.GetConnection(ctx, conn.GetID())
        require.NoError(t, err)

        assert.Equal(t, "Updated Name", updated.GetName())
    })

    t.Run("delete connection", func(t *testing.T) {
        conn := &connection.MySQLConnection{
            ID:       "test-delete",
            Name:     "To Delete",
            Host:     "localhost",
            Port:     3306,
            Database: "testdb",
            Username: "root",
            CreatedAt: time.Now(),
            UpdatedAt: time.Now(),
        }

        err := connUC.CreateConnection(ctx, conn)
        require.NoError(t, err)

        // 删除连接
        err = connUC.DeleteConnection(ctx, conn.GetID())
        require.NoError(t, err)

        // 验证删除
        _, err = connUC.GetConnection(ctx, conn.GetID())
        assert.Error(t, err)
    })

    t.Run("duplicate name error", func(t *testing.T) {
        conn1 := &connection.MySQLConnection{
            ID:       "dup-1",
            Name:     "Duplicate",
            Host:     "localhost",
            Port:     3306,
            Database: "db1",
            Username: "root",
            CreatedAt: time.Now(),
            UpdatedAt: time.Now(),
        }

        conn2 := &connection.MySQLConnection{
            ID:       "dup-2",
            Name:     "Duplicate",  // 同名
            Host:     "localhost",
            Port:     3306,
            Database: "db2",
            Username: "root",
            CreatedAt: time.Now(),
            UpdatedAt: time.Now(),
        }

        err := connUC.CreateConnection(ctx, conn1)
        require.NoError(t, err)

        err = connUC.CreateConnection(ctx, conn2)
        assert.Error(t, err)
        assert.ErrorIs(t, err, usecase.ErrConnectionAlreadyExists)
    })

    t.Run("validation error", func(t *testing.T) {
        invalidConn := &connection.MySQLConnection{
            ID:   "invalid",
            Name: "",  // 空名称
            Host: "localhost",
            Port: 3306,
        }

        err := connUC.CreateConnection(ctx, invalidConn)
        assert.Error(t, err)
    })
}

func TestIntegration_MultipleConnectionTypes(t *testing.T) {
    ctx := context.Background()
    db, err := database.InitializeSQLite(ctx, ":memory:")
    require.NoError(t, err)
    defer db.Close()

    connRepo := repository.NewSQLiteConnectionRepository(db)
    keyringProvider, err := keyring.NewFileFallback("", "test")
    require.NoError(t, err)
    connUC := usecase.NewConnectionUseCase(connRepo, keyringProvider)

    // 创建不同类型的连接
    connections := []connection.Connection{
        &connection.MySQLConnection{
            ID:       "mysql-1",
            Name:     "MySQL Test",
            Host:     "localhost",
            Port:     3306,
            Database: "mysql_db",
            Username: "root",
            CreatedAt: time.Now(),
            UpdatedAt: time.Now(),
        },
        &connection.PostgreSQLConnection{
            ID:       "pg-1",
            Name:     "PostgreSQL Test",
            Host:     "localhost",
            Port:     5432,
            Database: "pg_db",
            Username: "postgres",
            CreatedAt: time.Now(),
            UpdatedAt: time.Now(),
        },
        &connection.OracleConnection{
            ID:       "ora-1",
            Name:     "Oracle Test",
            Host:     "localhost",
            Port:     1521,
            SID:      "ORCL",
            Username: "system",
            CreatedAt: time.Now(),
            UpdatedAt: time.Now(),
        },
        &connection.SQLServerConnection{
            ID:       "mssql-1",
            Name:     "SQL Server Test",
            Host:     "localhost",
            Port:     1433,
            Database: "mssql_db",
            Username: "sa",
            CreatedAt: time.Now(),
            UpdatedAt: time.Now(),
        },
    }

    for _, conn := range connections {
        err := connUC.CreateConnection(ctx, conn)
        require.NoError(t, err, "should create %s connection", conn.GetType())
    }

    // 验证所有连接都已保存
    all, err := connUC.ListConnections(ctx)
    require.NoError(t, err)
    assert.Len(t, all, 4)
}
```

### 运行集成测试

```bash
# 运行所有测试（包括集成测试）
go test ./... -tags=integration

# 仅运行集成测试
go test ./test/integration/...

# 详细输出
go test -v ./test/integration/...
```

---

## E2E 测试

### 目标

- 测试完整用户场景
- 可以使用真实工具（sysbench）
- 最少但最重要

### 示例：完整测试流程

```go
// test/e2e/benchmark_test.go
package e2e

import (
    "context"
    "os"
    "os/exec"
    "path/filepath"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"

    "github.com/whhaicheng/DB-BenchMind/internal/app/usecase"
    "github.com/whhaicheng/DB-BenchMind/internal/domain/connection"
    "github.com/whhaicheng/DB-BenchMind/internal/domain/execution"
    "github.com/whhaicheng/DB-BenchMind/internal/infra/adapter"
    "github.com/whhaicheng/DB-BenchMind/internal/infra/database"
    "github.com/whhaicheng/DB-BenchMind/internal/infra/database/repository"
    "github.com/whhaicheng/DB-BenchMind/internal/infra/keyring"
)

func TestE2E_SysbenchBenchmark(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping E2E test")
    }

    // 跳过如果 sysbench 未安装
    if _, err := exec.LookPath("sysbench"); err != nil {
        t.Skip("sysbench not found")
    }

    ctx := context.Background()

    // 创建临时目录
    tmpDir, err := os.MkdirTemp("", "db-benchmind-e2e-*")
    require.NoError(t, err)
    defer os.RemoveAll(tmpDir)

    dbPath := filepath.Join(tmpDir, "test.db")
    resultsDir := filepath.Join(tmpDir, "results")
    os.MkdirAll(resultsDir, 0755)

    // 初始化数据库
    db, err := database.InitializeSQLite(ctx, dbPath)
    require.NoError(t, err)
    defer db.Close()

    // 初始化组件
    connRepo := repository.NewSQLiteConnectionRepository(db)
    runRepo := repository.NewSQLiteRunRepository(db)
    templateRepo := repository.NewTemplateRepository(db)

    keyringProvider, err := keyring.NewFileFallback(tmpDir, "test")
    require.NoError(t, err)

    adapterRegistry := adapter.NewRegistry()
    sysbenchAdapter := adapter.NewSysbenchAdapter("sysbench")
    adapterRegistry.Register("sysbench", sysbenchAdapter)

    connUC := usecase.NewConnectionUseCase(connRepo, keyringProvider)
    templateUC := usecase.NewTemplateUseCase(templateRepo)
    benchmarkUC := usecase.NewBenchmarkUseCase(runRepo, adapterRegistry, keyringProvider)

    // 1. 创建连接（使用 SQLite 作为测试数据库）
    testConn := &connection.MySQLConnection{
        ID:       "e2e-test-conn",
        Name:     "E2E Test Connection",
        Host:     "localhost",
        Port:     3306,
        Database: "sbtest",  // 假设已有测试数据库
        Username: "root",
        Password: "",
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }

    // 如果没有真实的 MySQL，跳过
    t.Run("prepare database connection", func(t *testing.T) {
        if testing.Short() {
            t.Skip("need real database")
        }
    })

    // 2. 获取内置模板
    t.Run("get builtin template", func(t *testing.T) {
        templates, err := templateUC.ListBuiltinTemplates(ctx)
        require.NoError(t, err)
        assert.NotEmpty(t, templates)
    })

    // 3. 创建任务（需要真实数据库）
    t.Run("execute benchmark", func(t *testing.T) {
        if testing.Short() {
            t.Skip("need real database")
        }

        // 获取模板
        templates, err := templateUC.ListBuiltinTemplates(ctx)
        require.NoError(t, err)

        sysbenchTemplate := templates[0]
        assert.Equal(t, "sysbench", sysbenchTemplate.Tool)

        // 创建任务
        task := &execution.Task{
            ID:           "e2e-task-1",
            Name:         "E2E Benchmark Test",
            ConnectionID: testConn.GetID(),
            TemplateID:   sysbenchTemplate.ID,
            Parameters: map[string]interface{}{
                "threads": 2,
                "time":    10,
            },
            CreatedAt: time.Now(),
        }

        // 执行任务
        run, err := benchmarkUC.ExecuteTask(ctx, task)
        require.NoError(t, err)

        assert.Equal(t, execution.StatePending, run.State)

        // 等待完成（最多 2 分钟）
        timeout := time.After(2 * time.Minute)
        ticker := time.NewTicker(5 * time.Second)
        defer ticker.Stop()

        for {
            select {
            case <-timeout:
                t.Fatal("benchmark timeout")
            case <-ticker.C:
                status, err := benchmarkUC.GetRunStatus(ctx, run.ID)
                require.NoError(t, err)

                t.Logf("Status: %s, Progress: %.1f%%", status.State, status.Progress)

                if status.State == execution.StateCompleted {
                    // 获取结果
                    result, err := benchmarkUC.GetRunResult(ctx, run.ID)
                    require.NoError(t, err)

                    assert.Greater(t, result.TPSCalculated, 0.0)
                    assert.Greater(t, result.LatencyAvg, 0.0)

                    t.Logf("TPS: %.2f", result.TPSCalculated)
                    t.Logf("Avg Latency: %.2f ms", result.LatencyAvg)

                    return
                }

                if status.State == execution.StateFailed {
                    t.Fatal("benchmark failed")
                }
            }
        }
    })
}
```

### 运行 E2E 测试

```bash
# 运行 E2E 测试
go test ./test/e2e/...

# 跳过 E2E 测试（快速测试）
go test ./... -short

# 运行特定 E2E 测试
go test -v ./test/e2e/ -run TestE2E_SysbenchBenchmark
```

---

## 性能测试

### Benchmark 测试

```go
// internal/domain/connection/mysql_benchmark_test.go
package connection

import (
    "testing"
)

func BenchmarkMySQLConnection_Validate(b *testing.B) {
    conn := &MySQLConnection{
        ID:       "bench-1",
        Name:     "Benchmark Connection",
        Host:     "localhost",
        Port:     3306,
        Database: "testdb",
        Username: "root",
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        conn.Validate()
    }
}

func BenchmarkMySQLConnection_ToJSON(b *testing.B) {
    conn := &MySQLConnection{
        ID:       "bench-1",
        Name:     "Benchmark Connection",
        Host:     "localhost",
        Port:     3306,
        Database: "testdb",
        Username: "root",
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, _ = conn.ToJSON()
    }
}
```

### 运行 Benchmark

```bash
# 运行 benchmark
go test -bench=. -benchmem ./internal/domain/connection

# 运行特定 benchmark
go test -bench=BenchmarkMySQLConnection -benchmem ./internal/domain/connection

# CPU profile
go test -bench=. -cpuprofile=cpu.prof ./internal/domain/connection
go tool pprof cpu.prof
```

---

## 测试覆盖率

### 查看覆盖率

```bash
# 生成覆盖率报告
go test -coverprofile=coverage.out ./...

# 查看总体覆盖率
go tool cover -func=coverage.out

# 生成 HTML 报告
go tool cover -html=coverage.out -o coverage.html

# 按包查看覆盖率
go test -cover ./internal/...
```

### 覆盖率目标

- **总体**: ≥ 80%
- **Domain 层**: ≥ 90%
- **UseCase 层**: ≥ 85%
- **Infra 层**: ≥ 75%

---

## CI/CD 集成

### GitHub Actions 示例

```yaml
# .github/workflows/test.yml
name: Test

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest

    steps:
    - uses: actions/checkout@v3

    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.22.2'

    - name: Download dependencies
      run: go mod download

    - name: Verify dependencies
      run: go mod verify

    - name: Run gofmt
      run: test -z $(gofmt -l .)

    - name: Run go vet
      run: go vet ./...

    - name: Run golangci-lint
      uses: golangci/golangci-lint-action@v3
      with:
        version: latest

    - name: Run tests
      run: go test -v -race -coverprofile=coverage.out ./...

    - name: Upload coverage to Codecov
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.out

    - name: Run security scan
      run: |
        go install golang.org/x/vuln/cmd/govulncheck@latest
        govulncheck ./...
```

---

## 测试最佳实践

### 1. 命名规范

```go
// ✅ 好：清晰表达测试意图
func TestMySQLConnection_Validate_ValidConnection(t *testing.T)
func TestMySQLConnection_Validate_InvalidPort(t *testing.T)
func TestConnectionUseCase_CreateConnection_DuplicateName(t *testing.T)

// ❌ 差：不清晰
func TestConnection1(t *testing.T)
func TestCreate(t *testing.T)
```

### 2. 表格驱动测试

```go
// ✅ 好：表格驱动，易于扩展
func TestValidate(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {"valid input", "valid", false},
        {"empty input", "", true},
        {"invalid input", "invalid", true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // ...
        })
    }
}

// ❌ 差：重复代码
func TestValidate(t *testing.T) {
    err := Validate("valid")
    if err != nil { t.Error() }

    err = Validate("")
    if err == nil { t.Error() }

    // ...
}
```

### 3. 使用 t.Helper()

```go
func setupTestDB(t *testing.T) *sql.DB {
    t.Helper()

    db, err := sql.Open("sqlite", ":memory:")
    if err != nil {
        t.Fatal(err)
    }
    return db
}

// 错误会报告到调用位置，而不是 setupTestDB 内部
```

### 4. 清理资源

```go
func TestWithCleanup(t *testing.T) {
    tmpDir, err := os.MkdirTemp("", "test")
    require.NoError(t, err)

    // 确保清理
    t.Cleanup(func() {
        os.RemoveAll(tmpDir)
    })

    // 测试代码...
}
```

---

**版本 1.0.0 - 完**
