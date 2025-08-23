package executor

import (
	"strings"
)

// CommandParser handles parsing of command sequences into segments
type CommandParser struct{}

// NewCommandParser creates a new command parser
func NewCommandParser() *CommandParser {
	return &CommandParser{}
}

// parseCommandSegments splits commands into SSH and shell segments
func (p *CommandParser) parseCommandSegments(commands []string) []CommandSegment {
	if len(commands) == 0 {
		return nil
	}

	var segments []CommandSegment
	var currentSegment CommandSegment

	for i := 0; i < len(commands); i++ {
		cmd := commands[i]

		if p.isSSHCommand(cmd) {
			// Save previous segment if exists
			if len(currentSegment.Commands) > 0 {
				segments = append(segments, currentSegment)
				currentSegment = CommandSegment{}
			}

			// Start new SSH segment
			currentSegment = CommandSegment{
				Type:     "ssh",
				Commands: []string{cmd},
			}

			// Next command is likely password if it's not a shell command
			if i+1 < len(commands) && !p.isShellCommand(commands[i+1]) {
				i++
				currentSegment.Commands = append(currentSegment.Commands, commands[i])
			}
		} else {
			// This is not an SSH command
			if currentSegment.Type == "ssh" && p.isRemoteCommand(cmd) {
				// We're in an SSH segment and this looks like a remote command
				currentSegment.Commands = append(currentSegment.Commands, cmd)
			} else {
				// This is a local shell command - start/continue shell segment
				if currentSegment.Type != "shell" {
					// Save any existing segment first
					if len(currentSegment.Commands) > 0 {
						segments = append(segments, currentSegment)
					}
					currentSegment = CommandSegment{
						Type:     "shell",
						Commands: []string{},
					}
				}
				currentSegment.Commands = append(currentSegment.Commands, cmd)
			}
		}
	}

	// Don't forget the last segment
	if len(currentSegment.Commands) > 0 {
		segments = append(segments, currentSegment)
	}

	return segments
}

// isSSHCommand checks if a command is an SSH command
func (p *CommandParser) isSSHCommand(cmd string) bool {
	trimmed := strings.TrimSpace(cmd)
	return strings.HasPrefix(trimmed, "ssh ") || strings.HasPrefix(trimmed, "ssh\t")
}

// isShellCommand checks if a command looks like a typical shell command
func (p *CommandParser) isShellCommand(cmd string) bool {
	trimmed := strings.TrimSpace(cmd)
	shellPrefixes := []string{
		"cd ", "ls", "pwd", "echo ", "cat ", "grep ", "mkdir ",
		"rm ", "mv ", "cp ", "chmod ", "chown ", "find ", "which ",
		"ps ", "top ", "kill ", "killall ", "tail ", "head ",
		"less ", "more ", "wget ", "curl ", "git ", "npm ", "node ",
		"python ", "java ", "make ", "cmake ", "docker ", "kubectl ",
	}

	for _, prefix := range shellPrefixes {
		if strings.HasPrefix(trimmed, prefix) {
			return true
		}
	}

	// Also check for command that starts with path separators or environment variables
	return strings.HasPrefix(trimmed, "/") ||
		strings.HasPrefix(trimmed, "./") ||
		strings.HasPrefix(trimmed, "~/") ||
		strings.Contains(trimmed, "=") // Environment variable assignment
}

// isRemoteCommand checks if a command should be executed remotely via SSH
func (p *CommandParser) isRemoteCommand(cmd string) bool {
	trimmed := strings.TrimSpace(cmd)
	
	// SSH commands within SSH context are remote (SSH jumps)
	if p.isSSHCommand(cmd) {
		return true
	}
	
	// Commands that clearly indicate local execution
	localIndicators := []string{
		"local", "echo local", "pwd local", // Contains "local"
	}
	
	for _, indicator := range localIndicators {
		if strings.Contains(strings.ToLower(trimmed), indicator) {
			return false
		}
	}
	
	// If it's a shell command, it could be remote (default assumption after SSH)
	// But we need to be smarter about when to transition back to local
	return p.isShellCommand(cmd)
}