package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"

	"hama-shell/internal/core/config"
)

// AppConfig holds the parsed and validated configuration
var AppConfig *config.Config

// configFile holds the path to the configuration file
var configFile string

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

	// Add config file flag
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file (default is $HOME/hama-shell.yaml or ./hama-shell.yaml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// Create validator and parse config
	validator := config.NewValidator()

	// Try to parse and validate config
	var err error
	AppConfig, err = validator.ParseAndValidate(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
		fmt.Fprintf(os.Stderr, "Please check your configuration file or run 'hama-shell config generate' to create one.\n")
		os.Exit(1)
	}

	// Also keep viper config for backward compatibility during transition
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get user home directory: %v\n", err)
		os.Exit(1)
	}

	viper.AddConfigPath(home)
	viper.AddConfigPath(".")
	viper.SetConfigType("yaml")
	viper.SetConfigName("hama-shell")

	viper.AutomaticEnv()
	_ = viper.ReadInConfig() // Ignore errors for now as we have static config
}
