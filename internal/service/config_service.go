package service

import (
	"fmt"
	"hama-shell/pkg/types"
	"path/filepath"
)

// ConfigService provides configuration management operations
type ConfigService struct {
	configLoader    types.ConfigLoader
	configValidator types.ConfigValidator
	aliasManager    types.AliasManager
	config          *types.Config
}

// NewConfigService creates a new configuration service
func NewConfigService(loader types.ConfigLoader, validator types.ConfigValidator, aliasManager types.AliasManager) *ConfigService {
	return &ConfigService{
		configLoader:    loader,
		configValidator: validator,
		aliasManager:    aliasManager,
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
func (s *ConfigService) GetConfig() *types.Config {
	return s.config
}

// ResolveSessionPath resolves a session path or alias to a full session path
func (s *ConfigService) ResolveSessionPath(pathOrAlias string) (string, error) {
	// Try to resolve as alias first
	if resolved, err := s.aliasManager.Resolve(pathOrAlias); err == nil {
		return resolved, nil
	}

	// If not an alias, return as-is (assume it's a path)
	return pathOrAlias, nil
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

// AddAlias adds a new alias
func (s *ConfigService) AddAlias(alias, sessionPath string) error {
	return s.aliasManager.Add(alias, sessionPath)
}

// RemoveAlias removes an alias
func (s *ConfigService) RemoveAlias(alias string) error {
	return s.aliasManager.Remove(alias)
}

// ListAliases returns all configured aliases
func (s *ConfigService) ListAliases() (map[string]string, error) {
	return s.aliasManager.List()
}
