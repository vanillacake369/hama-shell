package executor

import (
	"fmt"
	"sync"
)

// SessionRegistry manages the lifecycle of session groups
type SessionRegistry struct {
	sessions sync.Map // key: string, value: *SessionGroup
	mu       sync.RWMutex
}

// NewSessionRegistry creates a new session registry
func NewSessionRegistry() *SessionRegistry {
	return &SessionRegistry{}
}

// registerSession adds a session group to the registry
func (sr *SessionRegistry) registerSession(key string, session *SessionGroup) error {
	if key == "" {
		return fmt.Errorf("session key cannot be empty")
	}
	
	if session == nil {
		return fmt.Errorf("session cannot be nil")
	}
	
	sr.mu.Lock()
	defer sr.mu.Unlock()
	
	// Check if session already exists
	if _, exists := sr.sessions.Load(key); exists {
		return fmt.Errorf("session with key '%s' already exists", key)
	}
	
	sr.sessions.Store(key, session)
	return nil
}

// getSession retrieves a session group by key
func (sr *SessionRegistry) getSession(key string) (*SessionGroup, bool) {
	if key == "" {
		return nil, false
	}
	
	sr.mu.RLock()
	defer sr.mu.RUnlock()
	
	value, exists := sr.sessions.Load(key)
	if !exists {
		return nil, false
	}
	
	session, ok := value.(*SessionGroup)
	if !ok {
		// This should never happen, but handle gracefully
		sr.sessions.Delete(key)
		return nil, false
	}
	
	return session, true
}

// unregisterSession removes a session group from the registry
func (sr *SessionRegistry) unregisterSession(key string) (*SessionGroup, error) {
	if key == "" {
		return nil, fmt.Errorf("session key cannot be empty")
	}
	
	sr.mu.Lock()
	defer sr.mu.Unlock()
	
	value, exists := sr.sessions.Load(key)
	if !exists {
		return nil, fmt.Errorf("session with key '%s' not found", key)
	}
	
	sr.sessions.Delete(key)
	
	session, ok := value.(*SessionGroup)
	if !ok {
		return nil, fmt.Errorf("invalid session type for key '%s'", key)
	}
	
	return session, nil
}

// listSessions returns all active session keys
func (sr *SessionRegistry) listSessions() []string {
	sr.mu.RLock()
	defer sr.mu.RUnlock()
	
	var keys []string
	sr.sessions.Range(func(key, value interface{}) bool {
		if keyStr, ok := key.(string); ok {
			keys = append(keys, keyStr)
		}
		return true
	})
	
	return keys
}

// getAllSessions returns all active sessions
func (sr *SessionRegistry) getAllSessions() map[string]*SessionGroup {
	sr.mu.RLock()
	defer sr.mu.RUnlock()
	
	result := make(map[string]*SessionGroup)
	sr.sessions.Range(func(key, value interface{}) bool {
		if keyStr, ok := key.(string); ok {
			if session, ok := value.(*SessionGroup); ok {
				result[keyStr] = session
			}
		}
		return true
	})
	
	return result
}

// sessionCount returns the number of active sessions
func (sr *SessionRegistry) sessionCount() int {
	count := 0
	sr.sessions.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}

// hasSession checks if a session exists
func (sr *SessionRegistry) hasSession(key string) bool {
	_, exists := sr.getSession(key)
	return exists
}

// clearAllSessions removes all sessions (for cleanup/testing)
func (sr *SessionRegistry) clearAllSessions() {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	
	sr.sessions.Range(func(key, value interface{}) bool {
		sr.sessions.Delete(key)
		return true
	})
}