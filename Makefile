# Makefile for checkpoint + guardrail project

# Variables
MODULE_NAME := github.com/dmoose/checkpoint
BIN_DIR := bin
BUILD_DIR := build
VERSION := 0.2.0
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE) -s -w"
PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64
GO_FILES := $(shell find . -name '*.go' -not -path './vendor/*')
INSTALL_PATH := $(HOME)/.local/bin

# Default target
.PHONY: all
all: build

# Build all binaries
.PHONY: build
build: build-checkpoint build-guardrail build-mcp

.PHONY: build-checkpoint
build-checkpoint: $(BIN_DIR)/checkpoint

$(BIN_DIR)/checkpoint: $(GO_FILES)
	@mkdir -p $(BIN_DIR)
	@echo "Building checkpoint..."
	go build $(LDFLAGS) -o $(BIN_DIR)/checkpoint ./cmd/checkpoint

.PHONY: build-guardrail
build-guardrail: $(BIN_DIR)/guardrail

$(BIN_DIR)/guardrail: $(GO_FILES)
	@mkdir -p $(BIN_DIR)
	@echo "Building guardrail..."
	go build $(LDFLAGS) -o $(BIN_DIR)/guardrail ./cmd/guardrail

.PHONY: build-mcp
build-mcp: $(BIN_DIR)/checkpoint-mcp

$(BIN_DIR)/checkpoint-mcp: $(GO_FILES)
	@mkdir -p $(BIN_DIR)
	@echo "Building checkpoint-mcp..."
	go build $(LDFLAGS) -o $(BIN_DIR)/checkpoint-mcp ./cmd/checkpoint-mcp

# Build for development (with race detector and debug info)
.PHONY: build-dev
build-dev:
	@mkdir -p $(BIN_DIR)
	@echo "Building for development..."
	go build -race -o $(BIN_DIR)/checkpoint ./cmd/checkpoint
	go build -race -o $(BIN_DIR)/guardrail ./cmd/guardrail
	go build -race -o $(BIN_DIR)/checkpoint-mcp ./cmd/checkpoint-mcp

# Cross-compile for multiple platforms
.PHONY: build-all
build-all: clean-build
	@mkdir -p $(BUILD_DIR)
	@echo "Cross-compiling for multiple platforms..."
	@for platform in $(PLATFORMS); do \
		OS=$${platform%/*}; \
		ARCH=$${platform#*/}; \
		ext=""; \
		if [ "$$OS" = "windows" ]; then ext=".exe"; fi; \
		echo "Building for $$OS/$$ARCH..."; \
		GOOS=$$OS GOARCH=$$ARCH go build $(LDFLAGS) -o $(BUILD_DIR)/checkpoint-$$OS-$$ARCH/checkpoint$$ext ./cmd/checkpoint; \
		GOOS=$$OS GOARCH=$$ARCH go build $(LDFLAGS) -o $(BUILD_DIR)/guardrail-$$OS-$$ARCH/guardrail$$ext ./cmd/guardrail; \
	done

# Install both binaries to GOPATH/bin
.PHONY: install
install:
	@echo "Installing checkpoint and guardrail to GOPATH/bin..."
	go install $(LDFLAGS) ./cmd/checkpoint
	go install $(LDFLAGS) ./cmd/guardrail
	go install $(LDFLAGS) ./cmd/checkpoint-mcp

# Install to user's local bin directory (~/.local/bin)
.PHONY: install-user
install-user: build
	@echo "Installing to $(INSTALL_PATH)..."
	@mkdir -p $(INSTALL_PATH)
	@cp $(BIN_DIR)/checkpoint $(INSTALL_PATH)/checkpoint
	@cp $(BIN_DIR)/guardrail $(INSTALL_PATH)/guardrail
	@cp $(BIN_DIR)/checkpoint-mcp $(INSTALL_PATH)/checkpoint-mcp
	@echo "Installed checkpoint, guardrail, and checkpoint-mcp to $(INSTALL_PATH)"
	@echo "Ensure $(INSTALL_PATH) is in your PATH"

# Uninstall from user's local bin directory
.PHONY: uninstall-user
uninstall-user:
	@echo "Removing binaries from $(INSTALL_PATH)..."
	@rm -f $(INSTALL_PATH)/checkpoint
	@rm -f $(INSTALL_PATH)/guardrail

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run tests with race detector
.PHONY: test-race
test-race:
	@echo "Running tests with race detector..."
	go test -race -v ./...

# Run benchmarks
.PHONY: bench
bench:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Required tool versions (keep in sync with CI)
GOLANGCI_LINT_VERSION := v2.8.0

# Run linter (requires golangci-lint v2)
.PHONY: lint
lint:
	@echo "Running linter..."
	@if ! which golangci-lint > /dev/null 2>&1; then \
		echo "golangci-lint not found. Installing $(GOLANGCI_LINT_VERSION)..."; \
		go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION); \
	fi
	@installed_version=$$(golangci-lint version 2>/dev/null | grep -oE 'v[0-9]+\.[0-9]+\.[0-9]+' | head -1); \
	if [ "$${installed_version%%.*}" != "v2" ]; then \
		echo "golangci-lint v2 required (found $$installed_version). Installing $(GOLANGCI_LINT_VERSION)..."; \
		go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION); \
	fi
	golangci-lint run ./...

# Vet code
.PHONY: vet
vet:
	@echo "Vetting code..."
	go vet ./...

# Run all quality checks
.PHONY: check
check: fmt vet lint test

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BIN_DIR)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# Clean build directory only
.PHONY: clean-build
clean-build:
	@echo "Cleaning build directory..."
	rm -rf $(BUILD_DIR)

# Download dependencies
.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	go mod download

# Tidy dependencies
.PHONY: tidy
tidy:
	@echo "Tidying dependencies..."
	go mod tidy

# Update dependencies
.PHONY: update
update:
	@echo "Updating dependencies..."
	go get -u ./...
	go mod tidy

# Run checkpoint with arguments (usage: make run-checkpoint ARGS="start")
.PHONY: run-checkpoint
run-checkpoint: build-checkpoint
	@$(BIN_DIR)/checkpoint $(ARGS)

# Run guardrail with arguments (usage: make run-guardrail ARGS="explain")
.PHONY: run-guardrail
run-guardrail: build-guardrail
	@$(BIN_DIR)/guardrail $(ARGS)

# Development workflow: format, vet, test, and build
.PHONY: dev
dev: fmt vet test build

# Show help
.PHONY: help
help:
	@echo "Available targets:"
	@echo ""
	@echo "Build:"
	@echo "  build              - Build both checkpoint and guardrail binaries"
	@echo "  build-checkpoint   - Build only checkpoint binary"
	@echo "  build-guardrail    - Build only guardrail binary"
	@echo "  build-dev          - Build both with race detector"
	@echo "  build-all          - Cross-compile for multiple platforms"
	@echo ""
	@echo "Install:"
	@echo "  install            - Install both binaries to GOPATH/bin"
	@echo "  install-user       - Install both to ~/.local/bin"
	@echo "  uninstall-user     - Remove both from ~/.local/bin"
	@echo ""
	@echo "Test & Quality:"
	@echo "  test               - Run tests"
	@echo "  test-coverage      - Run tests with coverage report"
	@echo "  test-race          - Run tests with race detector"
	@echo "  bench              - Run benchmarks"
	@echo "  fmt                - Format code"
	@echo "  lint               - Run linter"
	@echo "  vet                - Run go vet"
	@echo "  check              - Run fmt, vet, lint, and test"
	@echo ""
	@echo "Run:"
	@echo "  run-checkpoint     - Build and run checkpoint (use ARGS='...')"
	@echo "  run-guardrail      - Build and run guardrail (use ARGS='...')"
	@echo ""
	@echo "Maintenance:"
	@echo "  clean              - Clean all build artifacts"
	@echo "  deps               - Download dependencies"
	@echo "  tidy               - Tidy dependencies"
	@echo "  update             - Update dependencies"
	@echo "  dev                - Development workflow (fmt, vet, test, build)"

# Version info
.PHONY: version
version:
	@echo "Version: $(VERSION)"
	@echo "Commit:  $(COMMIT)"
	@echo "Date:    $(DATE)"
	@echo "Module:  $(MODULE_NAME)"

# Create release archives (requires build-all)
.PHONY: release
release: build-all
	@echo "Creating release archives..."
	@cd $(BUILD_DIR) && for dir in */; do \
		if [ -d "$$dir" ]; then \
			tar -czf "$${dir%/}.tar.gz" -C "$$dir" .; \
		fi \
	done
	@echo "Release archives created in $(BUILD_DIR)/"
