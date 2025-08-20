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
	@echo ''
	@echo 'SSH Testing targets:'
	@echo '  ssh-test-all  Complete SSH test environment setup with Zellij'
	@echo '  ssh-setup     Initialize SSH test environment'
	@echo '  ssh-server    Start SSH server on port 2222'
	@echo '  ssh-client    Connect to test SSH server'
	@echo '  ssh-status    Check SSH server status'
	@echo '  ssh-clean     Clean up SSH test environment'

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

# SSH Testing Environment Variables
SSH_TEST_DIR := .ssh-test
SSH_TEST_PORT := 2222
SSH_TEST_USER := $(shell whoami)
SSH_TEST_HOST := 127.0.0.1
ZELLIJ_SESSION := hama-ssh-test

## ssh-test-all: Complete SSH test environment setup with Zellij
.PHONY: ssh-test-all
ssh-test-all: ssh-clean ssh-setup
	@echo "${BLUE}Starting complete SSH test environment...${NC}"
	@$(MAKE) ssh-server-bg
	@sleep 2
	@echo "${GREEN}✓ SSH test environment ready${NC}"
	@echo "${BLUE}Launching Zellij with SSH test panes...${NC}"
	@$(MAKE) ssh-test-zellij

## ssh-setup: Initialize SSH test environment
.PHONY: ssh-setup
ssh-setup:
	@echo "${BLUE}Setting up SSH test environment...${NC}"
	@mkdir -p "$(SSH_TEST_DIR)/sshd" "$(SSH_TEST_DIR)/client"
	@chmod 700 "$(SSH_TEST_DIR)/client"
	@echo "Generating SSH host keys..."
	@ssh-keygen -t ed25519 -f "$(SSH_TEST_DIR)/sshd/ssh_host_ed25519_key" -N "" -q
	@ssh-keygen -t rsa -b 4096 -f "$(SSH_TEST_DIR)/sshd/ssh_host_rsa_key" -N "" -q
	@echo "Generating client key..."
	@[ -f "$(SSH_TEST_DIR)/client/id_ed25519" ] || ssh-keygen -t ed25519 -f "$(SSH_TEST_DIR)/client/id_ed25519" -N "" -q
	@echo "Setting up authorized_keys..."
	@cat "$(SSH_TEST_DIR)/client/id_ed25519.pub" > "$(SSH_TEST_DIR)/client/authorized_keys"
	@chmod 600 "$(SSH_TEST_DIR)/client/authorized_keys"
	@echo "Creating sshd_config..."
	@cat > "$(SSH_TEST_DIR)/sshd/sshd_config" <<< 'Port $(SSH_TEST_PORT)\nListenAddress $(SSH_TEST_HOST)\nProtocol 2\nPidFile $(PWD)/$(SSH_TEST_DIR)/sshd/sshd.pid\nHostKey $(PWD)/$(SSH_TEST_DIR)/sshd/ssh_host_ed25519_key\nHostKey $(PWD)/$(SSH_TEST_DIR)/sshd/ssh_host_rsa_key\nPasswordAuthentication no\nPubkeyAuthentication yes\nAuthorizedKeysFile $(PWD)/$(SSH_TEST_DIR)/client/authorized_keys\nUsePAM no\nChallengeResponseAuthentication no\nPermitRootLogin no\nLogLevel VERBOSE'
	@echo "${GREEN}✓ SSH test environment setup complete${NC}"

## ssh-server-bg: Start SSH server in background
.PHONY: ssh-server-bg
ssh-server-bg: ssh-setup
	@echo "${BLUE}Starting SSH server on port $(SSH_TEST_PORT)...${NC}"
	@if [ -f "$(SSH_TEST_DIR)/sshd/sshd.pid" ] && kill -0 `cat "$(SSH_TEST_DIR)/sshd/sshd.pid"` 2>/dev/null; then \
		echo "SSH server already running"; \
	else \
		$(shell which sshd) -f "$(PWD)/$(SSH_TEST_DIR)/sshd/sshd_config"; \
		echo "${GREEN}✓ SSH server started${NC}"; \
	fi

## ssh-server: Start SSH server in foreground (for monitoring)
.PHONY: ssh-server
ssh-server: ssh-setup
	@echo "${BLUE}Starting SSH server on port $(SSH_TEST_PORT) (foreground)...${NC}"
	@$(shell which sshd) -f "$(PWD)/$(SSH_TEST_DIR)/sshd/sshd_config" -D

## ssh-client: Connect to test SSH server
.PHONY: ssh-client
ssh-client:
	@echo "${BLUE}Connecting to SSH test server...${NC}"
	@ssh-keygen -R "[$(SSH_TEST_HOST)]:$(SSH_TEST_PORT)" 2>/dev/null || true
	@ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -i "$(SSH_TEST_DIR)/client/id_ed25519" -p $(SSH_TEST_PORT) $(SSH_TEST_USER)@$(SSH_TEST_HOST)

## ssh-test-zellij: Launch Zellij with SSH test panes
.PHONY: ssh-test-zellij
ssh-test-zellij:
	@echo "${BLUE}Launching Zellij session: $(ZELLIJ_SESSION)${NC}"
	@zellij kill-session $(ZELLIJ_SESSION) 2>/dev/null || true
	@zellij --session $(ZELLIJ_SESSION) --layout - <<< 'layout {\n  pane size=1 borderless=true {\n    plugin location="zellij:compact-bar"\n  }\n  pane split_direction="vertical" {\n    pane {\n      name "SSH Server"\n      command "make"\n      args "ssh-server"\n    }\n    pane {\n      name "SSH Client"\n      command "bash"\n      args "-c" "echo '\''SSH Client Ready. Run: make ssh-client'\'' && bash"\n    }\n  }\n}'

## ssh-status: Check SSH server status
.PHONY: ssh-status
ssh-status:
	@echo "${BLUE}Checking SSH server status...${NC}"
	@if [ -f "$(SSH_TEST_DIR)/sshd/sshd.pid" ]; then \
		if kill -0 `cat "$(SSH_TEST_DIR)/sshd/sshd.pid"` 2>/dev/null; then \
			echo "${GREEN}✓ SSH server is running (PID: `cat "$(SSH_TEST_DIR)/sshd/sshd.pid"`)${NC}"; \
			echo "Port $(SSH_TEST_PORT) status:"; \
			netstat -ln | grep ":$(SSH_TEST_PORT)" || echo "Port not found in netstat"; \
		else \
			echo "SSH server PID file exists but process is not running"; \
		fi; \
	else \
		echo "SSH server is not running"; \
	fi

## ssh-clean: Clean up SSH test environment
.PHONY: ssh-clean
ssh-clean:
	@echo "${BLUE}Cleaning SSH test environment...${NC}"
	@if [ -f "$(SSH_TEST_DIR)/sshd/sshd.pid" ]; then \
		if kill -0 `cat "$(SSH_TEST_DIR)/sshd/sshd.pid"` 2>/dev/null; then \
			kill `cat "$(SSH_TEST_DIR)/sshd/sshd.pid"` 2>/dev/null || true; \
			echo "Stopped SSH server"; \
		fi; \
	fi
	@rm -rf "$(SSH_TEST_DIR)" 2>/dev/null || true
	@ssh-keygen -R "[$(SSH_TEST_HOST)]:$(SSH_TEST_PORT)" 2>/dev/null || true
	@zellij kill-session $(ZELLIJ_SESSION) 2>/dev/null || true
	@echo "${GREEN}✓ SSH test environment cleaned${NC}"