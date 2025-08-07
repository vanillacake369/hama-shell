# 🦛 HamaShell

A **session and connection manager** designed for developers who need reliable, secure access to various hosts in single CLI command.

## 🚀 Project Overview

HamaShell simplifies complex multi-step SSH tunneling and session setup by letting developers define their connections declaratively in a YAML file. Unlike ad-hoc scripts, it offers **structured, secure, and controllable workflows**, making it easier to manage connections across projects and environments.

## ✨ Why use this tool?

✅ **Declarative & reproducible** — Define connections once in YAML and reuse them easily

✅ **Secure by design** — Uses system environment variables to keep secrets hidden and safe  

✅ **Full process control** — Start, stop, check status, and manage connections interactively

✅ **Hierarchical organization** — Organize connections using project.stage.service pattern

✅ **Parallel execution** — Run multiple sessions simultaneously with process isolation

✅ **Cross-platform ready** — Works on Unix/Linux and Windows systems

✅ **Simple & focused** — Clean, minimal implementation that's easy to understand and extend

## 🏗️ Current Implementation

HamaShell is actively developed with a focus on core functionality. The current implementation includes:

### ✅ Process Executor
- **Hierarchical process management** using project.stage.service keys
- **Thread-safe operations** with concurrent process handling  
- **Platform-aware process control** (Unix signals, Windows termination)
- **Process isolation** using process groups

### ✅ Configuration System
- **YAML-based configuration** with validation
- **Environment variable support** for secure credential handling
- **Type-safe parsing** with clear error messages

### ✅ CLI Framework  
- **Cobra-based commands** for start, stop, status, config operations
- **Configuration loading** with validation on startup
- **Clear error reporting** and help system

## ⚙️ Configuration

Configure your connections using the simple **project.stage.service** pattern:

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
      prod:
        services:
          database:
            description: "Production database connection"
            commands:
              - "ssh -i ${SSH_KEY_PATH} ${DB_USER}@prod-db.example.com"
              - "mysql -u root -p${PROD_DB_PASSWORD}"

global_settings:
  timeout: 30
  retries: 3  
  auto_restart: true
```

**Configuration Structure:**
- **`projects`** - Your project name (e.g., `myapp`, `ecommerce`)
- **`stages`** - Environment stage (e.g., `dev`, `staging`, `prod`)  
- **`services`** - Service type (e.g., `database`, `api`, `cache`, `queue`)

Each service defines:
- **`description`** - Human-readable service description
- **`commands`** - Sequential commands to execute for the connection

## ⬇️ Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/hama-shell
cd hama-shell

# Build the application
go build -o hama-shell

# Run directly with Go
go run main.go
```

## ⌨️ Commands

**Session Management:**
```bash
# Start a session (planned)
hama-shell start myapp.dev.database

# Stop a session (planned)  
hama-shell stop myapp.dev.database

# Check session status (planned)
hama-shell status myapp.dev.database
```

**Configuration Management:**
```bash
# Validate configuration (current)
hama-shell config validate

# Show configuration help (current)
hama-shell config --help
```

**General:**
```bash
# Show help
hama-shell --help

# Show version
hama-shell --version
```

## 🍀 Example Usage

```bash
# Build and run
go build -o hama-shell
./hama-shell --help

# Validate your configuration
./hama-shell config validate

# Check specific command help
./hama-shell start --help
```

## 🏗️ Architecture

HamaShell follows a simple, focused architecture:

```
┌─────────────────┐
│   CLI Commands  │  ← Cobra-based command interface
└─────────────────┘
         │
┌─────────────────┐
│ Config Validator│  ← YAML parsing and validation  
└─────────────────┘
         │
┌─────────────────┐
│ Process Executor│  ← Hierarchical process management
└─────────────────┘
         │
┌─────────────────┐
│ OS Integration  │  ← Platform-specific process control
└─────────────────┘
```

### Key Components

#### Process Executor (`internal/core/executor/`)
- Manages processes with hierarchical keys (project.stage.service)
- Thread-safe operations using `sync.Map`
- Platform-specific process handling (Unix/Windows)
- Graceful shutdown with proper signal handling

#### Configuration System (`internal/core/config/`)  
- Parses and validates YAML configuration files
- Type-safe Go structs with clear error messages
- Environment variable substitution support

#### CLI Framework (`cmd/`)
- Cobra-based command structure
- Configuration loading and validation
- Help system and error reporting

## 🚧 Development Status

### ✅ Completed
- [x] Basic CLI structure with Cobra
- [x] Configuration validation system
- [x] Process executor with hierarchical management
- [x] Cross-platform process handling
- [x] Thread-safe operations

### 🔄 In Progress  
- [ ] Session state persistence
- [ ] Enhanced error handling
- [ ] Configuration file generation

### 📋 Planned
- [ ] SSH connection management
- [ ] Port forwarding and tunneling  
- [ ] Terminal multiplexer integration
- [ ] Interactive TUI mode
- [ ] Shell completion scripts

## 🛠️ Development

### Requirements
- Go 1.24+
- Unix/Linux or Windows environment

### Building
```bash
# Format code
go fmt ./...

# Run tests
go test ./...

# Build application
go build -o hama-shell

# Run with go
go run main.go
```

### Testing
```bash  
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific component tests
go test ./internal/core/executor/
go test ./internal/core/config/
```

### Project Structure
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

## 📋 Roadmap

### Phase 1: Core Foundation (Current)
Focus on reliable process management and configuration handling

### Phase 2: Connection Management  
SSH client, tunneling, port forwarding, and connection monitoring

### Phase 3: Advanced Features
Terminal multiplexer integration, TUI mode, and shell completion

### Phase 4: Polish & Distribution
Documentation, packages, CI/CD, and performance optimization

---

**Status**: 🚧 Active Development | **License**: MIT | **Go Version**: 1.24+