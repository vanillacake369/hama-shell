package executor

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSSHExecutorWithPTY(t *testing.T) {
	// GIVEN
	executor := NewSSHExecutor("127.0.0.1", "limjihoon", "1026", 3*time.Second)
	executor.Commands = []string{
		"whoami",
		"pwd",
		"ls -la | head -5",
		"cd ~/dev/",
		"pwd",
		"ls | head -5",
		"echo 'All commands completed'",
	}
	executor.Timeout = 10 * time.Second

	// WHEN
	t.Log("Starting PTY SSH test...")
	err := executor.ExecuteWithPTY()

	// THEN
	assert.NoError(t, err)
}
