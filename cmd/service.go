package cmd

import "github.com/spf13/cobra"

// sessionCmd represents the config command
var sessionCmd = &cobra.Command{}

// sessionStartCmd represents the config command
var sessionStartCmd = &cobra.Command{}

func init() {
	rootCmd.AddCommand(sessionCmd)

	sessionCmd.AddCommand(sessionStartCmd)
}
