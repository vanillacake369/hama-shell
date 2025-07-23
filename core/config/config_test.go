package config

import (
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func TestGetConfig(t *testing.T) {
	// Create a temporary YAML file for testing
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test_config.yaml")

	yamlContent := `
projects:
  myapp:
    description: "Main application project"
    stages:
      dev:
        description: "Development environment"
        services:
          db:
            description: "PostgreSQL database connection"
            command:
              - "ssh -L 5432:dev-db.myapp.com:5432 dbuser@dev-db.myapp.com -i /path/to/ssh/key"
          server:
            description: "Application server"
            command:
              - "ssh appuser@dev-app.myapp.com -i /path/to/ssh/key"
      prod:
        description: "Production environment"
        services:
          db:
            description: "Production database"
            steps:
              - command: 
                  - "ssh -i /path/to/key user@bastion.prod.com"
              - command: 
                  - "ssh -L 5432:prod-db:5432 db-reader@prod-db-proxy"

aliases:
  myapp-prod-db: "myapp.prod.db"

global_settings:
  retries: 3
  timeout: 30
  auto_restart: true
`

	// Write test YAML content to file
	err := os.WriteFile(configFile, []byte(yamlContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Test GetConfig function
	config, err := GetConfig(configFile)
	if err != nil {
		t.Fatalf("GetConfig failed: %v", err)
	}

	// Verify parsed config using testify assertions
	assert.Len(t, config.Projects, 1, "Expected exactly 1 project")
	myappProject, exists := config.Projects["myapp"]
	assert.True(t, exists, "Project 'myapp' should exist")
	assert.Equal(t, "Main application project", myappProject.Description, "Project description should match")
	assert.Len(t, myappProject.Stages, 2, "Expected exactly 2 stages")

	// Verify global settings
	assert.Equal(t, 3, config.GlobalSettings.Retries, "Retries should be 3")
	assert.Equal(t, 30, config.GlobalSettings.Timeout, "Timeout should be 30")
	assert.True(t, config.GlobalSettings.AutoRestart, "AutoRestart should be true")

	// Verify aliases
	assert.Equal(t, "myapp.prod.db", config.Aliases["myapp-prod-db"], "Alias should match expected value")

	// Verify stage details
	devStage, devExists := myappProject.Stages["dev"]
	prodStage, prodExists := myappProject.Stages["prod"]
	assert.True(t, devExists, "Dev stage should exist")
	assert.True(t, prodExists, "Prod stage should exist")
	assert.Equal(t, "Development environment", devStage.Description, "Dev stage description should match")
	assert.Equal(t, "Production environment", prodStage.Description, "Prod stage description should match")

	// Verify dev services
	assert.Len(t, devStage.Services, 2, "Dev stage should have 2 services")
	dbService, dbExists := devStage.Services["db"]
	serverService, serverExists := devStage.Services["server"]
	assert.True(t, dbExists, "DB service should exist")
	assert.True(t, serverExists, "Server service should exist")

	// Verify DB service details
	assert.Equal(t, "PostgreSQL database connection", dbService.Description, "DB service description should match")
	assert.Len(t, dbService.Command, 1, "DB service should have 1 command")
	assert.Equal(t, "ssh -L 5432:dev-db.myapp.com:5432 dbuser@dev-db.myapp.com -i /path/to/ssh/key", dbService.Command[0], "DB command should match")

	// Verify server service details
	assert.Equal(t, "Application server", serverService.Description, "Server service description should match")
	assert.Len(t, serverService.Command, 1, "Server service should have 1 command")
	assert.Equal(t, "ssh appuser@dev-app.myapp.com -i /path/to/ssh/key", serverService.Command[0], "Server command should match")

	// Verify prod stage with steps
	prodDbService, prodDbExists := prodStage.Services["db"]
	assert.True(t, prodDbExists, "Prod DB service should exist")
	assert.Equal(t, "Production database", prodDbService.Description, "Prod DB description should match")
	assert.Len(t, prodDbService.Steps, 2, "Prod DB should have 2 steps")
	assert.Len(t, prodDbService.Steps[0].Commands, 1, "First step should have 1 command")
	assert.Equal(t, "ssh -i /path/to/key user@bastion.prod.com", prodDbService.Steps[0].Commands[0], "First step command should match")
	assert.Len(t, prodDbService.Steps[1].Commands, 1, "Second step should have 1 command")
	assert.Equal(t, "ssh -L 5432:prod-db:5432 db-reader@prod-db-proxy", prodDbService.Steps[1].Commands[0], "Second step command should match")
}

func TestGetConfigNonexistentFile(t *testing.T) {
	_, err := GetConfig("/nonexistent/file.yaml")
	assert.Error(t, err, "Should return error for nonexistent file")
}

func TestGetConfigInvalidYAML(t *testing.T) {
	// Create a temporary file with invalid YAML
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "invalid_config.yaml")

	invalidYAML := `
projects:
  myapp:
    description: "Main application project"
    stages:
      dev:
        description: "Development environment"
        services:
          db:
            description: "unclosed quote
`

	err := os.WriteFile(configFile, []byte(invalidYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to create invalid config file: %v", err)
	}

	_, err = GetConfig(configFile)
	assert.Error(t, err, "Should return error for invalid YAML")
	assert.Contains(t, err.Error(), "yaml", "Error should mention YAML parsing issue")
}

func TestMultipleCommandsInStep(t *testing.T) {
	// Create a temporary YAML file with multiple commands in a step
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "multi_command_config.yaml")

	yamlContent := `
projects:
  testapp:
    description: "Test application with multiple commands per step"
    stages:
      dev:
        description: "Development environment"
        services:
          complex:
            description: "Complex service with multiple commands per step"
            steps:
              - command:
                  - "echo 'Starting setup...'"
                  - "mkdir -p /tmp/testdir"
                  - "cd /tmp/testdir"
              - command:
                  - "ssh -i /path/to/key user@bastion.com"
                  - "export ENV_VAR=value"
                  - "ssh -L 5432:db:5432 user@internal-db"

aliases:
  test-complex: "testapp.dev.complex"

global_settings:
  retries: 2
  timeout: 60
  auto_restart: false
`

	// Write test YAML content to file
	err := os.WriteFile(configFile, []byte(yamlContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Test GetConfig function
	config, err := GetConfig(configFile)
	if err != nil {
		t.Fatalf("GetConfig failed: %v", err)
	}

	// Verify parsed config
	assert.Len(t, config.Projects, 1, "Expected exactly 1 project")
	testappProject, exists := config.Projects["testapp"]
	assert.True(t, exists, "Project 'testapp' should exist")
	assert.Equal(t, "Test application with multiple commands per step", testappProject.Description, "Project description should match")

	// Verify dev stage
	devStage, devExists := testappProject.Stages["dev"]
	assert.True(t, devExists, "Dev stage should exist")
	assert.Equal(t, "Development environment", devStage.Description, "Dev stage description should match")

	// Verify complex service
	complexService, complexExists := devStage.Services["complex"]
	assert.True(t, complexExists, "Complex service should exist")
	assert.Equal(t, "Complex service with multiple commands per step", complexService.Description, "Complex service description should match")
	assert.Len(t, complexService.Steps, 2, "Complex service should have 2 steps")

	// Verify first step with multiple commands
	firstStep := complexService.Steps[0]
	assert.Len(t, firstStep.Commands, 3, "First step should have 3 commands")
	assert.Equal(t, "echo 'Starting setup...'", firstStep.Commands[0], "First command should match")
	assert.Equal(t, "mkdir -p /tmp/testdir", firstStep.Commands[1], "Second command should match")
	assert.Equal(t, "cd /tmp/testdir", firstStep.Commands[2], "Third command should match")

	// Verify second step with multiple commands
	secondStep := complexService.Steps[1]
	assert.Len(t, secondStep.Commands, 3, "Second step should have 3 commands")
	assert.Equal(t, "ssh -i /path/to/key user@bastion.com", secondStep.Commands[0], "First command of second step should match")
	assert.Equal(t, "export ENV_VAR=value", secondStep.Commands[1], "Second command of second step should match")
	assert.Equal(t, "ssh -L 5432:db:5432 user@internal-db", secondStep.Commands[2], "Third command of second step should match")

	// Verify alias
	assert.Equal(t, "testapp.dev.complex", config.Aliases["test-complex"], "Alias should match expected value")

	// Verify global settings
	assert.Equal(t, 2, config.GlobalSettings.Retries, "Retries should be 2")
	assert.Equal(t, 60, config.GlobalSettings.Timeout, "Timeout should be 60")
	assert.False(t, config.GlobalSettings.AutoRestart, "AutoRestart should be false")
}

func TestExampleYAML(t *testing.T) {
	// Test that the example.yaml file can be parsed correctly
	config, err := GetConfig("../../example.yaml")
	if err != nil {
		t.Fatalf("Failed to parse example.yaml: %v", err)
	}

	// Verify we have the expected projects
	assert.Len(t, config.Projects, 2, "Expected exactly 2 projects in example.yaml")

	// Verify myapp project
	myappProject, exists := config.Projects["myapp"]
	assert.True(t, exists, "Project 'myapp' should exist")
	assert.Equal(t, "Main application project", myappProject.Description, "myapp description should match")
	assert.Len(t, myappProject.Stages, 3, "myapp should have 3 stages")

	// Verify myapp dev stage
	devStage, devExists := myappProject.Stages["dev"]
	assert.True(t, devExists, "myapp dev stage should exist")
	assert.Equal(t, "Development environment", devStage.Description, "dev stage description should match")
	assert.Len(t, devStage.Services, 4, "dev stage should have 4 services")

	// Verify specific services
	dbService, dbExists := devStage.Services["db"]
	assert.True(t, dbExists, "db service should exist")
	assert.Equal(t, "PostgreSQL database connection", dbService.Description, "db service description should match")
	assert.Len(t, dbService.Command, 1, "db service should have 1 command")
	assert.Equal(t, "ssh -L 3306:${TARGET_HOST}:3306 ubuntu@${BASTION_HOST} -N", dbService.Command[0], "db command should match")

	jenkinsService, jenkinsExists := devStage.Services["jenkins"]
	assert.True(t, jenkinsExists, "jenkins service should exist")
	assert.Equal(t, "CI/CD Jenkins server", jenkinsService.Description, "jenkins description should match")
	assert.Len(t, jenkinsService.Command, 1, "jenkins service should have 1 command")
	assert.Equal(t, "ssh -L 3306:${TARGET_HOST}:3306 ubuntu@${BASTION_HOST} -N", jenkinsService.Command[0], "jenkins command should match")

	// Verify ecommerce project
	ecomProject, ecomExists := config.Projects["ecommerce"]
	assert.True(t, ecomExists, "Project 'ecommerce' should exist")
	assert.Equal(t, "E-commerce platform project", ecomProject.Description, "ecommerce description should match")
	assert.Len(t, ecomProject.Stages, 2, "ecommerce should have 2 stages")

	// Verify production services with steps
	prodStage, prodExists := myappProject.Stages["prod"]
	assert.True(t, prodExists, "prod stage should exist")
	prodDbService, prodDbExists := prodStage.Services["db"]
	assert.True(t, prodDbExists, "prod db service should exist")
	assert.Equal(t, "Production database via bastion", prodDbService.Description, "prod db description should match")
	assert.Len(t, prodDbService.Steps, 2, "prod db should have 2 steps")
	assert.Len(t, prodDbService.Steps[0].Commands, 1, "first step should have 1 command")
	assert.Equal(t, "ssh -i ${SSH_KEY_PATH} ${BASTION_USER}@bastion.prod.com", prodDbService.Steps[0].Commands[0], "first step should match")
	assert.Len(t, prodDbService.Steps[1].Commands, 1, "second step should have 1 command")
	assert.Equal(t, "ssh -L 5432:prod-db:5432 db-reader@prod-db-proxy", prodDbService.Steps[1].Commands[0], "second step should match")

	// Verify aliases
	assert.Len(t, config.Aliases, 4, "Should have 4 aliases")
	assert.Equal(t, "myapp.dev.db", config.Aliases["myapp-dev-db"], "myapp-dev-db alias should match")
	assert.Equal(t, "myapp.prod.db", config.Aliases["myapp-prod-db"], "myapp-prod-db alias should match")
	assert.Equal(t, "ecommerce.dev.api", config.Aliases["ecom-dev-api"], "ecom-dev-api alias should match")
	assert.Equal(t, "ecommerce.dev.mongodb", config.Aliases["ecom-mongo"], "ecom-mongo alias should match")

	// Verify global settings
	assert.Equal(t, 30, config.GlobalSettings.Timeout, "Timeout should be 30")
	assert.Equal(t, 3, config.GlobalSettings.Retries, "Retries should be 3")
	assert.True(t, config.GlobalSettings.AutoRestart, "AutoRestart should be true")
}
