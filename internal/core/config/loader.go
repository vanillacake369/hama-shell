package config

import (
	"fmt"
	"hama-shell/pkg/types"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Loader implements the ConfigLoader interface
type Loader struct {
	configPath    string
	searchPaths   []string
	currentConfig *Config
}

// NewLoader creates a new configuration loader
func NewLoader(configPath string) *Loader {
	loader := &Loader{
		configPath: configPath,
	}

	// Set up default search paths
	loader.setupSearchPaths()

	return loader
}

// Load loads configuration from the specified path
func (l *Loader) Load(path string) (*Config, error) {
	if path == "" {
		path = l.findConfigFile()
		if path == "" {
			return nil, fmt.Errorf("no configuration file found in search paths")
		}
	}

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("configuration file not found: %s", path)
	}

	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read configuration file: %w", err)
	}

	// Parse YAML
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML config: %w", err)
	}

	return &config, nil
}

// Reload reloads the configuration from the current path
func (l *Loader) Reload() (*Config, error) {
	if l.configPath == "" {
		return nil, fmt.Errorf("no configuration path set")
	}

	return l.Load(l.configPath)
}

// GetCurrentConfig returns the currently loaded configuration
func (l *Loader) GetCurrentConfig() *Config {
	return l.currentConfig
}

// GetConfigPath returns the current configuration file path
func (l *Loader) GetConfigPath() string {
	return l.configPath
}

// setupSearchPaths sets up default configuration search paths
func (l *Loader) setupSearchPaths() {
	l.searchPaths = []string{}

	// Add explicit config path if provided
	if l.configPath != "" {
		l.searchPaths = append(l.searchPaths, l.configPath)
	}

	// Add current directory
	l.searchPaths = append(l.searchPaths, ".hama-shell.yaml")
	l.searchPaths = append(l.searchPaths, ".hama-shell.yml")
	l.searchPaths = append(l.searchPaths, "hama-shell.yaml")
	l.searchPaths = append(l.searchPaths, "hama-shell.yml")

	// Add home directory
	if homeDir, err := os.UserHomeDir(); err == nil {
		l.searchPaths = append(l.searchPaths, filepath.Join(homeDir, ".hama-shell.yaml"))
		l.searchPaths = append(l.searchPaths, filepath.Join(homeDir, ".hama-shell.yml"))
		l.searchPaths = append(l.searchPaths, filepath.Join(homeDir, ".config", "hama-shell", "config.yaml"))
		l.searchPaths = append(l.searchPaths, filepath.Join(homeDir, ".config", "hama-shell", "config.yml"))
	}

	// Add system-wide paths
	l.searchPaths = append(l.searchPaths, "/etc/hama-shell/config.yaml")
	l.searchPaths = append(l.searchPaths, "/etc/hama-shell/config.yml")
}

// findConfigFile finds the first existing configuration file in search paths
func (l *Loader) findConfigFile() string {
	for _, path := range l.searchPaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return ""
}

// processProject processes and validates a project configuration
func (l *Loader) processProject(project *Project, projectName string) error {

	// Initialize stages if nil
	if project.Stages == nil {
		project.Stages = make(map[string]Stage)
	}

	return nil
}

// processDeveloper processes and validates a developer configuration
func (l *Loader) processDeveloper(developer *Developer, developerName string) error {
	// Set developer name if not set
	if developer.Name == "" {
		developer.Name = developerName
	}

	// Initialize sessions if nil
	if developer.Sessions == nil {
		developer.Sessions = make(map[string]types.SessionConfig)
	}

	// Process each session
	for sessionName, sessionConfig := range developer.Sessions {
		if err := l.processSessionConfig(&sessionConfig, sessionName); err != nil {
			return fmt.Errorf("error processing session %s: %w", sessionName, err)
		}
		developer.Sessions[sessionName] = sessionConfig
	}

	return nil
}

// processSessionConfig processes and validates a session configuration
func (l *Loader) processSessionConfig(sessionConfig *types.SessionConfig, sessionName string) error {
	// Set session name if not set
	if sessionConfig.Name == "" {
		sessionConfig.Name = sessionName
	}

	// Initialize maps if nil
	if sessionConfig.Environment == nil {
		sessionConfig.Environment = make(map[string]string)
	}
	if sessionConfig.Options == nil {
		sessionConfig.Options = make(map[string]interface{})
	}

	// Process commands
	for i := range sessionConfig.Commands {
		cmd := &sessionConfig.Commands[i]
		if cmd.Environment == nil {
			cmd.Environment = make(map[string]string)
		}
	}

	return nil
}

// GetSearchPaths returns the configuration search paths
func (l *Loader) GetSearchPaths() []string {
	return l.searchPaths
}
