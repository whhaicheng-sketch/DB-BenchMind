# Report 技术规格说明

## 1. 产物列表

每次压测建议产生以下产物：

1. `report.md`
   - 面向人类阅读
   - 聚焦摘要、结论、证据和建议

2. `report_bundle.json.zst`
   - 面向 AI 分析
   - 结构化、压缩、体积受限

3. `report_meta.json`（可选）
   - 记录生成时间、版本、schema 版本等

## 2. Human Report 规格

## 2.1 内容章节

Human Report 至少包含：

- Title / Run Meta
- Overall Status
- One-line Conclusion
- Core KPIs
- Resource Bottleneck Summary
- Bottleneck Judgment
- Comparison
- Trend Highlights
- Recommendation
- Detailed Data（折叠区摘要）

## 2.2 字段要求

### Run Meta
- run_id
- benchmark_type
- template_name
- target_name
- start_time
- end_time
- duration

### Core KPIs
- tps
- tpm
- latency_avg
- latency_p95
- latency_p99
- error_rate
- cpu_peak
- disk_util_peak

### Bottleneck Judgment
- primary_bottleneck
- confidence
- evidence[]

### Comparison
- vs_previous
- vs_baseline

## 3. AI Bundle 规格

## 3.1 文件体积限制

- 压缩后必须 `<= 1MB`

## 3.2 文件格式

推荐使用：

- JSON 作为逻辑结构
- Zstandard 作为压缩格式

文件名示例：

```text
report_bundle_20260402_143501.json.zst
```

## 3.3 Schema 示例

```json
{
  "schema_version": "1.0.0",
  "run_meta": {},
  "benchmark_summary": {},
  "resource_summary": {},
  "timeseries_downsampled": {},
  "phase_breakdown": {},
  "retained_windows": [],
  "anomaly_windows": [],
  "comparison": {},
  "raw_samples": [],
  "ai_meta": {}
}
```

## 4. 统一字段定义

## 4.1 run_meta

```json
{
  "run_id": "string",
  "benchmark_type": "sysbench|swingbench|hammerdb",
  "db_type": "mysql|oracle|postgresql|...",
  "template_name": "string",
  "target_name": "string",
  "start_time": "RFC3339",
  "end_time": "RFC3339",
  "duration_seconds": 0
}
```

## 4.2 benchmark_summary

```json
{
  "throughput": {
    "tps": 0,
    "tpm": 0,
    "qps": 0
  },
  "latency": {
    "avg_ms": 0,
    "p50_ms": 0,
    "p95_ms": 0,
    "p99_ms": 0,
    "max_ms": 0
  },
  "errors": {
    "total": 0,
    "error_rate": 0,
    "timeouts": 0
  },
  "tool_specific": {}
}
```

## 4.3 resource_summary

```json
{
  "cpu": {"avg": 0, "p95": 0, "peak": 0},
  "memory": {"avg": 0, "peak": 0, "used_percent": 0},
  "disk": {
    "read_iops_peak": 0,
    "write_iops_peak": 0,
    "util_peak": 0,
    "await_peak_ms": 0
  },
  "network": {
    "rx_peak_mb_s": 0,
    "tx_peak_mb_s": 0,
    "retransmits": 0,
    "errors": 0
  }
}
```

## 4.4 timeseries_downsampled

对以下指标允许统一降采样：

- TPS
- TPM
- latency_p95
- latency_p99
- cpu_usage
- memory_used
- disk_util
- disk_await
- network_rx
- network_tx

每个时间序列推荐为固定点数，字段示例：

```json
{
  "metric": "tps",
  "bucket_seconds": 30,
  "points": [
    {"ts": "2026-04-02T14:35:00Z", "avg": 320.1, "min": 315.0, "max": 326.2}
  ]
}
```

## 5. 保留窗口规格

## 5.1 普通窗口

建议默认比例：

- Front Window：0% ~ 15%
- Middle Window：45% ~ 55%
- Tail Window：85% ~ 100%

## 5.2 异常窗口触发条件

可配置阈值，示例：

- P95 / P99 超阈值
- Error Rate 突增
- TPS 突降
- CPU 或 Disk Util 长时间高位
- IO Await 明显抬升

异常窗口字段建议：

```json
{
  "type": "latency_spike",
  "start_time": "RFC3339",
  "end_time": "RFC3339",
  "evidence": {
    "p99_ms": 82.1,
    "disk_util": 91.2,
    "cpu": 63.5
  }
}
```

## 6. 原始样本保留规格

不保留全量 stdout/stderr 到 AI 包，只保留高价值样本。

建议每类样本保留：

- front：若干条
- middle：若干条
- tail：若干条
- anomaly：若干条

字段示例：

```json
{
  "position": "middle",
  "ts": "2026-04-02T15:48:00Z",
  "source": "sysbench",
  "content": "[5480s] tps: 327.8 qps: 6551.0 latency(ms,95%): 51.2"
}
```

## 7. 体积控制策略

## 7.1 体积预算建议

- meta + summary: 50KB 内
- resource summary: 50KB 内
- downsampled time series: 400KB 内
- retained/anomaly windows: 250KB 内
- raw samples: 150KB 内
- 预留与压缩波动空间: 100KB 内

## 7.2 超限时的处理顺序

1. 降低普通时间序列密度
2. 缩减普通窗口样本数量
3. 缩减非关键扩展字段
4. 保留异常窗口与核心 summary 不动

## 8. Spec 验收

- 同一套 schema 可承载 sysbench / Swingbench / HammerDB 数据。
- Human Report 不依赖 AI Bundle 才能阅读。
- AI Bundle 展开预览可清楚展示采样和保留策略。
- 长时压测场景下压缩后文件仍满足体积约束。
