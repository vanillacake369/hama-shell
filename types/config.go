package types

import (
	"time"
)

// CommandConfig represents a command configuration
type CommandConfig struct {
	ID            string            `json:"id"`
	Command       string            `json:"command"`
	Args          []string          `json:"args,omitempty"`
	WorkingDir    string            `json:"working_dir,omitempty"`
	Env           map[string]string `json:"env,omitempty"`
	AutoStart     bool              `json:"auto_start"`
	RestartPolicy RestartPolicy     `json:"restart_policy"`
	MaxRetries    int               `json:"max_retries,omitempty"`
	RetryDelay    time.Duration     `json:"retry_delay,omitempty"`
	LogFile       string            `json:"log_file,omitempty"`
	Description   string            `json:"description,omitempty"`
	Tags          []string          `json:"tags,omitempty"`
}

// RestartPolicy defines how a session should be restarted
type RestartPolicy string

const (
	RestartNever     RestartPolicy = "never"
	RestartAlways    RestartPolicy = "always"
	RestartOnFailure RestartPolicy = "on-failure"
)

// Config represents the main configuration
type Config struct {
	Version  string          `json:"version"`
	Commands []CommandConfig `json:"commands"`
	Settings Settings        `json:"settings"`
}

// Settings represents global settings
type Settings struct {
	LogDir           string        `json:"log_dir"`
	DefaultShell     string        `json:"default_shell"`
	DefaultWorkDir   string        `json:"default_work_dir,omitempty"`
	AutoRestart      bool          `json:"auto_restart"`
	MaxRetries       int           `json:"max_retries"`
	RetryDelay       time.Duration `json:"retry_delay"`
	SessionTimeout   time.Duration `json:"session_timeout,omitempty"`
	EnableMetrics    bool          `json:"enable_metrics"`
	MetricsPort      int           `json:"metrics_port,omitempty"`
	TTYRows          int           `json:"tty_rows,omitempty"`
	TTYCols          int           `json:"tty_cols,omitempty"`
	HistoryLimit     int           `json:"history_limit,omitempty"`
	NotifyOnFailure  bool          `json:"notify_on_failure"`
	NotifyCommand    string        `json:"notify_command,omitempty"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		Version:  "1.0.0",
		Commands: []CommandConfig{},
		Settings: Settings{
			LogDir:          "/var/log/hama-shell",
			DefaultShell:    "/bin/bash",
			AutoRestart:     false,
			MaxRetries:      3,
			RetryDelay:      5 * time.Second,
			EnableMetrics:   false,
			TTYRows:         24,
			TTYCols:         80,
			HistoryLimit:    1000,
			NotifyOnFailure: false,
		},
	}
}

// ConfigManager interface for managing configuration
type ConfigManager interface {
	// Load loads configuration from file
	Load(path string) (*Config, error)
	
	// Save saves configuration to file
	Save(config *Config, path string) error
	
	// Create creates a new configuration file
	Create(path string) error
	
	// AddCommand adds a new command to configuration
	AddCommand(cmd CommandConfig) error
	
	// RemoveCommand removes a command from configuration
	RemoveCommand(id string) error
	
	// UpdateCommand updates an existing command
	UpdateCommand(id string, updates map[string]interface{}) error
	
	// GetCommand retrieves a command by ID
	GetCommand(id string) (*CommandConfig, error)
	
	// ListCommands lists all commands
	ListCommands() ([]CommandConfig, error)
	
	// GetConfigPath returns the current configuration file path
	GetConfigPath() string
	
	// Validate validates the configuration
	Validate(config *Config) error
}

// ConfigError represents configuration-related errors
type ConfigError struct {
	Type    string
	Message string
	Field   string
}

func (e ConfigError) Error() string {
	if e.Field != "" {
		return e.Type + ": " + e.Field + " - " + e.Message
	}
	return e.Type + ": " + e.Message
}