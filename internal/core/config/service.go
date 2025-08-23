package config

import (
	"fmt"
	"strings"
)

// ConfigService provides business operations for configuration management
type ConfigService struct {
	validator *Validator
}

// NewService creates a new config service instance
func NewService() *ConfigService {
	return &ConfigService{
		validator: NewValidator(),
	}
}

// Load and validate configuration from the specified path
func (s *ConfigService) Load(configPath string) (*Config, error) {
	if configPath == "" {
		return nil, fmt.Errorf("config path is required")
	}
	return s.validator.ParseAndValidate(configPath)
}

// List returns all available targets in project.stage.service format
func (s *ConfigService) List(cfg *Config) []string {
	if cfg == nil {
		return []string{}
	}

	var targets []string
	for projName, proj := range cfg.Projects {
		for stageName, stage := range proj.Stages {
			for svcName := range stage.Services {
				targets = append(targets, fmt.Sprintf("%s.%s.%s",
					projName, stageName, svcName))
			}
		}
	}
	return targets
}

// ResolveTarget converts project.stage.service notation to service configuration
func (s *ConfigService) ResolveTarget(target string, cfg *Config) (*Service, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}

	parts := strings.Split(target, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid target format: expected project.stage.service, got: %s", target)
	}

	projName, stageName, svcName := parts[0], parts[1], parts[2]

	proj, ok := cfg.Projects[projName]
	if !ok {
		return nil, fmt.Errorf("project not found: %s", projName)
	}

	stage, ok := proj.Stages[stageName]
	if !ok {
		return nil, fmt.Errorf("stage not found: %s.%s", projName, stageName)
	}

	svc, ok := stage.Services[svcName]
	if !ok {
		return nil, fmt.Errorf("service not found: %s.%s.%s", projName, stageName, svcName)
	}

	return &svc, nil
}