package config

import (
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

	// Verify parsed config
	if len(config.Projects) != 1 {
		t.Errorf("Expected 1 project, got %d", len(config.Projects))
	}

	if config.Projects[0].Name != "myapp" {
		t.Errorf("Expected project name 'myapp', got '%s'", config.Projects[0].Name)
	}

	if len(config.Projects[0].Stages) != 2 {
		t.Errorf("Expected 2 stages, got %d", len(config.Projects[0].Stages))
	}

	if config.GlobalSettings.Retries != 3 {
		t.Errorf("Expected retries 3, got %d", config.GlobalSettings.Retries)
	}

	if config.GlobalSettings.Timeout != 30 {
		t.Errorf("Expected timeout 30, got %d", config.GlobalSettings.Timeout)
	}

	if !config.GlobalSettings.AutoRestart {
		t.Errorf("Expected auto_restart true, got %v", config.GlobalSettings.AutoRestart)
	}

	if config.Aliases.MyAppProd != "myapp.prod.bob.backend" {
		t.Errorf("Expected alias 'myapp.prod.bob.backend', got '%s'", config.Aliases.MyAppProd)
	}
}

func TestGetConfigNonexistentFile(t *testing.T) {
	_, err := GetConfig("/nonexistent/file.yaml")
	if err == nil {
		t.Error("Expected error for nonexistent file, got nil")
	}
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
	if err == nil {
		t.Error("Expected error for invalid YAML, got nil")
	}
}
