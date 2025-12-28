// Package api implements the core API functionality for the template plugin.
// It provides the main API structure and configuration for HTTP routing.
package api

import (
	_ "embed"

	pluginCore "go.lumeweb.com/portal-plugin-template/core"
	"go.lumeweb.com/portal-plugin-template/internal/protocol"
	"go.lumeweb.com/portal-plugin-template/internal/webapp"
	router "go.lumeweb.com/portal-router"
	"go.lumeweb.com/portal/core"
)

// Verify API implements the core.API interface
var _ core.API = (*API)(nil)

// API represents the main API structure for the template plugin.
// It holds references to core services and configuration needed for operation.
type API struct {
	*core.BaseComponent
	itemSvc pluginCore.ItemService
	proto   *protocol.Protocol
}

func (a *API) OpenAPIInfo() router.APIInfoDefinition {
	return router.APIInfo().
		Title("Template Plugin API").
		Version("1.0.0").
		Description("API for managing template items").
		Contact("support@lumeweb.com", "LumeWeb Support").
		License("MIT", "https://opensource.org/licenses/MIT")
}

// NewAPI creates a new instance of the template plugin API.
// It returns the API instance and context builder options needed for initialization.
func NewAPI() (core.API, []core.ContextBuilderOption, error) {
	instance := &API{}

	// Define startup configuration
	opts := core.ContextOptions(
		core.ContextWithStartupFunc(func(ctx core.Context) error {
			// Initialize API with context and services
			instance.BaseComponent = core.NewBaseComponent(ctx)
			instance.itemSvc = core.GetService[pluginCore.ItemService](ctx, pluginCore.ITEM_SERVICE)
			return nil
		}),
	)

	return instance, opts, nil
}

// Configure sets up all routes and middleware for the API.
// It handles both API endpoints and static file serving for the webapp.
func (a *API) Configure(gRouter router.Router, accessSvc core.AccessService) error {

	// Register all API routes with access control
	a.registerItemHandlers(gRouter, accessSvc)

	router.MustDefaultStaticSetup(gRouter, webapp.Files)

	return nil
}
