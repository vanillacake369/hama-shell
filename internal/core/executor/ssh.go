package executor

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/creack/pty"
)

// SSHManager handles SSH connections with PTY support
type SSHManager struct {
	ptyInterface PTYInterface
}

// NewSSHManager creates a new SSH manager
func NewSSHManager() *SSHManager {
	return &SSHManager{
		ptyInterface: &RealPTY{},
	}
}

// NewSSHManagerWithPTY creates SSH manager with custom PTY interface (for testing)
func NewSSHManagerWithPTY(ptyInterface PTYInterface) *SSHManager {
	return &SSHManager{
		ptyInterface: ptyInterface,
	}
}

// executeSSHWithPTY executes SSH command with password authentication via PTY
func (sm *SSHManager) executeSSHWithPTY(sshCmd, password string, remoteCmds []string) error {
	if sshCmd == "" {
		return fmt.Errorf("SSH command cannot be empty")
	}

	// Create SSH command
	cmd := exec.Command("sh", "-c", sshCmd)

	// Start PTY
	ptmx, err := sm.ptyInterface.Start(cmd)
	if err != nil {
		return fmt.Errorf("failed to start PTY: %w", err)
	}
	defer ptmx.Close()

	// Handle SSH interaction
	return sm.handleSSHSession(ptmx, password, remoteCmds)
}

// handleSSHSession manages the SSH session with password auth and remote commands
func (sm *SSHManager) handleSSHSession(ptmx PTYFile, password string, remoteCmds []string) error {
	// Use a single goroutine to handle both input and output
	done := make(chan error, 1)
	
	go func() {
		defer close(done)
		
		buf := make([]byte, 1024)
		var accumulated strings.Builder
		passwordSent := false
		commandsSent := false
		
		for {
			n, err := ptmx.Read(buf)
			if err != nil {
				if err == io.EOF {
					done <- nil
				} else {
					done <- fmt.Errorf("PTY read error: %w", err)
				}
				return
			}
			
			// Copy output to stdout
			os.Stdout.Write(buf[:n])
			
			// Accumulate output for prompt detection
			output := string(buf[:n])
			accumulated.WriteString(output)
			currentOutput := accumulated.String()
			
			// Handle password prompt
			if !passwordSent && sm.isPasswordPrompt(currentOutput) {
				if password != "" {
					if _, err := ptmx.Write([]byte(password + "\n")); err != nil {
						done <- fmt.Errorf("failed to send password: %w", err)
						return
					}
					passwordSent = true
					accumulated.Reset() // Clear accumulated output
					time.Sleep(1 * time.Second) // Wait for authentication
				} else {
					done <- fmt.Errorf("password prompt detected but no password provided")
					return
				}
			}
			
			// Handle shell prompt (after successful login)
			if passwordSent && !commandsSent && sm.isShellPrompt(currentOutput) {
				// Send remote commands
				for _, cmd := range remoteCmds {
					if _, err := ptmx.Write([]byte(cmd + "\n")); err != nil {
						done <- fmt.Errorf("failed to send command '%s': %w", cmd, err)
						return
					}
					time.Sleep(100 * time.Millisecond) // Brief pause between commands
				}
				commandsSent = true
				
				// If no commands to send, we can exit
				if len(remoteCmds) == 0 {
					done <- nil
					return
				}
			}
			
			// Handle authentication failure
			if passwordSent && strings.Contains(strings.ToLower(currentOutput), "permission denied") {
				done <- fmt.Errorf("SSH authentication failed")
				return
			}
		}
	}()

	// Wait for completion or timeout
	select {
	case err := <-done:
		return err
	case <-time.After(30 * time.Second):
		return fmt.Errorf("SSH session timeout")
	}
}

// isPasswordPrompt checks if a line contains a password prompt
func (sm *SSHManager) isPasswordPrompt(line string) bool {
	lowerLine := strings.ToLower(line)
	passwordIndicators := []string{
		"password:",
		"password for",
		"enter password",
		"'s password:",
	}
	
	for _, indicator := range passwordIndicators {
		if strings.Contains(lowerLine, indicator) {
			return true
		}
	}
	return false
}

// isShellPrompt checks if a line contains a shell prompt
func (sm *SSHManager) isShellPrompt(line string) bool {
	// Common shell prompt patterns
	shellIndicators := []string{
		"$",  // Bash/sh prompt
		"#",  // Root prompt
		"%",  // Zsh prompt
		">",  // PowerShell/cmd
	}
	
	trimmedLine := strings.TrimSpace(line)
	if len(trimmedLine) == 0 {
		return false
	}
	
	// Check if line ends with shell prompt indicators
	for _, indicator := range shellIndicators {
		if strings.HasSuffix(trimmedLine, indicator) {
			return true
		}
	}
	
	return false
}

// RealPTY implements PTYInterface using the actual pty library
type RealPTY struct{}

func (r *RealPTY) Start(cmd *exec.Cmd) (PTYFile, error) {
	return pty.Start(cmd)
}