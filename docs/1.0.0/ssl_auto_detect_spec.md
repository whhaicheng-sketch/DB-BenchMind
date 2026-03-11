# SSL/TLS Auto Detect Specification

## Overview
Automatically detect and try SSL/TLS modes for database connections without user manual selection.

## Requirements

### Target Databases
- MySQL
- PostgreSQL
- SQL Server
- Oracle

### Connection Flow (From Most Secure to Least)
1. **TLS_VERIFIED** - Encrypted with certificate/hostname verification
2. **TLS_ENCRYPTED_UNVERIFIED** - Encrypted without verification
3. **PLAINTEXT** - Plain text (only if explicitly enabled)

### Error Handling Rules
- **Stop immediately** for: authentication errors, database not found, network unreachable
- **Try next level** only for: TLS/certificate errors
- Final status must indicate: TLS_VERIFIED / TLS_ENCRYPTED_UNVERIFIED / PLAINTEXT / FAILED

## Database-Specific Implementation

### MySQL
SSL modes in order (most secure to least):
1. `skip-verify` (TLS encrypted, verify server certificate)
2. `preferred` (TLS encrypted, don't verify)
3. `false` (plain text - explicit flag only)

### PostgreSQL
SSL modes in order (most secure to least):
1. `verify-full` (TLS with full verification)
2. `verify-ca` (TLS with CA verification)
3. `require` (TLS without verification)
4. `disable` (plain text - explicit flag only)

### SQL Server
Using `encrypt` + `trustServerCertificate` combination:
1. `encrypt=true, trustServerCertificate=false` (TLS verified)
2. `encrypt=true, trustServerCertificate=true` (TLS unverified)
3. `encrypt=false` (plain text - explicit flag only)

### Oracle
TCP vs TCPS:
1. TCPS (TLS with verification)
2. TCP (plain text - explicit flag only)

## Test Result Enhancement

Add TLS status to TestResult:
```go
type TestResult struct {
    Success bool
    LatencyMs int64
    DatabaseVersion string
    Error string
    TLSMode string // TLS_VERIFIED, TLS_ENCRYPTED_UNVERIFIED, PLAINTEXT, FAILED
}
```
