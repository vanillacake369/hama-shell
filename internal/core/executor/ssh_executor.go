package executor

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/creack/pty"
)

// SSHExecutor handles SSH connections with command execution
type SSHExecutor struct {
	Host     string
	User     string
	Password string
	Commands []string
	Timeout  time.Duration
}

// NewSSHExecutor creates a new SSH executor with defaults
func NewSSHExecutor(host, user, password string, timeout time.Duration) *SSHExecutor {
	return &SSHExecutor{
		Host:     host,
		User:     user,
		Password: password,
		Timeout:  timeout,
	}
}

// ExecuteWithPTY runs commands via SSH with PTY (interactive)
func (e *SSHExecutor) ExecuteWithPTY() error {
	ctx, cancel := context.WithTimeout(context.Background(), e.Timeout)
	defer cancel()

	// Start pty
	cmd := exec.CommandContext(
		ctx,
		"ssh",
		"-tt",
		fmt.Sprintf("%s@%s", e.User, e.Host),
		"bash",
		"--norc",
		"-i",
	)
	ptySession, err := pty.Start(cmd)
	if err != nil {
		return fmt.Errorf("start pty: %w", err)
	}
	defer func() {
		closeErr := ptySession.Close()
		if closeErr != nil {
			panic(closeErr)
		}
	}()

	// Handle I/O in background
	done := make(chan error, 1)
	go e.handlePTYSession(ptySession, done)

	// Wait for completion or timeout
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		if cmd.Process != nil {
			killErr := cmd.Process.Kill()
			if killErr != nil {
				return killErr
			}
		}
		return fmt.Errorf("timeout after %v", e.Timeout)
	}
}

// sessionState manages the state of an interactive PTY session
type sessionState struct {
	ptySession     *os.File
	buffer         []byte
	output         string
	passwordSent   bool
	commandIndex   int
	lastPromptTime time.Time
	done           chan<- error
}

// newSessionState creates a new session state with sensible defaults
func newSessionState(ptySession *os.File, done chan<- error) *sessionState {
	return &sessionState{
		ptySession:     ptySession,
		buffer:         make([]byte, 4096),
		output:         "",
		passwordSent:   false,
		commandIndex:   0,
		lastPromptTime: time.Now(),
		done:           done,
	}
}

// handlePTYSession orchestrates the PTY session lifecycle
func (e *SSHExecutor) handlePTYSession(ptySession *os.File, done chan<- error) {
	state := newSessionState(ptySession, done)
	
	for {
		if !e.readAndProcessOutput(state) {
			return // Session ended
		}
		
		e.preventBufferOverflow(state)
	}
}

// readAndProcessOutput reads data from PTY and processes it based on current state
func (e *SSHExecutor) readAndProcessOutput(state *sessionState) bool {
	chunk, err := e.readChunkFromPTY(state)
	if err != nil {
		e.handleReadError(state, err)
		return false
	}
	
	state.output += chunk
	e.echoToConsole(chunk)
	
	return e.processSessionOutput(state)
}

// readChunkFromPTY reads a chunk of data from the PTY session
func (e *SSHExecutor) readChunkFromPTY(state *sessionState) (string, error) {
	n, err := state.ptySession.Read(state.buffer)
	if err != nil {
		return "", err
	}
	return string(state.buffer[:n]), nil
}

// handleReadError processes read errors and signals completion
func (e *SSHExecutor) handleReadError(state *sessionState, err error) {
	if err == io.EOF {
		state.done <- nil
	} else {
		state.done <- fmt.Errorf("read pty: %w", err)
	}
}

// echoToConsole displays output to the user's console
func (e *SSHExecutor) echoToConsole(chunk string) {
	fmt.Print(chunk)
}

// processSessionOutput determines next action based on accumulated output
func (e *SSHExecutor) processSessionOutput(state *sessionState) bool {
	if e.shouldHandlePasswordPrompt(state) {
		return e.handlePasswordPrompt(state)
	}
	
	if e.shouldHandleShellPrompt(state) {
		return e.handleShellPrompt(state)
	}
	
	return true // Continue session
}

// shouldHandlePasswordPrompt checks if we need to send password
func (e *SSHExecutor) shouldHandlePasswordPrompt(state *sessionState) bool {
	return !state.passwordSent && e.containsPasswordPrompt(state.output)
}

// containsPasswordPrompt detects password prompt in output
func (e *SSHExecutor) containsPasswordPrompt(output string) bool {
	return strings.Contains(strings.ToLower(output), "password:")
}

// handlePasswordPrompt sends the password and updates session state
func (e *SSHExecutor) handlePasswordPrompt(state *sessionState) bool {
	time.Sleep(100 * time.Millisecond) // Brief delay for stability
	
	if err := e.writeToSession(state, e.Password); err != nil {
		state.done <- fmt.Errorf("write password: %w", err)
		return false
	}
	
	e.updateStateAfterPassword(state)
	return true
}

// updateStateAfterPassword cleans state after successful password entry
func (e *SSHExecutor) updateStateAfterPassword(state *sessionState) {
	state.passwordSent = true
	state.output = ""
	state.lastPromptTime = time.Now()
}

// shouldHandleShellPrompt checks if we should process shell commands
func (e *SSHExecutor) shouldHandleShellPrompt(state *sessionState) bool {
	return state.passwordSent && 
		   e.hasShellPrompt(state.output) && 
		   e.hasEnoughTimePassed(state)
}

// hasShellPrompt detects common shell prompt patterns
func (e *SSHExecutor) hasShellPrompt(output string) bool {
	promptPatterns := []string{"$", "#", ">", "bash-"}
	
	for _, pattern := range promptPatterns {
		if strings.Contains(output, pattern) {
			return true
		}
	}
	return false
}

// hasEnoughTimePassed ensures sufficient delay between prompt detection and action
func (e *SSHExecutor) hasEnoughTimePassed(state *sessionState) bool {
	return time.Since(state.lastPromptTime) > 100*time.Millisecond
}

// handleShellPrompt processes shell prompt and executes next command or exits
func (e *SSHExecutor) handleShellPrompt(state *sessionState) bool {
	e.clearOutputAndResetTimer(state)
	
	if e.hasMoreCommands(state) {
		return e.executeNextCommand(state)
	}
	
	return e.exitSession(state)
}

// clearOutputAndResetTimer cleans accumulated output and resets timing
func (e *SSHExecutor) clearOutputAndResetTimer(state *sessionState) {
	state.output = ""
	state.lastPromptTime = time.Now()
}

// hasMoreCommands checks if there are remaining commands to execute
func (e *SSHExecutor) hasMoreCommands(state *sessionState) bool {
	return state.commandIndex < len(e.Commands)
}

// executeNextCommand sends the next command in the sequence
func (e *SSHExecutor) executeNextCommand(state *sessionState) bool {
	time.Sleep(100 * time.Millisecond) // Brief delay for stability
	
	command := e.Commands[state.commandIndex]
	if err := e.writeToSession(state, command); err != nil {
		state.done <- fmt.Errorf("write command: %w", err)
		return false
	}
	
	state.commandIndex++
	return true
}

// exitSession cleanly closes the SSH session
func (e *SSHExecutor) exitSession(state *sessionState) bool {
	time.Sleep(200 * time.Millisecond) // Brief delay before exit
	
	if err := e.writeToSession(state, "exit"); err != nil {
		state.done <- fmt.Errorf("write exit: %w", err)
		return false
	}
	
	// Allow time for clean connection closure
	time.Sleep(500 * time.Millisecond)
	state.done <- nil
	return false
}

// writeToSession sends a command or input to the PTY session
func (e *SSHExecutor) writeToSession(state *sessionState, text string) error {
	_, err := state.ptySession.Write([]byte(text + "\n"))
	return err
}

// preventBufferOverflow manages memory usage by trimming large output buffers
func (e *SSHExecutor) preventBufferOverflow(state *sessionState) {
	const maxBufferSize = 8192
	const trimSize = 4096
	
	if len(state.output) > maxBufferSize {
		state.output = state.output[trimSize:]
	}
}
