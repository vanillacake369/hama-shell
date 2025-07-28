//go:build !windows

package executor

import (
	"os/exec"
	"syscall"
)

// setupCommand configures Unix-specific command settings
func (ce *commandExecutor) setupCommand(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

// terminateProcess terminates a process using Unix signals
func (ce *commandExecutor) terminateProcess(cmd *exec.Cmd) error {
	if cmd.Process != nil {
		return cmd.Process.Signal(syscall.SIGTERM)
	}
	return nil
}
