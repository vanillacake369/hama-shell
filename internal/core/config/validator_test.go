package config

import (
	"os"
	"testing"

	"github.com/spf13/viper"
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
              - "aws configure ,,,,"
              - "aws ssm ,,,,"
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

func TestValidateViper(t *testing.T) {
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

	// Load config into viper
	viper.Reset()
	viper.SetConfigFile(tmpFile.Name())
	if err := viper.ReadInConfig(); err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	// Test successful config validation
	validator := NewValidator()
	err = validator.ValidateViper()
	if err != nil {
		t.Fatalf("Validation failed: %v", err)
	}

	// Verify validator helper functions
	projects := validator.GetProjects()
	if len(projects) != 1 {
		t.Errorf("Expected 1 project, got %d", len(projects))
	}

	stages := validator.GetStages("myapp")
	if len(stages) != 2 {
		t.Errorf("Expected 2 stages, got %d", len(stages))
	}

	services := validator.GetServices("myapp", "dev")
	if len(services) != 2 {
		t.Errorf("Expected 2 services in dev stage, got %d", len(services))
	}
}

func TestValidateViper_MissingProjects(t *testing.T) {
	// Test config without projects section
	invalidConfig := `global_settings:
  timeout: 30`

	tmpFile, err := os.CreateTemp("", "invalid-config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(invalidConfig)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	viper.Reset()
	viper.SetConfigFile(tmpFile.Name())
	if err := viper.ReadInConfig(); err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	validator := NewValidator()
	err = validator.ValidateViper()
	if err == nil {
		t.Error("Expected validation error for missing projects, got nil")
	}
}

func TestValidateViper_EmptyCommands(t *testing.T) {
	// Test config with empty commands
	invalidConfig := `
projects:
  myapp:
    stages:
      dev:
        services:
          db:
            description: "Database"
            commands: []`

	tmpFile, err := os.CreateTemp("", "invalid-config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(invalidConfig)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	viper.Reset()
	viper.SetConfigFile(tmpFile.Name())
	if err := viper.ReadInConfig(); err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	validator := NewValidator()
	err = validator.ValidateViper()
	if err == nil {
		t.Error("Expected validation error for empty commands, got nil")
	}
}
