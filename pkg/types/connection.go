package types

import "time"

// ConnectionStatus represents the current state of a connection
type ConnectionStatus string

const (
	ConnectionStatusDisconnected ConnectionStatus = "disconnected"
	ConnectionStatusConnecting   ConnectionStatus = "connecting"
	ConnectionStatusConnected    ConnectionStatus = "connected"
	ConnectionStatusReconnecting ConnectionStatus = "reconnecting"
	ConnectionStatusFailed       ConnectionStatus = "failed"
)

// HealthStatus represents the health state of a connection
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusUnknown   HealthStatus = "unknown"
)

// Connection represents a connection instance
type Connection struct {
	ID         string           `json:"id" yaml:"id"`
	Name       string           `json:"name" yaml:"name"`
	Type       string           `json:"type" yaml:"type"`
	Status     ConnectionStatus `json:"status" yaml:"status"`
	Config     ConnectionConfig `json:"config" yaml:"config"`
	CreatedAt  time.Time        `json:"created_at" yaml:"created_at"`
	UpdatedAt  time.Time        `json:"updated_at" yaml:"updated_at"`
	LastHealth HealthStatus     `json:"last_health" yaml:"last_health"`
}

// ConnectionConfig defines the configuration for a connection
type ConnectionConfig struct {
	Type    string            `json:"type" yaml:"type"`
	Host    string            `json:"host" yaml:"host"`
	Port    int               `json:"port" yaml:"port"`
	User    string            `json:"user" yaml:"user"`
	SSH     *SSHConfig        `json:"ssh,omitempty" yaml:"ssh,omitempty"`
	Tunnels []TunnelConfig    `json:"tunnels,omitempty" yaml:"tunnels,omitempty"`
	Options map[string]string `json:"options,omitempty" yaml:"options,omitempty"`
	Timeout time.Duration     `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	Retry   int               `json:"retry,omitempty" yaml:"retry,omitempty"`
}

// SSHConfig defines SSH connection configuration
type SSHConfig struct {
	KeyPath            string        `json:"key_path,omitempty" yaml:"key_path,omitempty"`
	Password           string        `json:"password,omitempty" yaml:"password,omitempty"`
	Agent              bool          `json:"agent,omitempty" yaml:"agent,omitempty"`
	HostKeyCallback    string        `json:"host_key_callback,omitempty" yaml:"host_key_callback,omitempty"`
	Timeout            time.Duration `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	KeepAlive          time.Duration `json:"keep_alive,omitempty" yaml:"keep_alive,omitempty"`
	MaxRetries         int           `json:"max_retries,omitempty" yaml:"max_retries,omitempty"`
	StrictHostChecking bool          `json:"strict_host_checking,omitempty" yaml:"strict_host_checking,omitempty"`
}

// TunnelConfig defines tunnel configuration
type TunnelConfig struct {
	Name       string `json:"name" yaml:"name"`
	LocalHost  string `json:"local_host" yaml:"local_host"`
	LocalPort  int    `json:"local_port" yaml:"local_port"`
	RemoteHost string `json:"remote_host" yaml:"remote_host"`
	RemotePort int    `json:"remote_port" yaml:"remote_port"`
	Type       string `json:"type" yaml:"type"` // forward, reverse
}

// Tunnel represents an active tunnel instance
type Tunnel struct {
	ID     string       `json:"id" yaml:"id"`
	Config TunnelConfig `json:"config" yaml:"config"`
	Status string       `json:"status" yaml:"status"`
}

// ConnectionManager interface defines connection management operations
type ConnectionManager interface {
	Connect(config ConnectionConfig) (Connection, error)
	Disconnect(connectionID string) error
	GetStatus(connectionID string) (ConnectionStatus, error)
	List() ([]Connection, error)
}

// SSHClient interface defines SSH client operations
type SSHClient interface {
	Connect(host string, config SSHConfig) error
	Execute(command string) ([]byte, error)
	Disconnect() error
}

// TunnelManager interface defines tunnel management operations
type TunnelManager interface {
	CreateTunnel(config TunnelConfig) (Tunnel, error)
	CloseTunnel(tunnelID string) error
	ListTunnels() ([]Tunnel, error)
}

// HealthMonitor interface defines health monitoring operations
type HealthMonitor interface {
	Monitor(connectionID string) (<-chan HealthStatus, error)
	CheckHealth(connectionID string) (HealthStatus, error)
	StopMonitoring(connectionID string) error
}
