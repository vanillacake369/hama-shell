# HamaShell Component Architecture

## Overview
HamaShell is designed with a clean, component-based architecture that promotes flexibility, maintainability, and cross-platform compatibility. The architecture centers around four core component groups with clear interfaces and responsibilities.

## Core Component Architecture

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

## Layered Architecture

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

## Package Structure

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

## Component Flow Architecture

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

## Core Component Interfaces

### Session Management
```go
type SessionManager interface {
    Create(config SessionConfig) (*Session, error)
    Start(sessionID string) error
    Stop(sessionID string) error
    GetStatus(sessionID string) (SessionStatus, error)
    List() ([]*Session, error)
}

type SessionState interface {
    Save(session *Session) error
    Load(sessionID string) (*Session, error)
    Delete(sessionID string) error
}

type SessionPersistence interface {
    Store(session *Session) error
    Retrieve(sessionID string) (*Session, error)
    Remove(sessionID string) error
}
```

### Connection Management
```go
type ConnectionManager interface {
    Connect(config ConnectionConfig) (Connection, error)
    Disconnect(connectionID string) error
    GetStatus(connectionID string) (ConnectionStatus, error)
    List() ([]Connection, error)
}

type SSHClient interface {
    Connect(host string, config SSHConfig) error
    Execute(command string) ([]byte, error)
    Disconnect() error
}

type TunnelManager interface {
    CreateTunnel(config TunnelConfig) (Tunnel, error)
    CloseTunnel(tunnelID string) error
    ListTunnels() ([]Tunnel, error)
}

type HealthMonitor interface {
    Monitor(connectionID string) (<-chan HealthStatus, error)
    CheckHealth(connectionID string) (HealthStatus, error)
    StopMonitoring(connectionID string) error
}
```

### Configuration
```go
type ConfigLoader interface {
    Load(path string) (*Config, error)
    LoadFromBytes(data []byte) (*Config, error)
    Reload() (*Config, error)
}

type ConfigValidator interface {
    Validate(config *Config) error
    ValidateSession(session SessionConfig) error
}

type AliasManager interface {
    Resolve(alias string) (string, error)
    List() (map[string]string, error)
    Add(alias, path string) error
    Remove(alias string) error
}
```

### Terminal Integration
```go
type TerminalInterface interface {
    Attach(sessionID string) error
    Detach(sessionID string) error
    SendInput(sessionID string, input []byte) error
    GetOutput(sessionID string) (<-chan []byte, error)
}

type MultiplexerIntegration interface {
    CreateSession(name string, config MultiplexerConfig) error
    AttachToSession(sessionID string) error
    DetachFromSession(sessionID string) error
    ListSessions() ([]MultiplexerSession, error)
}

type ShellIntegration interface {
    ExecuteCommand(command string) ([]byte, error)
    SetEnvironment(env map[string]string) error
    GetCompletion(input string) ([]string, error)
}
```

## Cross-Platform Integration Points

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

## Integration Features

### Terminal Multiplexer Support
- **Tmux**: Session creation, window management, pane splitting, layout management
- **Zellij**: Layout configuration, plugin integration, session persistence
- **Screen**: Basic session support and window management

### Shell Integration
- **Completion Scripts**: Auto-completion for all commands and aliases
- **Environment Variables**: Seamless integration with shell environments
- **Path Resolution**: Smart path handling across different shells

### Cross-Platform Features
- **Process Management**: Unified process handling across OS platforms
- **File System**: Cross-platform file operations and path handling
- **Network Stack**: Platform-specific network optimizations
- **Terminal Handling**: Native terminal integration per platform

## Architecture Benefits

### Component-Based Design
- **Clear Responsibilities**: Each component has a focused, well-defined purpose
- **Loose Coupling**: Components interact through well-defined interfaces
- **High Cohesion**: Related functionality grouped within components
- **Easy Testing**: Interface-driven design enables comprehensive unit testing

### Scalability & Extensibility
- **Service Layer**: Clean abstraction between CLI and core components
- **Interface-Driven**: Easy to add new implementations and protocols
- **Modular Structure**: Components can be developed and tested independently
- **Plugin Architecture**: Clean interfaces enable easy plugin development

### Cross-Platform Compatibility
- **OS Abstraction**: Platform-specific code isolated in infrastructure layer
- **Terminal Agnostic**: Works with tmux, zellij, screen, and native terminals
- **Shell Universal**: Supports bash, zsh, fish with auto-completion
- **Multiplexer Integration**: Seamless integration with popular multiplexers

### Maintainability & Reliability
- **Separation of Concerns**: Core logic separated from infrastructure details
- **Dependency Injection**: Enables mocking and comprehensive testing
- **Error Handling**: Consistent error propagation across components
- **State Management**: Clear session state handling and persistence

This component-based architecture provides a solid foundation for HamaShell that maintains simplicity while supporting complex session management scenarios across multiple platforms and terminal environments.