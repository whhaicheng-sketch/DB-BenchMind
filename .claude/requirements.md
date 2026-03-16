# DB-BenchMind Requirements

## SSH Connection Usage

### CRITICAL: SSH is NOT for Database Benchmark Connections

**SSH tunnels MUST NOT be used for database benchmark connections.**

SSH connections are reserved **ONLY** for:
- Collecting performance metrics during benchmark execution
- Retrieving system statistics from the database server
- Monitoring server health (CPU, memory, disk, network)

### In Tasks & Monitor Page:
- **DO NOT** use SSH tunnel for database connections
- **DO NOT** create SSH tunnels for prepare/run/cleanup phases
- **ALWAYS** connect directly to database: `host:port`
- SSH in connection configuration should be **IGNORED** for benchmark execution

### Example:
```
✗ WRONG: localhost:39197 (via SSH tunnel)
✓ CORRECT: 192.168.134.129:3306 (direct connection)
```

### Reason:
SSH adds overhead and complexity that is not needed for benchmark connections.
The benchmark tools (sysbench, swingbench, hammerdb) should connect directly
to the database to measure true performance.

### Implementation Note:
When implementing `checkTablesConfigForRun` or similar validation methods,
always use the direct database host:port, never the SSH tunnel endpoint.

---

## Template Behavior

### Prepare Phase:
- **MySQL/PostgreSQL**: Use UI parameters (tables, table_size, threads, duration, db_name)
- **Oracle**: Uses template-defined workload weights and scale
- **SQL Server**: Uses template-defined warehouse count and parameters
- **DO NOT** override UI parameters with hardcoded values

### Run Phase:
- **MUST** validate that existing tables match expected configuration
- If table count mismatch: Error and ask user to run Cleanup then Prepare
- **DO NOT** proceed with wrong configuration

### Cleanup Phase:
- **MySQL/PostgreSQL**: Drop all tables in the specified database
- **Oracle**: Drop SOE schema/user
- **SQL Server**: Drop benchmark database/tables
