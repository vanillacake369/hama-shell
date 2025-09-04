package cmd

import (
	"bufio"
	"fmt"
	"hama-shell/internal/core/config"
	"hama-shell/types"
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
		cmd.Help()
	},
}

// configAddCmd represents the config add command
var configAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new command to configuration",
	Long: `Add a new command to be executed in a hama-shell session.

Interactive mode:
  - Enter command details when prompted
  - Press Alt+C to cancel at any time
  - Press Alt+F to finish and save

You can also provide command details via flags for non-interactive mode.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		manager := config.GetInstance()

		// Load existing configuration
		if err := manager.Load(); err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		reader := bufio.NewReader(os.Stdin)

		fmt.Println("Add to existing configuration")
		fmt.Println("==============================")

		// Show existing projects
		existingConfig := manager.GetConfig()
		if existingConfig != nil && len(existingConfig.Projects) > 0 {
			fmt.Println("\nExisting projects:")
			for _, p := range existingConfig.Projects {
				fmt.Printf("  - %s\n", p.Name)
			}
		}

		// Get project name
		fmt.Print("\nEnter project name (or new project name): ")
		projectName, _ := reader.ReadString('\n')
		projectName = strings.TrimSpace(projectName)

		// Check if project exists
		var projectExists bool
		if existingConfig != nil {
			for _, p := range existingConfig.Projects {
				if p.Name == projectName {
					projectExists = true
					fmt.Printf("Adding service to existing project '%s'\n", projectName)
					break
				}
			}
		}

		if !projectExists {
			fmt.Printf("Creating new project '%s'\n", projectName)
		}

		// Get service name
		fmt.Print("Enter service name: ")
		serviceName, _ := reader.ReadString('\n')
		serviceName = strings.TrimSpace(serviceName)

		// Get commands
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

		// Add to configuration
		if projectExists {
			// Add service to existing project
			service := types.Service{
				Name:     serviceName,
				Commands: commands,
			}
			if err := manager.AddService(projectName, service); err != nil {
				return err
			}
		} else {
			// Create new project with service
			project := types.Project{
				Name: projectName,
				Services: []types.Service{
					{
						Name:     serviceName,
						Commands: commands,
					},
				},
			}
			if err := manager.AddProject(project); err != nil {
				return err
			}
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
		// home directory 의 설정 파일을 읽어옴
		home := os.Getenv("HOME")
		fileName := "hama-shell.yaml"
		filePath := home + "/" + fileName

		// Check if file already exists
		if _, err := os.Stat(filePath); err == nil {
			fmt.Println("Configuration file already exists")
			return nil
		}

		// 단계 별로 DTO 로 입력받아 file 저장
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

		// Create configuration object
		config := types.Config{
			Projects: []types.Project{
				{
					Name: projectName,
					Services: []types.Service{
						{
							Name:     serviceName,
							Commands: commands,
						},
					},
				},
			},
		}

		// Convert to YAML
		data, err := yaml.Marshal(&config)
		if err != nil {
			return fmt.Errorf("failed to marshal configuration: %w", err)
		}

		// Write the configuration file
		err = os.WriteFile(filePath, data, 0644)
		if err != nil {
			return fmt.Errorf("failed to create configuration file: %w", err)
		}

		fmt.Printf("\nConfiguration file created at: %s\n", filePath)
		fmt.Println("\nGenerated configuration:")
		fmt.Println("------------------------")
		fmt.Print(string(data))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)

	configCmd.AddCommand(configAddCmd)
	configCmd.AddCommand(configCreateCmd)
}
