package terminal

import "time"

// Client manages connection to a terminal server session
type Client interface {
	// Connect connects the client to the terminal session
	Connect() error

	// Disconnect disconnects the client from the terminal session
	Disconnect() error

	// Wait waits for the client session to end
	Wait() error

	// SendInput sends input data to the terminal session
	SendInput(data []byte) error

	// GetSessionInfo returns information about the connected session
	GetSessionInfo() (map[string]interface{}, error)

	// ResizeTerminal updates the terminal size
	ResizeTerminal(rows, cols uint16) error

	// ConnectAndWait is a convenience method that connects and waits
	ConnectAndWait() error

	// IsConnected returns true if client is connected
	IsConnected() bool

	// GetSessionID returns the session ID this client is connected to
	GetSessionID() string

	// GetClientID returns this client's ID
	GetClientID() string
}

// ClientConfig holds configuration for terminal client
type ClientConfig struct {
	SessionID string
	ClientID  string
	Timeout   time.Duration
}
