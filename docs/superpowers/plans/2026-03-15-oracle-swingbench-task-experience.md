# Oracle Swingbench Task Experience Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix Oracle + Swingbench task timing semantics, prepare-phase throughput messaging, Oracle prepare permission guidance, and real log viewer opening behavior.

**Architecture:** Keep phase timing derived from `PhaseHistory` as the source of truth, expose aggregated phase/total durations on the task DTO, and render them explicitly in the tasks monitor UI. Add a dedicated Wails log-opening binding and classify Oracle prepare privilege failures at the task layer so the UI can surface actionable guidance without changing non-Oracle tools.

**Tech Stack:** Go, Wails, Vue 3, Pinia, Vitest/Jest-style frontend tests if present, Go `testing`

---

## Chunk 1: Backend Semantics

### Task 1: Add failing backend tests for phase timing and duration semantics

**Files:**
- Modify: `internal/transportwails/bindings/task_test.go`
- Test: `internal/transportwails/bindings/task_test.go`

- [ ] **Step 1: Write failing tests for phase duration aggregation**
- [ ] **Step 2: Run targeted Go tests and confirm failures**
- [ ] **Step 3: Implement task timing aggregation sourced from `PhaseHistory`**
- [ ] **Step 4: Re-run targeted Go tests and confirm pass**

### Task 2: Add failing backend tests for Oracle prepare privilege classification

**Files:**
- Modify: `internal/transportwails/bindings/task_test.go`
- Modify: `internal/infra/adapter/swingbench_adapter_test.go`

- [ ] **Step 1: Write failing tests covering ORA-01031 post-schema setup guidance**
- [ ] **Step 2: Run targeted Go tests and confirm failures**
- [ ] **Step 3: Implement Oracle prepare error classification and messaging**
- [ ] **Step 4: Re-run targeted Go tests and confirm pass**

### Task 3: Add failing backend tests for log viewer path selection/open behavior

**Files:**
- Modify: `internal/transportwails/bindings/task_test.go`
- Modify: `internal/transportwails/bindings/task.go`

- [ ] **Step 1: Write failing tests for current-phase-first and recent-phase fallback selection**
- [ ] **Step 2: Run targeted Go tests and confirm failures**
- [ ] **Step 3: Implement Wails binding for opening task logs with explicit errors**
- [ ] **Step 4: Re-run targeted Go tests and confirm pass**

## Chunk 2: Frontend Behavior

### Task 4: Add failing frontend tests or assertions for phase-aware rendering

**Files:**
- Modify: `frontend/src/components/tabs/TasksMonitorTab.vue`
- Modify: `frontend/src/stores/task.js`

- [ ] **Step 1: Add failing coverage or deterministic assertions for prepare messaging and timing labels**
- [ ] **Step 2: Run targeted frontend tests if available, otherwise use build/lint verification**
- [ ] **Step 3: Implement explicit Prepare / Run / Total rendering and prepare-phase TPS/TPM messaging**
- [ ] **Step 4: Re-run frontend verification**

### Task 5: Wire real log opening feedback in UI

**Files:**
- Modify: `frontend/src/lib/taskBinding.js`
- Modify: `frontend/src/stores/task.js`
- Modify: `frontend/src/components/tabs/TasksMonitorTab.vue`

- [ ] **Step 1: Add failing coverage or deterministic checks for log-open error handling**
- [ ] **Step 2: Implement frontend binding/store call and user-visible error feedback**
- [ ] **Step 3: Re-run frontend verification**

## Chunk 3: Verification and Docs

### Task 6: Keep Oracle documentation aligned with code semantics

**Files:**
- Modify: `docs/USER_GUIDE.md`
- Modify: `contracts/templates/README.md`

- [ ] **Step 1: Document run-duration semantics and Oracle prepare privilege requirement**
- [ ] **Step 2: Verify wording matches implemented behavior**

### Task 7: Final verification

**Files:**
- Verify only

- [ ] **Step 1: Run targeted Go tests**
- [ ] **Step 2: Run frontend verification command**
- [ ] **Step 3: Summarize automated and manual verification gaps, especially live Oracle checks**
