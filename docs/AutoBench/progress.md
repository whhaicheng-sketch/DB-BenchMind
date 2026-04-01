# AutoBench 架构重构进度

**状态: AutoBench Usable Implementation - ✅ 已完成**

---

## 完成状态说明

### Phase 1 - Reports 持久化层 ✅

- M0-M8 全部完成
- Reports 数据模型、Usecase、前端页面已完成
- 单次 Benchmark 自动生成 report (suite_id="standalone")
- suite_manifest.json 结构已定义

### AutoBench 可用化 ✅

**AutoBench 页面已实现为可用产品**：
- Create Suite 按钮调用真实 API
- Start Suite 按钮调用真实 API
- 使用真实 Connections 数据
- Suite 状态面板展示进度
- Item 级别状态展示
- Reports 联动导航

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
| M10 | Suite 创建功能 | ✅ 完成 |
| M11 | Suite 执行功能 | ✅ 完成 |
| M12 | AutoBench UI 激活 | ✅ 完成 |
| M13 | 文档修正与最终验收 | ✅ 完成 |
| M14 | AutoBench UI Overhaul + Report Delete | ✅ 完成 |
| M15 | Reports Page AutoBench Grouping Fix | ✅ 完成 |

---

## 测试状态

### 后端 AutoBench 测试 (12 个)
- `TestAutoBenchBinding_CreateSuite` - 3 子测试 ✅
- `TestAutoBenchBinding_GetSuiteStatus` ✅
- `TestAutoBenchBinding_GetSuiteStatus_NotFound` ✅
- `TestAutoBenchBinding_GetExecutionPlan` ✅
- `TestAutoBenchBinding_ListProfiles` ✅
- `TestAutoBenchBinding_StartSuite_NoRunner` ✅
- `TestSuiteRepository_Save` ✅
- `TestSuiteRepository_FindByID_NotFound` ✅
- `TestSuiteRepository_FindAll` ✅
- `TestSuiteRepository_UpdateStatus` ✅
- `TestSuiteRepository_Delete` ✅
- `TestSuiteRepository_Delete_NotFound` ✅

### 前端 AutoBench 测试 (26 个 + M15 新增)

M15 新增测试:
- source_type=autobench 分组测试 ✅
- autobench 报告永不显示为 standalone ✅
- benchmark source_type standalone 分组 ✅
- 3 completed + 1 running 状态聚合 = running ✅
- orphan suite group 使用 deriveGroupStatus ✅
- computeSuiteProgress 计数测试 ✅
- canViewReport 完成项可查看、运行/失败/待定不可查看 ✅
- pendingReportId 跨标签页导航 ✅
- T12.2: Suite 状态面板展示 ✅
- T12.2: Item 级别状态展示 ✅
- T12.2: 进度条展示 ✅
- T12.2: 状态徽章样式 ✅
- T12.3: viewReport 函数 ✅
- T12.3: goToReports 函数 ✅
- T12.3: View Report 按钮 ✅
- T12.3: View All Reports 按钮 ✅
- T12.1: 真实数据源 ✅
- T12.1: 加载连接 ✅
- M10: CreateSuite API ✅
- M11: StartSuite API ✅
- M11: GetSuiteStatus API ✅
- T12.2: suiteSummary computed ✅
- T12.2: 按钮状态 ✅
- T12.2: 错误处理 ✅
- T12.2: 清理 polling ✅
- T12.2: 连接类型过滤 ✅
- T12.2: wizard draft ✅
- UI Overhaul: Two-column wizard layout ✅
- UI Overhaul: Compact connection rows ✅
- UI Overhaul: Profile toggle pills ✅
- UI Overhaul: Active run section with left border ✅
- UI Overhaul: Elapsed time tracking ✅
- UI Overhaul: currentItem computed ✅
- UI Overhaul: connNameMap connection resolution ✅

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
| D016 | AutoBench 当前状态 | 可用实现已完成 |
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

以下能力不在本轮范围（可在后续迭代中实现）：

- Memory 使用率
- Network IO (rx/tx)
- Load Average
- DB 连接数
- DB 活跃会话
- stdout.log / stderr.log 单独文件
- suite_items 表

---

### Phase E2E - 端到端真实验证 ✅ (2026-03-27)

**四库全部通过，真实压测执行成功**

| 数据库 | 工具 | TPS | 平均延迟 | 状态 |
|--------|------|-----|----------|------|
| MySQL 5.7.44 | sysbench oltp_read_write | 73.86 | 13.53ms | ✅ 通过 |
| Oracle 11g | swingbench charbench | 490.98 | 1.95ms | ✅ 通过 |
| PostgreSQL 13.14 | sysbench oltp_read_write | 70.18 | - | ✅ 通过 |
| SQL Server 2019 | hammerdb TPROC-C | 217.30 | - | ✅ 通过 |

**发现并修复的问题**:
- P1: ReportCollector 只写文件不写 SQLite reports 表 → 新增 persistToDB() 方法
- P2: main.go 未调用 SetReportCollector() → 已修复
- P3: main.go 未传入 SuiteRepository → 已修复 WithSuiteRepository(suiteRepo)

**验证结果**:
- 4/4 连接全部可达（含 SSH 隧道）
- 4/4 压测全部成功（test profile）
- 4/4 Reports 写入 SQLite + 文件落地（metrics.json, monitoring.json, raw.json）
- 4/4 初始连接完整保留
- 连接恢复脚本验证通过：`scripts/restore_initial_connections.sh`

### Phase E2E-2 - 参数归一化修复与二次验证 ✅ (2026-03-29)

**修复 AutoBench 空参数传递问题，四库重新验证通过**

**发现并修复的问题**:
- P4 (Critical): `autobench_runner.go` 创建 `BenchmarkTask` 时使用空 `Parameters: map[string]interface{}{}`，
  导致 sysbench adapter 的 `ValidateConfig` 因缺少 `threads`/`time` 参数而失败。
  → 重构 `selectTemplateIDForItem` 为 `selectTemplateForItem`（返回完整模板对象），
  新增 `resolveDefaultRunParams` 从模板 Runtime/ToolConfig 提取默认参数。

**二次验证结果 (2026-03-29)**:

| 数据库 | 工具 | TPS | 平均延迟 | 状态 |
|--------|------|-----|----------|------|
| MySQL 5.7.44 | sysbench oltp_read_write | 545.56 | 14.65ms | ✅ 通过 |
| Oracle 11g | swingbench charbench | 1371.63 | 22.82ms | ✅ 通过 |
| PostgreSQL 13.14 | sysbench oltp_read_write | 648.76 | 12.33ms | ✅ 通过 |
| SQL Server 2019 | hammerdb TPROC-C | 673.07 | - | ✅ 通过 |

**数据产物验证**:
- 4/4 Reports 写入 SQLite (status=completed, TPS/latency 正确)
- 4/4 metrics.json 文件存在 (含 percentiles/summary)
- 4/4 monitoring.json 文件存在 (结构正确, time_series=0 为 Phase 2 预期)
- 4/4 raw.json 文件存在 (含 parsed_result: transactions/queries/errors)
- 4/4 初始连接完整保留 (IDs 一致)
- 连接恢复脚本验证通过

---

### Phase M15 - Reports Page AutoBench Grouping Fix ✅ (2026-03-31)

**修复 Reports 页面 AutoBench 套件报告分组显示问题**

**发现并修复的问题**:
- P5: `reportGroups` computed 仅按 `suite_id` 分组，未使用 `source_type` 区分 AutoBench 与 standalone 报告
  → 重写分组逻辑：按 `source_type=autobench` 或非 `standalone` 的 `suite_id` 进行分区
  → AutoBench 报告永不显示为 standalone 卡片（去重保障）

- P6: 套件组状态使用存储的 `suites.status` 列，不反映子报告实时状态
  → 使用 `deriveGroupStatus(group.reports)` 计算实时聚合状态
  → 新增 `computeSuiteProgress()` 提供 completed/running/failed 计数器
  → 套件行显示 "4 items · 3 completed · 1 running"

- P7: `viewReport()` 忽略 `reportId` 参数，仅切换标签页
  → 新增 `pendingReportId` 状态和跨标签页导航机制
  → `AutoBenchTab.vue` 调用 `setPendingReportId`，`ReportsTab.vue` 在 `onMounted` 消费

- P8: 运行中/失败/待定项显示 "View Report" 按钮但打开空白页
  → 新增 `canViewReport()` 条件渲染：只有 completed/success 才显示查看按钮
  → running/pending 显示 "Running" 文本，failed 显示 "Failed" 文本

**测试验证**:
- 353 前端测试全部通过（含 8 个 M15 新增测试）

### Phase M16 - 缺陷收敛：监控、观察态、状态链路修复 ✅ (2026-03-31)

**一次性修复 6 类联动问题，完成完整缺陷收敛**

| 问题 | 根因 | 修复 |
|------|------|------|
| P9: Monitor "All" 按钮 | TIME_WINDOWS 含 value:0 的 "All" 条目 | 移除 "All" 条目，所有图表统一使用滑动窗口 |
| P10: Pause 只冻结 TPS/TPM | monitorPaused 仅影响 displayedBusinessMetrics | 新增 displayedCpuChart/displayedDiskChart 全局冻结 |
| P11: Reports deriveGroupStatus 运行时错误 | analyzeGroupReports 重命名时孤立分支未更新 | 修复孤立分支调用 |
| P12: AutoBench flags | connCapabilitiesMap 逻辑已正确 | 确认无需修复 |
| P13: AutoBench 观察态不联动 | statusSummary 未感知 AutoBench 状态 | 添加 AutoBench 状态摘要和观察态增强 |
| P14: Suite success 但 Reports 仍显示 running | collectStandaloneReport 硬编码 StandaloneSuiteID，runner 不传 suite context | 传递 SuiteID/SuiteItemID 到 BenchmarkTask，collectStandaloneReport 使用 suite context 更新现有 running 报告 |

**测试验证**:
- 396 前端测试全部通过（含 M16 新增 43 个测试）
- 后端 usecase 测试全部通过

---
### Phase M18 - 二次缺陷收敛：Re-run 修复、标签刷新、Phase Timing 同步 ✅ (2026-04-01)

**修复 3 个联动问题，分析 1 个后端架构限制**

本轮处理范围: P10 (Re-run TypeError), P11 (标签页切换后数据不刷新), P12 (Phase Timing 不同步), P7 root cause analysis (AutoBench 系统指标缺失)

| 问题 | 严重度 | 状态 | 根因 | 修复 |
|------|--------|------|------|------|
| P10: Re-run 点击后 TypeError | high | ✅ 已修复 | `resetSuite()` 先将 `suiteStatus` 置 null，随后读取 `suiteStatus.value.name` 抛出 TypeError | 在 `resetSuite()` 调用前保存 `suiteName = suiteStatus.value.name \|\| 'Re-run Suite'` |
| P11: 标签页切换后数据不刷新 | high | ✅ 已修复 | `<KeepAlive>` 阻止组件重新挂载，`onMounted` 不再触发，数据过时 | 为 5 个 tab 组件 (AutoBench/Reports/TasksMonitor/Connections/Templates) 添加 `onActivated` 钩子，刷新对应数据 |
| P12: AutoBench 观察 Phase Timing 不同步 | medium | ✅ 已修复 | `monitorTask` virtual task 缺少 `timing` 字段；后端提供 `phase_timings` 数组但前端未映射为 `{prepare_ms, run_ms, total_ms}` | 在 `monitorTask` computed 中新增 `phase_timings → timing` 映射逻辑 |
| P7: AutoBench 系统指标缺失 | high | ⏸ 已分析，延期 | `autobench_runner.go:collectItemMetrics()` 硬编码 `system_enabled: false`；runner 缺少 SSH 指标收集器基础设施（无 SSH 连接凭据传递路径） | 延期至后续迭代（需后端架构扩展：runner 需 SSH 连接信息 + SystemMetricsCollector 实例） |

**P7 延期说明**:

根因已确认：`autobench_runner.go` 的 `collectItemMetrics()` 方法（约第 452-467 行）硬编码了 `system_enabled: false` 和 `system_message: "System metrics not available in AutoBench mode"`。AutoBench runner 通过 `BenchmarkUseCase.StartBenchmark()` 直接执行压测，完全绕过 `TaskBinding`，因此无法访问 `SSHMetricsCollector`。

修复需要后端架构变更：
1. Runner 需要获取目标连接的 SSH 凭据
2. Runner 需要创建并管理 `SSHMetricsCollector` 实例
3. `collectItemMetrics()` 需要集成系统指标采集

这是 Phase 2 监控扩展的工作范畴，不在当前缺陷收敛范围内。

**测试验证**:
- 9 个新增测试（P10: 4 个, P12: 5 个）
- 49 个收敛测试全部通过
- 220 个前端测试全部通过
- 9 个后端 usecase 测试全部通过

**修改文件**:
- `frontend/src/components/tabs/AutoBenchTab.vue` — P10 (suiteName 保存), P11 (onActivated)
- `frontend/src/components/tabs/TasksMonitorTab.vue` — P12 (phase_timings 映射), P11 (onActivated)
- `frontend/src/components/tabs/ReportsTab.vue` — P11 (onActivated)
- `frontend/src/components/tabs/ConnectionsTab.vue` — P11 (onActivated)
- `frontend/src/components/tabs/TemplatesTab.vue` — P11 (onActivated)
- `frontend/tests/defect-convergence-m17.test.mjs` — 新增 9 个测试

---
**最后更新**: 2026-04-01
**状态**: ✅ 完成（含端到端真实验证 + 参数归一化修复 + UI Overhaul + Report Delete + Reports Grouping Fix + 缺陷收敛 M16 + M17 + M18 + M19）

### Phase M19 - 缺陷收敛 P13-P16: 视觉层次、系统指标、动态计时、删除 View Report ✅ (2026-04-01)

**修复 4 个问题，新增完整后端 SSH 指标采集链路**

| 问题 | 严重度 | 状态 | 根因 | 修复 |
|------|--------|------|------|------|
| P13: 界面层次感不够 | low | ✅ 已修复 | 所有区块使用相同背景/边框/间距，缺少视觉分层 | AutoBenchTab 增强卡片阴影/边框/背景/标题权重/表头样式；TasksMonitorTab 状态条添加阴影 |
| P14: AutoBench 托管期间无系统资源监控 | high | ✅ 已修复 | `collectItemMetrics()` 硬编码 `system_enabled: false`；runner 无 SSH 指标收集器 | 后端：runner 新增 `sshConfigFromConnection()` + SSH 收集器生命周期管理 + CPU/Disk 序列填充；前端：复用现有 system metric 图表逻辑 |
| P15: PREPARE/RUN/TOTAL 不动态变化 | high | ✅ 已修复 | `phase_timings` 仅在阶段完成时记录 duration_ms，进行中阶段 duration 为 0 | 前端 `monitorTask` 新增实时计时计算：`nowTick - started_at - completedPhases = currentPhaseMs` |
| P16: View Report 按钮不稳定 | medium | ✅ 已修复 | View Report 跳转逻辑不稳定 | 删除 Report 列、viewReport 函数、link-button 样式；保留 View All Reports 按钮 |

**P14 后端变更详情**:
- `autobench_runner.go`: 新增 `sshConfigFromConnection()` 从连接提取 SSH 配置；每个 item 开始时创建 `SSHMetricsCollector`；`collectItemMetrics()` 集成 SSH snapshot，填充 cpu_user/cpu_sys/cpu_iowait/cpu_steal/disk_read_bps/disk_write_bps/disk_read_latency_ms/disk_write_latency_ms 序列；item 完成后停止收集器
- 新增 `collector` 包依赖

**P15 前端变更详情**:
- `TasksMonitorTab.vue` `monitorTask` computed: 从仅累积已完成 `phase_timings` 扩展为同时计算进行中阶段的实时 duration
- 使用 `nowTick`（每秒刷新）驱动实时更新

**测试验证**:
- 458 前端测试全部通过（含 M19 新增 14 个测试：P14: 7 个, P15: 7 个）
- 24 个后端 usecase 测试全部通过（含 2 个新测试：`TestSSHConfigFromConnection` 6 子测试 + `TestCollectItemMetricsWithSystemMetrics`）
- 预有 `TestAITestConnection` 和 `TestIntegration_ConnectionValidation` 失败（非本次变更引入）

**修改文件**:
- `frontend/src/components/tabs/AutoBenchTab.vue` — P13 (视觉层次), P16 (删除 Report 列)
- `frontend/src/components/tabs/TasksMonitorTab.vue` — P13 (状态条阴影), P15 (实时计时)
- `internal/app/usecase/autobench_runner.go` — P14 (SSH 收集器 + 系统指标)
- `internal/app/usecase/autobench_runner_test.go` — P14 新增 2 个测试
- `frontend/tests/autoBenchTab.test.mjs` — P16 更新 3 个测试
- `frontend/tests/defect-convergence-m17.test.mjs` — P14/P15 新增 14 个测试
