# Makefile for checkpoint project

# Variables
BINARY_NAME := checkpoint
MODULE_NAME := github.com/dmoose/checkpoint
BIN_DIR := bin
BUILD_DIR := build
MAIN_FILE := main.go
VERSION := $(shell grep 'const version' main.go | cut -d'"' -f2)
LDFLAGS := -ldflags "-X main.version=$(VERSION) -s -w"
PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64
GO_FILES := $(shell find . -name '*.go' -not -path './vendor/*')
INSTALL_PATH := $(HOME)/.local/bin

# Default target
.PHONY: all
all: build

# Build the binary
.PHONY: build
build: $(BIN_DIR)/$(BINARY_NAME)

$(BIN_DIR)/$(BINARY_NAME): $(GO_FILES)
	@mkdir -p $(BIN_DIR)
	@echo "Building $(BINARY_NAME)..."
	go build $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_NAME) $(MAIN_FILE)

# Build for development (with race detector and debug info)
.PHONY: build-dev
build-dev:
	@mkdir -p $(BIN_DIR)
	@echo "Building $(BINARY_NAME) for development..."
	go build -race -o $(BIN_DIR)/$(BINARY_NAME) $(MAIN_FILE)

# Cross-compile for multiple platforms
.PHONY: build-all
build-all: clean-build
	@mkdir -p $(BUILD_DIR)
	@echo "Cross-compiling for multiple platforms..."
	@for platform in $(PLATFORMS); do \
		OS=$${platform%/*}; \
		ARCH=$${platform#*/}; \
		binary_name=$(BINARY_NAME); \
		if [ "$$OS" = "windows" ]; then binary_name=$(BINARY_NAME).exe; fi; \
		echo "Building for $$OS/$$ARCH..."; \
		GOOS=$$OS GOARCH=$$ARCH go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-$$OS-$$ARCH/$$binary_name $(MAIN_FILE); \
	done

# Install the binary to GOPATH/bin or GOBIN
.PHONY: install
install:
	@echo "Installing $(BINARY_NAME) to GOPATH/bin..."
	go install $(LDFLAGS) .

# Install to user's local bin directory (~/.local/bin)
.PHONY: install-user
install-user: build
	@echo "Installing $(BINARY_NAME) to $(INSTALL_PATH)..."
	@mkdir -p $(INSTALL_PATH)
	@cp $(BIN_DIR)/$(BINARY_NAME) $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "Installed to $(INSTALL_PATH)/$(BINARY_NAME)"
	@echo "Ensure $(INSTALL_PATH) is in your PATH"

# Uninstall from user's local bin directory
.PHONY: uninstall-user
uninstall-user:
	@echo "Removing $(BINARY_NAME) from $(INSTALL_PATH)..."
	@rm -f $(INSTALL_PATH)/$(BINARY_NAME)

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

# Run linter (requires golangci-lint)
.PHONY: lint
lint:
	@echo "Running linter..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not found. Install it with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; exit 1)
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

# Run the application
.PHONY: run
run: build
	@$(BIN_DIR)/$(BINARY_NAME)

# Run with arguments (usage: make run-with ARGS="check .")
.PHONY: run-with
run-with: build
	@$(BIN_DIR)/$(BINARY_NAME) $(ARGS)

# Development workflow: format, vet, test, and build
.PHONY: dev
dev: fmt vet test build

# Show help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build         - Build the binary to bin/ directory"
	@echo "  build-dev     - Build with race detector and debug info"
	@echo "  build-all     - Cross-compile for multiple platforms"
	@echo "  install       - Install binary to GOPATH/bin"
	@echo "  install-user  - Install binary to ~/.local/bin"
	@echo "  uninstall-user- Remove binary from ~/.local/bin"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  test-race     - Run tests with race detector"
	@echo "  bench         - Run benchmarks"
	@echo "  fmt           - Format code"
	@echo "  lint          - Run linter (requires golangci-lint)"
	@echo "  vet           - Run go vet"
	@echo "  check         - Run fmt, vet, lint, and test"
	@echo "  clean         - Clean all build artifacts"
	@echo "  clean-build   - Clean build directory only"
	@echo "  deps          - Download dependencies"
	@echo "  tidy          - Tidy dependencies"
	@echo "  update        - Update dependencies"
	@echo "  run           - Build and run the application"
	@echo "  run-with      - Build and run with arguments (use ARGS='...')"
	@echo "  dev           - Development workflow (fmt, vet, test, build)"
	@echo "  help          - Show this help message"

# Version info
.PHONY: version
version:
	@echo "Version: $(VERSION)"
	@echo "Module: $(MODULE_NAME)"
	@echo "Binary: $(BINARY_NAME)"

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
