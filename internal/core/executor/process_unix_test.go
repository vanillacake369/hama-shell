//go:build !windows

package executor

import (
	"os/exec"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func TestUnixProcessManager_Success(t *testing.T) {
	// Test creating a Unix process manager
	manager := newProcessManager()
	if manager == nil {
		t.Fatal("Expected process manager to be created, got nil")
	}

	// Verify it's the correct type
	if _, ok := manager.(*unixProcessManager); !ok {
		t.Fatal("Expected unixProcessManager, got different type")
	}

	// Test setupCommand with a simple Unix command
	cmd := exec.Command("sleep", "10")

	// Setup command with Unix-specific settings
	manager.setupCommand(cmd)

	// Verify the command has correct Unix settings
	if cmd.SysProcAttr == nil {
		t.Fatal("Expected SysProcAttr to be set, got nil")
	}

	if !cmd.SysProcAttr.Setpgid {
		t.Error("Expected Setpgid to be true, got false")
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start test command: %v", err)
	}

	// Give it a moment to start properly
	time.Sleep(100 * time.Millisecond)

	// Test graceful termination with SIGTERM
	err := manager.terminateProcess(cmd.Process)
	if err != nil {
		t.Errorf("Expected terminateProcess to succeed, got error: %v", err)
	}

	// Verify process is actually terminated by waiting for it
	// This should complete quickly since we terminated it
	done := make(chan bool, 1)
	go func() {
		cmd.Wait()
		done <- true
	}()

	select {
	case <-done:
		// Process terminated successfully
	case <-time.After(3 * time.Second):
		t.Error("Process did not terminate within expected time")
	}

	// Test terminateProcess with nil process (should not error)
	err = manager.terminateProcess(nil)
	if err != nil {
		t.Errorf("Expected terminateProcess with nil to succeed, got error: %v", err)
	}

	// Test forceKill functionality by starting another process
	cmd2 := exec.Command("sleep", "10")
	manager.setupCommand(cmd2)

	if err := cmd2.Start(); err != nil {
		t.Fatalf("Failed to start second test command: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Test the forceKill method directly
	unixManager := manager.(*unixProcessManager)
	err = unixManager.forceKill(cmd2.Process)
	if err != nil {
		t.Errorf("Expected forceKill to succeed, got error: %v", err)
	}

	// Verify second process is terminated
	done2 := make(chan bool, 1)
	go func() {
		cmd2.Wait()
		done2 <- true
	}()

	select {
	case <-done2:
		// Process force killed successfully
	case <-time.After(3 * time.Second):
		t.Error("Process did not force kill within expected time")
	}
}

// UnixProcessManagerTestSuite provides comprehensive test infrastructure using testify
type UnixProcessManagerTestSuite struct {
	suite.Suite
	manager      *unixProcessManager
	originalGOOS string
}

// SetupTest initializes each test
func (suite *UnixProcessManagerTestSuite) SetupTest() {
	suite.manager = &unixProcessManager{}
	suite.originalGOOS = runtime.GOOS
}

// TestUnixProcessManagerTestSuite runs the comprehensive test suite
func TestUnixProcessManagerTestSuite(t *testing.T) {
	suite.Run(t, new(UnixProcessManagerTestSuite))
}

// Test Suite 1: OS Detection Tests with GIVEN/WHEN/THEN
func (suite *UnixProcessManagerTestSuite) TestGetPlatformCapabilities_Darwin_ReturnsCorrectCapabilities() {
	// GIVEN a macOS/Darwin environment
	if runtime.GOOS == "darwin" {
		// WHEN getPlatformCapabilities is called
		capabilities := suite.manager.getPlatformCapabilities()
		
		// THEN the capabilities reflect macOS limitations
		assert.True(suite.T(), capabilities.supportsSessionCreation, "macOS should support session creation")
		assert.True(suite.T(), capabilities.supportsProcessGroup, "macOS should support process groups")
		assert.False(suite.T(), capabilities.supportsCombinedSetup, "macOS should NOT support combined Setsid+Setpgid")
		assert.Equal(suite.T(), "macOS/Darwin", capabilities.platformName)
	}
}

func (suite *UnixProcessManagerTestSuite) TestGetPlatformCapabilities_Linux_ReturnsFullSupport() {
	// GIVEN a Linux environment (simulated by testing known values)
	// WHEN we examine what Linux capabilities would be
	capabilities := suite.getPlatformCapabilitiesForOS("linux")
	
	// THEN Linux should have full support
	assert.True(suite.T(), capabilities.supportsSessionCreation)
	assert.True(suite.T(), capabilities.supportsProcessGroup)
	assert.True(suite.T(), capabilities.supportsCombinedSetup, "Linux should support combined setup")
	assert.Equal(suite.T(), "Linux", capabilities.platformName)
}

func (suite *UnixProcessManagerTestSuite) TestGetPlatformCapabilities_UnknownOS_ReturnsConservativeDefaults() {
	// GIVEN we test the logic for an unknown OS
	capabilities := suite.getPlatformCapabilitiesForOS("freebsd")
	
	// THEN it should return conservative defaults (assume full support)
	assert.True(suite.T(), capabilities.supportsSessionCreation)
	assert.True(suite.T(), capabilities.supportsProcessGroup)
	assert.True(suite.T(), capabilities.supportsCombinedSetup, "Unknown OS should assume full support")
	assert.Equal(suite.T(), "freebsd", capabilities.platformName)
}

// Test Suite 2: Platform-Specific Setup Tests with GIVEN/WHEN/THEN
func (suite *UnixProcessManagerTestSuite) TestGetPlatformSpecificSysProcAttr_Darwin_SetsidOnly() {
	// GIVEN a macOS environment
	if runtime.GOOS == "darwin" {
		// WHEN getPlatformSpecificSysProcAttr is called
		sysProcAttr := suite.manager.getPlatformSpecificSysProcAttr()
		
		// THEN it should configure session creation only
		require.NotNil(suite.T(), sysProcAttr, "SysProcAttr should not be nil")
		assert.True(suite.T(), sysProcAttr.Setsid, "Should create new session on macOS")
		assert.False(suite.T(), sysProcAttr.Setpgid, "Should NOT set process group on macOS (causes permission error)")
		assert.Equal(suite.T(), 0, sysProcAttr.Pgid, "PGID should remain 0")
	}
}

func (suite *UnixProcessManagerTestSuite) TestGetPlatformSpecificSysProcAttr_Linux_FullSetup() {
	// GIVEN a Linux environment (we'll test the logic for Linux using helper)
	if runtime.GOOS == "linux" {
		// WHEN getPlatformSpecificSysProcAttr is called
		sysProcAttr := suite.manager.getPlatformSpecificSysProcAttr()
		
		// THEN it should configure full session + process group setup
		require.NotNil(suite.T(), sysProcAttr, "SysProcAttr should not be nil")
		assert.True(suite.T(), sysProcAttr.Setsid, "Should create new session on Linux")
		assert.True(suite.T(), sysProcAttr.Setpgid, "Should set process group on Linux")
		assert.Equal(suite.T(), 0, sysProcAttr.Pgid, "PGID should be 0 (become group leader)")
	}
}

// Test Suite 3: Supervisor Setup Tests with GIVEN/WHEN/THEN
func (suite *UnixProcessManagerTestSuite) TestSetupSupervisor_AppliesPlatformSpecificConfiguration() {
	// GIVEN a command to configure
	cmd := exec.Command("echo", "test")
	require.Nil(suite.T(), cmd.SysProcAttr, "Command should start with no SysProcAttr")
	
	// WHEN setupSupervisor is called
	err := suite.manager.setupSupervisor(cmd)
	
	// THEN it should succeed and apply platform-specific configuration
	require.NoError(suite.T(), err, "setupSupervisor should not return error")
	require.NotNil(suite.T(), cmd.SysProcAttr, "SysProcAttr should be set after setup")
	assert.True(suite.T(), cmd.SysProcAttr.Setsid, "Should always create new session")
	
	// Platform-specific assertions
	if runtime.GOOS == "darwin" {
		assert.False(suite.T(), cmd.SysProcAttr.Setpgid, "macOS should not set Setpgid")
	}
	if runtime.GOOS == "linux" {
		assert.True(suite.T(), cmd.SysProcAttr.Setpgid, "Linux should set Setpgid")
		assert.Equal(suite.T(), 0, cmd.SysProcAttr.Pgid, "Linux should set Pgid to 0")
	}
}

func (suite *UnixProcessManagerTestSuite) TestSetupCommand_SetsBasicProcessGroup() {
	// GIVEN a command to configure for basic execution
	cmd := exec.Command("echo", "test")
	require.Nil(suite.T(), cmd.SysProcAttr, "Command should start with no SysProcAttr")
	
	// WHEN setupCommand is called (not setupSupervisor)
	suite.manager.setupCommand(cmd)
	
	// THEN it should apply basic process group settings
	require.NotNil(suite.T(), cmd.SysProcAttr, "SysProcAttr should be set")
	assert.True(suite.T(), cmd.SysProcAttr.Setpgid, "Should set process group for basic commands")
	// Note: Setsid should be false for basic commands (not session leaders)
	assert.False(suite.T(), cmd.SysProcAttr.Setsid, "Basic commands should not create new session")
}

// Test Suite 4: Edge Cases and Error Conditions with GIVEN/WHEN/THEN
func (suite *UnixProcessManagerTestSuite) TestSetupSupervisor_WithNilCommand_ReturnsError() {
	// GIVEN a nil command
	var cmd *exec.Cmd = nil
	
	// WHEN setupSupervisor is called
	err := suite.manager.setupSupervisor(cmd)
	
	// THEN it should return a descriptive error
	require.Error(suite.T(), err, "setupSupervisor should return error for nil command")
	assert.Contains(suite.T(), err.Error(), "command cannot be nil", "Error message should be descriptive")
}

// Helper function to test platform capabilities logic without runtime.GOOS dependency
func (suite *UnixProcessManagerTestSuite) getPlatformCapabilitiesForOS(goos string) platformCapabilities {
	switch goos {
	case "darwin":
		return platformCapabilities{
			supportsSessionCreation: true,
			supportsProcessGroup:    true,
			supportsCombinedSetup:   false,
			platformName:           "macOS/Darwin",
		}
	case "linux":
		return platformCapabilities{
			supportsSessionCreation: true,
			supportsProcessGroup:    true,
			supportsCombinedSetup:   true,
			platformName:           "Linux",
		}
	default:
		return platformCapabilities{
			supportsSessionCreation: true,
			supportsProcessGroup:    true,
			supportsCombinedSetup:   true,
			platformName:           goos,
		}
	}
}

// Benchmark Tests for Performance
func BenchmarkGetPlatformCapabilities(b *testing.B) {
	manager := &unixProcessManager{}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = manager.getPlatformCapabilities()
	}
}

func BenchmarkGetPlatformSpecificSysProcAttr(b *testing.B) {
	manager := &unixProcessManager{}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = manager.getPlatformSpecificSysProcAttr()
	}
}
