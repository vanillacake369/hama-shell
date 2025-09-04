package types

import (
	"time"
)

// SessionStatus represents the current state of a session
type SessionStatus string

const (
	StatusRunning SessionStatus = "running"
	StatusStopped SessionStatus = "stopped"
	StatusFailed  SessionStatus = "failed"
	StatusPending SessionStatus = "pending"
)

// Session represents a managed process session
type Session struct {
	ID          string            `json:"id"`
	Status      SessionStatus     `json:"status"`
	PID         int               `json:"pid"`
	Command     string            `json:"command"`
	Args        []string          `json:"args"`
	WorkingDir  string            `json:"working_dir"`
	StartTime   time.Time         `json:"start_time"`
	EndTime     *time.Time        `json:"end_time,omitempty"`
	LogFile     string            `json:"log_file"`
	Environment map[string]string `json:"environment"`
	ExitCode    *int              `json:"exit_code,omitempty"`
	TTYDevice   string            `json:"tty_device,omitempty"`
}

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

// SessionFilter for filtering session lists
type SessionFilter struct {
	Status   *SessionStatus
	Since    *time.Time
	Before   *time.Time
	Pattern  string
}

// SessionEvent represents events that occur during session lifecycle
type SessionEvent struct {
	SessionID string    `json:"session_id"`
	Type      EventType `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	Data      string    `json:"data,omitempty"`
}

// EventType represents different types of session events
type EventType string

const (
	EventStarted   EventType = "started"
	EventStopped   EventType = "stopped"
	EventFailed    EventType = "failed"
	EventRestarted EventType = "restarted"
	EventAttached  EventType = "attached"
	EventDetached  EventType = "detached"
)