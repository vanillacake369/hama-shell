package api

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"hama-shell/internal/core/configuration/infra"
	"hama-shell/internal/core/configuration/model"
)

// ConfigAPI provides high-level configuration operations
type ConfigAPI struct {
	configMgr *infra.ConfigManagerWrapper
	reader    *bufio.Reader
}

// NewConfigAPI creates a new ConfigAPI instance
func NewConfigAPI() *ConfigAPI {
	return &ConfigAPI{
		configMgr: infra.NewConfigManager(),
		reader:    bufio.NewReader(os.Stdin),
	}
}

// ViewConfiguration displays the current configuration
func (api *ConfigAPI) ViewConfiguration() error {
	view, err := api.configMgr.ViewConfig()
	if err != nil {
		return err
	}

	if !view.Exists {
		fmt.Println("Configuration file not found!")
		fmt.Printf("Please create one using: hama-shell config create\n")
		fmt.Printf("Expected location: %s\n", view.FilePath)
		return nil
	}

	if view.IsEmpty {
		fmt.Println("Configuration file exists but is empty.")
		fmt.Printf("Add commands using: hama-shell config add\n")
		return nil
	}

	// Format and display
	yamlContent, err := api.configMgr.FormatAsYAML(view.Content)
	if err != nil {
		return fmt.Errorf("failed to format configuration: %w", err)
	}

	fmt.Printf("Configuration file: %s\n", view.FilePath)
	fmt.Println("=====================================")
	fmt.Print(yamlContent)

	return nil
}

// CreateConfiguration creates a new configuration interactively
func (api *ConfigAPI) CreateConfiguration() error {
	view, err := api.configMgr.ViewConfig()
	if err != nil {
		return err
	}

	if view.Exists {
		fmt.Println("Configuration file already exists")
		return nil
	}

	fmt.Println("Creating new hama-shell configuration")
	fmt.Println("=====================================")

	// Get project name
	fmt.Print("Enter project name: ")
	projectName, _ := api.reader.ReadString('\n')
	projectName = strings.TrimSpace(projectName)

	// Get service name
	fmt.Print("Enter service name: ")
	serviceName, _ := api.reader.ReadString('\n')
	serviceName = strings.TrimSpace(serviceName)

	// Get commands
	commands := api.readCommands()

	// Create configuration
	op := model.ConfigOperation{
		ProjectName: projectName,
		ServiceName: serviceName,
		Commands:    commands,
	}

	if err := api.configMgr.CreateConfig(op); err != nil {
		return err
	}

	fmt.Printf("\nConfiguration file created at: %s\n", view.FilePath)

	// Display the generated configuration
	if err := api.ViewConfiguration(); err != nil {
		return err
	}

	return nil
}

// AddToConfiguration adds a new service or commands to existing configuration
func (api *ConfigAPI) AddToConfiguration() error {
	fmt.Println("Add to existing configuration")
	fmt.Println("==============================")

	// Show existing projects
	projects := api.configMgr.GetExistingProjects()
	if len(projects) > 0 {
		fmt.Println("\nExisting projects:")
		for _, name := range projects {
			fmt.Printf("  - %s\n", name)
		}
	}

	// Get project name
	fmt.Print("\nEnter project name (or new project name): ")
	projectName, _ := api.reader.ReadString('\n')
	projectName = strings.TrimSpace(projectName)

	// Check if project exists
	projectExists := false
	for _, p := range projects {
		if p == projectName {
			projectExists = true
			fmt.Printf("Adding service to existing project '%s'\n", projectName)
			break
		}
	}

	if !projectExists {
		fmt.Printf("Creating new project '%s'\n", projectName)
	}

	// Get service name
	fmt.Print("Enter service name: ")
	serviceName, _ := api.reader.ReadString('\n')
	serviceName = strings.TrimSpace(serviceName)

	// Get commands
	commands := api.readCommands()

	// Add to configuration
	op := model.ConfigOperation{
		ProjectName: projectName,
		ServiceName: serviceName,
		Commands:    commands,
	}

	if err := api.configMgr.AddToConfig(op); err != nil {
		return err
	}

	view, _ := api.configMgr.ViewConfig()
	fmt.Printf("\nConfiguration updated successfully!\n")
	fmt.Printf("File saved at: %s\n", view.FilePath)

	return nil
}

// readCommands reads commands from user input until empty line
func (api *ConfigAPI) readCommands() []string {
	fmt.Println("Enter commands (one per line, empty line to finish):")
	var commands []string
	for {
		fmt.Print("> ")
		command, _ := api.reader.ReadString('\n')
		command = strings.TrimSpace(command)
		if command == "" {
			break
		}
		commands = append(commands, command)
	}
	return commands
}