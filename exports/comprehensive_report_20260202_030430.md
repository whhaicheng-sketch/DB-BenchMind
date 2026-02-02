# Sysbench Multi-Configuration Comparison Report

* **Generated at:** 2026-02-02 03:04:30
* **Group by:** threads
* **Report ID:** report-20260202-030430

---

## 1) Experiment Metadata

### 1.1 Basic Information

| Item | Value |
|------|-------|
| Report ID | report-20260202-030430 |
| Generated | 2026-02-02 03:04:30 |
| Group By | threads |
| Config Groups | 5 |
| Time Window | 5m0s |

### 1.3 Measurement Policy

* **Report interval:** 1s
* **Test duration:** Varies by run
* **Runs per config (N):** Varies by config
* **Execution order:** Based on start time
* **Acceptance criteria:** errors=0 && reconnects=0

## 2) Experiment Matrix

| Config ID | threads | Database | Template | Runs (N) | Tags |
|---------:|-------:|---------|----------|--------:|------|
| C1 | 1 | mysql | Sysbench OLTP Read-Write | 1 | worst-latency  baseline |
| C2 | 2 | mysql | Sysbench OLTP Read-Write | 1 |  |
| C3 | 4 | mysql | Sysbench OLTP Read-Write | 1 |  |
| C4 | 8 | mysql | Sysbench OLTP Read-Write | 1 |  |
| C5 | 16 | mysql | Sysbench OLTP Read-Write | 1 | best-tps  |

> **Definitions:**
> * **Run Summary Metrics** = sysbench summary statistics (SQL statistics, Latency, etc.)
> * **Config ID** = Unique identifier for each configuration group (C1, C2, C3...)
> * **Runs (N)** = Number of benchmark executions for this configuration

## 3) Main Comparison (Run Summary Metrics)

> **Note:** If N=1, StdDev = N/A; Min=Avg=Max=Single value
> Latency unit: milliseconds

### 3.1 Throughput & Latency Summary

| threads | N | TPS (mean ± sd) | TPS (min..max) | QPS (mean ± sd) | QPS (min..max) | Lat avg ms (mean ± sd) | Lat p95 ms (mean ± sd) | Lat max ms (max-of-max) |
|-------:|:-:|---------------:|--------------:|---------------:|--------------:|----------------------:|----------------------:|-----------------------:|
| 1 | 1 | 63.93 | 63.93 | 1286.67 | 1286.67 | 15.57 | 33.72 | 48.39 |
| 2 | 1 | 116.84 | 116.84 | 2360.00 | 2360.00 | 17.04 | 27.17 | 32.57 |
| 4 | 1 | 255.63 | 255.63 | 5173.33 | 5173.33 | 15.52 | 21.50 | 30.97 |
| 8 | 1 | 608.50 | 608.50 | 12273.33 | 12273.33 | 13.06 | 16.41 | 49.96 |
| 16 | 1 | 751.37 | 751.37 | 15333.33 | 15333.33 | 20.96 | 31.94 | 104.36 |

### 3.2 Reliability

| threads | N | Total Errors | Total Reconnects | Any non-zero? |
|-------:|:-:|------------:|---------------:|:-------------|
| 1 | 1 | 0 | 0 | NO |
| 2 | 1 | 0 | 0 | NO |
| 4 | 1 | 0 | 0 | NO |
| 8 | 1 | 0 | 0 | NO |
| 16 | 1 | 0 | 0 | NO |

### 3.3 Actual Query Mix (from SQL statistics)

| threads | Read %% | Write %% | Other %% | Queries / Transaction |
|-------:|------:|-------:|-------:|--------------------:|
| 1 | 70.0 | 20.0 | 10.0 | 20.00 |
| 2 | 70.0 | 20.0 | 10.0 | 20.00 |
| 4 | 70.0 | 20.0 | 10.0 | 20.00 |
| 8 | 70.0 | 20.0 | 10.0 | 20.00 |
| 16 | 70.0 | 20.0 | 10.0 | 20.00 |

## 5) Scaling & Efficiency (Threads Analysis)

**Baseline:** threads=1 (TPS=63.93)

| threads | TPS_mean | Speedup | Efficiency (Speedup / threads) | ΔTPS vs prev | Δp95 latency |
|-------:|--------:|-------:|-------------------------------:|------------:|-------------:|
| 1 | 63.93 | 1.00 | 1.00 | — | — |
| 2 | 116.84 | 1.83 | 0.91 | 52.91 | -6.55 |
| 4 | 255.63 | 4.00 | 1.00 | 138.79 | -5.67 |
| 8 | 608.50 | 9.52 | 1.19 | 352.87 | -5.09 |
| 16 | 751.37 | 11.75 | 0.73 | 142.87 | 15.53 |

**Analysis:**
- **Best throughput:** threads=16 (TPS=751.37)
- **Scaling knee:** threads=~16 (efficiency drops significantly)

## 6) Visuals (ASCII Charts)

### 6.1 TPS vs Threads
```
threads=1  |████                                               63.93
threads=2  |███████                                            116.84
threads=4  |█████████████████                                  255.63
threads=8  |████████████████████████████████████████           608.50
threads=16 |██████████████████████████████████████████████████ 751.37
```

### 6.2 p95 Latency vs Threads
```
threads=1  |██████████████████████████████████████████████████ 33.72ms
threads=2  |████████████████████████████████████████           27.17ms
threads=4  |███████████████████████████████                    21.50ms
threads=8  |████████████████████████                           16.41ms
threads=16 |███████████████████████████████████████████████    31.94ms
```


## Sanity Checks

⚠️  **SOME CHECKS FAILED**

│ Check │ Result │ Details │
├───────┼────────┼─────────┤
│ Config groups exist                                │ ✅ PASS   │ Found 5 config groups                              │
│ Latency min ≤ avg (Group C1)                       │ ✅ PASS   │ min=15.57, avg=15.57                               │
│ Latency avg ≤ p95 (Group C1)                       │ ✅ PASS   │ avg=15.57, p95=33.72                               │
│ Latency p95 ≤ p99 (Group C1)                       │ ❌ FAIL   │ p95=33.72, p99=0.00                                │
│ Latency p99 ≤ max (Group C1)                       │ ✅ PASS   │ p99=0.00, max=48.39                                │
│ QPS ≈ TPS × queries/tx (Group C1)                  │ ✅ PASS   │ Expected=1278.60, Actual=1286.67, Diff=0.63%       │
│ SQL total = read + write + other (Group C1)        │ ✅ PASS   │ Read=2702, Write=772, Other=386, Calculated=386... │
│ No errors/reconnects (Group C1)                    │ ✅ PASS   │ Errors=0, Reconnects=0                             │
│ N=1 stddev check (Group C1)                        │ ✅ PASS   │ StdDev=0.00 (expected 0 for N=1)                   │
│ Latency min ≤ avg (Group C2)                       │ ✅ PASS   │ min=17.04, avg=17.04                               │
│ Latency avg ≤ p95 (Group C2)                       │ ✅ PASS   │ avg=17.04, p95=27.17                               │
│ Latency p95 ≤ p99 (Group C2)                       │ ❌ FAIL   │ p95=27.17, p99=0.00                                │
│ Latency p99 ≤ max (Group C2)                       │ ✅ PASS   │ p99=0.00, max=32.57                                │
│ QPS ≈ TPS × queries/tx (Group C2)                  │ ✅ PASS   │ Expected=2336.80, Actual=2360.00, Diff=0.99%       │
│ SQL total = read + write + other (Group C2)        │ ✅ PASS   │ Read=4956, Write=1416, Other=708, Calculated=70... │
│ No errors/reconnects (Group C2)                    │ ✅ PASS   │ Errors=0, Reconnects=0                             │
│ N=1 stddev check (Group C2)                        │ ✅ PASS   │ StdDev=0.00 (expected 0 for N=1)                   │
│ Latency min ≤ avg (Group C3)                       │ ✅ PASS   │ min=15.52, avg=15.52                               │
│ Latency avg ≤ p95 (Group C3)                       │ ✅ PASS   │ avg=15.52, p95=21.50                               │
│ Latency p95 ≤ p99 (Group C3)                       │ ❌ FAIL   │ p95=21.50, p99=0.00                                │
│ Latency p99 ≤ max (Group C3)                       │ ✅ PASS   │ p99=0.00, max=30.97                                │
│ QPS ≈ TPS × queries/tx (Group C3)                  │ ✅ PASS   │ Expected=5112.60, Actual=5173.33, Diff=1.19%       │
│ SQL total = read + write + other (Group C3)        │ ✅ PASS   │ Read=10864, Write=3104, Other=1552, Calculated=... │
│ No errors/reconnects (Group C3)                    │ ✅ PASS   │ Errors=0, Reconnects=0                             │
│ N=1 stddev check (Group C3)                        │ ✅ PASS   │ StdDev=0.00 (expected 0 for N=1)                   │
│ Latency min ≤ avg (Group C4)                       │ ✅ PASS   │ min=13.06, avg=13.06                               │
│ Latency avg ≤ p95 (Group C4)                       │ ✅ PASS   │ avg=13.06, p95=16.41                               │
│ Latency p95 ≤ p99 (Group C4)                       │ ❌ FAIL   │ p95=16.41, p99=0.00                                │
│ Latency p99 ≤ max (Group C4)                       │ ✅ PASS   │ p99=0.00, max=49.96                                │
│ QPS ≈ TPS × queries/tx (Group C4)                  │ ✅ PASS   │ Expected=12170.00, Actual=12273.33, Diff=0.85%     │
│ SQL total = read + write + other (Group C4)        │ ✅ PASS   │ Read=25774, Write=7364, Other=3682, Calculated=... │
│ No errors/reconnects (Group C4)                    │ ✅ PASS   │ Errors=0, Reconnects=0                             │
│ N=1 stddev check (Group C4)                        │ ✅ PASS   │ StdDev=0.00 (expected 0 for N=1)                   │
│ Latency min ≤ avg (Group C5)                       │ ✅ PASS   │ min=20.96, avg=20.96                               │
│ Latency avg ≤ p95 (Group C5)                       │ ✅ PASS   │ avg=20.96, p95=31.94                               │
│ Latency p95 ≤ p99 (Group C5)                       │ ❌ FAIL   │ p95=31.94, p99=0.00                                │
│ Latency p99 ≤ max (Group C5)                       │ ✅ PASS   │ p99=0.00, max=104.36                               │
│ QPS ≈ TPS × queries/tx (Group C5)                  │ ✅ PASS   │ Expected=15027.40, Actual=15333.33, Diff=2.04%     │
│ SQL total = read + write + other (Group C5)        │ ✅ PASS   │ Read=32200, Write=9200, Other=4600, Calculated=... │
│ No errors/reconnects (Group C5)                    │ ✅ PASS   │ Errors=0, Reconnects=0                             │
│ N=1 stddev check (Group C5)                        │ ✅ PASS   │ StdDev=0.00 (expected 0 for N=1)                   │
│ Baseline exists                                    │ ✅ PASS   │ threads=1 group found: true                        │
└───────┴────────┴─────────┘
## 8) Findings & Recommendations

### 8.1 Key Findings

* **Best throughput point:** threads=16 (TPS=751.37, p95=31.94ms)
* **Scaling knee:** threads=~16 (efficiency drops significantly)
* **Latency risk point:** threads=1 (p95=33.72ms - highest latency)
* **Stability:** All configs stable (CV < 10%)

### 8.2 Recommendation

**Suggested:** threads=8

**Trade-off:** 9.52x speedup with 118.98% scaling efficiency at 16.41ms p95 latency

**Next experiment:** Repeat with N=5 runs per config for better statistics

