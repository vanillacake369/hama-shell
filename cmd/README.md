# Command Package Documentation

## Overview

This package implements the CLI commands for HamaShell using the Cobra framework. It provides the command-line interface for managing sessions and connections.

## Go's init() Function

### What is init()?

`init()` is a special function in Go that runs automatically when a package is imported. Key characteristics:

- **Automatic execution**: Never called explicitly, runs when package is imported
- **Multiple init() allowed**: A package can have multiple init() functions
- **Execution order**: Runs after package-level variables are initialized
- **Runs once**: Even if package is imported multiple times

### init() Example

```go
package example

import "fmt"

// 1. Package-level variables initialize first
var config = loadDefaultConfig()

// 2. init() functions run second (in source order)
func init() {
    fmt.Println("First init")
}

func init() {
    fmt.Println("Second init")
}

// 3. Package is ready for use
```

## Cobra Initialization Process

### Execution Timeline in HamaShell

```
1. Program starts
   ↓
2. Go runtime initializes
   ↓
3. main.go imports "hama-shell/cmd"
   ↓
4. cmd package initialization:
   a. Package variables (rootCmd, AppConfig, etc.)
   b. init() function runs:
      - Registers cobra.OnInitialize(initConfig)
      - Sets up persistent flags
   ↓
5. main() function executes
   ↓
6. cmd.Execute() called
   ↓
7. rootCmd.Execute() runs
   ↓
8. Cobra command parsing:
   a. Parse command-line flags
   b. Trigger OnInitialize callbacks
   c. initConfig() runs (loads configuration)
   d. Execute matched command
```

### Key Components

#### 1. init() Function (root.go:46-51)
```go
func init() {
    // Register callback to run before any command
    cobra.OnInitialize(initConfig)
    
    // Setup persistent flags available to all commands
    rootCmd.PersistentFlags().StringVar(&configFile, "config", "", 
        "config file (default is $HOME/hama-shell.yaml)")
}
```

**Purpose**: 
- Sets up command initialization hooks
- Defines global flags
- Runs automatically on package import

#### 2. cobra.OnInitialize()
```go
cobra.OnInitialize(initConfig)
```

**Purpose**:
- Registers functions to run after flag parsing but before command execution
- Allows configuration loading with user-provided flag values
- Can register multiple callbacks (executed in order)

#### 3. initConfig() Function (root.go:54-81)
```go
func initConfig() {
    // Runs when any command executes
    // Has access to parsed flag values
    validator := config.NewValidator()
    AppConfig, err = validator.ParseAndValidate(configFile)
    // ...
}
```

**Purpose**:
- Loads and validates configuration
- Runs after flags are parsed
- Makes config available globally via AppConfig

## Why This Architecture?

### Separation of Concerns

1. **init()**: Setup phase - registers callbacks and flags
2. **OnInitialize callbacks**: Configuration phase - loads config with flag values
3. **Command execution**: Action phase - performs the actual work

### Benefits

- **Flag access**: Config loading can use flag values (e.g., custom config path)
- **Single load**: Configuration loaded once, available to all subcommands
- **Error handling**: Config errors caught before command execution
- **Clean separation**: Each phase has clear responsibilities

## Command Structure

```
hama-shell                    # Root command (root.go)
├── start [project.stage.service]  # Start sessions (start.go)
├── stop [project.stage.service]   # Stop sessions (stop.go)
├── status                          # Check status (status.go)
└── config                          # Manage config (config.go)
    ├── validate
    └── generate
```

Each command file follows the pattern:
1. Define command struct with cobra.Command
2. Implement command logic in Run function
3. Register with parent command in init()

## Global Variables

- **AppConfig**: Holds parsed configuration, available to all commands
- **configFile**: Path to configuration file from --config flag
- **rootCmd**: Base command, parent of all subcommands

## Testing

The package includes tests for configuration validation and command behavior. Run tests with:

```bash
go test ./cmd/...
```