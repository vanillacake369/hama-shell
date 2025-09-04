package cmd

import (
	"github.com/spf13/cobra"
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

func init() {
	rootCmd.AddCommand(configCmd)
}