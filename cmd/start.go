package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"hama-shell/internal/core/executor"

	"hama-shell/internal/core/config"
)

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

		fmt.Printf("Starting session '%s' with %d commands in keep-alive mode...\n", sessionPath, len(commands))
		executor := executor.New()
		
		// Run each command independently
		for _, command := range commands {
			if err := executor.Run(sessionPath, command); err != nil {
				fmt.Printf("Error starting command '%s': %s\n", command, err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(startCmd)

	startCmd.Flags().BoolP("all", "a", false, "Start all configured sessions")
}
