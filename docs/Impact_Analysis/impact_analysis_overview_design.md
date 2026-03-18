# Impact Analysis 概要设计

版本：v0.1  
状态：草案  
适用项目：DB-BenchMind

---

## 1. 设计目标

在 DB-BenchMind 中以最小侵入方式嵌入独立的 `Impact Analysis` 模块，使其能够承载 MySQL 高可用 PoC 的实时分析页面，并与现有 Connections、导航体系和未来后端分析链路对接。

---

## 2. 总体设计原则

### 2.1 独立模块原则
Impact Analysis 前后端应作为独立模块建设，不直接揉进现有 benchmark 逻辑。

### 2.2 MySQL 特性隔离原则
MySQL cluster connection 改造通过数据库类型条件分支控制，避免影响其他数据库。

### 2.3 页面优先原则
第一期优先建设稳定、清晰、可演示的页面结构，后端分析逻辑延后补齐。

---

## 3. 系统结构设计

## 3.1 模块结构
建议结构：

- 导航层：新增 `Impact Analysis`
- 连接层：扩展 MySQL cluster connection
- 页面层：Impact Analysis 单屏大盘
- 数据层：session / event / trend / cluster status 模型
- 服务层：Start Analysis / Stop Analysis 协议预留

---

## 3.2 前端结构建议

建议新增独立目录：

- `frontend/src/modules/impact-analysis/pages`
- `frontend/src/modules/impact-analysis/components`
- `frontend/src/modules/impact-analysis/store`
- `frontend/src/modules/impact-analysis/services`
- `frontend/src/modules/impact-analysis/mock`
- `frontend/src/modules/impact-analysis/constants`

### 建议组件拆分
- `ImpactAnalysisPage`
- `ImpactAnalysisToolbar`
- `ImpactSummaryCards`
- `ImpactTrendChart`
- `ClusterStatusPanel`
- `ImpactEventStream`
- `ImpactEmptyState`

---

## 4. 页面设计

## 4.1 页面流转

### 状态 A：空状态
无可用 MySQL cluster connection，显示引导。

### 状态 B：待开始
已有配置，展示控制栏和 Start Analysis。

### 状态 C：分析中 / 已完成
展示完整大盘。

---

## 4.2 页面分区设计

### A. 顶部控制栏
职责：定义“本次分析对象和模式”。

建议内容：
- Cluster Connection 选择器
- VIP
- Primary Node
- Secondary Node
- Connection Mode
- Workload
- Write Rate
- Start Analysis
- Stop（预留）

### B. Summary Cards
职责：第一屏输出核心结论。

建议卡片：
- Business Interruption
- RTO
- Consistency
- Commit Success / Error Summary

### C. 主趋势图
职责：表达“切换事件 → 业务影响 → 恢复”的关系。

建议数据：
- Success TPS
- Error Count / Error Rate
- 关键事件标记线

### D. Cluster Status Panel
职责：表达当前集群角色与节点状态。

建议内容：
- VIP
- Current Primary
- Current Secondary
- Node Status
- Last Role Switch Time

### E. Event Stream
职责：实时展示过程关键事件。

建议事件：
- Start analysis
- Error spike
- Success TPS = 0
- Connection failure
- VIP switch detected
- Role changed
- Recovery detected
- Consistency passed / failed

---

## 5. 数据模型设计（概要）

## 5.1 Connection Model（MySQL 扩展）

建议在 MySQL connection 中支持：

- `vip`
- `primaryNodeIp`
- `secondaryNodeIp`
- `nodes[]`（可选，为扩展预留）
- `defaultConnectionMode`
- `defaultWorkloadType`
- `defaultWriteRate`
- `consistencyCheckEnabled`

说明：非 MySQL 不要求启用这些字段。

---

## 5.2 Analysis Session

建议建模一次实时分析会话：

- `sessionId`
- `connectionId`
- `status`
- `startTime`
- `workloadType`
- `connectionMode`
- `writeRate`
- `successCommitCount`
- `selectSuccessCount`
- `errorCount`
- `interruptionDurationMs`
- `rtoMs`
- `consistencyResult`

---

## 5.3 Runtime Event

建议建模事件流：

- `timestamp`
- `level`（info / warn / error）
- `type`
- `message`

---

## 5.4 Trend Point

建议主图点位模型：

- `timestamp`
- `successTps`
- `errorCount`
- `eventMarkers[]`

---

## 5.5 Cluster Status

建议建模节点与角色信息：

- `vip`
- `currentPrimary`
- `currentSecondary`
- `nodes[]`
- `lastRoleSwitchTime`

---

## 6. 交互设计

## 6.1 启动流程
用户进入页面后：

1. 选择 cluster connection
2. 确认 workload 与模式
3. 点击 `Start Analysis`
4. 页面进入实时分析状态

## 6.2 空状态引导
若无可用配置，引导用户前往 Connections 完成 MySQL cluster connection 配置。

## 6.3 数据刷新
第一期可先使用 mock 驱动；后续可通过轮询或事件推送更新。

---

## 7. 兼容设计

### 7.1 非 MySQL 兼容
- 非 MySQL connections 表单不启用 cluster 字段
- 非 MySQL 现有行为不变

### 7.2 原功能兼容
- 不入侵 Performance Analysis 的原执行链
- 不改 History 的职责
- 不影响现有 benchmark 结果页

---

## 8. 后续实现预留

后续可在此基础上扩展：

- Long / Short Connection 真实实现
- Insert / Select 真实 workload
- commit 成功采集
- consistency 校验
- failover 探测
- 多节点集群支持
- 报告导出

---

## 9. 一期推荐验收视图

首页完成后，至少应满足：

- 有 Start Analysis
- 有 4 张核心指标卡
- 有主图
- 有节点状态
- 有事件流
- 能通过 mock 数据清晰演示切换过程和业务影响

