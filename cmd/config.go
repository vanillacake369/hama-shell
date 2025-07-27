package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"hama-shell/internal/core/config"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
	Long: `Manage HamaShell configuration including validation, templates, and settings.

Examples:
  hama-shell config validate
  hama-shell config show
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
	Run: func(cmd *cobra.Command, args []string) {
		var configFile string
		if len(args) > 0 {
			configFile = args[0]
		}

		if configFile == "" {
			fmt.Println("Validating default configuration...")
		} else {
			fmt.Printf("Validating configuration file: %s\n", configFile)
		}

		// Create validator
		validator := config.NewValidator()

		// If specific config file provided, load it temporarily
		if configFile != "" {
			// Save current viper config
			currentConfig := viper.AllSettings()

			// Load specific file into viper temporarily
			viper.Reset()
			viper.SetConfigFile(configFile)
			if err := viper.ReadInConfig(); err != nil {
				fmt.Printf("Error reading config file: %s\n", err)
				return
			}

			// Validate the temporary config
			if err := validator.ValidateViper(); err != nil {
				fmt.Printf("Validation failed: %s\n", err)
				return
			}

			// Restore original viper config
			viper.Reset()
			for key, value := range currentConfig {
				viper.Set(key, value)
			}
		} else {
			// Validate current viper config
			if err := validator.ValidateViper(); err != nil {
				fmt.Printf("Validation failed: %s\n", err)
				return
			}
		}

		fmt.Println("✓ Configuration is valid")
	},
}

// configShowCmd shows .yaml holding config
var configShowCmd = &cobra.Command{
	Use:   "show [config-file]",
	Short: "Show configuration file",
	Long: `Show the configuration file syntax and structure.

Examples:
  hama-shell config show
  hama-shell config show /path/to/config.yaml`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var configFile string
		if len(args) > 0 {
			configFile = args[0]
		}

		if configFile == "" {
			fmt.Println("Showing default configuration...")
		} else {
			fmt.Printf("Showing configuration file: %s\n", configFile)
		}

		// Create validator
		validator := config.NewValidator()

		// If specific config file provided, show that file
		if configFile != "" {
			// Save current viper config
			currentConfig := viper.AllSettings()

			// Load specific file into viper temporarily
			viper.Reset()
			viper.SetConfigFile(configFile)
			if err := viper.ReadInConfig(); err != nil {
				fmt.Printf("Error reading config file: %s\n", err)
				return
			}

			// Show the config structure
			showConfigStructure(validator)

			// Restore original viper config
			viper.Reset()
			for key, value := range currentConfig {
				viper.Set(key, value)
			}
		} else {
			// Show current viper config
			showConfigStructure(validator)
		}
	},
}

// showConfigStructure displays the config structure in a readable format
func showConfigStructure(validator *config.Validator) {
	fmt.Println("Configuration Structure:")
	fmt.Println("========================")

	projects := validator.GetProjects()
	if len(projects) == 0 {
		fmt.Println("No projects found")
		return
	}

	for _, projectName := range projects {
		fmt.Printf("Project: %s\n", projectName)

		stages := validator.GetStages(projectName)
		for _, stageName := range stages {
			fmt.Printf("  Stage: %s\n", stageName)

			services := validator.GetServices(projectName, stageName)
			for _, serviceName := range services {
				fmt.Printf("    Service: %s\n", serviceName)

				// Show commands for this service
				commands := viper.GetStringSlice(fmt.Sprintf("projects.%s.stages.%s.services.%s.commands", projectName, stageName, serviceName))
				for i, command := range commands {
					fmt.Printf("      Command[%d]: %s\n", i+1, command)
				}
			}
		}
		fmt.Println()
	}

	// Show global settings
	fmt.Println("Global Settings:")
	fmt.Println("================")
	if viper.IsSet("global_settings.timeout") {
		fmt.Printf("Timeout: %d\n", viper.GetInt("global_settings.timeout"))
	}
	if viper.IsSet("global_settings.retries") {
		fmt.Printf("Retries: %d\n", viper.GetInt("global_settings.retries"))
	}
	if viper.IsSet("global_settings.auto_restart") {
		fmt.Printf("Auto Restart: %t\n", viper.GetBool("global_settings.auto_restart"))
	}
}

// configGenerateCmd generates .yaml holding config
var configGenerateCmd = &cobra.Command{
	Use:   "generate [config-file]",
	Short: "Generate configuration file",
	Long: `Generate the configuration file syntax and structure.

Examples:
  hama-shell config generate
  hama-shell config generate /path/to/config.yaml`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var configFile string
		if len(args) > 0 {
			configFile = args[0]
		}

		if configFile == "" {
			fmt.Println("Generating default configuration...")
		} else {
			fmt.Printf("Generating configuration file: %s\n", configFile)
		}

		// ToDo : Config 선언
		//	 - 어떤 파일명으로 config.yaml 선언할 것인지?
		//		Step 1: Configuration File Setup
		//      "Let's start by setting up your configuration. What would you like to name your config.yaml file?"
		//	 - 어떤 프로젝트?
		//		Step 2: Project Selection
		//      "Which project are you configuring? This helps organize your connections by project scope."
		//	 - 어떤 서비스? (db, api-server, gitlab runner ,,,)
		//		Step 3: Service Definition
		//      "What type of service are you connecting to? (e.g., database, API server, GitLab runner, etc.)"
		//	 - 어떤 스테이지? (dev, prod ,,)
		//		Step 4: Environment Stage
		//      "Which environment stage is this for? (e.g., development, production, staging, etc.)"
		//	 - 어떤 명령어? (한 줄 한 줄 입력받되, 빈 줄 입력 시 명령어 입력 단계 종료)
		//		Step 5: Commands Input
		//      "Now let's define the commands for this connection. Enter each command on a separate line. When you're finished, press Enter on an empty line to continue."
		//   - 입력한 명령어 최종 확인 (y -> yes 로 입력받아 다음 단계로 넘어감, n -> no 로 입력받아 명령어 다시 입력받게끔 처리)
		//		Step 6: Commands Confirmation
		//      "Please review your commands below. Type 'y' to confirm and proceed, or 'n' to edit them again."
		// 	 - 글로벌 세팅 (재)설정
		//		Step 7: Global Settings
		//      "Finally, let's configure your global settings. These will apply across all your connections."
	},
}

func init() {
	rootCmd.AddCommand(configCmd)

	// Add subcommands
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configValidateCmd)
	configCmd.AddCommand(configGenerateCmd)

	// Flags
	configShowCmd.Flags().BoolP("paths", "p", false, "Show configuration file paths")
}
