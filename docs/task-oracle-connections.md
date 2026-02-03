# Task: Oracle Database Connection Implementation

**Created**: 2026-02-03
**Status**: Pending
**Priority**: High
**Estimate**: 60 minutes
**References**: [plan-oracle-connections.md](./plan-oracle-connections.md)

## Task Breakdown

### Phase 1: Dependency Setup (15 minutes)

#### Task 1.1: Add Oracle Driver to go.mod
- [ ] Run `go get github.com/sijms/go-ora/v2`
- [ ] Run `go mod tidy` to ensure clean dependency tree
- [ ] Verify driver appears in go.mod
- [ ] Check for vulnerabilities: `govulncheck ./...`

**Acceptance**:
- `github.com/sijms/go-ora/v2` appears in go.mod
- `go mod tidy` completes without errors
- No critical vulnerabilities reported

#### Task 1.2: Import Driver in oracle.go
- [ ] Add `_ "github.com/sijms/go-ora/v2"` import
- [ ] Verify build succeeds: `go build ./...`

**Acceptance**:
- `go build ./...` completes without errors
- No "imported but not used" warnings

---

### Phase 2: Implement Test() Method (30 minutes)

#### Task 2.1: Implement Core Test Logic
- [ ] Read existing `oracle.go` to understand current structure
- [ ] Implement Test() method following MySQL/PostgreSQL/SQL Server pattern
- [ ] Add timing measurement
- [ ] Add 10-second timeout context
- [ ] Implement connection open with error handling
- [ ] Implement PingContext with error wrapping
- [ ] Implement version query: `SELECT * FROM v$version WHERE rownum = 1`

**Acceptance**:
- Test() follows established pattern
- Returns TestResult with all fields populated
- Errors are wrapped with context
- Resources cleaned up (defer db.Close())

#### Task 2.2: Verify GetDSNWithPassword() Support
- [ ] Review existing GetDSNWithPassword() implementation
- [ ] Verify it works with go-ora driver
- [ ] Test both Service Name and SID formats

**Acceptance**:
- DSN format is compatible with go-ora
- Both connection modes supported

---

### Phase 3: Testing (15 minutes)

#### Task 3.1: Build and Run Application
- [ ] Run `go build -o db-benchmind ./cmd/db-benchmind`
- [ ] Start application
- [ ] Navigate to Connections page
- [ ] Select Oracle11g connection

**Acceptance**:
- Application builds without errors
- Application starts successfully
- Connections page loads

#### Task 3.2: Test Connection List
- [ ] Click "Test" button on Oracle11g connection
- [ ] Monitor logs for test events
- [ ] Verify success message appears
- [ ] Verify latency is reported
- [ ] Verify database version is retrieved

**Acceptance**:
- Test succeeds with valid credentials
- Latency reported in milliseconds
- Version contains "Oracle Database 11g" or similar
- No sensitive information in logs

#### Task 3.3: Test Connection Edit
- [ ] Click "Edit" on Oracle11g connection
- [ ] Click "Test" button in edit dialog
- [ ] Verify test succeeds
- [ ] Verify behavior matches list test

**Acceptance**:
- Edit test succeeds
- Same latency/version as list test
- No password storage issues

#### Task 3.4: Test Error Cases
- [ ] Create new Oracle connection with invalid password
- [ ] Verify test fails with clear error message
- [ ] Create connection with invalid host
- [ ] Verify test fails with connection error

**Acceptance**:
- Clear error messages for all failure scenarios
- No panics or crashes
- Errors are logged appropriately

---

### Phase 4: Documentation (5 minutes)

#### Task 4.1: Update Implementation Status
- [ ] Mark Oracle as complete in connections analysis
- [ ] Update driver list
- [ ] Update test results

**Acceptance**:
- Documentation reflects new status
- All 4 database types marked as complete

#### Task 4.2: Create Commit
- [ ] Stage all changes: `git add go.mod internal/domain/connection/oracle.go`
- [ ] Commit with message: `feat(oracle): implement Oracle database connection testing`
- [ ] Include Conventional Commits format
- [ ] Reference plan document

**Acceptance**:
- Commit message follows convention
- All required files included
- No test files or temporary files in commit

---

## Verification Checklist

### Code Quality
- [ ] Code follows gofmt formatting
- [ ] No imports unused
- [ ] All errors wrapped with context
- [ ] Resources properly cleaned up
- [ ] No goroutine leaks
- [ ] Thread-safe

### Testing
- [ ] Test succeeds with valid credentials
- [ ] Test fails with invalid credentials
- [ ] Test fails with unreachable host
- [ ] Version query returns data
- [ ] Latency is measured
- [ ] Timeout is enforced

### Documentation
- [ ] plan-oracle-connections.md is complete
- [ ] task-oracle-connections.md is complete
- [ ] Commit message is clear
- [ ] Implementation status updated

### Security
- [ ] No passwords in logs
- [ ] No sensitive data in error messages
- [ ] Dependency checked for vulnerabilities
- [ ] License is compatible (MIT)

---

## Success Criteria

### Definition of Done
1. ✅ Oracle driver imported and registered
2. ✅ Test() method fully implemented
3. ✅ Connection test succeeds with Oracle11g instance
4. ✅ Version retrieved and displayed
5. ✅ Latency measured and reported
6. ✅ All error cases handled gracefully
7. ✅ Code follows existing patterns
8. ✅ Documentation updated
9. ✅ Changes committed to git

### Final Validation
```bash
# Build verification
go build ./...

# Unit tests
go test ./internal/domain/connection/...

# Security check
govulncheck ./...

# Run application
./db-benchmind

# Manual test: Test Oracle11g connection
# Expected: Success=true, latency_ms < 1000, version populated
```

---

## Notes

### Driver-Specific Considerations
1. **v$version query**: Oracle-specific system view, requires proper permissions
2. **Connection format**: go-ora uses `oracle://` prefix with standard Oracle syntax
3. **Default port**: 1521 (standard Oracle port)
4. **Service Name vs SID**: Handled by GetDSNWithPassword() logic

### Known Issues
- None at this time

### Risks
- Oracle connection may require TNS_ADMIN environment variable for advanced configurations
- Some Oracle installations may require wallet files for SSL connections
- v$version access may be restricted on some Oracle installations

### Mitigations
- Test with basic Service Name connection first
- SSL/wallet support can be added later if needed
- Fallback version query: `SELECT SYS_CONTEXT('USERENV', 'DB_VERSION') FROM dual`
