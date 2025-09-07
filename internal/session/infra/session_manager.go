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

	// If no real sessions, return demo data for now
	// TODO: Remove this when terminal server is fully integrated
	if len(result) == 0 {
		result = sm.getDemoSessions(filter)
	}

	return result, nil
}

// getDemoSessions returns demo sessions for testing
func (sm *SessionManager) getDemoSessions(filter model.SessionFilter) []model.SessionInfo {
	demoSessions := []model.SessionInfo{
		{
			ID:        "web-server",
			Status:    "running",
			StartTime: time.Now().Add(-2 * time.Hour),
			Command:   "npm run dev",
		},
		{
			ID:        "db-backup",
			Status:    "running",
			StartTime: time.Now().Add(-5 * time.Hour),
			Command:   "pg_dump mydb > backup.sql",
		},
		{
			ID:        "worker-1",
			Status:    "stopped",
			StartTime: time.Now().Add(-8 * time.Hour),
			Command:   "python worker.py",
		},
	}

	var result []model.SessionInfo
	for _, session := range demoSessions {
		// Apply filter
		if filter.Status != "" && session.Status != filter.Status {
			continue
		}

		if !filter.ShowAll && session.Status == "stopped" {
			continue
		}

		result = append(result, session)
	}

	return result
}
