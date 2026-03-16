# Unified Benchmark Lifecycle Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make `prepare`, `run`, and `cleanup` follow one consistent lifecycle across supported databases: `prepare` always rebuilds, `run` reuses the prepared environment, and `cleanup` fully removes it.

**Architecture:** Keep lifecycle orchestration centralized in the benchmark use case while pushing database-specific teardown/build details down into each adapter. Oracle Swingbench prepare becomes an explicit multi-step rebuild flow aligned to the validated manual baseline: `cleanup -> verify cleanup -> bootstrap -> create -> generate -> allindexes`; direct run remains guarded by preflight checks that require a prepared environment after cleanup.

**Tech Stack:** Go, Wails bindings, repository-backed execution logs, built-in template seeds, Markdown docs.

---

## Chunk 1: Lifecycle Semantics and Tests

### Task 1: Lock the desired prepare/cleanup behavior in failing tests

**Files:**
- Modify: `internal/infra/adapter/swingbench_adapter_test.go`
- Modify: `internal/app/usecase/oracle_swingbench_guard_test.go`
- Modify: `internal/domain/template/seeds_test.go`

- [ ] **Step 1: Write failing tests for Oracle Swingbench prepare/cleanup semantics**

Add assertions that `BuildPrepareCommand()` produces a cleanup-first rebuild sequence and that cleanup is a full teardown for SOE user/tablespaces/datafiles.

- [ ] **Step 2: Run targeted Go tests to verify they fail**

Run: `go test ./internal/infra/adapter ./internal/app/usecase ./internal/domain/template`
Expected: FAIL in new lifecycle assertions before implementation.

- [ ] **Step 3: Add/adjust tests for run-after-cleanup guards and test template expectations**

Cover the invariant that cleanup invalidates direct run until a later prepare, and keep built-in test templates aligned with the unified lifecycle wording.

- [ ] **Step 4: Re-run the same targeted tests**

Run: `go test ./internal/infra/adapter ./internal/app/usecase ./internal/domain/template`
Expected: FAIL only because production code has not been updated yet.

## Chunk 2: Adapter and Use Case Implementation

### Task 2: Implement Oracle Swingbench rebuild and full teardown

**Files:**
- Modify: `internal/infra/adapter/swingbench_adapter.go`

- [ ] **Step 1: Embed the bootstrap SQL and wire prepare to cleanup-first rebuild**

Implement embedded/non-interactive `cts_orac.sql` generation plus ordered steps:
cleanup -> verify cleanup -> bootstrap -> create -> generate -> allindexes.

- [ ] **Step 2: Implement full Oracle cleanup semantics**

Make cleanup drop the SOE user and SOE/SOE_IDX tablespaces including datafiles so the next prepare always rebuilds from scratch.

- [ ] **Step 3: Run adapter tests**

Run: `go test ./internal/infra/adapter -run Swingbench -v`
Expected: PASS.

### Task 3: Remove legacy “already exists” tolerance from orchestration

**Files:**
- Modify: `internal/app/usecase/benchmark_usecase.go`

- [ ] **Step 1: Update prepare-only logging/messages to describe rebuild semantics**

Ensure prepare-only mode explicitly logs cleanup-before-prepare as destructive rebuild behavior.

- [ ] **Step 2: Remove continue-on-existing-object behavior from full prepare**

Prepare errors should now fail unless cleanup/prepare itself succeeds; do not silently continue on “already exists”.

- [ ] **Step 3: Run use case tests**

Run: `go test ./internal/app/usecase -v`
Expected: PASS.

## Chunk 3: Templates, Docs, and Verification

### Task 4: Align built-in template/docs wording with the unified lifecycle

**Files:**
- Modify: `docs/USER_GUIDE.md`
- Modify: `docs/OPERATION.md`
- Modify: `contracts/templates/README.md`

- [ ] **Step 1: Update test template wording**

Describe prepare as rebuild/recreate, run as reusable after prepare, and cleanup as full teardown.

- [ ] **Step 2: Update user-facing docs**

Document the lifecycle explicitly, including destructive cleanup semantics, the `cts_orac.sql`-equivalent bootstrap behavior, and the workload-only `charbench` run phase.

- [ ] **Step 3: Run targeted template/doc-adjacent tests**

Run: `go test ./internal/domain/template -v`
Expected: PASS.

### Task 5: Final verification including built-in test-template coverage

**Files:**
- Test: `frontend/scripts/verify-template-test-coverage.mjs`

- [ ] **Step 1: Run focused backend test suites**

Run: `go test ./internal/infra/adapter ./internal/app/usecase ./internal/domain/template`
Expected: PASS.

- [ ] **Step 2: Run template coverage verification if dependencies are available**

Run: `node frontend/scripts/verify-template-test-coverage.mjs`
Expected: PASS.

- [ ] **Step 3: Record any environment-dependent gaps**

If Oracle/MySQL/PostgreSQL/SQL Server live services are unavailable, note that end-to-end database execution was not possible even though template/tests passed.
