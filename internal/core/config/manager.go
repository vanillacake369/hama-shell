package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
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

// ConfigManager manages configuration using Viper
type ConfigManager struct {
	v        *viper.Viper
	mu       sync.RWMutex
	filePath string

	// change callbacks (optional)
	onChangeCallbacks []func()
}

var (
	instance *ConfigManager
	once     sync.Once
)

// GetInstance returns the singleton instance of ConfigManager
func GetInstance() *ConfigManager {
	once.Do(func() {
		instance = newConfigManager()
		instance.initialize()
	})
	return instance
}

// newConfigManager creates a new ConfigManager instance
func newConfigManager() *ConfigManager {
	home := os.Getenv("HOME")
	if home == "" {
		home = os.Getenv("USERPROFILE") // Windows support
	}
	filePath := filepath.Join(home, "hama-shell.yaml")

	v := viper.New()
	v.SetConfigFile(filePath)
	v.SetConfigType("yaml")

	return &ConfigManager{
		v:                 v,
		filePath:          filePath,
		onChangeCallbacks: make([]func(), 0),
	}
}

// initialize sets up the configuration manager
func (cm *ConfigManager) initialize() {
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

	// Set up file change detection (optional)
	cm.v.WatchConfig()
	cm.v.OnConfigChange(func(e fsnotify.Event) {
		cm.mu.RLock()
		callbacks := cm.onChangeCallbacks
		cm.mu.RUnlock()

		for _, callback := range callbacks {
			callback()
		}
	})
}

// Load reads configuration - Viper already has it in memory so this is a no-op
func (cm *ConfigManager) Load() error {
	// Viper automatically caches so no separate load is needed
	// Method is maintained for backward compatibility
	return nil
}

// Reload forces a configuration reload from file
func (cm *ConfigManager) Reload() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	return cm.v.ReadInConfig()
}

// Save writes the current configuration to file
func (cm *ConfigManager) Save() error {
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
func (cm *ConfigManager) GetConfig() *Config {
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
func (cm *ConfigManager) AddProject(projectName string) error {
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
func (cm *ConfigManager) AddService(projectName, serviceName string, commands []string) error {
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
func (cm *ConfigManager) AppendToService(projectName, serviceName string, commands []string) error {
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

// UpdateService updates an existing service
func (cm *ConfigManager) UpdateService(projectName, serviceName string, commands []string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	commandsPath := fmt.Sprintf("projects.%s.services.%s.commands", projectName, serviceName)

	if !cm.v.IsSet(commandsPath) {
		return fmt.Errorf("service '%s' not found in project '%s'", serviceName, projectName)
	}

	cm.v.Set(commandsPath, commands)
	return nil
}

// DeleteProject removes a project from configuration
func (cm *ConfigManager) DeleteProject(projectName string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	projectPath := fmt.Sprintf("projects.%s", projectName)

	if !cm.v.IsSet(projectPath) {
		return fmt.Errorf("project '%s' not found", projectName)
	}

	// Delete key from Viper
	projects := cm.v.GetStringMap("projects")
	delete(projects, projectName)
	cm.v.Set("projects", projects)

	return nil
}

// DeleteService removes a service from a project
func (cm *ConfigManager) DeleteService(projectName, serviceName string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	servicesPath := fmt.Sprintf("projects.%s.services", projectName)

	if !cm.v.IsSet(servicesPath) {
		return fmt.Errorf("project '%s' not found", projectName)
	}

	services := cm.v.GetStringMap(servicesPath)
	if services == nil {
		return fmt.Errorf("no services in project '%s'", projectName)
	}

	if _, exists := services[serviceName]; !exists {
		return fmt.Errorf("service '%s' not found in project '%s'", serviceName, projectName)
	}

	delete(services, serviceName)
	cm.v.Set(servicesPath, services)

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

// OnConfigChange registers a callback for configuration changes
func (cm *ConfigManager) OnConfigChange(callback func()) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.onChangeCallbacks = append(cm.onChangeCallbacks, callback)
}

// Get returns a raw value from configuration (Viper style)
func (cm *ConfigManager) Get(key string) interface{} {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return cm.v.Get(key)
}

// GetString returns a string value from configuration
func (cm *ConfigManager) GetString(key string) string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return cm.v.GetString(key)
}

// GetStringSlice returns a string slice from configuration
func (cm *ConfigManager) GetStringSlice(key string) []string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return cm.v.GetStringSlice(key)
}

// GetStringMap returns a string map from configuration
func (cm *ConfigManager) GetStringMap(key string) map[string]interface{} {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return cm.v.GetStringMap(key)
}

// IsSet checks if a key is set in the configuration
func (cm *ConfigManager) IsSet(key string) bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return cm.v.IsSet(key)
}
