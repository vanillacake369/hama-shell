//go:build windows

package executor

import (
	"os/exec"
	"syscall"
)

// setupCommand configures Windows-specific command settings
func (ce *commandExecutor) setupCommand(cmd *exec.Cmd) {
	// Windows doesn't support Setpgid, so we use CREATE_NEW_PROCESS_GROUP instead
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}
}

// terminateProcess terminates a process using Windows-appropriate method
func (ce *commandExecutor) terminateProcess(cmd *exec.Cmd) error {
	if cmd.Process != nil {
		return cmd.Process.Kill()
	}
	return nil
}
