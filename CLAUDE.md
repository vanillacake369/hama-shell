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

## Development Commands

### Building and Running

- `go run main.go` - Compile and run the main program
- `go build` - Compile the program into an executable
- `go build -o hama-shell` - Build with custom executable name

### Testing and Code Quality

- `go test` - Run tests (no tests currently exist)
- `go fmt` - Format Go source code
- `go vet` - Report likely mistakes in packages
- `go mod tidy` - Clean up module dependencies

### Module Management

- `go mod init` - Initialize module (already done)
- `go mod download` - Download module dependencies
- `go get <package>` - Add new dependencies

## TODO: Implementation Tasks

### Phase 1: Core Infrastructure
- [ ] Design and implement YAML configuration parser
- [ ] Create session data structures (Project, Stage, Developer, Session)
- [ ] Implement global alias registry and resolution
- [ ] Build basic CLI command structure with cobra/cli library
- [ ] Set up logging and error handling framework

### Phase 2: Session Management
- [ ] Implement session lifecycle management (start, stop, restart)
- [ ] Create process management for SSH commands
- [ ] Add session state persistence and recovery
- [ ] Implement parallel session execution
- [ ] Build session monitoring and health checks

### Phase 3: Advanced Features
- [ ] Add interactive CLI mode and dashboard/TUI
- [ ] Implement comprehensive status reporting
- [ ] Create alias management commands
- [ ] Add configuration validation and testing
- [ ] Build log management and viewing capabilities

### Phase 4: Security & Operations
- [ ] Implement secure environment variable handling
- [ ] Add connection retry logic and error recovery
- [ ] Create port forwarding management
- [ ] Add comprehensive testing suite
- [ ] Build installation and deployment scripts

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
└── tests/                 # Test files
```

### Current State
- Basic Go module initialized with Go 1.24
- Single `main.go` file with placeholder code
- Ready for modular architecture implementation