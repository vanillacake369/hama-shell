package cmd

import (
	"github.com/spf13/cobra"
	
	"hama-shell/internal/core/configuration/api"
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
		configAPI := api.NewConfigAPI()
		return configAPI.ViewConfiguration()
	},
}

// configAddCmd represents the config add command
var configAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new command to configuration",
	Long:  `Add a new command to be executed in a hama-shell session.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		configAPI := api.NewConfigAPI()
		return configAPI.AddToConfiguration()
	},
}

// configCreateCmd represents the config create command
var configCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a configuration",
	Long: `Create a configuration that contains commands.

You can also provide command details via flags for non-interactive mode.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		configAPI := api.NewConfigAPI()
		return configAPI.CreateConfiguration()
	},
}


func init() {
	rootCmd.AddCommand(configCmd)

	configCmd.AddCommand(configViewCmd)
	configCmd.AddCommand(configAddCmd)
	configCmd.AddCommand(configCreateCmd)
}
