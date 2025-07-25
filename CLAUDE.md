# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

HamaShell is a **session and connection manager** designed for developers who need reliable, secure access to various hosts in single CLI command. It simplifies complex multi-step SSH tunneling and session setup by letting developers define their connections declaratively in a YAML file.

### Key Benefits
- **Declarative & reproducible** — Define connections once in YAML and reuse them easily
- **Secure by design** — Uses system environment variables to keep secrets hidden and safe
- **Full process control** — Start, stop, check status, and manage connections interactively
- **Alias & hierarchy support** — Organize connections by project, stage, and developer
- **Parallel execution** — Run multiple sessions simultaneously without manual orchestration
- **Multi-cloud ready** — Works seamlessly with AWS, Oracle Cloud, Naver Cloud, and on-premise
- **Portable** — Runs on any Linux distro, integrates into CI/CD pipelines, and supports local dashboards for visibility

## Core Features

### 1. YAML-based Configuration
- Define complex multi-step tunneling and SSH workflows declaratively
- Supports dynamic command steps and environment variable substitution
- Hierarchical structure: project > stage > developer > session
- Global aliases for quick access to session paths

### 2. Secure & Flexible Connections
- SSH with key-based authentication and multi-hop tunneling
- Leverages environment variables to avoid hardcoding secrets
- Multi-cloud support (AWS, Oracle Cloud, Naver Cloud, Homelab)
- Isolated user configurations

### 3. Powerful Connection Management
- Persistent and recoverable sessions
- Process control: start, stop, view status, and monitor connections
- Port forwarding and connection health checks with automatic retry
- Parallel execution support

### 4. Developer-friendly CLI
- Interactive commands for managing sessions
- Clear status reporting and logs for troubleshooting
- Alias management and resolution
- Dashboard/TUI mode for visual monitoring

### 5. Configuration Processing
- YAML validation and parsing
- Dynamic command substitution
- Environment variable resolution
- Alias path resolution

## Development Methodology

### Test-Driven Development (TDD) Approach

This project follows a strict TDD methodology to ensure code quality, maintainability, and reliability. **ALL new features must follow this process:**

#### 1. Test/Implementation Planning Phase
Before implementing any feature, create a detailed plan that includes:

**Test Plan Structure:**
```
Feature: [Feature Name]
Description: [Brief description of what the feature does]

Test Cases:
1. [Test Case 1] - [Expected behavior]
2. [Test Case 2] - [Edge case or error condition]
3. [Test Case 3] - [Integration test]

Implementation Plan:
1. [Step 1] - [What needs to be implemented]
2. [Step 2] - [Dependencies or prerequisites]
3. [Step 3] - [Integration points]

Acceptance Criteria:
- [ ] All unit tests pass
- [ ] Integration tests pass
- [ ] Code coverage > 80%
- [ ] No linting errors
- [ ] Documentation updated
```

#### 2. Review and Approval Process
- **MANDATORY:** Present the test/implementation plan to the user for review
- Wait for explicit approval before proceeding with implementation
- Address any feedback or concerns raised during review
- Only proceed to implementation after receiving approval

#### 3. TDD Implementation Cycle
Once approved, follow the Red-Green-Refactor cycle:

**Red Phase:** Write failing tests first
- Create comprehensive test cases that cover the planned functionality
- Ensure tests fail initially (proving they're testing the right thing)
- Include edge cases, error conditions, and boundary tests

**Green Phase:** Implement minimal code to pass tests
- Write the simplest code possible to make tests pass
- Focus on functionality, not optimization
- Ensure all tests pass before moving to refactor

**Refactor Phase:** Improve code quality
- Optimize performance and readability
- Ensure code follows project conventions
- Maintain passing tests throughout refactoring

#### 4. Quality Gates
Before considering a feature complete:
- [ ] **Unit Tests:** All tests pass with >80% code coverage
- [ ] **Integration Tests:** Feature works with existing systems
- [ ] **Code Quality:** Pass `go fmt`, `go vet`, and any linters
- [ ] **Documentation:** Update relevant documentation
- [ ] **Performance:** No significant performance regression

#### 5. Example TDD Activities

**Feature Planning Activities:**
- Write user stories and acceptance criteria
- Design API interfaces and data structures
- Plan test scenarios including happy path and edge cases
- Identify integration points and dependencies
- Create mock objects and test fixtures

**Test Implementation Activities:**
- Write unit tests for individual functions and methods
- Create integration tests for component interactions
- Design table-driven tests for multiple scenarios
- Implement benchmarks for performance-critical code
- Set up test fixtures and helper functions

**Implementation Activities:**
- Implement minimal code to pass each test
- Follow Go best practices and project conventions
- Handle errors gracefully with proper error messages
- Add appropriate logging and monitoring
- Ensure thread safety where needed

**Refactoring Activities:**
- Optimize algorithms and data structures
- Improve code readability and maintainability
- Extract common functionality into reusable components
- Update documentation and comments
- Verify performance benchmarks

#### 6. Testing Strategy

**Unit Tests:**
- Test individual functions and methods in isolation
- Use mocks for external dependencies
- Cover all public APIs and critical internal functions
- Include table-driven tests for multiple scenarios

**Integration Tests:**
- Test component interactions
- Verify end-to-end workflows
- Test with real (but controlled) external dependencies
- Include CLI command integration tests

**Performance Tests:**
- Benchmark critical paths and algorithms
- Test with realistic data sizes
- Monitor memory usage and goroutine leaks
- Set performance regression thresholds

## Development Commands

### Building and Running

- `go run main.go` - Compile and run the main program
- `go build` - Compile the program into an executable
- `go build -o hama-shell` - Build with custom executable name

### Testing and Code Quality

- `go test ./...` - Run all tests in the project
- `go test -v ./...` - Run tests with verbose output
- `go test -cover ./...` - Run tests with coverage report
- `go test -bench=.` - Run benchmark tests
- `go fmt ./...` - Format Go source code
- `go vet ./...` - Report likely mistakes in packages
- `go mod tidy` - Clean up module dependencies

### Module Management

- `go mod init` - Initialize module (already done)
- `go mod download` - Download module dependencies
- `go get <package>` - Add new dependencies

## Component-Based Architecture

HamaShell is designed with a clean, component-based architecture that promotes flexibility, maintainability, and cross-platform compatibility. The architecture centers around four core component groups with clear interfaces and responsibilities.

### Core Component Groups

#### 1. Session Management (`internal/core/session/`)
- **Session Manager** (`manager.go`) - Session lifecycle management and orchestration
- **Session State** (`state.go`) - In-memory session state management
- **Session Persistence** (`persistence.go`) - File-based session persistence and recovery

#### 2. Connection Management (`internal/core/connection/`)
- **Connection Manager** (`manager.go`) - Connection lifecycle and management
- **SSH Client** (`ssh.go`) - SSH connection handling and authentication
- **Tunnel Manager** (`tunnel.go`) - Port forwarding and tunnel management
- **Health Monitor** (`monitor.go`) - Connection health monitoring and auto-recovery

#### 3. Configuration (`internal/core/config/`)
- **Config Loader** (`loader.go`) - YAML configuration loading and parsing
- **Config Validator** (`validator.go`) - Configuration validation and schema checking
- **Alias Manager** (`alias.go`) - Global alias registry and resolution

#### 4. Terminal Integration (`internal/core/terminal/`)
- **Terminal Interface** (`interface.go`) - Terminal session management
- **Multiplexer Integration** (`multiplexer.go`) - Tmux/Zellij/Screen integration
- **Shell Integration** (`shell.go`) - Shell command execution and completion

### Service Layer (`internal/service/`)
- **Session Service** (`session_service.go`) - Session management business logic
- **Config Service** (`config_service.go`) - Configuration management operations
- **Connection Service** (`connection_service.go`) - Connection management business logic
- **Terminal Service** (`terminal_service.go`) - Terminal integration operations

### CLI Layer (`cmd/`)
- **Root Command** (`root.go`) - Main CLI entry point with configuration
- **Start Command** (`start.go`) - Session start operations
- **Stop Command** (`stop.go`) - Session stop operations
- **Status Command** (`status.go`) - Session status monitoring
- **Config Command** (`config.go`) - Configuration management
- **Alias Command** (`alias.go`) - Alias management
- **Interactive Command** (`interactive.go`) - TUI mode

### Type Definitions (`pkg/types/`)
- **Session Types** (`session.go`) - Session interfaces and data structures
- **Config Types** (`config.go`) - Configuration interfaces and data structures
- **Connection Types** (`connection.go`) - Connection interfaces and data structures
- **Terminal Types** (`terminal.go`) - Terminal integration interfaces

### Infrastructure Layer (`internal/infrastructure/`)
- **Storage** (`storage/`) - File system and state storage abstractions
- **Network** (`network/`) - Network client and port forwarding implementations
- **Process** (`process/`) - Process control and execution management
- **Platform** (`platform/`) - OS-specific abstractions and implementations

### Integration Layer (`pkg/integration/`)
- **Tmux Integration** (`tmux.go`) - Tmux-specific multiplexer implementation
- **Zellij Integration** (`zellij.go`) - Zellij-specific multiplexer implementation
- **Shell Integration** (`shell.go`) - Shell-specific integrations

### Project Structure
```
hama-shell/
├── main.go                 # Entry point
├── go.mod                  # Module definition
├── cmd/                    # CLI command implementations
│   ├── root.go
│   ├── start.go
│   ├── stop.go
│   ├── status.go
│   ├── config.go
│   ├── alias.go
│   └── interactive.go
├── internal/               # Internal packages
│   ├── service/           # Service layer business logic
│   ├── core/              # Core component implementations
│   │   ├── session/       # Session management
│   │   ├── connection/    # Connection management
│   │   ├── config/        # Configuration management
│   │   └── terminal/      # Terminal integration
│   └── infrastructure/    # Infrastructure layer
│       ├── storage/       # File system abstractions
│       ├── network/       # Network implementations
│       ├── process/       # Process management
│       └── platform/      # OS abstractions
├── pkg/                   # Public packages
│   ├── types/            # Type definitions and interfaces
│   └── integration/      # External integrations
├── scripts/              # Build and deployment scripts
│   ├── completion/       # Shell completion scripts
│   └── multiplexer/      # Multiplexer integration scripts
├── examples/             # Example configurations
├── docs/                 # Documentation
└── .github/
    └── workflows/        # CI/CD workflows
```

### Current Implementation State
- ✅ **Directory Structure**: Complete component-based structure created
- ✅ **Type Definitions**: All core interfaces and types implemented
- ✅ **CLI Layer**: All command structures with Cobra integration
- ✅ **Service Layer**: Business logic services implemented
- ✅ **Core Components**: Session management and config loader implemented
- 🚧 **Infrastructure Layer**: Ready for platform-specific implementations
- 🚧 **Integration Layer**: Ready for multiplexer and shell integrations

### Architecture Benefits Achieved
- **Component-Based Design**: Clear separation of concerns with focused responsibilities
- **Interface-Driven**: All components use well-defined interfaces for testability
- **Layered Architecture**: Clean dependency flow from CLI → Service → Core → Infrastructure
- **Cross-Platform Ready**: Infrastructure layer prepared for platform-specific implementations
- **Testable**: Interface-driven design enables comprehensive unit and integration testing