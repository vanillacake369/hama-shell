package executor

import (
	"fmt"
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
func (sm *SupervisorManager) createSupervisor(key string, segments []CommandSegment) (*SessionGroup, error) {
	if len(segments) == 0 {
		return nil, fmt.Errorf("no command segments provided")
	}

	// Build supervisor script
	script := sm.buildSupervisorScript(segments)
	
	// Create supervisor command
	supervisor := exec.Command("bash", "-c", script)
	
	// Configure supervisor with session/PGID settings
	if err := sm.manager.setupSupervisor(supervisor); err != nil {
		return nil, fmt.Errorf("failed to setup supervisor: %w", err)
	}

	// Start supervisor
	if err := supervisor.Start(); err != nil {
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
func (sm *SupervisorManager) buildSupervisorScript(segments []CommandSegment) string {
	var script strings.Builder
	
	// Script header with error handling
	script.WriteString("#!/bin/bash\n")
	script.WriteString("set -e\n") // Exit on error
	script.WriteString("trap 'kill -TERM -$$; exit' INT TERM EXIT\n") // Cleanup on signals/exit
	script.WriteString("\n")
	
	// Process each segment
	for i, segment := range segments {
		switch segment.Type {
		case "ssh":
			sm.addSSHSegmentToScript(&script, i, segment)
		case "shell":
			sm.addShellSegmentToScript(&script, i, segment)
		default:
			script.WriteString(fmt.Sprintf("echo 'Unknown segment type: %s'\n", segment.Type))
		}
		script.WriteString("\n")
	}
	
	// Wait for all background processes
	script.WriteString("wait\n")
	
	return script.String()
}

// addSSHSegmentToScript adds SSH segment handling to the supervisor script
func (sm *SupervisorManager) addSSHSegmentToScript(script *strings.Builder, index int, segment CommandSegment) {
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
	
	// Use a simple approach for now - we'll enhance with PTY later
	if password != "" {
		// For now, create a simple SSH helper call
		script.WriteString(fmt.Sprintf("hama_ssh_helper '%s' '%s'", 
			sshCmd, password))
		
		// Add remote commands
		for _, remoteCmd := range remoteCmds {
			script.WriteString(fmt.Sprintf(" '%s'", remoteCmd))
		}
		script.WriteString(" &\n")
	} else {
		// SSH without password (key-based auth)
		allCommands := strings.Join(segment.Commands, "; ")
		script.WriteString(fmt.Sprintf("(%s) &\n", allCommands))
	}
	
	script.WriteString(fmt.Sprintf("SSH_PID_%d=$!\n", index))
}

// addShellSegmentToScript adds shell segment handling to the supervisor script
func (sm *SupervisorManager) addShellSegmentToScript(script *strings.Builder, index int, segment CommandSegment) {
	if len(segment.Commands) == 0 {
		return
	}
	
	script.WriteString(fmt.Sprintf("# Shell Segment %d\n", index))
	
	// Join shell commands with &&
	joinedCommands := strings.Join(segment.Commands, " && ")
	script.WriteString(fmt.Sprintf("(%s) &\n", joinedCommands))
	script.WriteString(fmt.Sprintf("SHELL_PID_%d=$!\n", index))
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

// Simple SSH helper placeholder - will be replaced with PTY implementation
func sshHelper(sshCmd, password string, remoteCmds []string) error {
	// This is a placeholder - in the full implementation, this would use PTY
	// to handle password authentication and remote command execution
	fmt.Printf("SSH Helper called: %s with %d remote commands\n", sshCmd, len(remoteCmds))
	return fmt.Errorf("SSH helper not yet implemented - use PTY version")
}