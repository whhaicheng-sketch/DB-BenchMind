# Task: Edit Connection Modal UI/UX Fixes

**Created**: 2026-03-10
**Status**: In Progress
**Priority**: High

## Background

The Edit Connection modal has multiple issues that need to be fixed:
1. Database Type has white background instead of dark theme
2. SSH Host doesn't auto-fill from Host when in Edit mode
3. Password is not properly saved/filled - always resets to empty
4. Database Name field layout is misaligned
5. Test SSH doesn't work properly
6. SSH section layout is uneven
7. Modal is too loose/not compact enough
8. Connection Test area has no result feedback

## Current Issues Found

### Password Issue (Line 263, 271)
```javascript
password: '',  // Always resets to empty
ssh_password: '',  // Always resets to empty
```
**Root cause**: When entering Edit mode, password is always set to empty string instead of loading from saved connection.

### SSH Host Auto-fill (Line 246-247)
```javascript
watch(() => formData.value.ssh_enabled, (enabled) => {
  if (enabled && !formData.value.ssh_host && formData.value.host) {
    formData.value.ssh_host = formData.value.host
  }
})
```
**Issue**: Only triggers when toggling SSH checkbox, not when loading existing SSH-enabled connection in Edit mode.

## Implementation Plan

### Phase 1: Fix Password Saving and Loading
- [ ] Load password from saved connection in Edit mode
- [ ] Don't reset to empty when editing
- [ ] Fix placeholder to show actual password or proper mask
- [ ] Ensure password is not lost on update

### Phase 2: Fix SSH Host Auto-fill
- [ ] Add watcher for Edit mode initialization
- [ ] Auto-fill ssh_host from host when ssh_enabled is true and ssh_host is empty

### Phase 3: Fix Layout Issues
- [ ] Fix Database Type select dark theme
- [ ] Fix Database Name layout
- [ ] Fix SSH section grid alignment
- [ ] Make modal more compact

### Phase 4: Fix Test SSH and Test Feedback
- [ ] Ensure Test SSH uses actual saved credentials
- [ ] Add result feedback display

---

## Success Criteria

1. Password persists after save and re-edit
2. SSH Host auto-fills correctly in Edit mode
3. Database Type uses dark theme
4. SSH section is properly aligned
5. Test SSH shows real feedback
6. Modal is more compact
