package executor

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// SignalManager handles signal forwarding and graceful shutdown
type SignalManager struct {
	registry *SessionRegistry
}

// NewSignalManager creates a new signal manager
func NewSignalManager(registry *SessionRegistry) *SignalManager {
	return &SignalManager{
		registry: registry,
	}
}

// manageSupervisor monitors and manages a supervisor process lifecycle
func (sm *SignalManager) manageSupervisor(session *SessionGroup) {
	if session == nil {
		return
	}

	// Set up signal forwarding
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Monitor signals in background
	go func() {
		for sig := range sigChan {
			sm.forwardSignalToPGID(session.PGID, sig.(syscall.Signal))
		}
	}()

	// Wait for supervisor completion
	err := session.Supervisor.Wait()
	signal.Stop(sigChan)
	close(sigChan)

	// Mark session as done
	close(session.Done)

	// Log completion (in real implementation, use proper logger)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Session %s ended with error: %v\n", session.Key, err)
	}
}

// forwardSignalToPGID forwards a signal to an entire process group
func (sm *SignalManager) forwardSignalToPGID(pgid int, sig syscall.Signal) error {
	if pgid <= 0 {
		return fmt.Errorf("invalid PGID: %d", pgid)
	}

	// Send signal to entire process group (negative PID)
	if err := syscall.Kill(-pgid, sig); err != nil {
		return fmt.Errorf("failed to send signal %v to PGID %d: %w", sig, pgid, err)
	}

	return nil
}

// gracefulShutdown performs graceful shutdown of a session with timeout
func (sm *SignalManager) gracefulShutdown(session *SessionGroup, timeout time.Duration) error {
	if session == nil {
		return fmt.Errorf("session cannot be nil")
	}

	if session.PGID <= 0 {
		return fmt.Errorf("invalid PGID: %d", session.PGID)
	}

	// Step 0: Check if session already completed
	select {
	case <-session.Done:
		// Already completed - no shutdown needed
		return nil
	default:
		// Continue with shutdown
	}

	// Step 1: Send SIGTERM for graceful shutdown
	if err := sm.forwardSignalToPGID(session.PGID, syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to send SIGTERM: %w", err)
	}

	// Step 2: Wait for graceful shutdown with timeout
	select {
	case <-session.Done:
		// Graceful shutdown completed
		return nil
	case <-time.After(timeout):
		// Timeout - proceed to force kill
		break
	}

	// Step 3: Force kill with SIGKILL
	if err := sm.forwardSignalToPGID(session.PGID, syscall.SIGKILL); err != nil {
		return fmt.Errorf("failed to send SIGKILL: %w", err)
	}

	// Step 4: Final wait for process cleanup
	select {
	case <-session.Done:
		return nil
	case <-time.After(5 * time.Second):
		return fmt.Errorf("session %s failed to terminate after SIGKILL", session.Key)
	}
}

// shutdownAllSessions gracefully shuts down all active sessions
func (sm *SignalManager) shutdownAllSessions(timeout time.Duration) error {
	sessions := sm.registry.getAllSessions()
	if len(sessions) == 0 {
		return nil
	}

	errors := make(chan error, len(sessions))
	done := make(chan struct{})

	// Shutdown all sessions concurrently
	for key, session := range sessions {
		go func(k string, s *SessionGroup) {
			if err := sm.gracefulShutdown(s, timeout); err != nil {
				errors <- fmt.Errorf("failed to shutdown session %s: %w", k, err)
			} else {
				errors <- nil
			}
		}(key, session)
	}

	// Collect results
	go func() {
		var shutdownErrors []error
		for i := 0; i < len(sessions); i++ {
			if err := <-errors; err != nil {
				shutdownErrors = append(shutdownErrors, err)
			}
		}

		if len(shutdownErrors) > 0 {
			// Combine errors
			errorMsg := fmt.Sprintf("failed to shutdown %d sessions", len(shutdownErrors))
			for _, err := range shutdownErrors {
				errorMsg += "; " + err.Error()
			}
			errors <- fmt.Errorf("%s", errorMsg)
		} else {
			errors <- nil
		}
		close(done)
	}()

	// Wait for all shutdowns to complete
	select {
	case <-done:
		return <-errors
	case <-time.After(timeout + 10*time.Second):
		return fmt.Errorf("timeout waiting for all sessions to shutdown")
	}
}

// terminateSession immediately terminates a session (force kill)
func (sm *SignalManager) terminateSession(session *SessionGroup) error {
	if session == nil {
		return fmt.Errorf("session cannot be nil")
	}

	if session.PGID <= 0 {
		return fmt.Errorf("invalid PGID: %d", session.PGID)
	}

	// Send SIGKILL to entire process group
	if err := sm.forwardSignalToPGID(session.PGID, syscall.SIGKILL); err != nil {
		return fmt.Errorf("failed to terminate session %s: %w", session.Key, err)
	}

	// Wait for process cleanup
	if session.Supervisor != nil {
		// Give a short time for cleanup
		go func() {
			session.Supervisor.Wait()
		}()
	}

	return nil
}

// isProcessRunning checks if a process is still running
func (sm *SignalManager) isProcessRunning(pid int) bool {
	if pid <= 0 {
		return false
	}

	// Send signal 0 to check if process exists
	err := syscall.Kill(pid, 0)
	return err == nil
}

// waitForSessionCompletion waits for a session to complete with timeout
func (sm *SignalManager) waitForSessionCompletion(session *SessionGroup, timeout time.Duration) error {
	if session == nil {
		return fmt.Errorf("session cannot be nil")
	}

	select {
	case <-session.Done:
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("timeout waiting for session %s to complete", session.Key)
	}
}