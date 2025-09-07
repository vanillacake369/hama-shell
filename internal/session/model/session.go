package model

import "time"

// SessionInfo represents information about a session
type SessionInfo struct {
	ID        string
	Status    string
	StartTime time.Time
	Command   string
}

// SessionFilter represents filtering options for sessions
type SessionFilter struct {
	ShowAll bool
	Status  string
}
