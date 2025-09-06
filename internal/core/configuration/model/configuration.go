package model

// Service represents a service configuration with commands
type Service struct {
	Commands []string `yaml:"commands"`
}

// Project represents a project configuration with services
type Project struct {
	Services map[string]*Service `yaml:"services"`
}

// Config represents the main configuration structure
type Config struct {
	Projects map[string]*Project `yaml:"projects"`
}

// ConfigOperation represents a configuration operation
type ConfigOperation struct {
	ProjectName string
	ServiceName string
	Commands    []string
}

// ConfigView represents configuration view data
type ConfigView struct {
	FilePath string
	Content  interface{}
	Exists   bool
	IsEmpty  bool
}