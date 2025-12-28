package config

import "go.lumeweb.com/portal/config"

// Verify ServiceConfig implements config.ServiceConfig interface
var _ config.ServiceConfig = (*ServiceConfig)(nil)

// ServiceConfig defines configuration for the service
type ServiceConfig struct {
	MaxItems int `config:"max_items"` // Maximum number of items to store
}

// Defaults provides default configuration values
func (c ServiceConfig) Defaults() map[string]any {
	return map[string]any{
		"max_items": 1000,
	}
}
