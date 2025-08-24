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

// ToDo : 이제 이걸 1) 개선하고 2) executor 에 merge 하는 일만 남았다!!
// ToDo : 이제 이걸 1) 개선하고 2) executor 에 merge 하는 일만 남았다!!
// ToDo : 이제 이걸 1) 개선하고 2) executor 에 merge 하는 일만 남았다!!
// ToDo : 이제 이걸 1) 개선하고 2) executor 에 merge 하는 일만 남았다!!
// handlePTYSession manages the interactive PTY session
func (e *SSHExecutor) handlePTYSession(ptySession *os.File, done chan<- error) {
	buffer := make([]byte, 4096)
	accumulated := ""

	passwordSent := false
	commandIndex := 0
	lastPromptTime := time.Now()

	for {
		// Read available data (non-blocking with timeout)
		n, err := ptySession.Read(buffer)
		if err != nil {
			if err == io.EOF {
				done <- nil
			} else {
				done <- fmt.Errorf("read pty: %w", err)
			}
			return
		}

		// Accumulate output
		chunk := string(buffer[:n])
		accumulated += chunk

		// Echo output to console
		fmt.Print(chunk)

		// Handle password prompt
		if !passwordSent && strings.Contains(strings.ToLower(accumulated), "password:") {
			time.Sleep(100 * time.Millisecond) // Small delay
			_, err := ptySession.Write([]byte(e.Password + "\n"))
			if err != nil {
				done <- fmt.Errorf("write password: %w", err)
				return
			}
			passwordSent = true
			accumulated = "" // Clear buffer after password
			lastPromptTime = time.Now()
			continue
		}

		// Handle shell prompt (look for common prompt patterns)
		// Check for prompt patterns: $, #, >, or bash-X.X$
		hasPrompt := strings.Contains(accumulated, "$") ||
			strings.Contains(accumulated, "#") ||
			strings.Contains(accumulated, ">") ||
			strings.Contains(accumulated, "bash-")

		// Only process if we have a prompt and sufficient time has passed
		if passwordSent && hasPrompt && time.Since(lastPromptTime) > 100*time.Millisecond {
			// Clear accumulated buffer when we detect a prompt
			accumulated = ""
			lastPromptTime = time.Now()

			if commandIndex < len(e.Commands) {
				time.Sleep(100 * time.Millisecond) // Small delay before sending command
				_, err := ptySession.Write([]byte(e.Commands[commandIndex] + "\n"))
				if err != nil {
					done <- fmt.Errorf("write command: %w", err)
					return
				}
				commandIndex++
			} else {
				time.Sleep(200 * time.Millisecond) // Small delay before exit
				_, err := ptySession.Write([]byte("exit\n"))
				if err != nil {
					done <- fmt.Errorf("write exit: %w", err)
					return
				}
				// Wait a bit for the connection to close cleanly
				time.Sleep(500 * time.Millisecond)
				done <- nil
				return
			}
		}

		// Prevent accumulated buffer from growing too large
		if len(accumulated) > 8192 {
			accumulated = accumulated[4096:]
		}
	}
}
