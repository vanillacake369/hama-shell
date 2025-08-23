package executor

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSupervisorManager_createSupervisor_Success(t *testing.T) {
	// GIVEN: supervisor manager with mock process manager
	mockPM := &MockProcessManager{}
	sm := &SupervisorManager{
		parser:  NewCommandParser(),
		manager: mockPM,
	}
	segments := []CommandSegment{
		{Type: "shell", Commands: []string{"echo test"}},
	}

	mockPM.On("setupSupervisor", mock.AnythingOfType("*exec.Cmd")).Return(nil)

	// WHEN: creating supervisor
	session, err := sm.createSupervisor("test-key", segments)

	// THEN: supervisor should be created successfully
	assert.NoError(t, err)
	assert.NotNil(t, session)
	assert.Equal(t, "test-key", session.Key)
	assert.NotNil(t, session.Supervisor)
	assert.Len(t, session.Segments, 1)
	assert.Equal(t, "shell", session.Segments[0].Type)
	mockPM.AssertExpectations(t)
}

func TestSupervisorManager_createSupervisor_NoSegments(t *testing.T) {
	// GIVEN: supervisor manager with empty segments
	sm := NewSupervisorManager()
	segments := []CommandSegment{}

	// WHEN: creating supervisor with empty segments
	session, err := sm.createSupervisor("test-key", segments)

	// THEN: should return error
	assert.Error(t, err)
	assert.Nil(t, session)
	assert.Contains(t, err.Error(), "no command segments provided")
}

func TestSupervisorManager_createSupervisor_SetupError(t *testing.T) {
	// GIVEN: supervisor manager with mock that returns setup error
	mockPM := &MockProcessManager{}
	sm := &SupervisorManager{
		parser:  NewCommandParser(),
		manager: mockPM,
	}
	segments := []CommandSegment{
		{Type: "shell", Commands: []string{"echo test"}},
	}

	mockPM.On("setupSupervisor", mock.AnythingOfType("*exec.Cmd")).Return(assert.AnError)

	// WHEN: creating supervisor
	session, err := sm.createSupervisor("test-key", segments)

	// THEN: should return setup error
	assert.Error(t, err)
	assert.Nil(t, session)
	assert.Contains(t, err.Error(), "failed to setup supervisor")
	mockPM.AssertExpectations(t)
}

func TestSupervisorManager_buildSupervisorScript_ShellSegment(t *testing.T) {
	// GIVEN: supervisor manager and shell segment
	sm := NewSupervisorManager()
	segments := []CommandSegment{
		{Type: "shell", Commands: []string{"cd /app", "ls -la"}},
	}

	// WHEN: building supervisor script
	script := sm.buildSupervisorScript(segments)

	// THEN: script should contain shell commands joined with &&
	assert.Contains(t, script, "#!/bin/bash")
	assert.Contains(t, script, "set -e")
	assert.Contains(t, script, "trap 'kill -TERM -$$; exit' INT TERM EXIT")
	assert.Contains(t, script, "(cd /app && ls -la) &")
	assert.Contains(t, script, "SHELL_PID_0=$!")
	assert.Contains(t, script, "wait")
}

func TestSupervisorManager_buildSupervisorScript_SSHSegment(t *testing.T) {
	// GIVEN: supervisor manager and SSH segment with password
	sm := NewSupervisorManager()
	segments := []CommandSegment{
		{Type: "ssh", Commands: []string{"ssh user@host", "password123", "ls", "pwd"}},
	}

	// WHEN: building supervisor script
	script := sm.buildSupervisorScript(segments)

	// THEN: script should contain SSH helper call
	assert.Contains(t, script, "#!/bin/bash")
	assert.Contains(t, script, "hama_ssh_helper 'ssh user@host' 'password123' 'ls' 'pwd' &")
	assert.Contains(t, script, "SSH_PID_0=$!")
	assert.Contains(t, script, "wait")
}

func TestSupervisorManager_buildSupervisorScript_SSHWithoutPassword(t *testing.T) {
	// GIVEN: supervisor manager and SSH segment without password
	sm := NewSupervisorManager()
	segments := []CommandSegment{
		{Type: "ssh", Commands: []string{"ssh user@host", "ls -la"}},
	}

	// WHEN: building supervisor script
	script := sm.buildSupervisorScript(segments)

	// THEN: script should execute SSH directly
	assert.Contains(t, script, "#!/bin/bash")
	assert.Contains(t, script, "(ssh user@host; ls -la) &")
	assert.Contains(t, script, "SSH_PID_0=$!")
}

func TestSupervisorManager_buildSupervisorScript_MixedSegments(t *testing.T) {
	// GIVEN: supervisor manager with mixed segments
	sm := NewSupervisorManager()
	segments := []CommandSegment{
		{Type: "shell", Commands: []string{"echo start"}},
		{Type: "ssh", Commands: []string{"ssh user@host", "pass123", "ls"}},
		{Type: "shell", Commands: []string{"echo end"}},
	}

	// WHEN: building supervisor script
	script := sm.buildSupervisorScript(segments)

	// THEN: script should contain all segments properly formatted
	assert.Contains(t, script, "#!/bin/bash")
	assert.Contains(t, script, "set -e")
	assert.Contains(t, script, "trap 'kill -TERM -$$; exit' INT TERM EXIT")
	
	// Shell segment 0
	assert.Contains(t, script, "# Shell Segment 0")
	assert.Contains(t, script, "(echo start) &")
	assert.Contains(t, script, "SHELL_PID_0=$!")
	
	// SSH segment 1
	assert.Contains(t, script, "# SSH Segment 1")
	assert.Contains(t, script, "hama_ssh_helper 'ssh user@host' 'pass123' 'ls' &")
	assert.Contains(t, script, "SSH_PID_1=$!")
	
	// Shell segment 2
	assert.Contains(t, script, "# Shell Segment 2")
	assert.Contains(t, script, "(echo end) &")
	assert.Contains(t, script, "SHELL_PID_2=$!")
	
	assert.Contains(t, script, "wait")
}

func TestSupervisorManager_buildSupervisorScript_EmptySegments(t *testing.T) {
	// GIVEN: supervisor manager with segments containing empty commands
	sm := NewSupervisorManager()
	segments := []CommandSegment{
		{Type: "shell", Commands: []string{}},
		{Type: "ssh", Commands: []string{}},
	}

	// WHEN: building supervisor script
	script := sm.buildSupervisorScript(segments)

	// THEN: script should be created but segments should be skipped
	assert.Contains(t, script, "#!/bin/bash")
	assert.Contains(t, script, "wait")
	
	// Should not contain segment-specific content since commands are empty
	assert.NotContains(t, script, "SHELL_PID_")
	assert.NotContains(t, script, "SSH_PID_")
}

func TestSupervisorManager_buildSupervisorScript_UnknownSegmentType(t *testing.T) {
	// GIVEN: supervisor manager with unknown segment type
	sm := NewSupervisorManager()
	segments := []CommandSegment{
		{Type: "unknown", Commands: []string{"some command"}},
	}

	// WHEN: building supervisor script
	script := sm.buildSupervisorScript(segments)

	// THEN: script should handle unknown type gracefully
	assert.Contains(t, script, "#!/bin/bash")
	assert.Contains(t, script, "echo 'Unknown segment type: unknown'")
	assert.Contains(t, script, "wait")
}

func TestSupervisorManager_addSSHSegmentToScript(t *testing.T) {
	sm := NewSupervisorManager()
	
	tests := []struct {
		name     string
		segment  CommandSegment
		expected []string
	}{
		{
			name: "SSH with password and remote commands",
			segment: CommandSegment{
				Type: "ssh",
				Commands: []string{"ssh user@host", "password123", "cd /app", "ls"},
			},
			expected: []string{
				"# SSH Segment 0",
				"hama_ssh_helper 'ssh user@host' 'password123' 'cd /app' 'ls' &",
				"SSH_PID_0=$!",
			},
		},
		{
			name: "SSH with just password",
			segment: CommandSegment{
				Type: "ssh", 
				Commands: []string{"ssh user@host", "password123"},
			},
			expected: []string{
				"# SSH Segment 1",
				"hama_ssh_helper 'ssh user@host' 'password123' &",
				"SSH_PID_1=$!",
			},
		},
		{
			name: "SSH without password",
			segment: CommandSegment{
				Type: "ssh",
				Commands: []string{"ssh user@host", "ls -la"},
			},
			expected: []string{
				"# SSH Segment 2", 
				"(ssh user@host; ls -la) &",
				"SSH_PID_2=$!",
			},
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// GIVEN: SSH segment
			var script strings.Builder
			
			// WHEN: adding SSH segment to script
			sm.addSSHSegmentToScript(&script, i, tt.segment)
			result := script.String()
			
			// THEN: script should contain expected content
			for _, expected := range tt.expected {
				assert.Contains(t, result, expected, "Missing expected content: %s", expected)
			}
		})
	}
}

func TestSupervisorManager_addShellSegmentToScript(t *testing.T) {
	// GIVEN: supervisor manager and shell segment
	sm := NewSupervisorManager()
	segment := CommandSegment{
		Type: "shell",
		Commands: []string{"cd /app", "npm install", "npm start"},
	}
	var script strings.Builder

	// WHEN: adding shell segment to script
	sm.addShellSegmentToScript(&script, 0, segment)
	result := script.String()

	// THEN: script should contain shell commands joined with &&
	assert.Contains(t, result, "# Shell Segment 0")
	assert.Contains(t, result, "(cd /app && npm install && npm start) &")
	assert.Contains(t, result, "SHELL_PID_0=$!")
}

func TestSupervisorManager_addShellSegmentToScript_EmptyCommands(t *testing.T) {
	// GIVEN: supervisor manager and empty shell segment
	sm := NewSupervisorManager()
	segment := CommandSegment{
		Type: "shell",
		Commands: []string{},
	}
	var script strings.Builder

	// WHEN: adding empty shell segment to script
	sm.addShellSegmentToScript(&script, 0, segment)
	result := script.String()

	// THEN: script should be empty (segment skipped)
	assert.Empty(t, result)
}

func TestConvertToSegments(t *testing.T) {
	// GIVEN: command segments
	cmdSegments := []CommandSegment{
		{Type: "shell", Commands: []string{"echo test"}},
		{Type: "ssh", Commands: []string{"ssh user@host", "password"}},
	}

	// WHEN: converting to segments
	segments := convertToSegments(cmdSegments)

	// THEN: should create proper Segment structs
	assert.Len(t, segments, 2)
	
	assert.Equal(t, "shell", segments[0].Type)
	assert.Equal(t, []string{"echo test"}, segments[0].Commands)
	assert.Nil(t, segments[0].Process)
	assert.Nil(t, segments[0].PTY)
	
	assert.Equal(t, "ssh", segments[1].Type)
	assert.Equal(t, []string{"ssh user@host", "password"}, segments[1].Commands)
	assert.Nil(t, segments[1].Process)
	assert.Nil(t, segments[1].PTY)
}