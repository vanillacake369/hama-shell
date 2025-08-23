package executor

import (
	"os"
	"os/exec"

	"github.com/stretchr/testify/mock"
)

// MockProcessManager is a mock implementation of processManager
type MockProcessManager struct {
	mock.Mock
}

func (m *MockProcessManager) setupCommand(cmd *exec.Cmd) {
	m.Called(cmd)
}

func (m *MockProcessManager) terminateProcess(process *os.Process) error {
	args := m.Called(process)
	return args.Error(0)
}

func (m *MockProcessManager) setupSupervisor(cmd *exec.Cmd) error {
	args := m.Called(cmd)
	return args.Error(0)
}

// MockPTY is a mock implementation of PTYInterface
type MockPTY struct {
	mock.Mock
}

func (m *MockPTY) Start(cmd *exec.Cmd) (PTYFile, error) {
	args := m.Called(cmd)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(PTYFile), args.Error(1)
}

// MockSupervisor is a mock implementation of SupervisorInterface
type MockSupervisor struct {
	mock.Mock
}

func (m *MockSupervisor) Start() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockSupervisor) Wait() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockSupervisor) Signal(sig os.Signal) error {
	args := m.Called(sig)
	return args.Error(0)
}

func (m *MockSupervisor) GetPID() int {
	args := m.Called()
	return args.Int(0)
}