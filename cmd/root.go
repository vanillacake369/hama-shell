package cmd

import (
	"fmt"
	"os"
	"slices"

	"github.com/spf13/cobra"
)

// Version information
var (
	version = "v0.0.1"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "hs",
	Short: "Hama Shell - Session management tool",
	Long: `Hama Shell (hs) is a powerful session management tool that helps you
		manage long-running processes, attach/detach from sessions,
		and maintain command configurations.`,
	// SilenceUsage prevents usage from being printed on every error
	SilenceUsage: true,
	Run: func(cmd *cobra.Command, args []string) {
		// Check if version flag is set
		versionFlag, _ := cmd.Flags().GetBool("version")
		if versionFlag {
			fmt.Printf("hama-shell %s\n", version)
			return
		}
		// If no flags or subcommands, show help
		err := cmd.Help()
		if err != nil {
			return
		}
	},
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
		isKnownCommand := slices.Contains(knownCommands, firstArg)

		// If it's not a known command, treat it as a session ID
		if !isKnownCommand {
			err := handleDynamicSession(rootCmd, args)
			if err != nil {
				_, _ = fmt.Fprintln(os.Stderr, "Error:", err)
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

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("version", "v", false, "Show version information")
}
