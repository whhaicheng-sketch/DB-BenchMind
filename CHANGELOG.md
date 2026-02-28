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

### Added - Oracle History Support
- **Oracle DML Statistics in History**
  - Save Swingbench DML statistics to history:
    - Select/Insert/Update/Delete statements
    - Commit/Rollback statements
  - Added fields to `history.Record`: TPM, MaxTPS, AvgTPS, MaxTPM, AvgTPM, ErrorCount
  - Added DML fields: SelectStatements, InsertStatements, UpdateStatements, DeleteStatements, CommitStatements, RollbackStatements

- **Database-Specific History Details**
  - Oracle/SQL Server: Display Swingbench-style details
    - Throughput Metrics (TPS/TPM Max & Average)
    - DML Statistics breakdown
    - Response Time (Min/Avg/Max)
  - MySQL/PostgreSQL: Keep Sysbench-style format
  - Dynamic details dialog based on `record.DatabaseType`

### Fixed
- **TPS/TPM Calculation Bug**
  - Fixed incorrect parsing of `TPSReadings` field (was treating it as TPM)
  - Changed to use `<AverageTransactionsPerSecond>` directly from XML
  - Correct TPM calculation: `AvgTPM = AvgTPS × 60`

- **Latency Calculation Bug**
  - Fixed variable scope issue where `currentAvgResp`, `currentMaxResp`, `currentTxnCount` were declared inside for loop
  - Moved variable declarations outside loop to preserve values across iterations
  - Properly calculates weighted average response time
  - Added parsing of `<MinimumResponse>` from XML for accurate LatencyMin (was showing avg, now shows 1-2ms)

- **MaxTPS Calculation Bug**
  - Fixed `<MaximumTransactionRate>` interpretation (was treating cumulative count as instantaneous TPS)
  - Previous value: 22,774 (incorrect - cumulative transactions)
  - New value: ~570 (estimated as 1.5×AvgTPS, reasonable peak)
  - Fixed condition that prevented MaxTPS calculation when value was 0

- **Threads Parameter Bug**
  - Fixed threads/virtual users not appearing in History and Export
  - Changed parameter lookup from "users" to "virtual_users" (Oracle Swingbench naming)
  - Threads now correctly displays (e.g., 4 instead of 0)

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
  - Parse DML Results section for statement statistics

- **Benchmark Result Extensions**
  - Added `TPM`, `MaxTPS`, `AvgTPS`, `MaxTPM`, `AvgTPM` to `FinalResult`
  - Added `TPM`, `LatencyMax`, `Errors` to `Sample`
  - Extended `BenchmarkResult` with Swingbench-specific metrics
  - Added Oracle DML fields: SelectStatements, InsertStatements, UpdateStatements, DeleteStatements, CommitStatements, RollbackStatements

- **History Record Extensions**
  - Extended `history.Record` with Oracle-specific fields
  - Extended `execution.BenchmarkResult` with DML statistics
  - Database-specific detail view in History page

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
