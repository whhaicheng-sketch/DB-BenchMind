# AutoBench 架构重构进度

**状态: AutoBench Usable Implementation - 进行中**

当前阶段: AutoBench 可用化实现（非 Phase 2 监控扩展）
当前模块: M10
当前任务: T10.2
下一任务: T10.3

---

## 真实状态说明

### Phase 1 - Reports 持久化层 ✅ 已完成

- M0-M8 全部完成
- Reports 数据模型、Usecase、前端页面已完成
- 单次 Benchmark 自动生成 report (suite_id="standalone")
- suite_manifest.json 结构已定义

### AutoBench UI 当前状态 ⚠️

**AutoBench 页面是占位草稿，非可用产品**：
- "Create Suite (later task)" 按钮是 `disabled` 占位符
- Wizard 使用本地静态假数据，不调用真实 API
- 不能创建真实 Suite
- 不能执行 Suite

### 本轮目标

把 AutoBench 从"占位/草稿 UI"实现为"可创建 Suite、可执行 Suite、可写入 Reports 的真实可用功能"。

---

## 模块状态

| 模块 | 描述 | 状态 |
|------|------|------|
| M0 | 设计规范与进度基础设施 | ✅ 完成 |
| M1 | 导航 UI 标签重命名 | ✅ 完成 |
| M2 | Reports 数据模型与 SQLite 表 | ✅ 完成 |
| M3 | ReportCollector 包装器实现 | ✅ 完成 |
| M4 | Reports Usecase 与 API 绑定 | ✅ 完成 |
| M5 | Reports 前端页面（列表/详情/导出） | ✅ 完成 |
| M6 | AutoBench 集成 report_id | ✅ 完成 |
| M7 | 单次 Benchmark 集成 report | ✅ 完成 |
| M8 | 验收与兼容性回归 | ✅ 完成 |
| M9 | AutoBench 后端 API | ✅ 完成 |
| M10 | Suite 创建功能 | 🔄 进行中 |
| M11 | Suite 执行功能 | ⏳ 待开始 |
| M12 | AutoBench UI 激活 | ⏳ 待开始 |
| M13 | 文档修正与最终验收 | ⏳ 待开始 |

---

## 任务日志

### M9 - AutoBench 后端 API ✅

- [x] T9.1 定义 AutoBenchBinding API
  - CreateSuite: 创建 Suite（基于 connection_ids + template_ids）
  - StartSuite: 启动 Suite 执行
  - GetSuiteStatus: 获取 Suite 状态
  - GetSuiteItems: 获取 SuiteItem 列表
  - CancelSuite: 取消 Suite 执行
- [x] T9.2 实现 SuiteRepository
  - CreateSuite: 插入 suites 表
  - GetSuite: 查询单条
  - ListSuites: 分页查询
  - UpdateSuiteStatus: 更新状态
- [x] T9.3 后端单元测试

### M10 - Suite 创建功能 🔄

- [x] T10.1 Suite 创建逻辑
  - 基于 selected_connection_ids + selected_template_ids
  - 生成 SuiteItem 列表
  - 写入 suites 表
  - 生成初始 suite_manifest.json (on start)
- [ ] T10.2 前端 Create Suite 按钮
  - 移除 disabled 属性
  - 调用 CreateSuite API
  - 显示创建结果
- [ ] T10.3 后端单元测试

### M11 - Suite 执行功能 ⏳

- [x] T11.1 Suite 执行编排
  - 实现 AutoBenchSuiteRunner (已存在)
  - 逐个执行 SuiteItem
  - 调用现有 BenchmarkUseCase
  - 通过 ReportCollector 生成 report
  - 更新 item 状态
- [x] T11.2 状态回写
  - 更新 suite_manifest.json
  - 更新 suites 表状态
  - 实时进度更新
- [ ] T11.3 前端 Start Suite 按钮
  - 移除 disabled 属性
  - 调用 StartSuite API
  - 显示执行进度
- [ ] T11.4 后端单元测试

### M12 - AutoBench UI 激活 🔄

- [x] T12.1 真实数据源接入
  - Connections 列表从真实 API 获取
  - Templates 列表从真实 API 获取
  - 删除静态假数据
- [ ] T12.2 Suite 状态展示
  - Suite 级别状态
  - Item 级别状态
  - 执行进度条
- [ ] T12.3 Reports 联动
  - AutoBench 结果出现在 Reports 列表
  - 点击可查看报告详情
- [ ] T12.4 前端单元测试

### M13 - 文档修正与最终验收 ⏳

- [ ] T13.1 文档修正
  - 删除旧草稿术语
  - 准确描述当前状态
  - 更新架构图
- [ ] T13.2 兼容性回归测试
- [ ] T13.3 最终验收

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
| D009 | JSON 导出策略 | 运行时动态打包 |
| D010 | 原始输出存储 | stdout/stderr 内联 raw.json |
| D011 | JSON Schema 版本 | 所有 JSON 必须包含 schema_version |
| D012 | 报告来源类型 | reports.source_type 显式字段 |
| D013 | 监控取数路径 | 直接访问 SystemCollector，不依赖前端事件 |
| D014 | 持久化失败处理 | 记录日志 + PersistError 字段 |
| D015 | 写入顺序 | 文件先写，SQLite 后写 |
| D016 | AutoBench 当前状态 | 占位草稿，非可用产品 |
| D017 | 本轮目标 | AutoBench 可用化，非 Phase 2 监控扩展 |
| D018 | 数据源 | 真实 Connections/Templates API |
| D019 | 执行策略 | 复用 BenchmarkUseCase，不重写执行引擎 |
| D020 | SuiteItem 恢复 | 依赖 suite_manifest.json，不建 suite_items 表 |
| D021 | 标准产物 | metrics/monitoring/raw/summary/report.html |
| D022 | suite_id | standalone = "standalone"，AutoBench = UUID |

---

## 设计文档

- **Phase 1 设计**: docs/superpowers/specs/2026-03-25-autobench-reports-design.md
- **Phase 1 实现**: docs/superpowers/plans/2026-03-26-autobench-reports-implementation.md
- **本轮设计**: docs/superpowers/specs/2026-03-27-autobench-usable-implementation-design.md

---

## 本轮不做

以下能力不在本轮范围：

- Memory 使用率
- Network IO (rx/tx)
- Load Average
- DB 连接数
- DB 活跃会话
- stdout.log / stderr.log 单独文件
- suite_items 表

---

**创建日期**: 2026-03-25
**最后更新**: 2026-03-27
