package executor

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

// SupervisorManager handles creation and management of supervisor processes
type SupervisorManager struct {
	parser  *CommandParser
	manager processManager
}

// NewSupervisorManager creates a new supervisor manager
func NewSupervisorManager() *SupervisorManager {
	return &SupervisorManager{
		parser:  NewCommandParser(),
		manager: newProcessManager(),
	}
}

// createSupervisor creates a supervisor process with session/PGID for command segments
func (sm *SupervisorManager) createSupervisor(key string, segments []CommandSegment, mode ExecutionMode) (*SessionGroup, error) {
	if len(segments) == 0 {
		return nil, fmt.Errorf("no command segments provided")
	}

	// Build supervisor script based on execution mode
	script := sm.buildSupervisorScript(segments, mode)
	
	// Try multiple execution strategies with mode-specific I/O handling
	supervisor, err := sm.createSupervisorWithFallback(script, mode)
	if err != nil {
		return nil, fmt.Errorf("failed to start supervisor: %w", err)
	}

	// Get PGID for process group management
	pgid, err := syscall.Getpgid(supervisor.Process.Pid)
	if err != nil {
		// Attempt cleanup
		supervisor.Process.Kill()
		return nil, fmt.Errorf("failed to get PGID: %w", err)
	}

	// Create session group
	session := &SessionGroup{
		Key:        key,
		SessionID:  supervisor.Process.Pid, // Session leader PID
		PGID:       pgid,
		Supervisor: supervisor,
		Segments:   convertToSegments(segments),
		StartTime:  time.Now(),
		Done:       make(chan struct{}),
	}

	return session, nil
}

// buildSupervisorScript generates bash script to manage all command segments
func (sm *SupervisorManager) buildSupervisorScript(segments []CommandSegment, mode ExecutionMode) string {
	var script strings.Builder
	
	// Script header with error handling
	script.WriteString("#!/bin/bash\n")
	script.WriteString("set -e\n") // Exit on error
	
	// Mode-specific signal handling
	if mode == ExecutionModeBackground {
		script.WriteString("trap 'kill -TERM -$$; exit' INT TERM EXIT\n") // Cleanup for background
	} else {
		script.WriteString("trap 'exit' INT TERM\n") // Direct exit for foreground
	}
	script.WriteString("\n")
	
	// Process each segment with mode-specific behavior
	for i, segment := range segments {
		switch segment.Type {
		case "ssh":
			sm.addSSHSegmentToScript(&script, i, segment, mode)
		case "shell":
			sm.addShellSegmentToScript(&script, i, segment, mode)
		default:
			script.WriteString(fmt.Sprintf("echo 'Unknown segment type: %s'\n", segment.Type))
		}
		script.WriteString("\n")
	}
	
	// Mode-specific completion handling
	if mode == ExecutionModeBackground {
		script.WriteString("wait\n") // Wait for background processes
	}
	// For foreground mode, commands run sequentially, so no wait needed
	
	return script.String()
}

// createSupervisorWithFallback tries multiple execution strategies with mode-specific I/O handling
func (sm *SupervisorManager) createSupervisorWithFallback(script string, mode ExecutionMode) (*exec.Cmd, error) {
	// Get platform capabilities for better error messages
	manager := newProcessManager().(*unixProcessManager)
	capabilities := manager.getPlatformCapabilities()
	
	// First, check if we can create processes at all
	if !sm.canCreateProcesses() {
		return nil, fmt.Errorf("process creation not permitted in this environment (%s)", 
			capabilities.platformName)
	}
	
	strategies := []struct {
		name string
		cmd  func(string) *exec.Cmd
	}{
		{"bash", func(script string) *exec.Cmd { return exec.Command("bash", "-c", script) }},
		{"sh", func(script string) *exec.Cmd { return exec.Command("sh", "-c", script) }},
		{"zsh", func(script string) *exec.Cmd { return exec.Command("zsh", "-c", script) }},
	}
	
	var lastErr error
	
	for _, strategy := range strategies {
		supervisor := strategy.cmd(script)
		
		// Configure I/O forwarding based on execution mode
		if mode == ExecutionModeForeground {
			// Forward stdin/stdout/stderr for interactive use
			supervisor.Stdin = os.Stdin
			supervisor.Stdout = os.Stdout
			supervisor.Stderr = os.Stderr
		}
		// For background mode, keep default behavior (no I/O forwarding)
		
		// Configure supervisor with session/PGID settings
		if err := sm.manager.setupSupervisor(supervisor); err != nil {
			lastErr = fmt.Errorf("%s setup failed: %w", strategy.name, err)
			continue
		}

		// Try to start supervisor
		if err := supervisor.Start(); err != nil {
			lastErr = fmt.Errorf("%s start failed: %w", strategy.name, err)
			continue
		}
		
		// Success!
		return supervisor, nil
	}
	
	// Provide platform-specific error guidance
	platformGuidance := sm.getPlatformErrorGuidance(capabilities, lastErr)
	return nil, fmt.Errorf("all execution strategies failed on %s: %w\n%s", 
		capabilities.platformName, lastErr, platformGuidance)
}

// canCreateProcesses checks if the system allows process creation
func (sm *SupervisorManager) canCreateProcesses() bool {
	// Try creating a simple test process
	testCmd := exec.Command("echo", "test")
	if err := testCmd.Start(); err != nil {
		return false
	}
	testCmd.Wait() // Clean up
	return true
}

// addSSHSegmentToScript adds SSH segment handling to the supervisor script
func (sm *SupervisorManager) addSSHSegmentToScript(script *strings.Builder, index int, segment CommandSegment, mode ExecutionMode) {
	if len(segment.Commands) == 0 {
		return
	}
	
	script.WriteString(fmt.Sprintf("# SSH Segment %d\n", index))
	
	sshCmd := segment.Commands[0]
	var password string
	var remoteCmds []string
	
	// Extract password and remote commands
	if len(segment.Commands) > 1 && !sm.parser.isShellCommand(segment.Commands[1]) {
		password = segment.Commands[1]
		if len(segment.Commands) > 2 {
			remoteCmds = segment.Commands[2:]
		}
	} else {
		// No password, treat remaining as remote commands
		remoteCmds = segment.Commands[1:]
	}
	
	// Mode-specific SSH execution
	if password != "" {
		// For now, create a simple SSH helper call
		script.WriteString(fmt.Sprintf("hama_ssh_helper '%s' '%s'", 
			sshCmd, password))
		
		// Add remote commands
		for _, remoteCmd := range remoteCmds {
			script.WriteString(fmt.Sprintf(" '%s'", remoteCmd))
		}
		
		// Mode-specific execution
		if mode == ExecutionModeBackground {
			script.WriteString(" &\n")
			script.WriteString(fmt.Sprintf("SSH_PID_%d=$!\n", index))
		} else {
			script.WriteString("\n") // Foreground - run directly
		}
	} else {
		// SSH without password (key-based auth)
		allCommands := strings.Join(segment.Commands, "; ")
		
		if mode == ExecutionModeBackground {
			script.WriteString(fmt.Sprintf("(%s) &\n", allCommands))
			script.WriteString(fmt.Sprintf("SSH_PID_%d=$!\n", index))
		} else {
			script.WriteString(fmt.Sprintf("(%s)\n", allCommands)) // Foreground
		}
	}
}

// addShellSegmentToScript adds shell segment handling to the supervisor script
func (sm *SupervisorManager) addShellSegmentToScript(script *strings.Builder, index int, segment CommandSegment, mode ExecutionMode) {
	if len(segment.Commands) == 0 {
		return
	}
	
	script.WriteString(fmt.Sprintf("# Shell Segment %d\n", index))
	
	// Join shell commands with &&
	joinedCommands := strings.Join(segment.Commands, " && ")
	
	// Mode-specific execution
	if mode == ExecutionModeBackground {
		script.WriteString(fmt.Sprintf("(%s) &\n", joinedCommands))
		script.WriteString(fmt.Sprintf("SHELL_PID_%d=$!\n", index))
	} else {
		script.WriteString(fmt.Sprintf("(%s)\n", joinedCommands)) // Foreground
	}
}

// convertToSegments converts CommandSegments to Segments for runtime tracking
func convertToSegments(cmdSegments []CommandSegment) []*Segment {
	segments := make([]*Segment, len(cmdSegments))
	for i, cmdSeg := range cmdSegments {
		segments[i] = &Segment{
			Type:     cmdSeg.Type,
			Commands: cmdSeg.Commands,
			Process:  nil, // Will be set when process starts
			PTY:      nil, // Will be set for SSH segments if needed
		}
	}
	return segments
}

// getPlatformErrorGuidance provides platform-specific guidance for common errors
func (sm *SupervisorManager) getPlatformErrorGuidance(capabilities platformCapabilities, err error) string {
	if err == nil {
		return ""
	}
	
	errorStr := err.Error()
	
	// Platform-specific guidance based on common error patterns
	switch capabilities.platformName {
	case "macOS/Darwin":
		if strings.Contains(errorStr, "operation not permitted") {
			return "💡 macOS Troubleshooting:\n" +
				"  • This may be due to macOS security restrictions\n" +
				"  • Try running from Terminal.app instead of IDE terminals\n" +
				"  • Check if your terminal has 'Full Disk Access' in System Preferences\n" +
				"  • Consider running from a less restricted environment"
		}
		if strings.Contains(errorStr, "fork/exec") {
			return "💡 macOS Note:\n" +
				"  • Process creation is restricted in some macOS environments\n" +
				"  • The application uses session-only mode on macOS (Setsid without Setpgid)\n" +
				"  • This provides compatibility while maintaining process group functionality"
		}
	case "Linux":
		if strings.Contains(errorStr, "operation not permitted") {
			return "💡 Linux Troubleshooting:\n" +
				"  • Check if you have permission to create process groups\n" +
				"  • Verify you're not in a restricted container or chroot environment\n" +
				"  • Try running with different user privileges"
		}
	default:
		if strings.Contains(errorStr, "operation not permitted") {
			return "💡 Unix System Troubleshooting:\n" +
				"  • Check system permissions for process creation\n" +
				"  • Verify your user has appropriate privileges\n" +
				"  • Consider running in a less restricted environment"
		}
	}
	
	return "💡 General Troubleshooting:\n" +
		"  • Try running the commands manually to verify they work\n" +
		"  • Check system logs for additional error information\n" +
		"  • Ensure all required dependencies are installed"
}

// Simple SSH helper placeholder - will be replaced with PTY implementation
func sshHelper(sshCmd, password string, remoteCmds []string) error {
	// This is a placeholder - in the full implementation, this would use PTY
	// to handle password authentication and remote command execution
	fmt.Printf("SSH Helper called: %s with %d remote commands\n", sshCmd, len(remoteCmds))
	return fmt.Errorf("SSH helper not yet implemented - use PTY version")
}