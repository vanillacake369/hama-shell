package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"hama-shell/types"
)

// configAddCmd represents the config add command
var configAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new command to configuration",
	Long: `Add a new command to be executed in a hama-shell session.

Interactive mode:
  - Enter command details when prompted
  - Press Alt+C to cancel at any time
  - Press Alt+F to finish and save

You can also provide command details via flags for non-interactive mode.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get config file path
		configPath, err := getConfigPath(cmd)
		if err != nil {
			return err
		}

		// Check if config file exists
		var config types.Config
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			fmt.Printf("Configuration file not found: %s\n", configPath)
			fmt.Println("Creating new configuration file...")
			config = *types.DefaultConfig()
		} else {
			// Read existing config
			data, err := ioutil.ReadFile(configPath)
			if err != nil {
				return fmt.Errorf("failed to read config file: %v", err)
			}
			if err := json.Unmarshal(data, &config); err != nil {
				return fmt.Errorf("failed to parse config file: %v", err)
			}
		}

		// Check for non-interactive mode (flags provided)
		sessionID, _ := cmd.Flags().GetString("id")
		command, _ := cmd.Flags().GetString("command")
		
		var newCommand types.CommandConfig
		
		if sessionID != "" && command != "" {
			// Non-interactive mode
			workDir, _ := cmd.Flags().GetString("workdir")
			autoStart, _ := cmd.Flags().GetBool("auto-start")
			restartPolicy, _ := cmd.Flags().GetString("restart")
			description, _ := cmd.Flags().GetString("description")
			
			newCommand = types.CommandConfig{
				ID:          sessionID,
				Command:     command,
				WorkingDir:  workDir,
				AutoStart:   autoStart,
				Description: description,
			}
			
			// Set restart policy
			switch restartPolicy {
			case "always":
				newCommand.RestartPolicy = types.RestartAlways
			case "on-failure":
				newCommand.RestartPolicy = types.RestartOnFailure
			default:
				newCommand.RestartPolicy = types.RestartNever
			}
		} else {
			// Interactive mode
			fmt.Println("Add new command to configuration")
			fmt.Println("=====================================")
			fmt.Println("Press Ctrl+C to cancel")
			fmt.Println()

			// Set up signal handler for Ctrl+C
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
			
			// Create goroutine to handle cancellation
			cancelled := false
			go func() {
				<-sigChan
				cancelled = true
				fmt.Println("\n\nOperation cancelled")
				os.Exit(0)
			}()

			reader := bufio.NewReader(os.Stdin)
			
			// Get session ID
			fmt.Print("Session ID (required): ")
			sessionID, _ = reader.ReadString('\n')
			sessionID = strings.TrimSpace(sessionID)
			if sessionID == "" {
				return fmt.Errorf("session ID is required")
			}

			// Check if session ID already exists
			for _, cmd := range config.Commands {
				if cmd.ID == sessionID {
					fmt.Printf("Warning: Session ID '%s' already exists. Overwrite? (y/N): ", sessionID)
					response, _ := reader.ReadString('\n')
					response = strings.ToLower(strings.TrimSpace(response))
					if response != "y" && response != "yes" {
						fmt.Println("Operation cancelled")
						return nil
					}
					// Remove existing command
					newCommands := []types.CommandConfig{}
					for _, c := range config.Commands {
						if c.ID != sessionID {
							newCommands = append(newCommands, c)
						}
					}
					config.Commands = newCommands
					break
				}
			}

			// Get command
			fmt.Print("Command to execute (required): ")
			command, _ = reader.ReadString('\n')
			command = strings.TrimSpace(command)
			if command == "" {
				return fmt.Errorf("command is required")
			}

			// Get working directory
			fmt.Print("Working directory (optional, press Enter to skip): ")
			workDir, _ := reader.ReadString('\n')
			workDir = strings.TrimSpace(workDir)

			// Get description
			fmt.Print("Description (optional, press Enter to skip): ")
			description, _ := reader.ReadString('\n')
			description = strings.TrimSpace(description)

			// Get auto-start preference
			fmt.Print("Auto-start on system boot? (y/N): ")
			autoStartResp, _ := reader.ReadString('\n')
			autoStartResp = strings.ToLower(strings.TrimSpace(autoStartResp))
			autoStart := autoStartResp == "y" || autoStartResp == "yes"

			// Get restart policy
			fmt.Print("Restart policy (never/always/on-failure) [never]: ")
			restartResp, _ := reader.ReadString('\n')
			restartResp = strings.TrimSpace(restartResp)
			if restartResp == "" {
				restartResp = "never"
			}

			// Parse command and args
			parts := strings.Fields(command)
			cmdName := parts[0]
			cmdArgs := []string{}
			if len(parts) > 1 {
				cmdArgs = parts[1:]
			}

			newCommand = types.CommandConfig{
				ID:          sessionID,
				Command:     cmdName,
				Args:        cmdArgs,
				WorkingDir:  workDir,
				AutoStart:   autoStart,
				Description: description,
			}

			// Set restart policy
			switch restartResp {
			case "always":
				newCommand.RestartPolicy = types.RestartAlways
			case "on-failure":
				newCommand.RestartPolicy = types.RestartOnFailure
			default:
				newCommand.RestartPolicy = types.RestartNever
			}
			
			if cancelled {
				fmt.Println("Operation cancelled")
				return nil
			}
		}

		// Add command to config
		config.Commands = append(config.Commands, newCommand)

		// Save config
		data, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal config: %v", err)
		}

		if err := ioutil.WriteFile(configPath, data, 0644); err != nil {
			return fmt.Errorf("failed to write config file: %v", err)
		}

		fmt.Printf("\nCommand added successfully:\n")
		fmt.Printf("  Session ID:  %s\n", newCommand.ID)
		fmt.Printf("  Command:     %s %s\n", newCommand.Command, strings.Join(newCommand.Args, " "))
		if newCommand.WorkingDir != "" {
			fmt.Printf("  Working Dir: %s\n", newCommand.WorkingDir)
		}
		if newCommand.Description != "" {
			fmt.Printf("  Description: %s\n", newCommand.Description)
		}
		fmt.Printf("  Auto-start:  %v\n", newCommand.AutoStart)
		fmt.Printf("  Restart:     %s\n", newCommand.RestartPolicy)
		
		fmt.Printf("\nConfiguration saved to: %s\n", configPath)
		fmt.Printf("Use 'hs %s status' to check session status\n", newCommand.ID)

		return nil
	},
}

func init() {
	configCmd.AddCommand(configAddCmd)
	
	// Add flags for non-interactive mode
	configAddCmd.Flags().StringP("id", "i", "", "Session ID")
	configAddCmd.Flags().StringP("command", "c", "", "Command to execute")
	configAddCmd.Flags().StringP("workdir", "w", "", "Working directory")
	configAddCmd.Flags().StringP("description", "d", "", "Command description")
	configAddCmd.Flags().BoolP("auto-start", "a", false, "Auto-start on boot")
	configAddCmd.Flags().StringP("restart", "r", "never", "Restart policy (never/always/on-failure)")
	configAddCmd.Flags().StringSliceP("env", "e", []string{}, "Environment variables (key=value)")
	configAddCmd.Flags().StringSliceP("tags", "t", []string{}, "Tags for categorization")
}