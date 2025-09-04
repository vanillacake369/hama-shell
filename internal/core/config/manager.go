package config

import (
	"fmt"
	"hama-shell/types"
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

// ConfigManager manages configuration with memory caching and file persistence
type ConfigManager struct {
	config   *types.Config
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
	if err != nil {
		if os.IsNotExist(err) {
			// Initialize with empty config if file doesn't exist
			cm.config = &types.Config{
				Projects: []types.Project{},
			}
			return nil
		}
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var config types.Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
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
func (cm *ConfigManager) GetConfig() *types.Config {
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
func (cm *ConfigManager) AddProject(project types.Project) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.config == nil {
		cm.config = &types.Config{
			Projects: []types.Project{},
		}
	}

	// Check if project already exists
	for _, p := range cm.config.Projects {
		if p.Name == project.Name {
			return fmt.Errorf("project '%s' already exists", project.Name)
		}
	}

	cm.config.Projects = append(cm.config.Projects, project)
	return nil
}

// AddService adds a service to an existing project
func (cm *ConfigManager) AddService(projectName string, service types.Service) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.config == nil {
		return fmt.Errorf("configuration not loaded")
	}

	for i, p := range cm.config.Projects {
		if p.Name == projectName {
			// Check if service already exists
			for _, s := range p.Services {
				if s.Name == service.Name {
					return fmt.Errorf("service '%s' already exists in project '%s'", service.Name, projectName)
				}
			}
			cm.config.Projects[i].Services = append(p.Services, service)
			return nil
		}
	}

	return fmt.Errorf("project '%s' not found", projectName)
}

// UpdateService updates an existing service
func (cm *ConfigManager) UpdateService(projectName, serviceName string, commands []string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.config == nil {
		return fmt.Errorf("configuration not loaded")
	}

	for i, p := range cm.config.Projects {
		if p.Name == projectName {
			for j, s := range p.Services {
				if s.Name == serviceName {
					cm.config.Projects[i].Services[j].Commands = commands
					return nil
				}
			}
			return fmt.Errorf("service '%s' not found in project '%s'", serviceName, projectName)
		}
	}

	return fmt.Errorf("project '%s' not found", projectName)
}

// DeleteProject removes a project from configuration
func (cm *ConfigManager) DeleteProject(projectName string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.config == nil {
		return fmt.Errorf("configuration not loaded")
	}

	for i, p := range cm.config.Projects {
		if p.Name == projectName {
			cm.config.Projects = append(cm.config.Projects[:i], cm.config.Projects[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("project '%s' not found", projectName)
}

// DeleteService removes a service from a project
func (cm *ConfigManager) DeleteService(projectName, serviceName string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.config == nil {
		return fmt.Errorf("configuration not loaded")
	}

	for i, p := range cm.config.Projects {
		if p.Name == projectName {
			for j, s := range p.Services {
				if s.Name == serviceName {
					cm.config.Projects[i].Services = append(p.Services[:j], p.Services[j+1:]...)
					return nil
				}
			}
			return fmt.Errorf("service '%s' not found in project '%s'", serviceName, projectName)
		}
	}

	return fmt.Errorf("project '%s' not found", projectName)
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
