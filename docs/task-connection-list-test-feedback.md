# Task: Connection List Test Button Feedback

**Created**: 2026-03-10
**Status**: In Progress
**Priority**: High
**References**: This is a bug fix for A3 (列表页 Test 按钮结果反馈)

## Background

The Connections list page Test button has been fixed to call the correct store method (`testConnectionById`), but the test result (success/failed/loading) is not displayed to users.

## Current State

1. **Store**: Single global `testResult` field - insufficient for per-row display
2. **UI**: Test button doesn't show loading/success/failed states
3. **Gap**: Users can't see test results on the list page

## Implementation Plan

### Phase 1: Store Enhancement (10 min)

#### Task 1.1: Add per-connection test state
- Add `testingById: Record<string, boolean>` to track testing state per connection
- Add `testResultById: Record<string, TestResult>` to store results per connection
- Modify `testConnectionById` to update these new fields

**Acceptance**:
- Store can track multiple connection test states independently
- Testing state is cleared after test completes

### Phase 2: UI Update (15 min)

#### Task 2.1: Update ConnectionsTab.vue Test button
- Add visual loading state (spinner/disabled) when testing
- Display success/failed indicator on the row
- Show error message if available
- Use connection ID to isolate states

**Acceptance**:
- Clicking Test shows immediate feedback
- Success shows clear "connected" indicator
- Failure shows error message
- Multiple rows don't interfere with each other

### Phase 3: Verification (10 min)

#### Task 3.1: Single connection test
- Click Test on a connection
- Verify loading state appears
- Verify result is displayed

#### Task 3.2: Multiple connection independence
- Add multiple test connections
- Test each independently
- Verify results are isolated

#### Task 3.3: Regression check
- Verify Add/Edit modal Test Database still works
- Verify SSH Test still works

---

## Verification Checklist

- [ ] Single connection test shows loading state
- [ ] Single connection test shows success result
- [ ] Single connection test shows failed result with error message
- [ ] Multiple connections don't interfere with each other
- [ ] Clicking Test again refreshes the result
- [ ] Add/Edit modal Test Database still works
- [ ] SSH Test in modal still works

## Success Criteria

1. ✅ Clicking list Test shows immediate loading feedback
2. ✅ Success shows clear "OK/Connected" indicator
3. ✅ Failure shows "Failed" with error message
4. ✅ Multiple rows have independent states
5. ✅ Results don't interfere between rows
6. ✅ No regression in Add/Edit modal tests
