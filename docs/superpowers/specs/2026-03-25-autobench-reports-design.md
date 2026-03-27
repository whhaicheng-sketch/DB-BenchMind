# AutoBench 架构重构设计规范 - Reports 持久化层

**文档版本**: v1.0
**创建日期**: 2026-03-25
**状态**: 已确认
**Phase**: 1

---

## 目录

1. [概述](#1-概述)
2. [命名规范](#2-命名规范)
3. [架构设计](#3-架构设计)
4. [数据模型](#4-数据模型)
5. [执行流程](#5-执行流程)
6. [持久化规范](#6-持久化规范)
7. [Phase 1 / Phase 2 边界](#7-phase-1--phase-2-边界)
8. [验收标准](#8-验收标准)
9. [兼容性保证](#9-兼容性保证)
10. [实施任务拆解](#10-实施任务拆解)

---

## 1. 概述

### 1.1 目标

为 DB-BenchMind 新增 Reports 持久化层，实现完整的压测报告闭环：

```
Connections → Templates → Benchmark → AutoBench → Reports
    │            │            │            │           │
    │            │            │            │           └─ 存储全部数据
    │            │            │            └─ 批量调度
    │            │            └─ 单次执行
    │            └─ 压测方案
    └─ 目标数据库
```

### 1.2 核心原则

| 原则 | 说明 |
|------|------|
| Doc-First | 先更新文档再写代码 |
| 最小侵入 | 新增包装器，不改现有执行链路 |
| Additive Only | 新表/新目录/新字段均为增量，不破坏现有功能 |
| 可恢复 | 中断后可基于 progress.md / progress.json 继续 |
| 数据完整 | 所有压测数据必须持久化，不丢失 |

### 1.3 设计决策记录

| ID | 主题 | 决策 |
|----|------|------|
| D001 | 导航重命名策略 | 仅 UI 标签变更，内部 ID 不变 |
| D002 | 存储模式 | 混合：SQLite 元数据 + 文件系统完整数据 |
| D003 | 监控采集 | 分阶段：Phase 1 复用现有，Phase 2 扩展 |
| D004 | 架构模式 | 最小侵入包装器 (ReportCollector) |
| D005 | suite_id 策略 | standalone = "standalone"，不用 null |
| D006 | 持久化失败策略 | 不破坏主执行结果，记录错误可追踪 |
| D007 | suites 表字段 | 使用 suite_manifest_json_path |
| D008 | SuiteItem 持久化 | 使用 suite_manifest.json |
| D009 | JSON 导出策略 | Phase 1 运行时动态打包 |
| D010 | 原始输出存储 | stdout/stderr 内联 raw.json |
| D011 | JSON Schema 版本 | 所有 JSON 必须包含 schema_version |
| D012 | 报告来源类型 | reports.source_type 显式字段 |
| D013 | 监控取数路径 | 直接访问 SystemCollector，不依赖前端事件 |
| D014 | 持久化失败处理 | 记录日志 + PersistError 字段 |
| D015 | 写入顺序 | 文件先写，SQLite 后写 |

---

## 2. 命名规范

### 2.1 导航标签

| 内部 ID (Legacy Key) | 显示名称 (Display Name) | 变更范围 |
|----------------------|-------------------------|----------|
| `connections` | Connections | 无变更 |
| `templates` | Templates | 无变更 |
| `tasks` | **Benchmark** | 仅 label 文本 |
| `autobench` | AutoBench | 无变更 |
| `history` | **Reports** | 仅 label 文本 |

**实现位置**: 仅修改 `frontend/src/constants/navigationTabs.mjs` 中的 `label` 字段

### 2.2 suite_id 策略

| 场景 | suite_id 值 |
|------|-------------|
| 单次 Benchmark | `"standalone"` (字符串常量，非 null) |
| AutoBench 批量 | UUID (AutoBench.Suite.ID) |

**查询示例**:
```sql
-- 单次 Benchmark 报告
SELECT * FROM reports WHERE suite_id = 'standalone'

-- AutoBench 批量报告
SELECT * FROM reports WHERE suite_id = '{uuid}' AND suite_id != 'standalone'
```

---

## 3. 架构设计

### 3.1 模块职责

```
┌─────────────────────────────────────────────────────────────────┐
│                        UI Layer                                  │
│  (仅 navigationTabs.mjs 标签文本变更，路由 ID 不变)              │
└─────────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────▼───────────────────────────────────┐
│                      Usecase Layer (现有不动)                    │
├──────────┬──────────┬───────────┬───────────┬──────────────────┤
│Connection│Template  │ Benchmark │ AutoBench │   Report         │
│ Usecase  │ Usecase  │  Usecase  │   Runner  │   Usecase        │
│ (现有)   │ (现有)   │  (现有)   │  (现有)   │   (新增)         │
└──────────┴──────────┴───────────┴───────────┴──────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                  ReportCollector (新增包装器)                    │
│                                                                 │
│  输入:                                                │
│  ├─ ctx context.Context                                          │
│  ├─ runFn func() (Run, error)                                    │
│  └─ ReportContext { SuiteID, ConnectionID, TemplateID, Tags }   │
│                                                                 │
│  输出:                                                │
│  ├─ ReportID string                                              │
│  ├─ ReportPaths { MetricsJSON, MonitoringJSON, RawJSON, ... }   │
│  ├─ Summary { Status, TPM, TPS, Latency, Errors }               │
│  └─ PersistError error (可选)                                    │
│                                                                 │
│  禁止事项:                                                        │
│  ✗ 不替换 BenchmarkRunner / Adapter / Execution                  │
│  ✗ 不重构现有执行链路                                             │
│  ✗ 不依赖前端事件监听作为主持久化路径                              │
└─────────────────────────────────────────────────────────────────┘
                              │
              ┌───────────────┴───────────────┐
              ▼                               ▼
      ┌──────────────┐              ┌─────────────────┐
      │   SQLite     │              │   Filesystem    │
      │ reports      │              │ reports/        │
      │ suites       │              │ ├─ standalone/  │
      └──────────────┘              │ └─ {suite_id}/  │
                                    └─────────────────┘
```

### 3.2 数据流向

```
Benchmark 执行流程（现有不动）:
  User → BenchmarkUsecase → BenchmarkRunner → Adapter → Run

新增包装流程:
  User → BenchmarkUsecase → ReportCollector → BenchmarkRunner
                                  │
                                  ├─ 执行 runFn()
                                  ├─ 收集 Run.Summary
                                  ├─ 收集 Monitoring (SystemCollector 底层)
                                  ├─ 持久化文件
                                  ├─ 写入 SQLite
                                  └─ 返回 ReportResult

AutoBench 流程（现有不动 + 新增）:
  User → AutoBenchRunner → ReportCollector → BenchmarkRunner
                │
                ├─ 每个 SuiteItem → 生成 report_id
                ├─ 更新 suite_manifest.json
                └─ 最终 SuiteReport 汇总 report_ids
```

---

## 4. 数据模型

### 4.1 SQLite 表结构

#### 4.1.1 reports 表

```sql
CREATE TABLE IF NOT EXISTS reports (
    id              TEXT PRIMARY KEY,           -- UUID
    suite_id        TEXT NOT NULL,              -- 'standalone' 或 UUID
    suite_item_id   TEXT,                       -- AutoBench SuiteItem.ID
    source_type     TEXT NOT NULL,              -- 'benchmark' | 'autobench'

    -- 执行上下文
    connection_id   TEXT NOT NULL,
    connection_name TEXT,
    database_type   TEXT NOT NULL,
    template_id     TEXT,
    template_name   TEXT,

    -- 时间信息
    started_at      DATETIME NOT NULL,
    ended_at        DATETIME,
    duration_ms     INTEGER,

    -- 执行状态
    status          TEXT NOT NULL,
    error_message   TEXT,

    -- 核心指标快照
    tpm             REAL,
    tps             REAL,
    qps             REAL,                       -- 预留
    throughput      REAL,                       -- 预留
    latency_avg_ms  REAL,
    latency_p95_ms  REAL,
    latency_p99_ms  REAL,
    error_count     INTEGER,

    -- 文件路径
    metrics_json_path    TEXT,
    monitoring_json_path TEXT,
    raw_json_path        TEXT,
    report_html_path     TEXT,
    summary_json_path    TEXT,

    -- 元数据
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    tags            TEXT
);

-- 索引
CREATE INDEX idx_reports_suite_id ON reports(suite_id);
CREATE INDEX idx_reports_source_type ON reports(source_type);
CREATE INDEX idx_reports_connection_id ON reports(connection_id);
CREATE INDEX idx_reports_status ON reports(status);
CREATE INDEX idx_reports_started_at ON reports(started_at DESC);
CREATE INDEX idx_reports_suite_item_id ON reports(suite_item_id);
```

#### 4.1.2 suites 表

```sql
CREATE TABLE IF NOT EXISTS suites (
    id              TEXT PRIMARY KEY,
    name            TEXT,

    -- 执行策略快照
    execution_mode         TEXT DEFAULT 'serial',
    failure_policy         TEXT DEFAULT 'continue_by_connection',
    cleanup_enabled        INTEGER DEFAULT 1,

    -- Suite 全量快照路径
    suite_manifest_json_path TEXT,

    -- 状态
    status          TEXT NOT NULL,
    started_at      DATETIME,
    ended_at        DATETIME,

    -- 汇总统计
    total_items         INTEGER DEFAULT 0,
    completed_items     INTEGER DEFAULT 0,
    success_items       INTEGER DEFAULT 0,
    failed_items        INTEGER DEFAULT 0,
    skipped_items       INTEGER DEFAULT 0,

    -- 文件路径
    suite_report_json_path  TEXT,
    suite_report_html_path  TEXT,

    -- 元数据
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 索引
CREATE INDEX idx_suites_status ON suites(status);
CREATE INDEX idx_suites_started_at ON suites(started_at DESC);
```

### 4.2 文件系统结构

```
reports/
├── standalone/                          # 单次 Benchmark
│   └── {report_id}/
│       ├── metrics.json                 # 必须
│       ├── monitoring.json              # 必须
│       ├── raw.json                     # 必须
│       ├── report.html                  # 必须
│       └── summary.json                 # 必须
│
└── {suite_id}/                          # AutoBench 批量
    ├── {report_id_1}/
    │   ├── metrics.json
    │   ├── monitoring.json
    │   ├── raw.json
    │   ├── report.html
    │   └── summary.json
    ├── {report_id_2}/
    │   └── ...
    ├── suite_manifest.json              # 必须：完整 SuiteItem 快照
    ├── suite_report.json                # 可选
    └── suite_report.html                # 可选
```

### 4.3 JSON 文件 Schema

#### 4.3.1 metrics.json

```json
{
  "schema_version": "v1",
  "report_id": "uuid",
  "suite_id": "standalone|uuid",
  "suite_item_id": "uuid|null",
  "generated_at": "2026-03-25T10:00:00Z",

  "benchmark": {
    "connection_id": "uuid",
    "connection_name": "MySQL Prod",
    "database_type": "mysql",
    "template_id": "uuid",
    "template_name": "oltp_read_write",
    "parameters": {}
  },

  "execution": {
    "status": "completed",
    "started_at": "2026-03-25T10:00:00Z",
    "ended_at": "2026-03-25T10:01:00Z",
    "duration_ms": 60000
  },

  "summary": {
    "tpm": 15000.5,
    "tps": 250.5,
    "qps": null,
    "throughput": null,
    "latency_avg_ms": 15.8,
    "latency_p95_ms": 25.3,
    "latency_p99_ms": 45.2,
    "latency_max_ms": 120.5,
    "error_count": 0,
    "error_rate": 0.0
  },

  "percentiles": {
    "p50": 12.5,
    "p90": 22.1,
    "p95": 25.3,
    "p99": 45.2
  },

  "phase_durations_ms": {
    "prepare": 5000,
    "warmup": 10000,
    "run": 60000,
    "cleanup": 3000
  }
}
```

#### 4.3.2 monitoring.json

```json
{
  "schema_version": "v1",
  "report_id": "uuid",
  "generated_at": "2026-03-25T10:01:00Z",

  "sample_interval_ms": 1000,
  "total_samples": 60,

  "time_series": {
    "timestamps": [1711363200, 1711363201],
    "tpm": [14500.0, 14800.0],
    "tps": [241.5, 246.5],
    "latency_ms": [15.2, 15.5],
    "errors": [0, 0]
  },

  "system": {
    "timestamps": [1711363200, 1711363201],
    "cpu_percent": [45.2, 48.5],
    "disk_read_bps": [1024000, 2048000],
    "disk_write_bps": [512000, 1024000]
  },

  "phase_2_reserved": {
    "memory_percent": null,
    "net_rx_bps": null,
    "net_tx_bps": null,
    "load_avg_1m": null,
    "load_avg_5m": null,
    "load_avg_15m": null,
    "db_connections": null,
    "db_active_sessions": null
  }
}
```

#### 4.3.3 raw.json

```json
{
  "schema_version": "v1",
  "report_id": "uuid",
  "generated_at": "2026-03-25T10:01:00Z",

  "adapter_type": "sysbench|swingbench|hammerdb",
  "adapter_version": "1.0.0",

  "stdout": "[完整 stdout 输出]",
  "stderr": "[完整 stderr 输出]",

  "exit_code": 0,

  "parsed_result": {
    "raw": {
      "transactions": 15000,
      "queries": 150000,
      "errors": 0,
      "reconnects": 0
    }
  }
}
```

**注意**: Phase 1 不承诺额外 txt/log 文件，stdout/stderr 内联存储。

#### 4.3.4 summary.json

```json
{
  "schema_version": "v1",
  "report_id": "uuid",
  "suite_id": "standalone|uuid",
  "generated_at": "2026-03-25T10:01:00Z",

  "display": {
    "title": "MySQL Prod - oltp_read_write",
    "subtitle": "4 threads, 60s duration",
    "status": "completed"
  },

  "key_metrics": {
    "tpm": 15000.5,
    "tps": 250.5,
    "latency_avg_ms": 15.8,
    "latency_p95_ms": 25.3
  },

  "duration_display": "1m 0s",
  "database_type": "mysql",
  "connection_name": "MySQL Prod"
}
```

#### 4.3.5 suite_manifest.json

```json
{
  "schema_version": "v1",
  "suite_id": "uuid",
  "generated_at": "2026-03-25T10:00:00Z",

  "suite_info": {
    "name": "MySQL Cluster Benchmark",
    "execution_mode": "serial",
    "failure_policy": "continue_by_connection",
    "cleanup_enabled": true
  },

  "selected_connection_ids": ["conn-1", "conn-2"],

  "items": [
    {
      "id": "item-1",
      "connection_id": "conn-1",
      "database_type": "mysql",
      "profile_type": "test",
      "template_id": "tmpl-1",
      "status": "success",
      "report_id": "report-uuid-1",
      "started_at": "2026-03-25T10:00:00Z",
      "ended_at": "2026-03-25T10:01:00Z",
      "error_message": null
    }
  ],

  "statistics": {
    "total_items": 6,
    "pending": 0,
    "running": 0,
    "success": 5,
    "failed": 1,
    "skipped": 0
  }
}
```

### 4.4 关系模型

```
┌─────────────────────────────────────────────────────────────────┐
│                         suites 表                                │
│  (仅 AutoBench 场景使用，standalone 不在此表)                    │
│  ├─ suite_manifest_json_path → suite_manifest.json              │
└───────────────────────────┬─────────────────────────────────────┘
                            │ 1:N (通过 suite_id)
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│                         reports 表                               │
│  ├─ suite_id = 'standalone' → 单次 Benchmark                     │
│  ├─ suite_id = UUID → AutoBench SuiteItem                        │
│  ├─ source_type = 'benchmark' | 'autobench'                      │
│  └─ 文件路径 → reports/{suite_id}/{report_id}/*                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## 5. 执行流程

### 5.1 Benchmark 单次执行时序

```
User → BenchmarkUsecase.Run()
         │
         └─→ ReportCollector.CollectAndPersist(ctx, runFn, ReportContext)
                    │
                    ├─→ runFn() → BenchmarkRunner → Adapter → Run
                    │
                    ├─→ 收集 Run.Summary
                    ├─→ 收集 SystemCollector 数据
                    ├─→ 写入 reports/standalone/{report_id}/*
                    ├─→ 写入 SQLite reports 表
                    │
                    └─→ 返回 ReportResult
```

### 5.2 AutoBench 批量执行时序

```
User → AutoBenchRunner.RunSuite(suite_id)
         │
         ├─→ 写入 suite_manifest.json
         │
         └─→ 对每个 SuiteItem:
                    │
                    ├─→ ReportCollector.CollectAndPersist(...)
                    │         │
                    │         └─→ 返回 report_id
                    │
                    └─→ 更新 suite_manifest.json

         └─→ 所有 Item 完成:
                    │
                    ├─→ 生成 suite_report.json/html
                    └─→ 更新 suites 表
```

### 5.3 监控数据取数

**取数路径**: 直接访问 SystemCollector 底层，不依赖前端 MonitorBinding

```go
type MonitoringSnapshotProvider interface {
    GetSystemHistory() SystemHistoryDTO
    GetCPUPercent() float64
    GetDiskIOBps() (read, write float64)
}

// 实现: 包装现有 SystemCollector
type SystemCollectorAdapter struct {
    collector *collector.SystemCollector
}
```

**时间对齐**:
1. 记录 benchmark 开始时间
2. 记录 benchmark 结束时间
3. 按 [start, end] 过滤 history buffer

### 5.4 持久化顺序

```
严格顺序:
1. 创建目录 reports/{suite_id}/{report_id}/
2. 写入文件: metrics.json, monitoring.json, raw.json, summary.json, report.html
3. 写入 SQLite: INSERT reports
4. (AutoBench) 更新 suite_manifest.json

失败处理:
├─ 文件写入失败 → 不写 SQLite，返回 PersistError
├─ SQLite 写入失败 → 保留文件（孤儿报告，可定期清理）
└─ 最终一致性
```

### 5.5 失败降级

```
ReportCollector 降级保证:
├─ 持久化失败不破坏 Run 主结果
├─ 错误记录在日志 + PersistError 字段
└─ 前端可显示警告但不阻塞

日志示例:
slog.Error("report persistence failed",
    "report_id", reportID,
    "stage", "file_write",  // dir_create | file_write | sqlite_write
    "error", err)
```

### 5.6 Reports 读取链路

| 页面 | 数据来源 |
|------|----------|
| 列表页 | SQLite reports 表 (summary 字段) |
| 详情页-基本信息 | SQLite + summary.json |
| 详情页-指标 | metrics.json |
| 详情页-图表 | monitoring.json |
| 详情页-原始 | raw.json |
| JSON 导出 | 运行时打包 metrics + monitoring + raw |
| HTML 导出 | report.html 文件 |

---

## 6. 持久化规范

### 6.1 文件写入

```go
func writeJSON(path string, data interface{}) error {
    tmpPath := path + ".tmp"
    f, err := os.Create(tmpPath)
    if err != nil {
        return err
    }
    defer f.Close()

    enc := json.NewEncoder(f)
    enc.SetIndent("", "  ")
    if err := enc.Encode(data); err != nil {
        os.Remove(tmpPath)
        return err
    }

    return os.Rename(tmpPath, path)  // 原子操作
}
```

### 6.2 初始化时机

```
应用启动时:
├─ 数据库: CREATE TABLE IF NOT EXISTS reports / suites
├─ 索引: CREATE INDEX IF NOT EXISTS
└─ 目录: os.MkdirAll("reports/")

不需要数据迁移 (新表)
```

---

## 7. Phase 1 / Phase 2 边界

### 7.1 Phase 1 范围

| 能力 | 状态 |
|------|------|
| CPU 使用率 | ✓ 复用现有 |
| 磁盘 IO | ✓ 复用现有 |
| TPM/TPS | ✓ 复用现有 |
| Latency | ✓ 复用现有 |
| Error Count | ✓ 复用现有 |
| 报告持久化 | ✓ 新增 |
| Reports 列表/详情/导出 | ✓ 新增 |

### 7.2 Phase 2 预留

| 能力 | 状态 | JSON 字段预留 |
|------|------|---------------|
| 内存使用率 | Phase 2 | monitoring.json → phase_2_reserved.memory_percent |
| 网络 IO | Phase 2 | monitoring.json → phase_2_reserved.net_rx/tx_bps |
| Load Average | Phase 2 | monitoring.json → phase_2_reserved.load_avg_* |
| DB 连接数 | Phase 2 | monitoring.json → phase_2_reserved.db_connections |
| DB 活跃会话 | Phase 2 | monitoring.json → phase_2_reserved.db_active_sessions |

**约束**: Phase 1 不允许为补齐 Phase 2 而重构现有主链路。

---

## 8. 验收标准

### 8.1 必须通过项

| # | 验收项 | 验证方法 |
|---|--------|----------|
| 1 | Benchmark 执行后生成持久化产物 | 检查 reports/standalone/{id}/ 目录 |
| 2 | AutoBench 结果进入 Reports | 检查 SQLite reports 表有记录 |
| 3 | Reports 列表显示 summary | 前端 Reports 页面显示列表 |
| 4 | Reports 详情显示图表 | 点击报告显示 metrics 图表 |
| 5 | JSON 可导出 | 点击导出下载 JSON |
| 6 | HTML 可导出 | 点击导出下载/打开 HTML |
| 7 | 保存 metrics 时间序列 | monitoring.json 有 time_series |
| 8 | 保存 monitoring 数据 | monitoring.json 有 system |
| 9 | 保存 raw 输出 | raw.json 有 stdout/stderr |
| 10 | 不破坏已有功能 | make test 全部通过 |

### 8.2 回归测试

```bash
make test-backend   # 所有后端测试通过
make test-frontend  # 所有前端测试通过
```

---

## 9. 兼容性保证

### 9.1 不破坏现有功能

| 现有功能 | 保证 |
|----------|------|
| BenchmarkUsecase.Run() | 行为不变，返回值不变 |
| AutoBenchRunner.RunSuite() | 行为不变，返回值不变 |
| MonitorBinding | 不修改，不依赖 |
| SystemCollector | 不修改，只读取 |
| 现有 SQLite 表 | 不修改结构 |

### 9.2 新增能力开关

```go
type ReportCollectorConfig struct {
    Enabled         bool
    ReportsDir      string
    MaxRetries      int
    RetryDelay      time.Duration
}
```

---

## 10. 实施任务拆解

### M0 - 设计规范

- [x] T0.1 需求澄清与方案确认
- [x] T0.2 架构概览文档定稿
- [x] T0.3 数据模型设计定稿
- [x] T0.4 执行流程设计定稿
- [x] T0.5 持久化规范定稿
- [x] T0.6 设计文档写入 docs/superpowers/specs/

### M1 - 导航 UI 标签重命名

- [ ] T1.1 更新 navigationTabs.mjs 标签文本
- [ ] T1.2 前端测试验证

### M2 - Reports 数据模型与 SQLite 表

- [ ] T2.1 定义 Report 领域模型
- [ ] T2.2 SQLite reports/suites 表结构
- [ ] T2.3 数据库初始化逻辑
- [ ] T2.4 后端单元测试

### M3 - ReportCollector 包装器实现

- [ ] T3.1 ReportCollector 接口定义
- [ ] T3.2 收集 Benchmark 执行结果
- [ ] T3.3 收集监控数据 (SystemCollector 底层)
- [ ] T3.4 文件系统持久化逻辑
- [ ] T3.5 SQLite 元数据写入
- [ ] T3.6 后端单元测试

### M4 - Reports Usecase 与 API 绑定

- [ ] T4.1 ReportUsecase 实现
- [ ] T4.2 Wails bindings 暴露
- [ ] T4.3 后端单元测试

### M5 - Reports 前端页面

- [ ] T5.1 Reports 列表页 (HistoryTab 改造)
- [ ] T5.2 Reports 详情页
- [ ] T5.3 图表组件集成
- [ ] T5.4 JSON/HTML 导出功能
- [ ] T5.5 前端单元测试

### M6 - AutoBench 集成 report_id

- [ ] T6.1 SuiteItem 关联 report_id
- [ ] T6.2 suite_manifest.json 生成与更新
- [ ] T6.3 后端单元测试

### M7 - 单次 Benchmark 集成 report

- [ ] T7.1 单次执行生成 standalone report
- [ ] T7.2 suite_id = "standalone" 策略实现
- [ ] T7.3 后端单元测试

### M8 - 验收与兼容性回归

- [ ] T8.1 后端回归测试
- [ ] T8.2 前端回归测试
- [ ] T8.3 集成测试
- [ ] T8.4 文档同步更新

---

## 附录

### A. 断点恢复

中断后恢复:
1. 读取 `docs/AutoBench/progress.md` → 当前模块/任务
2. 读取 `docs/AutoBench/progress.json` → done_tasks / decisions
3. 从 pending_tasks 取下一个任务继续

### B. 关键文件路径

| 文件 | 路径 |
|------|------|
| 设计文档 | docs/superpowers/specs/2026-03-25-autobench-reports-design.md |
| 进度文档 | docs/AutoBench/progress.md |
| 进度 JSON | docs/AutoBench/progress.json |
| 报告目录 | reports/ |

---

**文档结束**
