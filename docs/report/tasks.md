# Report 任务拆解

## T1. 文档与边界确认

### T1.1 建立文档目录
- 创建 `docs/report`
- 建立 README / requirements / design / spec / plan / tasks

### T1.2 明确需求边界
- 确认人类展示层与 AI 分析层分离
- 确认 Markdown 为人类主报告格式
- 确认 AI Bundle 压缩后 <= 1MB

### T1.3 明确验收标准
- 30 秒可判断
- Detailed Data 默认收起
- 长时压测可控裁剪

---

## T2. 数据模型设计

### T2.1 定义统一 Report Schema
- run_meta
- benchmark_summary
- resource_summary
- comparison
- phases
- samples

### T2.2 定义 benchmark 归一化映射
- sysbench → 统一字段
- Swingbench → 统一字段
- HammerDB → 统一字段

### T2.3 定义 tool_specific 扩展字段
- 避免统一字段被 benchmark 差异污染

---

## T3. AI Bundle 设计

### T3.1 定义 AI Bundle Schema
- schema_version
- summary
- timeseries_downsampled
- retained_windows
- anomaly_windows
- raw_samples
- ai_meta

### T3.2 设计 1MB 控制策略
- 时间序列降采样
- 普通窗口裁剪
- 异常窗口优先保留
- 非关键字段淘汰顺序

### T3.3 定义长时压测保留策略
- front
- middle
- tail
- anomaly

### T3.4 定义原始样本保留策略
- 保留高价值文本样本
- 不保留全量长日志

---

## T4. Human Report 模板设计

### T4.1 设计 Markdown 模板骨架
- Title / Meta
- Overall Status
- One-line Conclusion
- Core KPIs
- Resource Summary
- Bottleneck Judgment
- Comparison
- Trend Highlights
- Recommendation
- Detailed Data

### T4.2 设计字段映射表
- 模板占位符
- 来源字段
- 缺省值策略

### T4.3 设计状态与结论规则
- Healthy / Warning / Critical
- 一句话结论拼装规则

---

## T5. Detailed Data 展示设计

### T5.1 设计折叠区内容
- 包大小
- 采样策略
- 窗口摘要
- 异常摘要
- 样本预览

### T5.2 设计折叠区交互
- 默认收起
- 支持查看/隐藏
- 不影响主报告阅读节奏

---

## T6. 对比与分析能力设计

### T6.1 定义 previous / baseline 对比维度
- TPS / TPM
- P95 / P99
- Error Rate
- CPU Peak
- Disk Util Peak

### T6.2 定义规则引擎初判
- CPU-bound
- IO-bound
- Lock / Contention
- Network / Connection
- Misconfiguration

### T6.3 定义建议生成逻辑
- 每种瓶颈对应建议动作
- 建议按优先级输出

---

## T7. 后续实现任务建议顺序

1. 完成文档评审
2. 定稿统一 schema
3. 实现 AI Bundle 生成器
4. 实现 Markdown Report 生成器
5. 实现页面 Detailed Data 折叠区
6. 接入 previous / baseline 对比
7. 接入规则引擎结论
8. 预留 AI 深度分析入口

---

## 建议任务优先级

### P0
- 文档定稿
- Schema 定稿
- 1MB 控制策略定稿
- Markdown 模板定稿

### P1
- Bundle 生成能力
- Detailed Data 折叠展示
- 对比能力

### P2
- 规则引擎结论
- AI 深度分析接入
- 后续趋势分析

---

## T8. 实现完成状态

### 已完成 (2026-04-02)

| 任务 | 状态 | 关键文件 |
|------|------|----------|
| T2.1 统一 Report Schema | ✅ | `internal/domain/report/bundle.go`, `bottleneck.go` |
| T3.1 AI Bundle Schema | ✅ | `internal/domain/report/bundle.go` |
| T3.2 1MB 控制策略 | ✅ | `internal/app/usecase/report_bundle.go` (4层裁剪) |
| T3.3 前/中/后+异常窗口 | ✅ | `internal/app/usecase/report_bundle.go` |
| T4.1 Markdown 模板 | ✅ | `internal/app/usecase/report_markdown.go` |
| T5 Detailed Data 折叠区 | ✅ | `frontend/src/components/report/ReportDetailPanel.vue` |
| T6.1 previous/baseline 对比 | ✅ | `internal/app/usecase/report_usecase.go` |
| T6.2 规则引擎初判 | ✅ | `internal/app/usecase/report_bottleneck.go` |
| T6.3 建议生成逻辑 | ✅ | `internal/app/usecase/report_bottleneck.go` |
| 编排器 | ✅ | `internal/app/usecase/report_assembler.go` |
| Wails 绑定 | ✅ | `internal/transportwails/bindings/report.go` (GetDetailedData) |
| 测试覆盖 | ✅ | 24 test cases 全部通过 |

### 待完成

| 任务 | 说明 |
|------|------|
| 接入 collector 管道 | ReportAssembler 集成到 DefaultReportCollector.CollectAndPersist |
| AI 深度分析 | 预留入口，后续接入外部 AI 服务 |
| 套件级聚合报告 | 多报告对比与趋势分析 |
