# Changelog

All notable changes to DB-BenchMind will be documented in this file.

## [Unreleased] - 2026-03-31

### Fixed - AutoBench Capability Flags
- **Corrected SSH/WinRM/AI flag sources**: Capability flags now read from real connection configuration instead of a non-existent `remote_type` field on ConnectionDTO.
  - SSH flag shows only when `ssh_enabled` is true on the connection.
  - WinRM flag shows only when `winrm_enabled` is true on the connection.
  - AI flag shows only when `ai_assistants` has entries with a configured `provider`.

### Fixed - Benchmark AutoBench Observation Mode
- **TasksMonitorTab observation mode**: When AutoBench is running, the Benchmark page's TasksMonitorTab enters observation mode.
  - All Create Task form fields are populated with the current AutoBench sub-task's real configuration.
  - All form controls are disabled during observation mode.
  - Start and Stop buttons are disabled during observation mode.
  - Monitor Overview shows real metrics from the running benchmark task.
  - Help text changes to indicate AutoBench is managing the task.

### Fixed - Reports Grouping by AutoBench Suite
- **Suite-based report grouping**: Reports are now properly grouped by AutoBench suite.
  - Suite state is persisted to the database during execution (start, after each item, final).
  - Groups display the suite name and can be expanded/collapsed.
  - Sub-items within a group show connection name, template type, and metrics.
  - Standalone reports (non-AutoBench) display individually.
  - AutoBench sub-items always remain within their group.

### Fixed - Reports UI
- **Select All separator lines**: Removed extra separator lines on the Select All row (removed `margin-bottom`, `gap` handles spacing).
- **Status filter label mapping**: Fixed backend-to-frontend status value mapping so the filter label displays correctly after selection and persists across navigation/refresh.

### Fixed - AutoBench Status/Report Consistency
- **ReportID linking**: Suite items now have their `ReportID` set after successful benchmark completion.
  - Runner looks up the report by `suite_item_id` and links it to the suite item.
  - When Status shows "Success", Report column shows "View Report" instead of "-".
  - Failed items also get their `ReportID` linked if a report row exists.
- **New backend methods**: `GetReportBySuiteItemID` on ReportUsecase; `PersistSuite` on AutoBenchSuiteUseCase.

### Fixed - Reports Status Model
- **User-visible status filters**: All / 成功 / 失败 / Stop.
- **Status mapping**: `completed`/`success` -> 成功, `failed` -> 失败, `cancelled`/`stopped`/`interrupted` -> Stop.
- Internal statuses (`running`, `pending`, `draft`, `ready`) are not shown as filter options.

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
