package executor

import (
	"os"
	"os/exec"
	"time"
)

// ExecutionMode defines how commands should be executed
type ExecutionMode int

const (
	// ExecutionModeForeground runs commands interactively with I/O forwarding
	ExecutionModeForeground ExecutionMode = iota
	// ExecutionModeBackground runs commands detached in background
	ExecutionModeBackground
)

// String returns the string representation of ExecutionMode
func (e ExecutionMode) String() string {
	switch e {
	case ExecutionModeForeground:
		return "foreground"
	case ExecutionModeBackground:
		return "background"
	default:
		return "unknown"
	}
}

// ProcessCommand represents a running command with its metadata (legacy)
type ProcessCommand struct {
	Cmd       string
	Process   *os.Process
	StartTime time.Time
	Key       string
}

// SessionGroup represents a group of processes running under a single session/PGID
type SessionGroup struct {
	Key        string            // Unique identifier for this session
	SessionID  int               // Session ID (SID)
	PGID       int               // Process Group ID
	Supervisor *exec.Cmd         // Main supervisor process
	Segments   []*Segment        // All command segments in this session
	StartTime  time.Time         // When the session started
	Done       chan struct{}     // Signal when session completes
}

// Segment represents a single command segment (SSH or shell)
type Segment struct {
	Type     string      // "ssh" or "shell"
	Commands []string    // Commands for this segment
	Process  *os.Process // Individual process for this segment
	PTY      *os.File    // PTY for SSH segments (nil for shell segments)
}

// CommandSegment represents parsed command information for building segments
type CommandSegment struct {
	Type     string   // "ssh" or "shell"
	Commands []string // Commands for this segment
}

// processManager defines platform-specific process management operations
type processManager interface {
	// setupCommand configures platform-specific settings for the command
	setupCommand(cmd *exec.Cmd)

	// terminateProcess gracefully terminates a process
	terminateProcess(process *os.Process) error

	// setupSupervisor configures supervisor with session/PGID settings
	setupSupervisor(cmd *exec.Cmd) error
}

// PTYInterface defines operations for pseudo-terminal management (for testing)
type PTYInterface interface {
	Start(cmd *exec.Cmd) (PTYFile, error)
}

// PTYFile defines file operations needed for PTY management
type PTYFile interface {
	Read(b []byte) (n int, err error)
	Write(b []byte) (n int, err error)
	Close() error
}

// SupervisorInterface defines operations for supervisor process management (for testing)
type SupervisorInterface interface {
	Start() error
	Wait() error
	Signal(sig os.Signal) error
	GetPID() int
}
