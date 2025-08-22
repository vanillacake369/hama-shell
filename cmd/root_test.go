package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitConfigWithExampleYaml(t *testing.T) {
	// GIVEN: A valid example.yaml config file is copied to the working directory
	viper.Reset()
	tempDir := t.TempDir()
	// Read the actual example.yaml content
	examplePath := "../example.yaml"
	exampleContent, err := os.ReadFile(examplePath)
	require.NoError(t, err, "Should be able to read example.yaml")
	// Write it as hama-shell.yaml in temp directory
	configPath := filepath.Join(tempDir, "hama-shell.yaml")
	err = os.WriteFile(configPath, exampleContent, 0644)
	require.NoError(t, err)
	// Change to temp directory so config file is discovered
	originalDir, _ := os.Getwd()
	chDirErr := os.Chdir(tempDir)
	if chDirErr != nil {
		return
	}
	t.Cleanup(func() {
		err := os.Chdir(originalDir)
		if err != nil {
			return
		}
	})

	// WHEN: initConfig() is called to load the configuration
	initConfig()

	// THEN: viper should successfully parse the YAML structure and make values accessible
	// Config file should be loaded
	assert.NotEmpty(t, viper.ConfigFileUsed(), "Config file should be loaded")
	// Project structure should be parsed correctly
	assert.Equal(t, "Main application project",
		viper.GetString("projects.myapp.description"))
	assert.Equal(t, "Development environment",
		viper.GetString("projects.myapp.stages.dev.description"))
	assert.Equal(t, "PostgreSQL database connection",
		viper.GetString("projects.myapp.stages.dev.services.db.description"))
	// Array commands should be parsed correctly
	commands := viper.GetStringSlice("projects.myapp.stages.dev.services.db.commands")
	assert.Len(t, commands, 3, "Should have 3 commands")
	assert.Equal(t, "ssh -L 3306:${TARGET_HOST}:3306 ubuntu@${BASTION_HOST} -N", commands[0])
	// Global settings should be parsed with correct types
	assert.Equal(t, 30, viper.GetInt("global_settings.timeout"))
	assert.Equal(t, 3, viper.GetInt("global_settings.retries"))
	assert.True(t, viper.GetBool("global_settings.auto_restart"))
	// Other services should be accessible
	assert.Equal(t, "Application server",
		viper.GetString("projects.myapp.stages.dev.services.server.description"))
	assert.Equal(t, "CI/CD Jenkins server",
		viper.GetString("projects.myapp.stages.dev.services.jenkins.description"))
}

func TestInitConfigNoFile(t *testing.T) {
	t.Skip("Config file fallback behavior not yet implemented - initConfig exits when no config file found")
	// TODO: Implement graceful config file fallback behavior
	// GIVEN: No config file exists, only environment variables are set
	viper.Reset()
	// Create empty temp directory (no config file)
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	err := os.Chdir(tempDir)
	require.NoError(t, err)
	t.Cleanup(func() {
		err := os.Chdir(originalDir)
		require.NoError(t, err)
	})
	// Set environment variables that should be picked up by viper
	t.Setenv("HAMA_SHELL_TEST", "test_value")
	t.Setenv("GLOBAL_SETTINGS_TIMEOUT", "120")

	// WHEN: initConfig() is called without any config file present
	initConfig()

	// THEN: viper should fallback to environment variables only
	// No config file should be loaded
	assert.Empty(t, viper.ConfigFileUsed(), "No config file should be loaded")
	// Environment variables should be accessible through viper
	assert.Equal(t, "test_value", viper.Get("hama_shell_test"))
	assert.Equal(t, "120", viper.Get("global_settings_timeout"))
}
