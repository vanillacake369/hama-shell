package service

import (
	"fmt"
	"hama-shell/pkg/types"
)

// SessionService provides session management operations
type SessionService struct {
	sessionManager types.SessionManager
	configService  *ConfigService
}

// NewSessionService creates a new session service
func NewSessionService(sessionManager types.SessionManager, configService *ConfigService) *SessionService {
	return &SessionService{
		sessionManager: sessionManager,
		configService:  configService,
	}
}

// StartSession starts a session by path or alias
func (s *SessionService) StartSession(pathOrAlias string) error {
	// Resolve alias if needed
	sessionPath, err := s.configService.ResolveSessionPath(pathOrAlias)
	if err != nil {
		return fmt.Errorf("failed to resolve session path: %w", err)
	}

	// Get session configuration
	sessionConfig, err := s.configService.GetSessionConfig(sessionPath)
	if err != nil {
		return fmt.Errorf("failed to get session config: %w", err)
	}

	// Create session
	session, err := s.sessionManager.Create(*sessionConfig)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	// Start session
	if err := s.sessionManager.Start(session.ID); err != nil {
		return fmt.Errorf("failed to start session: %w", err)
	}

	return nil
}

// StopSession stops a session by ID or path
func (s *SessionService) StopSession(sessionID string) error {
	if err := s.sessionManager.Stop(sessionID); err != nil {
		return fmt.Errorf("failed to stop session: %w", err)
	}
	return nil
}

// GetSessionStatus returns the status of a session
func (s *SessionService) GetSessionStatus(sessionID string) (types.SessionStatus, error) {
	status, err := s.sessionManager.GetStatus(sessionID)
	if err != nil {
		return "", fmt.Errorf("failed to get session status: %w", err)
	}
	return status, nil
}

// ListSessions returns all sessions
func (s *SessionService) ListSessions() ([]*types.Session, error) {
	sessions, err := s.sessionManager.List()
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}
	return sessions, nil
}

// StartAllSessions starts all configured sessions
func (s *SessionService) StartAllSessions() error {
	sessionPaths, err := s.configService.GetAllSessionPaths()
	if err != nil {
		return fmt.Errorf("failed to get session paths: %w", err)
	}

	var errors []error
	for _, path := range sessionPaths {
		if err := s.StartSession(path); err != nil {
			errors = append(errors, fmt.Errorf("failed to start session %s: %w", path, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("some sessions failed to start: %v", errors)
	}

	return nil
}

// StopAllSessions stops all running sessions
func (s *SessionService) StopAllSessions() error {
	sessions, err := s.ListSessions()
	if err != nil {
		return fmt.Errorf("failed to list sessions: %w", err)
	}

	var errors []error
	for _, session := range sessions {
		if session.Status == types.SessionStatusActive {
			if err := s.StopSession(session.ID); err != nil {
				errors = append(errors, fmt.Errorf("failed to stop session %s: %w", session.ID, err))
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("some sessions failed to stop: %v", errors)
	}

	return nil
}
