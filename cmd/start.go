package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"hama-shell/internal/core/config"
	"hama-shell/internal/core/executor"
	"hama-shell/internal/core/session"
)

// sessionManager is a singleton instance used across commands
var sessionManager = session.NewManager(
	executor.New(),
	config.NewService(),
)

var startCmd = &cobra.Command{
	Use:   "start [project.stage.service]",
	Short: "Start a session",
	Long: `Start a session for the specified target.

The target should be in the format: project.stage.service

Examples:
  hama-shell start myapp.dev.api                    # Foreground (interactive)
  hama-shell start myapp.dev.api --detached         # Background (detached)
  hama-shell start monitoring.prod.prometheus`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		target := args[0]

		// Get detached flag
		detached, _ := cmd.Flags().GetBool("detached")

		// Determine execution mode
		var mode executor.ExecutionMode
		if detached {
			mode = executor.ExecutionModeBackground
		} else {
			mode = executor.ExecutionModeForeground
		}

		// Delegate to session manager with execution mode
		if err := sessionManager.StartWithMode(target, configFile, mode); err != nil {
			return fmt.Errorf("failed to start session: %w", err)
		}

		// Different messages for different modes
		if detached {
			fmt.Printf("✓ Started session: %s (background)\n", target)
		} else {
			fmt.Printf("✓ Session completed: %s\n", target)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
	
	// Add --detached flag for background execution
	startCmd.Flags().BoolP("detached", "d", false, "Run session in background (detached mode)")
}
