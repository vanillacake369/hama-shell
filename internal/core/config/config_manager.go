package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/spf13/viper"
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

// ConfigManager interface for managing configuration
type ConfigManager interface {
	// GetConfig returns the current configuration
	GetConfig() *Config

	// Save writes the current configuration to file
	Save() error

	// FileExists checks if the configuration file exists
	FileExists() bool

	// GetFilePath returns the configuration file path
	GetFilePath() string

	// AddProject adds a new project to the configuration
	AddProject(projectName string) error

	// AddService adds a service to an existing project
	AddService(projectName, serviceName string, commands []string) error

	// AppendToService appends commands to an existing service
	AppendToService(projectName, serviceName string, commands []string) error
}

// configManager manages configuration using Viper (implementation)
type configManager struct {
	v        *viper.Viper
	mu       sync.RWMutex
	filePath string
}

var (
	instance ConfigManager
	once     sync.Once
)

// GetInstance returns the singleton instance of ConfigManager
func GetInstance() ConfigManager {
	once.Do(func() {
		instance = newConfigManager()
		instance.(*configManager).initialize()
	})
	return instance
}

// newConfigManager creates a new configManager instance
func newConfigManager() ConfigManager {
	home := os.Getenv("HOME")
	filePath := filepath.Join(home, "hama-shell.yaml")

	v := viper.New()
	v.SetConfigFile(filePath)
	v.SetConfigType("yaml")

	return &configManager{
		v:        v,
		filePath: filePath,
	}
}

// initialize sets up the configuration manager
func (cm *configManager) initialize() {
	// Initialize with empty config if file doesn't exist
	if !cm.FileExists() {
		cm.v.Set("projects", make(map[string]interface{}))
	} else {
		// Load file if it exists (once only)
		if err := cm.v.ReadInConfig(); err != nil {
			// Initialize with empty config even if there's an error
			cm.v.Set("projects", make(map[string]interface{}))
		}
	}

}

// Save writes the current configuration to file
func (cm *configManager) Save() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Create directory if it doesn't exist
	dir := filepath.Dir(cm.filePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Save file using Viper's WriteConfigAs
	return cm.v.WriteConfigAs(cm.filePath)
}

// GetConfig returns the current configuration
func (cm *configManager) GetConfig() *Config {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	var config Config
	if err := cm.v.Unmarshal(&config); err != nil {
		return &Config{
			Projects: make(map[string]*Project),
		}
	}

	if config.Projects == nil {
		config.Projects = make(map[string]*Project)
	}

	// Initialize nil Services maps
	for _, project := range config.Projects {
		if project.Services == nil {
			project.Services = make(map[string]*Service)
		}
	}

	return &config
}

// AddProject adds a new project to the configuration
func (cm *configManager) AddProject(projectName string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	projects := cm.v.GetStringMap("projects")
	if projects == nil {
		projects = make(map[string]interface{})
	}

	if _, exists := projects[projectName]; exists {
		return fmt.Errorf("project '%s' already exists", projectName)
	}

	projects[projectName] = map[string]interface{}{
		"services": make(map[string]interface{}),
	}

	cm.v.Set("projects", projects)
	return nil
}

// AddService adds a service to an existing project
func (cm *configManager) AddService(projectName, serviceName string, commands []string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	projectPath := fmt.Sprintf("projects.%s", projectName)
	if !cm.v.IsSet(projectPath) {
		return fmt.Errorf("project '%s' not found", projectName)
	}

	servicesPath := fmt.Sprintf("%s.services", projectPath)
	services := cm.v.GetStringMap(servicesPath)
	if services == nil {
		services = make(map[string]interface{})
	}

	if _, exists := services[serviceName]; exists {
		return fmt.Errorf("service '%s' already exists in project '%s'", serviceName, projectName)
	}

	services[serviceName] = map[string]interface{}{
		"commands": commands,
	}

	cm.v.Set(servicesPath, services)
	return nil
}

// AppendToService appends commands to an existing service
func (cm *configManager) AppendToService(projectName, serviceName string, commands []string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	commandsPath := fmt.Sprintf("projects.%s.services.%s.commands", projectName, serviceName)

	if !cm.v.IsSet(commandsPath) {
		return fmt.Errorf("service '%s' not found in project '%s'", serviceName, projectName)
	}

	existingCommands := cm.v.GetStringSlice(commandsPath)
	existingCommands = append(existingCommands, commands...)
	cm.v.Set(commandsPath, existingCommands)

	return nil
}

// FileExists checks if the configuration file exists
func (cm *configManager) FileExists() bool {
	_, err := os.Stat(cm.filePath)
	return err == nil
}

// GetFilePath returns the configuration file path
func (cm *configManager) GetFilePath() string {
	return cm.filePath
}
