package types

// TerminalInterface interface defines terminal integration operations
type TerminalInterface interface {
	Attach(sessionID string) error
	Detach(sessionID string) error
	SendInput(sessionID string, input []byte) error
	GetOutput(sessionID string) (<-chan []byte, error)
}

// MultiplexerIntegration interface defines multiplexer integration operations
type MultiplexerIntegration interface {
	CreateSession(name string, config MultiplexerConfig) error
	AttachToSession(sessionID string) error
	DetachFromSession(sessionID string) error
	ListSessions() ([]MultiplexerSession, error)
}

// ShellIntegration interface defines shell integration operations
type ShellIntegration interface {
	ExecuteCommand(command string) ([]byte, error)
	SetEnvironment(env map[string]string) error
	GetCompletion(input string) ([]string, error)
}
