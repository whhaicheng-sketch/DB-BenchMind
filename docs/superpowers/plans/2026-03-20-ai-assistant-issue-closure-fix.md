# AI Assistant Issue Closure Fix Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Remove AI-assistant-side save blockers, correct the AI action button order, and make GLM/MiniMax test flows work without regressing DeepSeek/Ollama.

**Architecture:** Keep the change set focused. Separate blocking database validation from advisory AI validation in the frontend state path, and make provider transport handling explicit in the backend compatibility layer where GLM and MiniMax diverge from naive OpenAI assumptions. Preserve existing Wails binding shapes unless a targeted extension is required for testability.

**Tech Stack:** Vue 3, Pinia, Wails bindings, Go HTTP client tests, Node built-in test runner, Go test.

---

### Task 1: Audit save blockers and capture frontend red tests

**Files:**
- Modify: `frontend/tests/connectionFormAi.test.mjs`
- Modify: `frontend/src/components/connection/connectionFormAiState.mjs`
- Modify: `frontend/src/components/connection/ConnectionForm.vue`

- [ ] **Step 1: Write failing tests for blocking vs advisory AI validation**
- [ ] **Step 2: Run `node --test frontend/tests/connectionFormAi.test.mjs` and confirm the new assertions fail**
- [ ] **Step 3: Add minimal validation helpers to separate blocking save errors from advisory AI errors**
- [ ] **Step 4: Update `ConnectionForm.vue` to use the new helpers, clear stale AI state on assistant/provider changes, and reorder buttons**
- [ ] **Step 5: Re-run `node --test frontend/tests/connectionFormAi.test.mjs` and confirm green**

### Task 2: Capture backend red tests for GLM and MiniMax transport behavior

**Files:**
- Modify: `internal/transportwails/bindings/aiprovider/builder_test.go`
- Modify: `internal/transportwails/bindings/aiprovider/builder.go`

- [ ] **Step 1: Write failing tests for GLM base-path preservation, MiniMax endpoint defaults, and non-blocking manual-model chat flow**
- [ ] **Step 2: Run `go test ./internal/transportwails/bindings/aiprovider -run 'Test(GLM|MiniMax|QueryModels|SendChat)'` and confirm failures**
- [ ] **Step 3: Implement the smallest provider-specific fixes in `builder.go`**
- [ ] **Step 4: Re-run the targeted Go tests and confirm green**

### Task 3: Verify focused build/test outcomes and summarize residual risk

**Files:**
- No planned source changes unless verification exposes a direct regression

- [ ] **Step 1: Run `node --test frontend/tests/connectionFormAi.test.mjs`**
- [ ] **Step 2: Run `npm run build` in `frontend/`**
- [ ] **Step 3: Run `go test ./internal/transportwails/bindings/aiprovider`**
- [ ] **Step 4: Run `go test ./internal/transportwails/...` and record any pre-existing unrelated failures**
- [ ] **Step 5: Summarize audit results, file changes, behavior, verification, and residual risks**
