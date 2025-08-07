package executor

import (
	"fmt"
	"os/exec"
	"runtime"
	"sync"
	"time"
)

// Executor defines the interface for process execution and management
type Executor interface {
	// Run starts a command associated with the given key
	Run(key, command string) error

	// StopAll terminates all running processes
	StopAll() error

	// StopByKey terminates all processes associated with the given key
	StopByKey(key string) error

	// GetStatus returns the current status of all processes
	GetStatus() map[string][]*ProcessInfo
}

// ProcessInfo contains information about a running process
type ProcessInfo struct {
	Command   string
	PID       int
	StartTime int64 // Unix timestamp
	Key       string
}

// executor is the main implementation of the Executor interface
type executor struct {
	registry sync.Map // key: string, value: []*ProcessCommand
	mu       sync.RWMutex
	manager  processManager
}

// New creates a new Executor instance
func New() Executor {
	return &executor{
		manager: newProcessManager(),
	}
}

// Run starts a new process and associates it with the given key
func (e *executor) Run(key, command string) error {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", command)
	} else {
		cmd = exec.Command("sh", "-c", command)
	}

	// Platform-specific command setup
	e.manager.setupCommand(cmd)

	// Start the command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	// Create process record
	proc := &ProcessCommand{
		Cmd:       command,
		Process:   cmd.Process,
		Key:       key,
		StartTime: time.Now(),
	}

	// Add to registry
	e.addToRegistry(key, proc)

	// Monitor process in background
	go e.waitForProcess(key, proc, cmd)

	return nil
}

// StopAll terminates all running processes
func (e *executor) StopAll() error {
	var errors []error

	e.registry.Range(func(key, value interface{}) bool {
		if err := e.StopByKey(key.(string)); err != nil {
			errors = append(errors, fmt.Errorf("failed to stop processes for key %s: %w", key, err))
		}
		return true
	})

	if len(errors) > 0 {
		return fmt.Errorf("failed to stop %d process groups", len(errors))
	}

	return nil
}

// StopByKey terminates all processes associated with the given key
func (e *executor) StopByKey(key string) error {
	value, exists := e.registry.Load(key)
	if !exists {
		return nil // No processes for this key
	}

	processes := value.([]*ProcessCommand)
	var errors []error

	for _, proc := range processes {
		if proc.Process != nil {
			if err := e.manager.terminateProcess(proc.Process); err != nil {
				errors = append(errors, fmt.Errorf("failed to terminate PID %d: %w", proc.Process.Pid, err))
			}
		}
	}

	// Remove from registry
	e.registry.Delete(key)

	if len(errors) > 0 {
		return fmt.Errorf("failed to terminate %d processes", len(errors))
	}

	return nil
}

// GetStatus returns the current status of all processes
func (e *executor) GetStatus() map[string][]*ProcessInfo {
	status := make(map[string][]*ProcessInfo)

	e.registry.Range(func(key, value interface{}) bool {
		processes := value.([]*ProcessCommand)
		infos := make([]*ProcessInfo, 0, len(processes))

		for _, proc := range processes {
			if proc.Process != nil {
				info := &ProcessInfo{
					Command:   proc.Cmd,
					PID:       proc.Process.Pid,
					StartTime: proc.StartTime.Unix(),
					Key:       proc.Key,
				}
				infos = append(infos, info)
			}
		}

		if len(infos) > 0 {
			status[key.(string)] = infos
		}

		return true
	})

	return status
}

// addToRegistry adds a process to the registry
func (e *executor) addToRegistry(key string, proc *ProcessCommand) {
	value, _ := e.registry.LoadOrStore(key, []*ProcessCommand{})

	e.mu.Lock()
	defer e.mu.Unlock()

	processes := value.([]*ProcessCommand)
	processes = append(processes, proc)
	e.registry.Store(key, processes)
}

// removeFromRegistry removes a specific process from the registry
func (e *executor) removeFromRegistry(key string, pid int) {
	value, exists := e.registry.Load(key)
	if !exists {
		return
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	processes := value.([]*ProcessCommand)
	filtered := make([]*ProcessCommand, 0, len(processes))

	for _, proc := range processes {
		if proc.Process == nil || proc.Process.Pid != pid {
			filtered = append(filtered, proc)
		}
	}

	if len(filtered) > 0 {
		e.registry.Store(key, filtered)
	} else {
		e.registry.Delete(key)
	}
}

// waitForProcess monitors a process and cleans up when it exits
func (e *executor) waitForProcess(key string, proc *ProcessCommand, cmd *exec.Cmd) {
	// Wait for process to complete
	cmd.Wait()

	// Remove from registry
	if proc.Process != nil {
		e.removeFromRegistry(key, proc.Process.Pid)
	}
}
