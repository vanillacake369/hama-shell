# HamaShell Simple Executor

A lightweight, Go-idiomatic process executor with hierarchical key-based management.

## Features

- **Simple API**: Just `Run()`, `StopAll()`, `StopByKey()`, and `GetStatus()`
- **Thread-safe**: Uses `sync.Map` for concurrent operations
- **Hierarchical keys**: Organize processes by project.stage.service pattern
- **Platform-aware**: Proper signal handling for Unix/Linux and Windows
- **Lightweight**: No complex abstractions, just clean Go code

## Architecture

```
executor/
├── executor.go          # Main executor with sync.Map registry
├── process_common.go    # Shared types and interfaces
├── process_unix.go      # Unix-specific process management (SIGTERM/SIGKILL)
└── process_windows.go   # Windows-specific process management
```

## Usage

```go
// Create executor
exec := executor.New()

// Run processes with hierarchical keys
exec.Run("project1.stage1.serviceA", "sleep 30")
exec.Run("project1.stage1.serviceA", "sleep 60")  // Multiple processes per key
exec.Run("project2.stage1.serviceB", "ping google.com")

// Check status
status := exec.GetStatus()
// Returns: map[string][]*ProcessInfo

// Stop specific service
exec.StopByKey("project1.stage1.serviceA")

// Stop everything
exec.StopAll()
```

## Platform-Specific Behavior

### Unix/Linux
- Uses process groups (`Setpgid`) for better process tree management
- Graceful shutdown: SIGTERM → 5s wait → SIGKILL
- Prevents zombie processes

### Windows
- Uses `CREATE_NEW_PROCESS_GROUP` for process isolation
- Direct termination (no SIGTERM equivalent)

## Design Principles

1. **Simplicity**: Minimal interface, maximum functionality
2. **Go-idiomatic**: Uses goroutines, channels, and standard library
3. **Thread-safe**: Safe for concurrent use without external locking
4. **No magic**: Clear, understandable code flow

## Testing

```bash
go test -v ./internal/core/executor/
```

## Demo

Run the demo to see it in action:

```bash
go run internal/core/executor/demo/main.go
```