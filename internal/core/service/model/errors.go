package model

import "errors"

// Domain errors
var (
	ErrEmptyProjectName = errors.New("project name cannot be empty")
	ErrEmptyServiceName = errors.New("service name cannot be empty")
	ErrNoCommands       = errors.New("service must have at least one command")
	ErrServiceNotFound  = errors.New("service not found")
)