package executor

import (
	"fmt"
	"strings"
	"time"
)

// SupervisorManagerInterface defines the interface for supervisor management
type SupervisorManagerInterface interface {
	createSupervisor(key string, segments []CommandSegment, mode ExecutionMode) (*SessionGroup, error)
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

	// RunSequence runs a sequence of commands in a single shell session (background mode)
	RunSequence(key string, commands []string) error

	// RunSequenceWithMode runs a sequence of commands with specified execution mode
	RunSequenceWithMode(key string, commands []string, mode ExecutionMode) error

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

// RunSequence runs a sequence of commands in a single shell session (background mode)
func (e *executor) RunSequence(key string, commands []string) error {
	return e.RunSequenceWithMode(key, commands, ExecutionModeBackground)
}

// RunSequenceWithMode runs a sequence of commands with specified execution mode
func (e *executor) RunSequenceWithMode(key string, commands []string, mode ExecutionMode) error {
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

	// Create supervisor session with session/PGID setup and execution mode
	session, err := e.supervisorMgr.createSupervisor(key, segments, mode)
	if err != nil {
		// Check if this is a process creation restriction
		if isProcessCreationError(err) {
			return e.executeDirectMode(key, commands)
		}
		return fmt.Errorf("failed to create supervisor session: %w", err)
	}

	// Handle execution mode-specific behavior
	switch mode {
	case ExecutionModeForeground:
		// Foreground: wait for completion (blocking)
		return e.waitForSessionCompletion(session)
	case ExecutionModeBackground:
		// Background: register session and return immediately
		if err := e.registry.registerSession(key, session); err != nil {
			return fmt.Errorf("failed to register session: %w", err)
		}
		// Start session management in background
		go e.signalMgr.manageSupervisor(session)
		return nil
	default:
		return fmt.Errorf("unsupported execution mode: %s", mode)
	}
}

// StopAll terminates all running processes
func (e *executor) StopAll() error {
	// Get all session keys before shutdown
	sessions := e.registry.getAllSessions()
	keys := make([]string, 0, len(sessions))
	for key := range sessions {
		keys = append(keys, key)
	}
	
	// Use signal manager to gracefully shutdown all sessions
	err := e.signalMgr.shutdownAllSessions(30 * time.Second)
	
	// Clean up registry regardless of shutdown result
	for _, key := range keys {
		e.registry.unregisterSession(key)
	}
	
	return err
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

// isProcessCreationError checks if an error is due to process creation restrictions
func isProcessCreationError(err error) bool {
	errStr := err.Error()
	return strings.Contains(errStr, "operation not permitted") ||
		   strings.Contains(errStr, "process creation not permitted") ||
		   strings.Contains(errStr, "fork/exec") ||
		   strings.Contains(errStr, "permission denied")
}

// executeDirectMode provides a fallback execution mode for restricted environments
func (e *executor) executeDirectMode(key string, commands []string) error {
	fmt.Printf("⚠️  Running in direct mode (process creation restricted)\n")
	fmt.Printf("📋 Commands for %s:\n", key)
	
	for i, cmd := range commands {
		if strings.TrimSpace(cmd) == "" {
			continue
		}
		
		// Check if it's an SSH command
		parser := NewCommandParser()
		if parser.isSSHCommand(cmd) {
			fmt.Printf("   %d. [SSH] %s\n", i+1, cmd)
			fmt.Printf("      ⏸️  SSH commands require interactive execution\n")
		} else {
			fmt.Printf("   %d. [SHELL] %s\n", i+1, cmd)
			fmt.Printf("      ℹ️  Would execute in session/PGID environment\n")
		}
	}
	
	fmt.Printf("\n💡 To run these commands:\n")
	fmt.Printf("   1. Execute them manually in your terminal\n")
	fmt.Printf("   2. Run HamaShell in a less restrictive environment\n")
	fmt.Printf("   3. Use a container or VM without process restrictions\n")
	
	return nil
}

// waitForSessionCompletion blocks until the session completes (for foreground mode)
func (e *executor) waitForSessionCompletion(session *SessionGroup) error {
	if session == nil || session.Supervisor == nil {
		return fmt.Errorf("invalid session for waiting")
	}
	
	// Wait for the supervisor process to complete
	err := session.Supervisor.Wait()
	
	// Signal completion
	close(session.Done)
	
	if err != nil {
		return fmt.Errorf("session completed with error: %w", err)
	}
	
	return nil
}
