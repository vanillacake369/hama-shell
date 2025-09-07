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

#### 1. Clean Architecture Foundation (`internal/core/`)
Domain-driven design with clear separation of concerns:
- **Domain separation**: Configuration, Service, Session, and Terminal domains
- **Layer separation**: API-Infrastructure-Model pattern per domain
- **Dependency inversion**: Interfaces define contracts between layers
- **Single responsibility**: Each domain handles specific business concerns

#### 2. Configuration Domain (`internal/core/configuration/`)
Complete configuration management system:
- **API Layer**: Configuration operation interfaces for dependency inversion
- **Infrastructure Layer**: Viper-based configuration management with file handling
- **Model Layer**: Type-safe configuration structures with validation
- **Project-Service-Stage hierarchy**: Flexible configuration organization

#### 3. Service Domain (`internal/core/service/`)
Service definition and management:
- **API Layer**: Service management interfaces
- **Infrastructure Layer**: Config reading and terminal management
- **Model Layer**: Service validation, session tracking, and error handling
- **Business logic**: Service naming conventions and validation rules

#### 4. Session Domain (`internal/core/session/`)
Session lifecycle management:
- **API Layer**: Session operation interfaces
- **Infrastructure Layer**: Session persistence and process management
- **Model Layer**: Session information and filtering capabilities

#### 5. Terminal Domain (`internal/core/terminal/`)
Terminal session handling for interactive processes:
- **Client/Server architecture**: Separation of terminal client and server concerns
- **Interactive session support**: Foundation for terminal multiplexer integration

#### 6. CLI Framework (`cmd/`)
Comprehensive command-line interface:
- **Root Command**: Base `hs` command with version support
- **Config Commands**: Configuration management (show, validate, set)
- **List Command**: Service listing and discovery
- **Service Commands**: Service lifecycle management (start, info)

### 🚧 In Progress
- Service session execution and terminal integration
- Session persistence and state management
- Terminal client/server implementation
- Process monitoring and lifecycle management

### 📋 Planned Features
- SSH connection management and tunneling
- Interactive terminal attachment and detachment
- Process health monitoring and auto-restart
- Configuration file generation tools
- Shell completion scripts and TUI mode

## Project Structure

```
hama-shell/
├── main.go                           # Application entry point
├── go.mod                           # Go module definition  
├── Makefile                         # Build automation
├── cmd/                             # CLI command implementations
│   ├── root.go                     # Root command (hs)
│   ├── config.go                   # Configuration commands
│   ├── list.go                     # List services command
│   └── service.go                  # Service management commands
├── internal/core/                   # Core domains (Clean Architecture)
│   ├── configuration/              # Configuration domain
│   │   ├── api/config_api.go      # Configuration API interface
│   │   ├── infra/                 # Infrastructure implementations
│   │   │   ├── config_manager.go  # Configuration management
│   │   │   └── viper_config.go    # Viper-based config handling
│   │   └── model/                 # Configuration domain models
│   │       └── configuration.go   # Config structures & validation
│   ├── service/                    # Service domain  
│   │   ├── api/service_api.go     # Service API interface
│   │   ├── infra/                 # Infrastructure implementations
│   │   │   ├── config_reader.go   # Service config reading
│   │   │   └── terminal_manager.go # Terminal session management
│   │   └── model/                 # Service domain models
│   │       ├── service.go         # Service structures & validation
│   │       └── errors.go          # Service-specific errors
│   ├── session/                    # Session domain
│   │   ├── api/session_api.go     # Session API interface  
│   │   ├── infra/session_manager.go # Session management implementation
│   │   ├── model/session.go       # Session domain models
│   │   └── session_manager.go     # Session manager
│   └── terminal/                   # Terminal domain
│       ├── client.go              # Terminal client
│       └── server.go              # Terminal server
```

## Configuration Format

HamaShell uses a simple YAML-based configuration format:

```yaml
projects:
  myapp:
    services:
      database:
        stages:
          dev:
            commands:
              - "ssh -i ${SSH_KEY_PATH} ${DB_USER}@dev-db.example.com"
              - "mysql -u root -p${DB_PASSWORD}"
          prod:
            commands:
              - "ssh -i ${SSH_KEY_PATH} ${DB_USER}@prod-db.example.com"
              - "mysql -u root -p${PROD_DB_PASSWORD}"
      api:
        stages:
          dev:
            commands:
              - "ssh -i ${SSH_KEY_PATH} ${API_USER}@dev-api.example.com"
              - "cd /app && npm start"
          prod:
            commands:
              - "ssh -i ${SSH_KEY_PATH} ${API_USER}@prod-api.example.com"
              - "cd /app && npm start"
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

### Phase 1: Architecture Foundation (Current)
- ✅ Clean Architecture implementation with domain separation
- ✅ Configuration domain with Viper integration
- ✅ Service domain with validation and session tracking
- ✅ CLI framework with config, list, and service commands
- 🚧 Session execution and terminal integration
- 🚧 Terminal client/server implementation

### Phase 2: Session Management  
- Terminal session persistence and state management
- Process lifecycle management and monitoring
- Interactive session attachment and detachment
- Session filtering and discovery

### Phase 3: Connection Features
- SSH client implementation and tunneling
- Port forwarding and connection health monitoring
- Connection configuration templates
- Automatic retry and reconnection logic

### Phase 4: Advanced Features
- Terminal multiplexer integration (tmux, zellij)
- Interactive TUI mode for session management
- Shell completion scripts and configuration helpers
- Configuration file generation and validation tools

### Phase 5: Polish & Distribution
- Comprehensive documentation and examples
- Installation packages and distribution
- CI/CD pipeline and automated testing
- Performance optimization and monitoring

This roadmap ensures steady progress toward the full vision while maintaining a stable, usable tool at each phase with proper Clean Architecture foundations.