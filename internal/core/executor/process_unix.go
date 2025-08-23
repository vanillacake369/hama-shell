//go:build !windows

package executor

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"syscall"
	"time"
)

// unixProcessManager implements processManager for Unix-like systems
type unixProcessManager struct{}

// platformCapabilities describes what process management features are supported
type platformCapabilities struct {
	supportsSessionCreation bool
	supportsProcessGroup    bool
	supportsCombinedSetup   bool
	platformName           string
}

// newProcessManager creates a new platform-specific process manager
func newProcessManager() processManager {
	return &unixProcessManager{}
}

// getPlatformCapabilities returns the capabilities of the current platform
func (m *unixProcessManager) getPlatformCapabilities() platformCapabilities {
	switch runtime.GOOS {
	case "darwin":
		// macOS: Session creation works, but Setsid+Setpgid combination fails
		return platformCapabilities{
			supportsSessionCreation: true,
			supportsProcessGroup:    true,
			supportsCombinedSetup:   false, // Key limitation on macOS
			platformName:           "macOS/Darwin",
		}
	case "linux":
		// Linux: Full session and process group support
		return platformCapabilities{
			supportsSessionCreation: true,
			supportsProcessGroup:    true,
			supportsCombinedSetup:   true,
			platformName:           "Linux",
		}
	default:
		// Other Unix systems: Conservative approach
		return platformCapabilities{
			supportsSessionCreation: true,
			supportsProcessGroup:    true,
			supportsCombinedSetup:   true,
			platformName:           runtime.GOOS,
		}
	}
}

// getPlatformSpecificSysProcAttr returns OS-appropriate SysProcAttr configuration
func (m *unixProcessManager) getPlatformSpecificSysProcAttr() *syscall.SysProcAttr {
	capabilities := m.getPlatformCapabilities()
	
	if capabilities.supportsCombinedSetup {
		// Full session + process group setup (Linux and other Unix)
		return &syscall.SysProcAttr{
			Setsid:  true, // Create new session
			Setpgid: true, // Create new process group
			Pgid:    0,    // PGID = PID (become group leader)
		}
	} else {
		// Session-only setup (macOS) - automatically becomes process group leader
		return &syscall.SysProcAttr{
			Setsid: true, // Create new session (sufficient on macOS)
			// Note: When Setsid=true, the process automatically becomes process group leader
			// Setting Setpgid=true causes "operation not permitted" on macOS
		}
	}
}

// setupCommand configures Unix-specific command settings
func (m *unixProcessManager) setupCommand(cmd *exec.Cmd) {
	// Set process group ID to enable killing the entire process tree
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
}

// setupSupervisor configures supervisor with session/PGID settings
func (m *unixProcessManager) setupSupervisor(cmd *exec.Cmd) error {
	if cmd == nil {
		return fmt.Errorf("command cannot be nil")
	}
	// Use platform-specific configuration to avoid macOS restrictions
	cmd.SysProcAttr = m.getPlatformSpecificSysProcAttr()
	return nil
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
