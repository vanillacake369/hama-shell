package api

import (
	"fmt"

	"hama-shell/internal/core/service/infra"
	"hama-shell/internal/core/service/model"
)

// ServiceAPI provides high-level service operations
type ServiceAPI struct {
	configReader *infra.ConfigReader
	terminalMgr  *infra.TerminalManager
}

// NewServiceAPI creates a new ServiceAPI instance
func NewServiceAPI() *ServiceAPI {
	return &ServiceAPI{
		configReader: infra.NewConfigReader(),
		terminalMgr:  infra.NewTerminalManager(),
	}
}

// StartService starts a service by project, service, and stage name
func (api *ServiceAPI) StartService(projectName, serviceName, stageName string) error {
	// Get service configuration
	service, err := api.configReader.GetService(projectName, serviceName, stageName)
	if err != nil {
		return fmt.Errorf("failed to get service '%s.%s.%s': %w", projectName, serviceName, stageName, err)
	}

	// Print service information
	fmt.Printf("🚀 Starting service: %s\n", service.GetFullName())
	fmt.Printf("📋 Commands to execute:\n")
	for i, cmd := range service.Commands {
		fmt.Printf("  [%d] %s\n", i+1, cmd)
	}
	fmt.Printf("\n🔗 Connecting to interactive terminal...\n\n")

	// Start interactive terminal session
	if err := api.terminalMgr.StartInteractiveSession(service); err != nil {
		return fmt.Errorf("failed to start terminal session: %w", err)
	}

	return nil
}

// ListServices returns all available services
func (api *ServiceAPI) ListServices() error {
	services, err := api.configReader.ListAllServices()
	if err != nil {
		return fmt.Errorf("failed to list services: %w", err)
	}

	if len(services) == 0 {
		fmt.Println("No projects found in configuration.")
		fmt.Printf("Configuration file: %s\n", api.configReader.GetConfigFilePath())
		return nil
	}

	fmt.Println("Available services:")
	fmt.Println()

	// Group services by project and service
	projectServiceStages := make(map[string]map[string][]model.Service)
	for _, service := range services {
		if projectServiceStages[service.ProjectName] == nil {
			projectServiceStages[service.ProjectName] = make(map[string][]model.Service)
		}
		projectServiceStages[service.ProjectName][service.ServiceName] = append(
			projectServiceStages[service.ProjectName][service.ServiceName], service)
	}

	// Display grouped services
	for projectName, servicePairs := range projectServiceStages {
		fmt.Printf("📁 Project: %s\n", projectName)

		if len(servicePairs) == 0 {
			fmt.Println("  (no services defined)")
		} else {
			for serviceName, stages := range servicePairs {
				fmt.Printf("  🔧 Service: %s\n", serviceName)
				for _, stage := range stages {
					fmt.Printf("    📋 %s\n", stage.GetFullName())
					for i, command := range stage.Commands {
						fmt.Printf("      [%d] %s\n", i+1, command)
					}
				}
			}
		}
		fmt.Println()
	}

	fmt.Println("Usage:")
	fmt.Println("  hama-shell service start <project>.<service>.<stage>")
	fmt.Printf("\nConfiguration file: %s\n", api.configReader.GetConfigFilePath())

	return nil
}

// Shutdown gracefully shuts down the API and its dependencies
func (api *ServiceAPI) Shutdown() error {
	return api.terminalMgr.Shutdown()
}
