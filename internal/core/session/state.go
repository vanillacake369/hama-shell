package session

import (
	"fmt"
	"hama-shell/pkg/types"
	"sync"
)

// State implements the SessionState interface for in-memory state management
type State struct {
	mu       sync.RWMutex
	sessions map[string]*types.Session
}

// NewState creates a new session state manager
func NewState() *State {
	return &State{
		sessions: make(map[string]*types.Session),
	}
}

// Save saves a session to the state store
func (s *State) Save(session *types.Session) error {
	if session == nil {
		return fmt.Errorf("session cannot be nil")
	}

	if session.ID == "" {
		return fmt.Errorf("session ID cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Create a copy to avoid external modifications
	sessionCopy := *session
	s.sessions[session.ID] = &sessionCopy

	return nil
}

// Load loads a session from the state store
func (s *State) Load(sessionID string) (*types.Session, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("session ID cannot be empty")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	// Return a copy to avoid external modifications
	sessionCopy := *session
	return &sessionCopy, nil
}

// Delete removes a session from the state store
func (s *State) Delete(sessionID string) error {
	if sessionID == "" {
		return fmt.Errorf("session ID cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.sessions[sessionID]; !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	delete(s.sessions, sessionID)
	return nil
}

// GetAll returns all sessions in the state store
func (s *State) GetAll() ([]*types.Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sessions := make([]*types.Session, 0, len(s.sessions))
	for _, session := range s.sessions {
		// Return copies to avoid external modifications
		sessionCopy := *session
		sessions = append(sessions, &sessionCopy)
	}

	return sessions, nil
}

// GetByStatus returns sessions with the specified status
func (s *State) GetByStatus(status types.SessionStatus) ([]*types.Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var sessions []*types.Session
	for _, session := range s.sessions {
		if session.Status == status {
			// Return a copy to avoid external modifications
			sessionCopy := *session
			sessions = append(sessions, &sessionCopy)
		}
	}

	return sessions, nil
}

// Count returns the total number of sessions
func (s *State) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.sessions)
}

// CountByStatus returns the number of sessions with the specified status
func (s *State) CountByStatus(status types.SessionStatus) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	count := 0
	for _, session := range s.sessions {
		if session.Status == status {
			count++
		}
	}

	return count
}

// Clear removes all sessions from the state store
func (s *State) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sessions = make(map[string]*types.Session)
	return nil
}

// Exists checks if a session exists in the state store
func (s *State) Exists(sessionID string) bool {
	if sessionID == "" {
		return false
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	_, exists := s.sessions[sessionID]
	return exists
}
