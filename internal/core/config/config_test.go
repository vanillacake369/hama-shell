package config

import (
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
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
	// GIVEN a validator with valid YAML configuration
	validator := NewValidator()
	tmpFile := createTempConfigFile(t, testYAMLConfig)
	defer os.Remove(tmpFile)
	
	// Load config into viper
	viper.Reset()
	viper.SetConfigFile(tmpFile)
	err := viper.ReadInConfig()
	assert.NoError(t, err, "Should successfully read config file")

	// WHEN we validate the viper configuration
	validationErr := validator.ValidateViper()

	// THEN it should validate successfully
	assert.NoError(t, validationErr, "Valid configuration should pass validation")
	
	// AND helper functions should return correct information
	projects := validator.GetProjects()
	assert.Len(t, projects, 1, "Should have exactly 1 project")
	assert.Contains(t, projects, "myapp", "Should contain myapp project")

	stages := validator.GetStages("myapp")
	assert.Len(t, stages, 2, "Should have exactly 2 stages")
	assert.Contains(t, stages, "dev", "Should contain dev stage")
	assert.Contains(t, stages, "prod", "Should contain prod stage")

	services := validator.GetServices("myapp", "dev")
	assert.Len(t, services, 2, "Should have exactly 2 services in dev stage")
	assert.Contains(t, services, "db", "Should contain db service")
	assert.Contains(t, services, "api-server", "Should contain api-server service")
}

func TestValidateViper_MissingProjects(t *testing.T) {
	// GIVEN a validator and config without projects section
	validator := NewValidator()
	invalidConfig := `global_settings:
  timeout: 30`
	
	tmpFile := createTempConfigFile(t, invalidConfig)
	defer os.Remove(tmpFile)

	// Load invalid config into viper
	viper.Reset()
	viper.SetConfigFile(tmpFile)
	err := viper.ReadInConfig()
	assert.NoError(t, err, "Should successfully read config file")

	// WHEN we validate the configuration with missing projects
	validationErr := validator.ValidateViper()

	// THEN it should return a validation error
	assert.Error(t, validationErr, "Should return error for missing projects section")
	assert.Contains(t, validationErr.Error(), "missing required 'projects' section", 
		"Error should mention missing projects section")
}

func TestValidateViper_EmptyCommands(t *testing.T) {
	// GIVEN a validator and config with empty commands list
	validator := NewValidator()
	invalidConfig := `
projects:
  myapp:
    stages:
      dev:
        services:
          db:
            description: "Database"
            commands: []`

	tmpFile := createTempConfigFile(t, invalidConfig)
	defer os.Remove(tmpFile)

	// Load config with empty commands into viper
	viper.Reset()
	viper.SetConfigFile(tmpFile)
	err := viper.ReadInConfig()
	assert.NoError(t, err, "Should successfully read config file")

	// WHEN we validate the configuration with empty commands
	validationErr := validator.ValidateViper()

	// THEN it should return a validation error
	assert.Error(t, validationErr, "Should return error for empty commands list")
	assert.Contains(t, validationErr.Error(), "must have at least one command", 
		"Error should mention missing commands")
}

func TestValidateViper_InvalidYAML(t *testing.T) {
	// GIVEN malformed YAML configuration
	invalidYAML := `
projects:
  myapp:
    stages:
      dev:
        services:
          db:
            description: "Database"
            commands:
              - "ssh command"
            invalid_yaml: [unclosed`

	tmpFile := createTempConfigFile(t, invalidYAML)
	defer os.Remove(tmpFile)

	// WHEN we try to read the malformed YAML into viper
	viper.Reset()
	viper.SetConfigFile(tmpFile)
	err := viper.ReadInConfig()

	// THEN it should fail to read the config
	assert.Error(t, err, "Should fail to read malformed YAML")
}

func TestValidateViper_MissingStages(t *testing.T) {
	// GIVEN a validator and config with project but no stages
	validator := NewValidator()
	invalidConfig := `
projects:
  myapp:
    description: "App without stages"
global_settings:
  timeout: 30`

	tmpFile := createTempConfigFile(t, invalidConfig)
	defer os.Remove(tmpFile)

	viper.Reset()
	viper.SetConfigFile(tmpFile)
	err := viper.ReadInConfig()
	assert.NoError(t, err, "Should successfully read config file")

	// WHEN we validate the configuration with missing stages
	validationErr := validator.ValidateViper()

	// THEN it should return a validation error
	assert.Error(t, validationErr, "Should return error for missing stages")
	assert.Contains(t, validationErr.Error(), "missing required 'stages' section", 
		"Error should mention missing stages section")
}

func TestValidateViper_MissingServices(t *testing.T) {
	// GIVEN a validator and config with stage but no services
	validator := NewValidator()
	invalidConfig := `
projects:
  myapp:
    stages:
      dev:
        description: "Development stage without services"`

	tmpFile := createTempConfigFile(t, invalidConfig)
	defer os.Remove(tmpFile)

	viper.Reset()
	viper.SetConfigFile(tmpFile)
	err := viper.ReadInConfig()
	assert.NoError(t, err, "Should successfully read config file")

	// WHEN we validate the configuration with missing services
	validationErr := validator.ValidateViper()

	// THEN it should return a validation error
	assert.Error(t, validationErr, "Should return error for missing services")
	assert.Contains(t, validationErr.Error(), "missing required 'services' section", 
		"Error should mention missing services section")
}

func TestValidator_HelperFunctions(t *testing.T) {
	// GIVEN a validator with loaded test configuration
	validator := NewValidator()
	tmpFile := createTempConfigFile(t, testYAMLConfig)
	defer os.Remove(tmpFile)

	viper.Reset()
	viper.SetConfigFile(tmpFile)
	err := viper.ReadInConfig()
	assert.NoError(t, err, "Should successfully read config file")

	t.Run("GetStages with non-existent project", func(t *testing.T) {
		// WHEN we get stages for non-existent project
		stages := validator.GetStages("nonexistent")

		// THEN it should return empty list
		assert.Empty(t, stages, "Should return empty list for non-existent project")
	})

	t.Run("GetServices with non-existent project or stage", func(t *testing.T) {
		// WHEN we get services for non-existent project or stage
		services1 := validator.GetServices("nonexistent", "dev")
		services2 := validator.GetServices("myapp", "nonexistent")

		// THEN it should return empty lists
		assert.Empty(t, services1, "Should return empty list for non-existent project")
		assert.Empty(t, services2, "Should return empty list for non-existent stage")
	})

	t.Run("GetServices with valid project and stage", func(t *testing.T) {
		// WHEN we get services for valid project and stage
		prodServices := validator.GetServices("myapp", "prod")

		// THEN it should return correct services
		assert.Len(t, prodServices, 1, "Should have exactly 1 service in prod stage")
		assert.Contains(t, prodServices, "db", "Should contain db service")
	})
}

// Helper function to create temporary config files
func createTempConfigFile(t *testing.T, content string) string {
	tmpFile, err := os.CreateTemp("", "config-test-*.yaml")
	assert.NoError(t, err, "Should create temporary file successfully")

	_, err = tmpFile.Write([]byte(content))
	assert.NoError(t, err, "Should write content to temporary file")

	err = tmpFile.Close()
	assert.NoError(t, err, "Should close temporary file successfully")

	return tmpFile.Name()
}
