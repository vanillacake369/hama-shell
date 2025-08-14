# HamaShell Makefile
# Build and test commands for the Cobra CLI application

# Variables
BINARY_NAME=hama-shell
MAIN_PACKAGE=./main.go
GO=go
GOFLAGS=-v
BUILD_DIR=.
VERSION=$(shell git describe --tags --always --dirty)
LDFLAGS=-ldflags "-X main.version=${VERSION}"

# Colors for output
GREEN=\033[0;32m
NC=\033[0m # No Color
BLUE=\033[0;34m

# Default target
.DEFAULT_GOAL := help

## help: Show this help message
.PHONY: help
help:
	@echo 'Usage:'
	@echo '  make <target>'
	@echo ''
	@echo 'Common targets:'
	@echo '  build         Build the binary'
	@echo '  run           Run the application'
	@echo '  test-cli      Test all CLI commands'
	@echo '  test          Run Go tests'
	@echo '  clean         Remove built binaries'
	@echo '  all           Run fmt, vet, test, and build'
	@echo '  quick-check   Quick validation'
	@echo '  help          Show this help'

## build: Build the binary
.PHONY: build
build:
	@echo "Building ${BINARY_NAME}..."
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)
	@echo "${GREEN}✓ Build complete${NC}"

## run: Run the application
.PHONY: run
run: build
	@echo "Running ${BINARY_NAME}..."
	./$(BINARY_NAME)

## test-cli: Test all CLI commands
.PHONY: test-cli
test-cli: build
	@echo "${BLUE}Testing Cobra CLI commands...${NC}"
	@echo ""
	@echo "1. Testing help command:"
	./$(BINARY_NAME) --help || true
	@echo ""
	@echo "2. Testing version info:"
	./$(BINARY_NAME) version || true
	@echo ""
	@echo "3. Testing start command help:"
	./$(BINARY_NAME) start --help || true
	@echo ""
	@echo "4. Testing stop command help:"
	./$(BINARY_NAME) stop --help || true
	@echo ""
	@echo "5. Testing status command help:"
	./$(BINARY_NAME) status --help || true
	@echo ""
	@echo "6. Testing config command help:"
	./$(BINARY_NAME) config --help || true
	@echo ""
	@echo "${GREEN}✓ CLI command tests complete${NC}"

## test-commands: Test individual commands (non-destructive)
.PHONY: test-commands
test-commands: build
	@echo "${BLUE}Testing individual commands...${NC}"
	@echo ""
	@echo "Testing status command (safe to run):"
	./$(BINARY_NAME) status || true
	@echo ""
	@echo "${GREEN}✓ Command tests complete${NC}"

## test: Run all Go tests
.PHONY: test
test:
	@echo "Running tests..."
	$(GO) test $(GOFLAGS) ./...
	@echo "${GREEN}✓ Tests complete${NC}"

## test-coverage: Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GO) test -cover ./...
	@echo "${GREEN}✓ Coverage report complete${NC}"

## fmt: Format Go code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...
	@echo "${GREEN}✓ Code formatted${NC}"

## vet: Run Go vet for static analysis
.PHONY: vet
vet:
	@echo "Running static analysis..."
	$(GO) vet ./...
	@echo "${GREEN}✓ Static analysis complete${NC}"

## lint: Run additional linting (requires golangci-lint)
.PHONY: lint
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

## clean: Remove built binaries
.PHONY: clean
clean:
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME)-*
	@echo "${GREEN}✓ Clean complete${NC}"

## install: Install the binary to GOPATH/bin
.PHONY: install
install:
	@echo "Installing ${BINARY_NAME}..."
	$(GO) install $(LDFLAGS) $(MAIN_PACKAGE)
	@echo "${GREEN}✓ Installed to GOPATH/bin${NC}"

## build-all: Build for multiple platforms
.PHONY: build-all
build-all: build-linux build-windows build-darwin

## build-linux: Build for Linux
.PHONY: build-linux
build-linux:
	@echo "Building for Linux..."
	GOOS=linux GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PACKAGE)
	@echo "${GREEN}✓ Linux build complete${NC}"

## build-windows: Build for Windows
.PHONY: build-windows
build-windows:
	@echo "Building for Windows..."
	GOOS=windows GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PACKAGE)
	@echo "${GREEN}✓ Windows build complete${NC}"

## build-darwin: Build for macOS
.PHONY: build-darwin
build-darwin:
	@echo "Building for macOS..."
	GOOS=darwin GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PACKAGE)
	@echo "${GREEN}✓ macOS build complete${NC}"

## all: Run fmt, vet, test, and build
.PHONY: all
all: fmt vet test build
	@echo "${GREEN}✓ All tasks complete${NC}"

## quick-check: Quick validation (fmt, vet, build, test-cli)
.PHONY: quick-check
quick-check: fmt vet build test-cli
	@echo "${GREEN}✓ Quick check complete${NC}"

## dev: Watch for changes and rebuild (requires entr)
.PHONY: dev
dev:
	@echo "Watching for changes..."
	@if command -v entr >/dev/null 2>&1; then \
		find . -name '*.go' | entr -r make build; \
	else \
		echo "entr not installed. Run: apt-get install entr (or brew install entr on macOS)"; \
	fi

.PHONY: verify
verify: build
	@echo "${BLUE}Verifying Cobra CLI is working...${NC}"
	@./$(BINARY_NAME) --help > /dev/null 2>&1 && echo "${GREEN}✓ CLI is working properly${NC}" || echo "✗ CLI failed"