package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

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

// ConfigManager manages configuration using Viper
type ConfigManager struct {
	v        *viper.Viper
	mu       sync.RWMutex
	filePath string

	// 변경 콜백 (옵션)
	onChangeCallbacks []func()
}

var (
	instance *ConfigManager
	once     sync.Once
)

// GetInstance returns the singleton instance of ConfigManager
func GetInstance() *ConfigManager {
	once.Do(func() {
		instance = newConfigManager()
		instance.initialize()
	})
	return instance
}

// newConfigManager creates a new ConfigManager instance
func newConfigManager() *ConfigManager {
	home := os.Getenv("HOME")
	if home == "" {
		home = os.Getenv("USERPROFILE") // Windows support
	}
	filePath := filepath.Join(home, "hama-shell.yaml")

	v := viper.New()
	v.SetConfigFile(filePath)
	v.SetConfigType("yaml")

	return &ConfigManager{
		v:                 v,
		filePath:          filePath,
		onChangeCallbacks: make([]func(), 0),
	}
}

// initialize sets up the configuration manager
func (cm *ConfigManager) initialize() {
	// 파일이 없으면 빈 config로 초기화
	if !cm.FileExists() {
		cm.v.Set("projects", make(map[string]interface{}))
	} else {
		// 파일이 있으면 로드 (한 번만)
		if err := cm.v.ReadInConfig(); err != nil {
			// 에러가 있어도 빈 config로 초기화
			cm.v.Set("projects", make(map[string]interface{}))
		}
	}

	// 파일 변경 감지 설정 (옵션)
	cm.v.WatchConfig()
	cm.v.OnConfigChange(func(e fsnotify.Event) {
		cm.mu.RLock()
		callbacks := cm.onChangeCallbacks
		cm.mu.RUnlock()

		for _, callback := range callbacks {
			callback()
		}
	})
}

// Load reads configuration - Viper에서는 이미 메모리에 있으므로 no-op
func (cm *ConfigManager) Load() error {
	// Viper는 자동으로 캐싱하므로 별도 로드 불필요
	// 하위 호환성을 위해 메서드는 유지
	return nil
}

// Reload forces a configuration reload from file
func (cm *ConfigManager) Reload() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	return cm.v.ReadInConfig()
}

// Save writes the current configuration to file
func (cm *ConfigManager) Save() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// 디렉토리가 없으면 생성
	dir := filepath.Dir(cm.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Viper의 WriteConfigAs를 사용하여 파일 저장
	return cm.v.WriteConfigAs(cm.filePath)
}

// GetConfig returns the current configuration
func (cm *ConfigManager) GetConfig() *Config {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	var config Config
	if err := cm.v.Unmarshal(&config); err != nil {
		return &Config{
			Projects: make(map[string]*Project),
		}
	}

	if config.Projects == nil {
		config.Projects = make(map[string]*Project)
	}

	// nil Services 맵 초기화
	for _, project := range config.Projects {
		if project.Services == nil {
			project.Services = make(map[string]*Service)
		}
	}

	return &config
}

// AddProject adds a new project to the configuration
func (cm *ConfigManager) AddProject(projectName string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	projects := cm.v.GetStringMap("projects")
	if projects == nil {
		projects = make(map[string]interface{})
	}

	if _, exists := projects[projectName]; exists {
		return fmt.Errorf("project '%s' already exists", projectName)
	}

	projects[projectName] = map[string]interface{}{
		"services": make(map[string]interface{}),
	}

	cm.v.Set("projects", projects)
	return nil
}

// AddService adds a service to an existing project
func (cm *ConfigManager) AddService(projectName, serviceName string, commands []string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	projectPath := fmt.Sprintf("projects.%s", projectName)
	if !cm.v.IsSet(projectPath) {
		return fmt.Errorf("project '%s' not found", projectName)
	}

	servicesPath := fmt.Sprintf("%s.services", projectPath)
	services := cm.v.GetStringMap(servicesPath)
	if services == nil {
		services = make(map[string]interface{})
	}

	if _, exists := services[serviceName]; exists {
		return fmt.Errorf("service '%s' already exists in project '%s'", serviceName, projectName)
	}

	services[serviceName] = map[string]interface{}{
		"commands": commands,
	}

	cm.v.Set(servicesPath, services)
	return nil
}

// AppendToService appends commands to an existing service
func (cm *ConfigManager) AppendToService(projectName, serviceName string, commands []string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	commandsPath := fmt.Sprintf("projects.%s.services.%s.commands", projectName, serviceName)

	if !cm.v.IsSet(commandsPath) {
		return fmt.Errorf("service '%s' not found in project '%s'", serviceName, projectName)
	}

	existingCommands := cm.v.GetStringSlice(commandsPath)
	existingCommands = append(existingCommands, commands...)
	cm.v.Set(commandsPath, existingCommands)

	return nil
}

// UpdateService updates an existing service
func (cm *ConfigManager) UpdateService(projectName, serviceName string, commands []string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	commandsPath := fmt.Sprintf("projects.%s.services.%s.commands", projectName, serviceName)

	if !cm.v.IsSet(commandsPath) {
		return fmt.Errorf("service '%s' not found in project '%s'", serviceName, projectName)
	}

	cm.v.Set(commandsPath, commands)
	return nil
}

// DeleteProject removes a project from configuration
func (cm *ConfigManager) DeleteProject(projectName string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	projectPath := fmt.Sprintf("projects.%s", projectName)

	if !cm.v.IsSet(projectPath) {
		return fmt.Errorf("project '%s' not found", projectName)
	}

	// Viper에서 키 삭제
	projects := cm.v.GetStringMap("projects")
	delete(projects, projectName)
	cm.v.Set("projects", projects)

	return nil
}

// DeleteService removes a service from a project
func (cm *ConfigManager) DeleteService(projectName, serviceName string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	servicesPath := fmt.Sprintf("projects.%s.services", projectName)

	if !cm.v.IsSet(servicesPath) {
		return fmt.Errorf("project '%s' not found", projectName)
	}

	services := cm.v.GetStringMap(servicesPath)
	if services == nil {
		return fmt.Errorf("no services in project '%s'", projectName)
	}

	if _, exists := services[serviceName]; !exists {
		return fmt.Errorf("service '%s' not found in project '%s'", serviceName, projectName)
	}

	delete(services, serviceName)
	cm.v.Set(servicesPath, services)

	return nil
}

// FileExists checks if the configuration file exists
func (cm *ConfigManager) FileExists() bool {
	_, err := os.Stat(cm.filePath)
	return err == nil
}

// GetFilePath returns the configuration file path
func (cm *ConfigManager) GetFilePath() string {
	return cm.filePath
}

// OnConfigChange registers a callback for configuration changes
func (cm *ConfigManager) OnConfigChange(callback func()) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.onChangeCallbacks = append(cm.onChangeCallbacks, callback)
}

// Get returns a raw value from configuration (Viper style)
func (cm *ConfigManager) Get(key string) interface{} {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return cm.v.Get(key)
}

// GetString returns a string value from configuration
func (cm *ConfigManager) GetString(key string) string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return cm.v.GetString(key)
}

// GetStringSlice returns a string slice from configuration
func (cm *ConfigManager) GetStringSlice(key string) []string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return cm.v.GetStringSlice(key)
}

// GetStringMap returns a string map from configuration
func (cm *ConfigManager) GetStringMap(key string) map[string]interface{} {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return cm.v.GetStringMap(key)
}

// IsSet checks if a key is set in the configuration
func (cm *ConfigManager) IsSet(key string) bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return cm.v.IsSet(key)
}
