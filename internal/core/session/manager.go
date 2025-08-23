package session

import (
	"fmt"

	"hama-shell/internal/core/config"
	"hama-shell/internal/core/executor"
)

// ConfigService defines the interface for configuration operations
type ConfigService interface {
	Load(configPath string) (*config.Config, error)
	List(cfg *config.Config) []string
	ResolveTarget(target string, cfg *config.Config) (*config.Service, error)
}

// Manager handles session lifecycle management
type Manager struct {
	executor  executor.Executor
	configSvc ConfigService
}

// NewManager creates a new session manager instance
func NewManager(exec executor.Executor, cfgSvc ConfigService) *Manager {
	if exec == nil {
		exec = executor.New()
	}
	if cfgSvc == nil {
		cfgSvc = config.NewService()
	}
	return &Manager{
		executor:  exec,
		configSvc: cfgSvc,
	}
}

// Start initiates a session for the given target (background mode)
func (m *Manager) Start(target string, configPath string) error {
	return m.StartWithMode(target, configPath, executor.ExecutionModeBackground)
}

// StartWithMode initiates a session for the given target with specified execution mode
func (m *Manager) StartWithMode(target string, configPath string, mode executor.ExecutionMode) error {
	// Load configuration
	cfg, err := m.configSvc.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Resolve target to service configuration
	svc, err := m.configSvc.ResolveTarget(target, cfg)
	if err != nil {
		return fmt.Errorf("failed to resolve target: %w", err)
	}

	// Execute commands using RunSequenceWithMode
	if err := m.executor.RunSequenceWithMode(target, svc.Commands, mode); err != nil {
		return fmt.Errorf("failed to execute commands: %w", err)
	}

	return nil
}

// Stop terminates a session by target
func (m *Manager) Stop(target string) error {
	// Use StopByKey to stop the session
	if err := m.executor.StopByKey(target); err != nil {
		return fmt.Errorf("failed to stop session %s: %w", target, err)
	}
	return nil
}

// StopAll terminates all active sessions
func (m *Manager) StopAll() error {
	return m.executor.StopAll()
}

// GetStatus returns the status of all sessions
func (m *Manager) GetStatus() map[string][]*executor.ProcessInfo {
	return m.executor.GetStatus()
}

// GetTargetStatus returns the status of a specific target
func (m *Manager) GetTargetStatus(target string) []*executor.ProcessInfo {
	allStatus := m.executor.GetStatus()
	if processes, exists := allStatus[target]; exists {
		return processes
	}
	return []*executor.ProcessInfo{}
}