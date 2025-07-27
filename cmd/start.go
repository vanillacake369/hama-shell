package cmd

import (
	"fmt"
	"github.com/spf13/viper"

	"github.com/spf13/cobra"
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
		commands := viper.GetStringSlice(sessionPath + ".command")

		if len(commands) == 0 {
			fmt.Printf("No commands found for session: %s\n", sessionPath)
			return
		}

		fmt.Printf("Starting session '%s' with %d commands...\n", sessionPath, len(commands))
		// TODO: Execute commands
	},
}

func init() {
	rootCmd.AddCommand(startCmd)

	startCmd.Flags().BoolP("all", "a", false, "Start all configured sessions")
}
