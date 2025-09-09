package infra

import (
	"hama-shell/internal/core/terminal"
	"hama-shell/internal/session/model"
	"time"
)

// SessionManager handles session operations
type SessionManager struct {
	server terminal.Server
}

// NewSessionManager creates a new SessionManager instance
func NewSessionManager() *SessionManager {
	return &SessionManager{
		server: terminal.NewTerminalServer(),
	}
}

// ListSessions returns list of sessions based on filter
func (sm *SessionManager) ListSessions(filter model.SessionFilter) ([]model.SessionInfo, error) {
	// Get all sessions from terminal server
	sessions := sm.server.ListSessions()

	var result []model.SessionInfo
	for _, session := range sessions {
		info := session.GetInfo()

		sessionInfo := model.SessionInfo{
			ID:        info["id"].(string),
			StartTime: info["start_time"].(time.Time),
		}

		// Set status
		if running, ok := info["running"].(bool); ok && running {
			sessionInfo.Status = "running"
		} else {
			sessionInfo.Status = "stopped"
		}

		// Apply filter
		if filter.Status != "" && sessionInfo.Status != filter.Status {
			continue
		}

		if !filter.ShowAll && sessionInfo.Status == "stopped" {
			continue
		}

		result = append(result, sessionInfo)
	}

	return result, nil
}
