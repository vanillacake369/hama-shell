package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"hama-shell/internal/core/config"
)

// configService singleton instance (reused from start.go pattern)
var configServiceInstance = config.NewService()

// getConfigPath helper function to get config path from args or global flag
func getConfigPath(args []string) string {
	if len(args) > 0 {
		return args[0]
	}
	return configFile // Use global flag from root.go
}

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
	Long: `Manage HamaShell configuration including validation, display, listing, and generation.

Examples:
  hama-shell config validate
  hama-shell config show
  hama-shell config list
  hama-shell config generate`,
}

// configValidateCmd validates .yaml holding config
var configValidateCmd = &cobra.Command{
	Use:   "validate [config-file]",
	Short: "Validate configuration file",
	Long: `Validate the configuration file syntax and structure.

Examples:
  hama-shell config validate
  hama-shell config validate /path/to/config.yaml`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgPath := getConfigPath(args)

		if cfgPath != "" {
			fmt.Printf("Validating configuration file: %s\n", cfgPath)
		} else {
			fmt.Println("Validating default configuration...")
		}

		// Use config service to load and validate
		cfg, err := configServiceInstance.Load(cfgPath)
		if err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}

		fmt.Printf("✓ Configuration is valid (%d projects found)\n", len(cfg.Projects))
		return nil
	},
}

// configShowCmd shows .yaml holding config
var configShowCmd = &cobra.Command{
	Use:   "show [config-file]",
	Short: "Show configuration structure",
	Long: `Show the configuration file structure in a readable format.

Examples:
  hama-shell config show
  hama-shell config show /path/to/config.yaml`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgPath := getConfigPath(args)

		// Load configuration
		cfg, err := configServiceInstance.Load(cfgPath)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Display configuration structure
		showConfigStructure(cfg)
		return nil
	},
}

// showConfigStructure displays the config structure in a readable format
func showConfigStructure(cfg *config.Config) {
	fmt.Println("Configuration Structure:")
	fmt.Println("========================")

	if cfg == nil || len(cfg.Projects) == 0 {
		fmt.Println("No projects found")
		return
	}

	// Display projects with emojis for better UX
	for projectName, project := range cfg.Projects {
		fmt.Printf("\n📁 Project: %s\n", projectName)
		if project.Description != "" {
			fmt.Printf("   Description: %s\n", project.Description)
		}

		for stageName, stage := range project.Stages {
			fmt.Printf("\n   🔧 Stage: %s\n", stageName)
			if stage.Description != "" {
				fmt.Printf("      Description: %s\n", stage.Description)
			}

			for serviceName, service := range stage.Services {
				fmt.Printf("\n      💻 Service: %s\n", serviceName)
				if service.Description != "" {
					fmt.Printf("         Description: %s\n", service.Description)
				}

				fmt.Println("         Commands:")
				for i, command := range service.Commands {
					// Truncate long commands for display
					displayCmd := command
					if len(displayCmd) > 60 {
						displayCmd = displayCmd[:57] + "..."
					}
					fmt.Printf("         %d. %s\n", i+1, displayCmd)
				}
			}
		}
	}

	// Show global settings with emoji
	fmt.Println("\n⚙️  Global Settings:")
	fmt.Println("==================")
	fmt.Printf("Timeout: %d seconds\n", cfg.GlobalSettings.Timeout)
	fmt.Printf("Retries: %d\n", cfg.GlobalSettings.Retries)
	fmt.Printf("Auto Restart: %t\n", cfg.GlobalSettings.AutoRestart)
}

// configListCmd lists all available targets
var configListCmd = &cobra.Command{
	Use:   "list [config-file]",
	Short: "List all available targets",
	Long: `List all available targets in project.stage.service format.

Examples:
  hama-shell config list
  hama-shell config list /path/to/config.yaml`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgPath := getConfigPath(args)

		// Load configuration
		cfg, err := configServiceInstance.Load(cfgPath)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Get and display targets
		targets := configServiceInstance.List(cfg)
		if len(targets) == 0 {
			fmt.Println("No targets found in configuration")
			return nil
		}

		fmt.Println("Available targets:")
		for _, target := range targets {
			// Resolve to get description if available
			if svc, err := configServiceInstance.ResolveTarget(target, cfg); err == nil && svc.Description != "" {
				fmt.Printf("  %s - %s\n", target, svc.Description)
			} else {
				fmt.Printf("  %s\n", target)
			}
		}
		fmt.Printf("\nTotal: %d targets\n", len(targets))
		return nil
	},
}

// configGenerateCmd generates a configuration file interactively
var configGenerateCmd = &cobra.Command{
	Use:   "generate [output-file]",
	Short: "Generate configuration file interactively",
	Long: `Generate a configuration file through an interactive wizard.

Examples:
  hama-shell config generate
  hama-shell config generate ./my-config.yaml`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		outputPath := "hama-shell.yaml"
		if len(args) > 0 {
			outputPath = args[0]
		}

		fmt.Println("🚀 HamaShell Configuration Generator")
		fmt.Println("====================================")
		fmt.Println()

		// Create new config
		cfg := &config.Config{
			Projects:       make(map[string]config.Project),
			GlobalSettings: config.GlobalSettings{
				Timeout:     30,
				Retries:     3,
				AutoRestart: false,
			},
		}

		// Interactive configuration
		reader := bufio.NewReader(os.Stdin)

		// Add at least one project
		for {
			fmt.Print("\n📁 Enter project name (e.g., myapp): ")
			projectName, _ := reader.ReadString('\n')
			projectName = strings.TrimSpace(projectName)
			if projectName == "" {
				fmt.Println("Project name cannot be empty")
				continue
			}

			fmt.Print("   Project description (optional): ")
			projectDesc, _ := reader.ReadString('\n')
			projectDesc = strings.TrimSpace(projectDesc)

			project := config.Project{
				Description: projectDesc,
				Stages:      make(map[string]config.Stage),
			}

			// Add at least one stage
			for {
				fmt.Print("\n🔧 Enter stage name (e.g., dev, prod): ")
				stageName, _ := reader.ReadString('\n')
				stageName = strings.TrimSpace(stageName)
				if stageName == "" {
					fmt.Println("Stage name cannot be empty")
					continue
				}

				fmt.Print("   Stage description (optional): ")
				stageDesc, _ := reader.ReadString('\n')
				stageDesc = strings.TrimSpace(stageDesc)

				stage := config.Stage{
					Description: stageDesc,
					Services:    make(map[string]config.Service),
				}

				// Add at least one service
				for {
					fmt.Print("\n💻 Enter service name (e.g., api, database): ")
					serviceName, _ := reader.ReadString('\n')
					serviceName = strings.TrimSpace(serviceName)
					if serviceName == "" {
						fmt.Println("Service name cannot be empty")
						continue
					}

					fmt.Print("   Service description (optional): ")
					serviceDesc, _ := reader.ReadString('\n')
					serviceDesc = strings.TrimSpace(serviceDesc)

					// Add commands
					fmt.Println("\n📝 Enter commands (one per line, empty line to finish):")
					var commands []string
					for {
						fmt.Print("   > ")
						command, _ := reader.ReadString('\n')
						command = strings.TrimSpace(command)
						if command == "" {
							break
						}
						commands = append(commands, command)
					}

					if len(commands) == 0 {
						fmt.Println("At least one command is required")
						continue
					}

					service := config.Service{
						Description: serviceDesc,
						Commands:    commands,
					}
					stage.Services[serviceName] = service

					// Ask if want to add more services
					fmt.Print("\nAdd another service to this stage? (y/n): ")
					answer, _ := reader.ReadString('\n')
					if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(answer)), "y") {
						break
					}
				}

				project.Stages[stageName] = stage

				// Ask if want to add more stages
				fmt.Print("\nAdd another stage to this project? (y/n): ")
				answer, _ := reader.ReadString('\n')
				if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(answer)), "y") {
					break
				}
			}

			cfg.Projects[projectName] = project

			// Ask if want to add more projects
			fmt.Print("\nAdd another project? (y/n): ")
			answer, _ := reader.ReadString('\n')
			if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(answer)), "y") {
				break
			}
		}

		// Configure global settings
		fmt.Println("\n⚙️  Global Settings Configuration")
		fmt.Println("==================================")

		fmt.Printf("Timeout in seconds (default 30): ")
		timeoutStr, _ := reader.ReadString('\n')
		timeoutStr = strings.TrimSpace(timeoutStr)
		if timeoutStr != "" {
			var timeout int
			if _, err := fmt.Sscanf(timeoutStr, "%d", &timeout); err == nil && timeout > 0 {
				cfg.GlobalSettings.Timeout = timeout
			}
		}

		fmt.Printf("Number of retries (default 3): ")
		retriesStr, _ := reader.ReadString('\n')
		retriesStr = strings.TrimSpace(retriesStr)
		if retriesStr != "" {
			var retries int
			if _, err := fmt.Sscanf(retriesStr, "%d", &retries); err == nil && retries >= 0 {
				cfg.GlobalSettings.Retries = retries
			}
		}

		fmt.Print("Enable auto-restart? (y/n, default n): ")
		autoRestartStr, _ := reader.ReadString('\n')
		cfg.GlobalSettings.AutoRestart = strings.HasPrefix(strings.ToLower(strings.TrimSpace(autoRestartStr)), "y")

		// Write configuration to file
		data, err := yaml.Marshal(cfg)
		if err != nil {
			return fmt.Errorf("failed to marshal config: %w", err)
		}

		if err := os.WriteFile(outputPath, data, 0644); err != nil {
			return fmt.Errorf("failed to write config file: %w", err)
		}

		fmt.Printf("\n✅ Configuration saved to: %s\n", outputPath)
		fmt.Println("\nYou can now use:")
		fmt.Printf("  hama-shell config validate %s\n", outputPath)
		fmt.Printf("  hama-shell config show %s\n", outputPath)
		fmt.Printf("  hama-shell start <target> --config %s\n", outputPath)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)

	// Add subcommands
	configCmd.AddCommand(configValidateCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configGenerateCmd)
}
