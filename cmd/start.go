package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
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

		fmt.Printf("Starting session '%s' with %d commands in keep-alive mode...\n", sessionPath, len(commands))

		// Create command executor with configured timeout
		timeout := AppConfig.GlobalSettings.Timeout
		executor := executor.NewCommandExecutor(time.Duration(timeout) * time.Second)

		// ToDo : 어떻게 하면 테스트 할 수 있을까?
		// ToDo : 어떻게 하면 테스트 할 수 있을까?
		// ToDo : 어떻게 하면 테스트 할 수 있을까?
		// ToDo : 어떻게 하면 테스트 할 수 있을까?
		// Execute commands in keep-alive mode (background processes)
		err = executor.ExecuteCommandsKeepAlive(commands)
		if err != nil {
			fmt.Printf("Error starting keep-alive processes: %s\n", err)
			return
		}

		fmt.Printf("Session '%s' started with %d background processes\n", sessionPath, len(commands))
		fmt.Println("All processes are running in background with auto-restart enabled")
		fmt.Println("Press Ctrl+C to stop all processes...")

		// Set up signal catching for graceful shutdown
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)

		// ToDo : interrupt signal 을 Ctrl C 로 처리하는 게 맞을까?
		// ToDo : interrupt signal 을 Ctrl C 로 처리하는 게 맞을까?
		// ToDo : interrupt signal 을 Ctrl C 로 처리하는 게 맞을까?
		// ToDo : interrupt signal 을 Ctrl C 로 처리하는 게 맞을까?
		// Wait for interrupt signal
		<-c
		fmt.Println("\nReceived interrupt signal, stopping all processes...")

		// Stop all processes gracefully
		if err := executor.StopAll(); err != nil {
			fmt.Printf("Error stopping processes: %s\n", err)
		} else {
			fmt.Println("All processes stopped successfully")
		}
	},
}

func init() {
	rootCmd.AddCommand(startCmd)

	startCmd.Flags().BoolP("all", "a", false, "Start all configured sessions")
}
