# Impact Analysis 任务拆解（impact_analysis_tasks）

版本：v0.1  
状态：草案  
适用项目：DB-BenchMind

---

## 1. 任务目标

将 Impact Analysis 作为独立模块嵌入 DB-BenchMind，并完成 MySQL 集群连接模型改造、导航接入、页面壳搭建和后续真实分析能力预留。

---

## 2. 任务分组

## T1. 文档落地
- [ ] 输出 `impact_analysis_spec.md`
- [ ] 输出 `impact_analysis_plan.md`
- [ ] 输出 `impact_analysis_tasks.md`
- [ ] 输出“需求设计”文档
- [ ] 输出“概要设计”文档

交付标准：五份文档内容完整、术语统一、范围一致。

---

## T2. 顶部导航调整
- [ ] 新增顶部导航 `Impact Analysis`
- [ ] 位置放在 `History` 后
- [ ] `Tasks & Monitor` 改名为 `Performance Analysis`
- [ ] 检查导航宽度与样式是否仍正常

交付标准：导航结构可见、命名对齐、不影响现有跳转。

---

## T3. 独立模块目录初始化
- [ ] 创建 impact-analysis 前端独立目录
- [ ] 初始化 page / components / store / services / mock / constants 结构
- [ ] 与现有模块边界分离

交付标准：Impact Analysis 可独立维护，目录职责清晰。

---

## T4. 页面状态壳搭建
- [ ] 空状态页面
- [ ] 已配置未开始状态页面
- [ ] 分析中页面骨架
- [ ] 已完成状态展示占位

交付标准：至少三种状态可切换展示。

---

## T5. MySQL Connections 集群模型设计
- [ ] 梳理现有 connection model
- [ ] 设计 MySQL cluster connection 字段
- [ ] 定义兼容策略
- [ ] 明确旧数据升级方式

交付标准：字段清单完整，兼容方案明确。

---

## T6. Connections 表单改造（仅 MySQL）
- [ ] MySQL 表单支持 VIP
- [ ] MySQL 表单支持 Primary Node IP
- [ ] MySQL 表单支持 Secondary Node IP
- [ ] 支持 cluster connection 的保存与回显
- [ ] 非 MySQL 表单保持不变

交付标准：只影响 MySQL 配置路径，非 MySQL 无回归。

---

## T7. Impact Analysis 顶部控制栏
- [ ] Cluster Connection 选择器
- [ ] VIP 信息展示
- [ ] Primary / Secondary 摘要展示
- [ ] Connection Mode 展示
- [ ] Workload 展示
- [ ] Write Rate 展示
- [ ] Start Analysis 按钮
- [ ] Stop 按钮占位

交付标准：用户进入页面后可直接理解本次分析对象与模式。

---

## T8. 核心指标卡片
- [ ] Business Interruption 卡片
- [ ] RTO 卡片
- [ ] Consistency 卡片
- [ ] Commit Success / Error Summary 卡片

交付标准：第一屏一眼可读，指标名称和口径清晰。

---

## T9. 主趋势图
- [ ] Success TPS 曲线
- [ ] Error 趋势曲线
- [ ] 事件标记线
- [ ] 支持 mock 数据驱动

关键事件：
- [ ] Analysis Start
- [ ] Error Spike
- [ ] Success TPS = 0
- [ ] Connection Failure
- [ ] Failover Detected
- [ ] Recovery
- [ ] Consistency Check Start
- [ ] Consistency Check Done

交付标准：能清晰表达“切换事件 → 业务中断 → 恢复”的关系。

---

## T10. 节点状态面板
- [ ] VIP 展示
- [ ] Current Primary 展示
- [ ] Current Secondary 展示
- [ ] 节点状态展示
- [ ] Last Role Switch Time 展示

交付标准：可直观看到主从角色与切换状态。

---

## T11. 实时事件流
- [ ] 事件流组件
- [ ] 支持级别区分（info / warn / error）
- [ ] 支持时间戳
- [ ] 支持切换过程事件展示

建议首批事件：
- [ ] Start analysis
- [ ] Error rate spike detected
- [ ] Success TPS dropped to zero
- [ ] Connection failure detected
- [ ] VIP switch detected
- [ ] Primary role changed
- [ ] Recovery detected
- [ ] Consistency check passed / failed

交付标准：以事件流形式替代原始日志刷屏。

---

## T12. Mock 数据与状态切换
- [ ] 构建默认演示数据
- [ ] 构建切换中状态数据
- [ ] 构建恢复后状态数据
- [ ] 页面联动验证

交付标准：无需真实后端也可完成前端演示。

---

## T13. 数据模型预留
- [ ] 定义 Analysis Session 前端模型
- [ ] 定义 Runtime Event 模型
- [ ] 定义 Summary Card 模型
- [ ] 定义 Trend Point 模型
- [ ] 定义 Cluster Status 模型

交付标准：后续真实数据接入时无需重构页面。

---

## T14. Start Analysis 协议预留
- [ ] 定义 start analysis 请求结构
- [ ] 定义 stop analysis 请求结构
- [ ] 定义实时数据刷新协议草案

交付标准：接口契约清晰，可供后端后续对接。

---

## T15. 兼容性回归检查
- [ ] 非 MySQL connections 行为检查
- [ ] Performance Analysis 不受影响
- [ ] History 不受影响
- [ ] 导航与样式未破坏

交付标准：Impact Analysis 以最小侵入方式集成。

---

## T16. 后续真实逻辑预留任务（后续阶段）
- [ ] Long Connection 实现
- [ ] Short Connection 实现
- [ ] Insert workload 实现
- [ ] Select workload 实现
- [ ] commit 成功采集
- [ ] consistency 校验实现
- [ ] failover 事件识别实现

说明：该组任务暂不作为第一期页面优先开发阻塞项。

---

## 3. 第一优先级任务

建议本轮最先执行：

1. T1 文档落地
2. T2 导航调整
3. T3 独立模块目录初始化
4. T4 页面状态壳
5. T5 / T6 MySQL Connections 改造设计与落地
6. T7 ~ T12 页面主功能

---

## 4. 验收清单

### 必过项
- [ ] 导航新增 `Impact Analysis`
- [ ] `Tasks & Monitor` 已改名为 `Performance Analysis`
- [ ] MySQL 连接支持 cluster connection
- [ ] 非 MySQL 功能无回归
- [ ] 页面支持 Start Analysis
- [ ] 页面支持核心指标卡片、主图、节点状态、事件流
- [ ] 一致性结果支持总结果展示

### 非阻塞项
- [ ] 后端真实逻辑完整落地
- [ ] 长短连接真实打通
- [ ] 一致性明细下钻

