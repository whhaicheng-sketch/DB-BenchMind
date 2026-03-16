# Tasks & Monitor 实施计划 v0.1

## 1. 计划目标

按最小可交付闭环实现 **Tasks & Monitor 第一阶段**：

- 先打通任务创建、排队、执行、停止主链
- 再补统一监控与日志展示
- 最后做 Swingbench Overview 风格的监控总览收口和回归验证

---

## 2. 实施策略

遵循以下原则：

1. **先模型后界面**：先补 Task / Queue / Runtime 状态，再做页面。
2. **先执行链后可视化**：先保证任务能可靠运行，再做监控图形。
3. **先统一层后适配**：先定义统一指标结构，再接各 benchmark tool。
4. **先日志落盘后终端视图**：先保证日志可追溯，再做 tail 风格 viewer。

---

## 3. 阶段划分

## Phase 1：任务实体与队列基础

### 目标
建立 Task 独立模型与单执行器队列。

### 产出
- Task 实体
- Task 状态机
- Queue store / scheduler
- 任务 snapshot 结构

### 核心点
- Task 与 Template 解耦
- 支持 queued / running / success / failed / stopped
- 同时只允许一个任务活跃

---

## Phase 2：任务创建与 Preview

### 目标
用户可在 Tasks & Monitor 页面创建任务，并通过 Preview 确认。

### 产出
- Create Task Card
- Preview Modal
- 后端 Task Name 自动生成
- 少量 Runtime Overrides

### 核心点
- 从页面直接选择 Template 和 Connection
- 支持 prepare / run / cleanup / `Run Full Flow (Prepare -> Run -> Cleanup)`
- Confirm 后自动启动或入队
- 不在 Create Task 区单独展示 Task Name
- 任务草稿在 tab 切换后可恢复，优先采用轻量本地持久化

---

## Phase 3：执行前验证与任务执行链

### 目标
打通 DB 校验、SSH 校验与任务执行动作。

### 产出
- DB validation flow
- SSH soft-check flow
- prepare / run / cleanup / full pipeline 执行链
- Stop 逻辑

### 核心点
- DB 校验失败必须阻断执行
- SSH 校验失败只影响监控
- Stop 不自动 cleanup
- Full pipeline 下 run failed / stopped 不自动 cleanup

---

## Phase 4：日志系统

### 目标
打通完整日志落盘与前端 tail 风格查看器。

### 产出
- 任务日志落盘
- 最近 N 行缓存策略
- 页面日志摘要区
- Terminal 风格日志弹框

### 核心点
- 原始日志完整落盘
- benchmark tool stdout / stderr / summary 完整落盘
- viewer 同时展示系统事件日志与 benchmark 原始输出
- UI 仅保留最近 N 行
- 默认弹框展示最近 500 行
- 支持自动滚动、暂停、搜索、过滤、error 高亮

---

## Phase 5：统一指标层与监控接入

### 目标
为不同 benchmark tool 建立统一指标适配层。

### 产出
- 统一指标模型
- Sysbench 适配
- Swingbench 适配
- HammerDB 适配

### 核心点
- 前端只消费统一后的 TPS / TPM / CPU / Disk IO 指标
- Sysbench TPM 支持由 TPS × 60 计算
- 系统指标仅在 SSH 可用时输出
- 监控区优先占用右侧主要空间
- Disk latency 继续使用 SSH + `/proc/diskstats` 差值法计算，不引入额外系统依赖
- CPU 需要补充 steal time 并纳入统一指标模型

---

## Phase 6：监控 UI 实现

### 目标
将统一指标渲染为实时监控工作台。

### 产出
- 当前任务头部信息区
- TPS / TPM 吞吐卡（当前值 + 历史 / 稳定性图）
- CPU 综合监控图
- DISK IO 四线同图（read / write / r_lat / w_lat）
- 基于 `1280 × 800` 的固定视口紧凑布局

### 核心点
- 业务指标区与系统指标区分离，但整体作为总览监控板呈现
- TPS / TPM 不再拆独立趋势区，历史表现收敛进卡片内部，并保留明显波动线
- SSH 不可用时系统指标降级展示；SSH 正常时弱化提示
- 历史窗口默认沿用现有 5 分钟
- Disk latency 与带宽共处同一 Disk IO 图，但通过双轴分离
- DISK IO 图例收敛为单行显示
- 页面空间优先让给 CPU / DISK IO
- 顶部主操作区包含 Start / Stop / Open Log Viewer
- 页面内默认不展示 Log Output，占高优先让位给首屏操作与监控
- 当前阶段默认窗口调整为 `1200 × 800`，并优先把右侧 Monitor Overview 作为主展示区
- 左侧 Create Task 区固定宽度为 `250px`
- `1280 × 800` 固定视口紧凑收口后置，后续单独完成
- 全局字号与区块留白进一步收敛
- Capacity 模块正式移除，相关采集、绑定、文档同步删除

---

## Phase 7：联调与回归

### 目标
完成端到端链路验证。

### 产出
- 单任务执行验证
- 多任务排队验证
- Stop 验证
- SSH 降级验证
- 日志验证
- 指标展示验证

---

## 4. 模块拆分建议

### 4.1 前端模块
建议拆分为：
- `TasksMonitorPage`
- `TaskCreateCard`
- `TaskPreviewModal`
- `TaskReadinessCard`
- `TaskQueueCard`
- `CurrentTaskHeader`
- `BusinessMetricsPanel`
- `SystemMetricsPanel`
- `ThroughputOverviewPanel`
- `TaskLogSummary`
- `TaskLogViewerModal`

### 4.2 Store / State
建议拆分为：
- task store
- queue store
- monitoring store
- log store

### 4.3 后端模块
建议拆分为：
- task domain
- task usecase
- queue scheduler
- execution orchestrator
- benchmark adapter
- metrics collector
- log writer / log tail provider

---

## 5. 实施顺序

建议按以下顺序推进，不要打乱：

1. Task / Queue 数据结构
2. 任务创建 + Preview
3. 调度器 + 状态流转
4. DB / SSH 验证
5. prepare / run / cleanup / full pipeline 执行链
6. Stop
7. 日志落盘与 tail provider
8. 统一指标层
9. 监控 UI
10. 联调回归

---

## 6. 风险与控制

### 风险 1：Task 与 Template 耦合过深
**控制**：任务创建时强制固化 snapshot。

### 风险 2：前端直接依赖不同工具原始输出
**控制**：必须通过统一指标层。

### 风险 3：日志导致页面卡顿
**控制**：日志完整落盘，前端只保留最近 N 行。

### 风险 4：Stop 与 Cleanup 语义混淆
**控制**：文案与状态机明确区分，Stop 不自动 cleanup。

### 风险 5：SSH 失败误阻断任务
**控制**：将 SSH 定义为软依赖，失败只影响系统监控。

### 风险 6：SSH 采样缺失 steal time
**控制**：CPU steal time 采样缺失时统一降级为 0，不中断监控链路，也不打乱图例和图表。

### 风险 7：固定视口下新增 CPU 指标导致头部拥挤
**控制**：仅最小调整 CPU 图例与摘要密度，不改 TPS / TPM、DISK IO 和 Create Task 既有视觉体系。

---

## 7. 回归策略

### 7.1 核心回归链路
至少验证以下场景：

1. 选择 Template + Connection 创建 run 任务并立即启动。
2. 创建第二个任务进入队列，首个任务结束后自动启动。
3. DB 连接失败时任务被阻断。
4. 启用 SSH 但 SSH 失败时，任务仍能运行，系统指标置灰。
5. full pipeline 中 run 被 stop 后不自动 cleanup。
6. 日志弹框支持 tail 风格查看。
7. TPS / TPM 能持续刷新，Throughput 卡内历史图正常更新。
8. tab 切换后返回 Tasks 页面，draft 不丢失。
9. viewer 可看到 run 结束后的 summary / statistics。
10. 任务结束后顶部时间显示固定 duration，不再持续增长。
11. 当前默认 `1200 × 800` 窗口下，TPS / TPM / CPU / DISK IO 可读可用，且无新增横向滚动、遮挡、裁切、重叠。
12. 后续单独完成 `1280 × 800` 固定视口收口。
13. 顶部可直接看到 Start / Stop / Open Log Viewer / READY。
14. CPU 颜色符合绿 / 红 / 蓝 / 黄规则，并包含 st。
15. DISK IO 为单面板双层展示：带宽 + 时延。
16. Capacity 模块已正式移除，不再保留前后端相关项。

### 7.2 工具适配验证
至少验证：
- sysbench
- swingbench
- hammerdb

重点验证：
- TPM / TPS 解析是否正确
- Sysbench 的 TPM 衍生值是否正确
- Disk latency 为采样差值计算而非累计值直显
- 多盘 Disk latency 聚合为按 I/O 次数加权而非简单平均

---

## 8. 里程碑定义

### M1：任务创建与队列可用
- 能创建任务
- 能预览
- 能入队
- 只允许一个运行任务

### M2：执行链路可用
- prepare / run / cleanup / `Run Full Flow (Prepare -> Run -> Cleanup)` 可触发
- DB / SSH 校验规则正确
- Stop 语义正确

### M3：监控与日志可用
- TPS / TPM 可见
- SSH 可用时 CPU / Disk 可见
- 日志弹框可用
- `Current Throughput` / `System Metrics` 命名清晰
- benchmark 原始 stdout / stderr / summary 可查看
- 顶部极简状态条可用
- Start / Stop 为显著主操作按钮

### M4：端到端回归通过
- 多工具、多场景验证通过
