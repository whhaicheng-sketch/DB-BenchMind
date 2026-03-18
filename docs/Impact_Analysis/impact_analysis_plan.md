# Impact Analysis 实施计划（impact_analysis_plan）

版本：v0.1  
状态：草案  
适用项目：DB-BenchMind

---

## 1. 计划目标

在不破坏现有 DB-BenchMind 功能和认知的前提下，新增一个独立的 `Impact Analysis` 模块，用于支持 MySQL 高可用 / 主备 / 容灾 PoC 的实时影响分析，并完成与现有 Connections 的最小侵入式集成。

---

## 2. 实施原则

### 2.1 独立模块、最小侵入
- 新模块独立目录开发
- 与现有 benchmark / monitor 链路解耦
- 仅通过导航和共享连接配置接入

### 2.2 MySQL 优先、其他数据库不动
- 所有新增行为只对 MySQL 生效
- Oracle / PostgreSQL / SQL Server 等逻辑保持原样

### 2.3 页面先行
- 第一优先级：页面结构、视觉表达、模块定位
- 第二优先级：连接模型扩展
- 第三优先级：分析链路和数据采集逻辑

---

## 3. 分阶段计划

## 阶段 1：产品壳与导航接入

### 目标
完成模块露出和页面承载框架。

### 工作内容
- 新增顶部导航 `Impact Analysis`
- 调整 `Tasks & Monitor` 为 `Performance Analysis`
- 新建独立页面模块目录
- 完成空状态、未开始状态、分析中状态三种页面框架
- 搭建单屏大盘布局

### 交付结果
- 页面可进入
- 结构清晰
- 不依赖真实后端也能演示静态状态

---

## 阶段 2：MySQL Connections 集群模型改造

### 目标
让 MySQL 连接从单库模型升级为集群模型，为 Impact Analysis 提供配置基础。

### 工作内容
- 梳理现有 Connection 数据结构
- 增加 MySQL 专属 cluster 字段
- 表单支持 VIP、Primary Node、Secondary Node
- 保持非 MySQL 表单逻辑不变
- 明确旧连接兼容策略

### 风险点
- 连接模型改造可能影响现有表单与序列化
- 必须通过数据库类型分支控制行为

### 交付结果
- MySQL 可表达 cluster connection
- 其他数据库功能不回归

---

## 阶段 3：Impact Analysis 前端完整页面

### 目标
完成面向 PoC 演示的单屏实时分析界面。

### 工作内容
- 顶部控制栏
- 核心指标卡片
- 主趋势图
- 节点状态卡片
- 事件流区域
- 页面状态流转
- mock 数据接入与切换

### 交付结果
- 能清晰表达业务影响与中断关系
- 能清晰表达 RTO 与 Consistency
- 能进行前端演示

---

## 阶段 4：分析链路与数据模型预留

### 目标
为后续真实实现预留前后端接口和数据结构。

### 工作内容
- 定义 analysis session 模型
- 定义 runtime event 模型
- 定义 summary card 数据模型
- 定义 chart 点位模型
- 为 Start Analysis 预留调用协议

### 交付结果
- 页面不再是纯静态壳
- 可挂接未来后端逻辑

---

## 阶段 5：真实分析逻辑（后续）

### 目标
实现 Insert / Select 驱动、commit 成功判定、一致性校验和切换事件识别。

### 工作内容
- 长连接 / 短连接模式实现
- Insert workload 实现
- Select workload 实现
- commit 成功记录采集
- 成功 TPS / Error 统计
- 连接失败 / 恢复事件识别
- 一致性最终校验

### 说明
该阶段不是当前第一优先级，但接口和页面必须为其预留。

---

## 4. 建议开发顺序

### 4.1 推荐顺序
1. 导航与模块壳
2. 页面结构与视觉样式
3. MySQL Connections 集群模型改造
4. mock 数据驱动的前端联动
5. 分析协议与模型预留
6. 后端真实逻辑

### 4.2 不建议顺序
不建议先深挖后端逻辑再做页面，否则很容易在视觉和交互上反复返工。

---

## 5. 目录与模块规划建议

### 前端建议目录
- `frontend/src/modules/impact-analysis/pages`
- `frontend/src/modules/impact-analysis/components`
- `frontend/src/modules/impact-analysis/store`
- `frontend/src/modules/impact-analysis/services`
- `frontend/src/modules/impact-analysis/mock`
- `frontend/src/modules/impact-analysis/constants`

### 后端建议目录（预留）
- `internal/app/usecase/impact_analysis_*`
- `internal/domain/impact_analysis`
- `internal/infra/adapter/mysql_cluster_*`
- `internal/transport/wails/bindings/impact_analysis`

说明：目录命名可根据项目现有规范调整，但原则是不与 benchmark 链路强耦合。

---

## 6. 风险与控制

## 6.1 最大风险

### 风险 1：Connection 改造影响现有功能
控制方式：
- MySQL 专属分支渲染
- 非 MySQL 不进入 cluster 模型逻辑
- 保留兼容字段

### 风险 2：Impact Analysis 与 Performance Analysis 职责混淆
控制方式：
- 页面文案严格区分
- 数据口径严格区分
- 不复用错误的性能页面概念

### 风险 3：前端界面好看但口径不准
控制方式：
- commit 成功定义写入规范
- 中断时间定义写入规范
- consistency 结果定义写入规范

---

## 7. 里程碑建议

### M1：导航与页面壳完成
- 新 tab 可见
- 命名已对齐
- 页面状态齐全

### M2：MySQL Cluster Connection 完成
- Connections 支持 MySQL 集群字段
- 非 MySQL 回归通过

### M3：Impact Analysis 单页演示完成
- 顶部控制区
- summary cards
- chart
- cluster status
- event stream

### M4：真实数据接入协议完成
- session 模型
- event 模型
- chart 模型

### M5：后续真实分析实现
- insert/select
- commit 成功判定
- consistency 校验

---

## 8. 当前阶段建议输出物

第一轮先完成：

- spec
- plan
- tasks
- 需求设计
- 概要设计

然后进入：

- 页面低保真
- 连接模型字段设计
- 前端壳实现

