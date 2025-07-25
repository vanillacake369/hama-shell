package types

// Config represents the top-level YAML configuration
// pulled from file into Go structs.
type Config struct {
	Projects       map[string]Project `yaml:"projects"`
	GlobalSettings GlobalSettings     `yaml:"global_settings"`
}

// Project groups stages under a project name.
type Project struct {
	Description string           `yaml:"description"`
	Stages      map[string]Stage `yaml:"stages"`
}

// Stage represents a deployment or build stage within a project.
type Stage struct {
	Description string             `yaml:"description,omitempty"`
	Services    map[string]Service `yaml:"services"`
}

// Service defines connection details for a specific service.
type Service struct {
	Description string   `yaml:"description"`
	Commands    []string `yaml:"commands"`
}

// GlobalSettings configures retry logic, timeouts, and auto-restart behavior.
type GlobalSettings struct {
	Retries     int  `yaml:"retries"`
	Timeout     int  `yaml:"timeout"`
	AutoRestart bool `yaml:"auto_restart"`
}

// Developer represents a developer's configuration
type Developer struct {
	Name     string                   `json:"name" yaml:"name"`
	Email    string                   `json:"email,omitempty" yaml:"email,omitempty"`
	Sessions map[string]SessionConfig `json:"sessions" yaml:"sessions"`
	Settings map[string]interface{}   `json:"settings,omitempty" yaml:"settings,omitempty"`
}

// ConfigSettings represents global configuration settings
type ConfigSettings struct {
	DefaultShell       string            `json:"default_shell,omitempty" yaml:"default_shell,omitempty"`
	DefaultMultiplexer string            `json:"default_multiplexer,omitempty" yaml:"default_multiplexer,omitempty"`
	LogLevel           string            `json:"log_level,omitempty" yaml:"log_level,omitempty"`
	StateDir           string            `json:"state_dir,omitempty" yaml:"state_dir,omitempty"`
	ConfigDir          string            `json:"config_dir,omitempty" yaml:"config_dir,omitempty"`
	Environment        map[string]string `json:"environment,omitempty" yaml:"environment,omitempty"`
}

// TerminalConfig defines terminal-specific configuration
type TerminalConfig struct {
	Multiplexer  string                 `json:"multiplexer,omitempty" yaml:"multiplexer,omitempty"`
	Shell        string                 `json:"shell,omitempty" yaml:"shell,omitempty"`
	Layout       string                 `json:"layout,omitempty" yaml:"layout,omitempty"`
	WindowName   string                 `json:"window_name,omitempty" yaml:"window_name,omitempty"`
	SessionName  string                 `json:"session_name,omitempty" yaml:"session_name,omitempty"`
	AutoAttach   bool                   `json:"auto_attach,omitempty" yaml:"auto_attach,omitempty"`
	DetachOnExit bool                   `json:"detach_on_exit,omitempty" yaml:"detach_on_exit,omitempty"`
	Options      map[string]interface{} `json:"options,omitempty" yaml:"options,omitempty"`
}

// MultiplexerConfig defines multiplexer-specific configuration
type MultiplexerConfig struct {
	Type        string                 `json:"type" yaml:"type"`
	SessionName string                 `json:"session_name" yaml:"session_name"`
	WindowName  string                 `json:"window_name,omitempty" yaml:"window_name,omitempty"`
	Layout      string                 `json:"layout,omitempty" yaml:"layout,omitempty"`
	Options     map[string]interface{} `json:"options,omitempty" yaml:"options,omitempty"`
}

// MultiplexerSession represents a multiplexer session
type MultiplexerSession struct {
	ID          string `json:"id" yaml:"id"`
	Name        string `json:"name" yaml:"name"`
	Type        string `json:"type" yaml:"type"`
	WindowCount int    `json:"window_count" yaml:"window_count"`
	Attached    bool   `json:"attached" yaml:"attached"`
}

// ConfigLoader interface defines configuration loading operations
type ConfigLoader interface {
	Load(path string) (*Config, error)
	LoadFromBytes(data []byte) (*Config, error)
	Reload() (*Config, error)
}

// ConfigValidator interface defines configuration validation operations
type ConfigValidator interface {
	Validate(config *Config) error
	ValidateSession(session SessionConfig) error
}
