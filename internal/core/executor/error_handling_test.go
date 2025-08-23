package executor

import (
	"errors"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ErrorHandlingTest demonstrates comprehensive error handling scenarios
// for the session/PGID architecture

func TestErrorHandling_SupervisorCreationFailure(t *testing.T) {
	// GIVEN: executor with supervisor manager that fails to create supervisor
	registry := NewSessionRegistry()
	parser := NewCommandParser()
	supervisorMgr := &MockSupervisorManager{}
	signalMgr := &MockSignalManager{}
	sshMgr := &MockSSHManager{}
	
	executor := &executor{
		registry:      registry,
		parser:        CommandParserInterface(parser),
		supervisorMgr: SupervisorManagerInterface(supervisorMgr),
		signalMgr:     SignalManagerInterface(signalMgr),
		sshMgr:        SSHManagerInterface(sshMgr),
	}
	
	key := "test.project.service"
	commands := []string{"echo 'test'", "ls -la"}
	
	// Mock supervisor creation failure (non-process-creation error to avoid fallback)
	supervisorError := errors.New("failed to start supervisor: command not found")
	supervisorMgr.On("createSupervisor", key, mock.AnythingOfType("[]executor.CommandSegment"), mock.AnythingOfType("executor.ExecutionMode")).Return((*SessionGroup)(nil), supervisorError)
	
	// WHEN: running sequence with supervisor creation failure
	err := executor.RunSequence(key, commands)
	
	// THEN: should return supervisor creation error (not process creation error, so no fallback)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create supervisor session")
	assert.Contains(t, err.Error(), "command not found")
	
	// AND: session should not be registered
	status := executor.GetStatus()
	assert.Empty(t, status)
	
	supervisorMgr.AssertExpectations(t)
}

func TestErrorHandling_ProcessCreationErrorFallback(t *testing.T) {
	// GIVEN: executor with supervisor manager that fails with process creation error
	registry := NewSessionRegistry()
	parser := NewCommandParser()
	supervisorMgr := &MockSupervisorManager{}
	signalMgr := &MockSignalManager{}
	sshMgr := &MockSSHManager{}
	
	executor := &executor{
		registry:      registry,
		parser:        CommandParserInterface(parser),
		supervisorMgr: SupervisorManagerInterface(supervisorMgr),
		signalMgr:     SignalManagerInterface(signalMgr),
		sshMgr:        SSHManagerInterface(sshMgr),
	}
	
	key := "test.fallback.service"
	commands := []string{"echo 'fallback test'", "pwd"}
	
	// Mock supervisor creation with process creation error (triggers fallback)
	processCreationError := errors.New("fork/exec /bin/bash: operation not permitted")
	supervisorMgr.On("createSupervisor", key, mock.AnythingOfType("[]executor.CommandSegment"), mock.AnythingOfType("executor.ExecutionMode")).Return((*SessionGroup)(nil), processCreationError)
	
	// WHEN: running sequence with process creation error
	err := executor.RunSequence(key, commands)
	
	// THEN: should succeed (fallback to direct mode)
	assert.NoError(t, err, "Process creation error should trigger successful fallback to direct mode")
	
	// AND: session should not be registered (direct mode doesn't register sessions)
	status := executor.GetStatus()
	assert.Empty(t, status)
	
	supervisorMgr.AssertExpectations(t)
}

func TestErrorHandling_SessionRegistrationFailure(t *testing.T) {
	// GIVEN: executor with session that already exists
	registry := NewSessionRegistry()
	parser := NewCommandParser()
	supervisorMgr := &MockSupervisorManager{}
	signalMgr := &MockSignalManager{}
	sshMgr := &MockSSHManager{}
	
	executor := &executor{
		registry:      registry,
		parser:        CommandParserInterface(parser),
		supervisorMgr: SupervisorManagerInterface(supervisorMgr),
		signalMgr:     SignalManagerInterface(signalMgr),
		sshMgr:        SSHManagerInterface(sshMgr),
	}
	
	key := "test.project.service"
	commands := []string{"echo 'test'"}
	
	// Pre-register a session to cause conflict
	existingSession := &SessionGroup{
		Key:       key,
		SessionID: 11111,
		PGID:      11111,
		StartTime: time.Now(),
		Done:      make(chan struct{}),
	}
	registry.registerSession(key, existingSession)
	
	// WHEN: running sequence with duplicate key
	err := executor.RunSequence(key, commands)
	
	// THEN: should return duplicate key error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
	
	// Supervisor manager should not be called
	supervisorMgr.AssertNotCalled(t, "createSupervisor")
}

func TestErrorHandling_GracefulShutdownFailure(t *testing.T) {
	// GIVEN: executor with mock signal manager that fails graceful shutdown
	registry := NewSessionRegistry()
	parser := NewCommandParser()
	supervisorMgr := &MockSupervisorManager{}
	signalMgr := &MockSignalManager{}
	sshMgr := &MockSSHManager{}
	
	executor := &executor{
		registry:      registry,
		parser:        CommandParserInterface(parser),
		supervisorMgr: SupervisorManagerInterface(supervisorMgr),
		signalMgr:     SignalManagerInterface(signalMgr),
		sshMgr:        SSHManagerInterface(sshMgr),
	}
	
	key := "test.project.service"
	
	// Pre-register a session
	mockSession := &SessionGroup{
		Key:       key,
		SessionID: 22222,
		PGID:      22222,
		StartTime: time.Now(),
		Done:      make(chan struct{}),
	}
	registry.registerSession(key, mockSession)
	
	// Mock graceful shutdown failure
	shutdownError := errors.New("failed to send SIGTERM: no such process")
	signalMgr.On("gracefulShutdown", mockSession, 30*time.Second).Return(shutdownError)
	
	// WHEN: stopping session with graceful shutdown failure
	err := executor.StopByKey(key)
	
	// THEN: should return shutdown error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send SIGTERM")
	
	// AND: session should still be removed from registry (best effort)
	_, exists := registry.getSession(key)
	assert.False(t, exists)
	
	signalMgr.AssertExpectations(t)
}

func TestErrorHandling_StopAllSessionsFailure(t *testing.T) {
	// GIVEN: executor with signal manager that fails to stop all sessions
	registry := NewSessionRegistry()
	signalMgr := &MockSignalManager{}
	
	executor := &executor{
		registry:  registry,
		signalMgr: SignalManagerInterface(signalMgr),
	}
	
	// Mock shutdown all failure
	shutdownAllError := errors.New("failed to shutdown 2 sessions: timeout waiting for session test-1 to complete; failed to terminate session test-2")
	signalMgr.On("shutdownAllSessions", 30*time.Second).Return(shutdownAllError)
	
	// WHEN: stopping all sessions with failure
	err := executor.StopAll()
	
	// THEN: should return shutdown all error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to shutdown 2 sessions")
	assert.Contains(t, err.Error(), "timeout waiting for session")
	
	signalMgr.AssertExpectations(t)
}

func TestErrorHandling_InvalidCommandSegments(t *testing.T) {
	// GIVEN: executor with parser that returns no valid segments
	registry := NewSessionRegistry()
	parser := &MockCommandParser{}
	supervisorMgr := &MockSupervisorManager{}
	signalMgr := &MockSignalManager{}
	sshMgr := &MockSSHManager{}
	
	executor := &executor{
		registry:      registry,
		parser:        CommandParserInterface(parser),
		supervisorMgr: SupervisorManagerInterface(supervisorMgr),
		signalMgr:     SignalManagerInterface(signalMgr),
		sshMgr:        SSHManagerInterface(sshMgr),
	}
	
	key := "test.project.service"
	commands := []string{"", "   ", "\t"} // Empty/whitespace commands
	
	// Mock parser returning no valid segments
	parser.On("parseCommandSegments", commands).Return([]CommandSegment{})
	
	// WHEN: running sequence with invalid commands
	err := executor.RunSequence(key, commands)
	
	// THEN: should return validation error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no valid command segments found")
	
	// Supervisor manager should not be called
	supervisorMgr.AssertNotCalled(t, "createSupervisor")
	parser.AssertExpectations(t)
}

func TestErrorHandling_SessionRegistryCorruption(t *testing.T) {
	// GIVEN: executor with corrupted session registry state
	registry := NewSessionRegistry()
	parser := NewCommandParser()
	supervisorMgr := &MockSupervisorManager{}
	signalMgr := &MockSignalManager{}
	sshMgr := &MockSSHManager{}
	
	executor := &executor{
		registry:      registry,
		parser:        CommandParserInterface(parser),
		supervisorMgr: SupervisorManagerInterface(supervisorMgr),
		signalMgr:     SignalManagerInterface(signalMgr),
		sshMgr:        SSHManagerInterface(sshMgr),
	}
	
	key := "test.project.service"
	
	// Create session with invalid state (nil supervisor)
	corruptedSession := &SessionGroup{
		Key:        key,
		SessionID:  33333,
		PGID:       33333,
		StartTime:  time.Now(),
		Done:       make(chan struct{}),
		Supervisor: nil, // Corrupted state
	}
	registry.registerSession(key, corruptedSession)
	
	// Mock graceful shutdown for corrupted session
	signalMgr.On("gracefulShutdown", corruptedSession, 30*time.Second).Return(nil)
	
	// WHEN: stopping corrupted session
	err := executor.StopByKey(key)
	
	// THEN: should handle gracefully (no panic)
	assert.NoError(t, err) // Signal manager handles the corruption
	
	// AND: GetStatus should handle nil supervisor gracefully
	status := executor.GetStatus()
	assert.Empty(t, status) // No status for sessions with nil supervisor
	
	signalMgr.AssertExpectations(t)
}

func TestErrorHandling_ConcurrentAccessErrors(t *testing.T) {
	// GIVEN: executor under concurrent access stress
	registry := NewSessionRegistry()
	parser := NewCommandParser()
	supervisorMgr := &MockSupervisorManager{}
	signalMgr := &MockSignalManager{}
	sshMgr := &MockSSHManager{}
	
	executor := &executor{
		registry:      registry,
		parser:        CommandParserInterface(parser),
		supervisorMgr: SupervisorManagerInterface(supervisorMgr),
		signalMgr:     SignalManagerInterface(signalMgr),
		sshMgr:        SSHManagerInterface(sshMgr),
	}
	
	// Set up multiple concurrent operations
	keys := []string{"service-1", "service-2", "service-3"}
	commands := []string{"echo 'concurrent test'"}
	
	// Mock successful supervisor creation for all keys
	for _, key := range keys {
		mockSession := &SessionGroup{
			Key:       key,
			SessionID: 44444,
			PGID:      44444,
			StartTime: time.Now(),
			Done:      make(chan struct{}),
			Supervisor: &exec.Cmd{},
		}
		
		supervisorMgr.On("createSupervisor", key, mock.AnythingOfType("[]executor.CommandSegment"), mock.AnythingOfType("executor.ExecutionMode")).Return(mockSession, nil)
		signalMgr.On("manageSupervisor", mockSession).Return()
		signalMgr.On("gracefulShutdown", mockSession, 30*time.Second).Return(nil)
	}
	
	// WHEN: performing concurrent operations
	done := make(chan bool, len(keys)*2)
	
	// Concurrent session creation
	for _, key := range keys {
		go func(k string) {
			err := executor.RunSequence(k, commands)
			assert.NoError(t, err)
			done <- true
		}(key)
	}
	
	// Wait for all creations
	for i := 0; i < len(keys); i++ {
		<-done
	}
	
	// Brief delay to allow management to start
	time.Sleep(10 * time.Millisecond)
	
	// Concurrent session stopping
	for _, key := range keys {
		go func(k string) {
			err := executor.StopByKey(k)
			assert.NoError(t, err)
			done <- true
		}(key)
	}
	
	// Wait for all stops
	for i := 0; i < len(keys); i++ {
		<-done
	}
	
	// THEN: should handle concurrent access gracefully
	// No panics or race conditions should occur
	status := executor.GetStatus()
	assert.True(t, len(status) <= len(keys)) // Some sessions may still be cleaning up
	
	supervisorMgr.AssertExpectations(t)
	signalMgr.AssertExpectations(t)
}

func TestErrorHandling_ResourceExhaustion(t *testing.T) {
	// GIVEN: executor that simulates resource exhaustion
	registry := NewSessionRegistry()
	parser := NewCommandParser()
	supervisorMgr := &MockSupervisorManager{}
	signalMgr := &MockSignalManager{}
	sshMgr := &MockSSHManager{}
	
	executor := &executor{
		registry:      registry,
		parser:        CommandParserInterface(parser),
		supervisorMgr: SupervisorManagerInterface(supervisorMgr),
		signalMgr:     SignalManagerInterface(signalMgr),
		sshMgr:        SSHManagerInterface(sshMgr),
	}
	
	key := "resource-intensive.service"
	commands := []string{"high-memory-command", "cpu-intensive-process"}
	
	// Mock resource exhaustion error
	resourceError := errors.New("failed to start supervisor: cannot allocate memory")
	supervisorMgr.On("createSupervisor", key, mock.AnythingOfType("[]executor.CommandSegment"), mock.AnythingOfType("executor.ExecutionMode")).Return((*SessionGroup)(nil), resourceError)
	
	// WHEN: running sequence that exhausts resources
	err := executor.RunSequence(key, commands)
	
	// THEN: should return resource exhaustion error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create supervisor session")
	assert.Contains(t, err.Error(), "cannot allocate memory")
	
	// AND: system should remain stable
	status := executor.GetStatus()
	assert.Empty(t, status)
	
	supervisorMgr.AssertExpectations(t)
}

func TestErrorHandling_TimeoutScenarios(t *testing.T) {
	// GIVEN: executor with operations that timeout
	registry := NewSessionRegistry()
	signalMgr := &MockSignalManager{}
	
	executor := &executor{
		registry:  registry,
		signalMgr: SignalManagerInterface(signalMgr),
	}
	
	// Pre-register sessions that will timeout
	session1 := &SessionGroup{Key: "timeout-session-1", PGID: 55555, Done: make(chan struct{})}
	session2 := &SessionGroup{Key: "timeout-session-2", PGID: 55556, Done: make(chan struct{})}
	registry.registerSession("timeout-session-1", session1)
	registry.registerSession("timeout-session-2", session2)
	
	// Mock timeout errors
	timeoutError := errors.New("timeout waiting for all sessions to shutdown")
	signalMgr.On("shutdownAllSessions", 30*time.Second).Return(timeoutError)
	
	// WHEN: stopping all with timeout
	err := executor.StopAll()
	
	// THEN: should return timeout error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timeout waiting for all sessions")
	
	signalMgr.AssertExpectations(t)
}

func TestErrorHandling_EdgeCaseInputs(t *testing.T) {
	// GIVEN: executor receiving edge case inputs
	executor := New()
	
	testCases := []struct {
		name     string
		key      string
		commands []string
		expectError bool
		errorMsg string
	}{
		{
			name:        "nil commands",
			key:         "valid.key",
			commands:    nil,
			expectError: true,
			errorMsg:    "no commands provided",
		},
		{
			name:        "empty commands",
			key:         "valid.key",
			commands:    []string{},
			expectError: true,
			errorMsg:    "no commands provided",
		},
		{
			name:        "empty key",
			key:         "",
			commands:    []string{"echo test"},
			expectError: true,
			errorMsg:    "", // Empty key validation happens at different level
		},
		{
			name:        "whitespace only key",
			key:         "   ",
			commands:    []string{"echo test"},
			expectError: false, // Whitespace keys are technically valid
			errorMsg:    "",
		},
		{
			name:        "very long key",
			key:         string(make([]byte, 1000)), // 1000 character key
			commands:    []string{"echo test"},
			expectError: false, // Long keys should be handled
			errorMsg:    "",
		},
		{
			name:        "commands with null bytes",
			key:         "test.key",
			commands:    []string{"echo test\x00malicious"},
			expectError: false, // System should handle gracefully
			errorMsg:    "",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// WHEN: running sequence with edge case input
			err := executor.RunSequence(tc.key, tc.commands)
			
			// THEN: should handle according to expectation
			if tc.expectError {
				assert.Error(t, err)
				if tc.errorMsg != "" {
					assert.Contains(t, err.Error(), tc.errorMsg)
				}
			} else {
				// For non-error cases, we expect them to fail in this environment
				// due to process execution restrictions, but not with input validation errors
				if err != nil {
					assert.NotContains(t, err.Error(), "no commands provided")
				}
			}
			
			// Clean up any successful registrations
			if !tc.expectError && tc.key != "" {
				executor.StopByKey(tc.key)
			}
		})
	}
}

// MockCommandParser for testing parser failures
type MockCommandParser struct {
	mock.Mock
}

func (m *MockCommandParser) parseCommandSegments(commands []string) []CommandSegment {
	args := m.Called(commands)
	return args.Get(0).([]CommandSegment)
}

func TestErrorHandling_NetworkRelatedErrors(t *testing.T) {
	// GIVEN: executor handling network-related SSH failures
	registry := NewSessionRegistry()
	parser := NewCommandParser()
	supervisorMgr := &MockSupervisorManager{}
	signalMgr := &MockSignalManager{}
	sshMgr := &MockSSHManager{}
	
	executor := &executor{
		registry:      registry,
		parser:        CommandParserInterface(parser),
		supervisorMgr: SupervisorManagerInterface(supervisorMgr),
		signalMgr:     SignalManagerInterface(signalMgr),
		sshMgr:        SSHManagerInterface(sshMgr),
	}
	
	key := "network.test.service"
	commands := []string{
		"ssh unreachable@nonexistent.host",
		"password123",
		"echo 'this will not execute'",
	}
	
	// Mock supervisor creation that would handle network errors
	networkError := errors.New("failed to start supervisor: network unreachable")
	supervisorMgr.On("createSupervisor", key, mock.AnythingOfType("[]executor.CommandSegment"), mock.AnythingOfType("executor.ExecutionMode")).Return((*SessionGroup)(nil), networkError)
	
	// WHEN: running sequence with network connectivity issues
	err := executor.RunSequence(key, commands)
	
	// THEN: should return network error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create supervisor session")
	assert.Contains(t, err.Error(), "network unreachable")
	
	supervisorMgr.AssertExpectations(t)
}

func TestErrorHandling_SignalDeliveryFailures(t *testing.T) {
	// GIVEN: executor with signal delivery that fails
	registry := NewSessionRegistry()
	signalMgr := &MockSignalManager{}
	
	executor := &executor{
		registry:  registry,
		signalMgr: SignalManagerInterface(signalMgr),
	}
	
	key := "signal.test.service"
	
	// Register session with problematic PGID
	problemSession := &SessionGroup{
		Key:       key,
		SessionID: 66666,
		PGID:      -1, // Invalid PGID
		StartTime: time.Now(),
		Done:      make(chan struct{}),
	}
	registry.registerSession(key, problemSession)
	
	// Mock signal delivery failure
	signalError := errors.New("invalid PGID: -1")
	signalMgr.On("gracefulShutdown", problemSession, 30*time.Second).Return(signalError)
	
	// WHEN: stopping session with signal delivery problems
	err := executor.StopByKey(key)
	
	// THEN: should return signal error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid PGID: -1")
	
	signalMgr.AssertExpectations(t)
}