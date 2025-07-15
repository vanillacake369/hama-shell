# 🦛 HamaShell

## 🚀 Project Overview

This project is a **session and connection manager** designed for developers who need reliable, secure access to various hosts in single cli command.

It simplifies complex multi-step SSH tunneling and session setup by letting developers define their connections declaratively in a YAML file.

Unlike ad-hoc scripts, it offers **structured, secure, and controllable workflows**, making it easier to manage connections across projects and environments.

## ✨ Why use this tool?

✅ **Declarative & reproducible** — Define connections once in YAML and reuse them easily.

✅ **Secure by design** — Uses system environment variables to keep secrets hidden and safe.

✅ **Full process control** — Start, stop, check status, and manage connections interactively.

✅ **Alias & hierarchy support** — Organize connections by project, stage, and developer.

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

First, you have to declare your yaml like this in order to apply command step by step.

```yaml
projects:
  - name: "<project-name>"
    stages:
      - name: "<stage-name>"
        developers:
          - name: "<developer-name>"
            sessions:
              - name: "<session-name>"
                description: "<brief-description>"
                steps:
                  - command: "<shell-or-ssh-command>"
                  - command: "<next-step>"
                parallel: <true|false>
aliases:
  - myapp-prod: "myapp.production.alice.db-tunnel"
global_settings:
  retries: <number>
  timeout: <seconds>
  auto_restart: <true|false>
```

Here's the definition of each declaration.

* **`projects`**
  An **array** of project objects, each declared via `- name: "<project-name>"` to instantiate and name the project (e.g. `my-awesome-app`).

* **`stages`**
  An **array** under each project, each declared via `- name: "<stage-name>"` to group its environments (e.g. `development`, `staging`, `production`).

* **`developers`**
  An **array** under each stage, each declared via `- name: "<developer-name>"` to list team members who need sessions there.

* **`sessions`**
  An **array** of connection workflows, each declared via `- name: "<session-name>"` to that the developer can run:

    * **`description: "<brief-description>"`**
      What this session does, in a nutshell.
    * **`steps`**
      Ordered **array** of commands to execute.
    * **`parallel: <true|false>`**
      Whether this session can run concurrently (`true`) or must run sequentially (`false`).

* **`aliases`**
  An **array** of aliases on groups of shell commands
  
* **`global_settings`**
  Defaults applied across **all** projects, stages, and developers:

    * `retries`: max retry attempts per failed step
    * `timeout`: per‑step timeout in seconds
    * `auto_restart`: auto‑restart dropped sessions (`true`/`false`)

### ⌨ Command

Core Session Management

```shell
# Start sessions using aliases
hama-shell start aws-dev
hama-shell start db-dev

# Start multiple sessions
hama-shell start aws-dev db-dev
hama-shell start aws-dev,db-dev

# Start with full path (fallback)
hama-shell start myapp.development.alice.aws-bastion-tunnel

Enhanced Status Commands

# View status of specific alias
hama-shell status aws-dev

# View status of multiple aliases
hama-shell status aws-dev db-dev

# View all active sessions
hama-shell status --all

# Detailed status with step information
hama-shell status aws-dev --detailed
```

Alias Management Commands

```shell
# List all available aliases
hama-shell alias list

# Show what an alias points to
hama-shell alias show aws-dev
# Output: aws-dev -> myapp.development.alice.aws-bastion-tunnel

# Search aliases by pattern
hama-shell alias search dev
hama-shell alias search "*-prod"

# Validate all aliases (check if sessions exist)
hama-shell alias validate
```

Advanced Session Operations

```shell
# Stop sessions by alias
hama-shell stop aws-dev
hama-shell stop --all

# Restart sessions
hama-shell restart aws-dev
hama-shell restart aws-dev --force

# Test connection without starting
hama-shell test aws-dev
hama-shell test aws-dev --dry-run
```


Logging and Monitoring

```shell
# View logs for aliased sessions
hama-shell logs aws-dev
hama-shell logs aws-dev --tail 50
hama-shell logs aws-dev --follow

# Monitor session health
hama-shell monitor aws-dev
hama-shell monitor --all
```

Configuration Management

```shell
# Show configuration for alias
hama-shell config show aws-dev

# Validate configuration
hama-shell config validate
hama-shell config validate --alias aws-dev

# List sessions that can be aliased
hama-shell config sessions
```

Interactive Features

```shell
# Interactive alias selection
hama-shell interactive
hama-shell i

# Dashboard with alias support
hama-shell dashboard
hama-shell dash
```

Global Flags Enhanced for Aliases

```shell
--alias-only        # Only work with aliases, not full paths                                                                                                                                                                                                                                                                                                                                                                      
--resolve          # Show full session path for aliases                                                                                                                                                                                                                                                                                                                                                                           
--no-alias         # Force using full session paths
```

### 🍀 Example Usage Scenarios

```shell
# Quick development workflow
hama-shell start aws-dev
hama-shell status aws-dev
hama-shell logs aws-dev --follow
```

```shell
# Production deployment
hama-shell start prod-db
hama-shell test prod-api --dry-run
hama-shell start prod-api
```

```shell
# Batch operations
hama-shell start aws-dev db-dev
hama-shell status aws-dev db-dev
hama-shell stop aws-dev db-dev
```