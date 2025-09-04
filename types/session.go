package types

import (
	"time"
)

// Session represents a managed process session
type Session struct {
	ID        string        `json:"id"`
	Status    SessionStatus `json:"status"`
	PID       int           `json:"pid"`
	Commands  []string      `json:"commands"`
	StartTime time.Time     `json:"start_time"`
	EndTime   *time.Time    `json:"end_time,omitempty"`
	TTYDevice string        `json:"tty_device,omitempty"`
}

// SessionStatus represents the current state of a session
type SessionStatus string

const (
	StatusRunning SessionStatus = "running"
	StatusStopped SessionStatus = "stopped"
	StatusFailed  SessionStatus = "failed"
	StatusPending SessionStatus = "pending"
)

// SessionManager interface for managing sessions
type SessionManager interface {
	// List returns all sessions
	List() ([]Session, error)

	// Get returns a specific session by ID
	Get(id string) (*Session, error)

	// Create starts a new session
	Create(id string, cmd string, args []string, workDir string) (*Session, error)

	// Attach connects to a session's TTY
	Attach(id string) error

	// Detach disconnects from a session's TTY
	Detach(id string) error

	// Kill terminates a session
	Kill(id string) error

	// GetCommands returns registered commands for a session
	GetCommands(id string) ([]string, error)

	// GetLogs returns the log file path for a session
	GetLogs(id string) (string, error)

	// Restart restarts a stopped session
	Restart(id string) error

	// Update updates session configuration
	Update(id string, updates map[string]interface{}) error
}
