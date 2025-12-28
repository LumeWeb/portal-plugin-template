package config

import "go.lumeweb.com/portal/config"

// Verify ProtocolConfig implements config.ProtocolConfig interface
var _ config.ProtocolConfig = (*ProtocolConfig)(nil)

// ProtocolConfig defines configuration for the protocol
type ProtocolConfig struct {
	StoragePath  string `config:"storage_path"`  // Path to store protocol data
	MaxItems     int    `config:"max_items"`     // Maximum number of items to store
	CacheEnabled bool   `config:"cache_enabled"` // Whether to enable caching
}

// Defaults provides default configuration values
func (c ProtocolConfig) Defaults() map[string]any {
	return map[string]any{
		"storage_path":  "data/template",
		"max_items":     1000,
		"cache_enabled": true,
	}
}
