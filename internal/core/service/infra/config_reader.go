package infra

import (
	"hama-shell/internal/core/config"
	"hama-shell/internal/core/service/model"
)

// ConfigReader handles configuration reading operations
type ConfigReader struct {
	manager *config.Config
}

// NewConfigReader creates a new ConfigReader instance
func NewConfigReader() *ConfigReader {
	manager := config.GetInstance()
	return &ConfigReader{
		manager: manager.GetConfig(),
	}
}

// GetService retrieves a specific service configuration
func (c *ConfigReader) GetService(projectName, serviceName string) (*model.Service, error) {
	cfg := c.manager

	// Find project
	project, exists := cfg.Projects[projectName]
	if !exists {
		return nil, model.ErrServiceNotFound
	}

	// Find service
	serviceConfig, exists := project.Services[serviceName]
	if !exists {
		return nil, model.ErrServiceNotFound
	}

	// Create service model
	service := &model.Service{
		ProjectName: projectName,
		ServiceName: serviceName,
		Commands:    serviceConfig.Commands,
	}

	// Validate service
	if err := service.Validate(); err != nil {
		return nil, err
	}

	return service, nil
}

// ListAllServices returns all available services
func (c *ConfigReader) ListAllServices() ([]model.Service, error) {
	cfg := c.manager
	var services []model.Service

	for projectName, project := range cfg.Projects {
		for serviceName, serviceConfig := range project.Services {
			service := model.Service{
				ProjectName: projectName,
				ServiceName: serviceName,
				Commands:    serviceConfig.Commands,
			}
			services = append(services, service)
		}
	}

	return services, nil
}

// GetConfigFilePath returns the configuration file path
func (c *ConfigReader) GetConfigFilePath() string {
	manager := config.GetInstance()
	return manager.GetFilePath()
}