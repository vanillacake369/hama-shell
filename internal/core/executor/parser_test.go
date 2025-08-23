package executor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommandParser_parseCommandSegments_EmptyCommands(t *testing.T) {
	// GIVEN: empty command list
	parser := NewCommandParser()
	commands := []string{}

	// WHEN: parsing segments
	segments := parser.parseCommandSegments(commands)

	// THEN: should return empty segments
	assert.Nil(t, segments)
}

func TestCommandParser_parseCommandSegments_SSHOnly(t *testing.T) {
	// GIVEN: commands with single SSH segment
	parser := NewCommandParser()
	commands := []string{"ssh user@host", "password123", "ls -la", "pwd"}

	// WHEN: parsing segments
	segments := parser.parseCommandSegments(commands)

	// THEN: should have one SSH segment
	assert.Len(t, segments, 1)
	assert.Equal(t, "ssh", segments[0].Type)
	assert.Equal(t, "ssh user@host", segments[0].Commands[0])
	assert.Equal(t, "password123", segments[0].Commands[1])
	assert.Equal(t, "ls -la", segments[0].Commands[2])
	assert.Equal(t, "pwd", segments[0].Commands[3])
}

func TestCommandParser_parseCommandSegments_ShellOnly(t *testing.T) {
	// GIVEN: commands with only shell commands
	parser := NewCommandParser()
	commands := []string{"cd /app", "ls -la", "echo hello"}

	// WHEN: parsing segments
	segments := parser.parseCommandSegments(commands)

	// THEN: should have one shell segment
	assert.Len(t, segments, 1)
	assert.Equal(t, "shell", segments[0].Type)
	assert.Equal(t, []string{"cd /app", "ls -la", "echo hello"}, segments[0].Commands)
}

func TestCommandParser_parseCommandSegments_MixedCommands(t *testing.T) {
	// GIVEN: mixed SSH and shell commands
	parser := NewCommandParser()
	commands := []string{
		"ssh user@host1", "pass1", "cd /app", "ls",
		"echo local", "pwd",
		"ssh user@host2", "pass2", "tail -f log",
	}

	// WHEN: parsing segments
	segments := parser.parseCommandSegments(commands)

	// THEN: should have 3 segments (ssh, shell, ssh)
	assert.Len(t, segments, 3)

	// First SSH segment
	assert.Equal(t, "ssh", segments[0].Type)
	assert.Equal(t, []string{"ssh user@host1", "pass1", "cd /app", "ls"}, segments[0].Commands)

	// Shell segment
	assert.Equal(t, "shell", segments[1].Type)
	assert.Equal(t, []string{"echo local", "pwd"}, segments[1].Commands)

	// Second SSH segment
	assert.Equal(t, "ssh", segments[2].Type)
	assert.Equal(t, []string{"ssh user@host2", "pass2", "tail -f log"}, segments[2].Commands)
}

func TestCommandParser_parseCommandSegments_MultipleSSH(t *testing.T) {
	// GIVEN: commands with multiple consecutive SSH commands
	parser := NewCommandParser()
	commands := []string{
		"ssh user@host1", "pass1", "echo first",
		"ssh user@host2", "pass2", "echo second",
	}

	// WHEN: parsing segments
	segments := parser.parseCommandSegments(commands)

	// THEN: should have 2 SSH segments
	assert.Len(t, segments, 2)

	assert.Equal(t, "ssh", segments[0].Type)
	assert.Equal(t, []string{"ssh user@host1", "pass1", "echo first"}, segments[0].Commands)

	assert.Equal(t, "ssh", segments[1].Type)
	assert.Equal(t, []string{"ssh user@host2", "pass2", "echo second"}, segments[1].Commands)
}

func TestCommandParser_parseCommandSegments_SSHWithoutPassword(t *testing.T) {
	// GIVEN: SSH command followed by shell command (no password)
	parser := NewCommandParser()
	commands := []string{"ssh user@host", "ls -la", "pwd"}

	// WHEN: parsing segments
	segments := parser.parseCommandSegments(commands)

	// THEN: should have one SSH segment (ls is treated as remote command)
	assert.Len(t, segments, 1)
	assert.Equal(t, "ssh", segments[0].Type)
	assert.Equal(t, []string{"ssh user@host", "ls -la", "pwd"}, segments[0].Commands)
}

func TestCommandParser_parseCommandSegments_ComplexWorkflow(t *testing.T) {
	// GIVEN: complex workflow with multiple transitions
	parser := NewCommandParser()
	commands := []string{
		"echo starting",
		"cd /tmp",
		"ssh user@bastion", "bastionpass", "ssh user@target", "targetpass", "cd /app", "ls",
		"echo back local",
		"pwd",
	}

	// WHEN: parsing segments
	segments := parser.parseCommandSegments(commands)

	// THEN: should have 4 segments (shell, ssh-bastion, ssh-target, shell)
	assert.Len(t, segments, 4)

	// Initial shell commands
	assert.Equal(t, "shell", segments[0].Type)
	assert.Equal(t, []string{"echo starting", "cd /tmp"}, segments[0].Commands)

	// SSH to bastion
	assert.Equal(t, "ssh", segments[1].Type)
	assert.Equal(t, []string{"ssh user@bastion", "bastionpass"}, segments[1].Commands)

	// SSH from bastion to target
	assert.Equal(t, "ssh", segments[2].Type)
	assert.Equal(t, []string{"ssh user@target", "targetpass", "cd /app", "ls"}, segments[2].Commands)

	// Final local commands
	assert.Equal(t, "shell", segments[3].Type)
	assert.Equal(t, []string{"echo back local", "pwd"}, segments[3].Commands)
}

func TestCommandParser_isSSHCommand(t *testing.T) {
	parser := NewCommandParser()

	tests := []struct {
		name     string
		command  string
		expected bool
	}{
		{"basic ssh", "ssh user@host", true},
		{"ssh with options", "ssh -i key user@host", true},
		{"ssh with tab", "ssh\tuser@host", true},
		{"ssh with spaces", "  ssh user@host  ", true},
		{"not ssh", "echo hello", false},
		{"contains ssh but not prefix", "echo ssh command", false},
		{"empty command", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// GIVEN: command string
			// WHEN: checking if it's SSH command
			result := parser.isSSHCommand(tt.command)
			// THEN: should match expected result
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCommandParser_isShellCommand(t *testing.T) {
	parser := NewCommandParser()

	tests := []struct {
		name     string
		command  string
		expected bool
	}{
		{"cd command", "cd /app", true},
		{"ls command", "ls -la", true},
		{"echo command", "echo hello", true},
		{"path command", "/usr/bin/app", true},
		{"relative path", "./script.sh", true},
		{"home path", "~/script.sh", true},
		{"env var", "VAR=value", true},
		{"git command", "git status", true},
		{"docker command", "docker ps", true},
		{"random string", "randompassword123", false},
		{"number", "12345", false},
		{"ip address", "192.168.1.1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// GIVEN: command string
			// WHEN: checking if it's shell command
			result := parser.isShellCommand(tt.command)
			// THEN: should match expected result
			assert.Equal(t, tt.expected, result, "Command: %s", tt.command)
		})
	}
}

func TestCommandParser_parseCommandSegments_EdgeCases(t *testing.T) {
	parser := NewCommandParser()

	t.Run("single ssh command", func(t *testing.T) {
		// GIVEN: just SSH command
		commands := []string{"ssh user@host"}

		// WHEN: parsing segments
		segments := parser.parseCommandSegments(commands)

		// THEN: should have one SSH segment
		assert.Len(t, segments, 1)
		assert.Equal(t, "ssh", segments[0].Type)
		assert.Equal(t, []string{"ssh user@host"}, segments[0].Commands)
	})

	t.Run("ssh with shell command as password", func(t *testing.T) {
		// GIVEN: SSH followed by shell command that looks like password
		commands := []string{"ssh user@host", "cd /app"}

		// WHEN: parsing segments
		segments := parser.parseCommandSegments(commands)

		// THEN: should treat cd as remote command, not password
		assert.Len(t, segments, 1)
		assert.Equal(t, "ssh", segments[0].Type)
		assert.Equal(t, []string{"ssh user@host", "cd /app"}, segments[0].Commands)
	})

	t.Run("whitespace handling", func(t *testing.T) {
		// GIVEN: commands with extra whitespace
		commands := []string{"  ssh user@host  ", "  password123  ", "  ls -la  "}

		// WHEN: parsing segments
		segments := parser.parseCommandSegments(commands)

		// THEN: should handle whitespace correctly
		assert.Len(t, segments, 1)
		assert.Equal(t, "ssh", segments[0].Type)
		assert.Equal(t, []string{"  ssh user@host  ", "  password123  ", "  ls -la  "}, segments[0].Commands)
	})
}