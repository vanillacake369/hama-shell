package session

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"hama-shell/internal/core/config"
	"hama-shell/internal/core/executor"
)

// MockExecutor is a mock implementation of executor.Executor
type MockExecutor struct {
	mock.Mock
}

func (m *MockExecutor) RunSequence(key string, commands []string) error {
	args := m.Called(key, commands)
	return args.Error(0)
}

func (m *MockExecutor) StopByKey(key string) error {
	args := m.Called(key)
	return args.Error(0)
}

func (m *MockExecutor) StopAll() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockExecutor) GetStatus() map[string][]*executor.ProcessInfo {
	args := m.Called()
	if result := args.Get(0); result == nil {
		return nil
	}
	return args.Get(0).(map[string][]*executor.ProcessInfo)
}

// MockConfigService is a mock implementation of config.Service
type MockConfigService struct {
	mock.Mock
}

func (m *MockConfigService) Load(path string) (*config.Config, error) {
	args := m.Called(path)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*config.Config), args.Error(1)
}

func (m *MockConfigService) List(cfg *config.Config) []string {
	args := m.Called(cfg)
	if result := args.Get(0); result == nil {
		return []string{}
	}
	return args.Get(0).([]string)
}

func (m *MockConfigService) ResolveTarget(target string, cfg *config.Config) (*config.Service, error) {
	args := m.Called(target, cfg)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*config.Service), args.Error(1)
}

func TestManager_Start(t *testing.T) {
	t.Run("successful session start", func(t *testing.T) {
		// GIVEN a manager with mocked dependencies
		mockExec := new(MockExecutor)
		mockConfig := new(MockConfigService)
		manager := NewManager(mockExec, mockConfig)

		testConfig := &config.Config{}
		testService := &config.Service{
			Description: "Test service",
			Commands:    []string{"echo hello", "echo world"},
		}

		mockConfig.On("Load", "/path/to/config.yaml").Return(testConfig, nil)
		mockConfig.On("ResolveTarget", "myapp.dev.api", testConfig).Return(testService, nil)
		mockExec.On("RunSequence", "myapp.dev.api", []string{"echo hello", "echo world"}).Return(nil)

		// WHEN we start a session
		err := manager.Start("myapp.dev.api", "/path/to/config.yaml")

		// THEN it should succeed and execute all commands
		assert.NoError(t, err)
		mockConfig.AssertExpectations(t)
		mockExec.AssertExpectations(t)
	})

	t.Run("config load failure", func(t *testing.T) {
		// GIVEN a manager where config loading fails
		mockExec := new(MockExecutor)
		mockConfig := new(MockConfigService)
		manager := NewManager(mockExec, mockConfig)

		mockConfig.On("Load", "/invalid/config.yaml").Return(nil, errors.New("file not found"))

		// WHEN we start a session with invalid config
		err := manager.Start("myapp.dev.api", "/invalid/config.yaml")

		// THEN it should return config load error
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to load config")
		mockConfig.AssertExpectations(t)
		mockExec.AssertNotCalled(t, "RunSequence")
	})

	t.Run("target resolution failure", func(t *testing.T) {
		// GIVEN a manager where target resolution fails
		mockExec := new(MockExecutor)
		mockConfig := new(MockConfigService)
		manager := NewManager(mockExec, mockConfig)

		testConfig := &config.Config{}
		mockConfig.On("Load", "/path/to/config.yaml").Return(testConfig, nil)
		mockConfig.On("ResolveTarget", "invalid.target", testConfig).Return(nil, errors.New("invalid target"))

		// WHEN we start a session with invalid target
		err := manager.Start("invalid.target", "/path/to/config.yaml")

		// THEN it should return target resolution error
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to resolve target")
		mockConfig.AssertExpectations(t)
		mockExec.AssertNotCalled(t, "RunSequence")
	})

	t.Run("command execution failure", func(t *testing.T) {
		// GIVEN a manager where command execution fails
		mockExec := new(MockExecutor)
		mockConfig := new(MockConfigService)
		manager := NewManager(mockExec, mockConfig)

		testConfig := &config.Config{}
		testService := &config.Service{
			Commands: []string{"echo hello", "failing command", "echo never"},
		}

		mockConfig.On("Load", "/path/to/config.yaml").Return(testConfig, nil)
		mockConfig.On("ResolveTarget", "myapp.dev.api", testConfig).Return(testService, nil)
		mockExec.On("RunSequence", "myapp.dev.api", []string{"echo hello", "failing command", "echo never"}).Return(errors.New("command failed"))

		// WHEN we start a session that fails
		err := manager.Start("myapp.dev.api", "/path/to/config.yaml")

		// THEN it should return error
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to execute commands")
		mockExec.AssertExpectations(t)
	})
}

func TestManager_Stop(t *testing.T) {
	t.Run("successful session stop", func(t *testing.T) {
		// GIVEN a manager with active sessions
		mockExec := new(MockExecutor)
		mockConfig := new(MockConfigService)
		manager := NewManager(mockExec, mockConfig)

		mockExec.On("StopByKey", "myapp.dev.api").Return(nil)

		// WHEN we stop a session
		err := manager.Stop("myapp.dev.api")

		// THEN it should stop the session
		assert.NoError(t, err)
		mockExec.AssertExpectations(t)
	})

	t.Run("stop failure", func(t *testing.T) {
		// GIVEN a manager where stop fails
		mockExec := new(MockExecutor)
		mockConfig := new(MockConfigService)
		manager := NewManager(mockExec, mockConfig)

		mockExec.On("StopByKey", "myapp.dev.api").Return(errors.New("no such session"))

		// WHEN we try to stop a session
		err := manager.Stop("myapp.dev.api")

		// THEN it should return error
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to stop session")
		mockExec.AssertExpectations(t)
	})
}

func TestManager_GetStatus(t *testing.T) {
	// GIVEN a manager with some active sessions
	mockExec := new(MockExecutor)
	mockConfig := new(MockConfigService)
	manager := NewManager(mockExec, mockConfig)

	expectedStatus := map[string][]*executor.ProcessInfo{
		"myapp.dev.api": {
			{Command: "echo hello", PID: 1234, StartTime: 12345, Key: "myapp.dev.api"},
			{Command: "echo world", PID: 1235, StartTime: 12346, Key: "myapp.dev.api"},
		},
	}

	mockExec.On("GetStatus").Return(expectedStatus)

	// WHEN we get status
	status := manager.GetStatus()

	// THEN it should return all statuses
	assert.Equal(t, expectedStatus, status)
	mockExec.AssertExpectations(t)
}

func TestManager_GetTargetStatus(t *testing.T) {
	// GIVEN a manager with mixed active sessions
	mockExec := new(MockExecutor)
	mockConfig := new(MockConfigService)
	manager := NewManager(mockExec, mockConfig)

	process1 := &executor.ProcessInfo{Command: "echo hello", PID: 1234, StartTime: 12345, Key: "myapp.dev.api"}
	process2 := &executor.ProcessInfo{Command: "echo world", PID: 1235, StartTime: 12346, Key: "myapp.dev.api"}
	
	allStatus := map[string][]*executor.ProcessInfo{
		"myapp.dev.api": {process1, process2},
		"myapp.prod.api": {
			{Command: "prod cmd", PID: 1236, StartTime: 12347, Key: "myapp.prod.api"},
		},
	}

	mockExec.On("GetStatus").Return(allStatus)

	// WHEN we get status for a specific target
	targetStatus := manager.GetTargetStatus("myapp.dev.api")

	// THEN it should return only matching processes
	assert.Len(t, targetStatus, 2)
	assert.Equal(t, process1, targetStatus[0])
	assert.Equal(t, process2, targetStatus[1])
	mockExec.AssertExpectations(t)
}

func TestManager_GetTargetStatus_NotFound(t *testing.T) {
	// GIVEN a manager with no matching sessions
	mockExec := new(MockExecutor)
	mockConfig := new(MockConfigService)
	manager := NewManager(mockExec, mockConfig)

	allStatus := map[string][]*executor.ProcessInfo{
		"other.service": {
			{Command: "other cmd", PID: 1237, StartTime: 12348, Key: "other.service"},
		},
	}

	mockExec.On("GetStatus").Return(allStatus)

	// WHEN we get status for a non-existent target
	targetStatus := manager.GetTargetStatus("myapp.dev.api")

	// THEN it should return empty list
	assert.Empty(t, targetStatus)
	mockExec.AssertExpectations(t)
}

func TestManager_StopAll(t *testing.T) {
	// GIVEN a manager
	mockExec := new(MockExecutor)
	mockConfig := new(MockConfigService)
	manager := NewManager(mockExec, mockConfig)

	mockExec.On("StopAll").Return(nil)

	// WHEN we stop all sessions
	err := manager.StopAll()

	// THEN it should delegate to executor
	assert.NoError(t, err)
	mockExec.AssertExpectations(t)
}

func TestNewManager_NilDependencies(t *testing.T) {
	// GIVEN nil dependencies
	// WHEN we create a manager with nil dependencies
	manager := NewManager(nil, nil)

	// THEN it should create default instances
	assert.NotNil(t, manager)
	assert.NotNil(t, manager.executor)
	assert.NotNil(t, manager.configSvc)
}