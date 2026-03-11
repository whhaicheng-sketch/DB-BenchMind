# SSL/TLS Auto Detect Tasks

## COMPLETED

### Task 1: Enhance TestResult ✅
- Added `TLSMode` field to TestResult struct
- Values: TLS_VERIFIED, TLS_ENCRYPTED_UNVERIFIED, PLAINTEXT, FAILED

### Task 2: MySQL SSL Detection ✅
- Order: skip-verify → preferred → false (fixed from false→preferred→skip-verify)
- Sets TLSMode on success
- **Updated**: Added TLS error filtering - only continue for TLS/certificate errors

### Task 3: PostgreSQL SSL Detection ✅
- Order: verify-full → verify-ca → require → disable
- Sets TLSMode on success
- **Updated**: Added TLS error filtering - only continue for TLS/certificate errors

### Task 4: SQL Server TLS Detection ✅
- Order: encrypt=true + trustServerCertificate=false → encrypt=true + trustServerCertificate=true → encrypt=disable
- Sets TLSMode on success
- **Updated**: Fixed "disabled" → "disable", added TLS error filtering

### Task 5: Oracle TLS Detection ✅
- Marked as PLAINTEXT by default
- (TCPS requires special port configuration)

## Implementation Details

### Error Handling (All Databases)
Only TLS/certificate errors trigger automatic降级:
- TLS errors: contains "tls", "certificate", "x509", "ssl", "handshake"
- Non-TLS errors (auth, network, database not found): stop immediately

### SQL Server Fix
- Changed `encrypt=disabled` → `encrypt=disable` (correct driver parameter)
- Added handshake error detection
