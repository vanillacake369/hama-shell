# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

HamaShell is a **session and connection manager** designed for developers who need reliable, secure access to various hosts in single CLI command. It simplifies complex multi-step SSH tunneling and session setup by letting developers define their connections declaratively in a YAML file.

### Key Benefits
- **Declarative & reproducible** — Define connections once in YAML and reuse them easily
- **Secure by design** — Uses system environment variables to keep secrets hidden and safe
- **Full process control** — Start, stop, check status, and manage connections interactively
- **Hierarchical organization** — Organize connections using project.stage.service pattern
- **Parallel execution** — Run multiple sessions simultaneously with process isolation
- **Cross-platform ready** — Works on Unix/Linux and Windows systems
- **Simple & focused** — Clean, minimal implementation that's easy to understand and extend

## Current Implementation Status

HamaShell is currently in active development with a focus on core functionality. The implementation prioritizes simplicity, reliability, and ease of use over complex abstractions.

### ✅ Implemented Components

#### 1. Process Executor (`internal/core/executor/`)
A lightweight, thread-safe process executor with hierarchical key-based management:
- **Simple API**: `Run()`, `StopAll()`, `StopByKey()`, and `GetStatus()`
- **Thread-safe operations**: Uses `sync.Map` for concurrent access
- **Hierarchical keys**: Organizes processes by project.stage.service pattern
- **Platform-aware**: Proper signal handling for Unix/Linux (SIGTERM/SIGKILL) and Windows
- **Process isolation**: Uses process groups for better management

#### 2. Configuration System (`internal/core/config/`)
YAML-based configuration with validation:
- **Config Validator**: Parses and validates YAML configuration files
- **Type-safe structures**: Well-defined Go structs for configuration
- **Environment variable support**: Integration with system environment
- **Error handling**: Clear error messages for configuration issues

#### 3. CLI Framework (`cmd/`)
Cobra-based command-line interface:
- **Root Command**: Base command with configuration loading
- **Start Command**: Session start operations
- **Stop Command**: Session stop operations  
- **Status Command**: Session status monitoring
- **Config Command**: Configuration management

### 🚧 In Progress
- Session state management and persistence
- Enhanced error handling and recovery
- Configuration file generation and management

### 📋 Planned Features
- SSH connection management
- Port forwarding and tunneling
- Terminal multiplexer integration
- Interactive TUI mode
- Shell completion scripts

## Project Structure

```
hama-shell/
├── main.go                    # Application entry point
├── go.mod                     # Go module definition  
├── cmd/                       # CLI command implementations
│   ├── root.go               # Root command with config loading
│   ├── start.go              # Start command
│   ├── stop.go               # Stop command
│   ├── status.go             # Status command
│   └── config.go             # Config command
├── internal/                  # Internal packages
│   └── core/                 # Core components
│       ├── executor/         # Process execution management
│       │   ├── executor.go           # Main executor implementation
│       │   ├── process_common.go     # Shared types and interfaces
│       │   ├── process_unix.go       # Unix-specific process handling
│       │   ├── process_windows.go    # Windows-specific process handling
│       │   └── README.md             # Executor documentation
│       └── config/           # Configuration management
│           ├── validator.go          # Config parsing and validation
│           └── validator_test.go     # Config validation tests
├── docs/                     # Documentation
└── example.yaml              # Example configuration file
```

## Configuration Format

HamaShell uses a simple YAML-based configuration format:

```yaml
projects:
  myapp:
    description: "Main application project"
    stages:
      dev:
        services:
          database:
            description: "Development database connection"
            commands:
              - "ssh -i ${SSH_KEY_PATH} ${DB_USER}@dev-db.example.com"
              - "mysql -u root -p${DB_PASSWORD}"
          api:
            description: "Development API server"
            commands:
              - "ssh -i ${SSH_KEY_PATH} ${API_USER}@dev-api.example.com"
              - "cd /app && npm start"

global_settings:
  timeout: 30
  retries: 3  
  auto_restart: true
```

## Development Methodology

### Test-Driven Development (TDD) Approach

This project follows TDD methodology to ensure code quality and maintainability:

#### 1. Red-Green-Refactor Cycle
- **Red**: Write failing tests first
- **Green**: Write minimal code to pass tests  
- **Refactor**: Improve code quality while maintaining tests

#### 2. Quality Gates
Before considering a feature complete:
- [ ] Unit tests pass with >80% coverage
- [ ] Integration tests pass  
- [ ] Code quality passes (`go fmt`, `go vet`)
- [ ] Documentation updated
- [ ] No performance regression

#### 3. Testing Strategy
- **Unit Tests**: Test individual functions and methods in isolation
- **Integration Tests**: Test component interactions
- **Table-Driven Tests**: Multiple scenarios in single test functions
- **Mock External Dependencies**: Use interfaces for testability

## Development Commands

### Building and Running
```bash
go run main.go              # Run the application
go build                    # Build executable
go build -o hama-shell      # Build with custom name
```

### Testing and Code Quality  
```bash
go test ./...               # Run all tests
go test -v ./...            # Verbose test output
go test -cover ./...        # Test coverage report
go test -bench=.            # Run benchmarks
go fmt ./...                # Format source code
go vet ./...                # Static analysis
go mod tidy                 # Clean dependencies
```

### Module Management
```bash
go mod download             # Download dependencies
go get <package>            # Add new dependency
```

## Architecture Principles

### 1. Simplicity First
- Minimal abstractions and interfaces
- Clear, readable code over clever optimizations
- Direct implementation over complex patterns

### 2. Platform Awareness
- Proper handling of OS differences (Unix vs Windows)
- Native process management per platform
- Cross-platform file and path handling

### 3. Thread Safety
- Safe concurrent access to shared resources
- Use of Go's built-in synchronization primitives
- No data races or deadlock conditions

### 4. Error Handling
- Clear, actionable error messages
- Graceful degradation where possible
- Proper resource cleanup on errors

### 5. Testability
- Interface-driven design where beneficial
- Dependency injection for external resources
- Comprehensive test coverage

## Current Focus Areas

### 1. Core Stability
- Robust process management
- Reliable configuration parsing
- Error handling and recovery

### 2. User Experience  
- Clear command-line interface
- Helpful error messages
- Intuitive configuration format

### 3. Cross-Platform Support
- Consistent behavior across operating systems
- Platform-specific optimizations
- Native integration patterns

## Future Roadmap

### Phase 1: Core Functionality (Current)
- ✅ Process executor with hierarchical management
- ✅ Basic CLI structure
- ✅ Configuration validation
- 🚧 Session persistence
- 🚧 Enhanced error handling

### Phase 2: Connection Management  
- SSH client implementation
- Port forwarding and tunneling
- Connection health monitoring
- Automatic retry logic

### Phase 3: Advanced Features
- Terminal multiplexer integration (tmux, zellij)
- Interactive TUI mode
- Shell completion scripts
- Configuration generation tools

### Phase 4: Polish & Distribution
- Comprehensive documentation
- Installation packages
- CI/CD pipeline
- Performance optimization

This roadmap ensures steady progress toward the full vision while maintaining a stable, usable tool at each phase.