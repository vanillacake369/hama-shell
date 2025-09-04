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

// configViewCmd represents the config view command
var configViewCmd = &cobra.Command{
	Use:   "view",
	Short: "Display configuration file contents",
	Long:  `Display the contents of the hama-shell configuration file.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get config file path
		configPath, err := getConfigPath(cmd)
		if err != nil {
			return err
		}

		// Check if config file exists
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			fmt.Printf("Configuration file not found: %s\n", configPath)
			fmt.Println("Use 'hs config create' to create a new configuration file")
			return nil
		}

		// Read config file
		data, err := ioutil.ReadFile(configPath)
		if err != nil {
			return fmt.Errorf("failed to read config file: %v", err)
		}

		// Parse JSON to check if it's valid
		var config types.Config
		if err := json.Unmarshal(data, &config); err != nil {
			// If not valid JSON, display raw content
			fmt.Println("Configuration file (raw):")
			fmt.Println("=====================================")
			fmt.Println(string(data))
			return nil
		}

		// Display formatted JSON
		formatted, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format config: %v", err)
		}

		fmt.Printf("Configuration file: %s\n", configPath)
		fmt.Println("=====================================")
		fmt.Println(string(formatted))

		// Show summary
		fmt.Println("\n=====================================")
		fmt.Printf("Version:  %s\n", config.Version)
		fmt.Printf("Commands: %d registered\n", len(config.Commands))
		fmt.Printf("Log dir:  %s\n", config.Settings.LogDir)
		
		return nil
	},
}

// getConfigPath determines the configuration file path
func getConfigPath(cmd *cobra.Command) (string, error) {
	// Check if config flag is set
	configPath, _ := cmd.Flags().GetString("config")
	if configPath != "" {
		return configPath, nil
	}

	// Check environment variable
	if envPath := os.Getenv("HAMA_SHELL_CONFIG"); envPath != "" {
		return envPath, nil
	}

	// Use default path in home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %v", err)
	}

	return filepath.Join(homeDir, ".hama-shell.yaml"), nil
}

func init() {
	configCmd.AddCommand(configViewCmd)
}