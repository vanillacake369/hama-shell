//go:build windows

package executor

import (
	"os/exec"
	"syscall"
	"testing"
	"time"
)

func TestWindowsProcessManager_Success(t *testing.T) {
	// Test creating a Windows process manager
	manager := newProcessManager()
	if manager == nil {
		t.Fatal("Expected process manager to be created, got nil")
	}

	// Verify it's the correct type
	if _, ok := manager.(*windowsProcessManager); !ok {
		t.Fatal("Expected windowsProcessManager, got different type")
	}

	// Test setupCommand with a simple Windows command
	cmd := exec.Command("ping", "-n", "2", "127.0.0.1")

	// Setup command with Windows-specific settings
	manager.setupCommand(cmd)

	// Verify the command has correct Windows settings
	if cmd.SysProcAttr == nil {
		t.Fatal("Expected SysProcAttr to be set, got nil")
	}

	if cmd.SysProcAttr.CreationFlags != syscall.CREATE_NEW_PROCESS_GROUP {
		t.Errorf("Expected CreationFlags to be CREATE_NEW_PROCESS_GROUP, got %d", cmd.SysProcAttr.CreationFlags)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start test command: %v", err)
	}

	// Give it a moment to start properly
	time.Sleep(100 * time.Millisecond)

	// Test terminateProcess
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
	case <-time.After(5 * time.Second):
		t.Error("Process did not terminate within expected time")
	}

	// Test terminateProcess with nil process (should not error)
	err = manager.terminateProcess(nil)
	if err != nil {
		t.Errorf("Expected terminateProcess with nil to succeed, got error: %v", err)
	}
}
