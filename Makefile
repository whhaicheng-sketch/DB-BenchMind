# DB-BenchMind Makefile
# DB-BenchMind Makefile

.PHONY: build test lint check clean run help

# Variables
BINARY_NAME=db-benchmind
CMD_DIR=./cmd/db-benchmind
BUILD_DIR=./bin
GO=go
GOFLAGS=-v
LINT=golangci-lint

## build: Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)/main.go

## test: Run all tests
test:
	@echo "Running tests..."
	$(GO) test -v -race -cover ./...

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
check: format-check test lint
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
	@echo "Running $(BINARY_NAME)..."
	@echo "IMPORTANT: Must run from project root directory!"
	@if [ "$$(pwd)" != "$$(cd "$(dirname "$(realpath "$(MAKEFILE_LIST))")" && pwd)" ]; then \
		echo "ERROR: Must run from project root directory. Current: $$(pwd)"; \
		exit 1; \
	fi
	$(BUILD_DIR)/$(BINARY_NAME) gui

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
