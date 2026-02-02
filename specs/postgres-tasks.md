# DB-BenchMind PostgreSQL 支持任务分解列表

**版本**: 1.0.0
**日期**: 2026-02-02
**状态**: 待执行
**技术负责人**: AI Assistant

---

## 📋 文档说明

本文档将 `specs/postgres-plan.md` 中的技术方案分解为**原子化、可执行的任务列表**，确保每个任务都可以被 AI 独立完成。

### 任务格式说明

```markdown
#### Task X.Y: [任务标题] [P]
**Type**: test/impl
**File**: 文件路径
**Depends**: X.A, X.B（依赖的任务ID）
**Estimate**: 估算时间（分钟）
**Description**: 详细描述
**Acceptance**: 验收标准
**Implementation**: 实现要点
```

- **[P]** 标记表示该任务可与其他标记 `[P]` 的任务并行执行
- **Type**: `test` 表示测试任务，`impl` 表示实现任务
- **TDD 强制**: 测试任务必须在实现任务之前（Red → Green → Refactor）

---

## Phase 1: 驱动集成与测试实现

**目标**: 实现核心 PostgreSQL 连接测试功能

---

#### Task 1.1: 添加 PostgreSQL 驱动依赖 [P]
**Type**: impl
**File**: `go.mod`
**Depends**: 无
**Estimate**: 5 分钟
**Description**: 在 go.mod 中添加 github.com/lib/pq PostgreSQL 驱动依赖
**Acceptance**:
- `go.mod` 包含 `github.com/lib/pq` 依赖
- `go mod tidy` 执行成功
- `go build ./...` 编译成功
- 无依赖冲突
**Implementation**:
```bash
cd /opt/project/DB-BenchMind
go get github.com/lib/pq
go mod tidy
go build ./...
```
**Verification**:
```bash
grep "lib/pq" go.mod
go test ./... -run=nonexistent  # Quick build check
```

---

#### Task 1.2: 导入 PostgreSQL 驱动 [P]
**Type**: impl
**File**: `internal/domain/connection/postgresql.go`
**Depends**: 1.1
**Estimate**: 5 分钟
**Description**: 在 postgresql.go 文件中添加 database/sql 和 lib/pq 驱动的导入语句
**Acceptance**:
- 文件编译成功
- `database/sql` 标准库已导入
- `_ "github.com/lib/pq"` 驱动已导入（匿名导入）
- 导入语句按规范顺序排列（标准库 → 第三方 → 内部）
**Implementation**:
在文件开头修改导入块：
```go
import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq" // Register PostgreSQL driver
)
```
**Verification**:
```bash
go build ./internal/domain/connection
```

---

#### Task 1.3: 实现 PostgreSQL 连接测试
**Type**: impl
**File**: `internal/domain/connection/postgresql.go:107-127`
**Depends**: 1.2
**Estimate**: 30 分钟
**Description**: 实现 PostgreSQLConnection.Test() 方法，替换当前的占位符实现，参考 MySQL 实现模式
**Acceptance**:
- `Test()` 方法完整实现
- 成功路径：返回 `Success=true`, `Version`, `LatencyMs`
- 失败路径：返回 `Success=false`, `Error` 字符串
- 超时控制：5 秒上下文超时
- 正确关闭数据库连接（defer db.Close()）
- 使用 `GetDSNWithPassword()` 获取完整 DSN
- SQL 查询：`SELECT version()` 获取版本
**Implementation**:
```go
func (c *PostgreSQLConnection) Test(ctx context.Context) (*TestResult, error) {
	start := time.Now()

	// Build DSN with password
	dsn := c.GetDSNWithPassword()

	// Open connection
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return &TestResult{
			Success:   false,
			LatencyMs: time.Since(start).Milliseconds(),
			Error:     fmt.Sprintf("Failed to open connection: %v", err),
		}, nil
	}
	defer db.Close()

	// Set timeout
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Test connection
	if err := db.PingContext(ctx); err != nil {
		return &TestResult{
			Success:   false,
			LatencyMs: time.Since(start).Milliseconds(),
			Error:     fmt.Sprintf("Connection failed: %v", err),
		}, nil
	}

	// Query PostgreSQL version
	var version string
	err = db.QueryRow("SELECT version()").Scan(&version)
	if err != nil {
		version = "Unknown"
	}

	latency := time.Since(start).Milliseconds()

	return &TestResult{
		Success:   true,
		LatencyMs: latency,
		Version:   version,
		Error:     "",
	}, nil
}
```
**Verification**:
```bash
go build ./internal/domain/connection
```

---

#### Task 1.4: 编写 PostgreSQL 连接单元测试
**Type**: test
**File**: `internal/domain/connection/postgresql_test.go` [NEW]
**Depends**: 1.3
**Estimate**: 60 分钟
**Description**: 创建 PostgreSQL 连接的单元测试，覆盖参数验证、DSN 生成、密码管理等功能
**Acceptance**:
- 所有测试通过
- 测试覆盖率 > 80%
- 使用表格驱动测试（Table-Driven Tests）
- 测试边界条件（空值、无效值、极端值）
**Implementation**:
创建新文件 `internal/domain/connection/postgresql_test.go`：

```go
package connection

import (
	"context"
	"testing"
	"time"
)

// TestPostgreSQLConnection_Validate_ValidInput tests validation with valid input
func TestPostgreSQLConnection_Validate_ValidInput(t *testing.T) {
	conn := &PostgreSQLConnection{
		BaseConnection: BaseConnection{
			Name: "Test PG",
		},
		Host:     "localhost",
		Port:     5432,
		Database: "testdb",
		Username: "postgres",
		SSLMode:  "prefer",
	}

	err := conn.Validate()
	if err != nil {
		t.Errorf("Validate() should succeed with valid input, got error: %v", err)
	}
}

// TestPostgreSQLConnection_Validate_MissingRequiredFields tests validation with missing required fields
func TestPostgreSQLConnection_Validate_MissingRequiredFields(t *testing.T) {
	tests := []struct {
		name    string
		conn    *PostgreSQLConnection
		wantErr bool
	}{
		{
			name: "Missing Name",
			conn: &PostgreSQLConnection{
				Host:     "localhost",
				Port:     5432,
				Username: "postgres",
			},
			wantErr: true,
		},
		{
			name: "Missing Host",
			conn: &PostgreSQLConnection{
				BaseConnection: BaseConnection{Name: "Test"},
				Port:           5432,
				Username:      "postgres",
			},
			wantErr: true,
		},
		{
			name: "Missing Username",
			conn: &PostgreSQLConnection{
				BaseConnection: BaseConnection{Name: "Test"},
				Host:           "localhost",
				Port:           5432,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.conn.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestPostgreSQLConnection_Validate_InvalidPort tests validation with invalid port
func TestPostgreSQLConnection_Validate_InvalidPort(t *testing.T) {
	tests := []struct {
		name    string
		port    int
		wantErr bool
	}{
		{"Port zero", 0, true},
		{"Port negative", -1, true},
		{"Port too large", 65536, true},
		{"Valid port", 5432, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn := &PostgreSQLConnection{
				BaseConnection: BaseConnection{Name: "Test"},
				Host:           "localhost",
				Port:           tt.port,
				Username:       "postgres",
			}
			err := conn.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestPostgreSQLConnection_Validate_InvalidSSLMode tests validation with invalid SSL mode
func TestPostgreSQLConnection_Validate_InvalidSSLMode(t *testing.T) {
	conn := &PostgreSQLConnection{
		BaseConnection: BaseConnection{Name: "Test"},
		Host:           "localhost",
		Port:           5432,
		Username:       "postgres",
		SSLMode:        "invalid-mode",
	}

	err := conn.Validate()
	if err == nil {
		t.Error("Validate() should fail with invalid SSL mode")
	}
}

// TestPostgreSQLConnection_GetDSN tests DSN generation
func TestPostgreSQLConnection_GetDSN(t *testing.T) {
	conn := &PostgreSQLConnection{
		BaseConnection: BaseConnection{Name: "Test"},
		Host:           "localhost",
		Port:           5432,
		Database:       "testdb",
		Username:       "postgres",
	}

	expected := "host=localhost port=5432 database=testdb user=postgres"
	got := conn.GetDSN()

	if got != expected {
		t.Errorf("GetDSN() = %q, want %q", got, expected)
	}
}

// TestPostgreSQLConnection_GetDSNWithPassword tests DSN generation with password
func TestPostgreSQLConnection_GetDSNWithPassword(t *testing.T) {
	conn := &PostgreSQLConnection{
		BaseConnection: BaseConnection{Name: "Test"},
		Host:           "localhost",
		Port:           5432,
		Database:       "testdb",
		Username:       "postgres",
		Password:       "secret",
		SSLMode:        "require",
	}

	expected := "host=localhost port=5432 database=testdb user=postgres password=secret sslmode=require"
	got := conn.GetDSNWithPassword()

	if got != expected {
		t.Errorf("GetDSNWithPassword() = %q, want %q", got, expected)
	}
}

// TestPostgreSQLConnection_Redact tests redaction for display
func TestPostgreSQLConnection_Redact(t *testing.T) {
	conn := &PostgreSQLConnection{
		BaseConnection: BaseConnection{
			Name: "Production DB",
		},
		Host:     "prod.example.com",
		Port:     5432,
		Database: "production",
	}

	expected := "Production DB (***@prod.example.com:5432/production)"
	got := conn.Redact()

	if got != expected {
		t.Errorf("Redact() = %q, want %q", got, expected)
	}
}

// TestPostgreSQLConnection_SetPassword_GetPassword tests password management
func TestPostgreSQLConnection_SetPassword_GetPassword(t *testing.T) {
	conn := &PostgreSQLConnection{
		BaseConnection: BaseConnection{Name: "Test"},
	}

	// Test SetPassword
	conn.SetPassword("my-secret-password")
	if conn.Password != "my-secret-password" {
		t.Errorf("SetPassword() failed, Password = %q", conn.Password)
	}

	// Test GetPassword
	got := conn.GetPassword()
	if got != "my-secret-password" {
		t.Errorf("GetPassword() = %q, want %q", got, "my-secret-password")
	}

	// Verify UpdatedAt is set
	if conn.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should be set after SetPassword")
	}
}

// TestPostgreSQLConnection_Test_Success tests successful connection test
// NOTE: This test requires a running PostgreSQL server
// Skip if not available
func TestPostgreSQLConnection_Test_Success(t *testing.T) {
	// This test is optional and requires a real PostgreSQL server
	t.Skip("Skipping - requires PostgreSQL server")

	conn := &PostgreSQLConnection{
		BaseConnection: BaseConnection{
			Name: "Test",
		},
		Host:     "localhost",
		Port:     5432,
		Database: "postgres",
		Username: "postgres",
		Password: "postgres", // Use env variable in real test
		SSLMode:  "disable",
	}

	ctx := context.Background()
	result, err := conn.Test(ctx)

	if err != nil {
		t.Fatalf("Test() failed: %v", err)
	}

	if !result.Success {
		t.Errorf("Test() Success = false, error = %q", result.Error)
	}

	if result.LatencyMs <= 0 {
		t.Errorf("Test() LatencyMs = %d, want > 0", result.LatencyMs)
	}

	if result.Version == "" {
		t.Error("Test() Version should not be empty on success")
	}
}

// TestPostgreSQLConnection_Test_Failure tests failed connection test
func TestPostgreSQLConnection_Test_Failure(t *testing.T) {
	conn := &PostgreSQLConnection{
		BaseConnection: BaseConnection{
			Name: "Test",
		},
		Host:     "invalid-host-that-does-not-exist.local",
		Port:     5432,
		Database: "testdb",
		Username: "postgres",
		Password: "test",
		SSLMode:  "disable",
	}

	ctx := context.Background()
	result, err := conn.Test(ctx)

	if err != nil {
		t.Fatalf("Test() should not return error, got: %v", err)
	}

	if result.Success {
		t.Error("Test() Success = true, want false")
	}

	if result.Error == "" {
		t.Error("Test() Error should not be empty on failure")
	}
}
```

**Verification**:
```bash
go test -v ./internal/domain/connection -run TestPostgreSQL
go test ./internal/domain/connection -coverprofile=coverage.out
go tool cover -func=coverage.out | grep postgresql
```

---

## Phase 2: UI 修复

**目标**: 修复 SSL Mode 选项不匹配问题

---

#### Task 2.1: 修复 SSL Mode 下拉选项
**Type**: impl
**File**: `internal/transport/ui/pages/connection_page.go:296`
**Depends**: 无
**Estimate**: 10 分钟
**Description**: 更新 SSL Mode 下拉选项，使其与 PostgreSQL 规范一致
**Acceptance**:
- 选项包含：disable, allow, prefer, require, verify-ca, verify-full
- 默认值为 "prefer"
- UI 显示正常
- 不影响 MySQL 和其他数据库类型
**Implementation**:
在 `onAddConnection()` 方法中修改（约 line 296）：

```go
d.sslSelect = widget.NewSelect([]string{
	"disable",     // No SSL
	"allow",       // Try SSL, fallback to non-SSL
	"prefer",      // Try SSL first, fallback to non-SSL (default)
	"require",     // Force SSL, no certificate verification
	"verify-ca",   // Force SSL, verify CA certificate
	"verify-full", // Force SSL, verify CA and hostname
}, nil)
d.sslSelect.SetSelected("prefer")  // Set default
```

**Verification**:
```bash
go build ./cmd/db-benchmind
```

---

#### Task 2.2: 更新默认配置加载逻辑 [P]
**Type**: impl
**File**: `internal/domain/connection/default_config.go`
**Depends**: 无
**Estimate**: 10 分钟
**Description**: 确保默认配置中的 PostgreSQL SSL Mode 使用正确值 "prefer"
**Acceptance**:
- `DefaultPostgreSQLConfig` SSLMode = "prefer"
- 与 UI 默认值一致
**Implementation**:
检查 `internal/domain/connection/default_config.go`，确保：

```go
var DefaultPostgreSQLConfig = &PostgreSQLConnection{
    Host:     "localhost",
    Port:     5432,
    Database: "postgres",
    Username: "postgres",
    SSLMode:  "prefer",  // Ensure this is set
}
```

**Verification**:
```bash
go test -v ./internal/domain/connection -run TestDefaultConfig
```

---

## Phase 3: 集成测试

**目标**: 端到端验证 PostgreSQL 支持

---

#### Task 3.1: 手动 E2E 测试 - 连接管理
**Type**: test
**File**: 手动测试
**Depends**: 1.1, 1.2, 1.3, 2.1
**Estimate**: 30 分钟
**Description**: 手动测试 PostgreSQL 连接的完整 CRUD 流程
**Test Cases**:
1. ✅ 创建 PostgreSQL 连接
   - Name: Test PG
   - Host: localhost (or real PostgreSQL host)
   - Port: 5432
   - Database: postgres
   - Username: postgres
   - Password: ***
   - SSL Mode: prefer
2. ✅ 点击 "Test Connection"
   - 验证显示成功
   - 验证显示版本号
   - 验证延迟 < 5000ms
3. ✅ 点击 "Save"
   - 验证连接出现在列表中
   - 验证显示格式正确：`Test PG (postgres@localhost:5432/postgres)`
4. ✅ 编辑连接
   - 修改 Port
   - 重新测试
   - 保存
5. ✅ 删除连接
   - 验证从列表消失
6. ✅ 测试错误场景
   - 错误 Host: 验证显示错误信息
   - 错误 Port: 验证显示错误信息
**Acceptance**:
- 所有测试用例通过
- UI 响应正常
- 错误提示友好
**Verification**:
```bash
# 启动 GUI
./db-benchmind

# 手动执行上述测试用例
# 记录结果和截图
```

---

#### Task 3.2: 手动 E2E 测试 - 压测执行
**Type**: test
**File**: 手动测试
**Depends**: 3.1
**Estimate**: 40 分钟
**Description**: 测试完整的 Sysbench PostgreSQL 压测流程
**Test Cases**:
1. ✅ 配置压测任务
   - Connection: Test PG (from Task 3.1)
   - Tool: Sysbench
   - Template: OLTP Read-Write
   - Threads: 4
   - Time: 30 (短时间测试)
   - Tables: 4
   - Table Size: 10000
2. ✅ 执行 Prepare 阶段
   - 点击 "Prepare"
   - 验证数据库创建成功
3. ✅ 执行 Run 阶段
   - 点击 "Start"
   - 验证 Sysbench 命令包含 `--pgsql-*` 参数
   - 验证 `PGPASSWORD` 环境变量已设置
   - 等待完成
4. ✅ 查看结果
   - 进入 History 页面
   - 验证新记录显示
   - 验证 Database Type = "postgresql"
   - 验证指标正确：TPS, QPS, Latency
5. ✅ 执行 Cleanup 阶段
   - 点击 "Cleanup"
   - 验证数据库删除成功
**Acceptance**:
- 所有测试用例通过
- Sysbench 命令正确
- 结果正确解析和存储
**Verification**:
```bash
# 检查 Sysbench 命令日志
tail -f data/logs/db-benchmind-*.log | grep pgsql

# 检查结果存储
sqlite3 data/db-benchmind.db "SELECT * FROM runs WHERE id LIKE '%run%' ORDER BY created_at DESC LIMIT 5"
```

---

#### Task 3.3: 验证 Sysbench 命令生成 [P]
**Type**: test
**File**: `internal/infra/adapter/sysbench_adapter_test.go`
**Depends**: 无
**Estimate**: 20 分钟
**Description**: 添加单元测试验证 PostgreSQL Sysbench 命令生成正确
**Acceptance**:
- 测试 `BuildRunCommand()` 生成正确命令
- 测试 `BuildPrepareCommand()` 生成正确命令
- 测试 `BuildCleanupCommand()` 生成正确命令
- 测试环境变量包含 `PGPASSWORD`
**Implementation**:
在 `sysbench_adapter_test.go` 中添加：

```go
func TestSysbenchAdapter_PostgreSQLCommands(t *testing.T) {
	ctx := context.Background()
	adapter := NewSysbenchAdapter()

	conn := &connection.PostgreSQLConnection{
		BaseConnection: connection.BaseConnection{
			Name: "Test PG",
		},
		Host:     "localhost",
		Port:     5432,
		Database: "testdb",
		Username: "postgres",
		Password: "secret",
	}

	template := &domaintemplate.Template{
		ID:          "sysbench-oltp-read-write",
		Name:        "Sysbench OLTP Read-Write",
		Tool:        "sysbench",
		Script:      "/usr/share/sysbench/oltp_read_write.lua",
		CommandTemplate: domaintemplate.CommandTemplate{
			Run: "sysbench {script} --pgsql-host={host} --pgsql-port={port} --pgsql-user={user} --pgsql-password={password} --pgsql-db={database} --threads={threads} --time={time} run",
		},
	}

	params := map[string]any{
		"threads": 8,
		"time":    60,
		"tables":  10,
	}

	t.Run("BuildRunCommand", func(t *testing.T) {
		cmd, err := adapter.BuildRunCommand(ctx, conn, template, params)
		if err != nil {
			t.Fatalf("BuildRunCommand() failed: %v", err)
		}

		// Verify command
		if !strings.Contains(cmd.String(), "sysbench") {
			t.Error("Command should contain 'sysbench'")
		}
		if !strings.Contains(cmd.String(), "--pgsql-host=localhost") {
			t.Error("Command should contain --pgsql-host=localhost")
		}
		if !strings.Contains(cmd.String(), "--pgsql-port=5432") {
			t.Error("Command should contain --pgsql-port=5432")
		}
		if !strings.Contains(cmd.String(), "--pgsql-user=postgres") {
			t.Error("Command should contain --pgsql-user=postgres")
		}
		if !strings.Contains(cmd.String(), "--pgsql-db=testdb") {
			t.Error("Command should contain --pgsql-db=testdb")
		}
		// Password should NOT be in command, but in env
		if strings.Contains(cmd.String(), "--pgsql-password=secret") {
			t.Error("Password should not be in command string")
		}

		// Verify environment
		foundPassword := false
		for _, env := range cmd.Env {
			if env == "PGPASSWORD=secret" {
				foundPassword = true
			}
		}
		if !foundPassword {
			t.Error("PGPASSWORD should be set in environment")
		}
	})

	t.Run("BuildPrepareCommand", func(t *testing.T) {
		cmd, err := adapter.BuildPrepareCommand(ctx, conn, template, params)
		if err != nil {
			t.Fatalf("BuildPrepareCommand() failed: %v", err)
		}

		// Verify psql command for CREATE DATABASE
		if !strings.Contains(cmd.String(), "psql") {
			t.Error("Prepare command should contain 'psql'")
		}
		if !strings.Contains(cmd.String(), "-c") {
			t.Error("Prepare command should contain '-c' flag")
		}
		if !strings.Contains(cmd.String(), "CREATE DATABASE") {
			t.Error("Prepare command should contain 'CREATE DATABASE'")
		}
	})
}
```

**Verification**:
```bash
go test -v ./internal/infra/adapter -run TestSysbenchAdapter_PostgreSQLCommands
```

---

## Phase 4: 回归测试

**目标**: 确保 MySQL 功能不受影响

---

#### Task 4.1: MySQL 连接回归测试
**Type**: test
**File**: 手动测试
**Depends**: 3.1, 3.2
**Estimate**: 30 分钟
**Description**: 完整测试 MySQL 连接和压测功能，确保无回归
**Test Cases**:
1. ✅ 创建 MySQL 连接
2. ✅ 测试连接
3. ✅ 执行 Sysbench MySQL 压测
4. ✅ 查看结果
5. ✅ 验证所有功能正常
**Acceptance**:
- 所有 MySQL 功能正常
- 无性能退化
- 无 UI 问题
**Verification**:
```bash
# 通过 GUI 完整测试 MySQL 流程
# 对比执行时间，无显著增加
```

---

#### Task 4.2: 单元测试回归 [P]
**Type**: test
**File**: 所有单元测试
**Depends**: 1.4, 3.3
**Estimate**: 10 分钟
**Description**: 运行所有单元测试，确保无破坏性变更
**Acceptance**:
- 所有现有测试通过
- 无新增失败
**Verification**:
```bash
go test ./... -v
go test ./... -race
```

---

## Phase 5: 文档与提交

**目标**: 完成文档更新和代码提交

---

#### Task 5.1: 更新追溯文档 [P]
**Type**: impl
**File**: `specs/traceability.md` 或新建
**Depends**: 3.3
**Estimate**: 20 分钟
**Description**: 创建需求 → 测试 → 实现的追溯映射表
**Acceptance**:
- 所有需求映射到测试
- 所有测试映射到实现文件
- 格式清晰可读
**Implementation**:
创建 `specs/postgres-traceability.md`：

```markdown
# PostgreSQL 支持追溯性文档

## 需求 → 测试 → 实现映射

| 需求 ID | 需求描述 | 测试 | 实现文件 |
|---------|---------|------|---------|
| REQ-PG-CONN-001 | PostgreSQL 连接表单显示 | 手动测试 | connection_page.go |
| REQ-PG-CONN-002 | 默认端口 5432 | TestPostgreSQLConnection_Validate_ValidInput | postgresql.go |
| REQ-PG-CONN-003 | 字段验证 | TestPostgreSQLConnection_Validate_MissingRequiredFields | postgresql.go:Validate() |
| REQ-PG-CONN-004 | Database 可选 | TestPostgreSQLConnection_Validate_ValidInput | postgresql.go:Validate() |
| REQ-PG-CONN-005 | SSL Mode 选项 | TestPostgreSQLConnection_Validate_InvalidSSLMode | postgresql.go:Validate() |
| REQ-PG-CONN-010 | 连接测试成功 | TestPostgreSQLConnection_Test_Success | postgresql.go:Test() |
| REQ-PG-CONN-011 | 连接测试失败 | TestPostgreSQLConnection_Test_Failure | postgresql.go:Test() |
| REQ-PG-CONN-013 | 默认数据库 postgres | TestPostgreSQLConnection_GetDSN | postgresql.go:GetDSNWithPassword() |
| REQ-PG-SYS-001 | Sysbench pgsql 参数 | TestSysbenchAdapter_PostgreSQLCommands | sysbench_adapter.go:BuildRunCommand() |
| REQ-PG-SYS-003 | PGPASSWORD 环境变量 | TestSysbenchAdapter_PostgreSQLCommands | sysbench_adapter.go:BuildRunCommand() |
| REQ-PG-UI-001 | 连接表单字段 | 手动测试 | connection_page.go |
| REQ-PG-UI-002 | 自动设置端口 5432 | 手动测试 | connection_page.go |
```

**Verification**:
```bash
# 检查文档完整性
cat specs/postgres-traceability.md
```

---

#### Task 5.2: 代码质量检查 [P]
**Type**: test
**File**: 所有代码
**Depends**: 1.4, 2.1, 3.3, 4.2
**Estimate**: 10 分钟
**Description**: 运行所有质量门禁检查
**Acceptance**:
- `go build ./...` 成功
- `go test ./...` 全部通过
- `go test ./... -race` 无竞态
- `gofmt -l .` 无输出
- `go vet ./...` 无警告
- `golangci-lint run` 无错误
- `govulncheck ./...` 无漏洞
**Implementation**:
```bash
# Build check
go build ./...

# Test check
go test ./... -v

# Race check
go test ./... -race

# Format check
test -z "$(gofmt -l .)"

# Vet check
go vet ./...

# Lint check
golangci-lint run ./...

# Security check
govulncheck ./...
```

**Verification**:
所有检查通过，输出无错误

---

#### Task 5.3: 提交代码
**Type**: impl
**File**: Git
**Depends**: 5.1, 5.2
**Estimate**: 10 分钟
**Description**: 提交所有变更到 Git 仓库
**Acceptance**:
- Git commit 创建成功
- Commit message 符合 Conventional Commits 格式
- 包含 Co-Authored-By
- 无敏感信息提交
**Implementation**:
```bash
# Stage all changes
git add specs/postgres-*.md
git add internal/domain/connection/postgresql.go
git add internal/domain/connection/postgresql_test.go
git add internal/transport/ui/pages/connection_page.go
git add go.mod go.sum

# Commit
git commit -m "$(cat <<'EOF'
feat(postgres): add PostgreSQL connection support

- Add github.com/lib/pq PostgreSQL driver
- Implement PostgreSQLConnection.Test() method
- Add comprehensive unit tests for PostgreSQL connection
- Fix SSL Mode options to match PostgreSQL spec
- Update UI to support all SSL modes (disable, allow, prefer, require, verify-ca, verify-full)
- Add E2E test verification for PostgreSQL connections and benchmarks

Implements: REQ-PG-CONN-010, REQ-PG-CONN-011, REQ-PG-SYS-001, REQ-PG-UI-001
Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
EOF
)"
```

**Verification**:
```bash
git log -1 --stat
git show HEAD --stat
```

---

## 附录

### A. 任务依赖图

```
Phase 1: 驱动集成与测试实现
  1.1 [P] ──→ 1.2 [P] ──→ 1.3 ──→ 1.4

Phase 2: UI 修复
  2.1 [P]
  2.2 [P]

Phase 3: 集成测试
  3.1 ──→ 3.2
  3.3 [P]

Phase 4: 回归测试
  4.1 ──→ (depends on 3.1, 3.2)
  4.2 [P] ──→ (depends on 1.4, 3.3)

Phase 5: 文档与提交
  5.1 [P] ──→ (depends on 3.3)
  5.2 [P] ──→ (depends on 1.4, 2.1, 3.3, 4.2)
  5.3 ──→ (depends on 5.1, 5.2)
```

### B. 总时间估算

| Phase | 任务数 | 总时间 |
|-------|-------|--------|
| Phase 1 | 4 | 100 分钟 |
| Phase 2 | 2 | 20 分钟 |
| Phase 3 | 3 | 90 分钟 |
| Phase 4 | 2 | 40 分钟 |
| Phase 5 | 3 | 50 分钟 |
| **总计** | **14** | **300 分钟 (5 小时)**

### C. 并行执行机会

标记 `[P]` 的任务可以并行执行：
- **第一批**: 1.1, 1.2, 2.1, 2.2, 3.3, 4.2, 5.1, 5.2（依赖少，可并行）
- **第二批**: 1.3（等待 1.2）
- **第三批**: 1.4, 3.1, 3.2, 4.1（等待 1.3）
- **第四批**: 5.3（等待所有其他任务）

**理论最短时间**: 约 2-3 小时（高度并行）
**实际建议时间**: 4-5 小时（考虑串行和验证）

### D. 验收标准总结

**功能验收**:
- ✅ PostgreSQL 连接可以创建、编辑、删除
- ✅ PostgreSQL 连接测试成功/失败正确反馈
- ✅ Sysbench PostgreSQL 压测可以执行
- ✅ 结果正确解析和存储
- ✅ MySQL 功能无回归

**质量验收**:
- ✅ 所有单元测试通过
- ✅ 代码覆盖率 > 80%
- ✅ 无竞态条件
- ✅ 无安全漏洞
- ✅ 代码格式符合规范

**文档验收**:
- ✅ spec.md 完整
- ✅ plan.md 完整
- ✅ tasks.md 完整
- ✅ traceability.md 完整
- ✅ Git commit message 规范
