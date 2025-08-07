package executor

import (
	"os"
	"os/exec"
	"time"
)

// ProcessCommand represents a running command with its metadata
type ProcessCommand struct {
	Cmd       string
	Process   *os.Process
	StartTime time.Time
	Key       string
}

// processManager defines platform-specific process management operations
type processManager interface {
	// setupCommand configures platform-specific settings for the command
	setupCommand(cmd *exec.Cmd)

	// terminateProcess gracefully terminates a process
	terminateProcess(process *os.Process) error
}
