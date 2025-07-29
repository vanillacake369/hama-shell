package executor

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// ProcessStatusType represents the status of a keep-alive process
type ProcessStatusType int

const (
	ProcessStatusStarting ProcessStatusType = iota
	ProcessStatusRunning
	ProcessStatusFailed
	ProcessStatusRestarting
	ProcessStatusStopped
	ProcessStatusCompleted
)

// String returns the string representation of ProcessStatusType
func (pst ProcessStatusType) String() string {
	switch pst {
	case ProcessStatusStarting:
		return "starting"
	case ProcessStatusRunning:
		return "running"
	case ProcessStatusFailed:
		return "failed"
	case ProcessStatusRestarting:
		return "restarting"
	case ProcessStatusStopped:
		return "stopped"
	case ProcessStatusCompleted:
		return "completed"
	default:
		return "unknown"
	}
}

// CommandExecutor interface defines the contract for command execution
type CommandExecutor interface {
	ExecuteCommands(commands []string) ([]ExecutionResult, error)
	ExecuteCommand(command string) ExecutionResult
	ExecuteCommandInteractive(command string) ExecutionResult
	ExecuteCommandsParallel(commands []string, maxConcurrent int) ([]ExecutionResult, error)
	ExecuteWithRealtimeOutput(command string) ExecutionResult
	ExecuteCommandsKeepAlive(commands []string) error
	StopAll() error
	IsRunning() bool
	GetRunningCommands() []string
	GetProcessStatus() map[string]ProcessStatus
}

// ProcessStatus represents the status of a keep-alive process
type ProcessStatus struct {
	Command      string
	Status       ProcessStatusType
	RestartCount int
	LastRestart  time.Time
	StartTime    time.Time
	PID          int
}

// ExecutionResult contains the result of command execution
type ExecutionResult struct {
	Command   string
	ExitCode  int
	Output    string
	Error     error
	Duration  time.Duration
	StartTime time.Time
}

// commandExecutor is the base implementation
type commandExecutor struct {
	timeout    time.Duration
	ctx        context.Context
	cancel     context.CancelFunc
	mu         sync.RWMutex
	running    map[string]*exec.Cmd
	processes  map[string]*ProcessStatus
	stopSignal chan struct{}
}

// NewCommandExecutor creates a new command executor with default timeout
func NewCommandExecutor(timeout time.Duration) CommandExecutor {
	ctx, cancel := context.WithCancel(context.Background())
	return &commandExecutor{
		timeout:    timeout,
		ctx:        ctx,
		cancel:     cancel,
		running:    make(map[string]*exec.Cmd),
		processes:  make(map[string]*ProcessStatus),
		stopSignal: make(chan struct{}),
	}
}

// ExecuteCommands runs a list of commands sequentially
func (ce *commandExecutor) ExecuteCommands(commands []string) ([]ExecutionResult, error) {
	results := make([]ExecutionResult, 0, len(commands))

	for i, command := range commands {
		fmt.Printf("[%d/%d] Executing: %s\n", i+1, len(commands), command)

		result := ce.ExecuteCommand(command)
		results = append(results, result)

		// Stop execution if a command fails
		if result.Error != nil {
			fmt.Printf("Command failed: %s\n", result.Error)
			break
		}

		// Print command output if there is any
		if result.Output != "" {
			fmt.Printf("Output:\n%s\n", result.Output)
		}
	}

	return results, nil
}

// ExecuteCommand runs a single command with timeout and context
func (ce *commandExecutor) ExecuteCommand(command string) ExecutionResult {
	startTime := time.Now()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ce.ctx, ce.timeout)
	defer cancel()

	// Parse command and arguments
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return ExecutionResult{
			Command:   command,
			ExitCode:  1,
			Error:     fmt.Errorf("empty command"),
			StartTime: startTime,
			Duration:  time.Since(startTime),
		}
	}

	// Create command with context
	cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
	ce.setupCommand(cmd)

	// Store running command for potential cancellation
	ce.mu.Lock()
	ce.running[command] = cmd
	ce.mu.Unlock()

	// Clean up after execution
	defer func() {
		ce.mu.Lock()
		delete(ce.running, command)
		ce.mu.Unlock()
	}()

	// Capture output
	output, err := cmd.CombinedOutput()
	duration := time.Since(startTime)

	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			exitCode = 1
		}
	}

	return ExecutionResult{
		Command:   command,
		ExitCode:  exitCode,
		Output:    string(output),
		Error:     err,
		Duration:  duration,
		StartTime: startTime,
	}
}

// ExecuteCommandInteractive runs a command with interactive input/output
func (ce *commandExecutor) ExecuteCommandInteractive(command string) ExecutionResult {
	startTime := time.Now()

	// Parse command and arguments
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return ExecutionResult{
			Command:   command,
			ExitCode:  1,
			Error:     fmt.Errorf("empty command"),
			StartTime: startTime,
			Duration:  time.Since(startTime),
		}
	}

	// Create command
	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	ce.setupCommand(cmd)

	// Store running command
	ce.mu.Lock()
	ce.running[command] = cmd
	ce.mu.Unlock()

	// Clean up after execution
	defer func() {
		ce.mu.Lock()
		delete(ce.running, command)
		ce.mu.Unlock()
	}()

	// Run command
	err := cmd.Run()
	duration := time.Since(startTime)

	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			exitCode = 1
		}
	}

	return ExecutionResult{
		Command:   command,
		ExitCode:  exitCode,
		Error:     err,
		Duration:  duration,
		StartTime: startTime,
	}
}

// ExecuteCommandsParallel runs commands in parallel with a limit
func (ce *commandExecutor) ExecuteCommandsParallel(commands []string, maxConcurrent int) ([]ExecutionResult, error) {
	results := make([]ExecutionResult, len(commands))
	semaphore := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup

	for i, command := range commands {
		wg.Add(1)
		go func(index int, cmd string) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			fmt.Printf("[%d] Starting: %s\n", index+1, cmd)
			result := ce.ExecuteCommand(cmd)
			results[index] = result

			if result.Error != nil {
				fmt.Printf("[%d] Failed: %s\n", index+1, result.Error)
			} else {
				fmt.Printf("[%d] Completed: %s\n", index+1, cmd)
			}
		}(i, command)
	}

	wg.Wait()
	return results, nil
}

// StopAll cancels all running commands
func (ce *commandExecutor) StopAll() error {
	// Signal all keep-alive processes to stop
	close(ce.stopSignal)

	ce.mu.RLock()
	commands := make([]*exec.Cmd, 0, len(ce.running))
	for _, cmd := range ce.running {
		commands = append(commands, cmd)
	}
	ce.mu.RUnlock()

	var errors []error
	for _, cmd := range commands {
		if err := ce.terminateProcess(cmd); err != nil {
			errors = append(errors, err)
		}
	}

	// Cancel context to stop any new commands
	ce.cancel()

	if len(errors) > 0 {
		return fmt.Errorf("failed to stop %d commands", len(errors))
	}

	return nil
}

// IsRunning checks if any commands are currently running
func (ce *commandExecutor) IsRunning() bool {
	ce.mu.RLock()
	defer ce.mu.RUnlock()
	return len(ce.running) > 0
}

// GetRunningCommands returns a list of currently running commands
func (ce *commandExecutor) GetRunningCommands() []string {
	ce.mu.RLock()
	defer ce.mu.RUnlock()

	commands := make([]string, 0, len(ce.running))
	for cmd := range ce.running {
		commands = append(commands, cmd)
	}

	return commands
}

// ExecuteWithRealtimeOutput executes a command and streams output in real-time
func (ce *commandExecutor) ExecuteWithRealtimeOutput(command string) ExecutionResult {
	startTime := time.Now()

	// Parse command and arguments
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return ExecutionResult{
			Command:   command,
			ExitCode:  1,
			Error:     fmt.Errorf("empty command"),
			StartTime: startTime,
			Duration:  time.Since(startTime),
		}
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ce.ctx, ce.timeout)
	defer cancel()

	// Create command
	cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
	ce.setupCommand(cmd)

	// Set up pipes for real-time output
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return ExecutionResult{
			Command:   command,
			ExitCode:  1,
			Error:     fmt.Errorf("failed to create stdout pipe: %w", err),
			StartTime: startTime,
			Duration:  time.Since(startTime),
		}
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return ExecutionResult{
			Command:   command,
			ExitCode:  1,
			Error:     fmt.Errorf("failed to create stderr pipe: %w", err),
			StartTime: startTime,
			Duration:  time.Since(startTime),
		}
	}

	// Store running command
	ce.mu.Lock()
	ce.running[command] = cmd
	ce.mu.Unlock()

	// Clean up after execution
	defer func() {
		ce.mu.Lock()
		delete(ce.running, command)
		ce.mu.Unlock()
	}()

	// Start command
	if err := cmd.Start(); err != nil {
		return ExecutionResult{
			Command:   command,
			ExitCode:  1,
			Error:     fmt.Errorf("failed to start command: %w", err),
			StartTime: startTime,
			Duration:  time.Since(startTime),
		}
	}

	// Stream output in real-time
	var wg sync.WaitGroup
	var outputBuffer strings.Builder

	// Stream stdout
	wg.Add(1)
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Println(line)
			outputBuffer.WriteString(line + "\n")
		}
	}()

	// Stream stderr
	wg.Add(1)
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Fprintf(os.Stderr, "%s\n", line)
			outputBuffer.WriteString(line + "\n")
		}
	}()

	// Wait for command to complete
	err = cmd.Wait()
	wg.Wait()

	duration := time.Since(startTime)
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			exitCode = 1
		}
	}

	return ExecutionResult{
		Command:   command,
		ExitCode:  exitCode,
		Output:    outputBuffer.String(),
		Error:     err,
		Duration:  duration,
		StartTime: startTime,
	}
}

// ExecuteCommandsKeepAlive runs commands in background with auto-restart
func (ce *commandExecutor) ExecuteCommandsKeepAlive(commands []string) error {
	fmt.Printf("Starting %d commands in keep-alive mode...\n", len(commands))

	// Start all commands in background
	for i, command := range commands {
		fmt.Printf("[%d/%d] Starting background process: %s\n", i+1, len(commands), command)
		go ce.runKeepAliveProcess(command)
	}

	// Monitor processes and handle stop signal
	go ce.monitorProcesses()

	return nil
}

// runKeepAliveProcess runs a single command with auto-restart
func (ce *commandExecutor) runKeepAliveProcess(command string) {
	restartCount := 0
	maxRestarts := 10 // Prevent infinite restart loops

	// Initialize process status
	ce.mu.Lock()
	ce.processes[command] = &ProcessStatus{
		Command:      command,
		Status:       ProcessStatusStarting,
		StartTime:    time.Now(),
		RestartCount: 0,
	}
	ce.mu.Unlock()

	for restartCount <= maxRestarts {
		// Check if we should stop
		select {
		case <-ce.stopSignal:
			ce.updateProcessStatus(command, ProcessStatusStopped, restartCount)
			return
		default:
		}

		// Parse command and arguments
		parts := strings.Fields(command)
		if len(parts) == 0 {
			ce.updateProcessStatus(command, ProcessStatusFailed, restartCount)
			return
		}

		// Create command without timeout for keep-alive
		cmd := exec.Command(parts[0], parts[1:]...)
		ce.setupCommand(cmd)

		// Store running command
		ce.mu.Lock()
		ce.running[command] = cmd
		ce.mu.Unlock()

		// Update process status
		if cmd.Process != nil {
			ce.updateProcessStatusWithPID(command, ProcessStatusRunning, restartCount, cmd.Process.Pid)
		} else {
			ce.updateProcessStatus(command, ProcessStatusRunning, restartCount)
		}

		fmt.Printf("Process started: %s (attempt %d)\n", command, restartCount+1)

		// Run command and wait for completion
		err := cmd.Run()

		// Clean up from running map
		ce.mu.Lock()
		delete(ce.running, command)
		ce.mu.Unlock()

		if err != nil {
			restartCount++
			fmt.Printf("Process died: %s (error: %v), restart count: %d\n", command, err, restartCount)

			if restartCount <= maxRestarts {
				ce.updateProcessStatus(command, ProcessStatusRestarting, restartCount)

				// Exponential backoff: 5s, 10s, 20s, 40s, max 60s
				delay := time.Duration(5*(1<<uint(restartCount-1))) * time.Second
				if delay > 60*time.Second {
					delay = 60 * time.Second
				}

				fmt.Printf("Restarting in %v...\n", delay)
				time.Sleep(delay)
			} else {
				ce.updateProcessStatus(command, ProcessStatusFailed, restartCount)
				fmt.Printf("Process %s failed permanently after %d restarts\n", command, maxRestarts)
				return
			}
		} else {
			// Process completed normally
			ce.updateProcessStatus(command, ProcessStatusCompleted, restartCount)
			fmt.Printf("Process completed normally: %s\n", command)
			return
		}
	}
}

// updateProcessStatus updates the status of a process
func (ce *commandExecutor) updateProcessStatus(command string, status ProcessStatusType, restartCount int) {
	ce.mu.Lock()
	defer ce.mu.Unlock()

	if process, exists := ce.processes[command]; exists {
		process.Status = status
		process.RestartCount = restartCount
		process.LastRestart = time.Now()
	}
}

// updateProcessStatusWithPID updates the status of a process with PID
func (ce *commandExecutor) updateProcessStatusWithPID(command string, status ProcessStatusType, restartCount, pid int) {
	ce.mu.Lock()
	defer ce.mu.Unlock()

	if process, exists := ce.processes[command]; exists {
		process.Status = status
		process.RestartCount = restartCount
		process.LastRestart = time.Now()
		process.PID = pid
	}
}

// monitorProcesses monitors the health of keep-alive processes
func (ce *commandExecutor) monitorProcesses() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ce.stopSignal:
			return
		case <-ticker.C:
			ce.checkProcessHealth()
		}
	}
}

// checkProcessHealth checks if processes are still healthy
func (ce *commandExecutor) checkProcessHealth() {
	ce.mu.RLock()
	defer ce.mu.RUnlock()

	for command, process := range ce.processes {
		if process.Status == ProcessStatusRunning {
			// Check if process is still in running map
			if _, exists := ce.running[command]; !exists {
				fmt.Printf("Process %s appears to have died unexpectedly\n", command)
			}
		}
	}
}

// GetProcessStatus returns the status of all keep-alive processes
func (ce *commandExecutor) GetProcessStatus() map[string]ProcessStatus {
	ce.mu.RLock()
	defer ce.mu.RUnlock()

	status := make(map[string]ProcessStatus)
	for command, process := range ce.processes {
		status[command] = *process
	}

	return status
}

// Platform-specific methods to be implemented in platform files
// setupCommand configures platform-specific command settings
// terminateProcess terminates a process using platform-appropriate method
