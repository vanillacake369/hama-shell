package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "hama-shell",
	Short: "A session and connection manager for developers",
	Long: `HamaShell is a session and connection manager designed for developers who need reliable, secure access to various hosts in single CLI command.

It simplifies complex multi-step SSH tunneling and session setup by letting developers define their connections declaratively in a YAML file.

Key Features:
- Declarative YAML-based configuration
- Multi-step SSH tunneling and port forwarding
- Terminal multiplexer integration (tmux, zellij, screen)
- Session state management and persistence
- Cross-platform support`,
	Run: func(cmd *cobra.Command, args []string) {
		// This forces initConfig to run
		fmt.Println("Root command executed - config should be loaded")
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	home, err := os.UserHomeDir()
	cobra.CheckErr(err)

	viper.AddConfigPath(home)
	viper.AddConfigPath(".")
	viper.SetConfigType("yaml")
	viper.SetConfigName("hama-shell")

	viper.AutomaticEnv()
	err = viper.ReadInConfig()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "No config file found: %v\n", err)
		// ToDo : make a config file ${home}/hama-shell.yaml
	}
}
