package terminal

import (
	"context"
	"fmt"
	"time"
)

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

// terminalClient implements Client interface
type terminalClient struct {
	sessionID  string
	server     Server
	clientID   string
	ctx        context.Context
	cancel     context.CancelFunc
	isAttached bool
}


// NewTerminalClient creates a new terminal client
func NewTerminalClient(server Server, config ClientConfig) Client {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.ClientID == "" {
		config.ClientID = fmt.Sprintf("client-%d", time.Now().Unix())
	}

	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)

	return &terminalClient{
		sessionID: config.SessionID,
		server:    server,
		clientID:  config.ClientID,
		ctx:       ctx,
		cancel:    cancel,
	}
}

// Connect connects the client to the terminal session
func (tc *terminalClient) Connect() error {
	if tc.isAttached {
		return fmt.Errorf("client already connected to session %s", tc.sessionID)
	}

	// Check if session exists
	_, err := tc.server.GetSession(tc.sessionID)
	if err != nil {
		return fmt.Errorf("failed to find session: %w", err)
	}

	tc.isAttached = true
	return nil
}

// Disconnect disconnects the client from the terminal session
func (tc *terminalClient) Disconnect() error {
	if !tc.isAttached {
		return nil
	}



	// Cancel context
	tc.cancel()
	tc.isAttached = false

	return nil
}

// Wait waits for the client session to end
func (tc *terminalClient) Wait() error {
	if !tc.isAttached {
		return fmt.Errorf("client not connected")
	}

	// Wait for context to be done (timeout or cancellation)
	<-tc.ctx.Done()

	// Check if it was timeout or normal cancellation
	if tc.ctx.Err() == context.DeadlineExceeded {
		fmt.Println("\nSession timed out")
		return tc.ctx.Err()
	}

	return nil
}

// SendInput sends input data to the terminal session
func (tc *terminalClient) SendInput(data []byte) error {
	if !tc.isAttached {
		return fmt.Errorf("client not connected")
	}

	session, err := tc.server.GetSession(tc.sessionID)
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
	}

	return session.WriteInput(data)
}

// GetSessionInfo returns information about the connected session
func (tc *terminalClient) GetSessionInfo() (map[string]interface{}, error) {
	session, err := tc.server.GetSession(tc.sessionID)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	return session.GetInfo(), nil
}

// ResizeTerminal updates the terminal size
func (tc *terminalClient) ResizeTerminal(rows, cols uint16) error {
	if !tc.isAttached {
		return fmt.Errorf("client not connected")
	}

	return tc.server.ResizeSession(tc.sessionID, rows, cols)
}

// ConnectAndWait is a convenience method that connects and waits
func (tc *terminalClient) ConnectAndWait() error {
	if err := tc.Connect(); err != nil {
		return err
	}
	defer tc.Disconnect()
	return tc.Wait()
}

// IsConnected returns true if client is connected
func (tc *terminalClient) IsConnected() bool {
	return tc.isAttached
}

// GetSessionID returns the session ID this client is connected to
func (tc *terminalClient) GetSessionID() string {
	return tc.sessionID
}

// GetClientID returns this client's ID
func (tc *terminalClient) GetClientID() string {
	return tc.clientID
}

