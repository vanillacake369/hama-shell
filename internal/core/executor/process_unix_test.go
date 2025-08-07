//go:build !windows

package executor

import (
	"os/exec"
	"syscall"
	"testing"
	"time"
)

func TestUnixProcessManager_Success(t *testing.T) {
	// Test creating a Unix process manager
	manager := newProcessManager()
	if manager == nil {
		t.Fatal("Expected process manager to be created, got nil")
	}

	// Verify it's the correct type
	if _, ok := manager.(*unixProcessManager); !ok {
		t.Fatal("Expected unixProcessManager, got different type")
	}

	// Test setupCommand with a simple Unix command
	cmd := exec.Command("sleep", "10")

	// Setup command with Unix-specific settings
	manager.setupCommand(cmd)

	// Verify the command has correct Unix settings
	if cmd.SysProcAttr == nil {
		t.Fatal("Expected SysProcAttr to be set, got nil")
	}

	if !cmd.SysProcAttr.Setpgid {
		t.Error("Expected Setpgid to be true, got false")
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start test command: %v", err)
	}

	// Give it a moment to start properly
	time.Sleep(100 * time.Millisecond)

	// Test graceful termination with SIGTERM
	err := manager.terminateProcess(cmd.Process)
	if err != nil {
		t.Errorf("Expected terminateProcess to succeed, got error: %v", err)
	}

	// Verify process is actually terminated by waiting for it
	// This should complete quickly since we terminated it
	done := make(chan bool, 1)
	go func() {
		cmd.Wait()
		done <- true
	}()

	select {
	case <-done:
		// Process terminated successfully
	case <-time.After(3 * time.Second):
		t.Error("Process did not terminate within expected time")
	}

	// Test terminateProcess with nil process (should not error)
	err = manager.terminateProcess(nil)
	if err != nil {
		t.Errorf("Expected terminateProcess with nil to succeed, got error: %v", err)
	}

	// Test forceKill functionality by starting another process
	cmd2 := exec.Command("sleep", "10")
	manager.setupCommand(cmd2)

	if err := cmd2.Start(); err != nil {
		t.Fatalf("Failed to start second test command: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Test the forceKill method directly
	unixManager := manager.(*unixProcessManager)
	err = unixManager.forceKill(cmd2.Process)
	if err != nil {
		t.Errorf("Expected forceKill to succeed, got error: %v", err)
	}

	// Verify second process is terminated
	done2 := make(chan bool, 1)
	go func() {
		cmd2.Wait()
		done2 <- true
	}()

	select {
	case <-done2:
		// Process force killed successfully
	case <-time.After(3 * time.Second):
		t.Error("Process did not force kill within expected time")
	}
}
