package infra

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/creack/pty"
	"golang.org/x/term"

	"hama-shell/internal/core/service/model"
	"hama-shell/internal/core/terminal"
)

// TerminalManager handles terminal session operations
type TerminalManager struct {
	server terminal.Server
}

// NewTerminalManager creates a new TerminalManager instance
func NewTerminalManager() *TerminalManager {
	return &TerminalManager{
		server: terminal.NewTerminalServer(),
	}
}

// StartInteractiveSession starts an interactive terminal session for a service
func (t *TerminalManager) StartInteractiveSession(service *model.Service) error {
	sessionID := fmt.Sprintf("%s-%d", service.GetFullName(), time.Now().Unix())

	// Save original terminal state
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("failed to set raw mode: %w", err)
	}

	// Setup cleanup function
	cleanup := func() {
		_ = term.Restore(int(os.Stdin.Fd()), oldState)
		_ = t.server.KillSession(sessionID)
	}
	defer cleanup()

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		cleanup()
		os.Exit(0)
	}()

	// Create terminal session
	session, err := t.createSession(sessionID)
	if err != nil {
		return err
	}

	// Setup terminal I/O
	if err := t.setupTerminalIO(sessionID, session); err != nil {
		return err
	}

	// Execute service commands
	go t.executeCommands(session, service.Commands)

	// Wait for session to finish
	for session.IsRunning() {
		time.Sleep(100 * time.Millisecond)
	}

	// Restore terminal before final output
	_ = term.Restore(int(os.Stdin.Fd()), oldState)
	fmt.Printf("\n✅ Session ended normally\n")

	return nil
}

// createSession creates a new terminal session
func (t *TerminalManager) createSession(sessionID string) (terminal.Session, error) {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/bash"
	}

	session, err := t.server.CreateSession(sessionID, shell, []string{})
	if err != nil {
		return nil, fmt.Errorf("failed to create terminal session: %w", err)
	}

	return session, nil
}

// setupTerminalIO configures terminal input/output handling
func (t *TerminalManager) setupTerminalIO(sessionID string, session terminal.Session) error {
	ptyMaster := session.GetPTYMaster()

	// Set terminal size
	if size, err := pty.GetsizeFull(os.Stdin); err == nil {
		if err := t.server.ResizeSession(sessionID, size.Rows, size.Cols); err != nil {
			fmt.Printf("Warning: failed to set PTY size: %v\n", err)
		}
	}

	// Handle window size changes
	go func() {
		sigwinch := make(chan os.Signal, 1)
		signal.Notify(sigwinch, syscall.SIGWINCH)
		for range sigwinch {
			if size, err := pty.GetsizeFull(os.Stdin); err == nil {
				_ = t.server.ResizeSession(sessionID, size.Rows, size.Cols)
			}
		}
	}()

	// Copy stdin to ptyMaster (user input -> shell)
	go func() {
		_, _ = io.Copy(ptyMaster, os.Stdin)
	}()

	// Copy ptyMaster to stdout (shell output -> terminal)
	go func() {
		_, _ = io.Copy(os.Stdout, ptyMaster)
	}()

	return nil
}

// executeCommands sends commands to the terminal session
func (t *TerminalManager) executeCommands(session terminal.Session, commands []string) {
	time.Sleep(500 * time.Millisecond) // Wait for shell prompt

	for _, command := range commands {
		commandWithNewline := command + "\n"
		if err := session.WriteInput([]byte(commandWithNewline)); err != nil {
			fmt.Printf("Warning: failed to send command '%s': %v\n", command, err)
		}
		time.Sleep(200 * time.Millisecond) // Small delay between commands
	}
}

// Shutdown gracefully shuts down the terminal manager
func (t *TerminalManager) Shutdown() error {
	return t.server.Shutdown()
}