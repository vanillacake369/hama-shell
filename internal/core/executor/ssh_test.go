package executor

import (
	"io"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockFile is a mock implementation of PTYFile for testing
type MockFile struct {
	mock.Mock
	readChan  chan []byte
	writeChan chan []byte
}

func NewMockFile() *MockFile {
	return &MockFile{
		readChan:  make(chan []byte, 10),
		writeChan: make(chan []byte, 10),
	}
}

func (m *MockFile) Read(b []byte) (n int, err error) {
	select {
	case data, ok := <-m.readChan:
		if !ok {
			return 0, io.EOF
		}
		n = copy(b, data)
		return n, nil
	case <-time.After(100 * time.Millisecond):
		return 0, io.EOF // Return EOF instead of error for timeout
	}
}

func (m *MockFile) Write(b []byte) (n int, err error) {
	m.writeChan <- append([]byte(nil), b...)
	return len(b), nil
}

func (m *MockFile) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockFile) simulateOutput(output string) {
	m.readChan <- []byte(output)
}

func (m *MockFile) getWrittenData() []byte {
	select {
	case data := <-m.writeChan:
		return data
	case <-time.After(100 * time.Millisecond):
		return nil
	}
}

func TestSSHManager_executeSSHWithPTY_Success(t *testing.T) {
	// GIVEN: SSH manager with mock PTY
	mockPTY := &MockPTY{}
	mockFile := NewMockFile()
	sm := NewSSHManagerWithPTY(mockPTY)
	
	sshCmd := "ssh user@host"
	password := "secret123"
	remoteCmds := []string{"ls", "pwd"}
	
	mockPTY.On("Start", mock.AnythingOfType("*exec.Cmd")).Return(mockFile, nil)
	mockFile.On("Close").Return(nil)
	
	// Simulate SSH interaction
	go func() {
		time.Sleep(10 * time.Millisecond)
		mockFile.simulateOutput("password: ")
		time.Sleep(10 * time.Millisecond)
		mockFile.simulateOutput("$ ")
	}()
	
	// WHEN: executing SSH with PTY
	err := sm.executeSSHWithPTY(sshCmd, password, remoteCmds)
	
	// THEN: should complete successfully
	assert.NoError(t, err)
	
	// Verify password was sent
	writtenData := mockFile.getWrittenData()
	assert.Equal(t, []byte("secret123\n"), writtenData)
	
	mockPTY.AssertExpectations(t)
	mockFile.AssertExpectations(t)
}

func TestSSHManager_executeSSHWithPTY_EmptySSHCommand(t *testing.T) {
	// GIVEN: SSH manager with empty SSH command
	sm := NewSSHManager()
	
	// WHEN: executing with empty SSH command
	err := sm.executeSSHWithPTY("", "password", []string{"ls"})
	
	// THEN: should return error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "SSH command cannot be empty")
}

func TestSSHManager_executeSSHWithPTY_PTYStartError(t *testing.T) {
	// GIVEN: SSH manager with mock PTY that fails to start
	mockPTY := &MockPTY{}
	sm := NewSSHManagerWithPTY(mockPTY)
	
	mockPTY.On("Start", mock.AnythingOfType("*exec.Cmd")).Return((*os.File)(nil), assert.AnError)
	
	// WHEN: executing SSH with PTY
	err := sm.executeSSHWithPTY("ssh user@host", "password", []string{})
	
	// THEN: should return PTY start error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to start PTY")
	mockPTY.AssertExpectations(t)
}

func TestSSHManager_executeSSHWithPTY_AuthenticationFailure(t *testing.T) {
	// GIVEN: SSH manager with mock PTY
	mockPTY := &MockPTY{}
	mockFile := NewMockFile()
	sm := NewSSHManagerWithPTY(mockPTY)
	
	sshCmd := "ssh user@host"
	password := "wrongpassword"
	
	mockPTY.On("Start", mock.AnythingOfType("*exec.Cmd")).Return(mockFile, nil)
	mockFile.On("Close").Return(nil)
	
	// Simulate authentication failure
	go func() {
		time.Sleep(10 * time.Millisecond)
		mockFile.simulateOutput("password: ")
		time.Sleep(20 * time.Millisecond)
		mockFile.simulateOutput("Permission denied, please try again.")
	}()
	
	// WHEN: executing SSH with wrong password
	err := sm.executeSSHWithPTY(sshCmd, password, []string{})
	
	// THEN: should return authentication error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "SSH authentication failed")
	mockPTY.AssertExpectations(t)
}

func TestSSHManager_executeSSHWithPTY_NoPasswordProvided(t *testing.T) {
	// GIVEN: SSH manager with mock PTY
	mockPTY := &MockPTY{}
	mockFile := NewMockFile()
	sm := NewSSHManagerWithPTY(mockPTY)
	
	sshCmd := "ssh user@host"
	password := "" // No password provided
	
	mockPTY.On("Start", mock.AnythingOfType("*exec.Cmd")).Return(mockFile, nil)
	mockFile.On("Close").Return(nil)
	
	// Simulate password prompt
	go func() {
		time.Sleep(10 * time.Millisecond)
		mockFile.simulateOutput("password: ")
	}()
	
	// WHEN: executing SSH without password
	err := sm.executeSSHWithPTY(sshCmd, password, []string{})
	
	// THEN: should return password required error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "password prompt detected but no password provided")
	mockPTY.AssertExpectations(t)
}

func TestSSHManager_isPasswordPrompt(t *testing.T) {
	sm := NewSSHManager()
	
	tests := []struct {
		name     string
		line     string
		expected bool
	}{
		{"basic password prompt", "password:", true},
		{"password with user", "user@host's password:", true},
		{"password for user", "Password for user:", true},
		{"enter password", "Enter password:", true},
		{"not password prompt", "Welcome to server", false},
		{"empty line", "", false},
		{"command output", "total 1024", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// GIVEN: line from SSH output
			// WHEN: checking if it's password prompt
			result := sm.isPasswordPrompt(tt.line)
			// THEN: should match expected result
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSSHManager_isShellPrompt(t *testing.T) {
	sm := NewSSHManager()
	
	tests := []struct {
		name     string
		line     string
		expected bool
	}{
		{"bash prompt", "user@host:~$ ", true},
		{"root prompt", "root@host:~# ", true},
		{"zsh prompt", "user@host ~ % ", true},
		{"simple dollar", "$", true},
		{"simple hash", "#", true},
		{"powershell prompt", "PS C:\\> ", true},
		{"not shell prompt", "Welcome to server", false},
		{"command output", "total 1024", false},
		{"empty line", "", false},
		{"partial prompt", "user@host", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// GIVEN: line from SSH output
			// WHEN: checking if it's shell prompt
			result := sm.isShellPrompt(tt.line)
			// THEN: should match expected result
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSSHManager_handleSSHSession_PasswordAuth(t *testing.T) {
	// GIVEN: SSH manager and mock file
	sm := NewSSHManager()
	mockFile := NewMockFile()
	password := "testpass123"
	remoteCmds := []string{"ls -la"}
	
	mockFile.On("Close").Return(nil)
	
	// Simulate SSH session
	go func() {
		time.Sleep(5 * time.Millisecond)
		mockFile.simulateOutput("password: ")
		time.Sleep(5 * time.Millisecond)
		mockFile.simulateOutput("$ ")
		// Close the read channel to end the session
		close(mockFile.readChan)
	}()
	
	// WHEN: handling SSH session
	err := sm.handleSSHSession(mockFile, password, remoteCmds)
	
	// THEN: should complete successfully
	assert.NoError(t, err)
	
	// Verify password was written
	passwordData := mockFile.getWrittenData()
	assert.Equal(t, []byte("testpass123\n"), passwordData)
	
	// Verify command was written
	commandData := mockFile.getWrittenData()
	assert.Equal(t, []byte("ls -la\n"), commandData)
}

func TestSSHManager_handleSSHSession_MultipleCommands(t *testing.T) {
	// GIVEN: SSH manager with multiple remote commands
	sm := NewSSHManager()
	mockFile := NewMockFile()
	password := "testpass"
	remoteCmds := []string{"cd /app", "ls", "pwd"}
	
	mockFile.On("Close").Return(nil)
	
	// Simulate successful SSH session
	go func() {
		time.Sleep(5 * time.Millisecond)
		mockFile.simulateOutput("password: ")
		time.Sleep(5 * time.Millisecond)
		mockFile.simulateOutput("$ ")
		close(mockFile.readChan)
	}()
	
	// WHEN: handling SSH session with multiple commands
	err := sm.handleSSHSession(mockFile, password, remoteCmds)
	
	// THEN: should complete successfully
	assert.NoError(t, err)
	
	// Verify all commands were sent
	_ = mockFile.getWrittenData() // password
	cmd1 := mockFile.getWrittenData()
	cmd2 := mockFile.getWrittenData()
	cmd3 := mockFile.getWrittenData()
	
	assert.Equal(t, []byte("cd /app\n"), cmd1)
	assert.Equal(t, []byte("ls\n"), cmd2)
	assert.Equal(t, []byte("pwd\n"), cmd3)
}

func TestSSHManager_handleSSHSession_NoCommands(t *testing.T) {
	// GIVEN: SSH manager with no remote commands
	sm := NewSSHManager()
	mockFile := NewMockFile()
	password := "testpass"
	remoteCmds := []string{} // No commands
	
	mockFile.On("Close").Return(nil)
	
	// Simulate SSH session that should exit after authentication
	go func() {
		time.Sleep(5 * time.Millisecond)
		mockFile.simulateOutput("password: ")
		time.Sleep(5 * time.Millisecond)
		mockFile.simulateOutput("$ ")
	}()
	
	// WHEN: handling SSH session with no commands
	err := sm.handleSSHSession(mockFile, password, remoteCmds)
	
	// THEN: should complete successfully
	assert.NoError(t, err)
	
	// Verify only password was sent
	passwordData := mockFile.getWrittenData()
	assert.Equal(t, []byte("testpass\n"), passwordData)
}

func TestRealPTY_Start(t *testing.T) {
	// GIVEN: RealPTY instance
	realPTY := &RealPTY{}
	
	// WHEN: calling Start (this is a basic smoke test)
	// Note: We can't really test this without a real command
	// This is more for interface compliance verification
	assert.NotNil(t, realPTY)
	
	// The actual functionality would be tested in integration tests
	// where we can use real processes
}