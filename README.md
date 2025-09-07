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

### ✅ Domain-Driven Architecture
- **Clean Architecture pattern** with clear separation of concerns
- **Domain-based organization** (Configuration, Service, Session, Terminal)
- **API-Infrastructure-Model layering** for maintainable code

### ✅ Configuration Management (`internal/core/configuration/`)
- **YAML-based configuration** with validation using Viper
- **Type-safe domain models** with clear validation rules
- **Project-Service-Stage hierarchy** for organized configuration

### ✅ Service Management (`internal/core/service/`)
- **Service definition and validation** with domain models
- **Session management** for long-running processes
- **Terminal integration** for interactive sessions

### ✅ CLI Framework (`cmd/`)
- **Cobra-based commands** for config, list, and service operations
- **Version management** and help system
- **Modular command structure** for extensibility

## ⚙️ Configuration

Configure your connections using the simple **project.stage.service** pattern:

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

**Configuration Structure:**
- **`projects`** - Your project name (e.g., `myapp`, `ecommerce`)
- **`services`** - Service type (e.g., `database`, `api`, `cache`, `queue`)
- **`stages`** - Environment stage (e.g., `dev`, `staging`, `prod`)

Each stage defines:
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

**Service Management:**
```bash
# Start a service session
hs service start myapp.database.dev

# List services from configuration
hs list

# Show service information
hs service info myapp.database.dev
```

**Configuration Management:**
```bash
# Show current configuration
hs config show

# Validate configuration 
hs config validate

# Set configuration file path
hs config set /path/to/config.yaml
```

**Session Management:**
```bash
# List active sessions (planned)
hs list --sessions

# Attach to session (planned)
hs attach <session-id>
```

**General:**
```bash
# Show help
hs --help

# Show version
hs --version
```

## 🍀 Example Usage

```bash
# Build and run
go build -o hs
./hs --help

# Show current configuration
./hs config show

# List available services
./hs list

# Start a service
./hs service start myapp.database.dev
```

## 🏗️ Architecture

HamaShell follows Clean Architecture principles with domain-driven design:

```
┌─────────────────────────────────────────┐
│               CLI Layer                 │  ← Cobra commands (cmd/)
└─────────────────────────────────────────┘
                      │
┌─────────────────────────────────────────┐
│            API Layer                    │  ← Domain APIs
└─────────────────────────────────────────┘
                      │
┌─────────────────────────────────────────┐
│         Infrastructure Layer            │  ← Concrete implementations
└─────────────────────────────────────────┘
                      │
┌─────────────────────────────────────────┐
│           Model Layer                   │  ← Domain models & business logic
└─────────────────────────────────────────┘
```

### Key Domains

#### Configuration Domain (`internal/core/configuration/`)
- **API**: Configuration operations interface
- **Infrastructure**: Viper-based config management
- **Model**: Project-Service-Stage configuration structure

#### Service Domain (`internal/core/service/`)
- **API**: Service management interface  
- **Infrastructure**: Terminal management and config reading
- **Model**: Service definitions and validation

#### Session Domain (`internal/core/session/`)
- **API**: Session lifecycle management
- **Infrastructure**: Process session management
- **Model**: Session information and filtering

#### Terminal Domain (`internal/core/terminal/`)
- **Client/Server**: Terminal session handling for interactive processes

## 🚧 Development Status

### ✅ Completed
- [x] Clean Architecture implementation with domain separation
- [x] Configuration management with Viper integration  
- [x] Service definition and validation models
- [x] CLI structure with config, list, and service commands
- [x] Project-Service-Stage hierarchy support

### 🔄 In Progress  
- [ ] Service session execution and management
- [ ] Terminal session client/server implementation
- [ ] Session persistence and state management

### 📋 Planned
- [ ] SSH connection management
- [ ] Interactive terminal attachment
- [ ] Process monitoring and health checks
- [ ] Configuration file generation
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
├── main.go                           # Application entry point
├── go.mod                           # Go module definition  
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
└── Makefile                        # Build automation
```

## 📋 Roadmap

### Phase 1: Architecture Foundation (Current)
Clean Architecture implementation, domain separation, and configuration management

### Phase 2: Session Management  
Terminal sessions, process execution, and session persistence

### Phase 3: Connection Features
SSH client integration, port forwarding, and connection monitoring

### Phase 4: Advanced Features
Interactive terminal attachment, TUI mode, and shell completion

### Phase 5: Polish & Distribution
Documentation, packages, CI/CD, and performance optimization

---

**Status**: 🚧 Active Development | **License**: MIT | **Go Version**: 1.24+