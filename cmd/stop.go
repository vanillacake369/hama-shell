package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// stopCmd represents the stop command
var stopCmd = &cobra.Command{
	Use:   "stop [session-path]",
	Short: "Stop a running session",
	Long: `Stop a running session based on the session ID, path, or alias.

Examples:
  hama-shell stop project.stage.developer.session
  hama-shell stop my-alias
  hama-shell stop --all`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		all, _ := cmd.Flags().GetBool("all")
		force, _ := cmd.Flags().GetBool("force")

		if all {
			fmt.Println("Stopping all running sessions...")
			// TODO: Implement stop all logic
			return
		}

		if len(args) == 0 {
			fmt.Println("Error: session path, alias, or --all flag required")
			cmd.Help()
			return
		}

		sessionPath := args[0]
		if force {
			fmt.Printf("Force stopping session '%s'...\n", sessionPath)
		} else {
			fmt.Printf("Stopping session '%s'...\n", sessionPath)
		}

		// TODO: Implement stop session logic
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)

	stopCmd.Flags().BoolP("all", "a", false, "Stop all running sessions")
	stopCmd.Flags().BoolP("force", "f", false, "Force stop session")
}
