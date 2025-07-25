# 🦛 HamaShell

## 🚀 Project Overview

This project is a **session and connection manager** designed for developers who need reliable, secure access to various hosts in single cli command.

It simplifies complex multi-step SSH tunneling and session setup by letting developers define their connections declaratively in a YAML file.

Unlike ad-hoc scripts, it offers **structured, secure, and controllable workflows**, making it easier to manage connections across projects and environments.

## ✨ Why use this tool?

✅ **Declarative & reproducible** — Define connections once in YAML and reuse them easily.

✅ **Secure by design** — Uses system environment variables to keep secrets hidden and safe.

✅ **Full process control** — Start, stop, check status, and manage connections interactively.

✅ **Hierarchical organization** — Organize connections by project and stage.

✅ **Parallel execution** — Run multiple sessions simultaneously without manual orchestration.

✅ **Multi-cloud ready** — Works seamlessly with AWS, Oracle Cloud, Naver Cloud, and on-premise.

✅ **Portable** — Runs on any Linux distro, integrates into CI/CD pipelines, and supports local dashboards for visibility.

## 💡 Core Features

### ✅ YAML-based configuration

* Define complex multi-step tunneling and SSH workflows declaratively.
* Supports dynamic command steps and environment variable substitution.

### ✅ Secure & flexible connections

* SSH with key-based authentication and multi-hop tunneling.
* Leverages environment variables to avoid hardcoding secrets.

### ✅ Powerful connection management

* Persistent and recoverable sessions.
* Process control: start, stop, view status, and monitor connections.
* Port forwarding and connection health checks with automatic retry.

### ✅ Developer-friendly CLI

* Interactive commands for managing sessions.
* Clear status reporting and logs for troubleshooting.

### ✅ Extensible and cloud-ready

* Integrates easily with major cloud providers and on-premises setups.
* Flexible enough to be used in local development, on-premise, or CI/CD pipelines.

## ⬇️ Installation
## 📙 How to use

### ⚙️ Configure

Configure your connections using the simple **project.stage.service** pattern:

```yaml
projects:
  myapp:
    description: "Main application project"
    stages:
      dev:
        services:
          db:
            description: "Develop database"
            commands:
              - "ssh -i ${SSH_KEY_PATH} ${BASTION_USER}@bastion.dev.com"
              - "${DEV_DB_PW}"
          api-server:
            description: "Develop database"
            commands:
              - "aws configure ,,,"
              - "aws ssm ,,,"
      prod:
        services:
          db:
            description: "Production database"
            commands:
              - "ssh -i ${SSH_KEY_PATH} ${BASTION_USER}@bastion.prod.com"
              - "${PROD_DB_PW}"
# Global settings
global_settings:
  timeout: 30
  retries: 3
  auto_restart: true
```

**Configuration Structure:**

* **`projects`** - Your project name (e.g., `myapp`, `ecommerce`)
* **`stages`** - Environment stage (e.g., `dev`, `staging`, `prod`)
* **`services`** - Service type (e.g., `db`, `server`, `jenkins`, `redis`, `api`)

Each service can have:
- **`host`** - Target hostname
- **`user`** - SSH username (supports env vars like `${USER}`)
- **`key`** - SSH key path (supports env vars like `${SSH_KEY_PATH}`)
- **`tunnel`** - Port forwarding (format: `local_port:remote_host:remote_port`)
- **`steps`** - Multi-step commands for complex connections

### ⌨ Commands

**Configuration Management:**

```shell
# Initialize new configuration
hama-shell init                  # Create new config.yaml with interactive prompts
```

**Session Management:**

```shell
# Run shell sessions
hama-shell run [session-name]    # Start/run a configured session

# Kill active sessions
hama-shell kill [session-name]   # Stop/kill running sessions

# Explain session commands
hama-shell explain [session-name] # Show what commands a session will execute
```

**Monitoring and Dashboard:**

```shell
# View dashboard
hama-shell dashboard             # Show interactive dashboard of all sessions
```

### 🍀 Example Usage Scenarios

```shell
# Initial setup
hama-shell init                  # Interactive configuration setup

# Basic session management
hama-shell run myapp.dev.db      # Start database session
hama-shell explain myapp.dev.db  # See what commands will be executed
hama-shell dashboard             # Monitor all active sessions
hama-shell kill myapp.dev.db     # Stop the session
```

```shell
# Development workflow
hama-shell run myapp.dev.server  # Start application server
hama-shell run myapp.dev.db      # Start database connection
hama-shell dashboard             # Monitor both sessions
```

## 🏗️ Architecture

HamaShell is designed with a clean, component-based architecture that promotes flexibility, maintainability, and cross-platform compatibility. The architecture centers around four core component groups with clear interfaces and responsibilities.

### Core Component Architecture

```mermaid
graph LR
    subgraph "Session Management"
        SM[Session Manager]
        SS[Session State]
        SP[Session Persistence]
    end

    subgraph "Connection Management"
        CM[Connection Manager]
        SSH[SSH Client]
        TN[Tunnel Manager]
        HM[Health Monitor]
    end

    subgraph "Configuration"
        CF[Config Loader]
        CV[Config Validator]
        CA[Alias Manager]
    end

    subgraph "Terminal Integration"
        TI[Terminal Interface]
        MX[Multiplexer Integration]
        SI[Shell Integration]
    end

    SM --> SP
    SM --> SS
    CM --> SSH
    CM --> TN
    CM --> HM
    CF --> CV
    CF --> CA
    SM --> CM
    SM --> TI
    TI --> MX
    TI --> SI
```

### Layered Architecture

```mermaid
graph TB
    subgraph "CLI Layer"
        CLI[CLI Commands]
        TUI[Interactive Mode]
        Completion[Shell Completion]
    end
    
    subgraph "Service Layer"
        SessionSvc[Session Service]
        ConfigSvc[Config Service]
        ConnectionSvc[Connection Service]
        TerminalSvc[Terminal Service]
    end
    
    subgraph "Core Components"
        SM[Session Manager]
        CM[Connection Manager]
        CF[Config Loader]
        TI[Terminal Interface]
    end
    
    subgraph "Infrastructure Layer"
        Persistence[File System]
        Network[SSH/Network]
        Process[Process Control]
        Platform[OS Abstraction]
    end
    
    CLI --> SessionSvc
    CLI --> ConfigSvc
    CLI --> ConnectionSvc
    TUI --> TerminalSvc
    SessionSvc --> SM
    ConfigSvc --> CF
    ConnectionSvc --> CM
    TerminalSvc --> TI
    SM --> Persistence
    CM --> Network
    CM --> Process
    TI --> Platform
```

### Package Structure

```mermaid
graph TD
    subgraph "Project Root"
        Main[main.go]
        
        subgraph "cmd/"
            RootCmd[root.go]
            StartCmd[start.go]
            StopCmd[stop.go]
            StatusCmd[status.go]
            ConfigCmd[config.go]
            AliasCmd[alias.go]
            InteractiveCmd[interactive.go]
        end
        
        subgraph "internal/"
            subgraph "service/"
                SessionSvc[session_service.go]
                ConfigSvc[config_service.go]
                ConnSvc[connection_service.go]
                TerminalSvc[terminal_service.go]
            end
            
            subgraph "core/"
                subgraph "session/"
                    SessionMgr[manager.go]
                    SessionState[state.go]
                    SessionPersist[persistence.go]
                end
                subgraph "connection/"
                    ConnMgr[manager.go]
                    SSHClient[ssh.go]
                    TunnelMgr[tunnel.go]
                    HealthMon[monitor.go]
                end
                subgraph "config/"
                    ConfigLoader[loader.go]
                    ConfigValidator[validator.go]
                    AliasManager[alias.go]
                end
                subgraph "terminal/"
                    TerminalIntf[interface.go]
                    MultiplexerInteg[multiplexer.go]
                    ShellInteg[shell.go]
                end
            end
            
            subgraph "infrastructure/"
                subgraph "storage/"
                    FileSystem[filesystem.go]
                    StateStore[state.go]
                end
                subgraph "network/"
                    NetworkClient[client.go]
                    PortForward[forwarder.go]
                end
                subgraph "process/"
                    ProcessCtrl[control.go]
                    Executor[executor.go]
                end
                subgraph "platform/"
                    OSAbstract[abstraction.go]
                    PlatformSpec[platform.go]
                end
            end
        end
        
        subgraph "pkg/"
            subgraph "types/"
                SessionTypes[session.go]
                ConfigTypes[config.go]
                ConnTypes[connection.go]
            end
            subgraph "integration/"
                TmuxInteg[tmux.go]
                ZellijInteg[zellij.go]
                ShellInteg[shell.go]
            end
        end
        
        subgraph "scripts/"
            subgraph "completion/"
                BashComp[bash.sh]
                ZshComp[zsh.sh]
                FishComp[fish.sh]
            end
            subgraph "multiplexer/"
                TmuxScript[tmux.sh]
                ZellijScript[zellij.sh]
            end
        end
    end
```

### Component Flow Architecture

```mermaid
sequenceDiagram
    participant User
    participant CLI
    participant Service
    participant Core
    participant Infrastructure
    participant External
    
    User->>CLI: Command Input
    CLI->>Service: Parse & Route
    Service->>Core: Component Operations
    Core->>Infrastructure: System Operations
    Infrastructure->>External: SSH/File/Process
    External-->>Infrastructure: Response
    Infrastructure-->>Core: Processed Data
    Core-->>Service: Component Results
    Service-->>CLI: Formatted Output
    CLI-->>User: Terminal Output
```

### Cross-Platform Integration Points

```mermaid
graph TB
    subgraph "Terminal Multiplexers"
        Tmux[Tmux Integration]
        Zellij[Zellij Integration]
        Screen[Screen Integration]
    end
    
    subgraph "Shell Integration"
        Bash[Bash Completion]
        Zsh[Zsh Completion]
        Fish[Fish Completion]
        PowerShell[PowerShell Completion]
    end
    
    subgraph "Platform Support"
        Linux[Linux/Unix]
        MacOS[macOS/Darwin]
        Windows[Windows]
    end
    
    subgraph "HamaShell Core"
        Core[Core Engine]
    end
    
    Core --> Tmux
    Core --> Zellij
    Core --> Screen
    Core --> Bash
    Core --> Zsh
    Core --> Fish
    Core --> PowerShell
    Core --> Linux
    Core --> MacOS
    Core --> Windows
```

### Integration Features

#### Terminal Multiplexer Support
- **Tmux**: Session creation, window management, pane splitting, layout management
- **Zellij**: Layout configuration, plugin integration, session persistence
- **Screen**: Basic session support and window management

#### Shell Integration
- **Completion Scripts**: Auto-completion for all commands and aliases
- **Environment Variables**: Seamless integration with shell environments
- **Path Resolution**: Smart path handling across different shells

#### Cross-Platform Features
- **Process Management**: Unified process handling across OS platforms
- **File System**: Cross-platform file operations and path handling
- **Network Stack**: Platform-specific network optimizations
- **Terminal Handling**: Native terminal integration per platform

### Architecture Benefits

#### Component-Based Design
- **Clear Responsibilities**: Each component has a focused, well-defined purpose
- **Loose Coupling**: Components interact through well-defined interfaces
- **High Cohesion**: Related functionality grouped within components
- **Easy Testing**: Interface-driven design enables comprehensive unit testing

#### Scalability & Extensibility
- **Service Layer**: Clean abstraction between CLI and core components
- **Interface-Driven**: Easy to add new implementations and protocols
- **Modular Structure**: Components can be developed and tested independently
- **Plugin Architecture**: Clean interfaces enable easy plugin development

#### Cross-Platform Compatibility
- **OS Abstraction**: Platform-specific code isolated in infrastructure layer
- **Terminal Agnostic**: Works with tmux, zellij, screen, and native terminals
- **Shell Universal**: Supports bash, zsh, fish with auto-completion
- **Multiplexer Integration**: Seamless integration with popular multiplexers

#### Maintainability & Reliability
- **Separation of Concerns**: Core logic separated from infrastructure details
- **Dependency Injection**: Enables mocking and comprehensive testing
- **Error Handling**: Consistent error propagation across components
- **State Management**: Clear session state handling and persistence

This component-based architecture provides a solid foundation for HamaShell that maintains simplicity while supporting complex session management scenarios across multiple platforms and terminal environments.