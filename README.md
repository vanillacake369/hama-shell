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
        description: "Development environment"
        services:
          db:
            description: "PostgreSQL database connection"
            host: "dev-db.myapp.com"
            user: "${DB_USER}"
            key: "${SSH_KEY_PATH}"
            tunnel: "5432:localhost:5432"
          
          server:
            description: "Application server"
            host: "dev-app.myapp.com"
            user: "${APP_USER}"
            key: "${SSH_KEY_PATH}"
          
          jenkins:
            description: "CI/CD Jenkins server"
            host: "jenkins.myapp.com"
            user: "jenkins"
            key: "${JENKINS_KEY}"
            tunnel: "8080:localhost:8080"
            
      prod:
        services:
          db:
            description: "Production database"
            steps:
              - command: "ssh -i ${SSH_KEY_PATH} ${BASTION_USER}@bastion.prod.com"
              - command: "ssh -L 5432:prod-db:5432 db-reader@prod-db-proxy"

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

**Using project.stage.service pattern:**

```shell
# Connect to services using the dot notation
hama-shell myapp.dev.db          # Connect to development database
hama-shell myapp.prod.server     # Connect to production server
hama-shell ecommerce.dev.redis   # Connect to development Redis
```

**Core Session Management:**

```shell
# Start connections
hama-shell start myapp.dev.db

# Start multiple connections
hama-shell start myapp.dev.db myapp.dev.server

# Check status
hama-shell status myapp.dev.db
hama-shell status --all

# Stop connections
hama-shell stop myapp.dev.db
hama-shell stop --all
```

**Service Discovery:**

```shell
# List all available services
hama-shell list
hama-shell list myapp             # List services for specific project
hama-shell list myapp.dev         # List services for specific stage

# Show service details
hama-shell show myapp.dev.db

# Validate configuration
hama-shell validate
hama-shell validate myapp.dev.db
```

**Connection Testing:**

```shell
# Test connection without starting
hama-shell test myapp.dev.db
hama-shell test myapp.dev.db --dry-run

# Health check
hama-shell health myapp.dev.db
hama-shell health --all
```

**Logs and Monitoring:**

```shell
# View connection logs
hama-shell logs myapp.dev.db
hama-shell logs myapp.dev.db --tail 50
hama-shell logs myapp.dev.db --follow

# Monitor connection status
hama-shell monitor myapp.dev.db
hama-shell monitor --all
```

**Interactive Mode:**

```shell
# Interactive service selection
hama-shell interactive
hama-shell i

# Dashboard view
hama-shell dashboard
hama-shell dash
```

### 🍀 Example Usage Scenarios

```shell
# Quick development workflow
hama-shell start myapp.dev.db
hama-shell status myapp.dev.db
hama-shell logs myapp.dev.db --follow
```

```shell
# Production database access
hama-shell test myapp.prod.db --dry-run
hama-shell start myapp.prod.db
hama-shell monitor myapp.prod.db
```

```shell
# Multi-service development setup
hama-shell start myapp.dev.db myapp.dev.server myapp.dev.jenkins
hama-shell status myapp.dev.db myapp.dev.server myapp.dev.jenkins
hama-shell stop myapp.dev.db myapp.dev.server myapp.dev.jenkins
```

```shell
# Cross-project workflow
hama-shell start myapp.dev.db ecommerce.dev.redis
hama-shell list myapp
hama-shell list ecommerce
hama-shell stop --all
```