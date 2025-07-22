package config

import (
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

// ToDo : 테스트케이스부터 다시 짜기
// ToDo : Assertions 라이브러리 활용해서 THEN 절 개선하기
func TestGetConfig(t *testing.T) {
	// Create a temporary YAML file for testing
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test_config.yaml")

	yamlContent := `
projects:
  - name: myapp
    stages:
      - name: dev
        developers:
          - name: alice
            sessions:
              - name: frontend
                description: Run frontend development server
                steps:
                  - command: npm start
                parallel: false
      - name: prod
        developers:
          - name: bob
            sessions:
              - name: backend
                description: Deploy backend service
                steps:
                  - command: docker deploy
                parallel: true

aliases:
  myapp-prod: myapp.prod.bob.backend

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
	assert.Equal(t, "myapp", config.Projects[0].Name, "Project name should be 'myapp'")
	assert.Len(t, config.Projects[0].Stages, 2, "Expected exactly 2 stages")

	// Verify global settings
	assert.Equal(t, 3, config.GlobalSettings.Retries, "Retries should be 3")
	assert.Equal(t, 30, config.GlobalSettings.Timeout, "Timeout should be 30")
	assert.True(t, config.GlobalSettings.AutoRestart, "AutoRestart should be true")

	// Verify aliases
	assert.Equal(t, "myapp.prod.bob.backend", config.Aliases.MyAppProd, "Alias should match expected value")

	// Verify stage details
	devStage := config.Projects[0].Stages[0]
	prodStage := config.Projects[0].Stages[1]

	assert.Equal(t, "dev", devStage.Name, "First stage should be 'dev'")
	assert.Equal(t, "prod", prodStage.Name, "Second stage should be 'prod'")

	// Verify developers and sessions
	assert.Len(t, devStage.Developers, 1, "Dev stage should have 1 developer")
	assert.Equal(t, "alice", devStage.Developers[0].Name, "Dev developer should be 'alice'")
	assert.Len(t, devStage.Developers[0].Sessions, 1, "Alice should have 1 session")

	aliceSession := devStage.Developers[0].Sessions[0]
	assert.Equal(t, "frontend", aliceSession.Name, "Session name should be 'frontend'")
	assert.Equal(t, "Run frontend development server", aliceSession.Description, "Session description should match")
	assert.False(t, aliceSession.Parallel, "Frontend session should not be parallel")
	assert.Len(t, aliceSession.Steps, 1, "Frontend session should have 1 step")
	assert.Equal(t, "npm start", aliceSession.Steps[0].Command, "Command should be 'npm start'")

	// Verify prod stage
	assert.Len(t, prodStage.Developers, 1, "Prod stage should have 1 developer")
	assert.Equal(t, "bob", prodStage.Developers[0].Name, "Prod developer should be 'bob'")

	bobSession := prodStage.Developers[0].Sessions[0]
	assert.Equal(t, "backend", bobSession.Name, "Session name should be 'backend'")
	assert.True(t, bobSession.Parallel, "Backend session should be parallel")
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
  - name: myapp
    stages:
      - name: dev
        developers:
          - name: alice
            sessions:
              - name: frontend
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
