package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// helpCmd represents the help command
var helpCmd = &cobra.Command{
	Use:   "help",
	Short: "Display help information",
	Long:  `Display detailed help information about hama-shell and its commands.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Hama Shell (hs) - Session Management Tool")
		fmt.Println("==========================================")
		fmt.Println()
		fmt.Println("USAGE:")
		fmt.Println("  hs [command] [flags]")
		fmt.Println("  hs <session> [subcommand] [flags]")
		fmt.Println()
		fmt.Println("CORE COMMANDS:")
		fmt.Println("  list (ls)              List all sessions")
		fmt.Println("  config                 Manage configuration")
		fmt.Println("  help                   Show this help message")
		fmt.Println()
		fmt.Println("SESSION COMMANDS:")
		fmt.Println("  hs <session> status              Show session status (default)")
		fmt.Println("  hs <session> attach (a)          Attach to session TTY")
		fmt.Println("  hs <session> detach              Detach from session")
		fmt.Println("  hs <session> kill (k)            Terminate session")
		fmt.Println("  hs <session> commands (cmds)     List registered commands")
		fmt.Println("  hs <session> logs                Show session logs")
		fmt.Println("  hs <session> restart             Restart session")
		fmt.Println()
		fmt.Println("CONFIG SUBCOMMANDS:")
		fmt.Println("  hs config view         Display configuration")
		fmt.Println("  hs config edit         Edit configuration file")
		fmt.Println("  hs config create       Create new configuration")
		fmt.Println("  hs config add          Add new command")
		fmt.Println()
		fmt.Println("EXAMPLES:")
		fmt.Println("  # List all sessions")
		fmt.Println("  hs ls")
		fmt.Println()
		fmt.Println("  # Check status of web-server session")
		fmt.Println("  hs web-server status")
		fmt.Println()
		fmt.Println("  # Attach to web-server session")
		fmt.Println("  hs web-server attach")
		fmt.Println()
		fmt.Println("  # Kill a session forcefully")
		fmt.Println("  hs worker-1 kill --force")
		fmt.Println()
		fmt.Println("  # Add new command interactively")
		fmt.Println("  hs config add")
		fmt.Println()
		fmt.Println("  # Add command with flags")
		fmt.Println("  hs config add --id myapp --command \"npm run dev\" --workdir ~/projects/myapp")
		fmt.Println()
		fmt.Println("KEY BINDINGS:")
		fmt.Println("  When attached to a session:")
		fmt.Println("    Ctrl+B then D    Detach from session")
		fmt.Println()
		fmt.Println("  During 'config add' interactive mode:")
		fmt.Println("    Ctrl+C           Cancel operation")
		fmt.Println()
		fmt.Println("CONFIGURATION:")
		fmt.Println("  Config file location (in order of precedence):")
		fmt.Println("    1. --config flag")
		fmt.Println("    2. $HAMA_SHELL_CONFIG environment variable")
		fmt.Println("    3. ~/.hama-shell.yaml (default)")
		fmt.Println()
		fmt.Println("For more information about a specific command, use:")
		fmt.Println("  hs [command] --help")
		fmt.Println()
		fmt.Println("PROJECT INFORMATION:")
		fmt.Println("  Version:     1.0.0")
		fmt.Println("  Repository:  https://github.com/yourusername/hama-shell")
		fmt.Println("  Issues:      https://github.com/yourusername/hama-shell/issues")
	},
}

func init() {
	rootCmd.AddCommand(helpCmd)
}