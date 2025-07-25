package service

import (
	"fmt"
	"hama-shell/pkg/types"
)

// TerminalService provides terminal integration operations
type TerminalService struct {
	terminalInterface      types.TerminalInterface
	multiplexerIntegration types.MultiplexerIntegration
	shellIntegration       types.ShellIntegration
}

// NewTerminalService creates a new terminal service
func NewTerminalService(
	terminalInterface types.TerminalInterface,
	multiplexerIntegration types.MultiplexerIntegration,
	shellIntegration types.ShellIntegration,
) *TerminalService {
	return &TerminalService{
		terminalInterface:      terminalInterface,
		multiplexerIntegration: multiplexerIntegration,
		shellIntegration:       shellIntegration,
	}
}

// AttachToSession attaches to a session terminal
func (s *TerminalService) AttachToSession(sessionID string) error {
	if err := s.terminalInterface.Attach(sessionID); err != nil {
		return fmt.Errorf("failed to attach to session terminal: %w", err)
	}
	return nil
}

// DetachFromSession detaches from a session terminal
func (s *TerminalService) DetachFromSession(sessionID string) error {
	if err := s.terminalInterface.Detach(sessionID); err != nil {
		return fmt.Errorf("failed to detach from session terminal: %w", err)
	}
	return nil
}

// SendInput sends input to a session terminal
func (s *TerminalService) SendInput(sessionID string, input []byte) error {
	if err := s.terminalInterface.SendInput(sessionID, input); err != nil {
		return fmt.Errorf("failed to send input to session: %w", err)
	}
	return nil
}

// GetOutput gets output from a session terminal
func (s *TerminalService) GetOutput(sessionID string) (<-chan []byte, error) {
	outputChan, err := s.terminalInterface.GetOutput(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session output: %w", err)
	}
	return outputChan, nil
}

// CreateMultiplexerSession creates a new multiplexer session
func (s *TerminalService) CreateMultiplexerSession(name string, config types.MultiplexerConfig) error {
	if err := s.multiplexerIntegration.CreateSession(name, config); err != nil {
		return fmt.Errorf("failed to create multiplexer session: %w", err)
	}
	return nil
}

// AttachToMultiplexerSession attaches to an existing multiplexer session
func (s *TerminalService) AttachToMultiplexerSession(sessionID string) error {
	if err := s.multiplexerIntegration.AttachToSession(sessionID); err != nil {
		return fmt.Errorf("failed to attach to multiplexer session: %w", err)
	}
	return nil
}

// DetachFromMultiplexerSession detaches from a multiplexer session
func (s *TerminalService) DetachFromMultiplexerSession(sessionID string) error {
	if err := s.multiplexerIntegration.DetachFromSession(sessionID); err != nil {
		return fmt.Errorf("failed to detach from multiplexer session: %w", err)
	}
	return nil
}

// ListMultiplexerSessions lists all multiplexer sessions
func (s *TerminalService) ListMultiplexerSessions() ([]types.MultiplexerSession, error) {
	sessions, err := s.multiplexerIntegration.ListSessions()
	if err != nil {
		return nil, fmt.Errorf("failed to list multiplexer sessions: %w", err)
	}
	return sessions, nil
}

// ExecuteShellCommand executes a command in the shell
func (s *TerminalService) ExecuteShellCommand(command string) ([]byte, error) {
	output, err := s.shellIntegration.ExecuteCommand(command)
	if err != nil {
		return nil, fmt.Errorf("failed to execute shell command: %w", err)
	}
	return output, nil
}

// SetShellEnvironment sets environment variables in the shell
func (s *TerminalService) SetShellEnvironment(env map[string]string) error {
	if err := s.shellIntegration.SetEnvironment(env); err != nil {
		return fmt.Errorf("failed to set shell environment: %w", err)
	}
	return nil
}

// GetShellCompletion gets shell completion suggestions
func (s *TerminalService) GetShellCompletion(input string) ([]string, error) {
	completions, err := s.shellIntegration.GetCompletion(input)
	if err != nil {
		return nil, fmt.Errorf("failed to get shell completion: %w", err)
	}
	return completions, nil
}

// SetupSessionTerminal sets up terminal integration for a session
func (s *TerminalService) SetupSessionTerminal(sessionID string, config types.TerminalConfig) error {
	// Create multiplexer session if configured
	if config.Multiplexer != "" {
		multiplexerConfig := types.MultiplexerConfig{
			Type:        config.Multiplexer,
			SessionName: config.SessionName,
			WindowName:  config.WindowName,
			Layout:      config.Layout,
			Options:     config.Options,
		}

		if err := s.CreateMultiplexerSession(sessionID, multiplexerConfig); err != nil {
			return fmt.Errorf("failed to setup multiplexer session: %w", err)
		}

		// Auto-attach if configured
		if config.AutoAttach {
			if err := s.AttachToMultiplexerSession(sessionID); err != nil {
				return fmt.Errorf("failed to auto-attach to multiplexer session: %w", err)
			}
		}
	}

	return nil
}

// TeardownSessionTerminal cleans up terminal integration for a session
func (s *TerminalService) TeardownSessionTerminal(sessionID string, config types.TerminalConfig) error {
	// Detach from multiplexer if needed
	if config.Multiplexer != "" && !config.DetachOnExit {
		if err := s.DetachFromMultiplexerSession(sessionID); err != nil {
			// Log error but don't fail teardown
			fmt.Printf("Warning: failed to detach from multiplexer session: %v\n", err)
		}
	}

	// Detach from terminal
	if err := s.DetachFromSession(sessionID); err != nil {
		return fmt.Errorf("failed to detach from session terminal: %w", err)
	}

	return nil
}
