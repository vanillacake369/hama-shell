package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// aliasCmd represents the alias command
var aliasCmd = &cobra.Command{
	Use:   "alias",
	Short: "Manage session aliases",
	Long: `Manage session aliases for quick access to session paths.

Examples:
  hama-shell alias list
  hama-shell alias add my-dev project.dev.john.backend
  hama-shell alias remove my-dev`,
}

var aliasListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all aliases",
	Long:  `List all configured aliases and their corresponding session paths.`,
	Run: func(cmd *cobra.Command, args []string) {
		json, _ := cmd.Flags().GetBool("json")

		if json {
			fmt.Println("Listing aliases in JSON format...")
		} else {
			fmt.Println("Configured aliases:")
		}

		// TODO: Implement alias list logic
	},
}

var aliasAddCmd = &cobra.Command{
	Use:   "add [alias] [session-path]",
	Short: "Add a new alias",
	Long: `Add a new alias for a session path.

Examples:
  hama-shell alias add my-dev project.dev.john.backend
  hama-shell alias add prod-db project.prod.admin.database`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		alias := args[0]
		sessionPath := args[1]

		fmt.Printf("Adding alias '%s' for session path '%s'...\n", alias, sessionPath)

		// TODO: Implement alias add logic
	},
}

var aliasRemoveCmd = &cobra.Command{
	Use:     "remove [alias]",
	Aliases: []string{"rm", "delete"},
	Short:   "Remove an alias",
	Long: `Remove an existing alias.

Examples:
  hama-shell alias remove my-dev
  hama-shell alias rm my-dev`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		alias := args[0]
		force, _ := cmd.Flags().GetBool("force")

		if force {
			fmt.Printf("Force removing alias '%s'...\n", alias)
		} else {
			fmt.Printf("Removing alias '%s'...\n", alias)
		}

		// TODO: Implement alias remove logic
	},
}

var aliasResolveCmd = &cobra.Command{
	Use:   "resolve [alias]",
	Short: "Resolve an alias to its session path",
	Long: `Resolve an alias to show its corresponding session path.

Examples:
  hama-shell alias resolve my-dev`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		alias := args[0]

		fmt.Printf("Resolving alias '%s'...\n", alias)

		// TODO: Implement alias resolve logic
	},
}

func init() {
	rootCmd.AddCommand(aliasCmd)

	// Add subcommands
	aliasCmd.AddCommand(aliasListCmd)
	aliasCmd.AddCommand(aliasAddCmd)
	aliasCmd.AddCommand(aliasRemoveCmd)
	aliasCmd.AddCommand(aliasResolveCmd)

	// Flags
	aliasListCmd.Flags().BoolP("json", "j", false, "Output in JSON format")
	aliasRemoveCmd.Flags().BoolP("force", "f", false, "Force remove without confirmation")
}
