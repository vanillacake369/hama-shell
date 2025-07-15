# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This project is a session manager for developers at HAMALAB company to access to other host while developing.
Main purpose is to cope with any outdoor host even SaaS services (AWS, Oracle, Naver Cloud, Homelab).  
Each developer declares yaml file that claims each step of command for tunneling to other host.



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

## Architecture

This is currently a single-file Go application with:
- `main.go`: Entry point containing a simple greeting and loop demonstration
- `go.mod`: Module definition file specifying Go 1.24

The codebase is in its initial state and appears to be a template or starting point for a shell-related project based on the name "hama-shell".