# SSL/TLS Auto Detect Implementation Plan

## Phase 1: TestResult Enhancement
Add TLS status field to TestResult

**File**: `internal/domain/connection/connection.go`
- Add TLSMode field to TestResult struct

## Phase 2: MySQL Implementation
**File**: `internal/domain/connection/mysql.go`
- Current: Uses "false", "preferred", "skip-verify", "true"
- Change order to: "skip-verify" → "preferred" → "false"
- Add TLS mode detection logic

## Phase 3: PostgreSQL Implementation
**File**: `internal/domain/connection/postgresql.go`
- Change order to: "verify-full" → "verify-ca" → "require" → "disable"
- Add TLS mode detection logic

## Phase 4: SQL Server Implementation
**File**: `internal/domain/connection/sqlserver.go`
- Implement encrypt + trustServerCertificate combination
- Add TLS mode detection logic

## Phase 5: Oracle Implementation
**File**: `internal/domain/connection/oracle.go`
- Implement TCPS vs TCP detection

## Implementation Order

1. Update TestResult struct
2. Fix MySQL SSL order (skip-verify → preferred → false)
3. Fix PostgreSQL SSL order
4. Implement SQL Server TLS detection
5. Implement Oracle TLS detection
6. Test all database types
