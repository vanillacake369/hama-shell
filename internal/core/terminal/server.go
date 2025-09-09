package terminal

import (
	"os"
	"time"
)

// Server manages PTY sessions for multiple clients
type Server interface {
	// CreateSession creates a new PTY session
	CreateSession(sessionID, shell string, args []string) (Session, error)

	// KillSession terminates a session
	KillSession(sessionID string) error

	// ResizeSession updates the terminal size for a session
	ResizeSession(sessionID string, rows, cols uint16) error

	// GetSession returns session information
	GetSession(sessionID string) (Session, error)

	// ListSessions returns all active sessions
	ListSessions() map[string]Session

	// Shutdown gracefully shuts down the server
	Shutdown() error
}

// Session represents a single PTY session
type Session interface {
	// GetID returns the session ID
	GetID() string

	// GetStartTime returns when the session was started
	GetStartTime() time.Time

	// GetPID returns the process ID if available
	GetPID() int

	// IsRunning returns true if the session is currently running
	IsRunning() bool

	// GetInfo returns session information as a map
	GetInfo() map[string]interface{}

	// WriteInput writes data to the session's PTY
	WriteInput(data []byte) error

	// GetPTYMaster returns the PTY master file for direct I/O operations
	GetPTYMaster() *os.File
}

// ServerConfig holds configuration for terminal server
type ServerConfig struct {
	DefaultShell string
}
