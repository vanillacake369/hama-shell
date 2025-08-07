# Config Package

The config package provides YAML-based configuration validation and parsing for HamaShell. It ensures configuration files follow the correct schema and provides type-safe access to configuration data.

## Overview

HamaShell uses a hierarchical configuration structure organized as `project.stage.service`, allowing you to define connection commands in a structured, reusable way.

## Configuration Schema

### Core Structure
```yaml
projects:
  <project-name>:
    description: "Project description"
    stages:
      <stage-name>:
        description: "Stage description (optional)"
        services:
          <service-name>:
            description: "Service description"
            commands:
              - "command1"
              - "command2"

global_settings:
  timeout: 30          # Connection timeout in seconds (default: 30)
  retries: 3           # Number of retry attempts (default: 3)  
  auto_restart: false  # Auto-restart on failure (default: false)
```

### Configuration Files

The validator searches for configuration files in the following locations:
1. `~/hama-shell.yaml`
2. `~/hama-shell.yml` 
3. `./hama-shell.yaml`
4. `./hama-shell.yml`

## API Reference

### Validator

The main `Validator` struct provides configuration validation and parsing:

```go
// Create new validator
validator := config.NewValidator()

// Parse and validate config file
config, err := validator.ParseAndValidate("path/to/config.yaml")

// Validate viper configuration
err := validator.ValidateViper()

// Validate from map data
err := validator.ValidateFromMap(data)

// Validate Config struct directly
err := validator.ValidateConfig(config)
```

### Query Functions

```go
// Get available projects
projects := validator.GetProjects()

// Get stages for a project
stages := validator.GetStages("myproject")

// Get services for a project and stage  
services := validator.GetServices("myproject", "dev")

// Get commands for a session path
commands, err := config.GetCommands(config, "myproject.dev.database")
```

## Types

### Config
Main configuration structure containing projects and global settings.

### Project
Groups deployment stages under a project name with optional description.

### Stage  
Represents a deployment environment (dev, staging, prod) containing services.

### Service
Defines connection details and commands for a specific service.

### GlobalSettings
Configures timeout, retry logic, and auto-restart behavior.

## Validation Rules

### Required Fields
- `projects`: Must contain at least one project
- `stages`: Each project must have at least one stage
- `services`: Each stage must have at least one service  
- `commands`: Each service must have at least one non-empty command

### Data Types
- Project/stage/service names: strings (keys in YAML maps)
- Descriptions: strings (optional for stages)
- Commands: array of non-empty strings
- Timeout/retries: non-negative integers
- Auto-restart: boolean

### Hierarchical Path Format
Session paths follow the format: `project.stage.service`

Examples:
- `myapp.dev.database`
- `myapp.prod.api`
- `monitoring.staging.grafana`

## Environment Variables

Commands can reference environment variables using `${VAR_NAME}` syntax:

```yaml
commands:
  - "ssh -i ${SSH_KEY_PATH} ${USER}@${HOST}"
  - "mysql -u root -p${DB_PASSWORD}"
```

## Error Handling

The validator provides detailed error messages with context:
- Missing required sections
- Invalid data types
- Empty required fields
- Invalid session path formats
- File reading/parsing errors

## Example Usage

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/your-org/hama-shell/internal/core/config"
)

func main() {
    validator := config.NewValidator()
    
    // Parse config file
    cfg, err := validator.ParseAndValidate("hama-shell.yaml")
    if err != nil {
        log.Fatal(err)
    }
    
    // Get commands for a session
    commands, err := config.GetCommands(cfg, "myapp.dev.database")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("Commands:", commands)
}
```