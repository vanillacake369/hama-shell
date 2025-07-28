package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// Config represents the expected YAML structure (schema definition only)
type Config struct {
	Projects       map[string]Project `yaml:"projects"`
	GlobalSettings GlobalSettings     `yaml:"global_settings"`
}

// Project groups stages under a project name.
type Project struct {
	Description string           `yaml:"description"`
	Stages      map[string]Stage `yaml:"stages"`
}

// Stage represents a deployment or build stage within a project.
type Stage struct {
	Description string             `yaml:"description,omitempty"`
	Services    map[string]Service `yaml:"services"`
}

// Service defines connection details for a specific service.
type Service struct {
	Description string   `yaml:"description"`
	Commands    []string `yaml:"commands"`
}

// GlobalSettings configures retry logic, timeouts, and auto-restart behavior.
type GlobalSettings struct {
	Retries     int  `yaml:"retries"`
	Timeout     int  `yaml:"timeout"`
	AutoRestart bool `yaml:"auto_restart"`
}

// Validator validates viper config data against expected schema
type Validator struct{}

// NewValidator creates a new config validator
func NewValidator() *Validator {
	return &Validator{}
}

// ValidateViper validates the current viper configuration
func (v *Validator) ValidateViper() error {
	data := viper.AllSettings()
	return v.ValidateFromMap(data)
}

// ValidateFromMap validates config data from a map structure
func (v *Validator) ValidateFromMap(data map[string]interface{}) error {
	// Check if projects section exists
	projects, exists := data["projects"]
	if !exists {
		return fmt.Errorf("missing required 'projects' section")
	}

	projectsMap, ok := projects.(map[string]interface{})
	if !ok {
		return fmt.Errorf("'projects' must be an object")
	}

	if len(projectsMap) == 0 {
		return fmt.Errorf("'projects' section cannot be empty")
	}

	// Validate each project
	for projectName, projectData := range projectsMap {
		if err := v.validateProject(projectName, projectData); err != nil {
			return err
		}
	}

	// Validate global settings if present
	if globalSettings, exists := data["global_settings"]; exists {
		if err := v.validateGlobalSettings(globalSettings); err != nil {
			return err
		}
	}

	return nil
}

// validateProject validates a single project structure
func (v *Validator) validateProject(projectName string, projectData interface{}) error {
	project, ok := projectData.(map[string]interface{})
	if !ok {
		return fmt.Errorf("project '%s' must be an object", projectName)
	}

	// Check stages
	stages, exists := project["stages"]
	if !exists {
		return fmt.Errorf("project '%s' missing required 'stages' section", projectName)
	}

	stagesMap, ok := stages.(map[string]interface{})
	if !ok {
		return fmt.Errorf("project '%s' stages must be an object", projectName)
	}

	if len(stagesMap) == 0 {
		return fmt.Errorf("project '%s' must have at least one stage", projectName)
	}

	// Validate each stage
	for stageName, stageData := range stagesMap {
		if err := v.validateStage(projectName, stageName, stageData); err != nil {
			return err
		}
	}

	return nil
}

// validateStage validates a single stage structure
func (v *Validator) validateStage(projectName, stageName string, stageData interface{}) error {
	stage, ok := stageData.(map[string]interface{})
	if !ok {
		return fmt.Errorf("stage '%s.%s' must be an object", projectName, stageName)
	}

	// Check services
	services, exists := stage["services"]
	if !exists {
		return fmt.Errorf("stage '%s.%s' missing required 'services' section", projectName, stageName)
	}

	servicesMap, ok := services.(map[string]interface{})
	if !ok {
		return fmt.Errorf("stage '%s.%s' services must be an object", projectName, stageName)
	}

	if len(servicesMap) == 0 {
		return fmt.Errorf("stage '%s.%s' must have at least one service", projectName, stageName)
	}

	// Validate each service
	for serviceName, serviceData := range servicesMap {
		if err := v.validateService(projectName, stageName, serviceName, serviceData); err != nil {
			return err
		}
	}

	return nil
}

// validateService validates a single service structure
func (v *Validator) validateService(projectName, stageName, serviceName string, serviceData interface{}) error {
	service, ok := serviceData.(map[string]interface{})
	if !ok {
		return fmt.Errorf("service '%s.%s.%s' must be an object", projectName, stageName, serviceName)
	}

	// Check commands
	commands, exists := service["commands"]
	if !exists {
		return fmt.Errorf("service '%s.%s.%s' missing required 'commands' section", projectName, stageName, serviceName)
	}

	commandsList, ok := commands.([]interface{})
	if !ok {
		return fmt.Errorf("service '%s.%s.%s' commands must be a list", projectName, stageName, serviceName)
	}

	if len(commandsList) == 0 {
		return fmt.Errorf("service '%s.%s.%s' must have at least one command", projectName, stageName, serviceName)
	}

	// Validate each command is a string
	for i, cmd := range commandsList {
		if _, ok := cmd.(string); !ok {
			return fmt.Errorf("service '%s.%s.%s' command[%d] must be a string", projectName, stageName, serviceName, i)
		}
		if cmd.(string) == "" {
			return fmt.Errorf("service '%s.%s.%s' command[%d] cannot be empty", projectName, stageName, serviceName, i)
		}
	}

	return nil
}

// validateGlobalSettings validates global settings structure
func (v *Validator) validateGlobalSettings(globalData interface{}) error {
	settings, ok := globalData.(map[string]interface{})
	if !ok {
		return fmt.Errorf("global_settings must be an object")
	}

	// Validate timeout if present
	if timeout, exists := settings["timeout"]; exists {
		if _, ok := timeout.(int); !ok {
			return fmt.Errorf("global_settings.timeout must be an integer")
		}
	}

	// Validate retries if present
	if retries, exists := settings["retries"]; exists {
		if _, ok := retries.(int); !ok {
			return fmt.Errorf("global_settings.retries must be an integer")
		}
	}

	// Validate auto_restart if present
	if autoRestart, exists := settings["auto_restart"]; exists {
		if _, ok := autoRestart.(bool); !ok {
			return fmt.Errorf("global_settings.auto_restart must be a boolean")
		}
	}

	return nil
}

// GetProjects returns a list of available projects from viper
func (v *Validator) GetProjects() []string {
	projects := viper.GetStringMap("projects")
	var projectNames []string
	for name := range projects {
		projectNames = append(projectNames, name)
	}
	return projectNames
}

// GetStages returns a list of stages for a given project
func (v *Validator) GetStages(projectName string) []string {
	stages := viper.GetStringMap(fmt.Sprintf("projects.%s.stages", projectName))
	var stageNames []string
	for name := range stages {
		stageNames = append(stageNames, name)
	}
	return stageNames
}

// GetServices returns a list of services for a given project and stage
func (v *Validator) GetServices(projectName, stageName string) []string {
	services := viper.GetStringMap(fmt.Sprintf("projects.%s.stages.%s.services", projectName, stageName))
	var serviceNames []string
	for name := range services {
		serviceNames = append(serviceNames, name)
	}
	return serviceNames
}

// ParseAndValidate parses a config file and returns a validated Config struct
func (v *Validator) ParseAndValidate(configPath string) (*Config, error) {
	// If no config path provided, try to find default locations
	if configPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user home directory: %w", err)
		}

		// Try common locations
		possiblePaths := []string{
			filepath.Join(home, "hama-shell.yaml"),
			filepath.Join(home, "hama-shell.yml"),
			"hama-shell.yaml",
			"hama-shell.yml",
		}

		for _, path := range possiblePaths {
			if _, err := os.Stat(path); err == nil {
				configPath = path
				break
			}
		}

		if configPath == "" {
			return nil, fmt.Errorf("no config file found in default locations")
		}
	}

	// Read the config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	// Parse YAML
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML config: %w", err)
	}

	// Apply default values
	if config.GlobalSettings.Timeout == 0 {
		config.GlobalSettings.Timeout = 30 // default 30 seconds
	}
	if config.GlobalSettings.Retries == 0 {
		config.GlobalSettings.Retries = 3 // default 3 retries
	}
	// AutoRestart defaults to false, which is zero value

	// Validate the parsed config
	if err := v.ValidateConfig(&config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

// ValidateConfig validates a Config struct directly
func (v *Validator) ValidateConfig(config *Config) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// Check if projects section exists
	if len(config.Projects) == 0 {
		return fmt.Errorf("missing required 'projects' section")
	}

	// Validate each project
	for projectName, project := range config.Projects {
		if err := v.validateProjectStruct(projectName, project); err != nil {
			return err
		}
	}

	// Validate global settings
	if err := v.validateGlobalSettingsStruct(config.GlobalSettings); err != nil {
		return err
	}

	return nil
}

// validateProjectStruct validates a Project struct
func (v *Validator) validateProjectStruct(projectName string, project Project) error {
	if len(project.Stages) == 0 {
		return fmt.Errorf("project '%s' must have at least one stage", projectName)
	}

	// Validate each stage
	for stageName, stage := range project.Stages {
		if err := v.validateStageStruct(projectName, stageName, stage); err != nil {
			return err
		}
	}

	return nil
}

// validateStageStruct validates a Stage struct
func (v *Validator) validateStageStruct(projectName, stageName string, stage Stage) error {
	if len(stage.Services) == 0 {
		return fmt.Errorf("stage '%s.%s' must have at least one service", projectName, stageName)
	}

	// Validate each service
	for serviceName, service := range stage.Services {
		if err := v.validateServiceStruct(projectName, stageName, serviceName, service); err != nil {
			return err
		}
	}

	return nil
}

// validateServiceStruct validates a Service struct
func (v *Validator) validateServiceStruct(projectName, stageName, serviceName string, service Service) error {
	if len(service.Commands) == 0 {
		return fmt.Errorf("service '%s.%s.%s' must have at least one command", projectName, stageName, serviceName)
	}

	// Validate each command is not empty
	for i, cmd := range service.Commands {
		if cmd == "" {
			return fmt.Errorf("service '%s.%s.%s' command[%d] cannot be empty", projectName, stageName, serviceName, i)
		}
	}

	return nil
}

// validateGlobalSettingsStruct validates GlobalSettings struct
func (v *Validator) validateGlobalSettingsStruct(settings GlobalSettings) error {
	if settings.Timeout < 0 {
		return fmt.Errorf("global_settings.timeout must be non-negative")
	}
	if settings.Retries < 0 {
		return fmt.Errorf("global_settings.retries must be non-negative")
	}

	return nil
}

// GetCommands returns commands for a specific session path from parsed config
func GetCommands(config *Config, sessionPath string) ([]string, error) {
	if config == nil {
		return nil, fmt.Errorf("config is nil")
	}

	// Parse session path: project.stage.service
	parts := strings.Split(sessionPath, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid session path format, expected 'project.stage.service', got '%s'", sessionPath)
	}

	projectName, stageName, serviceName := parts[0], parts[1], parts[2]

	// Navigate config structure
	project, exists := config.Projects[projectName]
	if !exists {
		return nil, fmt.Errorf("project '%s' not found", projectName)
	}

	stage, exists := project.Stages[stageName]
	if !exists {
		return nil, fmt.Errorf("stage '%s' not found in project '%s'", stageName, projectName)
	}

	service, exists := stage.Services[serviceName]
	if !exists {
		return nil, fmt.Errorf("service '%s' not found in stage '%s.%s'", serviceName, projectName, stageName)
	}

	return service.Commands, nil
}
