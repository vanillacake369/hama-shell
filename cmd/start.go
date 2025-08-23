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
  hama-shell start myapp.dev.api
  hama-shell start monitoring.prod.prometheus`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		target := args[0]

		// Delegate to session manager
		if err := sessionManager.Start(target, configFile); err != nil {
			return fmt.Errorf("failed to start session: %w", err)
		}

		fmt.Printf("✓ Started session: %s\n", target)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
