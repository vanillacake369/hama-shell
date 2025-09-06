package cmd

import (
	"bufio"
	"fmt"
	"hama-shell/internal/core/config"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage hama-shell configuration",
	Long: `View, edit, and manage hama-shell configuration files.
	
Available subcommands:
  view    - Display configuration file contents
  edit    - Edit configuration file
  create  - Create a new configuration file
  add     - Add a new command to configuration`,
	Run: func(cmd *cobra.Command, args []string) {
		// If no subcommand is provided, show help
		_ = cmd.Help()
	},
}

// configViewCmd represents the config view command
var configViewCmd = &cobra.Command{
	Use:   "view",
	Short: "View command configuration",
	Long:  `View configuration file`,
	RunE: func(cmd *cobra.Command, args []string) error {
		manager := config.GetInstance()

		// Check if configuration file exists
		if !manager.FileExists() {
			fmt.Println("Configuration file not found!")
			fmt.Printf("Please create one using: hama-shell config create\n")
			fmt.Printf("Expected location: %s\n", manager.GetFilePath())
			return nil
		}

		// Get and display configuration
		cfg := manager.GetConfig()
		if cfg == nil || len(cfg.Projects) == 0 {
			fmt.Println("Configuration file exists but is empty.")
			fmt.Printf("Add commands using: hama-shell config add\n")
			return nil
		}

		// Convert to YAML and display
		data, err := yaml.Marshal(cfg)
		if err != nil {
			return fmt.Errorf("failed to format configuration: %w", err)
		}

		fmt.Printf("Configuration file: %s\n", manager.GetFilePath())
		fmt.Println("=====================================")
		fmt.Print(string(data))

		return nil
	},
}

// configAddCmd represents the config add command
var configAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new command to configuration",
	Long:  `Add a new command to be executed in a hama-shell session.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		manager := config.GetInstance()
		// Viper-based so no need to call Load() - already cached in memory

		reader := bufio.NewReader(os.Stdin)

		fmt.Println("Add to existing configuration")
		fmt.Println("==============================")

		// Show existing projects
		existingConfig := manager.GetConfig()
		if existingConfig != nil && len(existingConfig.Projects) > 0 {
			fmt.Println("\nExisting projects:")
			for name := range existingConfig.Projects {
				fmt.Printf("  - %s\n", name)
			}
		}

		// Get project name
		fmt.Print("\nEnter project name (or new project name): ")
		projectName, _ := reader.ReadString('\n')
		projectName = strings.TrimSpace(projectName)

		// Check if project exists
		var projectExists bool
		if existingConfig != nil {
			if _, exists := existingConfig.Projects[projectName]; exists {
				projectExists = true
				fmt.Printf("Adding service to existing project '%s'\n", projectName)
			}
		}

		if !projectExists {
			fmt.Printf("Creating new project '%s'\n", projectName)
		}

		// Get service name
		fmt.Print("Enter service name: ")
		serviceName, _ := reader.ReadString('\n')
		serviceName = strings.TrimSpace(serviceName)

		// Check if service exists in the project
		serviceExists := checkServiceExists(existingConfig, projectName, serviceName)
		if serviceExists {
			fmt.Printf("Service '%s' already exists. Adding commands to it.\n", serviceName)
		}

		// Get commands
		commands := readCommands(reader)

		// Add to configuration
		if err := processConfigUpdate(manager, projectName, serviceName, commands, projectExists, serviceExists); err != nil {
			return err
		}

		// Save configuration
		if err := manager.Save(); err != nil {
			return fmt.Errorf("failed to save configuration: %w", err)
		}

		fmt.Printf("\nConfiguration updated successfully!\n")
		fmt.Printf("File saved at: %s\n", manager.GetFilePath())

		return nil
	},
}

// configCreateCmd represents the config create command
var configCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a configuration",
	Long: `Create a configuration that contains commands.

You can also provide command details via flags for non-interactive mode.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		manager := config.GetInstance()
		// Viper-based so automatically initialized

		// Check if file already exists
		if manager.FileExists() {
			fmt.Println("Configuration file already exists")
			return nil
		}

		// Get input step by step via DTO and save to file
		reader := bufio.NewReader(os.Stdin)

		fmt.Println("Creating new hama-shell configuration")
		fmt.Println("=====================================")

		// 1. Get project name
		fmt.Print("Enter project name: ")
		projectName, _ := reader.ReadString('\n')
		projectName = strings.TrimSpace(projectName)

		// 2. Get service name
		fmt.Print("Enter service name: ")
		serviceName, _ := reader.ReadString('\n')
		serviceName = strings.TrimSpace(serviceName)

		// 3. Get commands
		commands := readCommands(reader)

		// Add project and service using ConfigManager
		if err := manager.AddProject(projectName); err != nil {
			return fmt.Errorf("failed to add project: %w", err)
		}

		if err := manager.AddService(projectName, serviceName, commands); err != nil {
			return fmt.Errorf("failed to add service: %w", err)
		}

		// Save configuration
		if err := manager.Save(); err != nil {
			return fmt.Errorf("failed to save configuration: %w", err)
		}

		fmt.Printf("\nConfiguration file created at: %s\n", manager.GetFilePath())

		// Display the generated configuration
		displayConfig(manager.GetConfig())

		return nil
	},
}

// Helper functions

// readCommands reads commands from user input until empty line
func readCommands(reader *bufio.Reader) []string {
	fmt.Println("Enter commands (one per line, empty line to finish):")
	var commands []string
	for {
		fmt.Print("> ")
		command, _ := reader.ReadString('\n')
		command = strings.TrimSpace(command)
		if command == "" {
			break
		}
		commands = append(commands, command)
	}
	return commands
}

// checkServiceExists checks if a service exists in a project
func checkServiceExists(cfg *config.Config, projectName, serviceName string) bool {
	if cfg == nil {
		return false
	}
	project, projectExists := cfg.Projects[projectName]
	if !projectExists {
		return false
	}
	_, serviceExists := project.Services[serviceName]
	return serviceExists
}

// displayConfig displays the configuration in YAML format
func displayConfig(cfg *config.Config) {
	if cfg == nil {
		return
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return
	}
	fmt.Println("\nGenerated configuration:")
	fmt.Println("------------------------")
	fmt.Print(string(data))
}

// processConfigUpdate handles adding/updating configuration based on existence
func processConfigUpdate(manager config.ConfigManager, projectName, serviceName string, commands []string, projectExists, serviceExists bool) error {
	switch {
	case !projectExists:
		// Create new project and service
		if err := manager.AddProject(projectName); err != nil {
			return err
		}
		if err := manager.AddService(projectName, serviceName, commands); err != nil {
			return err
		}
		fmt.Printf("Created new project '%s' with service '%s'\n", projectName, serviceName)

	case serviceExists:
		// Append to existing service
		if err := manager.AppendToService(projectName, serviceName, commands); err != nil {
			return err
		}
		fmt.Printf("Added %d commands to existing service '%s'\n", len(commands), serviceName)

	default:
		// Add new service to existing project
		if err := manager.AddService(projectName, serviceName, commands); err != nil {
			return err
		}
		fmt.Printf("Created new service '%s' with %d commands\n", serviceName, len(commands))
	}
	return nil
}

func init() {
	rootCmd.AddCommand(configCmd)

	configCmd.AddCommand(configViewCmd)
	configCmd.AddCommand(configAddCmd)
	configCmd.AddCommand(configCreateCmd)
}
