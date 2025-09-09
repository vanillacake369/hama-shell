package terminal

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/creack/pty"
)

// terminalServer implements Server interface
type terminalServer struct {
	mu       sync.RWMutex
	sessions map[string]*ptySession
	ctx      context.Context
	cancel   context.CancelFunc
	config   ServerConfig
}

// ptySession implements Session interface
type ptySession struct {
	id        string
	cmd       *exec.Cmd
	ptyMaster *os.File
	mu        sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc
	startTime time.Time
}

// NewTerminalServer creates a new terminal server
func NewTerminalServer() Server {
	return NewTerminalServerWithConfig(ServerConfig{})
}

// NewTerminalServerWithConfig creates a new terminal server with configuration
func NewTerminalServerWithConfig(config ServerConfig) Server {
	if config.DefaultShell == "" {
		config.DefaultShell = "/bin/bash"
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &terminalServer{
		sessions: make(map[string]*ptySession),
		ctx:      ctx,
		cancel:   cancel,
		config:   config,
	}
}

// CreateSession creates a new PTY session
func (ts *terminalServer) CreateSession(sessionID, shell string, args []string) (Session, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	// Check if session already exists
	if _, exists := ts.sessions[sessionID]; exists {
		return nil, fmt.Errorf("session %s already exists", sessionID)
	}

	// Use default shell if not provided
	if shell == "" {
		shell = os.Getenv("SHELL")
		if shell == "" {
			shell = "/bin/bash"
		}
	}

	// Create command
	cmd := exec.Command(shell, args...)
	ptyMaster, err := pty.Start(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to start PTY: %w", err)
	}

	// Set default terminal size
	if err := pty.Setsize(ptyMaster, &pty.Winsize{
		Rows: 24,
		Cols: 80,
	}); err != nil {
		fmt.Printf("Warning: failed to set PTY size: %v\n", err)
	}

	ctx, cancel := context.WithCancel(ts.ctx)
	session := &ptySession{
		id:        sessionID,
		cmd:       cmd,
		ptyMaster: ptyMaster,
		ctx:       ctx,
		cancel:    cancel,
		startTime: time.Now(),
	}

	ts.sessions[sessionID] = session

	// Start session management
	go ts.manageSession(session)

	return session, nil
}

// KillSession terminates a session
func (ts *terminalServer) KillSession(sessionID string) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	session, exists := ts.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session %s not found", sessionID)
	}

	// Cancel session context
	session.cancel()

	// Close PTY and kill process
	if session.ptyMaster != nil {
		session.ptyMaster.Close()
	}
	if session.cmd != nil && session.cmd.Process != nil {
		session.cmd.Process.Kill()
	}

	delete(ts.sessions, sessionID)
	return nil
}

// ResizeSession updates the terminal size for a session
func (ts *terminalServer) ResizeSession(sessionID string, rows, cols uint16) error {
	ts.mu.RLock()
	session, exists := ts.sessions[sessionID]
	ts.mu.RUnlock()

	if !exists {
		return fmt.Errorf("session %s not found", sessionID)
	}

	if session.ptyMaster == nil {
		return fmt.Errorf("session %s has no active PTY", sessionID)
	}

	return pty.Setsize(session.ptyMaster, &pty.Winsize{
		Rows: rows,
		Cols: cols,
	})
}

// GetSession returns session information
func (ts *terminalServer) GetSession(sessionID string) (Session, error) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	session, exists := ts.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	return session, nil
}

// ListSessions returns all active sessions
func (ts *terminalServer) ListSessions() map[string]Session {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	// Create copy to avoid race conditions
	result := make(map[string]Session)
	for id, session := range ts.sessions {
		result[id] = session
	}
	return result
}

// Shutdown gracefully shuts down the server
func (ts *terminalServer) Shutdown() error {
	ts.cancel()

	ts.mu.Lock()
	defer ts.mu.Unlock()

	// Kill all sessions
	for sessionID := range ts.sessions {
		if err := ts.KillSession(sessionID); err != nil {
			fmt.Printf("Warning: failed to kill session %s: %v\n", sessionID, err)
		}
	}

	return nil
}

// manageSession handles the lifecycle of a PTY session
func (ts *terminalServer) manageSession(session *ptySession) {
	defer func() {
		// Cleanup on session end
		if session.ptyMaster != nil {
			session.ptyMaster.Close()
		}
		if session.cmd != nil && session.cmd.Process != nil {
			session.cmd.Process.Kill()
		}
	}()

	// Wait for command to finish or context cancellation
	cmdDone := make(chan error, 1)
	go func() {
		cmdDone <- session.cmd.Wait()
	}()

	select {
	case err := <-cmdDone:
		if err != nil {
			fmt.Printf("Session %s ended with error: %v\n", session.id, err)
		} else {
			fmt.Printf("Session %s ended normally\n", session.id)
		}
	case <-session.ctx.Done():
		fmt.Printf("Session %s cancelled\n", session.id)
	}

	// Clean up session from server
	ts.mu.Lock()
	delete(ts.sessions, session.id)
	ts.mu.Unlock()
}

// Session interface implementation for ptySession

// GetID returns the session ID
func (s *ptySession) GetID() string {
	return s.id
}

// GetStartTime returns when the session was started
func (s *ptySession) GetStartTime() time.Time {
	return s.startTime
}

// GetPID returns the process ID if available
func (s *ptySession) GetPID() int {
	if s.cmd != nil && s.cmd.Process != nil {
		return s.cmd.Process.Pid
	}
	return 0
}

// IsRunning returns true if the session is currently running
func (s *ptySession) IsRunning() bool {
	return s.cmd != nil && s.cmd.ProcessState == nil
}

// WriteInput writes data to the session's PTY
func (s *ptySession) WriteInput(data []byte) error {
	if s.ptyMaster == nil {
		return fmt.Errorf("session has no active PTY")
	}
	_, err := s.ptyMaster.Write(data)
	return err
}

// GetPTYMaster returns the PTY master file for direct I/O operations
func (s *ptySession) GetPTYMaster() *os.File {
	return s.ptyMaster
}

// GetInfo returns session information as a map
func (s *ptySession) GetInfo() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	info := map[string]interface{}{
		"id":         s.id,
		"start_time": s.startTime,
		"running":    s.IsRunning(),
		"pid":        s.GetPID(),
	}

	return info
}
