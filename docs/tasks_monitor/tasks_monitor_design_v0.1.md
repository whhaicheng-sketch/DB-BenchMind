# Tasks & Monitor 设计文档 v0.1

## 1. 文档目的

本文档用于沉淀 **Tasks & Monitor 第一阶段** 的页面职责、交互边界、状态机、数据模型和监控表现方案。

当前阶段的目标不是完善 History、Reports，也不是继续打磨 Templates 页面细节，而是打通以下业务主链：

> Template / Connection 选择 → Task 创建预览 → 入队 / 启动 → Prepare / Run / Cleanup 执行 → 实时监控 → 日志查看

---

## 2. 页面定位

Tasks & Monitor 在当前阶段的定位明确为：

> **压测执行工作台 + 单执行器队列中心**

它不是单纯的任务列表页，也不是模板编辑页。

该页面需要承担以下职责：

1. 基于 Template + Connection 创建任务。
2. 支持任务预览确认。
3. 支持多个任务排队，但同一时刻只允许一个压测任务运行。
4. 支持执行 `prepare / run / cleanup / full pipeline`。
5. 运行中展示业务指标和系统指标。
6. 提供可视化趋势与终端风格日志查看能力。

---

## 3. 设计原则

### 3.1 主链优先
优先确保“可跑、可看、可停、可排队、可回溯”的主链闭环，不优先处理高级导出、复杂权限、结果报告。

### 3.2 Task 独立于 Template
Template 只负责定义压测模板，不存储运行态。Task 是独立实体，拥有自己的快照、状态和日志。

### 3.3 单执行器，多任务队列
系统允许创建多个任务，但同一时刻只允许一个任务进入活跃执行态，其余任务进入等待队列。

### 3.4 DB 校验硬阻断，SSH 校验软降级
- DB 连接校验失败：任务不可执行。
- SSH 校验失败：任务仍可执行，但 CPU / Disk 监控不可用。

### 3.5 监控统一层
不同压测工具的原始输出差异不直接暴露给前端。前端只消费统一后的指标模型，例如：
- TPS
- TPM
- CPU User / Sys / IOWait / Steal
- Disk Read / Write
- Disk Read Latency / Write Latency

### 3.6 日志完整落盘，前端仅维护尾部视图
原始日志必须完整落盘；UI 仅缓存最近 N 行，避免内存与渲染压力失控。

---

## 4. 当前阶段范围

### 4.1 本阶段要做

1. 在 Tasks & Monitor 页面中选择 Template 和 Connection。
2. 支持 Start Task 前的 Preview。
3. 支持多任务排队。
4. 支持 `prepare / run / cleanup / full pipeline` 四类动作。
5. 支持运行前 DB 连接验证。
6. 支持 SSH 可选监控。
7. 展示业务指标：TPS / TPM。
8. 展示系统指标：CPU / Disk IO（仅 SSH 可用时）。
9. 提供面向总览监控板的吞吐历史带与综合系统图。
10. 提供 terminal 风格日志弹框。

### 4.2 本阶段不做

1. 多任务并发压测。
2. 复杂任务编辑器。
3. 结果报告导出。
4. 高级任务重排和优先级策略。
5. 系统资源监控的历史长周期分析。
6. 更复杂的 ACL / shared 权限体系。
7. Import / Export。

---

## 5. 页面结构设计

页面采用 **左侧配置、顶部控制、右侧监控优先、日志按需弹出** 的结构。

固定视口收敛约束：
- 当前默认初始化窗口调整为 `1200 × 800`
- `1280 × 800` 固定视口紧凑布局目标未取消，但后置到后续单独的固定化界面收口阶段
- 首屏无需纵向滚动即可完成主要操作与监控查看
- 页面应优先表现为固定尺寸工作台，而非自然文档流页面
- 默认首屏必须可见：标题、主操作、状态条、Create Task、TPS、TPM、CPU、DISK IO、Open Log Viewer
- 页面内默认不展示日志摘要区，日志通过弹窗按需查看
- Create Task / TPS / TPM / CPU / DISK IO 需进一步紧凑化，整体字号统一缩小

### 5.1 左侧：任务配置与队列区

#### A. Create Task Card
用于创建任务，主要字段：

- Template 选择
- Connection 选择
- Action 选择
  - Prepare
  - Run
  - Cleanup
  - Run Full Flow（Prepare -> Run -> Cleanup）
- Runtime Overrides（仅开放少量可覆盖参数）
  - threads（sysbench）
  - virtual users（Swingbench / HammerDB）
  - duration
- Preview 按钮
- Start Task 按钮

说明：
- Task Name 仍由后端自动生成，用于执行、日志和状态追踪。
- Create Task 区不再单独展示或强调 Task Name 配置项。
- Tasks 页面草稿需在 tab 切换后保留，连接或模板失效时仅清理对应失效字段。
- 左侧 Create Task 区当前固定宽度为 `250px`，右侧空间优先留给监控板。
- 在当前 `1200 × 800` 默认窗口下，Create Task 与监控区应优先保证可读可用，且右侧 Monitor Overview 保持高度优先级。
- 后续固定化收口阶段，仍需回到 `1280 × 800` 视口下的紧凑表单密度目标，不改变字段顺序与联动行为。

#### B. Readiness / Validation Card
显示当前任务是否满足执行条件：

- Template selected
- Connection selected
- DB connection valid
- SSH monitoring available / unavailable
- Action supported by template
- Runtime overrides valid

注意：不应只在点击 Run 时才报错，而应显式展示 readiness 状态。

#### C. Queue Card
展示当前运行中的任务和等待队列。

每个任务条目展示：
- Task Name
- Template Name
- Connection Name
- Action
- Status
- Queue Position
- Created Time
- View Preview
- Cancel Queue（仅 queued）
- Stop（仅当前 running）

---

### 5.2 右侧：实时监控区
当前阶段不再保留右上 `Current Task` 大卡片，释放高度给监控图表。

顶部改为一行极简状态条，显示：
- 状态
- 当前阶段
- 时间（运行中显示 elapsed；结束后显示固定 duration）
- 可选次要信息：benchmark tool / db type

顶部主操作区必须包含：
- Start
- Stop
- Open Log Viewer
- READY / Warning / Blocked 状态入口

要求：
- 顶部字号、按钮尺寸、上下留白进一步压缩
- 在当前 `1200 × 800` 默认窗口下，上述操作必须首屏清晰可见

#### B. 监控总览板
当前阶段监控总览板扩展为 4 类核心图表 / 面板：
1. TPS
2. TPM
3. CPU
4. DISK IO
固定布局为：
- 第一行：TPS、TPM
- 第二行：CPU、DISK IO

#### C. TPS / TPM
每个吞吐卡包含：
- 当前值主显示
- Avg / Max 辅助信息
- 带填充感的历史 / 稳定性图形
- 明显可辨识的波动轮廓线，参考 Swingbench Overview 风格

布局约束：
- 吞吐卡在保留主数值视觉重心的前提下进一步压缩高度
- avg / max 与历史图保留但采用更紧凑的间距和字号
- 历史图中的波动线不得弱化到难以观察
- 可保留轻量 fill，但必须保留清晰的轮廓线

适配规则：
- HammerDB / Swingbench 若原生给出 TPM，则直接使用。
- Sysbench 若默认只有 TPS，则 TPM 使用 `TPS × 60` 的统一转换结果。

#### D. CPU
CPU 图表继续使用现有 SSH 采集与现有指标含义，不改数据来源和计算逻辑。

展示规则：
- User 使用绿色
- Sys 使用红色
- IOWait 使用蓝色
- Steal Time 使用黄色
- 图例颜色与曲线颜色一致
- 深色主题下保证足够对比度

#### E. DISK IO
DISK IO 继续作为单个面板，本轮收敛为四条线同图：
- Read
- Write
- Read Latency
- Write Latency

约束：
- read / write 使用带宽坐标轴
- read latency / write latency 使用时延坐标轴
- 带宽与时延必须采用双轴，不得强行共用同一标尺
- 图例必须单行显示 `READ / WRITE / R_LAT / W_LAT`
- 摘要区继续显示 `read / write / r_lat / w_lat`
- 布局必须紧凑，不得把单卡高度拉成长页面

Disk Latency 方案继续保持：
- 继续使用 SSH 远程执行读取 Linux `/proc/diskstats`
- 基于相邻两次采样差值计算平均时延，不直接展示累计值
- 多磁盘聚合按 I/O 次数加权：`sum(deltaTimeMs) / sum(deltaCompleted)`

固定视口约束：
- 当前 `1200 × 800` 默认窗口下优先保证 4 类面板可读可用，并把额外高度优先分配给图表绘图区
- 后续仍需单独完成 `1280 × 800` 固定视口收口
- CPU / DISK IO 必须优先保留可读面积
- 不依赖纵向滚动查看监控区

---

### 5.3 日志查看

#### A. 页面内默认不展示 Log Output
当前阶段页面首屏不再保留默认日志摘要区，不在主布局内占用固定高度。

#### B. Open Log Viewer Button
`Open Log Viewer` 作为顶部主操作区按钮之一，用户按需打开日志弹框。

#### C. Log Viewer Modal
采用 terminal 风格：
- 黑底
- 等宽字体
- 类似 `tail -f -n 500` 的视觉效果

P0 功能：
- 自动滚动到底部
- 暂停自动滚动
- 搜索
- 按阶段过滤
- error 高亮
- 默认打开时显示最近 500 行

---

## 6. 核心交互流程

### 6.1 创建任务流程

1. 用户选择 Template。
2. 用户选择 Connection。
3. 用户选择 Action。
4. 用户可调整少量运行参数。
5. 用户点击 Preview。
6. 弹出 Preview 弹框，展示：
   - Template 摘要
   - Connection 摘要
   - Action
   - Runtime Overrides
   - 预估监控能力（是否有 SSH）
7. 用户确认后：
   - 若当前无运行任务：立即启动。
   - 若当前有运行任务：进入队列。

### 6.2 执行前校验流程

所有 `prepare / run / cleanup / full pipeline` 在进入执行态前都必须做：
- DB Connection Test（硬阻断）

若启用了 SSH：
- SSH 校验（软校验）
- SSH 校验失败不阻断任务，只禁用系统监控

### 6.3 Stop 流程

Stop 仅停止当前任务的当前阶段执行进程，不自动 cleanup，不自动删除数据。

Stop 后：
- 当前任务进入 `stopped`
- 队列调度器自动拉起下一个 queued 任务

### 6.4 Full Pipeline 流程

UI 文案统一使用 `Run Full Flow (Prepare -> Run -> Cleanup)`。

其语义保持不变：
1. prepare
2. run
3. cleanup

当前阶段建议规则：
- prepare 成功后自动进入 run
- run 正常结束后自动进入 cleanup
- run failed / stopped 时，不自动 cleanup，由用户自行决定是否执行 cleanup

### 6.5 顶部状态条时间规则

- 运行中：`elapsed = now - startedAt`
- success / failed / stopped 后：`duration = completedAt - startedAt`
- 已结束任务不得继续增长 elapsed

---

## 7. 状态机设计

### 7.1 Task 状态

推荐状态集：
- `draft`（当前页面草稿，可本地持久化，用于 tab 切换与刷新恢复）
- `queued`
- `starting`
- `preparing`
- `running`
- `cleaning`
- `stopping`
- `stopped`
- `success`
- `failed`

### 7.2 当前阶段字段

建议单独维护 `currentPhase`：
- `none`
- `prepare`
- `run`
- `cleanup`

### 7.3 阶段结果记录

建议维护 phase history：
- phase
- startedAt
- finishedAt
- result
- summary

这样才能支持：
- Stop 后判断是否可直接继续 run
- Full Pipeline 的阶段回溯
- 后续历史页复用

---

## 8. 任务与模板关系设计

### 8.1 Template 不存运行状态
Template 只存定义信息：
- benchmark tool
- 参数模板
- capability
- 说明与标签

### 8.2 Task 创建时固化快照
Task 创建后必须固化：
- Template Snapshot
- Connection Snapshot
- Resolved Runtime Params

原因：
- 避免队列中等待的任务受后续模板编辑影响
- 保证实际执行参数可追溯

---

## 9. 监控视觉方案

### 9.1 设计目标
你给出的参考图说明本页监控不是传统 BI 仪表盘，而是“概览型运行看板”：
- 一眼看当前水平
- 一眼看是否稳定
- 能感知突刺、锯齿、波动

### 9.2 实现方式

每个核心指标建议拆成两层：

#### 第一层：当前值主数字
作用：表达当前吞吐水平，成为视觉重心。

#### 第二层：历史 / 稳定性图
作用：表达最近一段时间内的稳定性与波动。

这样可以满足：
- 稳定时图形平稳、连续
- 波动时历史图出现突刺、锯齿

### 9.3 CPU 与 Disk 表达
- CPU：建议使用一张综合折线图同时展示 user / sys / iowait / st
- Disk：建议使用一张综合折线图同时展示 read / write / r_lat / w_lat，并保持双轴分离

---

## 10. 日志设计原则

### 10.1 后端日志
- 原始日志完整落盘
- 按任务拆分日志文件
- 支持阶段标记
- 必须保留 benchmark tool 的完整 stdout / stderr
- 必须保留 run 结束后的 summary / statistics

### 10.2 前端日志
- 仅缓存最近 N 行
- 默认保留最近 10,000 行或 50,000 行中的一个固定值
- 首次打开弹框默认显示最近 500 行
- viewer 仅查看 tail，但磁盘日志必须可追溯完整原始输出

### 10.3 日志过滤
建议每条日志具备：
- timestamp
- phase
- stream / level
- line

便于实现：
- 阶段过滤
- error 高亮
- 搜索

---

## 11. 队列设计原则

### 11.1 单运行约束
同一时刻只允许一个任务处于活跃执行态：
- starting
- preparing
- running
- cleaning
- stopping

### 11.2 队列能力
当前阶段建议支持：
- 查看队列
- 取消排队
- 自动推进下一任务

当前阶段可不支持：
- 队列重排序
- 调整优先级
- 队列内直接编辑参数

### 11.3 队列中的任务变更策略
建议第一版不支持直接编辑队列任务；需要修改时：
- 取消排队
- 重新创建任务

---

## 12. 当前阶段的硬性结论

1. Tasks & Monitor 第一阶段应优先打通主链，而不是继续做页面装饰。
2. 任务是独立实体，必须有快照和状态机。
3. 支持多个任务，但同一时刻只能跑一个。
4. DB 校验是硬门槛；SSH 校验只影响监控。
5. 监控分为业务指标区和系统指标区。
6. 业务指标统一成 TPS / TPM。
7. 日志必须完整落盘，前端只维护尾部窗口。
8. 页面结构采用“左控右监、顶部主操作、日志按需弹出”的固定视口工作台。
