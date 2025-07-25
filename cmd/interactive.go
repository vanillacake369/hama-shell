package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// interactiveCmd represents the interactive command
var interactiveCmd = &cobra.Command{
	Use:     "interactive",
	Aliases: []string{"i", "tui"},
	Short:   "Start interactive TUI mode",
	Long: `Start HamaShell in interactive Terminal User Interface (TUI) mode.

The interactive mode provides:
- Visual session management dashboard
- Real-time status monitoring
- Easy session start/stop controls
- Configuration browsing
- Log viewing`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting HamaShell interactive mode...")
		fmt.Println("Press 'q' to quit, 'h' for help")

		// TODO: Implement TUI logic
		fmt.Println("Interactive mode not yet implemented")
	},
}

func init() {
	rootCmd.AddCommand(interactiveCmd)
}
