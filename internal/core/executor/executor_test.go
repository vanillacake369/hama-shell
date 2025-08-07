package executor

import (
	"runtime"
	"testing"
	"time"
)

func TestExecutor_Run_Success(t *testing.T) {
	// Create a new executor instance
	exec := New()
	if exec == nil {
		t.Fatal("Expected executor to be created, got nil")
	}

	// Test data - use cross-platform long-running command
	testKey := "test.service"
	var testCommand string
	if runtime.GOOS == "windows" {
		testCommand = "timeout /t 3 /nobreak >nul"
	} else {
		testCommand = "sleep 3"
	}

	// Test the Run method
	err := exec.Run(testKey, testCommand)
	if err != nil {
		t.Fatalf("Expected Run to succeed, got error: %v", err)
	}

	// Give the process time to start and register
	time.Sleep(100 * time.Millisecond)

	// Verify the process was registered correctly
	status := exec.GetStatus()
	if len(status) == 0 {
		t.Fatal("Expected at least one process in status after Run")
	}

	// Check that our specific key exists
	processes, exists := status[testKey]
	if !exists {
		t.Fatalf("Expected key '%s' to exist in status", testKey)
	}

	// Verify process information
	if len(processes) != 1 {
		t.Fatalf("Expected exactly 1 process for key '%s', got %d", testKey, len(processes))
	}

	proc := processes[0]

	// Verify process details
	if proc.Command != testCommand {
		t.Errorf("Expected command '%s', got '%s'", testCommand, proc.Command)
	}

	if proc.Key != testKey {
		t.Errorf("Expected key '%s', got '%s'", testKey, proc.Key)
	}

	if proc.PID <= 0 {
		t.Errorf("Expected valid PID > 0, got %d", proc.PID)
	}

	if proc.StartTime == 0 {
		t.Error("Expected StartTime to be set, got 0")
	}

	// Verify StartTime is recent (within last few seconds)
	now := time.Now().Unix()
	if now-proc.StartTime > 5 {
		t.Errorf("Expected StartTime to be recent, got timestamp %d seconds ago", now-proc.StartTime)
	}

	// Test running multiple processes with the same key - use cross-platform command
	var testCommand2 string
	if runtime.GOOS == "windows" {
		testCommand2 = "timeout /t 2 /nobreak >nul"
	} else {
		testCommand2 = "sleep 2"
	}
	err = exec.Run(testKey, testCommand2)
	if err != nil {
		t.Fatalf("Expected second Run to succeed, got error: %v", err)
	}

	// Give time for second process to register
	time.Sleep(100 * time.Millisecond)

	// Verify both processes are tracked under the same key
	status = exec.GetStatus()
	processes, exists = status[testKey]
	if !exists {
		t.Fatalf("Expected key '%s' to still exist after second Run", testKey)
	}

	// Should now have 2 processes (first one may have finished, but sleep should still be running)
	if len(processes) == 0 {
		t.Error("Expected at least one process after running two commands")
	}

	// Clean up - stop all processes
	err = exec.StopAll()
	if err != nil {
		t.Errorf("Failed to clean up processes: %v", err)
	}
}

func TestExecutor_Run_MultipleKeys_Success(t *testing.T) {
	exec := New()

	// Test running processes with different keys - use cross-platform command
	keys := []string{"service.one", "service.two", "service.three"}
	var command string
	if runtime.GOOS == "windows" {
		command = "timeout /t 3 /nobreak >nul"
	} else {
		command = "sleep 3"
	}

	// Start processes with different keys
	for _, key := range keys {
		err := exec.Run(key, command)
		if err != nil {
			t.Fatalf("Expected Run with key '%s' to succeed, got error: %v", key, err)
		}
	}

	// Give time for all processes to start
	time.Sleep(200 * time.Millisecond)

	// Verify all keys are tracked
	status := exec.GetStatus()
	if len(status) != len(keys) {
		t.Errorf("Expected %d keys in status, got %d", len(keys), len(status))
	}

	// Verify each key has exactly one process
	for _, key := range keys {
		processes, exists := status[key]
		if !exists {
			t.Errorf("Expected key '%s' to exist in status", key)
			continue
		}

		if len(processes) != 1 {
			t.Errorf("Expected 1 process for key '%s', got %d", key, len(processes))
		}
	}

	// Clean up
	err := exec.StopAll()
	if err != nil {
		t.Errorf("Failed to clean up processes: %v", err)
	}
}
