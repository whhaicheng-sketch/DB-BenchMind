# DB-BenchMind 产品需求文档 (Spec)

**版本**: 1.0.0
**日期**: 2026-01-27
**状态**: 待评审

---

## 文档变更历史

| 版本 | 日期 | 作者 | 变更说明 |
|------|------|------|---------|
| 1.0.0 | 2026-01-27 | Claude | 初始版本 |

---

## 1. 项目概述

### 1.1 项目背景

DB-BenchMind 是一款面向数据库工程师与性能测试工程师的**桌面压测工作台**，旨在通过统一的 GUI 界面编排与运行外部压测工具（Sysbench、Swingbench、HammerDB），对数据库进行性能压测、监控、结果归档与报告导出。

### 1.2 项目目标

- **可复现性**：每次压测配置与结果完整保存，支持精确复跑
- **可对比性**：多次运行结果可对比分析，支持基线管理
- **可交付性**：多格式报告导出（MD/HTML/JSON/PDF），满足不同场景需求
- **易用性**：桌面 GUI 操作，降低命令行工具使用门槛

### 1.3 运行环境

- **操作系统**: Ubuntu 24（带桌面 GUI）
- **Go 版本**: 1.22.2
- **GUI 框架**: Fyne v2.x
- **数据库**: SQLite（存储结果与配置）
- **密钥管理**: gnome-keyring（密码存储）

### 1.4 技术栈决策

| 技术选型 | 决策理由 |
|---------|---------|
| **GUI 框架**: Fyne | 跨平台、纯 Go、自绘制引擎，打包简单 |
| **存储**: SQLite (modernc.org/sqlite) | 无 CGO、纯 Go、高性能 |
| **密钥管理**: go-keyring | 支持 Ubuntu 24 gnome-keyring，安全可靠 |
| **报告导出**: pandoc/wkhtmltopdf | 支持多格式转换 |
| **图表**: fynesimplechart | Fyne 生态兼容 |

---

## 2. 核心概念定义

| 概念 | 定义 | 示例 |
|------|------|------|
| **Connection（连接）** | 数据库连接配置，包含连接参数与认证信息 | MySQL-生产库、Oracle-测试库 |
| **Template（模板）** | 工具与 workload 的预定义配置 | Sysbench OLTP Read-Write |
| **Scenario（场景）** | 具体的压测场景，由模板 + 参数组成 | 100 线程、300 秒 OLTP 读写 |
| **Task（任务）** | 一次压测的完整配置，包含连接、模板、参数快照 | Task-20260127-001 |
| **Run（运行）** | 任务的一次执行记录，包含状态、日志、结果 | Run-abc123-def456 |
| **Result（结果）** | 解析后的指标数据与摘要 | TPS: 1250.67, P95: 3.45ms |
| **Report（报告）** | 导出的测试报告文件 | report.md, report.html, report.pdf |
| **Baseline（基线）** | 作为对比基准的运行记录 | Run-previous-v1.0 |

---

## 3. 功能需求

### 3.1 信息架构与页面流程

#### 3.1.1 页面导航结构

主窗口采用 Fyne TabLayout，包含以下页面：

| Tab | 功能 | 核心操作 |
|-----|------|---------|
| **连接管理** | 数据库连接配置 | 新增/编辑/删除/测试连接 |
| **场景模板** | 压测场景与模板 | 查看/导入/导出/自定义模板 |
| **任务配置** | 配置压测任务参数 | 选择连接+工具+模板+参数 |
| **运行监控** | 实时监控执行状态 | 启动/停止/查看日志/实时指标 |
| **历史记录** | 查看历史执行记录 | 列表/详情/重跑/删除 |
| **结果对比** | 多次运行结果对比 | 选择基线/对比项/差异展示 |
| **报告导出** | 导出测试报告 | 选择运行/格式/导出 |
| **设置** | 工具路径与偏好配置 | 工具路径/结果目录/日志级别 |

#### 3.1.2 用户操作流程

**首次使用流程**:
```
启动应用 → 设置工具路径 → 添加数据库连接 → 查看内置模板 → 配置任务 → 启动压测 → 查看结果
```

**日常使用流程**:
```
启动应用 → 选择历史连接 → 选择模板 → 配置参数 → 启动任务 → 实时监控 → 导出报告
```

**对比分析流程**:
```
查看历史记录 → 选择多次运行 → 进入对比页面 → 查看差异 → 导出对比报告
```

---

### 3.2 数据库连接管理

#### 3.2.1 支持的数据库类型（按优先级）

1. **MySQL** - 最高优先级
2. **Oracle**
3. **SQL Server**
4. **PostgreSQL**

#### 3.2.2 连接字段定义

**MySQL 连接配置**:
```go
type MySQLConnection struct {
    Name     string // 连接名称（用户自定义）
    Host     string // 主机地址
    Port     int    // 端口（默认 3306）
    Database string // 数据库名
    Username string // 用户名
    Password string // 密码（存储到 keyring，不序列化）
    SSLMode  string // SSL模式：disabled/preferred/required
}
```

**Oracle 连接配置**:
```go
type OracleConnection struct {
    Name        string // 连接名称
    Host        string // 主机地址
    Port        int    // 端口（默认 1521）
    ServiceName string // 服务名
    SID         string // SID（与 ServiceName 二选一）
    Username    string // 用户名
    Password    string // 密码（存储到 keyring）
}
```

**SQL Server 连接配置**:
```go
type SQLServerConnection struct {
    Name                        string // 连接名称
    Host                        string // 主机地址
    Port                        int    // 端口（默认 1433）
    Database                    string // 数据库名
    Username                    string // 用户名
    Password                    string // 密码（存储到 keyring）
    TrustServerCertificate      bool   // 信任服务器证书
}
```

**PostgreSQL 连接配置**:
```go
type PostgreSQLConnection struct {
    Name     string // 连接名称
    Host     string // 主机地址
    Port     int    // 端口（默认 5432）
    Database string // 数据库名
    Username string // 用户名
    Password string // 密码（存储到 keyring）
    SSLMode  string // SSL模式：disable/allow/prefer/require/verify-ca/verify-full
}
```

#### 3.2.3 功能需求

| 需求 ID | 需求描述 | 优先级 |
|---------|---------|--------|
| REQ-CONN-001 | WHEN 用户点击"新增连接"，THE SYSTEM SHALL 显示连接配置表单 | P0 |
| REQ-CONN-002 | WHEN 用户选择数据库类型，THE SYSTEM SHALL 显示对应字段的配置表单 | P0 |
| REQ-CONN-003 | WHEN 用户填写连接信息并点击"测试连接"，THE SYSTEM SHALL 验证连接可用性 | P0 |
| REQ-CONN-004 | WHEN 连接测试成功，THE SYSTEM SHALL 显示"成功 (耗时 XXms)"及数据库版本 | P0 |
| REQ-CONN-005 | WHEN 连接测试失败，THE SYSTEM SHALL 显示具体错误信息（如"连接超时"、"认证失败"） | P0 |
| REQ-CONN-006 | WHEN 用户保存连接，THE SYSTEM SHALL 将密码加密存储到 keyring | P0 |
| REQ-CONN-007 | WHEN keyring 不可用，THE SYSTEM SHALL 使用加密文件降级方案并提示用户 | P1 |
| REQ-CONN-008 | WHEN 用户编辑连接，THE SYSTEM SHALL 加载连接信息（密码显示为 ••••••） | P0 |
| REQ-CONN-009 | WHEN 用户删除连接，THE SYSTEM SHALL 弹出确认对话框并删除 keyring 中的密码 | P0 |
| REQ-CONN-010 | WHILE 用户提供非法输入（如端口超出范围），THE SYSTEM SHALL 实时显示错误提示 | P0 |

#### 3.2.4 脱敏显示规则

- **密码输入框**: 始终显示为 `••••••••`
- **连接列表**: 显示格式为 `名称 (host:port/db)`
- **报告导出**: 连接信息中的密码必须脱敏或省略

---

### 3.3 场景与模板管理

#### 3.3.1 内置模板列表

| ID | 名称 | 工具 | 数据库 | 描述 |
|----|------|------|--------|------|
| `sysbench-oltp-read-write` | Sysbench OLTP 读写混合 | Sysbench | MySQL, PG | 标准 OLTP 场景，70% 读 30% 写 |
| `sysbench-oltp-read-only` | Sysbench OLTP 只读 | Sysbench | MySQL, PG | 纯读压测，100% SELECT |
| `sysbench-oltp-write-only` | Sysbench OLTP 只写 | Sysbench | MySQL, PG | 纯写压测，INSERT/UPDATE/DELETE |
| `swingbench-soe` | Swingbench Order Entry | Swingbench | Oracle | 模拟订单处理系统 |
| `swingbench-calling` | Swingbench Calling Circle | Swingbench | Oracle | 模拟电信话务系统 |
| `hammerdb-tpcc` | HammerDB TPROC-C | HammerDB | 全部 | 标准 TPC-C 基准测试 |
| `hammerdb-tpcb` | HammerDB TPROC-B | HammerDB | 全部 | 标准 TPC-B 基准测试 |

#### 3.3.2 模板 Schema（JSON 格式）

```json
{
  "$schema": "https://db-benchmind.dev/schemas/template/v1.json",
  "id": "sysbench-oltp-read-write",
  "name": "Sysbench OLTP Read-Write",
  "description": "标准的 OLTP 读写混合压测场景",
  "tool": "sysbench",
  "database_types": ["mysql", "postgresql"],
  "version": "1.0.0",
  "parameters": {
    "threads": {
      "type": "integer",
      "label": "线程数",
      "default": 8,
      "min": 1,
      "max": 1024
    },
    "time": {
      "type": "integer",
      "label": "运行时长（秒）",
      "default": 60,
      "min": 10,
      "max": 86400
    }
  },
  "command_template": {
    "prepare": "sysbench {db_type} --tables={tables} --table-size={table_size} {connection_string} prepare",
    "run": "sysbench {db_type} --threads={threads} --time={time} {connection_string} run",
    "cleanup": "sysbench {db_type} --tables={tables} {connection_string} cleanup"
  },
  "output_parser": {
    "type": "regex",
    "patterns": {
      "tps": "transactions:\\s*\\(\\s*(\\d+\\.\\d+)\\s*per sec.",
      "latency_avg": "latency:\\s*\\(ms\\).*?avg=\\s*(\\d+\\.\\d+)"
    }
  }
}
```

#### 3.3.3 功能需求

| 需求 ID | 需求描述 | 优先级 |
|---------|---------|--------|
| REQ-TMPL-001 | WHEN 用户进入"场景模板"页面，THE SYSTEM SHALL 显示内置模板列表 | P0 |
| REQ-TMPL-002 | WHEN 用户选择模板，THE SYSTEM SHALL 显示模板详细信息（参数、命令、解析规则） | P0 |
| REQ-TMPL-003 | WHEN 用户点击"导入模板"，THE SYSTEM SHALL 支持导入 JSON/YAML 格式文件 | P0 |
| REQ-TMPL-004 | WHEN 用户导入模板，THE SYSTEM SHALL 验证 JSON Schema 并检查工具可用性 | P0 |
| REQ-TMPL-005 | WHEN 用户导入模板失败，THE SYSTEM SHALL 显示具体错误（如"JSON 格式错误"、"工具未安装"） | P0 |
| REQ-TMPL-006 | WHEN 用户导出模板，THE SYSTEM SHALL 导出包含模板定义、使用统计的完整信息 | P0 |
| REQ-TMPL-007 | WHEN 用户配置任务，THE SYSTEM SHALL 保存模板快照以保证可复现 | P0 |

---

### 3.4 压测任务编排与执行

#### 3.4.1 任务配置结构

```go
type BenchmarkTask struct {
    ID           string            // UUID
    Name         string            // 任务名称
    ConnectionID string            // 连接 ID
    TemplateID   string            // 模板 ID
    Parameters   map[string]any    // 参数覆盖
    Options      TaskOptions       // 执行选项
    Tags         []string          // 标签
}

type TaskOptions struct {
    SkipPrepare     bool          // 跳过数据准备
    SkipCleanup     bool          // 跳过数据清理
    WarmupTime      int           // 预热时长（秒）
    SampleInterval  time.Duration // 采样间隔（默认 1s）
    DryRun          bool          // 仅显示命令不执行
    PrepareTimeout  time.Duration // 准备阶段超时（默认 30m）
    RunTimeout      time.Duration // 运行阶段超时（默认 24h）
}
```

#### 3.4.2 执行状态机

```go
type RunState string

const (
    StatePending      RunState = "pending"       // 已创建，待执行
    StatePreparing    RunState = "preparing"     // 准备数据中
    StatePrepared     RunState = "prepared"      // 准备完成
    StateWarmingUp    RunState = "warming_up"    // 预热中
    StateRunning      RunState = "running"       // 正式运行
    StateCompleted    RunState = "completed"     // 正常完成
    StateFailed       RunState = "failed"        // 执行失败
    StateCancelled    RunState = "cancelled"     // 用户取消
    StateTimeout      RunState = "timeout"       // 超时
    StateForceStopped RunState = "force_stopped" // 强制停止
)
```

**状态转换规则**:
- `pending` → `preparing` → `prepared` → `warming_up` → `running` → `completed`
- 任何状态 → `cancelled`（用户取消）
- 任何状态 → `failed`（执行失败）
- 任何状态 → `timeout`（超时）
- `running` → `force_stopped`（强制停止）

#### 3.4.3 功能需求

| 需求 ID | 需求描述 | 优先级 |
|---------|---------|--------|
| REQ-EXEC-001 | WHEN 用户配置任务并点击"启动"，THE SYSTEM SHALL 执行预检查 | P0 |
| REQ-EXEC-002 | WHEN 预检查通过，THE SYSTEM SHALL 依次执行 prepare → warmup → run → cleanup | P0 |
| REQ-EXEC-003 | WHEN 任务执行，THE SYSTEM SHALL 在运行监控页面显示实时状态 | P0 |
| REQ-EXEC-004 | WHILE 任务运行，THE SYSTEM SHALL 每秒采集并显示指标（TPS、延迟、错误率） | P0 |
| REQ-EXEC-005 | WHILE 任务运行，THE SYSTEM SHALL 实时滚动显示 stdout/stderr 日志 | P0 |
| REQ-EXEC-006 | WHEN 用户点击"停止"，THE SYSTEM SHALL 发送 SIGTERM 并等待最多 30 秒（优雅停止） | P0 |
| REQ-EXEC-007 | WHEN 优雅停止超时，THE SYSTEM SHALL 发送 SIGKILL（强制停止） | P0 |
| REQ-EXEC-008 | WHEN 任务完成/失败/取消，THE SYSTEM SHALL 保存完整结果到 SQLite | P0 |
| REQ-EXEC-009 | WHEN 任务超时，THE SYSTEM SHALL 自动停止任务并标记为 timeout | P0 |
| REQ-EXEC-010 | WHEN 用户启用"仅显示命令"，THE SYSTEM SHALL 显示完整命令但不执行 | P1 |

#### 3.4.4 预检查规则

| 检查项 | 规则 | 失败提示 |
|--------|------|---------|
| 工具存在 | 可执行文件在 PATH 或指定路径 | "工具未安装：Sysbench 未找到" |
| 工具版本 | 版本符合要求 | "工具版本过低：Sysbench >= 1.0 required" |
| 连接可用 | 能连接到数据库 | "数据库连接失败：connection refused" |
| 磁盘空间 | 剩余空间 >= 1GB | "磁盘空间不足：需要至少 1GB" |
| 参数合法性 | 参数在模板定义的范围内 | "线程数超出范围：1-1024" |
| 互斥检查 | 预热时长 < 运行时长 | "预热时长（60s）必须小于运行时长（30s）" |

---

### 3.5 指标采集与可视化

#### 3.5.1 统一指标结构

```go
type MetricSample struct {
    Timestamp    time.Time            // 采样时间
    RunID        string               // 运行 ID
    Phase        string               // 阶段：warmup/run/cooldown
    TPS          float64              // 每秒事务数
    QPS          float64              // 每秒查询数
    LatencyAvg   float64              // 平均延迟（毫秒）
    LatencyP95   float64              // 95分位延迟（毫秒）
    LatencyP99   float64              // 99分位延迟（毫秒）
    ErrorRate    float64              // 错误率（%）
}

type BenchmarkResult struct {
    RunID            string            // 运行 ID
    TPSCalculated    float64           // 计算得到的 TPS
    LatencyAvg       float64           // 平均延迟
    LatencyP95       float64           // 95分位延迟
    LatencyP99       float64           // 99分位延迟
    ErrorCount       int64             // 错误总数
    ErrorRate        float64           // 错误率
    Duration         time.Duration     // 运行时长
    TotalTransactions int64            // 总事务数
    TimeSeries       []MetricSample    // 时间序列数据
}
```

#### 3.5.2 指标口径定义

| 指标 | 定义 | 单位 | 计算方式 |
|------|------|------|---------|
| TPS | 每秒事务数 | tps | total_transactions / duration |
| QPS | 每秒查询数 | qps | total_queries / duration |
| Latency_Avg | 平均延迟 | ms | 所有事务延迟平均值 |
| Latency_P95 | 95分位延迟 | ms | 排序后 95% 位置的值 |
| Latency_P99 | 99分位延迟 | ms | 排序后 99% 位置的值 |
| Error_Rate | 错误率 | % | (errors / total) * 100 |

#### 3.5.3 可视化要求

| 图表类型 | 描述 | 实现方式 |
|---------|------|---------|
| TPS 趋势图 | 折线图，X 轴为时间，Y 轴为 TPS | fynesimplechart.LineChart |
| 延迟分布图 | 柱状图，显示 min/avg/p95/p99 | fynesimplechart.BarChart |
| 错误率趋势 | 折线图，X 轴为时间，Y 轴为错误率 | fynesimplechart.LineChart |

#### 3.5.4 功能需求

| 需求 ID | 需求描述 | 优先级 |
|---------|---------|--------|
| REQ-METRIC-001 | WHILE 任务运行，THE SYSTEM SHALL 每秒采集一次指标 | P0 |
| REQ-METRIC-002 | WHEN 用户查看运行监控页面，THE SYSTEM SHALL 显示实时指标数值 | P0 |
| REQ-METRIC-003 | WHEN 用户查看运行监控页面，THE SYSTEM SHALL 绘制实时趋势图 | P0 |
| REQ-METRIC-004 | WHEN 任务完成，THE SYSTEM SHALL 生成汇总指标（min/max/avg/p95/p99） | P0 |
| REQ-METRIC-005 | WHEN 用户查看历史记录详情，THE SYSTEM SHALL 显示完整趋势图 | P0 |

---

### 3.6 结果归档与本地存储

#### 3.6.1 SQLite 表结构（核心表）

```sql
-- 运行记录（核心表）
CREATE TABLE runs (
    id TEXT PRIMARY KEY,
    task_id TEXT NOT NULL,
    state TEXT NOT NULL,
    created_at TEXT NOT NULL,
    started_at TEXT,
    completed_at TEXT,
    duration_seconds REAL,
    result_json TEXT,
    error_message TEXT
);

-- 时间序列指标
CREATE TABLE metric_samples (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    run_id TEXT NOT NULL,
    timestamp TEXT NOT NULL,
    phase TEXT NOT NULL,
    tps REAL,
    latency_avg REAL,
    latency_p95 REAL,
    error_rate REAL,
    FOREIGN KEY (run_id) REFERENCES runs(id)
);

-- 运行日志
CREATE TABLE run_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    run_id TEXT NOT NULL,
    timestamp TEXT NOT NULL,
    stream TEXT NOT NULL,
    content TEXT NOT NULL,
    FOREIGN KEY (run_id) REFERENCES runs(id)
);
```

#### 3.6.2 Run 目录结构示例

```
results/
└── sysbench_oltp_abc123-def456_20260127_150000/
    ├── config_snapshot.json           # 配置快照
    ├── connection_snapshot.json       # 连接配置（脱敏）
    ├── template_snapshot.json         # 模板定义
    ├── run_metadata.json              # 运行元数据
    ├── stdout.log                     # 标准输出完整日志
    ├── stderr.log                     # 标准错误完整日志
    ├── raw_output.txt                 # 工具原始输出
    ├── parsed_result.json             # 解析后的结果
    ├── metrics.jsonl                  # 时间序列指标（JSON Lines）
    ├── charts/
    │   ├── tps_trend.png
    │   ├── latency_distribution.png
    │   └── error_rate.png
    └── reports/
        ├── report.md
        ├── report.html
        ├── report.json
        └── report.pdf
```

#### 3.6.3 功能需求

| 需求 ID | 需求描述 | 优先级 |
|---------|---------|--------|
| REQ-STORAGE-001 | WHEN 任务完成/失败/取消，THE SYSTEM SHALL 保存完整运行记录到 SQLite | P0 |
| REQ-STORAGE-002 | WHEN 保存运行记录，THE SYSTEM SHALL 同时保存配置快照到独立目录 | P0 |
| REQ-STORAGE-003 | WHEN 保存运行记录，THE SYSTEM SHALL 保存原始日志到文件 | P0 |
| REQ-STORAGE-004 | WHEN 用户查看历史记录，THE SYSTEM SHALL 从 SQLite 查询并按时间倒序显示 | P0 |
| REQ-STORAGE-005 | WHEN 用户选择某次运行，THE SYSTEM SHALL 显示详情（指标、日志、配置快照） | P0 |
| REQ-STORAGE-006 | WHEN 用户删除运行记录，THE SYSTEM SHALL 同时删除 SQLite 记录和文件目录 | P1 |

---

### 3.7 报告导出

#### 3.7.1 报告格式支持

MVP 必须支持以下 4 种格式：

| 格式 | 用途 | 生成方式 |
|------|------|---------|
| **Markdown** | 技术文档、版本控制 | 模板渲染 |
| **HTML** | 浏览器查看、在线分享 | Markdown → HTML 转换 |
| **JSON** | 程序化处理、API 集成 | 结构化数据导出 |
| **PDF** | 正式报告、打印归档 | Markdown/HTML → PDF 转换 |

#### 3.7.2 报告内容要求

每份报告必须包含以下章节：

1. **执行摘要**
   - 任务名称、数据库类型、测试工具、场景模板
   - 核心指标摘要（TPS、延迟、错误率）
   - 结论与建议

2. **测试环境**
   - 数据库连接信息（已脱敏）
   - 运行参数（并发、时长、数据规模）

3. **性能指标详情**
   - TPS/QPS 汇总与趋势图
   - 延迟分布（min/avg/median/p95/p99）
   - 错误统计与分析

4. **时间序列数据**
   - 关键指标随时间变化趋势

5. **运行日志**（关键片段）
   - 错误日志
   - 警告日志

#### 3.7.3 文件命名规则

```
格式: {task_name}_{run_id}_{timestamp}.{ext}

示例:
- sysbench_oltp_abc123_20260127_150000.md
- swingbench_soe_def456_20260127_153000.html
- hammerdb_tpcc_ghi789_20260127_160000.json
- comparison_baseline_vs_test_20260127_163000.pdf
```

#### 3.7.4 功能需求

| 需求 ID | 需求描述 | 优先级 |
|---------|---------|--------|
| REQ-RPT-001 | WHEN 用户进入"报告导出"页面并选择运行，THE SYSTEM SHALL 显示可导出格式 | P0 |
| REQ-RPT-002 | WHEN 用户选择 Markdown 格式并点击"导出"，THE SYSTEM SHALL 生成 .md 文件 | P0 |
| REQ-RPT-003 | WHEN 用户选择 HTML 格式并点击"导出"，THE SYSTEM SHALL 生成 .html 文件 | P0 |
| REQ-RPT-004 | WHEN 用户选择 JSON 格式并点击"导出"，THE SYSTEM SHALL 生成 .json 文件 | P0 |
| REQ-RPT-005 | WHEN 用户选择 PDF 格式并点击"导出"，THE SYSTEM SHALL 生成 .pdf 文件 | P0 |
| REQ-RPT-006 | WHEN 报告生成，THE SYSTEM SHALL 包含所有必需章节（摘要、环境、指标、日志） | P0 |
| REQ-RPT-007 | WHEN 报告生成，THE SYSTEM SHALL 包含趋势图（PNG 格式） | P0 |
| REQ-RPT-008 | WHEN 报告生成，THE SYSTEM SHALL 对敏感信息进行脱敏（密码、IP） | P0 |
| REQ-RPT-009 | WHEN PDF 导出失败（如 pandoc 未安装），THE SYSTEM SHALL 提示用户并提供替代方案 | P1 |
| REQ-RPT-010 | WHEN 用户导出对比报告，THE SYSTEM SHALL 生成差异分析（baseline vs compare） | P0 |

---

### 3.8 结果对比功能（完整系统）

#### 3.8.1 对比维度

| 对比类型 | 描述 | 实现方式 |
|---------|------|---------|
| **A/B 对比** | 两次运行的直接对比 | 差异计算、百分比变化 |
| **基线对比** | 多次运行与固定基线对比 | 趋势分析、偏差检测 |
| **趋势分析** | 同一任务多次运行的性能趋势 | 折线图、移动平均 |
| **图表对比** | 指标图表的直观对比 | 叠加图、差值图 |

#### 3.8.2 对比视图设计

```
+------------------------------------------------------------+
|                    结果对比                                 |
+------------------------------------------------------------+
| 基线: Run-abc123 (2026-01-26 10:00)                       |
| 对比: Run-def456 (2026-01-27 15:00)                       |
+------------------------------------------------------------+
| 核心指标差异                                                |
| +--------------------------------------------------------+ |
| | 指标       | 基线      | 对比      | 变化      | 趋势   | |
| |-----------|-----------|-----------|-----------|--------| |
| | TPS       | 1200.50   | 1250.75   | +4.18%    | ↑      | |
| | P95 延迟  | 3.45 ms   | 3.20 ms   | -7.25%    | ↓      | |
| | 错误率    | 0.05%     | 0.02%     | -60.00%   | ↓      | |
| +--------------------------------------------------------+ |
+------------------------------------------------------------+
| 图表对比                                                    |
| [TPS 趋势对比图] [延迟分布对比图] [错误率对比图]             |
+------------------------------------------------------------+
| 结论                                                        |
| 对比运行相比基线性能提升 4.18%，延迟降低 7.25%，错误率降低 60% |
+------------------------------------------------------------+
| [导出对比报告]                                             |
+------------------------------------------------------------+
```

#### 3.8.3 功能需求

| 需求 ID | 需求描述 | 优先级 |
|---------|---------|--------|
| REQ-COMP-001 | WHEN 用户进入"结果对比"页面，THE SYSTEM SHALL 显示历史运行列表 | P0 |
| REQ-COMP-002 | WHEN 用户选择两次运行并点击"对比"，THE SYSTEM SHALL 生成对比视图 | P0 |
| REQ-COMP-003 | WHEN 对比视图生成，THE SYSTEM SHALL 计算并显示核心指标的差异 | P0 |
| REQ-COMP-004 | WHEN 对比视图生成，THE SYSTEM SHALL 绘制对比图表（叠加显示） | P0 |
| REQ-COMP-005 | WHEN 用户选择"基线对比"，THE SYSTEM SHALL 支持选择固定基线并多次对比 | P0 |
| REQ-COMP-006 | WHEN 用户选择"趋势分析"，THE SYSTEM SHALL 绘制同一任务多次运行的趋势图 | P0 |
| REQ-COMP-007 | WHEN 用户点击"导出对比报告"，THE SYSTEM SHALL 生成包含差异分析的报告 | P0 |

---

### 3.9 设置与配置

#### 3.9.1 工具路径配置

| 工具 | 配置项 | 默认值 | 检测方式 |
|------|--------|--------|---------|
| **Sysbench** | `tool.sysbench.path` | `sysbench` | `sysbench --version` |
| **Swingbench** | `tool.swingbench.path` | 空 | 检查 JAR 文件是否存在 |
| **HammerDB** | `tool.hammerdb.path` | 空 | 检查可执行文件是否存在 |

#### 3.9.2 全局配置项

| 配置键 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `result_directory` | string | `./results` | 结果文件存储目录 |
| `log_level` | string | `info` | 日志级别：debug/info/warn/error |
| `sample_interval` | duration | `1s` | 默认采样间隔 |
| `max_history_runs` | int | `1000` | 最大保留历史记录数 |
| `auto_cleanup_days` | int | `30` | 自动清理天数（0=禁用） |
| `default_timeout_prepare` | duration | `30m` | 准备阶段默认超时 |
| `default_timeout_run` | duration | `24h` | 运行阶段默认超时 |
| `graceful_stop_timeout` | duration | `30s` | 优雅停止超时 |
| `enable_keyring` | bool | `true` | 是否启用 keyring |

#### 3.9.3 功能需求

| 需求 ID | 需求描述 | 优先级 |
|---------|---------|--------|
| REQ-SET-001 | WHEN 用户进入"设置"页面，THE SYSTEM SHALL 显示所有配置项 | P0 |
| REQ-SET-002 | WHEN 用户修改工具路径并点击"检测版本"，THE SYSTEM SHALL 执行工具并显示版本 | P0 |
| REQ-SET-003 | WHEN 工具未检测到，THE SYSTEM SHALL 显示错误提示（如"工具未找到"） | P0 |
| REQ-SET-004 | WHEN 用户修改配置并点击"保存"，THE SYSTEM SHALL 保存到 SQLite 并立即生效 | P0 |
| REQ-SET-005 | WHEN 修改结果目录，THE SYSTEM SHALL 验证目录可写性 | P0 |

---

## 4. 工具适配器设计

### 4.1 统一适配器接口

```go
// pkg/benchmark/adapter.go
package benchmark

type Adapter interface {
    // 基本信息
    ToolName() string
    Version() (string, error)
    Available() bool

    // 命令构建
    BuildPrepareCommand(ctx context.Context, cfg *Config) (*exec.Cmd, error)
    BuildRunCommand(ctx context.Context, cfg *Config) (*exec.Cmd, error)
    BuildCleanupCommand(ctx context.Context, cfg *Config) (*exec.Cmd, error)

    // 输出解析
    ParseRunOutput(stdout, stderr string) (*Result, error)

    // 实时采集
    StartRealtimeCollection(ctx context.Context, reader io.Reader) (<-chan Sample, error)

    // 配置验证
    ValidateConfig(cfg *Config) error
}

type Config struct {
    Connection   any                // 数据库连接配置
    Template     *Template          // 场景模板
    Parameters   map[string]any     // 参数覆盖
    WorkDir      string             // 工作目录
    Timeout      time.Duration      // 超时时间
}

type Result struct {
    Success     bool                   `json:"success"`
    Metrics     map[string]float64     `json:"metrics"`
    RawOutput   string                 `json:"raw_output"`
    TimeSeries  []Sample               `json:"time_series,omitempty"`
}

type Sample struct {
    Timestamp time.Time               `json:"timestamp"`
    Metrics   map[string]float64      `json:"metrics"`
}
```

### 4.2 Sysbench 适配器实现要点

**命令构建**:
```bash
# prepare
sysbench oltp_read_write \
  --mysql-host=localhost \
  --mysql-port=3306 \
  --mysql-user=test \
  --mysql-password=*** \
  --mysql-db=test \
  --tables=10 \
  --table-size=10000 \
  prepare

# run
sysbench oltp_read_write \
  --mysql-host=localhost \
  --mysql-port=3306 \
  --mysql-user=test \
  --mysql-password=*** \
  --mysql-db=test \
  --threads=8 \
  --time=300 \
  --report-interval=1 \
  run
```

**输出解析**:
- 使用正则表达式解析文本输出
- 关键指标：`tps`、`latency_avg`、`latency_p95`、`errors`
- 实时采集：解析 `report-interval` 输出

### 4.3 Swingbench 适配器实现要点

**命令构建**:
```bash
java -jar swingbench.jar \
  -cs //localhost:1521/ORCL \
  -u test \
  -p *** \
  -dt user \
  -uc 10 \
  -at tpcc
```

**输出解析**:
- Swingbench 输出 XML 或 JSON 格式
- 关键指标：`TPM`（Transactions Per Minute）、`Average Latency`、`Error Count`
- TPM 转换为 TPS：TPM / 60

### 4.4 HammerDB 适配器实现要点

**命令构建**:
```tcl
# 生成 TCL 脚本
dbset db mysql
diset connection mysql_host localhost
diset connection mysql_port 3306
vucreate
vurun
```

**输出解析**:
- HammerDB 使用 TCL 脚本控制
- 关键指标：`NOPM`（New Orders Per Minute）、`TPM`、`Response Time`
- TCL 脚本输出结果到文件，然后解析

---

## 5. 非功能需求

### 5.1 稳定性

| 需求 | 指标 | 验证方法 |
|------|------|---------|
| GUI 崩溃率 | < 0.1% | 长时间运行测试（24h） |
| 任务执行失败恢复 | 自动恢复或明确提示 | 异常注入测试 |
| 内存泄漏 | 无显著增长 | 长时间运行 + pprof |
| 数据库连接池 | 自动重连 | 网络故障模拟测试 |

### 5.2 性能

| 指标 | 目标 | 说明 |
|------|------|------|
| GUI 响应时间 | < 100ms | 页面切换、按钮点击 |
| 日志实时显示延迟 | < 500ms | 从采集到显示 |
| 数据库查询 | < 50ms | 历史记录列表查询 |
| 内存占用 | < 200MB（空闲） | 基础 GUI 启动 |
| 内存占用 | < 500MB（运行中） | 单任务执行 |

### 5.3 可维护性

| 需求 | 说明 |
|------|------|
| 模块化设计 | 核心逻辑与 GUI 分离 |
| 接口隔离 | 适配器接口独立，易于扩展新工具 |
| 日志分级 | debug/info/warn/error 四级 |
| 配置外部化 | 所有关键配置可调 |
| 文档完整 | 代码注释 + 用户手册 + 架构文档 |

### 5.4 安全性

| 需求 | 实现方式 |
|------|---------|
| 密码加密存储 | keyring + 加密文件降级 |
| 敏感信息脱敏 | 日志、报告、GUI 显示均脱敏 |
| SQL 注入防护 | 使用参数化查询 |
| 命令注入防护 | 参数校验 + 转义 |
| 文件权限 | 结果目录权限 700 |

---

## 6. 异常场景处理

### 6.1 工具相关异常

| 场景 | 检测方式 | 处理策略 |
|------|---------|---------|
| **工具未安装** | `exec.LookPath` 失败 | 提示用户安装并提供安装文档链接 |
| **工具路径错误** | 配置路径不存在 | 标记工具为"不可用"，禁用相关模板 |
| **版本不兼容** | 版本号不符合要求 | 警告用户，提供兼容性说明 |
| **工具执行失败** | 进程退出码非 0 | 保存错误日志，标记 Run 为 failed |

### 6.2 数据库连接异常

| 场景 | 检测方式 | 处理策略 |
|------|---------|---------|
| **连接失败** | Ping 超时或认证失败 | 显示具体错误信息（如"连接超时"、"密码错误"） |
| **权限不足** | CREATE/DROP 权限检查 | 提示用户授予权限，提供 SQL 脚本 |
| **SSL 配置错误** | SSL握手失败 | 提示检查 SSL 配置，提供示例 |

### 6.3 执行中断异常

| 场景 | 检测方式 | 处理策略 |
|------|---------|---------|
| **用户强制停止** | GUI 停止按钮 | 优雅停止（SIGTERM）→ 超时后强制停止（SIGKILL） |
| **进程被 kill** | 子进程异常退出 | 标记 Run 为 failed，保存已采集数据 |
| **系统重启** | 下次启动检测 | 恢复运行状态，标记为 interrupted |

### 6.4 资源异常

| 场景 | 检测方式 | 处理策略 |
|------|---------|---------|
| **磁盘空间不足** | 预检查失败 | 阻止启动，提示清理空间 |
| **内存不足** | OOM 检测 | 优雅退出，保存已采集数据 |
| **日志膨胀** | 日志文件大小监控 | 自动滚动日志，限制最大大小 |

### 6.5 输出解析异常

| 场景 | 检测方式 | 处理策略 |
|------|---------|---------|
| **输出格式变化** | 正则解析失败 | 保留原始输出，提示"解析失败"，提供原始日志链接 |
| **JSON/XML 格式错误** | 解析失败 | 回退到文本解析，标记结果为"部分解析" |

---

## 7. 项目结构设计

### 7.1 目录结构

```
DB-BenchMind/
├── cmd/
│   └── db-benchmind/
│       └── main.go                    # GUI 入口
├── internal/
│   ├── app/                           # 应用层
│   │   ├── usecase/                   # 用例
│   │   └── service/                   # 服务
│   ├── domain/                        # 领域层
│   │   ├── connection/
│   │   ├── template/
│   │   ├── execution/
│   │   └── metric/
│   ├── infra/                         # 基础设施层
│   │   ├── adapter/                   # 适配器
│   │   ├── database/                  # 数据库
│   │   ├── keyring/                   # 密钥管理
│   │   ├── report/                    # 报告生成
│   │   └── chart/                     # 图表生成
│   └── transport/                     # 传输层
│       ├── ui/                        # GUI 页面
│       └── controller/                # 控制器
├── pkg/                               # 对外可复用库
│   └── benchmark/
│       └── adapter.go                 # 适配器接口
├── contracts/                         # 契约定义
│   ├── templates/                     # 内置模板
│   ├── schemas/                       # Schema 定义
│   └── reports/                       # 报告模板
├── configs/                           # 配置样例
├── scripts/                           # 脚本
├── test/                              # 测试
├── docs/                              # 文档
├── results/                           # 结果目录
├── go.mod
├── go.sum
├── Makefile
├── CLAUDE.md
├── constitution.md
└── README.md
```

### 7.2 依赖方向

```
transport (GUI) → app → domain ← infra
                        ↑
                      pkg
```

**规则**:
- `transport` 可依赖 `app`、`domain`、`pkg`
- `app` 可依赖 `domain`、`pkg`、`infra`（通过接口）
- `domain` 不可依赖任何外部包（仅标准库）
- `infra` 实现 `domain` 或 `app` 定义的接口

---

## 8. 实现路线图

### Phase 1: 基础设施与连接管理（Week 1-2）

**目标**: 搭建项目骨架，实现连接管理功能

**交付物**:
- 项目目录结构
- SQLite 数据库设计与迁移
- Keyring 集成
- 连接管理 GUI 页面
- 四种数据库连接配置
- 连接测试功能
- 单元测试 + 集成测试

**验收标准**:
- 能添加/编辑/删除 MySQL/Oracle/SQL Server/PostgreSQL 连接
- 密码安全存储
- 测试连接成功/失败有明确提示

---

### Phase 2: 模板系统与任务配置（Week 3-4）

**目标**: 实现模板管理与任务配置功能

**交付物**:
- 内置模板（7个）
- 模板导入/导出
- 任务配置 GUI 页面
- 参数校验逻辑
- 配置快照机制
- 单元测试

**验收标准**:
- 能查看内置模板
- 能导入自定义模板并校验
- 能配置完整任务参数
- 参数错误有清晰提示

---

### Phase 3: 工具适配器与执行编排（Week 5-7）

**目标**: 实现三个工具的适配器与核心执行逻辑

**交付物**:
- Sysbench 适配器（完整实现）
- Swingbench 适配器（完整实现）
- HammerDB 适配器（完整实现）
- 执行编排器
- 状态机与取消逻辑
- 日志实时采集
- 指标实时采集
- 运行监控 GUI 页面
- 单元测试 + 集成测试

**验收标准**:
- 能跑通三个工具的完整流程
- 能启动/停止/取消任务
- 能实时查看日志和指标
- 异常场景有处理

---

### Phase 4: 结果存储与历史记录（Week 8）

**目标**: 实现结果持久化与历史记录查询

**交付物**:
- 结果解析与存储
- 时间序列数据存储
- 历史记录 GUI 页面
- 详情查看与日志回放
- 结果目录管理
- 单元测试

**验收标准**:
- 每次运行结果完整保存
- 能查看历史记录列表
- 能查看某次运行的详情

---

### Phase 5: 报告生成与导出（Week 9）

**目标**: 实现多格式报告导出

**交付物**:
- Markdown 报告生成
- HTML 报告转换
- JSON 报告导出
- PDF 报告导出
- 图表生成
- 报告导出 GUI 页面
- 单元测试

**验收标准**:
- 能导出四种格式报告
- 报告包含关键信息
- 文件命名规范正确

---

### Phase 6: 对比分析功能（Week 10）

**目标**: 实现完整的结果对比系统

**交付物**:
- A/B 对比视图
- 多次基线对比
- 趋势分析
- 图表对比
- 对比报告导出
- 对比 GUI 页面
- 单元测试

**验收标准**:
- 能选择两次运行并对比
- 能显示差异指标
- 能绘制对比图表

---

### Phase 7: 设置页面与优化（Week 11）

**目标**: 完善设置功能与性能优化

**交付物**:
- 设置 GUI 页面
- 工具路径配置与检测
- 全局配置项管理
- GUI 性能优化
- 内存占用优化
- 单元测试

**验收标准**:
- 能配置所有全局设置
- 工具版本自动检测
- GUI 流畅无卡顿

---

### Phase 8: 文档与测试完善（Week 12）

**目标**: 完善文档与测试覆盖

**交付物**:
- 用户手册
- 架构文档
- API 文档
- 单元测试覆盖率 > 80%
- 集成测试场景完整
- E2E 测试脚本
- 性能测试报告

**验收标准**:
- 文档完整且易懂
- 测试覆盖所有核心路径
- CI/CD 通过率 100%

---

## 9. 验收标准

### 9.1 核心功能验收

| ID | 验收标准 | 测试方法 |
|----|---------|---------|
| AC-001 | 能配置并测试 MySQL 数据库连接 | 手动测试 |
| AC-002 | 能配置并测试 Oracle 数据库连接 | 手动测试 |
| AC-003 | 能配置并测试 SQL Server 数据库连接 | 手动测试 |
| AC-004 | 能配置并测试 PostgreSQL 数据库连接 | 手动测试 |
| AC-005 | 能运行 Sysbench 对 MySQL 进行 OLTP 读写压测 | E2E 测试 |
| AC-006 | 能运行 Swingbench 对 Oracle 进行压测 | E2E 测试 |
| AC-007 | 能运行 HammerDB 对 SQL Server 进行压测 | E2E 测试 |
| AC-008 | 实时监控页面能显示 TPS、延迟、错误率 | E2E 测试 |
| AC-009 | 历史记录页面能显示所有运行记录 | 手动测试 |
| AC-010 | 能选择两次运行并进行对比分析 | 手动测试 |
| AC-011 | 能导出 Markdown 格式报告 | E2E 测试 |
| AC-012 | 能导出 HTML 格式报告 | E2E 测试 |
| AC-013 | 能导出 JSON 格式报告 | E2E 测试 |
| AC-014 | 能导出 PDF 格式报告 | E2E 测试 |
| AC-015 | 报告包含环境信息、参数、指标、图表 | 检查报告内容 |
| AC-016 | 工具未安装时有清晰提示 | 手动测试 |
| AC-017 | 数据库连接失败时有明确错误信息 | 手动测试 |
| AC-018 | 优雅停止功能正常工作（SIGTERM → SIGKILL） | E2E 测试 |
| AC-019 | 长时间运行（24h）无崩溃、无内存泄漏 | 性能测试 |
| AC-020 | GUI 响应流畅（< 100ms） | 性能测试 |

---

## 10. 依赖与风险

### 10.1 外部依赖

| 依赖 | 版本 | 用途 | 许可证 |
|------|------|------|--------|
| **Fyne** | v2.x | GUI 框架 | Apache 2.0 |
| **modernc.org/sqlite** | latest | SQLite 驱动 | BSD 3-Clause |
| **zalando/go-keyring** | latest | Keyring 集成 | Apache 2.0 |
| **pandoc**（可选） | any | Markdown → PDF | GPL-2.0 |
| **wkhtmltopdf**（可选） | any | HTML → PDF | LGPLv3 |

### 10.2 技术风险

| 风险 | 影响 | 概率 | 缓解措施 |
|------|------|------|---------|
| Fyne GUI 性能问题 | 高 | 中 | 早期原型验证；goroutine 异步加载 |
| 工具输出格式变化 | 高 | 中 | 灵活解析器；多版本兼容；失败时保留原始输出 |
| Swingbench/HammerDB 路径依赖 | 高 | 高 | 提供清晰的安装文档；自动检测提示 |
| SQLite 并发写入限制 | 中 | 低 | 写入队列；批量插入；WAL 模式 |
| PDF 转换失败 | 中 | 中 | 提供 Markdown/HTML 备选；清晰的依赖提示 |

---

## 11. 附录

### 11.1 术语表

| 术语 | 定义 |
|------|------|
| **TPS** | Transactions Per Second，每秒事务数 |
| **QPS** | Queries Per Second，每秒查询数 |
| **P95/P99** | 95th/99th percentile latency，95/99分位延迟 |
| **OLTP** | Online Transaction Processing，联机事务处理 |
| **Workload** | 工作负载，指具体的压测场景 |
| **Baseline** | 基线，作为对比基准的测试结果 |

### 11.2 参考资料

- [Fyne 官方文档](https://docs.fyne.io/)
- [Sysbench GitHub](https://github.com/akopytov/sysbench)
- [Swingbench 官方文档](https://www.dominicgiles.com/swingbench/)
- [HammerDB 官方网站](https://www.hammerdb.com/)
- [SQLite 官方文档](https://www.sqlite.org/docs.html)
- [go-keyring GitHub](https://github.com/zalando/go-keyring)

---

## 变更记录

| 版本 | 日期 | 变更内容 | 作者 |
|------|------|---------|------|
| 1.0.0 | 2026-01-27 | 初始版本，包含完整 MVP 需求 | Claude |

---

**文档结束**
