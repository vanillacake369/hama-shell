package infra

import (
	"hama-shell/internal/configuration/model"
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

	// ViewConfig returns the current configuration view
	ViewConfig() (*model.ConfigView, error)

	// CreateConfig creates a new configuration
	CreateConfig(op model.ConfigOperation) error

	// AddToConfig adds a service or updates existing configuration
	AddToConfig(op model.ConfigOperation) error

	// FormatAsYAML formats configuration as YAML string
	FormatAsYAML(content interface{}) (string, error)

	// GetExistingProjects returns list of existing project names
	GetExistingProjects() []string
}

// NewConfigManager creates a new ConfigManager instance
func NewConfigManager() ConfigManager {
	return GetInstance()
}
