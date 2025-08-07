//go:build !windows

package executor

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"
)

// unixProcessManager implements processManager for Unix-like systems
type unixProcessManager struct{}

// newProcessManager creates a new platform-specific process manager
func newProcessManager() processManager {
	return &unixProcessManager{}
}

// setupCommand configures Unix-specific command settings
func (m *unixProcessManager) setupCommand(cmd *exec.Cmd) {
	// Set process group ID to enable killing the entire process tree
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
}

// terminateProcess gracefully terminates a process on Unix
func (m *unixProcessManager) terminateProcess(process *os.Process) error {
	if process == nil {
		return nil
	}

	// First, try graceful termination with SIGTERM
	if err := process.Signal(syscall.SIGTERM); err != nil {
		// If SIGTERM fails, the process might already be dead
		if err == os.ErrProcessDone {
			return nil
		}
		// For other errors, try SIGKILL as fallback
		return m.forceKill(process)
	}

	// Give the process time to clean up
	done := make(chan error, 1)
	go func() {
		_, err := process.Wait()
		done <- err
	}()

	select {
	case <-done:
		// Process terminated gracefully
		return nil
	case <-time.After(5 * time.Second):
		// Timeout - force kill
		return m.forceKill(process)
	}
}

// forceKill forcefully terminates a process using SIGKILL
func (m *unixProcessManager) forceKill(process *os.Process) error {
	if err := process.Signal(syscall.SIGKILL); err != nil {
		if err == os.ErrProcessDone {
			return nil
		}
		return fmt.Errorf("failed to kill process: %w", err)
	}

	// Wait for process to be reaped to avoid zombies
	process.Wait()
	return nil
}
