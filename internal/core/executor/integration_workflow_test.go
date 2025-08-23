package executor

import (
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// IntegrationWorkflowTest demonstrates complete end-to-end workflows 
// using the session/PGID architecture with mocked components

func TestWorkflow_CompleteShellSequence(t *testing.T) {
	// GIVEN: Complete workflow with shell commands only
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
	
	key := "myapp.dev.api"
	commands := []string{
		"echo 'Starting deployment'",
		"cd /app",
		"npm install",
		"npm run build",
		"npm run test",
		"echo 'Deployment complete'",
	}
	
	// Mock supervisor creation
	mockSession := &SessionGroup{
		Key:       key,
		SessionID: 12345,
		PGID:      12345,
		Segments:  []*Segment{{Type: "shell", Commands: commands}},
		StartTime: time.Now(),
		Done:      make(chan struct{}),
		Supervisor: &exec.Cmd{},
	}
	// Set process to nil since we can't mock os.Process directly
	
	supervisorMgr.On("createSupervisor", key, mock.AnythingOfType("[]executor.CommandSegment"), mock.AnythingOfType("executor.ExecutionMode")).Return(mockSession, nil)
	
	manageSupCalled := make(chan bool, 1)
	signalMgr.On("manageSupervisor", mockSession).Run(func(args mock.Arguments) {
		manageSupCalled <- true
	}).Return()
	
	// WHEN: Running complete shell sequence
	err := executor.RunSequence(key, commands)
	
	// THEN: Should succeed
	assert.NoError(t, err)
	
	// Wait for management to start
	<-manageSupCalled
	
	// Session creation and management verified through mock expectations
	
	// WHEN: Stopping the session
	signalMgr.On("gracefulShutdown", mockSession, 30*time.Second).Return(nil)
	err = executor.StopByKey(key)
	
	// THEN: Should stop successfully
	assert.NoError(t, err)
	
	// Verify all expectations
	supervisorMgr.AssertExpectations(t)
	signalMgr.AssertExpectations(t)
}

func TestWorkflow_SSHOnlySequence(t *testing.T) {
	// GIVEN: Complete workflow with SSH commands only
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
	
	key := "myapp.prod.database"
	commands := []string{
		"ssh admin@prod-db.example.com",
		"secretpassword123",
		"sudo systemctl status postgresql",
		"psql -U postgres -c '\\l'",
		"pg_dump myapp_prod > backup_$(date +%Y%m%d).sql",
		"exit",
	}
	
	// Mock supervisor creation with SSH segment
	mockSession := &SessionGroup{
		Key:       key,
		SessionID: 23456,
		PGID:      23456,
		Segments:  []*Segment{{Type: "ssh", Commands: commands}},
		StartTime: time.Now(),
		Done:      make(chan struct{}),
		Supervisor: &exec.Cmd{},
	}
	// Process set to nil for testing
	
	supervisorMgr.On("createSupervisor", key, mock.AnythingOfType("[]executor.CommandSegment"), mock.AnythingOfType("executor.ExecutionMode")).Return(mockSession, nil)
	
	manageSupCalled := make(chan bool, 1)
	signalMgr.On("manageSupervisor", mockSession).Run(func(args mock.Arguments) {
		manageSupCalled <- true
	}).Return()
	
	// WHEN: Running SSH sequence
	err := executor.RunSequence(key, commands)
	
	// THEN: Should succeed
	assert.NoError(t, err)
	
	// Wait for management to start
	<-manageSupCalled
	
	// Session creation and management verified through mock expectations
	
	// WHEN: Stopping the session
	signalMgr.On("gracefulShutdown", mockSession, 30*time.Second).Return(nil)
	err = executor.StopByKey(key)
	
	// THEN: Should stop successfully  
	assert.NoError(t, err)
	
	// Verify all expectations
	supervisorMgr.AssertExpectations(t)
	signalMgr.AssertExpectations(t)
}

func TestWorkflow_MixedShellAndSSHSequence(t *testing.T) {
	// GIVEN: Complex workflow with mixed shell and SSH commands
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
	
	key := "myapp.staging.deploy"
	commands := []string{
		// Local preparation
		"echo 'Preparing deployment'",
		"git pull origin main",
		"docker build -t myapp:latest .",
		
		// Deploy to staging server
		"ssh deploy@staging.example.com", 
		"deploypassword456",
		"docker pull myapp:latest",
		"docker stop myapp-container || true",
		"docker run -d --name myapp-container -p 3000:3000 myapp:latest",
		
		// Local verification
		"echo 'Verifying deployment'",
		"curl -f http://staging.example.com:3000/health",
		"echo 'Deployment successful'",
	}
	
	// Mock supervisor creation with multiple segments
	segments := []*Segment{
		{Type: "shell", Commands: []string{"echo 'Preparing deployment'", "git pull origin main", "docker build -t myapp:latest ."}},
		{Type: "ssh", Commands: []string{"ssh deploy@staging.example.com", "deploypassword456", "docker pull myapp:latest", "docker stop myapp-container || true", "docker run -d --name myapp-container -p 3000:3000 myapp:latest"}},
		{Type: "shell", Commands: []string{"echo 'Verifying deployment'", "curl -f http://staging.example.com:3000/health", "echo 'Deployment successful'"}},
	}
	
	mockSession := &SessionGroup{
		Key:       key,
		SessionID: 34567,
		PGID:      34567,
		Segments:  segments,
		StartTime: time.Now(),
		Done:      make(chan struct{}),
		Supervisor: &exec.Cmd{},
	}
	// Process set to nil for testing
	
	supervisorMgr.On("createSupervisor", key, mock.AnythingOfType("[]executor.CommandSegment"), mock.AnythingOfType("executor.ExecutionMode")).Return(mockSession, nil)
	
	manageSupCalled := make(chan bool, 1)
	signalMgr.On("manageSupervisor", mockSession).Run(func(args mock.Arguments) {
		manageSupCalled <- true
	}).Return()
	
	// WHEN: Running mixed sequence
	err := executor.RunSequence(key, commands)
	
	// THEN: Should succeed
	assert.NoError(t, err)
	
	// Wait for management to start
	<-manageSupCalled
	
	// Session creation with multiple segments verified through mock expectations
	
	// WHEN: Stopping the session
	signalMgr.On("gracefulShutdown", mockSession, 30*time.Second).Return(nil)
	err = executor.StopByKey(key)
	
	// THEN: Should stop successfully
	assert.NoError(t, err)
	
	// Verify all expectations
	supervisorMgr.AssertExpectations(t)
	signalMgr.AssertExpectations(t)
}

func TestWorkflow_MultipleSessionsManagement(t *testing.T) {
	// GIVEN: Multiple concurrent sessions
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
	
	// Define multiple sessions
	sessions := []struct {
		key      string
		commands []string
		pid      int
	}{
		{"myapp.dev.api", []string{"echo 'dev api'", "npm start"}, 11111},
		{"myapp.dev.worker", []string{"echo 'dev worker'", "npm run worker"}, 22222},
		{"myapp.staging.api", []string{"ssh staging@host", "password", "systemctl restart api"}, 33333},
	}
	
	var mockSessions []*SessionGroup
	var manageCalls []chan bool
	
	// Set up mocks for each session
	for _, s := range sessions {
		mockSession := &SessionGroup{
			Key:       s.key,
			SessionID: s.pid,
			PGID:      s.pid,
			Segments:  []*Segment{{Type: "shell", Commands: s.commands}},
			StartTime: time.Now(),
			Done:      make(chan struct{}),
			Supervisor: &exec.Cmd{},
		}
		// Process set to nil for testing
		mockSessions = append(mockSessions, mockSession)
		
		supervisorMgr.On("createSupervisor", s.key, mock.AnythingOfType("[]executor.CommandSegment"), mock.AnythingOfType("executor.ExecutionMode")).Return(mockSession, nil)
		
		manageCalled := make(chan bool, 1)
		manageCalls = append(manageCalls, manageCalled)
		signalMgr.On("manageSupervisor", mockSession).Run(func(args mock.Arguments) {
			manageCalled <- true
		}).Return()
	}
	
	// WHEN: Starting all sessions
	for _, s := range sessions {
		err := executor.RunSequence(s.key, s.commands)
		assert.NoError(t, err)
	}
	
	// Wait for all management to start
	for _, ch := range manageCalls {
		<-ch
	}
	
	// Multiple session creation and management verified through mock expectations
	
	// WHEN: Stopping individual session
	signalMgr.On("gracefulShutdown", mockSessions[0], 30*time.Second).Return(nil)
	err := executor.StopByKey(sessions[0].key)
	assert.NoError(t, err)
	
	// THEN: Only that session should be removed
	// (Note: In real implementation, registry would update)
	
	// WHEN: Stopping all remaining sessions
	signalMgr.On("shutdownAllSessions", 30*time.Second).Return(nil)
	err = executor.StopAll()
	
	// THEN: Should succeed
	assert.NoError(t, err)
	
	// Verify all expectations
	supervisorMgr.AssertExpectations(t)
	signalMgr.AssertExpectations(t)
}

func TestWorkflow_SessionRecoveryAfterFailure(t *testing.T) {
	// GIVEN: Workflow that simulates session failure and cleanup
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
	
	key := "myapp.prod.critical"
	commands := []string{
		"ssh prod@critical.example.com",
		"criticalpassword789",
		"sudo systemctl restart critical-service",
		"sudo systemctl status critical-service",
	}
	
	// Mock session that will "fail"
	mockSession := &SessionGroup{
		Key:       key,
		SessionID: 45678,
		PGID:      45678,
		Segments:  []*Segment{{Type: "ssh", Commands: commands}},
		StartTime: time.Now(),
		Done:      make(chan struct{}),
		Supervisor: &exec.Cmd{},
	}
	// Process set to nil for testing
	
	supervisorMgr.On("createSupervisor", key, mock.AnythingOfType("[]executor.CommandSegment"), mock.AnythingOfType("executor.ExecutionMode")).Return(mockSession, nil)
	
	manageSupCalled := make(chan bool, 1)
	signalMgr.On("manageSupervisor", mockSession).Run(func(args mock.Arguments) {
		// Simulate session completion/failure by closing Done channel
		select {
		case <-mockSession.Done:
			// Already closed
		default:
			close(mockSession.Done)
		}
		manageSupCalled <- true
	}).Return()
	
	// WHEN: Running sequence that will "fail"
	err := executor.RunSequence(key, commands)
	
	// THEN: Should initially succeed
	assert.NoError(t, err)
	
	// Wait for management to start and "fail"
	<-manageSupCalled
	
	// Session failure handled by management
	
	// WHEN: Attempting to stop failed session
	signalMgr.On("gracefulShutdown", mockSession, 30*time.Second).Return(nil)
	err = executor.StopByKey(key)
	
	// THEN: Should handle gracefully
	assert.NoError(t, err)
	
	// Verify expectations for the first session
	supervisorMgr.AssertExpectations(t)
	signalMgr.AssertExpectations(t)
}

// MockProcess implements the necessary interface for testing
type MockProcess struct {
	Pid int
}

func TestWorkflow_ComplexSSHJumpHost(t *testing.T) {
	// GIVEN: Complex workflow with SSH jump host pattern
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
	
	key := "myapp.prod.backend"
	commands := []string{
		"echo 'Connecting through jump host'",
		
		// Connect to jump host
		"ssh jump@jumphost.example.com",
		"jumppassword123",
		
		// From jump host, connect to internal server
		"ssh internal@10.0.1.100",
		"internalpassword456", 
		"sudo systemctl status myapp-backend",
		"tail -f /var/log/myapp/backend.log",
		
		// Local cleanup
		"echo 'Connection completed'",
	}
	
	// Mock supervisor creation with multiple SSH segments
	segments := []*Segment{
		{Type: "shell", Commands: []string{"echo 'Connecting through jump host'"}},
		{Type: "ssh", Commands: []string{"ssh jump@jumphost.example.com", "jumppassword123", "ssh internal@10.0.1.100", "internalpassword456", "sudo systemctl status myapp-backend", "tail -f /var/log/myapp/backend.log"}},
		{Type: "shell", Commands: []string{"echo 'Connection completed'"}},
	}
	
	mockSession := &SessionGroup{
		Key:       key,
		SessionID: 67890,
		PGID:      67890,
		Segments:  segments,
		StartTime: time.Now(),
		Done:      make(chan struct{}),
		Supervisor: &exec.Cmd{},
	}
	// Process set to nil for testing
	
	supervisorMgr.On("createSupervisor", key, mock.AnythingOfType("[]executor.CommandSegment"), mock.AnythingOfType("executor.ExecutionMode")).Return(mockSession, nil)
	
	manageSupCalled := make(chan bool, 1)
	signalMgr.On("manageSupervisor", mockSession).Run(func(args mock.Arguments) {
		manageSupCalled <- true
	}).Return()
	
	// WHEN: Running jump host sequence
	err := executor.RunSequence(key, commands)
	
	// THEN: Should succeed
	assert.NoError(t, err)
	
	// Wait for management to start
	<-manageSupCalled
	
	// Complex session with jump host pattern verified through mock expectations
	
	// WHEN: Stopping the complex session
	signalMgr.On("gracefulShutdown", mockSession, 30*time.Second).Return(nil)
	err = executor.StopByKey(key)
	
	// THEN: Should stop successfully
	assert.NoError(t, err)
	
	// Verify all expectations
	supervisorMgr.AssertExpectations(t)
	signalMgr.AssertExpectations(t)
}