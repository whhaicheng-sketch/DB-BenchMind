# AutoBench 设计文档

> **状态**: ✅ 已实现 (2026-03-27)
> **实现版本**: AutoBench Usable Implementation

## 1. 目标与定位
AutoBench 是一个**批量压测编排模块**，用于批量选择多个数据库连接，并基于现有模板与现有执行能力完成自动化测试与报告输出。AutoBench 只做编排与聚合，不改写当前单任务链路。

### 核心定位
- 独立页面，不替代现有 Tasks & Monitor
- 只调用现有 Connections、Templates、任务执行、指标采集能力
- 不影响当前任何已有功能、默认值、交互语义与执行链路
- 产出 Suite 级别结果与报告

### 命名约定

| 用户可见术语 | 内部字段 | Legacy/内部别名 | 说明 |
|-------------|---------|----------------|------|
| Benchmark Types | `profile_type` | profiles, profile_type | 测试类型：test, cpu_bound, io_bound |
| Templates | - | - | 实际压测配置，由系统根据 benchmark type 自动匹配 |
| Connections | `connection_id` | - | 数据库连接 |
| Reports | `report_id` | - | 压测报告 |

**重要说明**：
- 用户在 AutoBench 选择的是 **Benchmark Types**（测试类型）
- 系统根据 (connection.db_type, benchmark_type) 自动匹配对应的 **Template**
- `profile_type` 作为内部字段名保留，是 Template 的分类属性

## 2. 设计原则
1. **兼容性优先**：所有改动必须向后兼容。
2. **新增优先于改造**：优先新增页面、编排层、报告层，不直接修改现有页面职责。
3. **模块化交付**：每次只做一个模块或一个子任务。
4. **续跑友好**：每轮必须更新 progress.md / progress.json。
5. **最小侵入**：仅在确有缺口时新增最小适配层。

## 3. 用户目标
用户只需：
- 配置好 connections
- 选择一个或多个 connection
- 选择测试范围：test / cpu_bound / io_bound
- 点击开始

系统自动完成：
- 预校验
- 任务编排
- 调用现有单任务执行能力
- 汇总结果
- 输出报告

## 4. 页面结构
### 4.1 AutoBenchPage
新页面，分四个区块：
1. 连接选择区
2. 测试范围与执行策略区
3. 执行监控区
4. 报告区

### 4.2 子组件建议
- AutoBenchWizard：多选 connection、选择 profile 范围、执行策略
- AutoBenchMonitor：显示 suite 进度、当前 item、日志入口、结果统计
- AutoBenchReportView：展示汇总报告与导出入口

## 5. 领域模型

### 5.1 Suite
表示一轮 AutoBench 批量执行。

**数据库表 (suites)**:
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
    created_at      TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

**关键字段说明**:
- `id`: UUID，AutoBench Suite 唯一标识
- `suite_manifest_json_path`: 恢复 SuiteItem 状态的快照锚点
- `status`: draft/ready/running/success/partial_success/failed/cancelled

### 5.2 SuiteItem
一个 connection + 一个 profile_type 的执行单元。

**注意**: SuiteItem 不单独建表，状态存储在 `suite_manifest.json`。

**关键字段**:
- `id`: UUID
- `suite_id`: 关联的 Suite ID
- `connection_id`: 数据库连接 ID
- `profile_type`: 压测类型 (test/cpu_bound/io_bound)
- `template_id`: 自动匹配的模板 ID（运行时确定）
- `status`: pending/running/success/failed/skipped
- `report_id`: 执行完成后生成的报告 ID

### 5.3 SuiteManifest (suite_manifest.json)
恢复完整 Suite 状态的唯一快照锚点。

```json
{
  "schema_version": "1.0",
  "suite_id": "uuid",
  "name": "AutoBench Suite",
  "execution_policy": {...},
  "items": [...],
  "created_at": "...",
  "updated_at": "..."
}
```

### 5.4 suite_id 策略

| 场景 | suite_id 值 |
|------|-------------|
| 单次 Benchmark | `"standalone"` (字符串常量) |
| AutoBench Suite | UUID |

## 6. 编排策略
### 默认执行策略
- 默认串行执行，同一时间只跑一个 suite item
- 推荐顺序：test -> cpu_bound -> io_bound
- 默认失败策略：同 connection fail 后，其后续 profile 标记 skipped，并继续下一个 connection
- cleanup 默认开启

### 非目标范围
以下能力不在第一版范围：
- 并行调度
- 自定义复杂报告模板
- 邮件推送
- 历史趋势对比大屏

## 7. 与现有系统的关系
### 必须复用
- 现有 connection 数据
- 现有 template 数据
- 现有单任务执行链路
- 现有日志/指标能力

### 禁止改写
- Performance Analysis 页面职责
- 单任务默认行为
- Templates 的 built-in / copy / readonly 语义
- 现有 Connections / History 页面交互

## 8. 报告设计
### 面向用户的报告
建议 HTML/PDF，包含：
- Executive Summary
- Connection Summary
- Performance Highlights
- Failure Analysis
- Recommendations

### 面向系统的结果
JSON，包含：
- suite metadata
- suite items
- metrics summary
- logs
- failures
- report metadata

## 9. 兼容性红线
- AutoBench 是新增能力，不得改变任何已有页面、已有接口、已有单任务执行行为的默认语义
- 复用现有逻辑时，应通过新增编排层调用，不得直接改写原页面职责
- 对现有能力的任何变动都必须证明是向后兼容且最小侵入

## 10. MVP 建议
第一版仅实现：
- 多选 connection
- 选择 test / cpu_bound / io_bound
- 串行执行
- 调用现有单任务能力
- 生成 HTML + JSON 报告
