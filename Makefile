.PHONY: build test clean install run lint schema

# Build variables
BINARY_NAME=airulesync
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_TIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-X github.com/upamune/airulesync/internal/version.Version=$(VERSION) -X github.com/upamune/airulesync/internal/version.Commit=$(COMMIT) -X github.com/upamune/airulesync/internal/version.BuildTime=$(BUILD_TIME)"

# Default target
all: clean build test

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/airulesync

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Run short tests (skips integration tests)
test-short:
	@echo "Running short tests..."
	@go test -v -short ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)
	@go clean

# Install the application
install:
	@echo "Installing $(BINARY_NAME)..."
	@go install $(LDFLAGS) ./cmd/airulesync

# Run the application
run:
	@go run $(LDFLAGS) ./cmd/airulesync $(ARGS)

# Run linters
lint:
	@echo "Running linters..."
	@golangci-lint run ./...

# Generate test coverage
coverage:
	@echo "Generating test coverage..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated at coverage.html"

# Generate JSON Schema
schema:
	@echo "Generating JSON schema..."
	@go run ./cmd/jsonschema/main.go > schema.json

# Help target
help:
	@echo "Available targets:"
	@echo "  all        - Clean, build, and test the application"
	@echo "  build      - Build the application"
	@echo "  test       - Run all tests"
	@echo "  test-short - Run short tests (skips integration tests)"
	@echo "  clean      - Clean build artifacts"
	@echo "  install    - Install the application"
	@echo "  run        - Run the application (use ARGS=\"arg1 arg2\" to pass arguments)"
	@echo "  lint       - Run linters"
	@echo "  coverage   - Generate test coverage report"
	@echo "  schema     - Generate JSON schema for configuration files"
	@echo "  help       - Show this help message"