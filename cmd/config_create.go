package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"hama-shell/types"
)

var (
	createForce    bool
	createTemplate string
)

// configCreateCmd represents the config create command
var configCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new configuration file",
	Long: `Create a new hama-shell configuration file with default settings.
	
Examples:
  hs config create
  hs config create --force
  hs config create --template minimal`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get config file path
		configPath, err := getConfigPath(cmd)
		if err != nil {
			return err
		}

		// Check if config file already exists
		if _, err := os.Stat(configPath); err == nil && !createForce {
			fmt.Printf("Configuration file already exists: %s\n", configPath)
			fmt.Println("Use --force flag to overwrite")
			return nil
		}

		// Create config directory if it doesn't exist
		configDir := filepath.Dir(configPath)
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %v", err)
		}

		// Create config based on template
		var config *types.Config
		switch createTemplate {
		case "minimal":
			config = createMinimalConfig()
		case "full":
			config = createFullConfig()
		default:
			config = types.DefaultConfig()
		}

		// Set default paths
		homeDir, _ := os.UserHomeDir()
		if config.Settings.LogDir == "/var/log/hama-shell" {
			// Use user-specific log directory if /var/log is not writable
			config.Settings.LogDir = filepath.Join(homeDir, ".hama-shell", "logs")
		}

		// Convert to JSON
		data, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal config: %v", err)
		}

		// Write to file
		if err := ioutil.WriteFile(configPath, data, 0644); err != nil {
			return fmt.Errorf("failed to write config file: %v", err)
		}

		fmt.Printf("Configuration file created: %s\n", configPath)
		fmt.Println("\nDefault configuration:")
		fmt.Println("=====================================")
		fmt.Println(string(data))
		fmt.Println("\nUse 'hs config edit' to modify the configuration")
		fmt.Println("Use 'hs config add' to add new commands")

		return nil
	},
}

// createMinimalConfig creates a minimal configuration
func createMinimalConfig() *types.Config {
	return &types.Config{
		Version:  "1.0.0",
		Commands: []types.CommandConfig{},
		Settings: types.Settings{
			LogDir:       "~/.hama-shell/logs",
			DefaultShell: "/bin/bash",
			AutoRestart:  false,
			MaxRetries:   3,
		},
	}
}

// createFullConfig creates a full example configuration
func createFullConfig() *types.Config {
	config := types.DefaultConfig()
	
	// Add example commands
	config.Commands = []types.CommandConfig{
		{
			ID:            "web-server",
			Command:       "npm",
			Args:          []string{"run", "dev"},
			WorkingDir:    "~/projects/my-app",
			AutoStart:     true,
			RestartPolicy: types.RestartOnFailure,
			MaxRetries:    5,
			Description:   "Development web server",
			Tags:          []string{"web", "development"},
		},
		{
			ID:            "db-backup",
			Command:       "pg_dump",
			Args:          []string{"mydb", "-f", "backup.sql"},
			WorkingDir:    "~/backups",
			AutoStart:     false,
			RestartPolicy: types.RestartNever,
			Description:   "Database backup job",
			Tags:          []string{"database", "backup"},
		},
		{
			ID:            "worker",
			Command:       "python",
			Args:          []string{"worker.py"},
			WorkingDir:    "~/projects/worker",
			AutoStart:     true,
			RestartPolicy: types.RestartAlways,
			MaxRetries:    10,
			Description:   "Background worker process",
			Tags:          []string{"worker", "python"},
			Env: map[string]string{
				"PYTHONPATH": "~/projects/worker/lib",
				"LOG_LEVEL":  "INFO",
			},
		},
	}
	
	return config
}

func init() {
	configCmd.AddCommand(configCreateCmd)
	
	// Add flags
	configCreateCmd.Flags().BoolVarP(&createForce, "force", "f", false, "Overwrite existing configuration file")
	configCreateCmd.Flags().StringVarP(&createTemplate, "template", "t", "default", "Template to use (minimal, default, full)")
}