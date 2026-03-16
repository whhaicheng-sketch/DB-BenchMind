# Tasks & Monitor 任务拆解 v0.1

## 1. 说明

本文档用于将 Tasks & Monitor 第一阶段落地为可执行任务列表。

原则：
- 任务顺序按实施依赖排列
- 先打通主链，再补监控表现
- 每个任务都应具备明确完成标准

---

## 2. 任务清单

## T1. 建立 Task 领域模型

### 目标
新增独立的 Task 实体，不再让运行态依附于 Template。

### 内容
- 定义 Task 基础字段
- 定义 action
- 定义 status
- 定义 currentPhase
- 定义 phaseHistory
- 定义 template snapshot / connection snapshot / resolved params

### 完成标准
- 代码中存在明确 Task 模型
- Task 能独立持久化或被 store 管理
- 不依赖 Template 的运行态字段

---

## T2. 建立 Queue 模型与单执行器调度器

### 目标
支持多个任务排队，但同一时刻只允许一个任务执行。

### 内容
- 新增 queue state
- 标记 running task
- 管理 queued task 列表
- 当前任务结束后自动推进下一个 queued 任务

### 完成标准
- 同时创建多个任务时，仅一个任务进入活跃态
- 后续任务自动排队
- 当前任务结束后下一个自动启动

---

## T3. 新增 Tasks & Monitor 页面创建区

### 目标
在页面左侧实现任务创建能力。

### 内容
- Template 选择器
- Connection 选择器
- Action 选择器
- 后端 Task Name 自动生成
- Runtime Overrides 输入区

### 完成标准
- 页面中可填写创建任务所需字段
- Create Task 区不单独展示 Task Name
- Runtime Overrides 仅开放少量参数
- tab 切换后返回时 draft 不丢失
- 顶部 Start / Stop 为主操作按钮
- 顶部主操作区同时包含 Open Log Viewer
- 左侧 Create Task 区固定宽度为 `250px`
- 当前默认 `1200 × 800` 窗口下首屏应优先保证 Create Task 与监控区可读可用，且右侧 Monitor Overview 保持主展示区
- `1280 × 800` 固定视口收口后置到后续单独阶段

---

## T4. 实现 Preview Modal

### 目标
Start 前必须先让用户预览任务。

### 内容
- 展示 Template 摘要
- 展示 Connection 摘要
- 展示 Action
- 展示 Resolved Params
- 展示 SSH 监控是否可用的预期说明

### 完成标准
- 点击 Preview 能弹框
- 点击 Confirm 后创建任务
- 当前无运行任务时立即启动，否则进入队列

---

## T5. 实现 Readiness / Validation Card

### 目标
在页面上显式展示执行准备情况。

### 内容
- Template selected
- Connection selected
- Action supported
- Runtime overrides valid
- DB valid（执行前触发）
- SSH monitoring available / unavailable

### 完成标准
- 页面中有独立 readiness 区
- 用户不需要只依赖报错弹窗理解为什么不能执行

---

## T6. 实现 DB 校验逻辑

### 目标
所有动作进入执行态前必须完成 DB Connection Test。

### 内容
- prepare 前 DB test
- run 前 DB test
- cleanup 前 DB test
- full pipeline 启动前 DB test

### 完成标准
- DB 不通时任务不启动
- UI 明确给出失败反馈

---

## T7. 实现 SSH 校验与软降级逻辑

### 目标
SSH 仅影响监控，不影响压测执行。

### 内容
- SSH 配置存在时尝试校验
- SSH 成功则启用系统监控
- SSH 失败则禁用系统监控并提示

### 完成标准
- SSH 失败不阻断任务
- UI 中系统指标区置灰并说明原因

---

## T8. 实现 prepare / run / cleanup / full pipeline 执行链

### 目标
让四种动作都能真实进入执行流程。

### 内容
- prepare 执行
- run 执行
- cleanup 执行
- full pipeline 执行串联

### 完成标准
- 四种动作都能触发
- full pipeline 顺序正确
- run failed / stopped 时不自动 cleanup

---

## T9. 实现 Stop 语义

### 目标
Stop 只停止当前任务，不自动 cleanup。

### 内容
- 当前 running / preparing / cleaning 任务支持 stop
- 停止后状态流转到 stopping → stopped
- Stop 后调度器继续推进后续队列

### 完成标准
- 点击 Stop 后当前任务结束
- 不自动 cleanup
- 队列中的后续任务能继续推进

---

## T10. 实现 Queue Card

### 目标
在页面左侧显示当前任务和排队任务。

### 内容
- 当前 running task 展示
- queued task 列表展示
- queued task 取消按钮
- 查看 preview 摘要入口

### 完成标准
- 队列状态清晰可见
- 支持取消 queued 任务

---

## T11. 建立统一指标模型

### 目标
屏蔽不同 benchmark tool 的输出差异。

### 内容
- 定义统一业务指标：TPS / TPM
- 定义统一系统指标：CPU User / Sys / IOWait / Steal, Disk Read / Write, Disk Read Latency / Write Latency
- 定义 current / avg / max 统计结构

### 完成标准
- 前端不直接依赖各工具原始输出字段
- 统一结构可供监控 UI 使用
- Disk latency 使用 SSH + `/proc/diskstats` 采样差值法表达，而非累计值直显

---

## T12. 实现 sysbench 指标适配

### 目标
将 sysbench 输出适配为统一指标。

### 内容
- 解析 TPS
- 在无原生 TPM 时由 TPS × 60 计算 TPM

### 完成标准
- sysbench 运行时页面可见 TPS / TPM

---

## T13. 实现 swingbench 指标适配

### 目标
将 swingbench 输出适配为统一指标。

### 内容
- 解析 swingbench 业务输出
- 提供 TPS / TPM 统一结果

### 完成标准
- swingbench 运行时页面可见 TPS / TPM

---

## T14. 实现 hammerdb 指标适配

### 目标
将 hammerdb 输出适配为统一指标。

### 内容
- 解析 hammerdb 业务输出
- 提供 TPS / TPM 统一结果

### 完成标准
- hammerdb 运行时页面可见 TPS / TPM

---

## T15. 实现系统监控采集

### 目标
通过 SSH 获取系统资源数据。

### 内容
- 采集 CPU user / sys / iowait / steal
- 采集 Disk read / write
- 采集 Disk read latency / write latency
- 定时推送到前端

### 完成标准
- SSH 可用时页面能看到 CPU / DISK IO
- SSH 不可用时不输出系统类指标，但任务仍可执行

---

## T16. 实现 Top Status Strip

### 目标
在顶部展示极简任务状态信息，不再保留右上大卡片。

### 内容
- status
- currentPhase
- elapsed / duration
- 可选次要信息：tool / db type

### 完成标准
- 用户能一眼识别当前状态与真实 phase
- 任务结束后时间不再继续增长

---

## T17. 实现 Current Throughput Panel

### 目标
展示 TPM / TPS 业务指标，并以总览板风格表达当前值与近期稳定性。

### 内容
- current
- avg
- max
- 当前值主显示
- 带填充感的历史 / 稳定性图形

### 完成标准
- 页面可见 TPS / TPM 实时变化
- avg / max 保留但弱化
- 稳定时表现平稳，波动时能看出突刺/锯齿趋势

---

## T18. 实现 System Metrics Panel

### 目标
将系统指标收敛为 CPU / DISK IO 两类面板。

### 内容
- CPU 综合图同时展示 user / sys / iowait / st，且颜色为绿 / 红 / 蓝 / 黄
- TPS / TPM 历史图保留明显波动轮廓线
- DISK IO 综合图作为单面板四线同图展示：
  - read / write
  - r_lat / w_lat
  - 双轴分离带宽与时延
  - 图例单行显示
- SSH 可用时直接展示图，不重复强调 SSH active
- 固定视口下收紧 panel 高度、图例、摘要与图表留白
- 页面内默认移除 Log Output 区块

### 完成标准
- SSH 可用时指标可见
- SSH 不可用时降级表现正确
- steal time 不可用时 CPU 摘要、图例、图表仍安全展示
- TPS / TPM 波动线清晰可见
- CPU 颜色符合绿 / 红 / 蓝 / 黄
- DISK IO 为单面板四线同图
- READ / WRITE / R_LAT / W_LAT 图例单行显示
- 当前默认 `1200 × 800` 窗口下 4 类监控面板应保持可读可用
- 后续单独完成 `1280 × 800` 固定视口收口
- 页面默认不展示日志摘要，只通过顶部 Open Log Viewer 查看日志

---

## T20. 实现任务日志落盘

### 目标
为每个任务完整保存原始日志。

### 内容
- 按任务拆分日志文件
- 支持阶段标记
- 记录 stdout / stderr
- 保留 benchmark tool summary / statistics

### 完成标准
- 每个任务都有完整日志文件
- 任务结束后日志仍可追溯
- run 结束后的总结输出可回看

---

## T21. 实现顶部日志入口

### 目标
通过顶部主操作区进入日志查看，不在页面首屏默认展开日志内容。

### 内容
- 在顶部主操作区提供 Open Log Viewer
- 与当前任务联动启用 / 禁用
- 不在页面主布局中保留默认日志摘要区

### 完成标准
- 用户可从顶部直接进入日志弹框
- 页面首屏不再被默认日志内容占高

---

## T22. 实现 Terminal 风格日志弹框

### 目标
提供类似 `tail -f -n 500` 的黑底终端风格日志 viewer。

### 内容
- 黑底等宽字体
- 默认展示最近 500 行
- 自动滚动
- 暂停自动滚动
- 搜索
- 阶段过滤
- error 高亮
- 展示系统事件日志
- 展示 benchmark tool 原始 stdout / stderr / summary

### 完成标准
- 日志弹框交互完整可用
- 在长日志场景下仍保持流畅

---

## T23. 实现 UI 最近 N 行缓存策略

### 目标
避免日志过大导致前端卡死。

### 内容
- 前端仅缓存最近 N 行
- 超过阈值时丢弃旧行
- 磁盘仍保留全量

### 完成标准
- 长时间运行不会导致前端日志区失控

---

## T24. 端到端联调：单任务执行

### 目标
验证完整单任务主链。

### 内容
- 创建任务
- Preview
- DB 校验
- 执行 run
- 查看指标
- 查看日志
- Stop

### 完成标准
- 单任务链路完整可用

---

## T25. 端到端联调：多任务排队

### 目标
验证队列逻辑。

### 内容
- 连续创建多个任务
- 验证只有一个任务运行
- 验证后续任务自动推进
- 验证 queued 任务可取消

### 完成标准
- 队列行为符合预期

---

## T26. 工具回归：sysbench / swingbench / hammerdb

### 目标
验证统一指标与执行链在不同工具上都工作正常。

### 内容
- sysbench 回归
- swingbench 回归
- hammerdb 回归

### 完成标准
- 三种工具均能正确展示 TPM / TPS
- SSH 监控开启与关闭均可正确工作

---

## 3. 验收清单

### 页面与交互
- [ ] 页面采用左控右监、顶部主操作、日志按需弹出的固定视口布局
- [ ] 页面采用固定视口紧凑工作台布局
- [ ] 当前默认窗口调整为 `1200 × 800`
- [ ] 后续单独完成 `1280 × 800` 固定视口界面收口
- [ ] 可选择 Template / Connection / Action
- [ ] Create Task 区不单独展示 Task Name
- [ ] 左侧 Create Task 区固定宽度为 `250px`
- [ ] 支持 Preview
- [ ] tab 切换后返回 Tasks 页面，draft 不丢失
- [ ] 当前默认 `1200 × 800` 窗口下核心内容可读可用
- [ ] 顶部清晰可见 Start / Stop / Open Log Viewer / READY

### 队列与执行
- [ ] 可同时创建多个任务
- [ ] 同时仅一个任务运行
- [ ] queued 任务可取消
- [ ] 任务结束后自动推进下一个
- [ ] Stop 不自动 cleanup

### 校验规则
- [ ] DB 校验失败会阻断执行
- [ ] SSH 校验失败不阻断执行
- [ ] SSH 失败时系统指标区置灰并提示

### 监控
- [ ] 显示 TPM / TPS current / avg / max
- [ ] `Throughput Trend` 已整块移除
- [ ] `Current Throughput` 与 `System Metrics` 命名清晰
- [ ] TPS / TPM 在 Throughput 卡内显示当前值 + 历史表现
- [ ] Current Throughput 不再使用普通进度条表现
- [ ] 页面整体字号统一缩小并保持一致
- [ ] SSH 可用时显示 CPU user / sys / iowait / st
- [ ] steal time 不可用时前后端安全兜底
- [ ] SSH 可用时显示 Disk read / write
- [ ] SSH 可用时显示 Disk `r_lat / w_lat`
- [ ] System Metrics 收敛为 CPU / Disk IO 两类监控区
- [ ] TPS / TPM 历史图具有明显波动线
- [ ] Disk IO 单图显示 READ / WRITE / R_LAT / W_LAT 四条线
- [ ] Disk latency 与带宽通过双轴分离，不共用同一标尺
- [ ] READ / WRITE / R_LAT / W_LAT 图例单行显示
- [ ] Capacity 模块已从前端与后端彻底移除
- [ ] 多盘 Disk latency 聚合不是简单平均
- [ ] Current Task 大卡片已移除
- [ ] 顶部极简状态条正确显示 status / phase / elapsed 或 duration
- [ ] 任务结束后时间不再继续增长

### 日志
- [ ] 日志完整落盘
- [ ] 页面内默认不显示 Log Output 摘要
- [ ] 顶部 Open Log Viewer 可打开日志弹框
- [ ] 日志弹框具备 terminal 风格
- [ ] 支持自动滚动、暂停、搜索、阶段过滤、error 高亮
- [ ] sysbench / swingbench / hammerdb 原始 stdout / stderr 在 viewer 可见
- [ ] run 结束后的 summary / statistics 未被摘要截断
