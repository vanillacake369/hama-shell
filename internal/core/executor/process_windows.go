//go:build windows

package executor

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

// windowsProcessManager implements processManager for Windows systems
type windowsProcessManager struct{}

// newProcessManager creates a new platform-specific process manager
func newProcessManager() processManager {
	return &windowsProcessManager{}
}

// setupCommand configures Windows-specific command settings
func (m *windowsProcessManager) setupCommand(cmd *exec.Cmd) {
	// Create new process group to enable killing the entire process tree
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}
}

// terminateProcess terminates a process on Windows
func (m *windowsProcessManager) terminateProcess(process *os.Process) error {
	if process == nil {
		return nil
	}

	// Windows doesn't have SIGTERM, so we use Kill directly
	if err := process.Kill(); err != nil {
		if err == os.ErrProcessDone {
			return nil
		}
		return fmt.Errorf("failed to kill process: %w", err)
	}

	// Wait for process to be cleaned up
	process.Wait()
	return nil
}
