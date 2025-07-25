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