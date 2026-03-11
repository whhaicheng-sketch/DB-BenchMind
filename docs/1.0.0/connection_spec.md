# Connection Page UI Specification (Updated 2026-03-08)
# REQ-001 ~ REQ-012

## Overview
This specification defines the UI changes for the Connections page.

---

## REQ-001: Connection Groups Default Collapsed ✅
- Groups collapsed by default on page load

## REQ-002: Toggle All Button ✅
- Single button toggles between Expand/Collapse

## REQ-003: Compact Form Layout ✅
- Two-column layout for form fields
- Reduced font sizes and spacing

## REQ-004: Dropdown White Background ✅
- White background for Database Type dropdown

## REQ-005: Username Default Values ✅
- MySQL: root, PostgreSQL: postgres, Oracle: system, SQL Server: sa

## REQ-006: Auto SSL Mode ✅
- SSL Mode dropdown removed from UI
- Backend auto-detects SSL mode

## REQ-007: Database Test Button ✅
- Test button in form

## REQ-008: SSH Test Button ✅
- SSH Test button when SSH enabled

---

## NEW ISSUE: REQ-011 - SSL Mode Auto Detection Bug

**Problem**: Database Test fails with SSL error:
```
ssl_mode=required: failed to open connection: invalid value / unknown config name: required
```

**Root Cause**: MySQL SSL mode "required" is invalid, should be "skip-verify" or "preferred"

**Fix Required**: Update backend SSL mode mapping

---

## NEW ISSUE: REQ-012 - Test Button Enable Logic

**Problem 1**: Database Test button disabled in New Connection mode
- Should be enabled when required fields (host, port, username) are filled

**Problem 2**: SSH Test button disabled in New Connection mode
- Should be enabled when SSH is enabled and username is filled

**Fix Required**: Update button enable logic to work in create mode (without saved connection)

---

## Technical Fix Required

### 1. Backend: Fix MySQL SSL Mode Mapping
**File**: `internal/domain/connection/mysql.go`

Current (wrong):
```go
sslModes := []string{"disabled", "preferred", "required"}
```

Should be:
```go
sslModes := []string{"false", "preferred", "skip-verify", "true"}
```

### 2. Frontend: Enable Test Buttons in Create Mode
**File**: `frontend/src/components/connection/ConnectionForm.vue`

- Database Test: Enable when host, port, username are filled (even in create mode)
- SSH Test: Enable when SSH enabled and username filled (even in create mode)

Note: For database test in create mode, we need to create a temporary connection and test it directly.
