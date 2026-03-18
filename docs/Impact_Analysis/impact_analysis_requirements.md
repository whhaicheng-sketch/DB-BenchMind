# Impact Analysis 需求设计

版本：v0.1  
状态：草案  
适用项目：DB-BenchMind

---

## 1. 背景

现有 DB-BenchMind 更偏向数据库性能压测与运行监控，对高可用 PoC 中“切换对业务造成什么影响”缺少独立而清晰的表达。用户当前希望新增一个独立模块，用于在 MySQL 高可用、主备、容灾类演示中，实时回答以下问题：

- 切换影响了业务多久
- RTO 是多少
- 主备是否已切换完成
- commit 成功的数据是否最终仍然存在
- 切换期间错误、连接失败、TPS 掉 0 等过程是否可见

---

## 2. 需求目标

新增 `Impact Analysis` 模块，重点围绕：

- **RTO**
- **Consistency**
- **业务影响**

构建一套可直接用于 MySQL 高可用 PoC 演示的实时分析界面。

---

## 3. 用户核心诉求

### 3.1 结果层诉求
用户更关心：
- 业务中断多久
- 切换恢复多久
- 是否满足一致性要求
- 客户能否一眼看懂切换与业务影响的关系

### 3.2 过程层诉求
用户希望实时看到：
- TPS 掉 0
- Error 突增
- 连接失败
- 主备角色切换
- 恢复完成
- 一致性校验结果

### 3.3 集成层诉求
用户要求：
- Impact Analysis 作为独立模块嵌入当前项目
- 不能影响 DB-BenchMind 现有功能
- 只优先改造 MySQL
- 其他数据库逻辑不要动

---

## 4. 适用范围

### 4.1 一期适用数据库
- MySQL

### 4.2 一期适用场景
- 高可用 PoC
- 主备切换演示
- 容灾切换演示
- 故障注入后的业务影响观测

### 4.3 一期 workload 范围
- Insert
- Select

---

## 5. 功能性需求

## 5.1 导航与入口
### R-01
系统应新增一级导航 `Impact Analysis`，位于 `History` 后。

### R-02
系统应将 `Tasks & Monitor` 重命名为 `Performance Analysis`，以保证产品命名体系一致。

### R-03
用户应可从顶部导航直接进入 Impact Analysis 页面。

---

## 5.2 Connections 改造
### R-04
系统应将 MySQL 连接模型从单库连接升级为集群连接模型。

### R-05
MySQL 集群连接至少应支持：
- VIP
- Primary Node IP
- Secondary Node IP
- Database Name
- Username / Password
- Port

### R-06
非 MySQL 连接逻辑与现有功能必须保持不变。

---

## 5.3 页面状态
### R-07
Impact Analysis 页面应支持以下状态：
- 无可用配置
- 已配置未开始
- 分析中 / 已完成

### R-08
当存在可用 MySQL cluster connection 时，页面应允许用户直接点击 `Start Analysis` 启动分析。

---

## 5.4 指标与口径
### R-09
业务成功必须以 `commit` 成功为准。

### R-10
系统应定义业务中断时间为“成功事务 TPS 掉到 0 的持续时间”。

### R-11
系统应展示 RTO。

### R-12
系统应展示一致性总结果，至少包括：
- Passed
- Failed

### R-13
系统应支持在过程区域展示：
- TPS 掉 0
- Error 突增
- 连接失败
- 角色切换
- 恢复完成
- 一致性校验完成

---

## 5.5 页面展示
### R-14
页面应以单屏形式展示，并包含：
- 顶部控制区
- 核心指标卡
- 主趋势图
- 节点状态区
- 实时事件流

### R-15
核心指标卡至少包括：
- Business Interruption
- RTO
- Consistency
- Commit Success / Error Summary

### R-16
主趋势图应能体现“切换事件与业务影响关系”。

### R-17
节点状态区应能体现：
- VIP
- Current Primary
- Current Secondary
- 节点状态
- Last Role Switch Time

---

## 6. 非功能性需求

### N-01 兼容性
Impact Analysis 的引入不得影响现有 DB-BenchMind 其他数据库和原有功能链路。

### N-02 独立性
Impact Analysis 应作为独立模块开发，避免与原性能压测模块强耦合。

### N-03 可扩展性
页面与数据模型应为未来支持更多 workload、更多数据库、更多分析明细预留扩展空间。

### N-04 可演示性
第一期即使后端未完全落地，页面也应能通过 mock 数据完成 PoC 演示。

---

## 7. 成功标准

### 7.1 产品成功标准
- 用户进入页面后，一眼能看出业务中断与切换的关系
- 用户能用该页完成高可用 PoC 演示
- 页面语义与 Performance Analysis 明确区分

### 7.2 技术成功标准
- MySQL cluster connection 可配置
- 页面可以独立显示分析流程
- 非 MySQL 回归不受影响

---

## 8. 暂不纳入本期

- 多数据库统一支持
- 导出报告
- 历史列表分析中心
- 一致性明细下钻
- 复杂业务场景建模

