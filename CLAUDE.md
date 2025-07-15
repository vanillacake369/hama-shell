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

## Module Architecture

### Core Modules

#### 1. Configuration Module (`pkg/config/`)
```go
// config.go - YAML parsing and validation
// alias.go - Global alias registry and resolution
// validation.go - Configuration validation
```

#### 2. Session Module (`pkg/session/`)
```go
// session.go - Session data structures and lifecycle
// manager.go - Session management and orchestration
// process.go - SSH process management
// state.go - Session state persistence
```

#### 3. CLI Module (`pkg/cli/`)
```go
// commands.go - CLI command definitions
// start.go - Start command implementation
// status.go - Status command implementation
// alias.go - Alias management commands
// interactive.go - Interactive mode
```

#### 4. Network Module (`pkg/network/`)
```go
// ssh.go - SSH connection handling
// tunnel.go - Tunneling and port forwarding
// monitor.go - Connection health monitoring
```

#### 5. Utils Module (`pkg/utils/`)
```go
// env.go - Environment variable handling
// log.go - Logging utilities
// path.go - Path resolution utilities
```

### Project Structure
```
hama-shell/
├── main.go                 # Entry point
├── go.mod                  # Module definition
├── pkg/
│   ├── config/            # Configuration management
│   ├── session/           # Session management
│   ├── cli/               # CLI commands
│   ├── network/           # Network operations
│   └── utils/             # Utilities
├── cmd/                   # Command implementations
├── internal/              # Internal packages
├── examples/              # Example configurations
├── docs/                  # Documentation
├── tests/                 # Test files
├── .github/
│   └── workflows/         # CI/CD workflows
├── scripts/               # Build and installation scripts
└── .goreleaser.yml        # Release configuration
```

### Current State
- Basic Go module initialized with Go 1.24
- Single `main.go` file with placeholder code
- Ready for modular architecture implementation