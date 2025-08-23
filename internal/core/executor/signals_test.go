package executor

import (
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSignalManager_forwardSignalToPGID_Success(t *testing.T) {
	// GIVEN: signal manager and valid PGID
	registry := NewSessionRegistry()
	sm := NewSignalManager(registry)
	
	// For testing, we'll use our own process group
	// This is safe because we're sending signal 0 (existence check)
	currentPGID, _ := syscall.Getpgid(os.Getpid())
	
	// WHEN: forwarding signal 0 (existence check) to current PGID
	err := sm.forwardSignalToPGID(currentPGID, 0)
	
	// THEN: should succeed (process group exists)
	assert.NoError(t, err)
}

func TestSignalManager_forwardSignalToPGID_InvalidPGID(t *testing.T) {
	// GIVEN: signal manager and invalid PGID
	registry := NewSessionRegistry()
	sm := NewSignalManager(registry)
	
	// WHEN: forwarding signal to invalid PGID
	err := sm.forwardSignalToPGID(-1, syscall.SIGTERM)
	
	// THEN: should return error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid PGID: -1")
}

func TestSignalManager_forwardSignalToPGID_NonExistentPGID(t *testing.T) {
	// GIVEN: signal manager and non-existent PGID
	registry := NewSessionRegistry()
	sm := NewSignalManager(registry)
	
	// WHEN: forwarding signal to non-existent PGID
	err := sm.forwardSignalToPGID(999999, syscall.SIGTERM)
	
	// THEN: should return error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send signal")
}

func TestSignalManager_gracefulShutdown_NilSession(t *testing.T) {
	// GIVEN: signal manager and nil session
	registry := NewSessionRegistry()
	sm := NewSignalManager(registry)
	
	// WHEN: attempting graceful shutdown of nil session
	err := sm.gracefulShutdown(nil, 5*time.Second)
	
	// THEN: should return error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "session cannot be nil")
}

func TestSignalManager_gracefulShutdown_InvalidPGID(t *testing.T) {
	// GIVEN: signal manager and session with invalid PGID
	registry := NewSessionRegistry()
	sm := NewSignalManager(registry)
	session := &SessionGroup{
		Key:  "test-session",
		PGID: -1,
		Done: make(chan struct{}),
	}
	
	// WHEN: attempting graceful shutdown
	err := sm.gracefulShutdown(session, 1*time.Second)
	
	// THEN: should return error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid PGID: -1")
}

func TestSignalManager_gracefulShutdown_QuickCompletion(t *testing.T) {
	// GIVEN: signal manager and session that completes quickly
	registry := NewSessionRegistry()
	sm := NewSignalManager(registry)
	
	// Use a fake PGID that doesn't exist, but simulate quick completion
	session := &SessionGroup{
		Key:  "test-session",
		PGID: 999999, // Non-existent PGID
		Done: make(chan struct{}),
	}
	
	// Simulate quick completion (close Done channel immediately)
	close(session.Done)
	
	// WHEN: attempting graceful shutdown
	// Since the Done channel is already closed, gracefulShutdown should
	// detect this immediately and return without sending any signals
	err := sm.gracefulShutdown(session, 5*time.Second)
	
	// THEN: should complete successfully (no signal sending needed)
	assert.NoError(t, err)
}

func TestSignalManager_terminateSession_Success(t *testing.T) {
	// GIVEN: signal manager and valid session
	registry := NewSessionRegistry()
	sm := NewSignalManager(registry)
	
	// Use a fake PGID to avoid actually sending signals
	session := &SessionGroup{
		Key:  "test-session",
		PGID: 999999, // Non-existent PGID for safe testing
		Done: make(chan struct{}),
	}
	
	// WHEN: terminating session
	err := sm.terminateSession(session)
	
	// THEN: should return error because PGID doesn't exist
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to terminate session")
}

func TestSignalManager_terminateSession_NilSession(t *testing.T) {
	// GIVEN: signal manager and nil session
	registry := NewSessionRegistry()
	sm := NewSignalManager(registry)
	
	// WHEN: terminating nil session
	err := sm.terminateSession(nil)
	
	// THEN: should return error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "session cannot be nil")
}

func TestSignalManager_terminateSession_InvalidPGID(t *testing.T) {
	// GIVEN: signal manager and session with invalid PGID
	registry := NewSessionRegistry()
	sm := NewSignalManager(registry)
	session := &SessionGroup{
		Key:  "test-session",
		PGID: 0,
		Done: make(chan struct{}),
	}
	
	// WHEN: terminating session
	err := sm.terminateSession(session)
	
	// THEN: should return error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid PGID: 0")
}

func TestSignalManager_isProcessRunning(t *testing.T) {
	registry := NewSessionRegistry()
	sm := NewSignalManager(registry)
	
	tests := []struct {
		name     string
		pid      int
		expected bool
	}{
		{"current process", os.Getpid(), true},
		{"invalid pid", -1, false},
		{"zero pid", 0, false},
		{"non-existent pid", 999999, false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// GIVEN: PID to check
			// WHEN: checking if process is running
			result := sm.isProcessRunning(tt.pid)
			// THEN: should match expected result
			assert.Equal(t, tt.expected, result, "PID %d running status", tt.pid)
		})
	}
}

func TestSignalManager_waitForSessionCompletion_Success(t *testing.T) {
	// GIVEN: signal manager and session that will complete
	registry := NewSessionRegistry()
	sm := NewSignalManager(registry)
	session := &SessionGroup{
		Key:  "test-session",
		Done: make(chan struct{}),
	}
	
	// Simulate completion after delay
	go func() {
		time.Sleep(100 * time.Millisecond)
		close(session.Done)
	}()
	
	// WHEN: waiting for session completion
	err := sm.waitForSessionCompletion(session, 1*time.Second)
	
	// THEN: should complete successfully
	assert.NoError(t, err)
}

func TestSignalManager_waitForSessionCompletion_Timeout(t *testing.T) {
	// GIVEN: signal manager and session that won't complete
	registry := NewSessionRegistry()
	sm := NewSignalManager(registry)
	session := &SessionGroup{
		Key:  "test-session",
		Done: make(chan struct{}), // Never closed
	}
	
	// WHEN: waiting for session completion with short timeout
	err := sm.waitForSessionCompletion(session, 100*time.Millisecond)
	
	// THEN: should timeout
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timeout waiting for session test-session to complete")
}

func TestSignalManager_waitForSessionCompletion_NilSession(t *testing.T) {
	// GIVEN: signal manager and nil session
	registry := NewSessionRegistry()
	sm := NewSignalManager(registry)
	
	// WHEN: waiting for nil session completion
	err := sm.waitForSessionCompletion(nil, 1*time.Second)
	
	// THEN: should return error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "session cannot be nil")
}

func TestSignalManager_shutdownAllSessions_Empty(t *testing.T) {
	// GIVEN: signal manager with empty registry
	registry := NewSessionRegistry()
	sm := NewSignalManager(registry)
	
	// WHEN: shutting down all sessions
	err := sm.shutdownAllSessions(1 * time.Second)
	
	// THEN: should complete successfully (nothing to shutdown)
	assert.NoError(t, err)
}

func TestSignalManager_shutdownAllSessions_WithSessions(t *testing.T) {
	// GIVEN: signal manager with registry containing sessions
	registry := NewSessionRegistry()
	sm := NewSignalManager(registry)
	
	// Create sessions that are already completed (Done channels closed)
	session1 := &SessionGroup{
		Key:  "session-1",
		PGID: 999999, // Non-existent PGID since we're pre-closing Done
		Done: make(chan struct{}),
	}
	session2 := &SessionGroup{
		Key:  "session-2", 
		PGID: 999998, // Non-existent PGID since we're pre-closing Done
		Done: make(chan struct{}),
	}
	
	// Pre-close the Done channels to simulate completed sessions
	close(session1.Done)
	close(session2.Done)
	
	registry.registerSession("session-1", session1)
	registry.registerSession("session-2", session2)
	
	// WHEN: shutting down all sessions
	err := sm.shutdownAllSessions(1 * time.Second)
	
	// THEN: should complete successfully (sessions already done)
	assert.NoError(t, err)
}

func TestSignalManager_manageSupervisor_NilSession(t *testing.T) {
	// GIVEN: signal manager and nil session
	registry := NewSessionRegistry()
	sm := NewSignalManager(registry)
	
	// WHEN: managing nil supervisor (should not panic)
	// THEN: should handle gracefully
	assert.NotPanics(t, func() {
		sm.manageSupervisor(nil)
	})
}

func TestSignalManager_Integration_SupervisorLifecycle(t *testing.T) {
	// GIVEN: signal manager and real supervisor process
	registry := NewSessionRegistry()
	sm := NewSignalManager(registry)
	
	// Create a simple command that runs briefly
	cmd := exec.Command("sleep", "0.1")
	if err := cmd.Start(); err != nil {
		t.Skip("Cannot create test process")
	}
	
	session := &SessionGroup{
		Key:        "test-session",
		SessionID:  cmd.Process.Pid,
		PGID:       cmd.Process.Pid, // Simplified for test
		Supervisor: cmd,
		Done:       make(chan struct{}),
	}
	
	// WHEN: managing supervisor
	start := time.Now()
	sm.manageSupervisor(session)
	duration := time.Since(start)
	
	// THEN: should complete within reasonable time
	assert.True(t, duration < 1*time.Second, "Supervisor management took too long")
	
	// Session should be marked as done
	select {
	case <-session.Done:
		// Good - session completed
	default:
		t.Error("Session was not marked as done")
	}
}