package api

import (
	"fmt"

	"hama-shell/internal/core/service/infra"
	"hama-shell/internal/core/service/model"
)

// ServiceAPI provides high-level service operations
type ServiceAPI struct {
	configReader  *infra.ConfigReader
	terminalMgr   *infra.TerminalManager
}

// NewServiceAPI creates a new ServiceAPI instance
func NewServiceAPI() *ServiceAPI {
	return &ServiceAPI{
		configReader: infra.NewConfigReader(),
		terminalMgr:  infra.NewTerminalManager(),
	}
}

// StartService starts a service by project and service name
func (api *ServiceAPI) StartService(projectName, serviceName string) error {
	// Get service configuration
	service, err := api.configReader.GetService(projectName, serviceName)
	if err != nil {
		return fmt.Errorf("failed to get service '%s.%s': %w", projectName, serviceName, err)
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

	// Group services by project
	projectServices := make(map[string][]model.Service)
	for _, service := range services {
		projectServices[service.ProjectName] = append(projectServices[service.ProjectName], service)
	}

	// Display grouped services
	for projectName, projectSvcs := range projectServices {
		fmt.Printf("📁 Project: %s\n", projectName)
		
		if len(projectSvcs) == 0 {
			fmt.Println("  (no services defined)")
		} else {
			for _, service := range projectSvcs {
				fmt.Printf("  🔧 %s\n", service.GetFullName())
				for i, command := range service.Commands {
					fmt.Printf("    [%d] %s\n", i+1, command)
				}
			}
		}
		fmt.Println()
	}

	fmt.Println("Usage:")
	fmt.Println("  hama-shell service start <project>.<service>")
	fmt.Printf("\nConfiguration file: %s\n", api.configReader.GetConfigFilePath())

	return nil
}

// Shutdown gracefully shuts down the API and its dependencies
func (api *ServiceAPI) Shutdown() error {
	return api.terminalMgr.Shutdown()
}