# DB-BenchMind 对比功能指南

**版本**: 1.0.0
**更新日期**: 2026-01-30

---

## 目录

1. [功能概述](#功能概述)
2. [GUI 使用指南](#gui-使用指南)
3. [API 使用指南](#api-使用指南)
4. [对比场景](#对比场景)
5. [结果解读](#结果解读)
6. [最佳实践](#最佳实践)

---

## 功能概述

DB-BenchMind 的对比功能允许您横向对比多个历史基准测试记录，快速识别性能差异。

### 核心特性

- **多配置对比**: 一次性对比 2-5 条历史记录
- **灵活分组**: 按线程数、数据库类型、模板名称或日期分组
- **丰富指标**: TPS、延迟（min/avg/max/P95/P99）、QPS、读写比例
- **可视化展示**: ASCII 表格和柱状图
- **搜索过滤**: 快速筛选目标记录
- **多格式导出**: 支持 TXT、Markdown、CSV 格式（即将推出）

### 应用场景

1. **性能调优**: 对比不同线程数、不同配置下的性能表现
2. **数据库对比**: 对比 MySQL、PostgreSQL、Oracle 等不同数据库的性能
3. **版本对比**: 对比数据库升级前后的性能差异
4. **回归测试**: 确保新版本没有引入性能回归
5. **容量规划**: 评估不同配置下的性能上限

---

## GUI 使用指南

### 启动应用

```bash
cd /path/to/DB-BenchMind
./bin/db-benchmind gui
```

### 界面布局

```
┌─────────────────────────────────────────────────────────────────┐
│  DB-BenchMind                                                  │
├─────────────────────────────────────────────────────────────────┤
│  Connections │ Templates │ Tasks & Monitor │ History │         │
│  Comparison │ Reports │ Settings                                   │
├─────────────────────────────────────────────────────────────────┤
│  Configuration & Selection                                      │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │ Filter by:                                             │   │
│  │ [Search...]                    Group By: [Threads ▼]    │   │
│  └─────────────────────────────────────────────────────────┘   │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │ ☑ MySQL | Sysbench OLTP | 4 threads | 1250 TPS      │   │
│  │ ☐ MySQL | Sysbench OLTP | 8 threads | 2100 TPS      │   │
│  │ ☑ PostgreSQL | Sysbench OLTP | 8 threads | 1980 TPS  │   │
│  │ ...                                                    │   │
│  └─────────────────────────────────────────────────────────┘   │
│  [🔄 Refresh] [📊 Compare Selected] [💾 Export Report] [🗑️ Clear] │
│  Comparison Results:                                          │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │ ╔════════════════════════════════════════════════════╗ │   │
│  │ ║       Multi-Configuration Comparison Results       ║ │   │
│  │ ╠════════════════════════════════════════════════════╣ │   │
│  │ ║ Summary...                                         ║ │   │
│  │ ╚════════════════════════════════════════════════════╝ │   │
│  └─────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

### 基本操作

#### 1. 选择记录

在记录列表中，勾选您想要对比的记录（2-5 条）。

每条记录显示：
- 数据库类型
- 模板名称
- 线程数
- TPS（每秒事务数）
- QPS（每秒查询数）
- 开始时间

**记录数量限制说明：**

- **最少选择**: 2 条记录（必须有对比意义）
- **最多选择**: 5 条记录（严格限制）
- **推荐范围**: 2-5 条记录

**为什么限制为 5 条？**

1. **可读性**: 超过 5 条记录时，ASCII 表格和柱状图会变得很长，难以阅读和比较
2. **性能**: 5 条记录的处理时间 < 1 秒，确保良好的用户体验
3. **实用性**: 5 条记录已覆盖绝大多数对比场景：
   - 线程数对比：1, 2, 4, 8, 16 线程
   - 数据库对比：MySQL, PostgreSQL, Oracle, SQL Server + 1 个基准
   - A/B/C/D/E 测试

**如果需要对比更多记录？**

建议采用**分步对比**策略：
1. 先对比 2-3 个关键配置，找出最优
2. 再用最优配置与其他配置对比
3. 逐步缩小范围，最终确定最佳配置

例如：
- 第一轮：对比 1, 4, 8, 16 线程（找出最佳线程数）
- 第二轮：用最佳线程数对比不同数据库类型
- 第三轮：对比不同参数配置

#### 2. 选择分组方式

从 "Group By" 下拉菜单中选择分组方式：

| 分组方式 | 说明 | 适用场景 |
|---------|------|----------|
| **Threads** | 按线程数升序排列 | 对比不同并发数的性能 |
| **Database Type** | 按数据库类型分组 | 对比不同数据库的性能 |
| **Template Name** | 按模板名称分组 | 对比不同测试类型的性能 |
| **Date** | 按开始时间排列 | 对比历史趋势 |

#### 3. 执行对比

点击 "📊 Compare Selected" 按钮，系统将：
1. 验证选择的记录数（2-5 条）
2. 按照选择的分组方式排序
3. 计算各项指标的统计值
4. 生成对比表格和图表

#### 4. 查看结果

结果包含以下部分：

**摘要信息**
```
## Summary

Total Records: 3
Group By: threads
```

**TPS 对比表格**
```
## TPS Comparison (Transactions Per Second)

┌─────────────────────────────────────────────────────────────────┐
│ Configuration                                                  │
├─────────────────────────────────────────────────────────────────┤
│ Config              │       Min │       Avg │       Max │   StdDev │
├─────────────────────────────────────────────────────────────────┤
│ MySQL (4 threads)   │   1250.50 │   2100.30 │   3500.80 │   1125.15 │
│ MySQL (8 threads)   │   1980.20 │   2100.30 │   3500.80 │    760.30 │
│ MySQL (16 threads)  │   3500.80 │   3500.80 │   3500.80 │      0.00 │
└─────────────────────────────────────────────────────────────────┘
```

**延迟对比表格**
```
## Latency Comparison (ms)

┌─────────────────────────────────────────────────────────────────┐
│ Configuration (Avg Latency)                                    │
├─────────────────────────────────────────────────────────────────┤
│ Config              │       Min │       Avg │       Max │   StdDev │
├─────────────────────────────────────────────────────────────────┤
│ MySQL (4 threads)   │      6.80 │      8.20 │      9.10 │      1.15 │
│ MySQL (8 threads)   │      6.80 │      7.85 │      9.10 │      1.15 │
│ MySQL (16 threads)  │      6.80 │      6.80 │      6.80 │      0.00 │
└─────────────────────────────────────────────────────────────────┘
```

**TPS 柱状图**
```
## TPS Bar Chart

MySQL (4 threads)   │██████████████████████████████████████████████████ 1250.50
MySQL (8 threads)   │█████████████████████████████████████████████████████████████████████████ 2100.30
MySQL (16 threads)  │████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████ 3500.80
```

**查询分布**
```
## Query Distribution

  Read:  30048 queries (66.7%)
  Write: 15024 queries (33.3%)
```

#### 5. 搜索过滤

使用搜索框快速过滤记录：

- 输入 "MySQL" - 只显示 MySQL 相关记录
- 输入 "8 threads" - 只显示 8 线程的记录
- 输入 "oltp" - 只显示 OLTP 相关的记录

搜索支持：数据库类型、模板名称、连接名、线程数的模糊匹配。

#### 6. 导出报告

点击 "💾 Export Report" 按钮，选择导出格式：
- **TXT**: 纯文本格式
- **Markdown**: Markdown 格式（推荐）
- **CSV**: CSV 格式（便于 Excel 分析）

---

## API 使用指南

### 初始化

```go
import (
    "context"
    "github.com/whhaicheng/DB-BenchMind/internal/app/usecase"
    "github.com/whhaicheng/DB-BenchMind/internal/infra/database/repository"
)

// 初始化 ComparisonUseCase
comparisonUC := usecase.NewComparisonUseCase(historyRepo)
```

### 获取记录列表

```go
// 获取所有记录
records, err := comparisonUC.GetAllRecords(ctx)
if err != nil {
    log.Fatal(err)
}

// 获取记录引用（轻量级）
refs, err := comparisonUC.GetRecordRefs(ctx)
if err != nil {
    log.Fatal(err)
}

// 遍历记录引用
for _, ref := range refs {
    fmt.Printf("ID: %s, DB: %s, Threads: %d, TPS: %.2f\n",
        ref.ID, ref.DatabaseType, ref.Threads, ref.TPS)
}
```

### 执行对比

```go
import "github.com/whhaicheng/DB-BenchMind/internal/domain/comparison"

// 选择要对比的记录 ID
recordIDs := []string{"hist-001", "hist-002", "hist-003"}

// 选择分组方式
groupBy := comparison.GroupByThreads

// 执行对比
result, err := comparisonUC.CompareRecords(ctx, recordIDs, groupBy)
if err != nil {
    log.Fatal(err)
}

// 查看结果
fmt.Printf("对比 ID: %s\n", result.ID)
fmt.Printf("记录数: %d\n", len(result.Records))
fmt.Printf("分组方式: %s\n", result.GroupBy)
```

### 访问统计指标

```go
// TPS 统计
if result.TPSComparison != nil {
    fmt.Printf("TPS 平均值: %.2f\n", result.TPSComparison.Avg)
    fmt.Printf("TPS 标准差: %.2f\n", result.TPSComparison.StdDev)
    fmt.Printf("TPS 最小值: %.2f\n", result.TPSComparison.Min)
    fmt.Printf("TPS 最大值: %.2f\n", result.TPSComparison.Max)
}

// 延迟统计
if result.LatencyCompare != nil {
    if result.LatencyCompare.Avg != nil {
        fmt.Printf("平均延迟: %.2f ms\n", result.LatencyCompare.Avg.Avg)
    }
    if result.LatencyCompare.P95 != nil {
        fmt.Printf("P95 延迟: %.2f ms\n", result.LatencyCompare.P95.Max)
    }
    if result.LatencyCompare.P99 != nil {
        fmt.Printf("P99 延迟: %.2f ms\n", result.LatencyCompare.P99.Max)
    }
}

// QPS 统计
if result.QPSComparison != nil {
    fmt.Printf("QPS 平均值: %.2f\n", result.QPSComparison.Avg)
}

// 读写比例
if result.ReadWriteRatio != nil {
    fmt.Printf("读查询: %d (%.1f%%)\n",
        result.ReadWriteRatio.ReadQueries,
        result.ReadWriteRatio.ReadPct)
    fmt.Printf("写查询: %d (%.1f%%)\n",
        result.ReadWriteRatio.WriteQueries,
        result.ReadWriteRatio.WritePct)
}
```

### 格式化输出

```go
// 生成表格
table := result.FormatTable()
fmt.Println(table)

// 生成 TPS 柱状图
tpsChart := result.FormatBarChart("TPS")
fmt.Println(tpsChart)

// 生成 QPS 柱状图
qpsChart := result.FormatBarChart("QPS")
fmt.Println(qpsChart)
```

### 过滤记录

```go
import "github.com/whhaicheng/DB-BenchMind/internal/app/usecase"

// 创建过滤器
filter := &usecase.ComparisonFilter{
    DatabaseType: "MySQL",
    MinThreads:   4,
    MaxThreads:   16,
}

// 应用过滤
filteredRefs := comparisonUC.FilterRecords(ctx, refs, filter)

// 遍历过滤结果
for _, ref := range filteredRefs {
    fmt.Printf("Filtered: %s (%d threads)\n", ref.DatabaseType, ref.Threads)
}
```

---

## 对比场景

### 场景 1: 线程数对比

**目标**: 找到最佳线程数配置

**步骤**:
1. 使用相同配置、不同线程数（4, 8, 16, 32）运行测试
2. 选择这些记录，按 "Threads" 分组
3. 查看 TPS 和 P99 延迟的变化趋势
4. 选择 TPS 高且延迟稳定的配置

**预期结果**:
```
Threads │  TPS    │ P99 Latency │ 结论
--------│---------│-------------│--------
4       │ 1250    │ 12.5 ms     │ 基准
8       │ 2100    │ 15.2 ms     │ ✓ 最佳
16      │ 3500    │ 28.1 ms     │ 延迟偏高
32      │ 3200    │ 45.3 ms     │ 过载
```

### 场景 2: 数据库类型对比

**目标**: 对比不同数据库的性能

**步骤**:
1. 在 MySQL、PostgreSQL、Oracle 上运行相同测试
2. 选择这些记录，按 "Database Type" 分组
3. 对比 TPS、延迟、资源消耗

**预期结果**:
```
Database     │ TPS    │ Avg Latency │ 结论
-------------│---------│-------------│--------
MySQL        │ 2100   │ 7.2 ms      │ 速度快
PostgreSQL   │ 1980   │ 9.1 ms      │ 稳定性好
Oracle       │ 2300   │ 6.8 ms      │ 性能最佳
```

### 场景 3: 性能回归测试

**目标**: 确保升级没有引入性能回归

**步骤**:
1. 升级前运行测试，记录 `before-upgrade`
2. 升级后运行相同测试，记录 `after-upgrade`
3. 对比这两条记录
4. 关注 TPS 变化和 P99 延迟

**判断标准**:
- TPS 下降 < 5%: 无回归
- TPS 下降 5-10%: 轻微回归，需要关注
- TPS 下降 > 10%: 严重回归，需要调查

### 场景 4: 配置优化对比

**目标**: 评估数据库配置优化的效果

**步骤**:
1. 使用默认配置运行测试
2. 优化配置（如 buffer pool、并发连接数）
3. 使用优化后的配置运行测试
4. 对比两次测试结果

### 场景 5: 硬件升级对比

**目标**: 评估硬件升级的性价比

**步骤**:
1. 在旧硬件上运行测试
2. 升级硬件（CPU、内存、磁盘）
3. 在新硬件上运行相同测试
4. 对比性能提升幅度

---

## 结果解读

### TPS (Transactions Per Second)

**定义**: 每秒完成的事务数

**解读**:
- **越高越好**: 表示系统吞吐能力强
- **关注趋势**: 不是绝对值，而是变化趋势
- **结合延迟**: TPS 高但延迟也高，可能是过载

### 延迟指标 (Latency)

**Avg (平均延迟)**:
- 所有请求的平均响应时间
- 参考，但不够全面

**P95 (95分位延迟)**:
- 95% 的请求延迟低于此值
- 更接近用户体验
- **重点关注**

**P99 (99分位延迟)**:
- 99% 的请求延迟低于此值
- 反映尾部延迟
- **SLA 关键指标**

**Min/Max**:
- 最小/最大延迟
- 帮助识别异常

**标准差 (StdDev)**:
- 延迟波动程度
- **越小越好**: 表示性能稳定

### QPS (Queries Per Second)

**定义**: 每秒执行的查询数

**与 TPS 的关系**:
- QPS ≈ TPS × 每事务查询数
- OLTP 场景: QPS 通常 = 2 × TPS (1 读 + 1 写)

### 读写比例 (Read/Write Ratio)

**定义**: 读查询与写查询的比例

**解读**:
- **Read 为主** (>80%): 适合读写分离、缓存优化
- **Write 为主** (>30%): 需关注写入性能、事务隔离级别
- **平衡读写** (50/50): 综合优化

### 统计指标的意义

**最小值 (Min)**:
- 最低性能
- 可能是异常值

**最大值 (Max)**:
- 最高性能
- 理想情况下的表现

**平均值 (Avg)**:
- 平均性能
- 容易受极端值影响

**标准差 (StdDev)**:
- 波动程度
- **小**: 性能稳定
- **大**: 性能不稳定，需要优化

---

## 最佳实践

### 1. 测试准备

**保持一致性**:
- 相同的测试数据量
- 相同的测试时长（建议 ≥ 5 分钟）
- 相同的预热时间
- 相同的系统负载（尽量无其他负载）

**多次运行**:
- 每个配置运行 3-5 次
- 取平均值或中位数
- 丢弃异常值

### 2. 记录选择

**可比性**:
- 选择相同模板的记录
- 选择相同测试时长的记录
- 注意数据库版本差异

**数量控制**:
- 建议 2-5 条记录
- 严格限制为 5 条（系统强制限制）
- 如需对比更多记录，请采用分步对比策略

### 3. 分组选择

**按 Threads 分组**:
- 适用: 线程数对比、并发调优
- 不适用: 不同数据库对比

**按 Database Type 分组**:
- 适用: 数据库选型
- 不适用: 线程数对比

**按 Date 分组**:
- 适用: 性能趋势分析、回归测试
- 不适用: 配置对比

### 4. 结果分析

**综合判断**:
- 不要只看 TPS
- 结合延迟、错误率、资源消耗
- 考虑业务场景需求

**关注稳定性**:
- P99 延迟比平均延迟更重要
- 标准差越小越好
- 多次运行结果应该接近

### 5. 常见误区

**❌ 错误做法**:
1. 只看平均 TPS，忽略 P99 延迟
2. 对比不同测试时长的记录
3. 对比不同数据量的记录
4. 只运行一次就下结论
5. 忽略系统负载差异

**✅ 正确做法**:
1. 综合考虑 TPS、P95、P99 延迟
2. 确保测试条件一致
3. 多次运行取平均值
4. 控制系统负载
5. 记录完整的测试环境信息

---

## 故障排查

### 问题 1: "No records to compare"

**原因**: 没有选择任何记录

**解决**: 勾选至少 2 条记录

### 问题 2: "Too many records selected (maximum 5)"

**原因**: 选择了超过 5 条记录

**解决**: 取消部分记录，只保留关键的 2-5 条

**为什么限制为 5 条？**
- 超过 5 条时，表格和图表会变得很长，难以阅读
- 推荐采用分步对比策略（见"选择记录"章节的说明）

### 问题 3: "Some records not found"

**原因**: 记录已被删除或 ID 错误

**解决**: 刷新记录列表，重新选择

### 问题 4: 对比结果不符合预期

**可能原因**:
1. 测试条件不一致
2. 系统负载不同
3. 数据库状态不同（如缓存）

**解决**:
1. 重新运行测试，保持条件一致
2. 多次运行取平均值
3. 记录详细的测试环境信息

---

## 进阶技巧

### 1. 自动化对比脚本

```bash
#!/bin/bash
# 自动对比最新两次测试

# 获取最新的两条记录 ID
LATESTtwo=$(db-benchmind-cli list | tail -n 2 | awk '{print $2}')

# 执行对比
db-benchmind-cli compare $LATEST_TWO --group-by date --output comparison.md
```

### 2. 定期回归测试

```bash
# 每天凌晨运行，对比与基线的差异
0 0 * * * /path/to/regression-test.sh
```

### 3. 集成到 CI/CD

```yaml
# .github/workflows/bench.yml
name: Benchmark Test
on: [push]
jobs:
  benchmark:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Run Benchmark
        run: |
          db-benchmind-cli run --config sysbench-oltp.json
          db-benchmind-cli compare baseline latest --fail-on-regression 5%
```

---

## 附录

### 完整示例代码

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/whhaicheng/DB-BenchMind/internal/app/usecase"
    "github.com/whhaicheng/DB-BenchMind/internal/domain/comparison"
    "github.com/whhaicheng/DB-BenchMind/internal/infra/database"
    "github.com/whhaicheng/DB-BenchMind/internal/infra/database/repository"
)

func main() {
    ctx := context.Background()

    // 初始化数据库
    db, err := database.InitializeSQLite(ctx, "./data/db-benchmind.db")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // 初始化 repository 和 use case
    historyRepo := repository.NewSQLiteHistoryRepository(db)
    comparisonUC := usecase.NewComparisonUseCase(historyRepo)

    // 获取所有记录
    records, err := comparisonUC.GetAllRecords(ctx)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Found %d records\n\n", len(records))

    // 选择最近的 3 条记录进行对比
    if len(records) < 3 {
        fmt.Println("Not enough records for comparison")
        return
    }

    recordIDs := []string{
        records[0].ID,
        records[1].ID,
        records[2].ID,
    }

    // 按线程数分组对比
    result, err := comparisonUC.CompareRecords(
        ctx,
        recordIDs,
        comparison.GroupByThreads,
    )
    if err != nil {
        log.Fatal(err)
    }

    // 输出结果
    fmt.Println(result.FormatTable())
    fmt.Println(result.FormatBarChart("TPS"))

    // 保存到文件
    f, _ := os.Create("comparison-result.txt")
    defer f.Close()
    f.WriteString(result.FormatTable())
    f.WriteString(result.FormatBarChart("TPS"))
}
```

---

**版本 1.0.0 - 完**
