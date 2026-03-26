# AutoBench 架构重构进度 — Reports 持久化层

**状态: Phase 1 - 实现中**

当前阶段: Phase 1 - 核心功能实现
当前模块: M5
当前任务: T5.1 (已完成)
下一任务: T5.2

---

## 模块状态

| 模块 | 描述 | 状态 |
|------|------|------|
| M0 | 设计规范与进度基础设施 | ✅ 完成 |
| M1 | 导航 UI 标签重命名 | ✅ 完成 |
| M2 | Reports 数据模型与 SQLite 表 | ✅ 完成 |
| M3 | ReportCollector 包装器实现 | ✅ 完成 |
| M4 | Reports Usecase 与 API 绑定 | ✅ 完成 |
| M5 | Reports 前端页面（列表/详情/导出） | 🔄 进行中 |
| M6 | AutoBench 集成 report_id | ✅ 完成 |
| M7 | 单次 Benchmark 集成 report | ✅ 完成 |
| M8 | 验收与兼容性回归 | ⏳ 待开始 |

---

## 任务日志

### M0 - 设计规范与进度基础设施 ✅

- [x] T0.1 需求澄清与方案确认
- [x] T0.2 架构概览文档定稿
- [x] T0.3 数据模型设计定稿
- [x] T0.4 执行流程设计定稿
- [x] T0.5 持久化规范定稿
- [x] T0.6 设计文档写入 docs/superpowers/specs/
- [x] T0.7 实现计划写入 docs/superpowers/plans/

### M1 - 导航 UI 标签重命名 ✅

- [x] T1.1 更新 navigationTabs.mjs 标签文本
- [x] T1.2 前端测试验证

### M2 - Reports 数据模型与 SQLite 表 ✅

- [x] T2.1 定义 Report 领域模型
- [x] T2.2 SQLite reports/suites 表结构
- [x] T2.3 集成报告 schema 到 SQLite 初始化
- [x] T2.4 扩展报告 schema 测试覆盖

### M3 - ReportCollector 包装器实现 ✅

- [x] T3.1 ReportCollector 接口定义
- [x] T3.2 实现 CollectAndPersist 核心逻辑
- [x] T3.3 文件持久化实现（metrics/monitoring/raw/summary）

### M4 - Reports Usecase 与 API 绑定 ✅

- [x] T4.1 ReportUsecase 实现
  - ListReports: 分页查询、过滤（suite_id/status/connection_id）
  - GetReport: 单条查询
  - GetReportMetrics: 从 JSON 文件加载完整指标
  - ListSuites: 分页查询、过滤（status）
  - GetSuite: 单条查询
  - 使用 SQLite 兼容的 `?` 占位符
  - 完整的表驱动测试覆盖
- [x] T4.2 Wails ReportBinding 暴露
  - ListReports: 暴露给前端，带 DTO 转换
  - GetReport: 返回单个报告
  - GetReportMetrics: 返回 MetricsData
  - ListSuites: 列出套件
  - GetSuite: 返回单个套件
- [x] T4.3 后端单元测试

### M5 - Reports 前端页面 🔄

- [x] T5.1 Reports 列表页（HistoryTab 改造为 ReportsTab）
- [ ] T5.2 Reports 详情页
- [ ] T5.3 图表组件集成
- [ ] T5.4 JSON/HTML 导出功能
- [ ] T5.5 前端单元测试

### M6 - AutoBench 集成 report_id ✅

- [x] T6.1 SuiteItem 关联 report_id
- [ ] T6.2 suite_manifest.json 生成与更新
- [ ] T6.3 后端单元测试

### M7 - 单次 Benchmark 集成 report ✅

- [x] T7.1 单次执行生成 standalone report
  - 添加 reportCollector 字段到 BenchmarkUseCase
  - 实现 WithReportCollector 选项模式
  - 实现 collectStandaloneReport 方法
  - 使用 suite_id='standalone' 标识独立报告
  - 在 goroutine 中执行，不阻塞响应
- [ ] T7.2 suite_id = "standalone" 策略实现
- [ ] T7.3 后端单元测试

### M8 - 验收与兼容性回归 ⏳

- [ ] T8.1 后端回归测试
- [ ] T8.2 前端回归测试
- [ ] T8.3 集成测试
- [ ] T8.4 文档同步更新

---

## 关键决策记录

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

## 设计文档

- **主文档**: docs/superpowers/specs/2026-03-25-autobench-reports-design.md
- **实现计划**: docs/superpowers/plans/2026-03-26-autobench-reports-implementation.md

---

## Phase 2 预留扩展

以下能力不在 Phase 1 范围，但在数据模型中预留：

- Memory 使用率
- Network IO (rx/tx)
- Load Average
- DB 连接数
- DB 活跃会话

---

## 最近提交

1. `87fc90d` feat(benchmark): integrate standalone report collection
2. `bc33ce5` feat(bindings): add ReportBinding for Wails frontend
3. `42a0d18` feat(usecase): add ReportUsecase for querying reports and suites
4. `537dfbc` feat(autobench): add ReportID field to SuiteItem
5. `4e62fd8` refactor(ui): rename HistoryTab to ReportsTab

---

**创建日期**: 2026-03-25
**最后更新**: 2026-03-26
