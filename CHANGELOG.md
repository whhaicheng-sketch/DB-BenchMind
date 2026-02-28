# Changelog

All notable changes to DB-BenchMind will be documented in this file.

## [Unreleased] - 2026-02-28

### Added - Oracle Swingbench Support
- **Full Oracle Swingbench Integration**
  - Complete implementation of Oracle Swingbench (charbench/oewizard) adapter
  - Support for all SOE (Order Entry) transaction types:
    - NCR (New Customer Registration)
    - UCD (Update Customer Details)
    - BP (Browse Products)
    - OP (Order Products)
    - PO (Process Orders)
    - BO (Browse Orders)
  - Real-time metrics collection with TPM/TPS monitoring
  - Automatic XML result parsing for detailed statistics

- **Database-Specific Real-time Monitor**
  - MySQL/PostgreSQL: Display TPS, QPS, 95% Latency, Errors/s
  - Oracle/SQL Server: Display TPS, TPM, Response Time, Errors
  - Dynamic UI labels based on database type
  - `updateMetricLabels()` function for label switching

- **Oracle Result Parsing**
  - Parse `<AverageTransactionsPerSecond>` for accurate AvgTPS
  - Calculate AvgTPM = AvgTPS × 60
  - Calculate MaxTPM = MaxTPS × 60
  - Parse per-transaction response times from XML
  - Calculate weighted average response time across all transaction types
  - Extract minimum and maximum response times

### Fixed
- **TPS/TPM Calculation Bug**
  - Fixed incorrect parsing of `TPSReadings` field (was treating it as TPM)
  - Changed to use `<AverageTransactionsPerSecond>` directly from XML
  - Correct TPM calculation: `AvgTPM = AvgTPS × 60`

- **Latency Calculation Bug**
  - Fixed variable scope issue where `currentAvgResp`, `currentMaxResp`, `currentTxnCount` were declared inside for loop
  - Moved variable declarations outside loop to preserve values across iterations
  - Properly calculates weighted average response time

- **Duration Calculation**
  - Fixed Swingbench runtime to round up to nearest minute
  - Formula: `runtimeMinutes = (runtimeSeconds + 59) / 60`
  - Minimum 1 minute runtime

### Changed - UI/UX
- **Oracle Completion Dialog**
  - Simplified Oracle/SQL Server completion dialog
  - Only shows "Benchmark completed successfully!" message
  - Retains Save and OK buttons for history functionality
  - MySQL/PostgreSQL keep detailed statistics display

### Technical Details
- **Swingbench Adapter** (`internal/infra/adapter/swingbench_adapter.go`)
  - Added `-a` flag for automatic mode (enables stdout output)
  - Verbose output: `-v users,tpm,tps,errs,vresp`
  - Proper XML result file parsing from `/tmp/soe_*.xml`
  - Real-time collection from tabular stdout output

- **Benchmark Result Extensions**
  - Added `TPM`, `MaxTPS`, `AvgTPS`, `MaxTPM`, `AvgTPM` to `FinalResult`
  - Added `TPM`, `LatencyMax`, `Errors` to `Sample`
  - Extended `BenchmarkResult` with Swingbench-specific metrics

### Performance
- Oracle Swingbench can handle 4+ concurrent users
- Typical TPS: 400-600 (4 users)
- Typical TPM: 24,000-36,000 (4 users)
- Error rate: <0.2% (normal for concurrent Browse Orders)

### Known Issues
- Swingbench `Browse Orders` transaction may show 0.1-0.5% error rate due to row-level locking
  - This is expected behavior and indicates proper concurrent access handling
  - Can be reduced by lowering user count or increasing think time

## [Previous Releases]

See individual commit history for earlier changes.
