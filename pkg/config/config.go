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
	Projects       []Project      `yaml:"projects"`
	Aliases        Aliases        `yaml:"aliases"`
	GlobalSettings GlobalSettings `yaml:"global_settings"`
}

// Project groups stages under a project name.
type Project struct {
	Name   string  `yaml:"name"`
	Stages []Stage `yaml:"stages"`
}

// Stage represents a deployment or build stage within a project.
type Stage struct {
	Name       string      `yaml:"name"`
	Developers []Developer `yaml:"developers"`
}

// Developer lists sessions they are responsible for.
type Developer struct {
	Name     string    `yaml:"name"`
	Sessions []Session `yaml:"sessions"`
}

// Session defines a set of steps (commands) to run, optionally in parallel.
type Session struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Steps       []Step `yaml:"steps"`
	Parallel    bool   `yaml:"parallel"`
}

// Step is a single shell or SSH command to execute within a session.
type Step struct {
	Command string `yaml:"command"`
}

// Aliases maps friendly names to full session identifiers or endpoints.
type Aliases struct {
	MyAppProd string `yaml:"myapp-prod"`
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
