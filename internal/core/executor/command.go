package executor

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"
)

// CommandExecutor handles execution of shell commands
type CommandExecutor struct {
	timeout time.Duration
	ctx     context.Context
	cancel  context.CancelFunc
	mu      sync.RWMutex
	running map[string]*exec.Cmd
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

// NewCommandExecutor creates a new command executor with default timeout
func NewCommandExecutor(timeout time.Duration) *CommandExecutor {
	ctx, cancel := context.WithCancel(context.Background())
	return &CommandExecutor{
		timeout: timeout,
		ctx:     ctx,
		cancel:  cancel,
		running: make(map[string]*exec.Cmd),
	}
}

// ToDo : 테스트 코드 만들기 :-) & 컴포넌트 마인드맵 그리기 (mermaid)
// ToDo : 테스트 코드 만들기 :-) & 컴포넌트 마인드맵 그리기 (mermaid)
// ToDo : 테스트 코드 만들기 :-) & 컴포넌트 마인드맵 그리기 (mermaid)
// ToDo : 테스트 코드 만들기 :-) & 컴포넌트 마인드맵 그리기 (mermaid)
// ExecuteCommands runs a list of commands sequentially
func (ce *CommandExecutor) ExecuteCommands(commands []string) ([]ExecutionResult, error) {
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
func (ce *CommandExecutor) ExecuteCommand(command string) ExecutionResult {
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
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

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
func (ce *CommandExecutor) ExecuteCommandInteractive(command string) ExecutionResult {
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
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

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
func (ce *CommandExecutor) ExecuteCommandsParallel(commands []string, maxConcurrent int) ([]ExecutionResult, error) {
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
func (ce *CommandExecutor) StopAll() error {
	ce.mu.RLock()
	commands := make([]*exec.Cmd, 0, len(ce.running))
	for _, cmd := range ce.running {
		commands = append(commands, cmd)
	}
	ce.mu.RUnlock()

	var errors []error
	for _, cmd := range commands {
		if cmd.Process != nil {
			if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
				errors = append(errors, err)
			}
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
func (ce *CommandExecutor) IsRunning() bool {
	ce.mu.RLock()
	defer ce.mu.RUnlock()
	return len(ce.running) > 0
}

// GetRunningCommands returns a list of currently running commands
func (ce *CommandExecutor) GetRunningCommands() []string {
	ce.mu.RLock()
	defer ce.mu.RUnlock()

	commands := make([]string, 0, len(ce.running))
	for cmd := range ce.running {
		commands = append(commands, cmd)
	}

	return commands
}

// ExecuteWithRealtimeOutput executes a command and streams output in real-time
func (ce *CommandExecutor) ExecuteWithRealtimeOutput(command string) ExecutionResult {
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
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

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
