package service

import (
	"fmt"
	"hama-shell/pkg/types"
)

// ConnectionService provides connection management operations
type ConnectionService struct {
	connectionManager types.ConnectionManager
	tunnelManager     types.TunnelManager
	healthMonitor     types.HealthMonitor
}

// NewConnectionService creates a new connection service
func NewConnectionService(
	connectionManager types.ConnectionManager,
	tunnelManager types.TunnelManager,
	healthMonitor types.HealthMonitor,
) *ConnectionService {
	return &ConnectionService{
		connectionManager: connectionManager,
		tunnelManager:     tunnelManager,
		healthMonitor:     healthMonitor,
	}
}

// Connect establishes a connection using the provided configuration
func (s *ConnectionService) Connect(config types.ConnectionConfig) (types.Connection, error) {
	connection, err := s.connectionManager.Connect(config)
	if err != nil {
		return types.Connection{}, fmt.Errorf("failed to establish connection: %w", err)
	}

	// Set up tunnels if configured
	if len(config.Tunnels) > 0 {
		for _, tunnelConfig := range config.Tunnels {
			if _, err := s.tunnelManager.CreateTunnel(tunnelConfig); err != nil {
				// Log error but don't fail the connection
				fmt.Printf("Warning: failed to create tunnel %s: %v\n", tunnelConfig.Name, err)
			}
		}
	}

	// Start health monitoring
	if _, err := s.healthMonitor.Monitor(connection.ID); err != nil {
		fmt.Printf("Warning: failed to start health monitoring for connection %s: %v\n", connection.ID, err)
	}

	return connection, nil
}

// Disconnect closes a connection
func (s *ConnectionService) Disconnect(connectionID string) error {
	// Stop health monitoring
	if err := s.healthMonitor.StopMonitoring(connectionID); err != nil {
		fmt.Printf("Warning: failed to stop health monitoring: %v\n", err)
	}

	// Close any associated tunnels
	tunnels, err := s.tunnelManager.ListTunnels()
	if err == nil {
		for _, tunnel := range tunnels {
			// Close tunnels associated with this connection
			if err := s.tunnelManager.CloseTunnel(tunnel.ID); err != nil {
				fmt.Printf("Warning: failed to close tunnel %s: %v\n", tunnel.ID, err)
			}
		}
	}

	// Disconnect the connection
	if err := s.connectionManager.Disconnect(connectionID); err != nil {
		return fmt.Errorf("failed to disconnect: %w", err)
	}

	return nil
}

// GetConnectionStatus returns the status of a connection
func (s *ConnectionService) GetConnectionStatus(connectionID string) (types.ConnectionStatus, error) {
	status, err := s.connectionManager.GetStatus(connectionID)
	if err != nil {
		return "", fmt.Errorf("failed to get connection status: %w", err)
	}
	return status, nil
}

// ListConnections returns all connections
func (s *ConnectionService) ListConnections() ([]types.Connection, error) {
	connections, err := s.connectionManager.List()
	if err != nil {
		return nil, fmt.Errorf("failed to list connections: %w", err)
	}
	return connections, nil
}

// CreateTunnel creates a new tunnel
func (s *ConnectionService) CreateTunnel(config types.TunnelConfig) (types.Tunnel, error) {
	tunnel, err := s.tunnelManager.CreateTunnel(config)
	if err != nil {
		return types.Tunnel{}, fmt.Errorf("failed to create tunnel: %w", err)
	}
	return tunnel, nil
}

// CloseTunnel closes an existing tunnel
func (s *ConnectionService) CloseTunnel(tunnelID string) error {
	if err := s.tunnelManager.CloseTunnel(tunnelID); err != nil {
		return fmt.Errorf("failed to close tunnel: %w", err)
	}
	return nil
}

// ListTunnels returns all active tunnels
func (s *ConnectionService) ListTunnels() ([]types.Tunnel, error) {
	tunnels, err := s.tunnelManager.ListTunnels()
	if err != nil {
		return nil, fmt.Errorf("failed to list tunnels: %w", err)
	}
	return tunnels, nil
}

// CheckHealth checks the health of a connection
func (s *ConnectionService) CheckHealth(connectionID string) (types.HealthStatus, error) {
	health, err := s.healthMonitor.CheckHealth(connectionID)
	if err != nil {
		return "", fmt.Errorf("failed to check connection health: %w", err)
	}
	return health, nil
}

// MonitorHealth starts monitoring the health of a connection
func (s *ConnectionService) MonitorHealth(connectionID string) (<-chan types.HealthStatus, error) {
	healthChan, err := s.healthMonitor.Monitor(connectionID)
	if err != nil {
		return nil, fmt.Errorf("failed to start health monitoring: %w", err)
	}
	return healthChan, nil
}

// StopHealthMonitoring stops health monitoring for a connection
func (s *ConnectionService) StopHealthMonitoring(connectionID string) error {
	if err := s.healthMonitor.StopMonitoring(connectionID); err != nil {
		return fmt.Errorf("failed to stop health monitoring: %w", err)
	}
	return nil
}
