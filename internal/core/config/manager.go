package config

import (
	"fmt"
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

// Service represents a service configuration with commands
type Service struct {
	Commands []string `yaml:"commands"`
}

// Project represents a project configuration with services
type Project struct {
	Services map[string]*Service `yaml:"services"`
}

// Config represents the main configuration structure
type Config struct {
	Projects map[string]*Project `yaml:"projects"`
}

// ConfigManager manages configuration with memory caching and file persistence
type ConfigManager struct {
	config   *Config
	filePath string
	mu       sync.RWMutex
}

var (
	instance *ConfigManager
	once     sync.Once
)

// GetInstance returns the singleton instance of ConfigManager
func GetInstance() *ConfigManager {
	once.Do(func() {
		home := os.Getenv("HOME")
		instance = &ConfigManager{
			filePath: home + "/hama-shell.yaml",
		}
	})
	return instance
}

// Load reads configuration from file into memory
func (cm *ConfigManager) Load() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	data, err := os.ReadFile(cm.filePath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	if err != nil {
		// Initialize with empty config if file doesn't exist
		cm.config = &Config{
			Projects: make(map[string]*Project),
		}
		return nil
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	// Initialize map if nil
	if config.Projects == nil {
		config.Projects = make(map[string]*Project)
	}

	cm.config = &config
	return nil
}

// Save writes the current configuration to file
func (cm *ConfigManager) Save() error {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if cm.config == nil {
		return fmt.Errorf("no configuration to save")
	}

	data, err := yaml.Marshal(cm.config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(cm.filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetConfig returns a copy of the current configuration
func (cm *ConfigManager) GetConfig() *Config {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if cm.config == nil {
		return nil
	}

	// Return a copy to prevent external modification
	configCopy := *cm.config
	return &configCopy
}

// AddProject adds a new project to the configuration
func (cm *ConfigManager) AddProject(projectName string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.config == nil {
		cm.config = &Config{
			Projects: make(map[string]*Project),
		}
	}

	// Check if project already exists
	if _, exists := cm.config.Projects[projectName]; exists {
		return fmt.Errorf("project '%s' already exists", projectName)
	}

	cm.config.Projects[projectName] = &Project{
		Services: make(map[string]*Service),
	}
	return nil
}

// AddService adds a service to an existing project
func (cm *ConfigManager) AddService(projectName, serviceName string, commands []string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.config == nil {
		return fmt.Errorf("configuration not loaded")
	}

	project, exists := cm.config.Projects[projectName]
	if !exists {
		return fmt.Errorf("project '%s' not found", projectName)
	}

	if project.Services == nil {
		project.Services = make(map[string]*Service)
	}

	// Check if service already exists
	if _, exists := project.Services[serviceName]; exists {
		return fmt.Errorf("service '%s' already exists in project '%s'", serviceName, projectName)
	}

	project.Services[serviceName] = &Service{
		Commands: commands,
	}
	return nil
}

// AppendToService appends commands to an existing service
func (cm *ConfigManager) AppendToService(projectName, serviceName string, commands []string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.config == nil {
		return fmt.Errorf("configuration not loaded")
	}

	project, exists := cm.config.Projects[projectName]
	if !exists {
		return fmt.Errorf("project '%s' not found", projectName)
	}

	service, exists := project.Services[serviceName]
	if !exists {
		return fmt.Errorf("service '%s' not found in project '%s'", serviceName, projectName)
	}

	// Append new commands to existing ones
	service.Commands = append(service.Commands, commands...)
	return nil
}

// UpdateService updates an existing service
func (cm *ConfigManager) UpdateService(projectName, serviceName string, commands []string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.config == nil {
		return fmt.Errorf("configuration not loaded")
	}

	project, exists := cm.config.Projects[projectName]
	if !exists {
		return fmt.Errorf("project '%s' not found", projectName)
	}

	service, exists := project.Services[serviceName]
	if !exists {
		return fmt.Errorf("service '%s' not found in project '%s'", serviceName, projectName)
	}

	service.Commands = commands
	return nil
}

// DeleteProject removes a project from configuration
func (cm *ConfigManager) DeleteProject(projectName string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.config == nil {
		return fmt.Errorf("configuration not loaded")
	}

	if _, exists := cm.config.Projects[projectName]; !exists {
		return fmt.Errorf("project '%s' not found", projectName)
	}

	delete(cm.config.Projects, projectName)
	return nil
}

// DeleteService removes a service from a project
func (cm *ConfigManager) DeleteService(projectName, serviceName string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.config == nil {
		return fmt.Errorf("configuration not loaded")
	}

	project, exists := cm.config.Projects[projectName]
	if !exists {
		return fmt.Errorf("project '%s' not found", projectName)
	}

	if _, exists := project.Services[serviceName]; !exists {
		return fmt.Errorf("service '%s' not found in project '%s'", serviceName, projectName)
	}

	delete(project.Services, serviceName)
	return nil
}

// FileExists checks if the configuration file exists
func (cm *ConfigManager) FileExists() bool {
	_, err := os.Stat(cm.filePath)
	return err == nil
}

// GetFilePath returns the configuration file path
func (cm *ConfigManager) GetFilePath() string {
	return cm.filePath
}
