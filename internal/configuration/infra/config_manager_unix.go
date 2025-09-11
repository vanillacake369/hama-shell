package infra

import (
	"fmt"
	"hama-shell/internal/configuration/model"
	"os"
	"path/filepath"
	"sync"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

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

// ViewConfig returns the current configuration view
func (cm *viperConfigManager) ViewConfig() (*model.ConfigView, error) {
	view := &model.ConfigView{
		FilePath: cm.GetFilePath(),
		Exists:   cm.FileExists(),
	}

	if !view.Exists {
		return view, nil
	}

	cfg := cm.GetConfig()
	view.Content = cfg
	view.IsEmpty = cfg == nil || len(cfg.Projects) == 0

	return view, nil
}

// CreateConfig creates a new configuration
func (cm *viperConfigManager) CreateConfig(op model.ConfigOperation) error {
	if cm.FileExists() {
		return fmt.Errorf("configuration file already exists")
	}

	// Add project and service
	if err := cm.AddProject(op.ProjectName); err != nil {
		return fmt.Errorf("failed to add project: %w", err)
	}

	if err := cm.AddService(op.ProjectName, op.ServiceName); err != nil {
		return fmt.Errorf("failed to add service: %w", err)
	}

	if err := cm.AddStage(op.ProjectName, op.ServiceName, op.StageName, op.Commands); err != nil {
		return fmt.Errorf("failed to add stage: %w", err)
	}

	// Save configuration
	if err := cm.Save(); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	return nil
}

// AddToConfig adds a service or updates existing configuration
func (cm *viperConfigManager) AddToConfig(op model.ConfigOperation) error {
	cfg := cm.GetConfig()

	// Check if project, service, and stage exist
	projectExists := false
	serviceExists := false
	stageExists := false

	if cfg != nil {
		if project, exists := cfg.Projects[op.ProjectName]; exists {
			projectExists = true
			if service, exists := project.Services[op.ServiceName]; exists {
				serviceExists = true
				if _, exists := service.Stages[op.StageName]; exists {
					stageExists = true
				}
			}
		}
	}

	// Create missing components step by step
	// 1. Create project if it doesn't exist
	if !projectExists {
		if err := cm.AddProject(op.ProjectName); err != nil {
			return fmt.Errorf("failed to add project: %w", err)
		}
	}

	// 2. Create service if it doesn't exist
	if !serviceExists {
		if err := cm.AddService(op.ProjectName, op.ServiceName); err != nil {
			return fmt.Errorf("failed to add service: %w", err)
		}
	}

	// 3. Handle stage creation/update
	if stageExists {
		// Stage exists - append to existing stage
		if err := cm.AppendToService(op.ProjectName, op.ServiceName, op.StageName, op.Commands); err != nil {
			return fmt.Errorf("failed to append to existing stage: %w", err)
		}
	} else {
		// Stage doesn't exist - add new stage
		if err := cm.AddStage(op.ProjectName, op.ServiceName, op.StageName, op.Commands); err != nil {
			return fmt.Errorf("failed to add stage: %w", err)
		}
	}

	// Save configuration
	if err := cm.Save(); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	return nil
}

// FormatAsYAML formats configuration as YAML string
func (cm *viperConfigManager) FormatAsYAML(content interface{}) (string, error) {
	data, err := yaml.Marshal(content)
	if err != nil {
		return "", fmt.Errorf("failed to format as YAML: %w", err)
	}
	return string(data), nil
}

// GetExistingProjects returns list of existing project names
func (cm *viperConfigManager) GetExistingProjects() []string {
	cfg := cm.GetConfig()
	if cfg == nil || cfg.Projects == nil {
		return []string{}
	}

	var projects []string
	for name := range cfg.Projects {
		projects = append(projects, name)
	}
	return projects
}
