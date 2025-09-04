package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

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
		// TODO: Implement actual session listing from session manager
		// For now, showing example output
		
		sessions := []struct {
			ID        string
			Status    string
			StartTime time.Time
			Command   string
		}{
			{
				ID:        "web-server",
				Status:    "running",
				StartTime: time.Now().Add(-2 * time.Hour),
				Command:   "npm run dev",
			},
			{
				ID:        "db-backup",
				Status:    "running",
				StartTime: time.Now().Add(-5 * time.Hour),
				Command:   "pg_dump mydb > backup.sql",
			},
			{
				ID:        "worker-1",
				Status:    "stopped",
				StartTime: time.Now().Add(-8 * time.Hour),
				Command:   "python worker.py",
			},
		}

		// Create tabwriter for aligned output
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "SESSION ID\tSTATUS\tSTART TIME\tCOMMAND")
		fmt.Fprintln(w, "----------\t------\t----------\t-------")
		
		for _, session := range sessions {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
				session.ID,
				session.Status,
				session.StartTime.Format("2006-01-02 15:04:05"),
				session.Command,
			)
		}
		
		w.Flush()
		
		// Show session count
		fmt.Printf("\nTotal sessions: %d\n", len(sessions))
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	
	// Add flags specific to list command
	listCmd.Flags().BoolP("all", "a", false, "Show all sessions including stopped ones")
	listCmd.Flags().StringP("status", "s", "", "Filter by status (running/stopped/failed)")
}