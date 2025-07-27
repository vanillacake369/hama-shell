package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

const testYAMLConfig = `
projects:
  myapp:
    description: "Main application project"
    stages:
      dev:
        services:
          db:
            description: "Develop database"
            commands:
              - "ssh -i ${SSH_KEY_PATH} ${BASTION_USER}@bastion.dev.com"
              - "${DEV_DB_PW}"
          api-server:
            description: "Develop database"
            commands:
              - "aws configure ,,,"
              - "aws ssm ,,,"
      prod:
        services:
          db:
            description: "Production database"
            commands:
              - "ssh -i ${SSH_KEY_PATH} ${BASTION_USER}@bastion.prod.com"
              - "${PROD_DB_PW}"
global_settings:
  timeout: 30
  retries: 3
  auto_restart: true
`

func TestLoad(t *testing.T) {
	// Create temporary file
	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write test YAML to file
	if _, err := tmpFile.Write([]byte(testYAMLConfig)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	// Test successful config loading
	loader := NewLoader(tmpFile.Name())
	config, err := loader.Load(tmpFile.Name())
	if err != nil {
		t.Fatalf("GetConfig failed: %v", err)
	}

	// Print parsed YAML result for debugging
	fmt.Printf("=== PARSED CONFIG DEBUG OUTPUT ===\n")
	fmt.Printf("Global Settings: Timeout=%d, Retries=%d, AutoRestart=%t\n",
		config.GlobalSettings.Timeout, config.GlobalSettings.Retries, config.GlobalSettings.AutoRestart)
	fmt.Printf("Number of Projects: %d\n", len(config.Projects))

	for projectName, project := range config.Projects {
		fmt.Printf("Project '%s': Description='%s', Stages=%d\n",
			projectName, project.Description, len(project.Stages))

		for stageName, stage := range project.Stages {
			fmt.Printf("  Stage '%s': Description='%s', Services=%d\n",
				stageName, stage.Description, len(stage.Services))

			for serviceName, service := range stage.Services {
				fmt.Printf("    Service '%s': Description='%s', Commands=%d\n",
					serviceName, service.Description, len(service.Commands))
				for i, cmd := range service.Commands {
					fmt.Printf("      Command[%d]: %s\n", i, cmd)
				}
			}
		}
	}
	fmt.Printf("=== END DEBUG OUTPUT ===\n")

	// Verify global settings
	if config.GlobalSettings.Timeout != 30 {
		t.Errorf("Expected timeout 30, got %d", config.GlobalSettings.Timeout)
	}
	if config.GlobalSettings.Retries != 3 {
		t.Errorf("Expected retries 3, got %d", config.GlobalSettings.Retries)
	}
	if !config.GlobalSettings.AutoRestart {
		t.Error("Expected auto_restart to be true")
	}

	// Verify project structure
	if len(config.Projects) != 1 {
		t.Errorf("Expected 1 project, got %d", len(config.Projects))
	}

	myapp, exists := config.Projects["myapp"]
	if !exists {
		t.Fatal("Project 'myapp' not found")
	}
	if myapp.Description != "Main application project" {
		t.Errorf("Expected project description 'Main application project', got '%s'", myapp.Description)
	}

	// Verify stages
	if len(myapp.Stages) != 2 {
		t.Errorf("Expected 2 stages, got %d", len(myapp.Stages))
	}

	// Test dev stage
	dev, exists := myapp.Stages["dev"]
	if !exists {
		t.Fatal("Stage 'dev' not found")
	}
	if len(dev.Services) != 2 {
		t.Errorf("Expected 2 services in dev stage, got %d", len(dev.Services))
	}

	// Test dev db service
	devDB, exists := dev.Services["db"]
	if !exists {
		t.Fatal("Service 'db' not found in dev stage")
	}
	if devDB.Description != "Develop database" {
		t.Errorf("Expected dev db description 'Develop database', got '%s'", devDB.Description)
	}
	if len(devDB.Commands) != 2 {
		t.Errorf("Expected 2 commands in dev db service, got %d", len(devDB.Commands))
	}
	expectedDevDBCmd1 := "ssh -i ${SSH_KEY_PATH} ${BASTION_USER}@bastion.dev.com"
	if devDB.Commands[0] != expectedDevDBCmd1 {
		t.Errorf("Expected first dev db command '%s', got '%s'", expectedDevDBCmd1, devDB.Commands[0])
	}

	// Test dev api-server service
	devAPI, exists := dev.Services["api-server"]
	if !exists {
		t.Fatal("Service 'api-server' not found in dev stage")
	}
	if devAPI.Description != "Develop database" {
		t.Errorf("Expected dev api-server description 'Develop database', got '%s'", devAPI.Description)
	}
	if len(devAPI.Commands) != 2 {
		t.Errorf("Expected 2 commands in dev api-server service, got %d", len(devAPI.Commands))
	}

	// Test prod stage
	prod, exists := myapp.Stages["prod"]
	if !exists {
		t.Fatal("Stage 'prod' not found")
	}
	if len(prod.Services) != 1 {
		t.Errorf("Expected 1 service in prod stage, got %d", len(prod.Services))
	}

	// Test prod db service
	prodDB, exists := prod.Services["db"]
	if !exists {
		t.Fatal("Service 'db' not found in prod stage")
	}
	if prodDB.Description != "Production database" {
		t.Errorf("Expected prod db description 'Production database', got '%s'", prodDB.Description)
	}
	if len(prodDB.Commands) != 2 {
		t.Errorf("Expected 2 commands in prod db service, got %d", len(prodDB.Commands))
	}
	expectedProdDBCmd1 := "ssh -i ${SSH_KEY_PATH} ${BASTION_USER}@bastion.prod.com"
	if prodDB.Commands[0] != expectedProdDBCmd1 {
		t.Errorf("Expected first prod db command '%s', got '%s'", expectedProdDBCmd1, prodDB.Commands[0])
	}
}

func TestGetConfig_FileNotFound(t *testing.T) {
	path := "nonexistent-file.yaml"
	loader := NewLoader(path)
	_, err := loader.Load(path)
	if err == nil {
		t.Error("Expected error for nonexistent file, got nil")
	}
}

func TestGetConfig_InvalidYAML(t *testing.T) {
	// Create temporary file with invalid YAML
	tmpFile, err := os.CreateTemp("", "invalid-config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	invalidYAML := `projects:
  myapp:
    description: "Test"
    stages:
      dev:
        services:
          db:
            invalid_yaml_syntax: [unclosed bracket
`
	if _, err := tmpFile.Write([]byte(invalidYAML)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	loader := NewLoader(tmpFile.Name())
	_, err = loader.Load(tmpFile.Name())
	if err == nil {
		t.Error("Expected error for invalid YAML, got nil")
	}
}

func TestGetConfig_RelativePath(t *testing.T) {
	// Create temporary file in current directory
	tmpFile, err := os.CreateTemp(".", "config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(testYAMLConfig)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	// Test with relative path
	relativePath := filepath.Base(tmpFile.Name())
	loader := NewLoader(relativePath)
	config, err := loader.Load(relativePath)
	if err != nil {
		t.Fatalf("GetConfig with relative path failed: %v", err)
	}

	if config == nil {
		t.Error("Expected config to be non-nil")
	}
}

func TestConfig_EmptyProjects(t *testing.T) {
	emptyProjectsYAML := `projects: {}
global_settings:
  timeout: 10
  retries: 1
  auto_restart: false
`
	tmpFile, err := os.CreateTemp("", "empty-config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(emptyProjectsYAML)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	loader := NewLoader(tmpFile.Name())
	config, err := loader.Load(tmpFile.Name())
	if err != nil {
		t.Fatalf("GetConfig failed: %v", err)
	}

	if len(config.Projects) != 0 {
		t.Errorf("Expected 0 projects, got %d", len(config.Projects))
	}

	if config.GlobalSettings.Timeout != 10 {
		t.Errorf("Expected timeout 10, got %d", config.GlobalSettings.Timeout)
	}
}
