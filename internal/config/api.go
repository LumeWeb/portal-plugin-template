package config

import "go.lumeweb.com/portal/config"

// Verify APIConfig implements config.APIConfig interface
var _ config.APIConfig = (*APIConfig)(nil)

// APIConfig defines API-specific configuration options
type APIConfig struct {
	ItemsPerPage int `config:"items_per_page"` // Number of items to return per page
	SearchLimit  int `config:"search_limit"`   // Maximum number of search results
}

// Defaults provides default configuration values for API settings
func (a APIConfig) Defaults() map[string]any {
	return map[string]any{
		"items_per_page": 10,  // Default page size
		"search_limit":   100, // Default search results limit
	}
}
