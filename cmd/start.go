package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"hama-shell/internal/core/config"
	"hama-shell/internal/core/executor"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start [session-path]",
	Short: "Start a session",
	Long: `Start a session based on the configuration path.

Examples:
  hama-shell start project.stage.developer.session
  hama-shell start --all`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		all, _ := cmd.Flags().GetBool("all")

		if all {
			fmt.Println("Starting all sessions...")
			// TODO: Implement start all logic
			return
		}

		if len(args) == 0 {
			fmt.Println("Error: session path required")
			cmd.Help()
			return
		}

		sessionPath := args[0]

		// Check if AppConfig is available
		if AppConfig == nil {
			fmt.Println("Error: Configuration not loaded")
			return
		}

		// Get commands from static config
		commands, err := config.GetCommands(AppConfig, sessionPath)
		if err != nil {
			fmt.Printf("Error getting commands for session '%s': %s\n", sessionPath, err)
			return
		}

		if len(commands) == 0 {
			fmt.Printf("No commands found for session: %s\n", sessionPath)
			return
		}

		fmt.Printf("Starting session '%s' with %d commands...\n", sessionPath, len(commands))

		// Create command executor with configured timeout
		timeout := AppConfig.GlobalSettings.Timeout
		executor := executor.NewCommandExecutor(time.Duration(timeout) * time.Second)

		// Execute commands sequentially
		// ToDo : How can I process keepAlive on executor ?
		// ToDo : How can I process keepAlive on executor ?
		// ToDo : How can I process keepAlive on executor ?
		results, err := executor.ExecuteCommands(commands)
		if err != nil {
			fmt.Printf("Error executing commands: %s\n", err)
			return
		}

		// Check if any command failed
		for _, result := range results {
			if result.Error != nil {
				fmt.Printf("Command failed: %s - %s\n", result.Command, result.Error)
				return
			}
		}

		fmt.Printf("Session '%s' completed successfully\n", sessionPath)
	},
}

func init() {
	rootCmd.AddCommand(startCmd)

	startCmd.Flags().BoolP("all", "a", false, "Start all configured sessions")
}
