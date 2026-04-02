# Report 方案设计

## 1. 设计概述

本方案将 Report 拆成两层：

1. **人类展示层（Human Report）**
2. **AI 分析层（AI Bundle）**

两层共享同一份基础数据来源，但呈现目标不同。

- Human Report 关注快速判断。
- AI Bundle 关注证据完整性与体积可控性。

## 2. 信息架构

## 2.1 人类展示层

建议固定为以下区块：

1. Title / Run Meta
2. Overall Status
3. One-line Conclusion
4. Core KPIs
5. Resource Bottleneck Summary
6. Bottleneck Judgment
7. Comparison
8. Trend Highlights
9. Recommendation
10. Detailed Data（默认收起）

### 为什么这样设计

- 先结论后证据，符合使用习惯。
- 把大量原始数据压到详细区，避免主页面失焦。
- 对比和趋势位于中部，既能判断现状，也能看变化。

## 2.2 AI 分析层

AI Bundle 不直接作为主页面内容，而是作为折叠区摘要 + 文件产物存在。

包含以下部分：

1. Meta
2. Benchmark Summary
3. Resource Summary
4. Downsampled Time Series
5. Retained Windows
6. Anomaly Windows
7. Raw Samples
8. Comparison
9. Phase Breakdown
10. AI Analysis Metadata

## 3. 页面交互设计

## 3.1 默认状态

- 默认只显示人类展示层。
- 详细数据区域默认收起。
- 用户不展开也可以完成绝大多数判断。

## 3.2 详细数据入口

建议使用：

- 按钮：`查看详细数据`
- 或折叠面板：`Detailed Data`

展开后展示：

- AI 包名称
- 压缩后大小
- 采样策略
- 前/中/后窗口摘要
- 异常窗口摘要
- 原始样本片段
- 结构示意

## 4. 报告结构设计

## 4.1 顶部摘要区

应包括：

- Benchmark 类型
- 模板名
- 目标连接
- Run ID
- 起止时间
- 总时长
- 总体健康状态
- 一句话结论

## 4.2 核心指标区

建议固定：

- TPS
- TPM
- Avg Latency
- P95
- P99
- Error Rate
- CPU Peak
- Disk Util Peak

这里的目标不是展示尽可能多，而是展示最有判断价值的指标。

## 4.3 资源瓶颈摘要区

按资源维度给出平均值、峰值与结论：

- CPU
- Memory
- Disk
- Network

每个资源项必须有一个简短 verdict，而不是只给数字。

## 4.4 瓶颈判断区

至少包括：

- Primary Bottleneck
- Confidence
- Why（证据列表）

这样后续 AI 结论能自然映射到 UI，不需要额外改版。

## 4.5 详细数据区

详细数据区是 AI Bundle 的“可视化预览”，而不是原始日志堆叠区。

应该展示：

- 采样策略
- 时间序列是否降采样
- 保留了哪些窗口
- 异常窗口有哪些
- 样本保留策略是什么
- 当前包体积多少

## 5. 长时压测数据保留设计

## 5.1 问题

压测持续时间很长时，原始监控与 benchmark 输出会快速膨胀，无法直接进入 AI 输入。

## 5.2 解决方案

采用“普通窗口 + 异常窗口 + 降采样”的三层策略。

### 普通窗口
- Front Window：前段
- Middle Window：中段
- Tail Window：后段

### 异常窗口
自动捕获以下窗口：
- 延迟尖峰窗口
- 错误突增窗口
- TPS 突降窗口
- CPU / IO 高压窗口

### 降采样
对长时间序列执行固定点数降采样，保留：
- avg
- min
- max
- p95（如适用）

## 5.3 优先级原则

保留顺序优先级建议为：

1. 异常窗口
2. 前 / 中 / 后窗口
3. 关键原始样本
4. 普通高频序列细节

若包体积超限，应优先降低普通窗口密度，而非删除异常窗口。

## 6. 统一归一化设计

为兼容不同 benchmark，需要引入统一归一化层。

### 输入来源
- sysbench
- Swingbench
- HammerDB

### 统一输出目标
- throughput
- latency
- errors
- resources
- phases
- comparison
- samples

### 原则

- UI 不因 benchmark 切换而完全变形。
- 差异保留在“原始样本”和“扩展字段”中。

## 7. 为什么 Markdown 仍然适合作为人类报告

Markdown 的优势：

- 简洁清晰
- 易于导出与归档
- 易于生成
- 可在后续继续导出 PDF/HTML

因此方案建议：

- 面向人类：`report.md`
- 面向 AI：`report_bundle.json.zst`

## 8. 设计决策结论

### 决策 D1
主报告只服务人类快速判断，不承担原始数据承载职责。

### 决策 D2
AI 输入必须隐藏在详细数据区，不默认展示。

### 决策 D3
长时压测不保留全量原始数据到 AI 包，而是采用前/中/后 + 异常窗口 + 降采样。

### 决策 D4
Markdown 作为人类导出主格式，结构化压缩包作为 AI 分析主输入。

## 9. 实现备注

### 压缩格式
原方案使用 zstd 压缩 (`report_bundle.json.zst`)，实际实现使用 gzip (`report_bundle.json.gz`)，因 zstd 依赖在网络受限环境不可用。gzip 为 Go 标准库自带，满足 1MB 体积控制要求。

### 文件产物
- 人类报告：`report.md`（Markdown 格式）
- AI Bundle：`report_bundle.json.gz`（gzip 压缩 JSON）
- 存储路径：`{reportsDir}/{suiteID}/{reportID}/`

### 关键实现文件
- `internal/domain/report/bundle.go` — Bundle 领域模型
- `internal/domain/report/bottleneck.go` — 瓶颈判断领域模型
- `internal/app/usecase/report_bundle.go` — Bundle 生成器（含降采样、异常检测、压缩）
- `internal/app/usecase/report_markdown.go` — Markdown 生成器
- `internal/app/usecase/report_bottleneck.go` — 规则引擎（CPU/IO/锁/网络/配置）
- `internal/app/usecase/report_assembler.go` — 编排器（组装所有产物）
- `internal/app/usecase/report_usecase.go` — GetDetailedData、CompareReports
- `internal/transportwails/bindings/report.go` — GetDetailedData 绑定
- `frontend/src/components/report/ReportDetailPanel.vue` — 详细数据折叠区
