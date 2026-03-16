# Oracle Admin Semantics Unification Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Ensure Oracle Swingbench administrative prepare steps consistently execute with `SYS AS SYSDBA` semantics and surface `sqlplus` failure details clearly.

**Architecture:** Keep the current prepare lifecycle intact, but tighten administrative credential resolution so cleanup, verify, and bootstrap use SYSDBA semantics even when the saved connection uses `SYSTEM`. Separately, improve command failure reporting so `sqlplus` stdout diagnostics are included in returned errors.

**Tech Stack:** Go, testify, existing Swingbench adapter and benchmark use case tests

---

## Chunk 1: Admin Credential Semantics

### Task 1: Write the failing adapter test

**Files:**
- Modify: `internal/infra/adapter/swingbench_adapter_test.go`

- [ ] **Step 1: Replace the current system-connection prepare test with one that expects SYSDBA semantics**
- [ ] **Step 2: Run `go test ./internal/infra/adapter -run 'TestSwingbenchAdapter_BuildPrepareCommand_.*System.*' -count=1` and confirm it fails**
- [ ] **Step 3: Update credential resolution logic with the minimal change required**
- [ ] **Step 4: Re-run the same test and confirm it passes**

## Chunk 2: Sqlplus Failure Visibility

### Task 2: Write the failing benchmark use case test

**Files:**
- Modify: `internal/app/usecase/benchmark_usecase_test.go`
- Modify: `internal/app/usecase/benchmark_usecase.go`

- [ ] **Step 1: Add a test that executes a failing `sqlplus`-like command and expects stdout diagnostics in the returned error**
- [ ] **Step 2: Run `go test ./internal/app/usecase -run TestExecuteCommandSync_.*Sqlplus.* -count=1` and confirm it fails**
- [ ] **Step 3: Update sync command failure reporting to include stdout when present**
- [ ] **Step 4: Re-run the same test and confirm it passes**

## Chunk 3: Verification

### Task 3: Run targeted regression coverage

**Files:**
- Modify: `internal/infra/adapter/swingbench_adapter.go`
- Modify: `internal/infra/adapter/swingbench_adapter_test.go`
- Modify: `internal/app/usecase/benchmark_usecase.go`
- Modify: `internal/app/usecase/benchmark_usecase_test.go`

- [ ] **Step 1: Run `go test ./internal/infra/adapter -run 'TestSwingbenchAdapter_BuildPrepareCommand|TestResolveOracleWizardCredentials' -count=1`**
- [ ] **Step 2: Run `go test ./internal/app/usecase -run 'TestExecuteCommandSequence_CreateFailureLogsDiagnosticOutput|TestExecuteCommandSync_.*Sqlplus.*' -count=1`**
- [ ] **Step 3: If both pass, summarize behavior changes and remaining risks**
