# Connection Page UI Implementation Plan

## Phase 1: SSL Mode Auto Detection Fix (Backend)

**Issue**: MySQL SSL mode "required" is invalid
- File: `internal/domain/connection/mysql.go`
- Change SSL mode list from: "disabled", "preferred", "required"
- To: "false", "preferred", "skip-verify", "true"

## Phase 2: Enable Test Buttons in Create Mode (Frontend)

**Issue 1**: Database Test button disabled in New Connection
- File: `frontend/src/components/connection/ConnectionForm.vue`
- Enable button when required fields filled (not requiring saved connection)

**Issue 2**: SSH Test button disabled in New Connection
- Enable when SSH enabled and username filled

---

## Implementation Order

1. Fix MySQL SSL mode in backend
2. Fix Database Test enable logic in create mode
3. Fix SSH Test enable logic in create mode

---

## Testing Points

1. Database Test in New Connection mode
2. Database Test in Edit mode
3. SSH Test in New Connection mode
4. SSH Test in Edit mode
