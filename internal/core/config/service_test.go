package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestService_List(t *testing.T) {
	// GIVEN a config service with a sample configuration
	service := NewService()
	cfg := &Config{
		Projects: map[string]Project{
			"myapp": {
				Description: "Test application",
				Stages: map[string]Stage{
					"dev": {
						Services: map[string]Service{
							"api": {
								Description: "API service",
								Commands:    []string{"echo api"},
							},
							"db": {
								Description: "Database",
								Commands:    []string{"echo db"},
							},
						},
					},
					"prod": {
						Services: map[string]Service{
							"api": {
								Description: "Production API",
								Commands:    []string{"echo prod-api"},
							},
						},
					},
				},
			},
			"monitoring": {
				Description: "Monitoring stack",
				Stages: map[string]Stage{
					"dev": {
						Services: map[string]Service{
							"prometheus": {
								Description: "Metrics collector",
								Commands:    []string{"echo prometheus"},
							},
						},
					},
				},
			},
		},
	}

	// WHEN we list all targets
	targets := service.List(cfg)

	// THEN we should get all services in project.stage.service format
	assert.Len(t, targets, 4)
	assert.Contains(t, targets, "myapp.dev.api")
	assert.Contains(t, targets, "myapp.dev.db")
	assert.Contains(t, targets, "myapp.prod.api")
	assert.Contains(t, targets, "monitoring.dev.prometheus")
}

func TestService_List_EmptyConfig(t *testing.T) {
	// GIVEN a config service with nil and empty configs
	service := NewService()

	t.Run("nil config", func(t *testing.T) {
		// WHEN we list targets from nil config
		targets := service.List(nil)

		// THEN we should get empty list
		assert.Empty(t, targets)
	})

	t.Run("empty config", func(t *testing.T) {
		// WHEN we list targets from empty config
		targets := service.List(&Config{})

		// THEN we should get empty list
		assert.Empty(t, targets)
	})
}

func TestService_ResolveTarget(t *testing.T) {
	tests := []struct {
		name         string
		target       string
		config       *Config
		wantErr      bool
		errContains  string
		wantCommands []string
	}{
		{
			name:   "valid target resolution",
			target: "myapp.dev.api",
			config: &Config{
				Projects: map[string]Project{
					"myapp": {
						Stages: map[string]Stage{
							"dev": {
								Services: map[string]Service{
									"api": {
										Description: "API service",
										Commands:    []string{"ssh api", "start server"},
									},
								},
							},
						},
					},
				},
			},
			wantErr:      false,
			wantCommands: []string{"ssh api", "start server"},
		},
		{
			name:        "invalid target format - too few parts",
			target:      "myapp.dev",
			config:      &Config{},
			wantErr:     true,
			errContains: "invalid target format",
		},
		{
			name:        "invalid target format - too many parts",
			target:      "myapp.dev.api.extra",
			config:      &Config{},
			wantErr:     true,
			errContains: "invalid target format",
		},
		{
			name:        "nil config",
			target:      "myapp.dev.api",
			config:      nil,
			wantErr:     true,
			errContains: "config is nil",
		},
		{
			name:        "project not found",
			target:      "unknown.dev.api",
			config:      &Config{Projects: map[string]Project{}},
			wantErr:     true,
			errContains: "project not found: unknown",
		},
		{
			name:   "stage not found",
			target: "myapp.unknown.api",
			config: &Config{
				Projects: map[string]Project{
					"myapp": {
						Stages: map[string]Stage{
							"dev": {
								Services: map[string]Service{},
							},
						},
					},
				},
			},
			wantErr:     true,
			errContains: "stage not found: myapp.unknown",
		},
		{
			name:   "service not found",
			target: "myapp.dev.unknown",
			config: &Config{
				Projects: map[string]Project{
					"myapp": {
						Stages: map[string]Stage{
							"dev": {
								Services: map[string]Service{
									"api": {
										Commands: []string{"echo api"},
									},
								},
							},
						},
					},
				},
			},
			wantErr:     true,
			errContains: "service not found: myapp.dev.unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// GIVEN a config service
			service := NewService()

			// WHEN we resolve the target
			svc, err := service.ResolveTarget(tt.target, tt.config)

			// THEN we should get expected result
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, svc)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, svc)
				assert.Equal(t, tt.wantCommands, svc.Commands)
			}
		})
	}
}

func TestService_Load(t *testing.T) {
	t.Run("empty config path", func(t *testing.T) {
		// GIVEN a config service
		service := NewService()

		// WHEN we load with empty path
		cfg, err := service.Load("")

		// THEN it should return error
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "config path is required")
		assert.Nil(t, cfg)
	})
}
