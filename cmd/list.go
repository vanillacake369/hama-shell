package cmd

import (
	"hama-shell/internal/session/api"
	"log"

	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all running sessions",
	Long: `Display all currently running hama-shell sessions with their ID, status, and start time.
	
Examples:
  hs list
  hs ls`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get flags
		showAll, _ := cmd.Flags().GetBool("all")
		statusFilter, _ := cmd.Flags().GetString("status")

		// Create session API
		sessionAPI := api.NewSessionAPI()

		// List sessions through API layer
		if err := sessionAPI.ListSessions(showAll, statusFilter); err != nil {
			log.Fatalf("Failed to list sessions: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	// Add flags specific to list command
	listCmd.Flags().BoolP("all", "a", false, "Show all sessions including stopped ones")
	listCmd.Flags().StringP("status", "s", "", "Filter by status (running/stopped/failed)")
}
