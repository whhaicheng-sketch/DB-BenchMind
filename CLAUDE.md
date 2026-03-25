# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

DB-BenchMind is a desktop database benchmark workbench built with **Wails v2** (Go backend + Vue 3 frontend). It provides a unified GUI for managing database connections, running benchmarks (Sysbench/Swingbench/HammerDB), and generating reports.

**Core principle documents** (read these first for new features):
- `constitution.md` — Non-negotiable development principles (TDD, library-first, CLI mandates)
- `.specify/steering/` — Architecture decisions, testing strategy, product scope

## Build & Test Commands

```bash
# Build (Wails + frontend)
make build          # Full build via scripts/entry/build

# Run application
make run            # Build and start

# Test commands
make test           # Run backend + frontend tests
make test-backend   # Go tests for specific packages (see scripts/gates/test_backend_gate.sh)
make test-frontend  # npm test in frontend/
make test-unit      # go test -short ./...
make lint           # golangci-lint run

# Release gates (CI)
make release-gate   # Full regression suite
make check          # format-check + regression + lint

# Development
make deps           # Download dependencies, go mod tidy
make vulncheck      # govulncheck ./...
```

## Architecture

```
cmd/db-benchmind-cli/     CLI entry point
main.go                   Wails GUI entry point
internal/
├── domain/               # Core business logic (NO external deps, stdlib only)
│   ├── connection/       MySQL, Oracle, PostgreSQL, SQL Server models
│   ├── template/         Benchmark templates
│   ├── execution/        Run state machine
│   └── autobench/        AutoBench domain models
├── app/usecase/          # Orchestration layer, defines interfaces
│   ├── connection_usecase.go
│   ├── template_usecase.go
│   ├── benchmark_usecase.go
│   └── autobench_usecase.go
├── infra/                # External dependencies implementation
│   ├── database/         SQLite repositories (modernc.org/sqlite)
│   ├── adapter/          Sysbench, Swingbench, HammerDB adapters
│   └── keyring/          go-keyring with file fallback
├── transportwails/       # Wails bindings and collectors
│   ├── app.go
│   ├── bindings/         Go-to-JS bindings
│   └── collector/        System metrics collection
frontend/
├── src/
│   ├── components/       Vue 3 components (tabs, charts, forms)
│   └── App.vue
├── tests/                Node.js test runner (*.test.mjs)
└── package.json
```

**Dependency flow**: `transportwails → app/usecase → domain ← infra` (infra implements usecase interfaces)

## Key Technical Decisions

1. **No ORM**: Direct `database/sql` with prepared statements (constitution.md Article VIII)
2. **No CGO**: Pure Go SQLite (`modernc.org/sqlite`), Wails for cross-platform GUI
3. **State machine**: Run execution uses explicit state transitions (see `internal/domain/execution/state.go`)
4. **Table-driven tests**: Required format for Go tests (constitution.md Article III)
5. **Fake over mock**: Prefer in-memory/httptest over interface mocks

## Frontend

- **Framework**: Vue 3 + Pinia + ECharts
- **Build**: Vite
- **Tests**: `node --test tests/*.test.mjs`
- **Wails bindings**: Auto-generated in `frontend/wailsjs/`

Run frontend tests: `npm --prefix frontend run test`

## Development Workflow

Per constitution.md, implement new features by:

1. **Read existing code** first (entry points, call chains, data structures)
2. **Write implementation plan** covering: scope, design, error handling, test cases, risks
3. **Wait for confirmation** before coding
4. **TDD**: Write failing test, implement, refactor

### Error Handling

Cross-boundary calls must wrap errors with context:
```go
fmt.Errorf("read config: %w", err)  // op describes what was being done
```

### Logging

Use `log/slog` with unified field keys: `trace_id`, `req_id`, `user_id`, `op`, `err`, `latency_ms`

### Context

First parameter, always propagated:
```go
func (s *Service) DoSomething(ctx context.Context, arg string) error
```

## Concurrency Safety

When using goroutines/channels/mutex, document:
- Race risk points
- Control measures (mutex/channel/atomic/errgroup)
- Ensure `go test ./... -race` passes

## Commit Convention

Conventional Commits required: `feat(scope): description`, `fix(scope): description`

## Pre-commit Checklist

- [ ] `gofmt` / `goimports` applied
- [ ] `go mod tidy` run (if deps changed)
- [ ] Errors wrapped with context at boundaries
- [ ] `make test` passes
- [ ] `make lint` passes
- [ ] Commit message follows Conventional Commits
