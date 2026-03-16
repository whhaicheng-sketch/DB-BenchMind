# Swingbench Direct Launcher Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make Oracle Swingbench steps bypass the bundled `bin/oewizard` wrapper so DB-BenchMind can keep the real process working directory in the run workspace and avoid `bin/debug.log` permission failures.

**Architecture:** Keep the adapter command shape unchanged, but normalize Swingbench command parts more aggressively inside the benchmark use case. When a command targets `charbench`, `minibench`, or `oewizard`, replace the shell wrapper entrypoint with a direct Java launcher invocation using absolute classpath/config paths while preserving the original runtime arguments and workspace `Dir`.

**Tech Stack:** Go, existing benchmark use case command normalization tests, Swingbench launcher bootstrap packaging.

---

## Chunk 1: Red Test for Direct Java Launch

### Task 1: Lock the launcher rewrite behavior in tests

**Files:**
- Modify: `internal/app/usecase/benchmark_usecase_test.go`
- Modify: `internal/app/usecase/benchmark_usecase.go`

- [ ] **Step 1: Write the failing test**

Add a focused test for `normalizeSwingbenchCommandParts()` asserting that `./oewizard -cl -c oewizard.xml ...` rewrites to a direct `java -cp <abs launcher> LauncherBootstrap ...` command instead of `/opt/benchtools/swingbench/bin/oewizard`.

- [ ] **Step 2: Run the focused test and verify it fails**

Run: `go test ./internal/app/usecase -run TestNormalizeSwingbenchCommandParts -count=1`
Expected: FAIL because the current code still rewrites `./oewizard` to the bundled wrapper script.

## Chunk 2: Minimal Launcher Rewrite

### Task 2: Replace wrapper-script execution with direct launcher execution

**Files:**
- Modify: `internal/app/usecase/benchmark_usecase.go`

- [ ] **Step 1: Implement the minimal rewrite**

Update `normalizeSwingbenchCommandParts()` so Swingbench wrapper script invocations are converted to direct Java launcher commands with absolute launcher/config paths while preserving non-bundled runtime file paths like result/debug outputs.

- [ ] **Step 2: Re-run the focused test**

Run: `go test ./internal/app/usecase -run TestNormalizeSwingbenchCommandParts -count=1`
Expected: PASS.

## Chunk 3: Targeted Regression Verification

### Task 3: Confirm nearby behavior still holds

**Files:**
- Test: `internal/app/usecase/benchmark_usecase_test.go`

- [ ] **Step 1: Run targeted Swingbench use case tests**

Run: `go test ./internal/app/usecase -run 'Test(NormalizeSwingbenchCommandParts|IsSwingbenchCommandLine|IngestSwingbenchDebugLogs)' -count=1`
Expected: PASS.

- [ ] **Step 2: Record any unverified gap**

Note that live Oracle/Swingbench end-to-end execution is not exercised by these unit tests.
