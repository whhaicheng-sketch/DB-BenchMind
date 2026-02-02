# Sysbench Multi-Configuration Comparison Report

* **Generated at:** 2026-02-02 06:47:50
* **Report ID:** report-20260202_064750
* **Group by:** threads
* **Config Groups:** 5

---

## 1) Experiment Metadata

### 1.1 Basic Information

| Item | Value |
|------|-------|
| Report ID | report-20260202_064750 |
| Generated | 2026-02-02 06:47:50 |
| Group By | threads |
| Config Groups | 5 |

### 1.3 Measurement Policy

* **Report interval:** 1s
* **Test duration:** Varies by run
* **Runs per config (N):** Varies by config
* **Execution order:** Based on start time
* **Acceptance criteria:** errors=0 && reconnects=0

## 2) Experiment Matrix

| Config ID | threads | Database | Template | Runs (N) | Tags |
|---------:|-------:|---------|----------|--------:|------|
| C1 | 1 | mysql | Sysbench OLTP Read-Write | 1 | baseline |
| C2 | 2 | mysql | Sysbench OLTP Read-Write | 1 |  |
| C3 | 4 | mysql | Sysbench OLTP Read-Write | 1 |  |
| C4 | 8 | mysql | Sysbench OLTP Read-Write | 1 | best-latency |
| C5 | 16 | mysql | Sysbench OLTP Read-Write | 1 | best-tps |

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

| threads | Read % | Write % | Other % | Queries / Transaction |
|-------:|------:|-------:|-------:|--------------------:|
| 1 | 70.0 | 20.0 | 10.0 | 60.38 |
| 2 | 70.0 | 20.0 | 10.0 | 60.60 |
| 4 | 70.0 | 20.0 | 10.0 | 60.71 |
| 8 | 70.0 | 20.0 | 10.0 | 60.51 |
| 16 | 70.0 | 20.0 | 10.0 | 61.22 |

## 5) Scaling & Efficiency (Threads Analysis)

**Baseline:** threads=1 (TPS=63.93)

| threads | TPS_mean | Speedup | Efficiency (Speedup / threads) | ΔTPS vs prev | Δp95 latency |
|-------:|--------:|-------:|-------------------------------:|------------:|-------------:|
| 1 | 63.93 | 1.00x | 100.00% | — | — |
| 2 | 116.84 | 1.83x | 91.38% | 52.91 | -6.55 |
| 4 | 255.63 | 4.00x | 99.96% | 138.79 | -5.67 |
| 8 | 608.50 | 9.52x | 118.98% | 352.87 | -5.09 |
| 16 | 751.37 | 11.75x | 73.46% | 142.87 | 15.53 |

## 6) Visuals (ASCII Charts)

### 6.1 TPS vs Threads
```text
threads=1  |████                                               63.93
threads=2  |███████                                            116.84
threads=4  |█████████████████                                  255.63
threads=8  |████████████████████████████████████████           608.50
threads=16  |██████████████████████████████████████████████████ 751.37
```

### 6.2 p95 Latency vs Threads
```text
threads=1  |██████████████████████████████████████████████████ 33.72ms
threads=2  |████████████████████████████████████████           27.17ms
threads=4  |███████████████████████████████                    21.50ms
threads=8  |████████████████████████                           16.41ms
threads=16  |███████████████████████████████████████████████    31.94ms
```

## 7) Sanity Checks

✅ **ALL CHECKS PASSED**

| Check | Result | Details |
|------|--------|----------|
| SQL total = read + write + other | ✅ PASS |  |
| QPS ≈ TPS × 20 | ✅ PASS |  |
| Latency min ≤ avg ≤ p95 | ✅ PASS |  |
| errors=0 & reconnects=0 | ✅ PASS |  |

## 8) Findings & Recommendations

### 8.1 Key Findings

* **Best throughput point:** threads=16 (TPS=751.37, p95=31.94ms)
* **Best latency point:** threads=8 (p95=16.41ms)
* **Stability:** All configs stable (CV < 10%)

### 8.2 Recommendation

**Suggested:** threads=16

**Trade-off:** 11.75x speedup with 73.46% scaling efficiency at 31.94ms p95 latency

**Next experiment:** Repeat with N=5 runs per config for better statistics
