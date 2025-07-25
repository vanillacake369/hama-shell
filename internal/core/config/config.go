package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

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

// GetConfig reads and parses a YAML configuration file
func GetConfig(filePath string) (*Config, error) {
	// Resolve absolute path
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	// Read file
	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", absPath, err)
	}

	// Parse YAML
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML config: %w", err)
	}

	return &config, nil
}
