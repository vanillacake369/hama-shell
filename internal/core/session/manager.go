package session

import (
	"fmt"
	"github.com/google/uuid"
	"hama-shell/pkg/types"
	"time"
)

// Manager implements the SessionManager interface
type Manager struct {
	state       types.SessionState
	persistence types.SessionPersistence
	sessions    map[string]*types.Session
}

// NewManager creates a new session manager
func NewManager(state types.SessionState, persistence types.SessionPersistence) *Manager {
	return &Manager{
		state:       state,
		persistence: persistence,
		sessions:    make(map[string]*types.Session),
	}
}

// Create creates a new session from the provided configuration
func (m *Manager) Create(config types.SessionConfig) (*types.Session, error) {
	sessionID := uuid.New().String()
	now := time.Now()

	session := &types.Session{
		ID:          sessionID,
		Name:        config.Name,
		ProjectPath: fmt.Sprintf("%s.%s", config.Name, sessionID[:8]),
		Status:      types.SessionStatusPending,
		Config:      config,
		CreatedAt:   now,
		UpdatedAt:   now,
		Metadata:    make(map[string]string),
	}

	// Validate session configuration
	if err := m.validateSessionConfig(config); err != nil {
		return nil, fmt.Errorf("invalid session configuration: %w", err)
	}

	// Store session in memory
	m.sessions[sessionID] = session

	// Persist session state
	if err := m.state.Save(session); err != nil {
		delete(m.sessions, sessionID)
		return nil, fmt.Errorf("failed to save session state: %w", err)
	}

	// Store session for long-term persistence
	if err := m.persistence.Store(session); err != nil {
		// Log warning but don't fail - state is saved
		fmt.Printf("Warning: failed to persist session %s: %v\n", sessionID, err)
	}

	return session, nil
}

// Start starts a session by ID
func (m *Manager) Start(sessionID string) error {
	session, exists := m.sessions[sessionID]
	if !exists {
		// Try to load from state
		loadedSession, err := m.state.Load(sessionID)
		if err != nil {
			return fmt.Errorf("session not found: %s", sessionID)
		}
		session = loadedSession
		m.sessions[sessionID] = session
	}

	// Check if session is already running
	if session.Status == types.SessionStatusActive || session.Status == types.SessionStatusStarting {
		return fmt.Errorf("session %s is already running or starting", sessionID)
	}

	// Update status to starting
	session.Status = types.SessionStatusStarting
	session.UpdatedAt = time.Now()

	if err := m.updateSessionState(session); err != nil {
		return fmt.Errorf("failed to update session state: %w", err)
	}

	// Execute session commands (simplified - actual implementation would be more complex)
	go func() {
		if err := m.executeSession(session); err != nil {
			session.Status = types.SessionStatusFailed
			session.UpdatedAt = time.Now()
			m.updateSessionState(session)
			return
		}

		// Mark as active
		now := time.Now()
		session.Status = types.SessionStatusActive
		session.UpdatedAt = now
		session.StartedAt = &now
		m.updateSessionState(session)
	}()

	return nil
}

// Stop stops a running session
func (m *Manager) Stop(sessionID string) error {
	session, exists := m.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	// Check if session can be stopped
	if session.Status != types.SessionStatusActive {
		return fmt.Errorf("session %s is not running", sessionID)
	}

	// Update status to stopping
	session.Status = types.SessionStatusStopping
	session.UpdatedAt = time.Now()

	if err := m.updateSessionState(session); err != nil {
		return fmt.Errorf("failed to update session state: %w", err)
	}

	// Stop session processes (simplified)
	go func() {
		m.stopSession(session)

		// Mark as stopped
		now := time.Now()
		session.Status = types.SessionStatusStopped
		session.UpdatedAt = now
		session.StoppedAt = &now
		m.updateSessionState(session)
	}()

	return nil
}

// GetStatus returns the current status of a session
func (m *Manager) GetStatus(sessionID string) (types.SessionStatus, error) {
	session, exists := m.sessions[sessionID]
	if !exists {
		// Try to load from state
		loadedSession, err := m.state.Load(sessionID)
		if err != nil {
			return "", fmt.Errorf("session not found: %s", sessionID)
		}
		session = loadedSession
		m.sessions[sessionID] = session
	}

	return session.Status, nil
}

// List returns all sessions
func (m *Manager) List() ([]*types.Session, error) {
	sessions := make([]*types.Session, 0, len(m.sessions))
	for _, session := range m.sessions {
		sessions = append(sessions, session)
	}
	return sessions, nil
}

// validateSessionConfig validates the session configuration
func (m *Manager) validateSessionConfig(config types.SessionConfig) error {
	if config.Name == "" {
		return fmt.Errorf("session name is required")
	}

	if len(config.Commands) == 0 {
		return fmt.Errorf("at least one command is required")
	}

	for i, cmd := range config.Commands {
		if cmd.Command == "" {
			return fmt.Errorf("command %d is empty", i)
		}
	}

	return nil
}

// updateSessionState updates the session state
func (m *Manager) updateSessionState(session *types.Session) error {
	if err := m.state.Save(session); err != nil {
		return err
	}

	// Also update persistence (best effort)
	if err := m.persistence.Store(session); err != nil {
		fmt.Printf("Warning: failed to persist session %s: %v\n", session.ID, err)
	}

	return nil
}

// executeSession executes the session commands (simplified implementation)
func (m *Manager) executeSession(session *types.Session) error {
	// This is a simplified implementation
	// In a real implementation, this would:
	// 1. Set up the execution environment
	// 2. Execute each command in sequence or parallel
	// 3. Handle command failures and retries
	// 4. Manage process lifecycle

	for _, cmd := range session.Config.Commands {
		fmt.Printf("Executing command: %s %v\n", cmd.Command, cmd.Args)
		// Simulate command execution
		time.Sleep(100 * time.Millisecond)
	}

	return nil
}

// stopSession stops the session processes (simplified implementation)
func (m *Manager) stopSession(session *types.Session) {
	// This is a simplified implementation
	// In a real implementation, this would:
	// 1. Send termination signals to running processes
	// 2. Wait for graceful shutdown
	// 3. Force kill if necessary
	// 4. Clean up resources

	fmt.Printf("Stopping session: %s\n", session.ID)
	time.Sleep(100 * time.Millisecond)
}
