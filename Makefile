# DB-BenchMind Makefile
# DB-BenchMind Makefile

.PHONY: build test test-backend test-frontend test-smoke regression release-gate lint check clean run help

# Variables
BINARY_NAME=db-benchmind
CMD_DIR=./cmd/db-benchmind
BUILD_DIR=./bin
GO=go
GOFLAGS=-v
LINT=golangci-lint

## build: Build the application
build:
	./scripts/entry/build

## test: Run all tests
test:
	$(MAKE) test-backend
	$(MAKE) test-frontend

## test-backend: Run backend release-gate test suite
test-backend:
	bash scripts/gates/test_backend_gate.sh

## test-frontend: Run frontend release-gate test suite
test-frontend:
	bash scripts/gates/test_frontend_gate.sh

## test-smoke: Run build/start smoke gate
test-smoke:
	bash scripts/gates/test_smoke_gate.sh

## regression: Run the automated regression gate used for validation
regression:
	./scripts/entry/regression

## release-gate: Mandatory release blocker suite
release-gate: regression

## test-unit: Run unit tests only (no integration)
test-unit:
	@echo "Running unit tests..."
	$(GO) test -v -short ./...

## test-integration: Run integration tests
test-integration:
	@echo "Running integration tests..."
	$(GO) test -v -tags=integration ./...

## lint: Run golangci-lint
lint:
	@echo "Running linter..."
	$(LINT) run

## check: Run all checks (format, test, lint)
check: format-check regression lint
	@echo "All checks passed!"

## format: Format code with gofmt and goimports
format:
	@echo "Formatting code..."
	gofmt -w -s .
	goimports -w .

## format-check: Check if code is formatted
format-check:
	@echo "Checking code format..."
	@test -z "$$(gofmt -l . | grep -v '^vendor' | grep -v '^testdata')" || \
		(echo "Code is not formatted. Please run 'make format'" && exit 1)

## clean: Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR)
	rm -rf results/*.log

## run: Build and run the application
run: build
	./scripts/entry/start

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GO) mod download
	$(GO) mod tidy

## vulncheck: Check for security vulnerabilities
vulncheck:
	@echo "Checking for vulnerabilities..."
	govulncheck ./...

## help: Show this help message
help:
	@echo "DB-BenchMind Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /'
