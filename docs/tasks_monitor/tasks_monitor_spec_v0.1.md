# Tasks & Monitor 需求规格说明书 v0.1

## 1. 目标

本规格说明书定义 **Tasks & Monitor 第一阶段** 的功能需求、非功能需求、业务规则与验收标准。

目标是交付一个可用的压测执行工作台，完成以下闭环：

- 选择 Template 与 Connection
- 创建任务并预览确认
- 入队或启动任务
- 执行 prepare / run / cleanup / full pipeline
- 在运行中查看 TPM / TPS 和可选系统监控
- 查看实时日志

---

## 2. 术语定义

### 2.1 Template
压测模板，定义 benchmark tool、参数模板、可执行能力、说明等。

### 2.2 Connection
数据库连接，包含数据库类型、连接信息、是否启用 SSH、SSH 连接信息等。

### 2.3 Task
一次具体压测执行对象，由 Template + Connection + Action + Runtime Overrides 解析生成。

### 2.4 Action
任务动作，当前阶段支持：
- prepare
- run
- cleanup
- full pipeline

### 2.5 Queue
任务队列。多个任务可排队，但同一时刻仅允许一个任务处于执行态。

---

## 3. 功能范围

### 3.1 必须实现

1. 在 Tasks & Monitor 页面创建任务。
2. 支持任务预览确认。
3. 支持多任务排队。
4. 同一时刻只允许运行一个压测任务。
5. 支持 prepare / run / cleanup / full pipeline。
6. 支持执行前 DB Connection Test。
7. 支持 SSH 监控可选启用。
8. 支持显示 TPS / TPM。
9. 支持显示 CPU / Disk IO（仅 SSH 可用时）。
10. 不再显示 Capacity；Tasks & Monitor 仅保留 TPS / TPM / CPU / Disk IO 四类监控。
11. 支持实时日志查看。

### 3.2 明确不在范围内

1. 多任务并发执行。
2. 报告导出。
3. 复杂任务编辑。
4. 队列优先级与拖拽排序。
5. 长周期资源历史分析。

---

## 4. 功能需求

## 4.1 任务创建

### FR-1 任务创建入口
系统必须允许用户在 Tasks & Monitor 页面中选择：
- Template
- Connection
- Action
- 少量 Runtime Overrides

### FR-2 Task Name
系统必须为新任务生成默认 Task Name，供执行、日志与状态追踪使用。

默认建议格式：
`{templateName}-{yyyyMMdd-HHmmss}`

补充约束：
- Tasks 页面 Create Task 区不再单独展示或要求用户配置 Task Name。
- Preview 中如展示 Task Name，只能作为内部解析结果，不得作为主要配置项。

### FR-2A Tasks 草稿保留
系统必须保留 Tasks 页面当前草稿状态，至少满足 tab 切换后返回不丢失。

至少包括：
- databaseType
- connectionId
- templateId
- action
- overrides
- 与当前页面配置直接相关的筛选值

恢复规则：
- 若 connection / template 仍存在且仍匹配，则自动恢复。
- 若已失效，只清除失效字段，不得整份草稿全部清空。

### FR-3 运行参数覆写
当前阶段仅允许覆写少量运行参数，至少包括：
- threads（sysbench）
- virtual users（Swingbench / HammerDB）
- duration

### FR-4 Preview
用户点击 Start 之前，系统必须提供 Preview 弹框，展示：
- Template 摘要
- Connection 摘要
- Action
- 解析后的运行参数
- 监控可用性摘要

### FR-5 Confirm 后启动或入队
用户确认 Preview 后：
- 若当前无运行任务，系统必须立即启动该任务。
- 若当前已有运行任务，系统必须将该任务置为 queued。

### FR-5A 顶部主操作按钮
Tasks 页面顶部的 Start / Stop 必须作为主操作按钮呈现。

要求：
- Start 使用显著绿色主按钮
- Stop 使用显著红色危险按钮
- Open Log Viewer 必须与 Start / Stop 同处顶部主操作区
- Start 点击后仍先进入 Preview，再由 Confirm 真正启动
- 有 active task 时 Start 置灰
- 无 active task 时 Stop 置灰

---

## 4.2 队列管理

### FR-6 多任务队列
系统必须支持多个任务存在于等待队列中。

### FR-7 单运行约束
系统必须保证任意时刻最多仅有一个任务处于活跃执行态。

### FR-8 自动推进
当前运行任务结束后，系统必须自动拉起队列中的下一个任务。

### FR-9 取消排队
系统必须支持取消 queued 状态的任务。

### FR-10 队列任务修改策略
当前阶段系统不需要支持直接编辑 queued 任务；若需修改，应通过取消后重新创建完成。

---

## 4.3 执行动作

### FR-11 支持动作
系统必须支持以下动作：
- prepare
- run
- cleanup
- full pipeline

说明：对于 Swingbench 这类底层工具带有 `build / generate / delete` 原生命令语义的场景，Tasks & Monitor 仍统一按 `prepare / run / cleanup` 暴露动作能力；模板能力与任务执行链必须映射到这三个动作语义，不能在任务页退化成仅 `run`。

### FR-12 Full Pipeline 顺序
UI 文案可使用 `Run Full Flow (Prepare -> Run -> Cleanup)`，其语义必须保持为 full pipeline。

full pipeline 必须按以下顺序执行：
1. prepare
2. run
3. cleanup

### FR-13 Run 中止后的 cleanup 规则
当 full pipeline 中的 run 阶段 failed 或 stopped 时，系统默认不自动 cleanup，应由用户自行决定是否执行 cleanup。

---

## 4.4 执行前校验

### FR-14 DB 校验
prepare / run / cleanup / full pipeline 在进入执行态前都必须通过 DB Connection Test。

### FR-15 DB 校验失败处理
若 DB 校验失败，系统必须阻断任务执行，并将任务标记为 failed 或 validation failed（实现可选，但需有明确失败反馈）。

### FR-16 SSH 校验
若连接启用了 SSH，系统可执行 SSH 可用性校验。

### FR-17 SSH 失败处理
SSH 校验失败不得阻断任务运行，但系统必须禁用 CPU / Disk 监控，并在 UI 上提示原因。

---

## 4.5 停止任务

### FR-18 Stop
系统必须支持停止当前运行中的任务。

### FR-19 Stop 语义
Stop 仅停止当前执行进程，不自动 cleanup，不自动清理数据。

### FR-20 Stop 后状态
被停止的任务必须进入 `stopped` 状态。

---

## 4.6 任务状态与阶段

### FR-21 状态集
系统至少应支持以下状态：
- queued
- starting
- preparing
- running
- cleaning
- stopping
- stopped
- success
- failed

### FR-22 当前阶段
系统必须维护当前阶段字段，至少支持：
- none
- prepare
- run
- cleanup

### FR-23 阶段历史
系统应记录任务阶段历史，用于展示和后续历史复用。

### FR-23A 顶部极简状态条
系统必须在顶部展示一行极简状态条，至少包含：
- 状态
- 当前阶段
- 时间

规则：
- `Run Full Flow` 下必须展示真实执行阶段：prepare / run / cleanup
- 结束后阶段可显示 `finished`
- 不再保留右上大面积 Current Task 详情卡片

---

## 4.7 监控指标

### FR-24 业务指标
系统必须统一输出以下业务指标：
- TPS
- TPM

### FR-25 工具适配
对于不同 benchmark tool，系统必须在后端进行统一转换：
- HammerDB / Swingbench：优先使用原生可得指标
- Sysbench：若仅有 TPS，则 TPM 应支持按 `TPS × 60` 计算

补充约束：
- Oracle Swingbench 的 `prepare` 必须对应 schema / data prepare，不得把 wizard 阶段当作 workload run。
- Oracle Swingbench 的 `run` 必须调用官方 workload frontend；当前自动化链路默认使用 `charbench`。
- Oracle Swingbench 的 TPS / TPM 优先来自 `charbench -v ...` 的 stdout 实时输出；run 结束后的汇总结果应继续读取 `results.xml` 或 `-r` 指定结果文件。
- Oracle Swingbench 在 `prepare` 阶段没有真实 workload 吞吐时，TPS / TPM 显示为 `0` 属于合理行为；`run` 阶段持续为 `0` 则视为执行链或解析链故障。

### FR-26 业务指标统计
每个业务指标至少应展示：
- current
- avg
- max

### FR-27 系统指标
在 SSH 可用时，系统必须支持展示：
- CPU User
- CPU Sys
- CPU IOWait
- CPU Steal
- Disk Read
- Disk Write
- Disk Read Latency
- Disk Write Latency

补充约束：
- Disk Latency 必须继续基于 SSH + Linux `/proc/diskstats` 采集
- read latency 采用 `deltaReadTimeMs / deltaReadsCompleted`
- write latency 采用 `deltaWriteTimeMs / deltaWritesCompleted`
- 若对应采样窗口内 completed 次数为 0，则 latency 显示为 0 或空值
- 多磁盘聚合必须按 I/O 次数加权，不得做简单平均

### FR-28 SSH 不可用时的展示
当 SSH 未配置或 SSH 不可用时，系统必须明确提示系统监控不可用，并以降级方式展示 UI，而不是静默消失。

补充约束：
- Disk latency 必须与现有 CPU / Disk 指标使用同一 SSH 降级逻辑
- 不得为 latency 单独弹错误或引入额外异常分支

---

## 4.8 监控可视化

### FR-29 指标概览卡
系统应为核心指标提供总览板风格概览卡，当前值必须成为视觉重心。

### FR-30 Throughput 历史表现
系统必须在 `Current Throughput` 中为以下指标提供最近窗口的历史表现图形：
- TPS
- TPM

补充约束：
- 历史图必须保留明显可辨识的波动轮廓线
- 可保留轻量 fill，但轮廓线必须优先保证可读性
- 图形表现应接近 Swingbench Overview 风格的明显波动线，而不是只有淡填充

### FR-31 趋势窗口
历史表现图默认采用最近 5 分钟滚动窗口。

### FR-31A 监控优先布局
页面必须将主要空间优先分配给 4 类核心图表 / 面板：
- TPS
- TPM
- CPU
- DISK IO

布局要求：
- 第一行：TPS、TPM
- 第二行：CPU、DISK IO
- 不得再保留独立 Throughput Trend 区块

补充约束：
- 当前阶段，Tasks & Monitor 页面默认窗口调整为 `1200 × 800`
- 在当前默认窗口下，首屏应优先保证顶部主操作、状态条、Create Task、4 类监控面板可读可用
- `1280 × 800` 固定视口紧凑布局目标未取消，后续将单独完成界面固定化收口
- 页面主容器不得继续以自然内容撑高为主要布局方式
- Create Task 与 4 类监控面板必须进一步紧凑化
- 页面整体字体需统一缩小，保持风格一致
- 左侧 Create Task 区固定宽度为 `250px`

### FR-31B CPU 配色规则
CPU 图表必须满足：
- user = 绿色
- sys = 红色
- iowait = 蓝色
- st = 黄色
- 图例与曲线颜色一致
- 深色主题下有明确对比度
- steal time 不可用时必须安全降级为 0 或 unavailable，不得导致前端报错、折线错乱或布局异常

### FR-31C Disk IO 展示策略
DISK IO 必须作为单个面板展示，并在同一张图中展示 4 条线：
- read
- write
- read latency
- write latency

要求：
- 必须采用双轴：
  - 带宽轴用于 `read / write`
  - 时延轴用于 `read latency / write latency`
- 不得把 latency 与带宽绘制在同一标尺上
- 摘要文案至少包含 `read` / `write` / `r_lat` / `w_lat`
- 图例必须单行显示 `READ / WRITE / R_LAT / W_LAT`
- 不得把带宽和 latency 拆成两个独立主面板

---

## 4.9 日志

### FR-32 完整落盘
系统必须为任务原始日志完整落盘。

### FR-33 UI 最近 N 行缓存
系统前端仅需缓存最近 N 行日志，避免内存压力。

### FR-34 日志弹框
系统必须提供 terminal 风格日志弹框，视觉接近 `tail -f -n 500`。

### FR-35 日志功能
日志弹框 P0 至少支持：
- 自动滚动
- 暂停自动滚动
- 搜索
- 按阶段过滤
- error 高亮

### FR-35A 日志内容范围
Log Viewer 必须同时展示：
- 系统事件日志：Task created、Preview confirmed、DB validation、SSH available / unavailable、Phase started / finished、Stop requested
- benchmark tool 原始日志：stdout、stderr、各阶段完整回显、summary / statistics、原始错误输出

要求：
- benchmark tool 原始日志为主内容，不得只保留阶段摘要。
- run 结束后的总结性输出必须可查看。
- 保持“UI 只看 tail / 磁盘完整落盘”的策略，不得把前端缓存改为无限增长。

### FR-36 默认展示最近日志
日志弹框首次打开时，应默认展示最近 500 行日志。

### FR-36A 页面内默认日志展示策略
Tasks 页面默认不得展示 `Log Output` 摘要区块。

要求：
- 页面内不默认渲染日志内容
- 日志查看能力仅通过顶部 `Open Log Viewer` 按钮进入
- 不得改坏现有日志弹框、日志加载与日志落盘能力

---

## 5. 非功能需求

### NFR-1 可读性
页面结构必须清晰区分：
- 任务配置区
- 队列区
- 监控区
- 日志入口与日志弹框

指标命名要求：
- TPS / TPM 表示当前吞吐与 avg / max，并内置历史水平 / 稳定性图形
- CPU / DISK IO 表示系统负载与 IO 状态；CPU 至少包含 user / sys / iowait / st

布局要求：
- 页面采用左侧配置、顶部控制、右侧监控优先、日志按需弹出
- Current Task 大卡片移除，改为顶部极简状态条

### NFR-2 性能
日志 UI 不得因长日志阻塞页面主渲染。

### NFR-3 可追溯性
任务一旦创建，必须保留 template snapshot、connection snapshot 和解析参数，以便后续追溯。

### NFR-4 降级友好
SSH 不可用时，页面必须保持可用，只禁用系统指标。
SSH 可用时，系统监控区应直接展示图表，不重复强调 `SSH active`。
Capacity 模块已删除，不再保留 Data Disk / Log Disk / SOE / UNDO / TEMP 卡片位置。
CPU steal time 不可用时，统一按安全兜底处理，不得为此引入独立失败分支。

### NFR-5 可扩展性
任务状态、阶段历史、指标统一层设计必须可扩展到 History / Reports 页面复用。

---

## 6. 数据规则

### DR-1 Task 独立实体
Task 必须独立于 Template 持久化。

### DR-2 Snapshot 固化
Task 创建时必须固化：
- template snapshot
- connection snapshot
- resolved runtime params

### DR-3 前端不可直接依赖工具原始输出
前端只能依赖统一指标模型，不应在模板层直接判断 sysbench / swingbench / hammerdb 的原始输出字段。

### DR-4 elapsed / duration 规则
时间显示必须遵循：
- 运行中：`elapsed = now - startedAt`
- success / failed / stopped 后：`duration = completedAt - startedAt`

---

## 7. UI 规则

### UR-1 布局
页面采用：
- 左侧配置与队列
- 右侧监控
- 顶部主操作区中的日志入口

### UR-2 系统指标降级
当 SSH 不可用时：
- 系统指标区置灰
- 顶部或区域内应有文案提示

### UR-3 Task Queue 展示
必须清晰区分：
- 当前运行任务
- 等待中的任务
- 已结束任务（可选，仅少量最近记录）

---

## 8. 验收标准

### AC-1 创建任务
可从页面选择 Template 和 Connection，并成功创建任务。

### AC-2 Preview
点击 Start 前必须能看到预览信息。

### AC-2A Tasks 草稿恢复
切换到其他 tab 再返回 Tasks 页面时，当前 draft 不丢失；若关联连接或模板失效，仅清除失效字段。

### AC-3 队列
当已有运行任务时，新任务进入队列而不是并发运行。

### AC-4 自动推进
当前任务结束后，下一个 queued 任务会自动开始。

### AC-5 DB 阻断
DB 校验失败时，任务不会启动。

### AC-6 SSH 降级
SSH 校验失败时，任务仍能执行，但系统指标区提示不可用。

### AC-7 业务指标
运行中可看到 TPS / TPM 的 current / avg / max。

### AC-8 监控图形
运行中可看到 TPS / TPM 当前值主显示、avg / max 辅助信息，以及卡内历史表现图。

### AC-8A 顶部状态条
顶部极简状态条能正确显示当前状态、真实 phase 和 elapsed / duration；任务结束后时间不再继续增长。

### AC-9 日志
可打开 terminal 风格日志弹框，支持自动滚动、暂停、搜索、阶段过滤、error 高亮。

### AC-9A 原始 benchmark 日志
sysbench / swingbench / hammerdb 的 stdout、stderr、summary / statistics 能在日志 viewer 中看到，且 run 结束后的总结输出未被摘要截断。

### AC-10 Stop
点击 Stop 可停止当前任务，任务不自动 cleanup。

### AC-11 命名收敛
Tasks 页面不再在 Create Task 区展示 Task Name，监控区改为 4 类核心图表 / 面板，避免拆出重复的 TPS / TPM 趋势区。

### AC-12 监控优先收敛
Current Task 大卡片已移除，Start / Stop 为顶部主操作按钮，监控区获得更大展示空间。
