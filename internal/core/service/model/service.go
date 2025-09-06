package model

import "time"

// Service represents a service configuration
type Service struct {
	ProjectName string
	ServiceName string
	StageName   string
	Commands    []string
}

// ServiceSession represents an active service session
type ServiceSession struct {
	ID        string
	Service   Service
	StartTime time.Time
}

// GetFullName returns project.service.stage format
func (s Service) GetFullName() string {
	return s.ProjectName + "." + s.ServiceName + "." + s.StageName
}

// Validate checks if service configuration is valid
func (s Service) Validate() error {
	if s.ProjectName == "" {
		return ErrEmptyProjectName
	}
	if s.ServiceName == "" {
		return ErrEmptyServiceName
	}
	if s.StageName == "" {
		return ErrEmptyStageName
	}
	if len(s.Commands) == 0 {
		return ErrNoCommands
	}
	return nil
}
