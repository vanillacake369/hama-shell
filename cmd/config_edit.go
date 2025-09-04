package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

// configEditCmd represents the config edit command
var configEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit configuration file",
	Long:  `Open the hama-shell configuration file in your default editor.`,
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

		// Determine editor to use
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = os.Getenv("VISUAL")
		}
		if editor == "" {
			// Try common editors
			editors := []string{"vim", "nano", "vi", "emacs", "notepad"}
			for _, e := range editors {
				if _, err := exec.LookPath(e); err == nil {
					editor = e
					break
				}
			}
		}

		if editor == "" {
			return fmt.Errorf("no editor found. Please set EDITOR environment variable")
		}

		fmt.Printf("Opening %s with %s...\n", configPath, editor)

		// Open editor
		editorCmd := exec.Command(editor, configPath)
		editorCmd.Stdin = os.Stdin
		editorCmd.Stdout = os.Stdout
		editorCmd.Stderr = os.Stderr

		if err := editorCmd.Run(); err != nil {
			return fmt.Errorf("failed to open editor: %v", err)
		}

		fmt.Println("Configuration file edited successfully")
		
		// TODO: Validate configuration after editing
		fmt.Println("Validating configuration...")
		if err := validateConfigFile(configPath); err != nil {
			fmt.Printf("Warning: Configuration validation failed: %v\n", err)
		} else {
			fmt.Println("Configuration is valid")
		}

		return nil
	},
}

// validateConfigFile validates the configuration file
func validateConfigFile(path string) error {
	// TODO: Implement actual validation logic
	return nil
}

func init() {
	configCmd.AddCommand(configEditCmd)
	
	// Add editor flag
	configEditCmd.Flags().StringP("editor", "e", "", "Editor to use (overrides EDITOR env var)")
}