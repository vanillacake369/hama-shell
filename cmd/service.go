package cmd

import (
	"log"
	"strings"

	"github.com/spf13/cobra"

	"hama-shell/internal/core/service/api"
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

// runServiceStart starts a service using API layer
func runServiceStart(_ *cobra.Command, args []string) {
	// Parse project.service format
	parts := strings.Split(args[0], ".")
	if len(parts) != 2 {
		log.Fatalf("Invalid service format. Use: <project>.<service>")
	}

	// Create service API
	serviceAPI := api.NewServiceAPI()
	defer serviceAPI.Shutdown()

	// Start service through API layer
	if err := serviceAPI.StartService(parts[0], parts[1]); err != nil {
		log.Fatalf("Failed to start service: %v", err)
	}
}

// runServiceList lists all available services using API layer
func runServiceList(_ *cobra.Command, _ []string) {
	// Create service API
	serviceAPI := api.NewServiceAPI()
	defer serviceAPI.Shutdown()

	// List services through API layer
	if err := serviceAPI.ListServices(); err != nil {
		log.Fatalf("Failed to list services: %v", err)
	}
}
