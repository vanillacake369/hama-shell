package executor

import (
	"fmt"
	"time"
)

// SupervisorManagerInterface defines the interface for supervisor management
type SupervisorManagerInterface interface {
	createSupervisor(key string, segments []CommandSegment) (*SessionGroup, error)
}

// SignalManagerInterface defines the interface for signal management  
type SignalManagerInterface interface {
	manageSupervisor(session *SessionGroup)
	gracefulShutdown(session *SessionGroup, timeout time.Duration) error
	shutdownAllSessions(timeout time.Duration) error
}

// SSHManagerInterface defines the interface for SSH management
type SSHManagerInterface interface {
	executeSSHWithPTY(sshCmd, password string, remoteCmds []string) error
}

// CommandParserInterface defines the interface for command parsing
type CommandParserInterface interface {
	parseCommandSegments(commands []string) []CommandSegment
}

// Executor defines the interface for process execution and management
type Executor interface {

	// RunSequence runs a sequence of commands in a single shell session
	RunSequence(key string, commands []string) error

	// StopAll terminates all running processes
	StopAll() error

	// StopByKey terminates all processes associated with the given key
	StopByKey(key string) error

	// GetStatus returns the current status of all processes
	GetStatus() map[string][]*ProcessInfo
}

// ProcessInfo contains information about a running process
type ProcessInfo struct {
	Command   string
	PID       int
	StartTime int64 // Unix timestamp
	Key       string
}

// executor is the main implementation of the Executor interface using session/PGID architecture
type executor struct {
	registry         *SessionRegistry
	parser          CommandParserInterface
	supervisorMgr   SupervisorManagerInterface
	signalMgr       SignalManagerInterface
	sshMgr          SSHManagerInterface
}

// New creates a new Executor instance
func New() Executor {
	registry := NewSessionRegistry()
	
	return &executor{
		registry:        registry,
		parser:          NewCommandParser(),
		supervisorMgr:   NewSupervisorManager(),
		signalMgr:       NewSignalManager(registry),
		sshMgr:          NewSSHManager(),
	}
}

// RunSequence runs a sequence of commands in a single shell session
func (e *executor) RunSequence(key string, commands []string) error {
	if len(commands) == 0 {
		return fmt.Errorf("no commands provided")
	}

	// Check if session already exists
	if e.registry.hasSession(key) {
		return fmt.Errorf("session with key '%s' already exists", key)
	}

	// Parse commands into segments (SSH vs shell commands)
	segments := e.parser.parseCommandSegments(commands)
	if len(segments) == 0 {
		return fmt.Errorf("no valid command segments found")
	}

	// Create supervisor session with session/PGID setup
	session, err := e.supervisorMgr.createSupervisor(key, segments)
	if err != nil {
		return fmt.Errorf("failed to create supervisor session: %w", err)
	}

	// Register session for management
	if err := e.registry.registerSession(key, session); err != nil {
		return fmt.Errorf("failed to register session: %w", err)
	}

	// Start session management in background
	go e.signalMgr.manageSupervisor(session)

	return nil
}

// StopAll terminates all running processes
func (e *executor) StopAll() error {
	// Use signal manager to gracefully shutdown all sessions
	return e.signalMgr.shutdownAllSessions(30 * time.Second)
}

// StopByKey terminates all processes associated with the given key
func (e *executor) StopByKey(key string) error {
	// Get session from registry
	session, exists := e.registry.getSession(key)
	if !exists {
		return nil // No session for this key
	}

	// Use signal manager for graceful shutdown
	err := e.signalMgr.gracefulShutdown(session, 30*time.Second)
	
	// Remove from registry regardless of shutdown result
	e.registry.unregisterSession(key)
	
	return err
}

// GetStatus returns the current status of all processes
func (e *executor) GetStatus() map[string][]*ProcessInfo {
	status := make(map[string][]*ProcessInfo)
	
	// Get all active sessions
	sessions := e.registry.getAllSessions()
	
	for key, session := range sessions {
		if session.Supervisor != nil && session.Supervisor.Process != nil {
			// Create ProcessInfo for the supervisor process
			info := &ProcessInfo{
				Command:   fmt.Sprintf("Session supervisor (%d segments)", len(session.Segments)),
				PID:       session.Supervisor.Process.Pid,
				StartTime: session.StartTime.Unix(),
				Key:       key,
			}
			status[key] = []*ProcessInfo{info}
		}
	}
	
	return status
}
