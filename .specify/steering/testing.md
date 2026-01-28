# Testing Strategy

## Testing Pyramid

```
                    /\
                   /  \
                  / E2E \         10% - 端到端测试
                 /--------\
                /          \
               / Integration\    30% - 集成测试
              /--------------\
             /                \
            /    Unit Tests     \  60% - 单元测试
           /--------------------\
```

### 目标覆盖率
| 层级 | 目标覆盖率 | 必须覆盖 |
|------|-----------|---------|
| domain/ | > 90% | 所有业务逻辑 |
| usecase/ | > 85% | 所有用例 |
| infra/database/ | > 80% | 所有仓储方法 |
| infra/adapter/ | > 75% | 命令构建、输出解析 |
| transport/ui/ | > 40% | 主要逻辑（手动为主） |

## Test Types

### 1. Unit Tests

**原则**：Test-First (constitution.md Article III)

**实现**：
- 表格驱动测试（Table-Driven Tests）
- 测试命名：`Test_<Func>_<Scenario>`
- 覆盖边界和错误路径

**示例**：
```go
func TestMySQLConnection_Validate(t *testing.T) {
    tests := []struct {
        name    string
        conn    *MySQLConnection
        wantErr bool
    }{
        {"valid", validConn, false},
        {"invalid port", invalidPortConn, true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.conn.Validate()
            if (err != nil) != tt.wantErr { ... }
        })
    }
}
```

### 2. Integration Tests

**原则**：Integration-First (constitution.md Article IX)

**策略**：
- 使用真实 SQLite（`:memory:`）
- 使用真实序列化/反序列化
- 使用临时目录（`t.TempDir()`）
- 外部服务使用 fake 或 skip（工具测试）

**Fake > Mock**：
- ✅ Fake: 内存实现、临时文件、httptest.Server
- ❌ Mock: 仅在必要时使用

**示例**：
```go
func TestSQLiteConnectionRepository(t *testing.T) {
    db, _ := sql.Open("sqlite", "file::memory:?mode=memory")
    // 执行 schema
    repo := NewSQLiteConnectionRepository(db)

    // 测试 Save 和 FindByID
    conn := &MySQLConnection{...}
    err := repo.Save(ctx, conn)
    // ...
}
```

### 3. E2E Tests

**场景**：端到端完整流程

**实现**：
- 使用真实的 SQLite
- 跳过需要外部工具的测试（或标记）
- 测试关键用户路径

**示例**：
```go
func TestE2E_ConnectionWorkflow(t *testing.T) {
    // 1. 初始化环境
    db := setupTestDB(t)
    repo := NewSQLiteConnectionRepository(db)
    keyring := NewMockKeyring()
    uc := NewConnectionUseCase(repo, keyring)

    // 2. 创建连接
    conn := NewMySQLConnection("Test", "localhost", "db", "root", 3306)
    conn.Password = "secret"
    err := uc.CreateConnection(ctx, conn)

    // 3. 验证保存
    saved, err := uc.GetConnectionByID(ctx, conn.ID)

    // 4. 测试连接（mock，不需要真实数据库）
    // ...
}
```

## Test Organization

```
test/
├── testdata/           # 测试数据
│   ├── connections/    # 测试用连接配置
│   ├── outputs/        # 工具输出示例
│   └── expected/       # 预期结果
└── integration/        # 集成测试
    ├── connection_test.go
    ├── template_test.go
    └── benchmark_test.go
```

## Test Execution

### Makefile Targets
```makefile
test:              # 运行所有测试
test-unit:         # 仅单元测试
test-integration:  # 仅集成测试
test-e2e:          # 仅 E2E 测试
```

### CI/CD
- PR 必须 `make check` 通过
- 包含：`format-check`, `test`, `lint`
- 必须运行的测试：`go test ./... -race`

## Test Data Management

### Golden Files
- 用于复杂输出验证
- 更新方式：`go test ./... -update-golden`
- 存放位置：`testdata/expected/`

### Fixtures
- 复杂测试数据使用 JSON 文件
- 避免硬编码在测试代码中
- 示例：`testdata/connections/mysql_valid.json`

## Tool-Specific Testing

### Sysbench Tests
- 跳过规则：`if !isSysbenchAvailable() { t.Skip() }`
- 输出示例：`testdata/outputs/sysbench_success.txt`

### Swingbench Tests
- 需要 Java 环境
- 跳过规则：同上

### HammerDB Tests
- 需要 TCL 环境
- 跳过规则：同上

## Coverage Requirements

### Minimum Coverage
```bash
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out | grep total
```

**目标**：总体覆盖率 > 75%

### Critical Paths
以下路径必须 100% 覆盖：
- Connection 验证逻辑
- 状态机转换
- 密码加密/解密
- SQL 参数化查询

## Testing Best Practices

### 1. Table-Driven Tests
```go
tests := []struct{...}{...}
for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) { ... })
}
```

### 2. Context with Timeout
```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
```

### 3. Clean Up
```go
defer func() {
    // 清理资源
}()
```

### 4. Error Messages
```go
if err != nil {
    t.Fatalf("Save() failed: %v", err)
}
```

## Continuous Testing

### Pre-Commit Hooks
```bash
#!/bin/bash
make format-check
go test ./...
go vet ./...
```

### Pre-Push Checklist
- [ ] All tests pass
- [ ] Coverage > 75%
- [ ] No lint errors
- [ ] Manual smoke test (GUI 启动)

## Testing Environments

### Local Development
- Go 1.22.2+
- Ubuntu 24 with GUI

### CI/CD
- Docker 容器
- 无 GUI 环境（跳过 GUI 测试）

### Production Testing
- 真实数据库环境
- 完整压测工具链
