package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var readConfigCmd = &cobra.Command{
	Use:   "read-config",
	Short: "Read config file",
	Long:  "Read config file",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Reading config...")
	},
}

func init() {
	rootCmd.AddCommand(readConfigCmd)
}
