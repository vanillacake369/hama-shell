package executor

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestExecutor_RunSequence_Success(t *testing.T) {
	// GIVEN: executor instance with shell commands
	e := New()
	key := "test.project.service"
	commands := []string{"echo 'hello'", "pwd", "ls -la"}
	
	// WHEN: running sequence
	err := e.RunSequence(key, commands)
	
	// THEN: should succeed
	assert.NoError(t, err)
	
	// AND: session should be registered
	status := e.GetStatus()
	assert.Contains(t, status, key)
	assert.Len(t, status[key], 1) // One supervisor process
	
	// Clean up
	e.StopByKey(key)
}

func TestExecutor_RunSequence_EmptyCommands(t *testing.T) {
	// GIVEN: executor instance with empty commands
	e := New()
	key := "test.project.service"
	commands := []string{}
	
	// WHEN: running sequence with empty commands
	err := e.RunSequence(key, commands)
	
	// THEN: should return error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no commands provided")
}

func TestExecutor_RunSequence_DuplicateKey(t *testing.T) {
	// GIVEN: executor instance with active session
	e := New()
	key := "test.project.service"
	commands := []string{"sleep 0.1"}
	
	// WHEN: running first sequence
	err1 := e.RunSequence(key, commands)
	assert.NoError(t, err1)
	
	// WHEN: running second sequence with same key
	err2 := e.RunSequence(key, commands)
	
	// THEN: second should fail due to duplicate key
	assert.Error(t, err2)
	assert.Contains(t, err2.Error(), "already exists")
	
	// Clean up
	e.StopByKey(key)
}

func TestExecutor_StopByKey_Success(t *testing.T) {
	// GIVEN: executor instance with running session
	e := New()
	key := "test.project.service"
	commands := []string{"sleep 1"}
	
	err := e.RunSequence(key, commands)
	assert.NoError(t, err)
	
	// Verify session exists
	status := e.GetStatus()
	assert.Contains(t, status, key)
	
	// WHEN: stopping by key
	err = e.StopByKey(key)
	
	// THEN: should succeed
	assert.NoError(t, err)
	
	// AND: session should be removed
	time.Sleep(100 * time.Millisecond) // Brief wait for cleanup
	status = e.GetStatus()
	assert.NotContains(t, status, key)
}

func TestExecutor_StopByKey_NonExistentKey(t *testing.T) {
	// GIVEN: executor instance
	e := New()
	
	// WHEN: stopping non-existent key
	err := e.StopByKey("non-existent-key")
	
	// THEN: should succeed (no-op)
	assert.NoError(t, err)
}

func TestExecutor_StopAll_Success(t *testing.T) {
	// GIVEN: executor instance with multiple running sessions
	e := New()
	key1 := "test.project.service1"
	key2 := "test.project.service2"
	commands := []string{"sleep 1"}
	
	err1 := e.RunSequence(key1, commands)
	err2 := e.RunSequence(key2, commands)
	assert.NoError(t, err1)
	assert.NoError(t, err2)
	
	// Verify sessions exist
	status := e.GetStatus()
	assert.Contains(t, status, key1)
	assert.Contains(t, status, key2)
	
	// WHEN: stopping all sessions
	err := e.StopAll()
	
	// THEN: should succeed
	assert.NoError(t, err)
	
	// AND: all sessions should be removed (with proper cleanup time)
	time.Sleep(500 * time.Millisecond) // Wait for cleanup to complete
	status = e.GetStatus()
	assert.Empty(t, status, "All sessions should be cleaned up after StopAll()")
}

func TestExecutor_StopAll_EmptyRegistry(t *testing.T) {
	// GIVEN: executor instance with no active sessions
	e := New()
	
	// WHEN: stopping all sessions
	err := e.StopAll()
	
	// THEN: should succeed
	assert.NoError(t, err)
}

func TestExecutor_GetStatus_Success(t *testing.T) {
	// GIVEN: executor instance with running session
	e := New()
	key := "test.project.service"
	commands := []string{"sleep 1"}
	
	err := e.RunSequence(key, commands)
	assert.NoError(t, err)
	
	// WHEN: getting status
	status := e.GetStatus()
	
	// THEN: should return session information
	assert.Contains(t, status, key)
	processInfo := status[key][0]
	assert.Equal(t, key, processInfo.Key)
	assert.Contains(t, processInfo.Command, "Session supervisor")
	assert.Greater(t, processInfo.PID, 0)
	assert.Greater(t, processInfo.StartTime, int64(0))
	
	// Clean up
	e.StopByKey(key)
}

func TestExecutor_GetStatus_EmptyRegistry(t *testing.T) {
	// GIVEN: executor instance with no active sessions
	e := New()
	
	// WHEN: getting status
	status := e.GetStatus()
	
	// THEN: should return empty status
	assert.Empty(t, status)
}

func TestExecutor_RunSequence_SSHCommands(t *testing.T) {
	// GIVEN: executor instance with SSH commands
	e := New()
	key := "test.ssh.service"
	commands := []string{
		"ssh user@host",
		"secretpassword",
		"ls -la",
		"pwd",
	}
	
	// WHEN: running SSH sequence
	err := e.RunSequence(key, commands)
	
	// THEN: should succeed (supervisor created)
	assert.NoError(t, err)
	
	// AND: session should be registered
	status := e.GetStatus()
	assert.Contains(t, status, key)
	
	// Clean up
	e.StopByKey(key)
}

func TestExecutor_RunSequence_MixedCommands(t *testing.T) {
	// GIVEN: executor instance with mixed SSH and shell commands
	e := New()
	key := "test.mixed.service"
	commands := []string{
		"echo 'local command'",
		"ssh user@host1",
		"password123",
		"ls -la",
		"ssh user@host2", 
		"password456",
		"pwd",
		"echo 'final local'",
	}
	
	// WHEN: running mixed sequence
	err := e.RunSequence(key, commands)
	
	// THEN: should succeed
	assert.NoError(t, err)
	
	// AND: session should be registered
	status := e.GetStatus()
	assert.Contains(t, status, key)
	
	// Clean up
	e.StopByKey(key)
}

// MockExecutorTestComponents provides mocked components for integration testing
type MockExecutorTestComponents struct {
	registry      *SessionRegistry
	parser        *CommandParser
	supervisorMgr *MockSupervisorManager
	signalMgr     *MockSignalManager
	sshMgr        *MockSSHManager
}

type MockSupervisorManager struct {
	mock.Mock
}

func (m *MockSupervisorManager) createSupervisor(key string, segments []CommandSegment, mode ExecutionMode) (*SessionGroup, error) {
	args := m.Called(key, segments, mode)
	return args.Get(0).(*SessionGroup), args.Error(1)
}

type MockSignalManager struct {
	mock.Mock
}

func (m *MockSignalManager) manageSupervisor(session *SessionGroup) {
	m.Called(session)
}

func (m *MockSignalManager) gracefulShutdown(session *SessionGroup, timeout time.Duration) error {
	args := m.Called(session, timeout)
	return args.Error(0)
}

func (m *MockSignalManager) shutdownAllSessions(timeout time.Duration) error {
	args := m.Called(timeout)
	return args.Error(0)
}

type MockSSHManager struct {
	mock.Mock
}

func (m *MockSSHManager) executeSSHWithPTY(sshCmd, password string, remoteCmds []string) error {
	args := m.Called(sshCmd, password, remoteCmds)
	return args.Error(0)
}

func TestExecutor_MockedComponents_RunSequence(t *testing.T) {
	// GIVEN: executor with mocked components
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
	
	// Mock expectations
	mockSession := &SessionGroup{
		Key:       key,
		SessionID: 12345,
		PGID:      12345,
		StartTime: time.Now(),
		Done:      make(chan struct{}),
	}
	supervisorMgr.On("createSupervisor", key, mock.AnythingOfType("[]executor.CommandSegment"), mock.AnythingOfType("executor.ExecutionMode")).Return(mockSession, nil)
	
	// Use a channel to signal when manageSupervisor is called
	manageSupCalled := make(chan bool, 1)
	signalMgr.On("manageSupervisor", mockSession).Run(func(args mock.Arguments) {
		manageSupCalled <- true
	}).Return()
	
	// WHEN: running sequence
	err := executor.RunSequence(key, commands)
	
	// THEN: should succeed
	assert.NoError(t, err)
	
	// Wait for manageSupervisor to be called (with timeout)
	select {
	case <-manageSupCalled:
		// Good - manageSupervisor was called
	case <-time.After(100 * time.Millisecond):
		t.Error("manageSupervisor was not called within timeout")
	}
	
	// Verify mocks
	supervisorMgr.AssertExpectations(t)
	signalMgr.AssertExpectations(t)
}

func TestExecutor_MockedComponents_StopAll(t *testing.T) {
	// GIVEN: executor with mocked signal manager
	registry := NewSessionRegistry()
	signalMgr := &MockSignalManager{}
	
	executor := &executor{
		registry:  registry,
		signalMgr: SignalManagerInterface(signalMgr),
	}
	
	// Mock expectations
	signalMgr.On("shutdownAllSessions", 30*time.Second).Return(nil)
	
	// WHEN: stopping all sessions
	err := executor.StopAll()
	
	// THEN: should succeed
	assert.NoError(t, err)
	
	// Verify mock
	signalMgr.AssertExpectations(t)
}