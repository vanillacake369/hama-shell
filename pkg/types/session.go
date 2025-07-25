package types

import "time"

// SessionStatus represents the current state of a session
type SessionStatus string

const (
	SessionStatusPending  SessionStatus = "pending"
	SessionStatusStarting SessionStatus = "starting"
	SessionStatusActive   SessionStatus = "active"
	SessionStatusStopping SessionStatus = "stopping"
	SessionStatusStopped  SessionStatus = "stopped"
	SessionStatusFailed   SessionStatus = "failed"
)

// Session represents a session instance
type Session struct {
	ID          string            `json:"id" yaml:"id"`
	Name        string            `json:"name" yaml:"name"`
	ProjectPath string            `json:"project_path" yaml:"project_path"`
	Status      SessionStatus     `json:"status" yaml:"status"`
	Config      SessionConfig     `json:"config" yaml:"config"`
	CreatedAt   time.Time         `json:"created_at" yaml:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at" yaml:"updated_at"`
	StartedAt   *time.Time        `json:"started_at,omitempty" yaml:"started_at,omitempty"`
	StoppedAt   *time.Time        `json:"stopped_at,omitempty" yaml:"stopped_at,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// SessionConfig defines the configuration for a session
type SessionConfig struct {
	Name        string                 `json:"name" yaml:"name"`
	Description string                 `json:"description,omitempty" yaml:"description,omitempty"`
	Commands    []Command              `json:"commands" yaml:"commands"`
	Environment map[string]string      `json:"environment,omitempty" yaml:"environment,omitempty"`
	WorkingDir  string                 `json:"working_dir,omitempty" yaml:"working_dir,omitempty"`
	Terminal    TerminalConfig         `json:"terminal,omitempty" yaml:"terminal,omitempty"`
	Connection  ConnectionConfig       `json:"connection,omitempty" yaml:"connection,omitempty"`
	Options     map[string]interface{} `json:"options,omitempty" yaml:"options,omitempty"`
}

// Command represents a command to execute in a session
type Command struct {
	Name        string            `json:"name" yaml:"name"`
	Command     string            `json:"command" yaml:"command"`
	Args        []string          `json:"args,omitempty" yaml:"args,omitempty"`
	Environment map[string]string `json:"environment,omitempty" yaml:"environment,omitempty"`
	WorkingDir  string            `json:"working_dir,omitempty" yaml:"working_dir,omitempty"`
	Timeout     time.Duration     `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	Retry       int               `json:"retry,omitempty" yaml:"retry,omitempty"`
}

// SessionManager interface defines session management operations
type SessionManager interface {
	Create(config SessionConfig) (*Session, error)
	Start(sessionID string) error
	Stop(sessionID string) error
	GetStatus(sessionID string) (SessionStatus, error)
	List() ([]*Session, error)
}

// SessionState interface defines session state management
type SessionState interface {
	Save(session *Session) error
	Load(sessionID string) (*Session, error)
	Delete(sessionID string) error
}

// SessionPersistence interface defines session persistence operations
type SessionPersistence interface {
	Store(session *Session) error
	Retrieve(sessionID string) (*Session, error)
	Remove(sessionID string) error
}
