.PHONY: help lint test build clean fmt vet lint-fix

# Default target
all: fmt lint test build

## help: Show this help message
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

## lint: Run golangci-lint
lint: ## Run golangci-lint
	@echo "Running golangci-lint..."
	@go tool golangci-lint run ./...

## lint-fix: Run golangci-lint with auto-fix
lint-fix: ## Run golangci-lint with auto-fix
	@echo "Running golangci-lint with auto-fix..."
	@go tool golangci-lint run --fix ./...

## test: Run all tests
test: ## Run all tests
	@echo "Running tests..."
	@go test -v ./...

## test-coverage: Run tests with coverage
test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"
