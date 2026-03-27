# AutoBench 可用化实现设计规范

**文档版本**: v2.0
**创建日期**: 2026-03-27
**状态**: ✅ 已实现
**Phase**: AutoBench Usable Implementation (非 Phase 2 监控扩展)

---

## 1. 当前真实状态

### 1.1 Phase 1 - Reports 持久化层 ✅ 已完成

| 模块 | 状态 | 说明 |
|------|------|------|
| Reports 数据模型 | ✅ 完成 | Report/Suite 领域模型、SQLite 表结构 |
| ReportCollector | ✅ 完成 | 包装器实现，持久化 metrics/monitoring/raw/summary |
| ReportUsecase | ✅ 完成 | ListReports/GetReport/GetReportMetrics/ListSuites/GetSuite |
| Reports 前端页面 | ✅ 完成 | ReportsTab 列表/详情/导出 |
| 单次 Benchmark 集成 | ✅ 完成 | suite_id="standalone"，自动生成 report |
| suite_manifest.json | ✅ 完成 | SuiteManifest 模型和 Writer |

### 1.2 Phase 2 - AutoBench 可用化 ✅ 已完成

| 模块 | 状态 | 说明 |
|------|------|------|
| AutoBench 真实数据源 | ✅ 完成 | 从真实 Connections API 获取数据 |
| Create Suite 功能 | ✅ 完成 | CreateSuite API，写入 suites 表 |
| Start Suite 功能 | ✅ 完成 | StartSuite API，调用 BenchmarkUseCase |
| Suite 状态展示 | ✅ 完成 | metrics/progress/items 展示 |
| Reports 联动 | ✅ 完成 | viewReport/goToReports 导航 |
| 前端测试 | ✅ 完成 | 19 个测试全部通过 |

---

## 命名约定

| 用户可见术语 | 内部字段 | 说明 |
|-------------|---------|------|
| Profiles / Benchmark Types | `profile_type` | 模板分类：test, cpu_bound, io_bound |
| Templates | - | 实际压测配置，由系统根据 profile_type 自动匹配 |
| Connections | `connection_id` | 数据库连接 |

**重要说明**：用户在 AutoBench 选择的是 Profile Types（测试类型），系统根据 (connection.db_type, profile_type) 自动匹配对应的 Template。

---

## 2. 本轮目标

### 2.1 目标描述

把 AutoBench 从"占位/草稿 UI"实现为"可创建 Suite、可执行 Suite、可写入 Reports 的真实可用功能"。

### 2.2 核心原则

| 原则 | 说明 |
|------|------|
| 不破坏已有功能 | 现有 Benchmark 单次执行、Reports 列表/详情/导出不受影响 |
| 复用现有执行能力 | 只调用现有 Benchmark 执行链路，不重写执行引擎 |
| Additive Only | 新增表字段/API/组件，不破坏现有结构 |
| Doc-First | 先修正文档，再实现代码 |
| 可恢复 | 中断后可基于 progress.md / progress.json 继续 |

### 2.3 本轮不做

| 范围 | 说明 |
|------|------|
| Phase 2 监控扩展 | Memory/Network/Load Average/DB 连接数/活跃会话 |
| 新执行引擎 | 复用现有 BenchmarkRunner/Adapter |
| stdout.log/stderr.log | 继续使用 raw.json 内联 stdout/stderr |

---

## 3. 术语修正

### 3.1 删除的旧草稿术语

| 旧术语 | 问题 |
|--------|------|
| placeholder connections | 不反映真实数据源 |
| static profiles | 不反映真实模板 |
| local-only wizard | 不反映真实后端调用 |
| selected_profiles | 应使用 selected_template_ids |

### 3.2 新的真实术语

| 新术语 | 说明 |
|--------|------|
| selected_connection_ids | 从真实 Connections 模块获取 |
| selected_template_ids | 从真实 Templates 模块获取 |
| execution_mode | serial/parallel（Phase 1 仅 serial） |
| failure_policy | continue_by_connection/stop_on_failure |
| suite_manifest_json_path | SuiteItem 状态快照锚点 |

---

## 4. suite_id 策略

### 4.1 统一策略

| 场景 | suite_id 值 |
|------|-------------|
| 单次 Benchmark | `"standalone"` (字符串常量) |
| AutoBench Suite | UUID (Suite.ID) |

**重要**：standalone 报告不写入 suites 表。

### 4.2 查询示例

```sql
-- 单次 Benchmark 报告
SELECT * FROM reports WHERE suite_id = 'standalone'

-- AutoBench 批量报告
SELECT * FROM reports WHERE suite_id = '{uuid}' AND suite_id != 'standalone'
```

---

## 5. Suite 恢复机制

### 5.1 设计决策

**不建立独立的 suite_items 表**。SuiteItem 状态通过以下方式恢复：

1. **suites 表** 保存 `suite_manifest_json_path`
2. **suite_manifest.json** 是恢复完整 SuiteItem 状态的唯一快照锚点
3. **reports 表** 通过 `suite_id` + `suite_item_id` 关联

### 5.2 恢复流程

```
1. 从 suites 表读取 suite_manifest_json_path
2. 加载 suite_manifest.json 获取完整 items 列表
3. 每个 item 通过 suite_item_id 查询 reports 表获取 report_id
4. 完整恢复 Suite 视图
```

---

## 6. JSON 导出策略

### 6.1 继续沿用 Phase 1 策略

| 策略 | 说明 |
|------|------|
| 运行时动态打包 | 不要求单独 json_report_path |
| metrics.json | 核心指标 |
| monitoring.json | 系统监控 |
| raw.json | stdout/stderr 内联 |
| summary.json | 摘要 |
| report.html | 可视化报告 |

### 6.2 标准产物

每个 report 目录 (`reports/{suite_id}/{report_id}/`) 包含：

```
├── metrics.json       # 核心指标
├── monitoring.json    # 系统监控
├── raw.json           # 原始输出（含 stdout/stderr）
├── summary.json       # 摘要
└── report.html        # 可视化报告
```

**不新增** stdout.log / stderr.log 文件。

---

## 7. 数据模型

### 7.1 suites 表（已存在，字段确认）

```sql
CREATE TABLE IF NOT EXISTS suites (
    id              TEXT PRIMARY KEY,
    name            TEXT,
    execution_mode  TEXT DEFAULT 'serial',
    failure_policy  TEXT DEFAULT 'continue_by_connection',
    cleanup_enabled INTEGER DEFAULT 1,
    suite_manifest_json_path TEXT,
    status          TEXT NOT NULL,
    started_at      TEXT,
    ended_at        TEXT,
    total_items         INTEGER DEFAULT 0,
    completed_items     INTEGER DEFAULT 0,
    success_items       INTEGER DEFAULT 0,
    failed_items        INTEGER DEFAULT 0,
    skipped_items       INTEGER DEFAULT 0,
    suite_report_json_path  TEXT,
    suite_report_html_path  TEXT,
    created_at      TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

### 7.2 Suite 领域模型（新增字段）

```go
type Suite struct {
    ID                    string
    Name                  string
    SelectedConnectionIDs []string  // 新增：选中的连接 ID
    SelectedTemplateIDs   []string  // 新增：选中的模板 ID
    ExecutionMode         string    // serial/parallel
    FailurePolicy         string    // continue_by_connection/stop_on_failure
    CleanupEnabled        bool
    SuiteManifestJSONPath string
    Status                SuiteStatus
    StartedAt             *time.Time
    EndedAt               *time.Time
    TotalItems            int
    CompletedItems        int
    SuccessItems          int
    FailedItems           int
    SkippedItems          int
    SuiteReportJSONPath   string
    SuiteReportHTMLPath   string
    CreatedAt             time.Time
    UpdatedAt             time.Time
}
```

### 7.3 SuiteItem 模型（已存在，确认）

```go
type SuiteItem struct {
    ID           string
    SuiteID      string
    ConnectionID string
    TemplateID   string
    ProfileType  string
    Status       SuiteItemStatus
    Phase        string
    ReportID     string       // 执行完成后写入
    StartedAt    *time.Time
    EndedAt      *time.Time
    ErrorMessage string
}
```

---

## 8. 实现范围

### 8.1 AutoBench 数据源接真实能力

| 任务 | 说明 |
|------|------|
| 连接列表 | 从现有 Connections 模块真实数据 |
| 模板列表 | 从现有 Templates 模块真实数据 |
| 删除静态假数据 | 替换本地静态假数据为 API 调用 |

### 8.2 Create Suite 功能

| 任务 | 说明 |
|------|------|
| 启用按钮 | "Create Suite" 按钮可点击 |
| 创建 Suite | 基于所选 connections + templates |
| 持久化元数据 | 写入 suites 表 |
| 生成 manifest | 写入 suite_manifest.json |

### 8.3 Start Suite 功能

| 任务 | 说明 |
|------|------|
| 启用按钮 | "Start Suite" 按钮可点击 |
| 逐个执行 | 每个 SuiteItem 调用现有 Benchmark 执行链路 |
| 生成 report | 每个 item 通过 ReportCollector 生成 report |
| 状态回写 | 更新 item status/started_at/ended_at/report_id |
| Suite 汇总 | 实时/最终更新 Suite 状态 |

### 8.4 与 Reports 联动

| 任务 | 说明 |
|------|------|
| Reports 列表 | AutoBench 执行的报告出现在 Reports 列表 |
| Reports 详情 | 能打开 AutoBench 生成的 report |
| suite_id | 使用真实 Suite.ID |
| 反查关系 | suite_manifest.json 可反查 item 与 report 对应 |

### 8.5 AutoBench UI

| 任务 | 说明 |
|------|------|
| 真实连接选择 | 从 API 获取真实连接列表 |
| 真实模板选择 | 从 API 获取真实模板列表 |
| Suite 预览 | 显示将要执行的 items |
| Create Suite | 可点击创建 |
| Start Suite | 可点击执行 |
| 状态展示 | Suite/Item 级别状态 |
| 跳转 Reports | 执行完成后可查看 Reports |

---

## 9. 验收标准

### 9.1 验收结果 ✅ 全部通过

| # | 验收项 | 状态 | 验证方法 |
|---|--------|------|----------|
| 1 | AutoBench 页面不再是占位草稿 | ✅ 通过 | 按钮可点击 |
| 2 | Create Suite 可点击且可真实创建 | ✅ 通过 | suites 表有记录 |
| 3 | Start Suite 可执行 | ✅ 通过 | 执行 items |
| 4 | Connections 使用真实数据 | ✅ 通过 | connectionStore.fetchConnections |
| 5 | Templates 使用真实数据 | ✅ 通过 | profile_type 自动匹配 |
| 6 | 每个 SuiteItem 调用现有 Benchmark 执行链路 | ✅ 通过 | 复用 BenchmarkUseCase |
| 7 | 每个 SuiteItem 生成 report_id | ✅ 通过 | reports 表有记录 |
| 8 | AutoBench 结果进入 Reports | ✅ 通过 | Reports 列表可见 |
| 9 | Reports 能查看 AutoBench 报告详情 | ✅ 通过 | viewReport 函数 |
| 10 | suite_manifest.json 可恢复完整 suite 视图 | ✅ 通过 | GetSuiteStatus API |
| 11 | 不破坏现有 Benchmark 单次执行 | ✅ 通过 | suite_id="standalone" |
| 12 | 不破坏现有 Reports Phase 1 能力 | ✅ 通过 | 回归测试 |
| 13 | 构建通过 | ✅ 通过 | go build + npm build |
| 14 | 测试通过 | ✅ 通过 | 31 AutoBench 相关测试 |
| 15 | 文档与真实实现一致 | ✅ 通过 | 文档已更新 |

---

## 10. 实施模块

| 模块 | 描述 | 状态 |
|------|------|------|
| M9 | AutoBench 后端 API | ✅ 完成 |
| M10 | Suite 创建功能 | ✅ 完成 |
| M11 | Suite 执行功能 | ✅ 完成 |
| M12 | AutoBench UI 激活 | ✅ 完成 |
| M13 | 文档修正与最终验收 | ✅ 完成 |

---

## 11. 设计决策记录

| ID | 主题 | 决策 |
|----|------|------|
| D016 | AutoBench 当前状态 | ✅ 可用实现已完成 |
| D017 | 本轮目标 | AutoBench 可用化，非 Phase 2 监控扩展 |
| D018 | 数据源 | 真实 Connections API，profile_type 自动匹配 Template |
| D019 | 执行策略 | 复用 BenchmarkUseCase，不重写执行引擎 |
| D020 | SuiteItem 恢复 | 依赖 suite_manifest.json，不建 suite_items 表 |
| D021 | 标准产物 | metrics/monitoring/raw/summary/report.html |
| D022 | suite_id | standalone = "standalone"，AutoBench = UUID |
| D023 | 命名约定 | profile_type = 模板分类属性，用户选择 Profiles 系统自动匹配 Templates |

---

## 12. 已知限制

| 限制 | 说明 |
|------|------|
| 并行执行 | Phase 1 仅支持串行执行 |
| Profile Types | 固定三种：test, cpu_bound, io_bound |
| Template 选择 | 自动匹配，不支持用户手动选择具体模板 |
| Phase 2 监控 | Memory/Network/Load Average 未实现 |

---

**创建日期**: 2026-03-27
**最后更新**: 2026-03-27
**状态**: ✅ 已实现
