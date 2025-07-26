package service

import (
	"fmt"
	"hama-shell/internal/core/config"
	"path/filepath"
)

// ConfigService provides configuration management operations
type ConfigService struct {
	configLoader    config.ConfigLoader
	configValidator config.ConfigValidator
	config          *config.Config
}

// NewConfigService creates a new configuration service
func NewConfigService(loader config.ConfigLoader, validator config.ConfigValidator) *ConfigService {
	return &ConfigService{
		configLoader:    loader,
		configValidator: validator,
	}
}

// LoadConfig loads the configuration from the specified path
func (s *ConfigService) LoadConfig(path string) error {
	config, err := s.configLoader.Load(path)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := s.configValidator.Validate(config); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	s.config = config
	return nil
}

// GetConfig returns the current configuration
func (s *ConfigService) GetConfig() *config.Config {
	return s.config
}

// ValidateConfig validates the current configuration
func (s *ConfigService) ValidateConfig() error {
	if s.config == nil {
		return fmt.Errorf("configuration not loaded")
	}

	return s.configValidator.Validate(s.config)
}

// ValidateConfigFile validates a configuration file
func (s *ConfigService) ValidateConfigFile(configPath string) error {
	config, err := s.configLoader.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config file: %w", err)
	}

	return s.configValidator.Validate(config)
}

// ReloadConfig reloads the configuration
func (s *ConfigService) ReloadConfig() error {
	config, err := s.configLoader.Reload()
	if err != nil {
		return fmt.Errorf("failed to reload config: %w", err)
	}

	if err := s.configValidator.Validate(config); err != nil {
		return fmt.Errorf("config validation failed after reload: %w", err)
	}

	s.config = config
	return nil
}

// GetConfigPaths returns configuration file paths
func (s *ConfigService) GetConfigPaths() map[string]string {
	paths := make(map[string]string)

	// Add common config paths
	homeDir, _ := filepath.Abs("~")
	paths["home"] = filepath.Join(homeDir, ".hama-shell.yaml")
	paths["local"] = filepath.Join(".", ".hama-shell.yaml")
	paths["system"] = "/etc/hama-shell/config.yaml"

	return paths
}
