package infra

import (
	"fmt"
	"hama-shell/internal/core/configuration/model"

	"gopkg.in/yaml.v3"
)

// ConfigManagerWrapper handles configuration file operations
type ConfigManagerWrapper struct {
	manager ConfigManager
}

// NewConfigManager creates a new ConfigManagerWrapper instance
func NewConfigManager() *ConfigManagerWrapper {
	return &ConfigManagerWrapper{
		manager: GetInstance(),
	}
}

// ViewConfig returns the current configuration view
func (cm *ConfigManagerWrapper) ViewConfig() (*model.ConfigView, error) {
	view := &model.ConfigView{
		FilePath: cm.manager.GetFilePath(),
		Exists:   cm.manager.FileExists(),
	}

	if !view.Exists {
		return view, nil
	}

	cfg := cm.manager.GetConfig()
	view.Content = cfg
	view.IsEmpty = cfg == nil || len(cfg.Projects) == 0

	return view, nil
}

// CreateConfig creates a new configuration
func (cm *ConfigManagerWrapper) CreateConfig(op model.ConfigOperation) error {
	if cm.manager.FileExists() {
		return fmt.Errorf("configuration file already exists")
	}

	// Add project and service
	if err := cm.manager.AddProject(op.ProjectName); err != nil {
		return fmt.Errorf("failed to add project: %w", err)
	}

	if err := cm.manager.AddService(op.ProjectName, op.ServiceName, op.Commands); err != nil {
		return fmt.Errorf("failed to add service: %w", err)
	}

	// Save configuration
	if err := cm.manager.Save(); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	return nil
}

// AddToConfig adds a service or updates existing configuration
func (cm *ConfigManagerWrapper) AddToConfig(op model.ConfigOperation) error {
	cfg := cm.manager.GetConfig()
	
	// Check if project exists
	projectExists := false
	serviceExists := false
	
	if cfg != nil {
		if project, exists := cfg.Projects[op.ProjectName]; exists {
			projectExists = true
			if _, exists := project.Services[op.ServiceName]; exists {
				serviceExists = true
			}
		}
	}

	// Process based on existence
	switch {
	case !projectExists:
		// Create new project and service
		if err := cm.manager.AddProject(op.ProjectName); err != nil {
			return err
		}
		if err := cm.manager.AddService(op.ProjectName, op.ServiceName, op.Commands); err != nil {
			return err
		}

	case serviceExists:
		// Append to existing service
		if err := cm.manager.AppendToService(op.ProjectName, op.ServiceName, op.Commands); err != nil {
			return err
		}

	default:
		// Add new service to existing project
		if err := cm.manager.AddService(op.ProjectName, op.ServiceName, op.Commands); err != nil {
			return err
		}
	}

	// Save configuration
	if err := cm.manager.Save(); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	return nil
}

// FormatAsYAML formats configuration as YAML string
func (cm *ConfigManagerWrapper) FormatAsYAML(content interface{}) (string, error) {
	data, err := yaml.Marshal(content)
	if err != nil {
		return "", fmt.Errorf("failed to format as YAML: %w", err)
	}
	return string(data), nil
}

// GetExistingProjects returns list of existing project names
func (cm *ConfigManagerWrapper) GetExistingProjects() []string {
	cfg := cm.manager.GetConfig()
	if cfg == nil || cfg.Projects == nil {
		return []string{}
	}

	var projects []string
	for name := range cfg.Projects {
		projects = append(projects, name)
	}
	return projects
}