package cmd

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/creack/pty"
	"github.com/spf13/cobra"
	"golang.org/x/term"

	"hama-shell/internal/core/config"
	"hama-shell/internal/core/terminal"
)

// serviceCmd represents the service command
var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "Manage services from configuration",
	Long:  "Commands to start and manage services defined in hama-shell.yaml configuration file",
}

// serviceStartCmd represents the service start command
var serviceStartCmd = &cobra.Command{
	Use:   "start <project>.<service>",
	Short: "Start a service from configuration",
	Long: `Start a service defined in the configuration file.

Examples:
  hama-shell service start myproject.database
  hama-shell service start myproject.api`,
	Args: cobra.ExactArgs(1),
	Run:  runServiceStart,
}

// serviceListCmd represents the service list command
var serviceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available services",
	Long:  "Display all projects and services from the configuration file",
	Run:   runServiceList,
}

func init() {
	rootCmd.AddCommand(serviceCmd)
	serviceCmd.AddCommand(serviceStartCmd)
	serviceCmd.AddCommand(serviceListCmd)
}

// runServiceStart starts a service using terminal server
func runServiceStart(_ *cobra.Command, args []string) {
	// Parse project.service format
	parts := strings.Split(args[0], ".")
	if len(parts) != 2 {
		log.Fatalf("Invalid service format. Use: <project>.<service>")
	}

	projectName := parts[0]
	serviceName := parts[1]

	// Get configuration
	configManager := config.GetInstance()
	cfg := configManager.GetConfig()

	// Find project
	project, exists := cfg.Projects[projectName]
	if !exists {
		log.Fatalf("Project '%s' not found in configuration", projectName)
	}

	// Find service
	service, exists := project.Services[serviceName]
	if !exists {
		log.Fatalf("Service '%s' not found in project '%s'", serviceName, projectName)
	}

	// Validate service has commands
	if len(service.Commands) == 0 {
		log.Fatalf("Service '%s.%s' has no commands defined", projectName, serviceName)
	}

	fmt.Printf("🚀 Starting service: %s.%s\n", projectName, serviceName)
	fmt.Printf("📋 Commands to execute:\n")
	for i, cmd := range service.Commands {
		fmt.Printf("  [%d] %s\n", i+1, cmd)
	}
	fmt.Printf("\n🔗 Connecting to interactive terminal...\n\n")

	// Create terminal server
	server := terminal.NewTerminalServer()
	defer server.Shutdown()

	// Create session ID
	sessionID := fmt.Sprintf("%s-%s-%d", projectName, serviceName, time.Now().Unix())

	// Save original terminal state
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		log.Fatalf("Failed to set raw mode: %v", err)
	}
	defer func() {
		_ = term.Restore(int(os.Stdin.Fd()), oldState)
	}()

	// Set up signal handling to restore terminal on exit
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		_ = term.Restore(int(os.Stdin.Fd()), oldState)
		_ = server.KillSession(sessionID)
		os.Exit(0)
	}()

	// Create session with shell
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/bash"
	}

	session, err := server.CreateSession(sessionID, shell, []string{})
	if err != nil {
		log.Fatalf("Failed to create terminal session: %v", err)
	}

	// Get PTY master for direct I/O
	ptyMaster := session.GetPTYMaster()

	// Set terminal size on PTY
	if size, err := pty.GetsizeFull(os.Stdin); err == nil {
		if err := server.ResizeSession(sessionID, size.Rows, size.Cols); err != nil {
			fmt.Printf("Warning: failed to set PTY size: %v\n", err)
		}
	}

	// Handle window size changes
	go func() {
		sigwinch := make(chan os.Signal, 1)
		signal.Notify(sigwinch, syscall.SIGWINCH)
		for range sigwinch {
			if size, err := pty.GetsizeFull(os.Stdin); err == nil {
				_ = server.ResizeSession(sessionID, size.Rows, size.Cols)
			}
		}
	}()

	// Copy stdin to ptyMaster (user input -> shell)
	go func() {
		_, _ = io.Copy(ptyMaster, os.Stdin)
	}()

	// Copy ptyMaster to stdout (shell output -> terminal)
	go func() {
		_, _ = io.Copy(os.Stdout, ptyMaster)
	}()

	// Wait a moment for shell to be ready, then send configured commands
	go func() {
		time.Sleep(500 * time.Millisecond) // Wait for shell prompt
		
		// Execute each command from configuration
		for _, command := range service.Commands {
			commandWithNewline := command + "\n"
			if err := session.WriteInput([]byte(commandWithNewline)); err != nil {
				fmt.Printf("Warning: failed to send command '%s': %v\n", command, err)
			}
			time.Sleep(200 * time.Millisecond) // Small delay between commands
		}
	}()

	// Wait for session to finish
	for session.IsRunning() {
		time.Sleep(100 * time.Millisecond)
	}

	// Restore terminal state before final output
	_ = term.Restore(int(os.Stdin.Fd()), oldState)
	fmt.Printf("\n✅ Session ended normally\n")
	
}

// runServiceList lists all available services
func runServiceList(_ *cobra.Command, _ []string) {
	configManager := config.GetInstance()
	cfg := configManager.GetConfig()

	if len(cfg.Projects) == 0 {
		fmt.Println("No projects found in configuration.")
		fmt.Printf("Configuration file: %s\n", configManager.GetFilePath())
		return
	}

	fmt.Println("Available services:")
	fmt.Println()

	for projectName, project := range cfg.Projects {
		fmt.Printf("📁 Project: %s\n", projectName)
		
		if len(project.Services) == 0 {
			fmt.Println("  (no services defined)")
		} else {
			for serviceName, service := range project.Services {
				fmt.Printf("  🔧 %s.%s\n", projectName, serviceName)
				for i, command := range service.Commands {
					fmt.Printf("    [%d] %s\n", i+1, command)
				}
			}
		}
		fmt.Println()
	}
	
	fmt.Println("Usage:")
	fmt.Println("  hama-shell service start <project>.<service>")
	fmt.Printf("\nConfiguration file: %s\n", configManager.GetFilePath())
}
