# Sysbench Comparison Report

- **Generated at:** 2026-02-02 06:47:35
- **Report ID:** report-20260202_064735
- **Group By:** threads
- **Selected Records:** 5
- **Notes:** Simplified report (no Template Variant, no time series)

---

## 0) Record Selection Summary

### 0.1 Filters (UI Inputs)
- Search Query: (none)
- Selected Templates: All templates
- Threads in Selection: 1, 2, 4, 8, 16
## 1) Parsing & Sanity Checks (Global)

### 1.1 Global Parse Summary
| Item | Value |
|------|-------|
| Runs parsed successfully | 5 |
| Runs failed to parse | 0 |
| Unknown/Unparsed lines (count) | 0 |

### 1.2 Sanity Checks (Must Pass)
| Check | Result | Details |
|------|--------|----------|
| SQL total = read + write + other | ✅ PASS |  |
| QPS ≈ TPS × 20 | ✅ PASS |  |
| Latency min ≤ avg ≤ p95 | ✅ PASS |  |
| errors=0 & reconnects=0 | ✅ PASS |  |

## 2) Executive Summary (Across All Included Runs)

### 2.1 Best Points (by common criteria)
- Highest TPS (run-summary mean): **threads=16** → TPS=751.37
- Lowest p95 latency (run-summary mean): **threads=8** → p95=16.41

## 4) Comparison Sections (Per Thread Count)

## 4.1 Thread Group: threads=1

### 4.1.1 Experiment Matrix

| threads | N | date_span | tags |
|-------:|-:|----------|------|
| 1 | 1 | 2026-02-02 01:14 to 2026-02-02 01:14 | - |

### 4.1.2 Main Comparison (Run Summary Metrics)

> Source: sysbench tail summary (SQL statistics / General statistics / Latency)

|                          threads |  N | TPS mean±sd | TPS min..max | QPS mean±sd | Lat avg mean±sd | Lat p95 mean±sd | Errors | Reconnects |
| -------------------------------: | -: | ----------: | -----------: | ----------: | --------------: | --------------: | -----: | ---------: |
|                              - | 1 | 63.93 | 63.93 | 1286.67 | 15.57 | 33.72 | 0 | 0 |

### 4.1.7 Visuals (ASCII)

#### TPS vs Threads (run summary mean)
```text
threads=1  |████                                               63.93
threads=2  |███████                                            116.84
threads=4  |█████████████████                                  255.63
threads=8  |████████████████████████████████████████           608.50
threads=16 |██████████████████████████████████████████████████ 751.37
```

#### p95 Latency vs Threads (run summary mean)
```text
threads=1  |██████████████████████████████████████████████████ 33.72ms
threads=2  |████████████████████████████████████████           27.17ms
threads=4  |███████████████████████████████                    21.50ms
threads=8  |████████████████████████                           16.41ms
threads=16 |███████████████████████████████████████████████    31.94ms
```

### 4.1.8 Findings & Recommendation

* Best throughput threads: 16 (TPS=751.37)
* Best latency threads: 8 (p95=16.41ms)
* Recommendation: threads=16 (TPS=751.37, p95=31.94ms)

## 4.2 Thread Group: threads=2

### 4.2.1 Experiment Matrix

| threads | N | date_span | tags |
|-------:|-:|----------|------|
| 2 | 1 | 2026-02-02 01:14 to 2026-02-02 01:14 | - |

### 4.2.2 Main Comparison (Run Summary Metrics)

> Source: sysbench tail summary (SQL statistics / General statistics / Latency)

|                          threads |  N | TPS mean±sd | TPS min..max | QPS mean±sd | Lat avg mean±sd | Lat p95 mean±sd | Errors | Reconnects |
| -------------------------------: | -: | ----------: | -----------: | ----------: | --------------: | --------------: | -----: | ---------: |
|                              - | 1 | 116.84 | 116.84 | 2360.00 | 17.04 | 27.17 | 0 | 0 |

### 4.%d.5 Scaling & Efficiency (Threads Analysis)

|                  threads | TPS_mean | Speedup | Efficiency | ΔTPS vs prev | Δp95 vs prev |
| -----------------------: | -------: | ------: | ---------: | -----------: | -----------: |
|                 threads=1 |    63.93 |    0.00x |     0.00% |           — |            — |
|                 threads=2 |   116.84 |    1.83x |    91.38% |       52.91 |        -6.55 |

### 4.2.7 Visuals (ASCII)

#### TPS vs Threads (run summary mean)
```text
threads=1  |████                                               63.93
threads=2  |███████                                            116.84
threads=4  |█████████████████                                  255.63
threads=8  |████████████████████████████████████████           608.50
threads=16 |██████████████████████████████████████████████████ 751.37
```

#### p95 Latency vs Threads (run summary mean)
```text
threads=1  |██████████████████████████████████████████████████ 33.72ms
threads=2  |████████████████████████████████████████           27.17ms
threads=4  |███████████████████████████████                    21.50ms
threads=8  |████████████████████████                           16.41ms
threads=16 |███████████████████████████████████████████████    31.94ms
```

### 4.2.8 Findings & Recommendation

* Best throughput threads: 16 (TPS=751.37)
* Best latency threads: 8 (p95=16.41ms)
* Recommendation: threads=16 (TPS=751.37, p95=31.94ms)

## 4.3 Thread Group: threads=4

### 4.3.1 Experiment Matrix

| threads | N | date_span | tags |
|-------:|-:|----------|------|
| 4 | 1 | 2026-02-02 01:14 to 2026-02-02 01:14 | - |

### 4.3.2 Main Comparison (Run Summary Metrics)

> Source: sysbench tail summary (SQL statistics / General statistics / Latency)

|                          threads |  N | TPS mean±sd | TPS min..max | QPS mean±sd | Lat avg mean±sd | Lat p95 mean±sd | Errors | Reconnects |
| -------------------------------: | -: | ----------: | -----------: | ----------: | --------------: | --------------: | -----: | ---------: |
|                              - | 1 | 255.63 | 255.63 | 5173.33 | 15.52 | 21.50 | 0 | 0 |

### 4.%d.5 Scaling & Efficiency (Threads Analysis)

|                  threads | TPS_mean | Speedup | Efficiency | ΔTPS vs prev | Δp95 vs prev |
| -----------------------: | -------: | ------: | ---------: | -----------: | -----------: |
|                 threads=1 |    63.93 |    0.00x |     0.00% |           — |            — |
|                 threads=2 |   116.84 |    1.83x |    91.38% |       52.91 |        -6.55 |
|                 threads=4 |   255.63 |    4.00x |    99.96% |      138.79 |        -5.67 |

### 4.3.7 Visuals (ASCII)

#### TPS vs Threads (run summary mean)
```text
threads=1  |████                                               63.93
threads=2  |███████                                            116.84
threads=4  |█████████████████                                  255.63
threads=8  |████████████████████████████████████████           608.50
threads=16 |██████████████████████████████████████████████████ 751.37
```

#### p95 Latency vs Threads (run summary mean)
```text
threads=1  |██████████████████████████████████████████████████ 33.72ms
threads=2  |████████████████████████████████████████           27.17ms
threads=4  |███████████████████████████████                    21.50ms
threads=8  |████████████████████████                           16.41ms
threads=16 |███████████████████████████████████████████████    31.94ms
```

### 4.3.8 Findings & Recommendation

* Best throughput threads: 16 (TPS=751.37)
* Best latency threads: 8 (p95=16.41ms)
* Recommendation: threads=16 (TPS=751.37, p95=31.94ms)

## 4.4 Thread Group: threads=8

### 4.4.1 Experiment Matrix

| threads | N | date_span | tags |
|-------:|-:|----------|------|
| 8 | 1 | 2026-02-02 01:34 to 2026-02-02 01:34 | best-latency |

### 4.4.2 Main Comparison (Run Summary Metrics)

> Source: sysbench tail summary (SQL statistics / General statistics / Latency)

|                          threads |  N | TPS mean±sd | TPS min..max | QPS mean±sd | Lat avg mean±sd | Lat p95 mean±sd | Errors | Reconnects |
| -------------------------------: | -: | ----------: | -----------: | ----------: | --------------: | --------------: | -----: | ---------: |
|                              - | 1 | 608.50 | 608.50 | 12273.33 | 13.06 | 16.41 | 0 | 0 |

### 4.%d.5 Scaling & Efficiency (Threads Analysis)

|                  threads | TPS_mean | Speedup | Efficiency | ΔTPS vs prev | Δp95 vs prev |
| -----------------------: | -------: | ------: | ---------: | -----------: | -----------: |
|                 threads=1 |    63.93 |    0.00x |     0.00% |           — |            — |
|                 threads=2 |   116.84 |    1.83x |    91.38% |       52.91 |        -6.55 |
|                 threads=4 |   255.63 |    4.00x |    99.96% |      138.79 |        -5.67 |
|                 threads=8 |   608.50 |    9.52x |   118.98% |      352.87 |        -5.09 |

### 4.4.7 Visuals (ASCII)

#### TPS vs Threads (run summary mean)
```text
threads=1  |████                                               63.93
threads=2  |███████                                            116.84
threads=4  |█████████████████                                  255.63
threads=8  |████████████████████████████████████████           608.50
threads=16 |██████████████████████████████████████████████████ 751.37
```

#### p95 Latency vs Threads (run summary mean)
```text
threads=1  |██████████████████████████████████████████████████ 33.72ms
threads=2  |████████████████████████████████████████           27.17ms
threads=4  |███████████████████████████████                    21.50ms
threads=8  |████████████████████████                           16.41ms
threads=16 |███████████████████████████████████████████████    31.94ms
```

### 4.4.8 Findings & Recommendation

* Best throughput threads: 16 (TPS=751.37)
* Best latency threads: 8 (p95=16.41ms)
* Recommendation: threads=16 (TPS=751.37, p95=31.94ms)

## 4.5 Thread Group: threads=16

### 4.5.1 Experiment Matrix

| threads | N | date_span | tags |
|-------:|-:|----------|------|
| 16 | 1 | 2026-02-02 01:35 to 2026-02-02 01:35 | best-tps |

### 4.5.2 Main Comparison (Run Summary Metrics)

> Source: sysbench tail summary (SQL statistics / General statistics / Latency)

|                          threads |  N | TPS mean±sd | TPS min..max | QPS mean±sd | Lat avg mean±sd | Lat p95 mean±sd | Errors | Reconnects |
| -------------------------------: | -: | ----------: | -----------: | ----------: | --------------: | --------------: | -----: | ---------: |
|                              - | 1 | 751.37 | 751.37 | 15333.33 | 20.96 | 31.94 | 0 | 0 |

### 4.%d.5 Scaling & Efficiency (Threads Analysis)

|                  threads | TPS_mean | Speedup | Efficiency | ΔTPS vs prev | Δp95 vs prev |
| -----------------------: | -------: | ------: | ---------: | -----------: | -----------: |
|                 threads=1 |    63.93 |    0.00x |     0.00% |           — |            — |
|                 threads=2 |   116.84 |    1.83x |    91.38% |       52.91 |        -6.55 |
|                 threads=4 |   255.63 |    4.00x |    99.96% |      138.79 |        -5.67 |
|                 threads=8 |   608.50 |    9.52x |   118.98% |      352.87 |        -5.09 |
|                threads=16 |   751.37 |   11.75x |    73.46% |      142.87 |        15.53 |

### 4.5.7 Visuals (ASCII)

#### TPS vs Threads (run summary mean)
```text
threads=1  |████                                               63.93
threads=2  |███████                                            116.84
threads=4  |█████████████████                                  255.63
threads=8  |████████████████████████████████████████           608.50
threads=16 |██████████████████████████████████████████████████ 751.37
```

#### p95 Latency vs Threads (run summary mean)
```text
threads=1  |██████████████████████████████████████████████████ 33.72ms
threads=2  |████████████████████████████████████████           27.17ms
threads=4  |███████████████████████████████                    21.50ms
threads=8  |████████████████████████                           16.41ms
threads=16 |███████████████████████████████████████████████    31.94ms
```

### 4.5.8 Findings & Recommendation

* Best throughput threads: 16 (TPS=751.37)
* Best latency threads: 8 (p95=16.41ms)
* Recommendation: threads=16 (TPS=751.37, p95=31.94ms)

