package config

import (
	"fmt"
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

// findConfigFile finds the first existing configuration file in search paths
func (l *Loader) findConfigFile() string {
	for _, path := range l.searchPaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return ""
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
