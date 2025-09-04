/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)



// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "hs",
	Short: "Hama Shell - Session management tool",
	Long: `Hama Shell (hs) is a powerful session management tool that helps you
manage long-running processes, attach/detach from sessions,
and maintain command configurations.

Examples:
  hs ls                    # List all sessions
  hs web-server status     # Check status of web-server session
  hs web-server attach     # Attach to web-server session
  hs config add           # Add new command to configuration`,
	// SilenceUsage prevents usage from being printed on every error
	SilenceUsage: true,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	// Handle dynamic session commands before executing
	args := os.Args[1:]
	if len(args) > 0 {
		// Check if the first argument might be a session ID
		firstArg := args[0]
		
		// Skip if it's a known command
		knownCommands := []string{"list", "ls", "config", "help", "completion", "--help", "-h", "--version", "-v"}
		isKnownCommand := false
		for _, cmd := range knownCommands {
			if firstArg == cmd {
				isKnownCommand = true
				break
			}
		}
		
		// If it's not a known command, treat it as a session ID
		if !isKnownCommand {
			if err := handleDynamicSession(rootCmd, args); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return
		}
	}
	
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringP("config", "c", "", "config file path (default is $HOME/.hama-shell.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("version", "v", false, "Show version information")
}


