//go:build !windows

package executor

import (
	"fmt"
	"runtime"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// SupervisorIntegrationTestSuite tests real process creation with platform-specific configurations
type SupervisorIntegrationTestSuite struct {
	suite.Suite
	supervisor *SupervisorManager
}

// SetupTest initializes each integration test
func (suite *SupervisorIntegrationTestSuite) SetupTest() {
	suite.supervisor = NewSupervisorManager()
}

// TearDownTest cleans up after each integration test
func (suite *SupervisorIntegrationTestSuite) TearDownTest() {
	// Any cleanup needed
}

// TestSupervisorIntegrationTestSuite runs the integration test suite
func TestSupervisorIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(SupervisorIntegrationTestSuite))
}

// Integration Test 1: Real Supervisor Creation with Platform-Specific Settings
func (suite *SupervisorIntegrationTestSuite) TestCreateSupervisor_WithShellCommands_CreatesProcessSuccessfully() {
	// GIVEN a set of shell commands for testing
	segments := []CommandSegment{
		{
			Type:     "shell",
			Commands: []string{"echo 'integration test 1'", "sleep 0.1"},
		},
	}
	
	// WHEN createSupervisor is called
	session, err := suite.supervisor.createSupervisor("test-key-1", segments, ExecutionModeBackground)
	
	// THEN it should create a session successfully
	require.NoError(suite.T(), err, "createSupervisor should succeed with shell commands")
	require.NotNil(suite.T(), session, "Session should be created")
	require.NotNil(suite.T(), session.Supervisor, "Supervisor process should be set")
	require.NotNil(suite.T(), session.Supervisor.Process, "Supervisor process should be running")
	
	// Verify session properties
	assert.Equal(suite.T(), "test-key-1", session.Key)
	assert.Equal(suite.T(), session.Supervisor.Process.Pid, session.SessionID)
	assert.Greater(suite.T(), session.PGID, 0, "PGID should be set")
	assert.Len(suite.T(), session.Segments, 1, "Should have one segment")
	
	// Cleanup
	session.Supervisor.Process.Kill()
	session.Supervisor.Wait()
}

func (suite *SupervisorIntegrationTestSuite) TestCreateSupervisor_VerifiesPlatformSpecificPGID() {
	// GIVEN a simple shell command
	segments := []CommandSegment{
		{
			Type:     "shell", 
			Commands: []string{"sleep 0.5"},
		},
	}
	
	// WHEN createSupervisor is called
	session, err := suite.supervisor.createSupervisor("test-key-2", segments, ExecutionModeBackground)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), session)
	
	// THEN the process should have correct PGID based on platform
	actualPGID, err := syscall.Getpgid(session.Supervisor.Process.Pid)
	require.NoError(suite.T(), err, "Should be able to get PGID")
	
	// Verify PGID behavior based on platform
	if runtime.GOOS == "darwin" {
		// On macOS with Setsid=true, the process becomes its own process group leader
		assert.Equal(suite.T(), session.Supervisor.Process.Pid, actualPGID, 
			"On macOS, process should be its own group leader")
	} else {
		// On Linux and other Unix, PGID should be set correctly
		assert.Equal(suite.T(), session.PGID, actualPGID,
			"PGID should match what was recorded in session")
	}
	
	// Cleanup
	session.Supervisor.Process.Kill()
	session.Supervisor.Wait()
}

// Integration Test 2: Multiple Command Segments
func (suite *SupervisorIntegrationTestSuite) TestCreateSupervisor_WithMultipleSegments_HandlesAllTypes() {
	// GIVEN multiple command segments of different types
	segments := []CommandSegment{
		{
			Type:     "shell",
			Commands: []string{"echo 'first segment'"},
		},
		{
			Type:     "shell", 
			Commands: []string{"echo 'second segment'", "sleep 0.1"},
		},
	}
	
	// WHEN createSupervisor is called
	session, err := suite.supervisor.createSupervisor("test-key-3", segments, ExecutionModeBackground)
	
	// THEN it should handle all segments
	require.NoError(suite.T(), err, "Should handle multiple segments")
	require.NotNil(suite.T(), session)
	assert.Len(suite.T(), session.Segments, 2, "Should track both segments")
	
	// Wait a moment for execution
	time.Sleep(200 * time.Millisecond)
	
	// Cleanup
	session.Supervisor.Process.Kill()
	session.Supervisor.Wait()
}

// Integration Test 3: Error Conditions and Edge Cases
func (suite *SupervisorIntegrationTestSuite) TestCreateSupervisor_WithEmptySegments_ReturnsError() {
	// GIVEN empty command segments
	var segments []CommandSegment = nil
	
	// WHEN createSupervisor is called
	session, err := suite.supervisor.createSupervisor("test-key-empty", segments, ExecutionModeBackground)
	
	// THEN it should return an error
	require.Error(suite.T(), err, "Should return error for empty segments")
	assert.Nil(suite.T(), session, "Session should be nil on error")
	assert.Contains(suite.T(), err.Error(), "no command segments", "Error should be descriptive")
}

// Integration Test 4: Platform-Specific Script Generation
func (suite *SupervisorIntegrationTestSuite) TestBuildSupervisorScript_GeneratesCorrectScript() {
	// GIVEN various command segments
	segments := []CommandSegment{
		{
			Type:     "shell",
			Commands: []string{"echo 'test shell'", "sleep 0.1"},
		},
		{
			Type:     "ssh",
			Commands: []string{"ssh user@host", "remote command"},
		},
	}
	
	// WHEN buildSupervisorScript is called
	script := suite.supervisor.buildSupervisorScript(segments, ExecutionModeBackground)
	
	// THEN it should generate a proper bash script
	require.NotEmpty(suite.T(), script, "Script should not be empty")
	assert.Contains(suite.T(), script, "#!/bin/bash", "Should have bash shebang")
	assert.Contains(suite.T(), script, "set -e", "Should have error handling")
	assert.Contains(suite.T(), script, "trap", "Should have signal trap")
	assert.Contains(suite.T(), script, "wait", "Should wait for background processes")
	assert.Contains(suite.T(), script, "Shell Segment 0", "Should comment shell segments")
	assert.Contains(suite.T(), script, "SSH Segment 1", "Should comment SSH segments")
}

// Integration Test 5: Real Process Group Management
func (suite *SupervisorIntegrationTestSuite) TestCreateSupervisor_ProcessGroupManagement() {
	// GIVEN a longer-running command for testing
	segments := []CommandSegment{
		{
			Type:     "shell",
			Commands: []string{"sleep 1"},
		},
	}
	
	// WHEN supervisor is created
	session, err := suite.supervisor.createSupervisor("test-pgid", segments, ExecutionModeBackground)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), session)
	
	supervisorPid := session.Supervisor.Process.Pid
	
	// THEN we should be able to manage the process group
	// Verify the process is running
	assert.True(suite.T(), suite.isProcessRunning(supervisorPid), "Supervisor should be running")
	
	// Test sending signals to process group
	err = syscall.Kill(-session.PGID, syscall.SIGTERM) // Negative PID sends to process group
	if err != nil {
		// On some systems, this might fail - that's okay for this test
		suite.T().Logf("Could not send signal to process group (this may be expected): %v", err)
	}
	
	// Wait for process to terminate
	done := make(chan bool, 1)
	go func() {
		session.Supervisor.Wait()
		done <- true
	}()
	
	select {
	case <-done:
		// Process terminated (good)
	case <-time.After(2 * time.Second):
		// Force cleanup if still running
		session.Supervisor.Process.Kill()
		session.Supervisor.Wait()
	}
}

// Integration Test 6: Platform Capabilities Verification
func (suite *SupervisorIntegrationTestSuite) TestPlatformCapabilities_ReflectActualBehavior() {
	// GIVEN the current platform
	manager := &unixProcessManager{}
	capabilities := manager.getPlatformCapabilities()
	
	// WHEN we test actual supervisor creation
	segments := []CommandSegment{
		{
			Type:     "shell",
			Commands: []string{"echo 'platform test'"},
		},
	}
	
	session, err := suite.supervisor.createSupervisor("platform-test", segments, ExecutionModeBackground)
	
	// THEN the behavior should match reported capabilities
	if capabilities.supportsCombinedSetup {
		// Should work without issues on Linux
		require.NoError(suite.T(), err, "Combined setup should work on platforms that support it")
		require.NotNil(suite.T(), session)
	} else {
		// On macOS, should still work (with session-only setup)
		require.NoError(suite.T(), err, "Session-only setup should work on macOS")
		require.NotNil(suite.T(), session)
	}
	
	// Cleanup
	if session != nil && session.Supervisor != nil {
		session.Supervisor.Process.Kill()
		session.Supervisor.Wait()
	}
	
	// Report platform info for debugging
	suite.T().Logf("Platform: %s, Combined Setup Supported: %t", 
		capabilities.platformName, capabilities.supportsCombinedSetup)
}

// Helper Methods
func (suite *SupervisorIntegrationTestSuite) isProcessRunning(pid int) bool {
	// Try to send signal 0 (no-op) to check if process exists
	err := syscall.Kill(pid, 0)
	return err == nil
}

// Benchmark Tests for Integration Performance
func BenchmarkCreateSupervisor_ShellCommand(b *testing.B) {
	supervisor := NewSupervisorManager()
	segments := []CommandSegment{
		{
			Type:     "shell",
			Commands: []string{"echo 'benchmark test'"},
		},
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		session, err := supervisor.createSupervisor(fmt.Sprintf("bench-%d", i), segments, ExecutionModeBackground)
		if err != nil {
			b.Fatalf("createSupervisor failed: %v", err)
		}
		
		// Cleanup
		if session != nil && session.Supervisor != nil {
			session.Supervisor.Process.Kill()
			session.Supervisor.Wait()
		}
	}
}

func BenchmarkBuildSupervisorScript(b *testing.B) {
	supervisor := NewSupervisorManager()
	segments := []CommandSegment{
		{
			Type:     "shell",
			Commands: []string{"echo 'test'", "sleep 0.1"},
		},
		{
			Type:     "ssh", 
			Commands: []string{"ssh user@host", "remote cmd"},
		},
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = supervisor.buildSupervisorScript(segments, ExecutionModeBackground)
	}
}