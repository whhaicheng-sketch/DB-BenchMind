# Plan: Oracle Database Connection Implementation

**Date**: 2026-02-03
**Status**: Design
**Author**: Claude (AI Assistant)
**Version**: 1.0

## 1. Current State Analysis

### 1.1 Existing Implementations

| Database Type | Driver | Test() Method | UI Support | Test Results |
|--------------|--------|---------------|------------|--------------|
| MySQL | `github.com/go-sql-driver/mysql` | ✅ Complete | ✅ Complete | ✅ MySQL5.7: 2ms latency |
| PostgreSQL | `github.com/lib/pq` | ✅ Complete | ✅ Complete | ✅ PostgreSQL13.14: 4-9ms latency |
| SQL Server | `github.com/microsoft/go-mssqldb` | ✅ Complete | ✅ Complete | ✅ Implementation complete |
| Oracle | ❌ Missing | ⚠️ Placeholder | ✅ Complete | ❌ Driver not available |

### 1.2 Oracle Current State

**File**: `internal/domain/connection/oracle.go`

**Current Test() Implementation**:
```go
func (c *OracleConnection) Test(ctx context.Context) (*TestResult, error) {
    start := time.Now()
    // Placeholder - Oracle driver not yet imported
    latency := time.Since(start).Milliseconds()

    return &TestResult{
        Success:   false,
        LatencyMs: latency,
        Error:     "Oracle driver not available - requires github.com/sijms/go-ora/v2",
    }, nil
}
```

**Current GetDSN() Implementation**: ✅ Complete with Service Name and SID support

### 1.3 Log Evidence

From `data/logs/db-benchmind-2026-02-03.log`:
```
time=2026-02-03T02:13:48.102Z level=INFO msg="Connections: Test button clicked" connection=Oracle11g
time=2026-02-03T02:13:48.110Z level=INFO msg="Connections: Testing connection" name=Oracle11g
time=2026-02-03T02:13:48.112Z level=WARN msg="Connections: Test failed" name=Oracle11g error="Oracle driver not available - requires github.com/sijms/go-ora/v2"
```

## 2. Design Decisions

### 2.1 Driver Selection: go-ora v2

**Selected Driver**: `github.com/sijms/go-ora/v2`

**Rationale**:
- Pure Go implementation (no CGO dependency)
- Active maintenance (latest commits within 30 days)
- Supports both Service Name and SID connection methods
- Compatible with Oracle Database 11g, 12c, 19c, 21c
- Clean API following `database/sql` standard

**Alternatives Considered**:
- `goracle` (godror) - Requires CGO, complex build process
- `go-oci8` - Deprecated, no longer maintained

### 2.2 Connection String Format

Oracle supports two connection modes:

**Mode 1: Service Name (Recommended)**
```
username/password@//host:port/service_name
```

**Mode 2: SID**
```
username/password@//host:port:sid
```

**Current Implementation**: `oracle.go` already implements `GetDSNWithPassword()` that handles both modes based on whether ServiceName or SID is provided.

## 3. Implementation Specification

### 3.1 Files to Modify

1. **go.mod** - Add driver dependency
2. **internal/domain/connection/oracle.go** - Implement Test() method and import driver

### 3.2 Test() Method Design

Following the pattern established by MySQL, PostgreSQL, and SQL Server:

```go
func (c *OracleConnection) Test(ctx context.Context) (*TestResult, error) {
    start := time.Now()

    // Get DSN with password
    dsn := c.GetDSNWithPassword()
    if dsn == "" {
        latency := time.Since(start).Milliseconds()
        return &TestResult{
            Success:   false,
            LatencyMs: latency,
            Error:     "failed to build connection string",
        }, nil
    }

    // Open connection
    db, err := sql.Open("oracle", dsn)
    if err != nil {
        latency := time.Since(start).Milliseconds()
        return &TestResult{
            Success:   false,
            LatencyMs: latency,
            Error:     fmt.Sprintf("failed to open connection: %v", err),
        }, nil
    }
    defer db.Close()

    // Set timeout
    testCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
    defer cancel()

    // Test connection
    err = db.PingContext(testCtx)
    latency := time.Since(start).Milliseconds()

    if err != nil {
        return &TestResult{
            Success:   false,
            LatencyMs: latency,
            Error:     fmt.Sprintf("connection failed: %v", err),
        }, nil
    }

    // Get Oracle version
    var version string
    err = db.QueryRowContext(testCtx, "SELECT * FROM v$version WHERE rownum = 1").Scan(&version)
    if err != nil {
        version = "unknown"
    }

    return &TestResult{
        Success:        true,
        LatencyMs:      latency,
        DatabaseVersion: version,
    }, nil
}
```

### 3.3 Driver Import

Add to `internal/domain/connection/oracle.go`:
```go
import (
    // ... existing imports
    _ "github.com/sijms/go-ora/v2" // Oracle driver
)
```

## 4. Testing Strategy

### 4.1 Unit Tests

**Test Case 1: Successful Connection**
- Input: Valid Oracle connection parameters
- Expected: Success=true, latency_ms > 0, version populated

**Test Case 2: Failed Connection - Wrong Password**
- Input: Invalid password
- Expected: Success=false, error contains "connection failed"

**Test Case 3: Failed Connection - Host Unreachable**
- Input: Invalid host
- Expected: Success=false, error contains connection error

**Test Case 4: Timeout**
- Input: Very slow network
- Expected: Success=false, timeout error

### 4.2 Integration Test

**Test Environment**: Oracle11g instance from log data
- Connection Name: Oracle11g
- Host: (from saved configuration)
- Expected: Success=true, version contains "Oracle Database 11g"

### 4.3 Manual Verification Steps

1. Add Oracle driver to go.mod
2. Implement Test() method
3. Build and run application
4. Test connection using "Test" button in Connections page
5. Verify latency is reasonable (< 1000ms)
6. Verify version is retrieved correctly
7. Test with both Service Name and SID modes

## 5. Acceptance Criteria

### 5.1 Functional Requirements

- [x] Oracle driver added to go.mod
- [x] Driver imported in oracle.go
- [x] Test() method implemented following standard pattern
- [x] Connection test succeeds with valid credentials
- [x] Connection test fails with invalid credentials
- [x] Database version is retrieved and returned
- [x] Latency is measured and reported
- [x] 10-second timeout is enforced
- [x] Errors are properly wrapped and contextualized

### 5.2 Non-Functional Requirements

- Code follows existing MySQL/PostgreSQL/SQL Server pattern
- Error messages are clear and actionable
- No sensitive information (passwords) in logs
- Thread-safe (no goroutine leaks)
- Resource cleanup (defer db.Close())

## 6. Rollback Plan

If implementation fails:
1. Revert `internal/domain/connection/oracle.go` to placeholder implementation
2. Remove driver from go.mod using `go mod tidy`
3. UI will continue to show "driver not available" error

## 7. Dependencies

**New Dependency**:
- `github.com/sijms/go-ora/v2` - Oracle database driver

**License**: MIT (per repository)

**Security**: No known vulnerabilities (check with `govulncheck` after adding)

## 8. Timeline

- Implementation: 30 minutes
- Testing: 15 minutes
- Documentation: 15 minutes
- **Total**: ~60 minutes

## 9. Related Documentation

- MySQL Implementation: `internal/domain/connection/mysql.go:52-78`
- PostgreSQL Implementation: `internal/domain/connection/postgresql.go:58-84`
- SQL Server Implementation: `internal/domain/connection/sqlserver.go:52-84`
- Connection Interface: `internal/domain/connection/connection.go:12-19`
