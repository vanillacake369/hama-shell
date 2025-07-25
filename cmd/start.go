package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start [session-path]",
	Short: "Start a session",
	Long: `Start a session based on the configuration path or alias.

Examples:
  hama-shell start project.stage.developer.session
  hama-shell start my-alias
  hama-shell start --all`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		all, _ := cmd.Flags().GetBool("all")
		background, _ := cmd.Flags().GetBool("background")

		if all {
			fmt.Println("Starting all sessions...")
			// TODO: Implement start all logic
			return
		}

		if len(args) == 0 {
			fmt.Println("Error: session path or alias required")
			cmd.Help()
			return
		}

		sessionPath := args[0]
		if background {
			fmt.Printf("Starting session '%s' in background...\n", sessionPath)
		} else {
			fmt.Printf("Starting session '%s'...\n", sessionPath)
		}

		// TODO: Implement start session logic
	},
}

func init() {
	rootCmd.AddCommand(startCmd)

	startCmd.Flags().BoolP("all", "a", false, "Start all configured sessions")
	startCmd.Flags().BoolP("background", "b", false, "Start session in background")
	startCmd.Flags().StringP("config", "c", "", "Config file path")
}
