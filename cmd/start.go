package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

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
		commands := viper.GetStringSlice(sessionPath + ".commands")

		if len(commands) == 0 {
			fmt.Printf("No commands found for session: %s\n", sessionPath)
			return
		}

		fmt.Printf("Starting session '%s' with %d commands...\n", sessionPath, len(commands))

		// Create command executor with 30 second timeout
		executor := executor.NewCommandExecutor(30 * time.Second)

		// Execute commands sequentially
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
