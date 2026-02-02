# Comparison Report Redesign Plan

## Current State Analysis

### What We Have Now:
- ✅ Basic comparison of 2-5 individual benchmark runs
- ✅ Simple statistics (Min/Avg/Max/StdDev) across selected records
- ✅ ASCII bar charts for TPS/QPS/Latency
- ❌ No grouping by configuration (e.g., all runs with threads=4)
- ❌ No support for N≥1 runs per configuration
- ❌ Missing: environment metadata, reliability checks, scaling analysis
- ❌ Limited report format (simple ASCII tables)

### What You Need:
Professional-grade multi-configuration comparison report with:
1. **Grouped Configurations** - e.g., all runs with threads=4 as "Config C4"
2. **Multiple Runs per Config** - Statistics across N runs (mean±stddev, min..max)
3. **Comprehensive Metadata** - Environment, sysbench command, measurement policy
4. **Advanced Analysis** - Scaling efficiency, steady-state, sanity checks
5. **Findings & Recommendations** - Best throughput point, scaling knee, suggestions

---

## Proposed Data Model Changes

### 1. New Domain Structure

```go
// ConfigGroup represents a group of runs with the same configuration
type ConfigGroup struct {
    GroupID        string    // e.g., "C1", "C2", "C3"
    Config         ConfigSpec // The configuration dimensions
    Runs           []*Run     // N runs with this configuration
    Statistics     RunStats   // Aggregated statistics across N runs
}

type ConfigSpec struct {
    Threads        int
    DatabaseType   string
    TemplateName   string
    ConnectionName string
    // Other config dimensions that define "same config"
}

type Run struct {
    RunID          string
    StartTime      time.Time
    // All existing metrics from history.Record
    TPS            float64
    QPS            float64
    LatencyAvg     float64
    LatencyP95     float64
    LatencyMax     float64
    Errors         int64
    Reconnects     int64
    // ... other metrics
}

type RunStats struct {
    // For each metric, calculate across N runs:
    N              int              // Number of runs

    // TPS statistics
    TPS_Mean       float64
    TPS_StdDev     float64
    TPS_Min        float64
    TPS_Max        float64

    // QPS statistics
    QPS_Mean       float64
    QPS_StdDev     float64
    QPS_Min        float64
    QPS_Max        float64

    // Latency statistics
    LatAvg_Mean    float64
    LatAvg_StdDev  float64
    LatP95_Mean    float64
    LatP95_StdDev  float64
    LatMax_Max     float64  // max-of-max

    // Reliability
    TotalErrors    int64
    TotalReconnects int64
    HasErrors      bool
}

// ComparisonReport is the new comprehensive report structure
type ComparisonReport struct {
    // Metadata
    GeneratedAt    time.Time
    GroupBy        GroupByField  // threads, database_type, template

    // Environment info (to be collected)
    Environment    EnvironmentInfo

    // Experiment matrix
    ConfigGroups   []*ConfigGroup

    // Analysis results
    ScalingAnalysis *ScalingAnalysis
    SanityChecks    *SanityCheckResults
    Findings        *ReportFindings
}

type EnvironmentInfo struct {
    ClientHost      string
    ServerHost      string
    CPU_Client      string
    CPU_Server      string
    Memory_Client   string
    Memory_Server   string
    Storage         string
    Network         string
    OS              string
    DBVersion       string
    DBParams        string
    Dataset         string
}

type ScalingAnalysis struct {
    BaselineTPS     float64  // TPS for threads=1
    Speedup         []float64 // Speedup for each config vs baseline
    Efficiency      []float64 // Speedup / threads
    ScalingKnee     int       // Thread count with diminishing returns
    BestTPSConfig   *ConfigGroup
    WorstLatencyConfig *ConfigGroup
}

type SanityCheckResults struct {
    AllPassed       bool
    Checks          []SanityCheck
}

type SanityCheck struct {
    Name            string
    Passed          bool
    Details         string
}

type ReportFindings struct {
    BestThroughput  string
    ScalingKnee     string
    LatencyRisk     string
    StabilityConcerns string
    Recommendation  string
    TradeoffStatement string
    NextExperiment  string
}
```

### 2. Report Format

Two formats:
1. **Markdown** - Rich, professional report (your template)
2. **TXT** - Plain text version (for CLI)

---

## Implementation Plan

### Phase 1: Data Model & Core Logic (1-2 days)

**Task 1.1**: Create new domain structures
- [ ] Add `ConfigGroup`, `ConfigSpec`, `RunStats` to `comparison.go`
- [ ] Create `comparison_report.go` with `ComparisonReport` structure
- [ ] Add environment info structures

**Task 1.2**: Implement grouping logic
- [ ] `GroupRecordsByConfig(records []*history.Record, groupBy GroupByField) []*ConfigGroup`
- [ ] Identify runs with same configuration
- [ ] Assign GroupID (C1, C2, C3...)

**Task 1.3**: Implement statistics calculation
- [ ] `CalculateRunStats(runs []*Run) RunStats`
- [ ] Calculate mean, stddev, min, max for each metric across N runs
- [ ] Handle N=1 case (stddev = N/A)

### Phase 2: Analysis Features (1-2 days)

**Task 2.1**: Scaling analysis
- [ ] Calculate speedup vs baseline (threads=1)
- [ ] Calculate efficiency (speedup / threads)
- [ ] Identify scaling knee (diminishing returns point)
- [ ] Find best TPS config, worst latency config

**Task 2.2**: Sanity checks
- [ ] Validate min ≤ avg ≤ max for latency
- [ ] Check QPS ≈ TPS × (queries/transaction)
- [ ] Verify SQL total = read + write + other
- [ ] Check errors=0 and reconnects=0

**Task 2.3**: Findings generation
- [ ] Auto-generate best throughput point
- [ ] Identify scaling knee
- [ ] Flag latency risk points
- [ ] Generate recommendations

### Phase 3: Report Generation (1 day)

**Task 3.1**: Markdown report generator
- [ ] Implement `FormatMarkdown() string` on `ComparisonReport`
- [ ] Include all sections from your template
- [ ] Generate ASCII charts for TPS vs Threads, Latency vs Threads

**Task 3.2**: Environment info collection
- [ ] Collect system info (CPU, memory, OS)
- [ ] Extract database version and parameters
- [ ] Store dataset information

**Task 3.3**: Update GUI and CLI
- [ ] GUI: Update export to use new report format
- [ ] CLI: Add `comparison report` command

### Phase 4: Testing & Documentation (0.5 day)

**Task 4.1**: Test with real data
- [ ] Test with N=1 (single run per config)
- [ ] Test with N>1 (multiple runs per config)
- [ ] Verify statistics calculations
- [ ] Validate report output

**Task 4.2**: Documentation
- [ ] Update user guide with new report format
- [ ] Document environment info collection
- [ ] Add examples

---

## Key Changes Required

### 1. Database Schema Enhancement

```sql
-- Add environment tracking
ALTER TABLE history_records ADD COLUMN sysbench_version TEXT;
ALTER TABLE history_records ADD COLUMN db_version TEXT;
ALTER TABLE history_records ADD COLUMN db_params TEXT;
ALTER TABLE history_records ADD COLUMN dataset_info TEXT;

-- Add raw command for reproducibility
ALTER TABLE history_records ADD COLUMN command_line TEXT;

-- Add execution metadata
ALTER TABLE history_records ADD COLUMN run_number INTEGER; -- Nth run of this config
ALTER TABLE history_records ADD COLUMN config_group_id TEXT; -- C1, C2, C3...
```

### 2. Template Execution Changes

When parsing sysbench output:
- Extract sysbench version
- Extract and store command line
- Parse and store DB version (if available)
- Store dataset info (tables, table size, total rows)
- Assign config_group_id based on config
- Track run_number for N runs

### 3. Use Case Layer Changes

```go
// New method in ComparisonUseCase
func (uc *ComparisonUseCase) GenerateComprehensiveReport(
    ctx context.Context,
    recordIDs []string,
    groupBy GroupByField,
) (*comparison.ComparisonReport, error) {
    // 1. Load records
    // 2. Group by configuration
    // 3. Calculate statistics for each group
    // 4. Perform scaling analysis
    // 5. Run sanity checks
    // 6. Generate findings
    // 7. Return comprehensive report
}
```

### 4. GUI Changes

Comparison page updates:
- Add environment info display section
- Show grouped configs (C1, C2, C3...) instead of individual runs
- Display N runs per config
- Show statistics in "mean ± stddev" format
- Add scaling efficiency table
- Display findings and recommendations

---

## Migration Strategy

### Option A: Incremental (Recommended)
1. Keep existing simple comparison for backward compatibility
2. Add new "Comprehensive Report" as separate feature
3. Allow user to choose: Simple vs Comprehensive
4. Gradually migrate to comprehensive as default

### Option B: Full Replacement
1. Replace entire comparison feature
2. All existing comparisons use new format
3. Risk: Breaking change for existing users

**Recommendation**: Option A with feature flag

---

## Example Output (Based on Your Template)

```markdown
# Sysbench Multi-Configuration Comparison Report

* **Generated at:** 2026-02-02 10:30:00
* **Tool:** sysbench 1.0.20 (LuaJIT 2.1.0-beta3)
* **Group by:** threads
* **Workload:** oltp_read_write
* **Run mode:** time-based
* **Notes:** Cold cache, randomized order

---

## 1) Experiment Metadata

### 1.1 Environment

| Item | Value |
|------|-------|
| Client Host | bench-client-01 |
| Server Host | mysql-server-prod |
| CPU / Memory (Client) | 8 cores / 32GB |
| CPU / Memory (Server) | 16 cores / 64GB |
| Storage | NVMe SSD |
| Network | 10Gbps |
| OS | Ubuntu 22.04 |
| MySQL Version | 8.0.35 |
| Key MySQL Params | innodb_buffer_pool_size=16GB |
| Dataset | tables=10, table_size=1000000 |

### 1.2 Sysbench Command

```bash
sysbench oltp_read_write \
  --threads=4 \
  --time=60 \
  --tables=10 \
  --table-size=1000000 \
  --db-driver=mysql \
  run
```

### 1.3 Measurement Policy

* **Report interval:** 1s
* **Test duration:** 60s
* **Warmup handling:** Discard first 10s
* **Runs per config (N):** 3
* **Execution order:** Randomized
* **Acceptance criteria:** errors=0 && reconnects=0

---

## 2) Experiment Matrix

| Config ID | threads | Runs (N) | Tags |
|---------:|-------:|--------:|------|
| C1 | 1 | 3 | baseline |
| C2 | 2 | 3 |  |
| C3 | 4 | 3 |  |
| C4 | 8 | 3 |  |
| C5 | 16 | 3 |  |

---

## 3) Main Comparison

### 3.1 Throughput & Latency Summary

| threads | N | TPS (mean ± sd) | TPS (min..max) | QPS (mean ± sd) | Lat avg ms (mean ± sd) | Lat p95 ms (mean ± sd) |
|-------:|:-:|---------------:|--------------:|---------------:|----------------------:|----------------------:|
| 1 | 3 | 75.2 ± 2.1 | 73.1 .. 77.5 | 1504 ± 42 | 13.3 ± 0.4 | 28.5 ± 1.2 |
| 2 | 3 | 142.8 ± 5.3 | 137.5 .. 148.2 | 2856 ± 106 | 14.0 ± 0.5 | 31.2 ± 1.8 |
| 4 | 3 | 255.6 ± 12.4 | 243.2 .. 268.1 | 5112 ± 248 | 15.7 ± 0.8 | 35.8 ± 2.1 |
| 8 | 3 | 423.5 ± 28.7 | 394.8 .. 452.1 | 8470 ± 574 | 19.2 ± 1.5 | 45.3 ± 3.8 |
| 16 | 3 | 512.3 ± 45.2 | 467.1 .. 558.9 | 10246 ± 904 | 31.8 ± 3.2 | 78.5 ± 8.4 |

### 3.2 Reliability

| threads | N | Total Errors | Total Reconnects | Any non-zero? |
|-------:|:-:|------------:|----------------:|:-------------|
| 1 | 3 | 0 | 0 | NO |
| 2 | 3 | 0 | 0 | NO |
| 4 | 3 | 0 | 0 | NO |
| 8 | 3 | 0 | 0 | NO |
| 16 | 3 | 0 | 0 | NO |

### 3.3 Actual Query Mix

| threads | Read % | Write % | Other % | Queries / Transaction |
|-------:|------:|-------:|-------:|---------------------:|
| 1 | 70.0 | 20.0 | 10.0 | 20.0 |
| 2 | 70.0 | 20.0 | 10.0 | 20.0 |
| 4 | 70.0 | 20.0 | 10.0 | 20.0 |
| 8 | 70.0 | 20.0 | 10.0 | 20.0 |
| 16 | 70.0 | 20.0 | 10.0 | 20.0 |

---

## 5) Scaling & Efficiency

| threads | TPS_mean | Speedup | Efficiency | ΔTPS vs prev | Δp95 latency |
|-------:|--------:|-------:|----------:|------------:|-------------:|
| 1 | 75.2 | 1.00 | 1.00 | — | — |
| 2 | 142.8 | 1.90 | 0.95 | +67.6 | +2.7 |
| 4 | 255.6 | 3.40 | 0.85 | +112.8 | +4.6 |
| 8 | 423.5 | 5.63 | 0.70 | +167.9 | +9.5 |
| 16 | 512.3 | 6.81 | 0.43 | +88.8 | +33.2 |

**Analysis:**
- Best throughput: threads=8 (5.63x speedup, still 70% efficiency)
- Scaling knee: threads=16 (efficiency drops to 43%)
- Diminishing returns after threads=8

---

## 7) Sanity Checks

| Check | Result | Details |
|-------|--------|---------|
| min ≤ avg ≤ max (latency) | PASS | All runs valid |
| QPS ≈ TPS × queries/transaction | PASS | Expected ≈ TPS × 20 |
| SQL total = read + write + other | PASS | All runs match |
| errors=0 and reconnects=0 | PASS | Clean runs |
| Unknown lines present | NO | Clean parse |

---

## 8) Findings & Recommendations

### 8.1 Key Findings

* **Best throughput point:** threads=8 (TPS=423.5, p95=45.3ms)
* **Scaling knee:** threads=~8-16 (efficiency drops from 70% to 43%)
* **Latency risk point:** threads=16 (p95 jumps 33ms from previous)
* **Stability:** CV increases with threads (8 threads: 6.8%, 16 threads: 8.8%)

### 8.2 Recommendation

* **Suggested production threads:** 8
* **Trade-off:** threads=8 gives 5.63x speedup with acceptable latency (45ms p95)
* **Next experiment:** Repeat with N=5 runs, test innodb_buffer_pool_size variations
```

---

## Questions & Clarifications

1. **Metadata Collection**: Where should we get environment info?
   - Auto-detect from system? (CPU, memory, OS, etc.)
   - User-provided via config file?
   - Mix of both?

2. **Grouping Strategy**: How to identify "same configuration"?
   - Exact match on (threads, database, template)?
   - Allow some dimensions to vary?
   - User-defined grouping rules?

3. **Run Numbering**: How to track N runs per config?
   - Manual tagging during benchmark execution?
   - Auto-detect based on similar timestamps?
   - User selects which runs belong to which config?

4. **Time-Series Data**: Do we need per-second interval data?
   - For steady-state analysis (CV% over time)
   - Requires parsing interval reports from sysbench
   - Increases storage requirements significantly

5. **Report Export Format**:
   - Markdown only? (Your template)
   - HTML for web viewing?
   - JSON for programmatic access?
   - PDF for sharing?

Please provide feedback on this plan before I start implementation.
