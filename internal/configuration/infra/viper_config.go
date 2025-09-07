package infra

import (
	"fmt"
	"hama-shell/internal/configuration/model"
	"os"
	"path/filepath"
	"sync"

	"github.com/spf13/viper"
)

// ConfigManager interface for managing configuration
type ConfigManager interface {
	// GetConfig returns the current configuration
	GetConfig() *model.Config

	// Save writes the current configuration to file
	Save() error

	// FileExists checks if the configuration file exists
	FileExists() bool

	// GetFilePath returns the configuration file path
	GetFilePath() string

	// AddProject adds a new project to the configuration
	AddProject(projectName string) error

	// AddService adds a service to an existing project
	AddService(projectName, serviceName string) error

	// AddStage adds a stage to an existing service
	AddStage(projectName, serviceName, stageName string, commands []string) error

	// AppendToService appends commands to an existing service stage
	AppendToService(projectName, serviceName, stageName string, commands []string) error
}

// viperConfigManager manages configuration using Viper (implementation)
type viperConfigManager struct {
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
		instance = newViperConfigManager()
		instance.(*viperConfigManager).initialize()
	})
	return instance
}

// newViperConfigManager creates a new viperConfigManager instance
func newViperConfigManager() ConfigManager {
	home := os.Getenv("HOME")
	filePath := filepath.Join(home, "hama-shell.yaml")

	v := viper.New()
	v.SetConfigFile(filePath)
	v.SetConfigType("yaml")

	return &viperConfigManager{
		v:        v,
		filePath: filePath,
	}
}

// initialize sets up the configuration manager
func (cm *viperConfigManager) initialize() {
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
func (cm *viperConfigManager) Save() error {
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
func (cm *viperConfigManager) GetConfig() *model.Config {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	var config model.Config
	if err := cm.v.Unmarshal(&config); err != nil {
		return &model.Config{
			Projects: make(map[string]*model.Project),
		}
	}

	if config.Projects == nil {
		config.Projects = make(map[string]*model.Project)
	}

	// Initialize nil Services and Stages maps
	for _, project := range config.Projects {
		if project.Services == nil {
			project.Services = make(map[string]*model.Service)
		}
		for _, service := range project.Services {
			if service.Stages == nil {
				service.Stages = make(map[string]*model.Stage)
			}
		}
	}

	return &config
}

// AddProject adds a new project to the configuration
func (cm *viperConfigManager) AddProject(projectName string) error {
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
func (cm *viperConfigManager) AddService(projectName, serviceName string) error {
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

	// Check if service already exists
	if _, exists := services[serviceName]; exists {
		return fmt.Errorf("service '%s' already exists in project '%s'", serviceName, projectName)
	}

	// Create new service with empty stages map
	services[serviceName] = map[string]interface{}{
		"stages": make(map[string]interface{}),
	}

	cm.v.Set(servicesPath, services)
	return nil
}

// AddStage adds a stage to an existing service
func (cm *viperConfigManager) AddStage(projectName, serviceName, stageName string, commands []string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	servicePath := fmt.Sprintf("projects.%s.services.%s", projectName, serviceName)
	if !cm.v.IsSet(servicePath) {
		return fmt.Errorf("service '%s' not found in project '%s'", serviceName, projectName)
	}

	stagesPath := fmt.Sprintf("%s.stages", servicePath)
	stages := cm.v.GetStringMap(stagesPath)
	if stages == nil {
		stages = make(map[string]interface{})
	}

	// Check if stage already exists
	if _, exists := stages[stageName]; exists {
		return fmt.Errorf("stage '%s' already exists in service '%s.%s'", stageName, projectName, serviceName)
	}

	// Add new stage
	stages[stageName] = map[string]interface{}{
		"commands": commands,
	}

	cm.v.Set(stagesPath, stages)
	return nil
}

// AppendToService appends commands to an existing service stage
func (cm *viperConfigManager) AppendToService(projectName, serviceName, stageName string, commands []string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	commandsPath := fmt.Sprintf("projects.%s.services.%s.stages.%s.commands", projectName, serviceName, stageName)

	if !cm.v.IsSet(commandsPath) {
		return fmt.Errorf("stage '%s' not found in service '%s.%s'", stageName, projectName, serviceName)
	}

	existingCommands := cm.v.GetStringSlice(commandsPath)
	existingCommands = append(existingCommands, commands...)
	cm.v.Set(commandsPath, existingCommands)

	return nil
}

// FileExists checks if the configuration file exists
func (cm *viperConfigManager) FileExists() bool {
	_, err := os.Stat(cm.filePath)
	return err == nil
}

// GetFilePath returns the configuration file path
func (cm *viperConfigManager) GetFilePath() string {
	return cm.filePath
}
