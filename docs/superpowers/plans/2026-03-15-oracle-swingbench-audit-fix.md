# Oracle Swingbench Audit Fix Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Audit and fix the Oracle + Swingbench prepare/run/cleanup/stop/metrics/logs/overlay/state-machine flow so prepare remains unified as `cleanup + init/create`, cleanup-invalidated runs fail fast, stop is real and idempotent, waiting overlays only appear for healthy short-lived startup, and failed status cards stay concise.

**Architecture:** Tighten backend state enforcement in the Oracle Swingbench-specific execution path, ensuring prepare performs exactly one cleanup, verifies residual state before create, and logs explicit credential/cleanup diagnostics. Surface those backend decisions in the task binding and frontend status-strip/overlay model while keeping scope constrained to Oracle + Swingbench and Tasks & Monitor status presentation.

**Tech Stack:** Go, Node test runner, Vue 3, Pinia, Wails task bindings

---

## Chunk 1: Backend Guards And Runtime Supervision

### Task 1: Lock Oracle Swingbench direct-run preflight semantics

**Files:**
- Modify: `internal/app/usecase/oracle_swingbench_guard_test.go`
- Modify: `internal/app/usecase/benchmark_usecase.go`

- [ ] **Step 1: Write failing tests for cleanup-aware invalidation and preflight result ordering**

Add backend tests covering:
- cleanup success followed by direct run returns `Cleanup removed required SOE objects. Please run Prepare first.`
- later successful prepare clears the invalidation guard
- preflight status model preserves discrete outcomes for user missing, account locked, invalid credentials, schema incomplete, cleanup invalidated

- [ ] **Step 2: Run backend Oracle Swingbench guard tests and confirm they fail**

Run: `go test ./internal/app/usecase -run 'TestOracleSwingbench' -count=1`
Expected: FAIL on missing cleanup-aware direct-run guard and missing discrete result handling

- [ ] **Step 3: Implement the minimal backend guard changes**

Update `benchmark_usecase.go` to:
- add a discrete Oracle Swingbench preflight result model
- enforce fixed-order checks for user existence, login, schema completeness, cleanup-history invalidation
- add helper(s) to inspect Oracle Swingbench environment history using connection/tool/workload identity
- return the exact cleanup-invalidated error message before starting the run phase

- [ ] **Step 4: Re-run backend Oracle Swingbench guard tests**

Run: `go test ./internal/app/usecase -run 'TestOracleSwingbench' -count=1`
Expected: PASS

### Task 2: Fail fast for zero-throughput dead state

**Files:**
- Modify: `internal/transportwails/bindings/task_test.go`
- Modify: `internal/transportwails/bindings/task.go`
- Modify: `internal/app/usecase/oracle_swingbench_guard_test.go`

- [ ] **Step 1: Write failing tests for Oracle Swingbench stalled run escalation**

Add tests covering:
- run phase with repeated `[0/N]`, zero TPS/TPM, no completed transactions escalates to failed quickly
- cleanup-invalidated zero-progress run points users to Prepare instead of remaining in waiting

- [ ] **Step 2: Run the targeted task binding and guard tests to confirm red**

Run: `go test ./internal/transportwails/bindings ./internal/app/usecase -run 'Test(OracleSwingbench|TaskBinding)' -count=1`
Expected: FAIL on missing runtime escalation path

- [ ] **Step 3: Implement minimal zero-throughput supervision**

Update task binding/runtime logic to:
- detect Oracle Swingbench stalled run state before the benchmark naturally exits
- classify cleanup-invalidated and generic zero-progress failures distinctly
- stop waiting-state propagation once fatal runtime evidence exists

- [ ] **Step 4: Re-run the targeted backend tests**

Run: `go test ./internal/transportwails/bindings ./internal/app/usecase -run 'Test(OracleSwingbench|TaskBinding)' -count=1`
Expected: PASS

### Task 3: Make stop idempotent and terminate real processes

**Files:**
- Modify: `internal/app/usecase/benchmark_usecase_test.go`
- Modify: `internal/transportwails/bindings/task_test.go`
- Modify: `internal/app/usecase/benchmark_usecase.go`
- Modify: `internal/transportwails/bindings/task.go`

- [ ] **Step 1: Write failing tests for one-shot stop and terminal state closure**

Add tests covering:
- repeated `StopTask` requests only append one stop event and call backend stop once
- stop transitions final task state away from `stopping`
- backend stop path kills the benchmark process/process group and marks run/task terminal

- [ ] **Step 2: Run the targeted stop tests and confirm failure**

Run: `go test ./internal/app/usecase ./internal/transportwails/bindings -run 'Test(BenchmarkUseCase_Stop|TaskBinding_Stop)' -count=1`
Expected: FAIL on duplicate stop handling and incomplete stop closure

- [ ] **Step 3: Implement minimal stop-state fixes**

Update Go code to:
- make `StopTask` idempotent while a stop is already in flight
- ensure task status/logging does not re-emit `Stop requested`
- terminate the actual process tree for Swingbench/charbench on Linux
- propagate cancelled/stopped completion back into phase history, `CompletedAt`, and task terminal state

- [ ] **Step 4: Re-run the targeted stop tests**

Run: `go test ./internal/app/usecase ./internal/transportwails/bindings -run 'Test(BenchmarkUseCase_Stop|TaskBinding_Stop)' -count=1`
Expected: PASS

## Chunk 2: Frontend Overlay And Active Task Binding

### Task 4: Tighten Oracle Swingbench overlay priority and failure summary ownership

**Files:**
- Modify: `frontend/tests/tasksMonitorOracleSwingbench.test.mjs`
- Modify: `frontend/src/components/tabs/tasksMonitorOracleSwingbenchState.mjs`

- [ ] **Step 1: Write failing frontend tests for stopping/error/cleanup-invalidated priority**

Add tests covering:
- Oracle error-like overlays are removed from TPS/TPM cards
- cleanup-invalidated state renders no error overlay
- stopping renders no error overlay
- prepare overlay still applies only to prepare phase
- failed status strip only shows short summary text and "See logs for details"

- [ ] **Step 2: Run the Oracle Swingbench frontend tests and confirm failure**

Run: `cd frontend && node --test tests/tasksMonitorOracleSwingbench.test.mjs`
Expected: FAIL on missing stopping/cleanup-invalidated overlay states

- [ ] **Step 3: Implement minimal overlay-state changes**

Update overlay/status-strip helpers to:
- keep only `prepare`, `run-waiting`, `none` metric overlays
- remove Oracle-specific error/stopping overlays from TPS/TPM cards
- move failure ownership to the top-left status summary with short bounded text only
- use backend/task history evidence rather than only “no samples yet”

- [ ] **Step 4: Re-run the Oracle Swingbench frontend tests**

Run: `cd frontend && node --test tests/tasksMonitorOracleSwingbench.test.mjs`
Expected: PASS

### Task 5: Keep top-bar actions bound to active task and block stop re-entry

**Files:**
- Modify: `frontend/tests/tasksMonitorTaskState.test.mjs`
- Modify: `frontend/tests/taskBinding.test.mjs`
- Modify: `frontend/src/components/tabs/TasksMonitorTab.vue`
- Modify: `frontend/src/components/tabs/tasksMonitorTaskState.mjs`
- Modify: `frontend/src/stores/task.js`

- [ ] **Step 1: Write failing frontend tests for stop de-duplication and active-task-only bindings**

Add tests covering:
- stop click enters stop-in-flight and prevents re-entry
- log viewer and status strip keep tracking active task, not draft
- stopping task still resolves as active/log-viewer target

- [ ] **Step 2: Run the targeted frontend binding tests and confirm red**

Run: `cd frontend && node --test tests/tasksMonitorTaskState.test.mjs tests/taskBinding.test.mjs`
Expected: FAIL on missing stop-in-flight and/or binding assertions

- [ ] **Step 3: Implement minimal binding/store/UI changes**

Update frontend code to:
- track stop-in-flight by active task id
- disable repeat stop dispatch while stopping
- preserve log viewer/status strip binding to active task/current task only

- [ ] **Step 4: Re-run the targeted frontend binding tests**

Run: `cd frontend && node --test tests/tasksMonitorTaskState.test.mjs tests/taskBinding.test.mjs`
Expected: PASS

## Chunk 3: Verification

### Task 6: Run focused verification and capture audit evidence

**Files:**
- Review only: `internal/app/usecase/benchmark_usecase.go`
- Review only: `internal/transportwails/bindings/task.go`
- Review only: `frontend/src/components/tabs/tasksMonitorOracleSwingbenchState.mjs`
- Review only: `frontend/src/components/tabs/TasksMonitorTab.vue`

- [ ] **Step 1: Run backend verification suite**

Run: `go test ./internal/app/usecase ./internal/transportwails/bindings -count=1`
Expected: PASS

- [ ] **Step 2: Run frontend verification suite**

Run: `cd frontend && node --test tests/tasksMonitorOracleSwingbench.test.mjs tests/tasksMonitorTaskState.test.mjs tests/taskBinding.test.mjs tests/tasksMonitorStatusStrip.test.mjs`
Expected: PASS

- [ ] **Step 3: Re-read changed code against the Oracle Swingbench audit checklist**

Confirm code evidence exists for:
- cleanup-after-run direct run fails before waiting
- stop requests are one-shot and terminal
- stopping/error overlays outrank waiting
- zero-throughput dead state fails fast

- [ ] **Step 4: Summarize findings and residual risks**

Prepare the final report in the required audit format with file references and verification evidence.
