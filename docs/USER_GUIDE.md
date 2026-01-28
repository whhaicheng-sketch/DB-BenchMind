# DB-BenchMind 用户手册

**版本**: 1.0.0
**更新日期**: 2026-01-28

---

## 目录

1. [快速开始](#快速开始)
2. [概念介绍](#概念介绍)
3. [连接管理](#连接管理)
4. [模板管理](#模板管理)
5. [运行基准测试](#运行基准测试)
6. [查看结果](#查看结果)
7. [生成报告](#生成报告)
8. [结果对比](#结果对比)
9. [系统设置](#系统设置)
10. [常见问题](#常见问题)

---

## 快速开始

### 安装

#### 从源码编译

```bash
# 克隆仓库
git clone https://github.com/whhaicheng/DB-BenchMind.git
cd DB-BenchMind

# 编译 CLI 版本
go build -o build/db-benchmind-cli ./cmd/db-benchmind-cli/

# 编译 GUI 版本（需要 GUI 依赖）
go build -o build/db-benchmind ./cmd/db-benchmind
```

#### 下载预编译版本

访问 [Releases](https://github.com/whhaicheng/DB-BenchMind/releases) 下载适合您平台的二进制文件。

### 首次运行

#### CLI 版本

```bash
# 查看版本
./build/db-benchmind-cli version

# 检测已安装的基准测试工具
./build/db-benchmind-cli detect

# 查看数据库连接
./build/db-benchmind-cli list
```

#### GUI 版本

```bash
# 启动 GUI
./build/db-benchmind
```

---

## 概念介绍

### 核心概念

#### 连接 (Connection)

数据库连接包含了连接到特定数据库所需的所有信息：

- **连接类型**: MySQL, Oracle, SQL Server, PostgreSQL
- **连接参数**: 主机、端口、用户名、密码、数据库名
- **安全存储**: 密码使用系统 keyring 加密存储

#### 模板 (Template)

模板定义了基准测试的配置：

- **测试工具**: Sysbench, Swingbench, HammerDB
- **测试类型**: OLTP 读写、只读、只写等
- **参数配置**: 线程数、时长、表大小等
- **内置模板**: 7 个常用场景的预配置模板

#### 任务 (Task)

任务是连接和模板的组合：

- 绑定特定的数据库连接
- 使用特定的测试模板
- 可以保存重复运行
- 支持自定义参数覆盖

#### 运行 (Run)

运行是一次实际的基准测试执行：

- 记录完整的状态变化
- 实时采集性能指标
- 保存完整的日志输出
- 生成结构化结果

### 支持的工具

#### Sysbench

- **支持数据库**: MySQL, PostgreSQL
- **测试场景**: OLTP 读写、只读、只写、非索引写入
- **输出指标**: TPS, 延迟, 百分位数, 错误率

#### Swingbench

- **支持数据库**: Oracle
- **测试场景**: SOE (Sales Order Entry), Calling Circle
- **输出指标**: TPM, 延迟, 错误率

#### HammerDB

- **支持数据库**: MySQL, Oracle, SQL Server, PostgreSQL
- **测试场景**: OLTP 读写、只读
- **输出指标**: NOPM, TPM, 延迟

---

## 连接管理

### 添加连接（程序化）

当前 CLI 版本需要通过 API 添加连接。以下是示例代码：

```go
package main

import (
    "context"
    "time"
    "github.com/whhaicheng/DB-BenchMind/internal/domain/connection"
    "github.com/whhaicheng/DB-BenchMind/internal/app/usecase"
    "github.com/whhaicheng/DB-BenchMind/internal/infra/database"
    "github.com/whhaicheng/DB-BenchMind/internal/infra/database/repository"
    "github.com/whhaicheng/DB-BenchMind/internal/infra/keyring"
)

func main() {
    ctx := context.Background()

    // 初始化数据库
    db, err := database.InitializeSQLite(ctx, "./data/db-benchmind.db")
    if err != nil {
        panic(err)
    }
    defer db.Close()

    // 初始化 repository 和 use case
    connRepo := repository.NewSQLiteConnectionRepository(db)
    keyringProvider, err := keyring.NewFileFallback("./data", "")
    if err != nil {
        panic(err)
    }
    connUC := usecase.NewConnectionUseCase(connRepo, keyringProvider)

    // 创建 MySQL 连接
    mysqlConn := &connection.MySQLConnection{
        ID:       "prod-mysql-01",
        Name:     "Production MySQL",
        Host:     "192.168.1.100",
        Port:     3306,
        Database: "sbtest",
        Username: "bench_user",
        Password: "secure_password",
        SSLMode:  "disabled",
    }

    // 保存连接
    err = connUC.CreateConnection(ctx, mysqlConn)
    if err != nil {
        panic(err)
    }

    // 测试连接
    result, err := connUC.TestConnection(ctx, mysqlConn.GetID())
    if err != nil {
        panic(err)
    }

    if result.Success {
        println("连接成功！延迟:", result.LatencyMs, "ms")
        println("数据库版本:", result.DatabaseVersion)
    }
}
```

### 连接类型

#### MySQL 连接

```go
conn := &connection.MySQLConnection{
    Host:     "localhost",
    Port:     3306,
    Database: "testdb",
    Username: "root",
    Password: "password",
    SSLMode:  "disabled", // or "required", "preferred"
}
```

#### Oracle 连接

```go
conn := &connection.OracleConnection{
    Host:         "localhost",
    Port:         1521,
    SID:          "ORCL",          // 使用 SID
    ServiceName:  "",              // 或使用 Service Name
    Username:     "system",
    Password:     "password",
}
```

#### SQL Server 连接

```go
conn := &connection.SQLServerConnection{
    Host:     "localhost",
    Port:     1433,
    Database: "master",
    Username: "sa",
    Password: "password",
}
```

#### PostgreSQL 连接

```go
conn := &connection.PostgreSQLConnection{
    Host:         "localhost",
    Port:         5432,
    Database:     "postgres",
    Username:     "postgres",
    Password:     "password",
    SSLMode:      "disable", // or "require", "verify-ca", "verify-full"
}
```

### 查看连接列表

```bash
./build/db-benchmind-cli list
```

输出示例：

```
Found 2 connection(s):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

[1] Production MySQL
    ID:   prod-mysql-01
    Type: mysql
    Host: 192.168.1.100:3306/sbtest

[2] Test Oracle
    ID:   test-ora-01
    Type: oracle
    Host: 192.168.1.101:1521:ORCL

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

---

## 模板管理

### 内置模板

系统提供 7 个预配置的内置模板：

| ID | 名称 | 工具 | 数据库 | 类型 | 描述 |
|----|------|------|--------|------|------|
| `sysbench-oltp-mixed` | Sysbench OLTP 混合 | Sysbench | MySQL, PostgreSQL | oltp_read_write | 读写混合场景 |
| `sysbench-oltp-read` | Sysbench OLTP 只读 | Sysbench | MySQL, PostgreSQL | oltp_read_only | 只读查询 |
| `sysbench-oltp-write` | Sysbench OLTP 只写 | Sysbench | MySQL, PostgreSQL | oltp_write_only | 只写操作 |
| `swingbench-soe` | Swingbench SOE | Swingbench | Oracle | soe | 销售订单录入 |
| `swingbench-calling` | Swingbench Calling | Swingbench | Oracle | calling | 呼叫中心模拟 |
| `hammerdb-tpcc` | HammerDB TPCC | HammerDB | MySQL, Oracle, SQL Server, PostgreSQL | tpcc | TPCC 基准 |
| `hammerdb-tpc-h` | HammerDB TPC-H | HammerDB | MySQL, Oracle, SQL Server, PostgreSQL | tpch | 决策支持查询 |

### 查看可用模板（通过 API）

```go
import "github.com/whhaicheng/DB-BenchMind/internal/app/usecase"

// 初始化
templateUC := usecase.NewTemplateUseCase(templateRepo)

// 获取所有模板
templates, err := templateUC.ListTemplates(ctx)

// 获取内置模板
builtinTemplates, err := templateUC.ListBuiltinTemplates(ctx)

// 获取自定义模板
customTemplates, err := templateUC.ListCustomTemplates(ctx)
```

### 创建自定义模板

```go
import "github.com/whhaicheng/DB-BenchMind/internal/domain/template"

customTemplate := &template.Template{
    ID:          "custom-stress-test",
    Name:        "自定义压力测试",
    Tool:        "sysbench",
    DatabaseTypes: []template.DatabaseType{
        template.DatabaseTypeMySQL,
    },
    BenchmarkType: "oltp_read_write",
    Parameters: map[string]interface{}{
        "threads":      64,
        "time":         600,
        "table_size":   10000000,
        "tables":       32,
    },
    Options: map[string]interface{}{
        "report_interval": 10,
        "forced_shutdown": "off",
    },
}

err := templateUC.CreateTemplate(ctx, customTemplate)
```

---

## 运行基准测试

### 基本流程（通过 API）

```go
import "github.com/whhaicheng/DB-BenchMind/internal/app/usecase"

// 1. 初始化
benchmarkUC := usecase.NewBenchmarkUseCase(
    runRepo,
    adapterRegistry,
    keyringProvider,
)

// 2. 创建任务
task := &execution.Task{
    ID:           "task-001",
    Name:         "生产环境压测",
    ConnectionID: "prod-mysql-01",
    TemplateID:   "sysbench-oltp-mixed",
    Parameters: map[string]interface{}{
        "threads": 16,
        "time":    300,
    },
}

// 3. 执行测试
run, err := benchmarkUC.ExecuteTask(ctx, task)
if err != nil {
    panic(err)
}

// 4. 监控运行状态
for {
    status := benchmarkUC.GetRunStatus(ctx, run.ID)
    println("状态:", status.State)
    println("进度:", status.Progress, "%")

    if status.State == execution.StateCompleted {
        break
    }

    time.Sleep(5 * time.Second)
}

// 5. 获取结果
result, err := benchmarkUC.GetRunResult(ctx, run.ID)
println("TPS:", result.TPSCalculated)
println("平均延迟:", result.LatencyAvg, "ms")
println("P95 延迟:", result.LatencyP95, "ms")
```

### 执行参数

#### Sysbench 参数

| 参数 | 类型 | 默认值 | 描述 |
|------|------|--------|------|
| `threads` | int | 1 | 并发线程数 |
| `time` | int | 10 | 测试时长（秒） |
| `table_size` | int | 10000 | 每张表的行数 |
| `tables` | int | 1 | 表数量 |
| `rate` | int | 0 | 事务速率限制（0=无限制） |

#### Swingbench 参数

| 参数 | 类型 | 默认值 | 描述 |
|------|------|--------|------|
| `threads` | int | 1 | 并发用户数 |
| `time` | int | 60 | 测试时长（秒） |
| `benchmark` | string | "soe" | 测试类型 (soe/calling) |

#### HammerDB 参数

| 参数 | 类型 | 默认值 | 描述 |
|------|------|--------|------|
| `threads` | int | 1 | 虚拟用户数 |
| `time` | int | 60 | 测试时长（分钟） |
| `warehouses` | int | 1 | 仓库数量（TPCC） |
| `scale_factor` | int | 1 | 扩展因子（TPC-H） |

---

## 查看结果

### 获取历史运行（通过 API）

```go
// 获取所有运行
runs, err := runRepo.FindAll(ctx, repository.FindOptions{
    Limit:    50,
    SortBy:   "created_at",
    SortOrder: "DESC",
})

// 按连接筛选
mysqlRuns, err := runRepo.FindByConnection(ctx, "prod-mysql-01", repository.FindOptions{
    Limit: 20,
})

// 按状态筛选
completedRuns, err := runRepo.FindByState(ctx, execution.StateCompleted, repository.FindOptions{
    Limit: 100,
})
```

### 查看运行详情

```go
run, err := runRepo.FindByID(ctx, "run-001")

// 基本信息
println("运行 ID:", run.ID)
println("任务名称:", run.TaskID)
println("状态:", run.State)
println("创建时间:", run.CreatedAt)

// 执行信息
if run.StartedAt != nil {
    println("开始时间:", *run.StartedAt)
}
if run.CompletedAt != nil {
    println("完成时间:", *run.CompletedAt)
}
println("耗时:", run.DurationSeconds, "秒")

// 结果
if run.Result != nil {
    println("TPS:", run.Result.TPSCalculated)
    println("平均延迟:", run.Result.LatencyAvg, "ms")
    println("P95 延迟:", run.Result.LatencyP95, "ms")
    println("P99 延迟:", run.Result.LatencyP99, "ms")
    println("错误率:", run.Result.ErrorRate, "%")
}

// 错误信息
if run.ErrorMessage != "" {
    println("错误:", run.ErrorMessage)
}
```

---

## 生成报告

### 报告格式

支持以下报告格式：

- **Markdown**: .md 文件，易于编辑和版本控制
- **HTML**: .html 文件，适合浏览器查看
- **JSON**: .json 文件，便于程序处理
- **PDF**: .pdf 文件，适合打印和分享（需要额外工具）

### 生成报告（通过 API）

```go
import "github.com/whhaicheng/DB-BenchMind/internal/app/usecase"

// 初始化报告用例
reportUC := usecase.NewReportUseCase(reportRepo, runRepo, generatorRegistry)

// 生成 Markdown 报告
err := reportUC.GenerateReport(ctx, "run-001", report.FormatMarkdown)
// 输出: results/run-001-report.md

// 生成 HTML 报告
err := reportUC.GenerateReport(ctx, "run-001", report.FormatHTML)
// 输出: results/run-001-report.html

// 生成 JSON 报告
err := reportUC.GenerateReport(ctx, "run-001", report.FormatJSON)
// 输出: results/run-001-report.json

// 生成 PDF 报告（需要 pandoc）
err := reportUC.GenerateReport(ctx, "run-001", report.FormatPDF)
// 输出: results/run-001-report.pdf
```

### 报告内容

报告包含以下部分：

1. **概要信息**
   - 运行 ID、名称
   - 数据库类型和连接信息
   - 测试工具和类型
   - 执行时间

2. **测试配置**
   - 使用的模板
   - 自定义参数
   - 执行选项

3. **性能指标**
   - TPS/QPS
   - 延迟统计（平均、P95、P99）
   - 错误率
   - 时间序列图表

4. **原始输出**
   - 完整的工具输出日志
   - 解析后的指标数据

---

## 结果对比

### 对比类型

#### 基线对比 (Baseline Comparison)

将多次运行与指定的基线运行进行对比：

```go
import "github.com/whhaicheng/DB-BenchMind/internal/domain/comparison"

compUC := usecase.NewComparisonUseCase(runRepo)

// 对比多次运行
result, err := compUC.CompareRuns(ctx, []string{
    "run-002",
    "run-003",
    "run-004",
}, "run-001", comparison.ComparisonTypeBaseline)

// 查看差异
for _, diff := range result.Metrics.TPSDiff {
    println(diff.RunName)
    println("  TPS:", diff.Value)
    println("  差异:", diff.Diff)
    println("  变化率:", diff.DiffPct, "%")
}
```

#### 趋势分析 (Trend Analysis)

分析特定指标在多次运行中的趋势：

```go
// 分析 TPS 趋势
analysis, err := compUC.GetTrendAnalysis(ctx, []string{
    "run-001",
    "run-002",
    "run-003",
    "run-004",
}, "tps")

println("趋势:", analysis.Trend)  // "increasing", "decreasing", "stable"
println("变化率:", analysis.ChangePct, "%")
println("最小值:", analysis.MinValue)
println("最大值:", analysis.MaxValue)
println("平均值:", analysis.AvgValue)
```

### 对比摘要

对比结果包含摘要信息：

```go
summary := result.Summary
println("总运行数:", summary.TotalRuns)
println("基线 ID:", summary.BaselineRunID)
println("TPS 趋势:", summary.OverallTpsTrend)
println("TPS 变化:", summary.TpsChangePct, "%")

// 最佳运行
if summary.BestRun != nil {
    println("最佳运行:", summary.BestRun.RunName)
    println("  原因:", summary.BestRun.Reason)
    println("  TPS:", summary.BestRun.TPS)
}

// 洞察
for _, insight := range summary.Insights {
    println("💡", insight)
}
```

---

## 系统设置

### 工具检测

系统会自动检测已安装的基准测试工具：

```bash
./build/db-benchmind-cli detect
```

输出示例：

```
Detecting benchmark tools...
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

✓ sysbench
  Path:    /usr/bin/sysbench
  Version: 1.0.20

✗ swingbench (not found)
✗ hammerdb (not found)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Tip: To install tools:
  Sysbench:   apt-get install sysbench
  Swingbench: Download from https://www.swingbench.com
  HammerDB:   Download from https://www.hammerdb.com
```

### 配置文件

数据存储位置：

```
./data/db-benchmind.db     # SQLite 数据库
./data/*.key              # 密钥存储（文件降级方案）
./results/                # 报告输出目录
```

### 环境变量

| 变量 | 描述 | 默认值 |
|------|------|--------|
| `DB_BENCHMIND_DB_PATH` | 数据库文件路径 | `./data/db-benchmind.db` |
| `DB_BENCHMIND_KEY_DIR` | 密钥存储目录 | `./data` |
| `DB_BENCHMIND_RESULTS_DIR` | 报告输出目录 | `./results` |

---

## 常见问题

### Q1: 如何安装 Sysbench？

**Ubuntu/Debian**:
```bash
sudo apt-get update
sudo apt-get install sysbench
```

**macOS**:
```bash
brew install sysbench
```

**验证安装**:
```bash
sysbench --version
```

### Q2: 如何安装 Swingbench？

Swingbench 需要手动下载：

1. 访问 https://www.swingbench.com
2. 下载最新版本的 zip 文件
3. 解压到任意目录
4. 系统会自动检测 `swingbench.jar`

### Q3: 如何安装 HammerDB？

1. 访问 https://www.hammerdb.com
2. 下载适合您平台的安装包
3. 安装 HammerDB
4. 系统会自动检测 `hammerdbcli` 或 `hammerdb.bat`

### Q4: 密码存储安全吗？

系统使用多层安全机制：

1. **优先使用系统 keyring**: gnome-keyring, macOS Keychain, Windows Credential Manager
2. **文件降级方案**: 如果系统 keyring 不可用，使用加密文件存储
3. **密码不在 JSON 中序列化**: 数据库中不存储明文密码

### Q5: 如何备份和恢复数据？

**备份数据库**:
```bash
# 备份 SQLite 数据库
cp ./data/db-benchmind.db ./backup/db-benchmind-$(date +%Y%m%d).db

# 备份密钥文件
cp ./data/*.key ./backup/
```

**恢复数据**:
```bash
# 停止程序
# 恢复数据库
cp ./backup/db-benchmind-20260128.db ./data/db-benchmind.db
# 恢复密钥文件
cp ./backup/*.key ./data/
# 重启程序
```

### Q6: 测试时出现 "connection refused" 错误

检查清单：

1. 数据库服务是否正在运行
2. 主机和端口配置是否正确
3. 防火墙是否允许连接
4. 数据库用户权限是否足够
5. 使用 `TestConnection` 功能验证连接

### Q7: 如何提高测试准确性？

建议：

1. **预热**: 先运行一段时间预热数据库
2. **多次测试**: 运行 3-5 次取平均值
3. **稳定环境**: 确保没有其他负载
4. **足够时长**: 测试时长建议 ≥ 5 分钟
5. **合理并发**: 线程数不超过 CPU 核心数的 2 倍

### Q8: 报告生成失败怎么办？

**Markdown/HTML/JSON**: 这些格式应该始终能生成。

**PDF 生成**: 需要额外工具：
- `pandoc`: Markdown → PDF
- `wkhtmltopdf`: HTML → PDF

安装 pandoc（Ubuntu）:
```bash
sudo apt-get install pandoc
```

如果 PDF 生成失败，可以使用 Markdown 或 HTML 格式，然后手动转换。

### Q9: 数据库连接失败

常见原因：

1. **网络问题**: 使用 `ping` 和 `telnet` 测试连通性
2. **认证失败**: 检查用户名和密码
3. **权限不足**: 确保用户有测试数据库的权限
4. **SSL 问题**: 调整 `SSLMode` 参数

### Q10: 如何对比不同版本的数据库性能？

使用结果对比功能：

1. 对版本 A 运行测试，记录 `run-id-v1`
2. 升级数据库到版本 B
3. 使用相同配置运行测试，记录 `run-id-v2`
4. 对比两次运行：
   ```bash
   # 通过 API 或未来 CLI 命令
   CompareRuns(run-id-v1, run-id-v2)
   ```

---

## 获取帮助

- **GitHub Issues**: https://github.com/whhaicheng/DB-BenchMind/issues
- **文档**: https://github.com/whhaicheng/DB-BenchMind/tree/main/docs
- **示例代码**: `test/integration/`

---

**版本 1.0.0 - 完**
