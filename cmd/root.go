package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "hama-shell",
	Short: "hama-shell is a session and connection manager for developers",
	Long:  "HamaShell is a session and connection manager designed for developers who need reliable, secure access to various hosts in single CLI command.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("HamaShell - Session Manager for Developers")
		fmt.Println("Use 'hama-shell --help' for available commands")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing HamaShell: %s\n", err)
		os.Exit(1)
	}
}
