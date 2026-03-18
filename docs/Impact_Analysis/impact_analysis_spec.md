# Impact Analysis 规格说明（impact_analysis_spec）

版本：v0.1  
状态：草案  
适用项目：DB-BenchMind

---

## 1. 文档目标

本文档定义 DB-BenchMind 中新模块 **Impact Analysis** 的产品规格。该模块用于在数据库高可用、主备切换、容灾演练等 PoC 场景下，实时展示切换对业务的影响，并重点围绕 **RTO** 与 **一致性** 两个核心维度进行分析。

---

## 2. 模块定位

Impact Analysis 是一个嵌入在 DB-BenchMind 内部、但职责独立的新工具模块，不属于现有性能压测链路的简单扩展。

### 2.1 模块职责

Impact Analysis 负责：

- 高可用 / 主备 / 容灾 PoC 实时分析
- 切换期间业务中断影响分析
- RTO 观测与展示
- commit 成功事务的一致性结果展示
- 节点状态、VIP、主从角色变化可视化
- 实时事件流打印（例如 TPS 掉 0、错误突增、连接失败、切换恢复）

### 2.2 与现有模块边界

- `Performance Analysis`（原 Tasks & Monitor）：负责性能压测、吞吐和运行监控
- `Impact Analysis`：负责高可用 PoC、切换影响分析、一致性和 RTO

两者职责必须严格分离，避免语义混淆。

---

## 3. 命名与导航

### 3.1 新增模块命名

- 顶部一级导航名称：`Impact Analysis`
- 插入位置：`History` 后面

### 3.2 现有模块改名建议

为保证产品命名对齐，建议将：

- `Tasks & Monitor` 改名为 `Performance Analysis`

推荐顶部导航顺序：

- Connections
- Templates
- Performance Analysis
- History
- Impact Analysis

---

## 4. 一期范围

### 4.1 一期必须支持

- 数据库类型：仅 **MySQL**
- 分析方式：**实时分析**
- 页面形态：**单屏实时分析页**
- workload：仅支持 `Insert` 与 `Select`
- 连接模式：支持 `Long Connection` / `Short Connection`
- 切换影响指标：RTO、业务中断时长、错误趋势、连接失败事件
- 一致性结果：首页只展示总结果（Passed / Failed）
- 节点状态：展示 VIP、主节点、从节点、当前角色状态

### 4.2 一期不做

- 非 MySQL 数据库逻辑扩展
- 导出报告
- 外部压测工具接入
- 复杂业务脚本编排
- 历史分析中心化列表页
- 细粒度一致性明细表

---

## 5. 核心业务定义

### 5.1 业务成功定义

**业务成功以 commit 为基准。**

对于写请求：

- SQL 执行成功但未 commit，不算成功
- 只有 commit 成功，才计入成功事务
- 只有 commit 成功，才进入一致性校验集合

对于读请求：

- Select 成功返回，计为成功读请求

### 5.2 业务中断定义

第一期定义为：

**成功事务 TPS 掉到 0 的持续时间 = 业务中断时间**

其中：

- Insert：以 commit 成功为准
- Select：以成功返回为准

### 5.3 一致性定义

第一期一致性定义为：

**凡是 commit 成功的 Insert 记录，最终必须在切换后数据库中验证存在。**

因此不允许出现：

- 成功提交但数据丢失
- 成功提交但切换后不可见

### 5.4 RTO 定义

RTO 指从切换异常开始，到业务成功事务恢复到可用状态的恢复时间窗口。

建议在界面中与“业务中断时长”同时展示，但语义区分如下：

- Business Interruption：业务侧真实不可用窗口
- RTO：切换恢复完成窗口

---

## 6. 连接模型改造要求

### 6.1 设计原则

Connections 对 MySQL 从“单库连接”升级为“集群连接模型”，但必须满足：

- 仅优先改造 MySQL
- 其他数据库逻辑一律不要动
- 不影响现有功能
- 非 MySQL 保持现有行为和表单逻辑

### 6.2 MySQL Cluster Connection 需要表达的能力

建议字段包括：

#### 基础字段
- Connection Name
- Database Type = MySQL
- Database Name
- Username
- Password
- Port

#### 集群字段
- VIP / Access Endpoint
- Primary Node IP
- Secondary Node IP
- 可扩展节点列表（为后续多节点预留）

#### 分析默认字段
- Connection Mode（Long / Short）
- Workload Type（Insert / Select）
- Read / Write Ratio
- Write Rate Per Second
- Consistency Check Enabled

---

## 7. 页面规格

### 7.1 页面入口

用户路径：

1. 在 Connections 配置好 MySQL Cluster Connection
2. 点击顶部导航 `Impact Analysis`
3. 在页面中选择连接
4. 点击 `Start Analysis`
5. 进入单屏实时分析

### 7.2 页面状态

页面需要支持三种状态：

#### A. 未配置
- 提示无可用 MySQL Cluster Connection
- 引导去 Connections 完成配置

#### B. 已配置未开始
- 展示当前连接摘要
- 展示 Start Analysis

#### C. 分析中 / 已完成
- 展示完整实时分析大盘

### 7.3 页面结构

#### 区块 1：顶部控制栏
- Cluster Connection 选择器
- VIP
- Primary Node
- Secondary Node
- Connection Mode
- Workload
- Write Rate
- Start Analysis
- Stop（预留）

#### 区块 2：核心结果卡片
建议固定展示：
- Business Interruption
- RTO
- Consistency
- Commit Success / Error Summary

#### 区块 3：主趋势图
展示：
- Success TPS
- Error 趋势
- 关键事件标记线

关键事件包括：
- Analysis Start
- Error Spike
- Success TPS = 0
- Connection Failure
- Failover Detected
- Recovery
- Consistency Check Start
- Consistency Check Done

#### 区块 4：节点状态面板
展示：
- VIP
- Current Primary
- Current Secondary
- Node 状态
- Last Role Switch Time

#### 区块 5：实时事件流
展示事件化过程打印，例如：
- Start analysis
- Error rate spike detected
- Success TPS dropped to zero
- Connection failure detected
- VIP switch detected
- Primary role changed
- Recovery detected
- Consistency check passed / failed

---

## 8. 展示原则

### 8.1 产品表达优先
一期重点在前端页面表达清晰准确，不要求复杂后端先完全落地。

### 8.2 客户视角优先
页面必须让用户快速理解：

- 业务断了多久
- 切换多久恢复
- 数据是否一致
- 当前主从是否已切换成功

### 8.3 事件流优先于原始日志
实时区域优先展示“事件流”，而非大段技术原始日志。

---

## 9. 兼容性约束

必须满足：

- 不影响现有 DB-BenchMind 其他功能
- 不改 Oracle / SQL Server / PostgreSQL 现有逻辑
- 不把 Impact Analysis 塞入 Performance Analysis 链路
- Impact Analysis 作为独立模块存在
- 只在 MySQL 连接模型上新增集群表达能力

---

## 10. 一期验收要点

### 功能验收
- 顶部导航可见 Impact Analysis，位置在 History 后
- Tasks & Monitor 改名为 Performance Analysis
- MySQL 连接可配置集群信息
- Impact Analysis 可直接从页面启动分析
- 页面存在实时核心指标卡、主图、节点状态、事件流
- 一致性结果可显示总结果 Passed / Failed

### 边界验收
- 非 MySQL 页面与逻辑行为不变
- 现有性能压测模块不受影响
- 现有其他数据库连接不受影响

