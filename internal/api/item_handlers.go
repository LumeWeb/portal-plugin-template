// Package api implements the REST API handlers for the template plugin
package api

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"go.lumeweb.com/httputil"
	"go.lumeweb.com/portal-plugin-template/internal/api/dto"
	"go.lumeweb.com/portal-plugin-template/internal/db/models"
	router "go.lumeweb.com/portal-router"
	"go.lumeweb.com/portal/core"
	queryHttp "go.lumeweb.com/queryutil/http"
	"go.uber.org/zap"
)

// registerItemHandlers sets up all item-related routes with Swagger documentation.
func (a *API) registerItemHandlers(gRouter router.Router, accessSvc core.AccessService) {
	routes := router.DefineRoutes(
		// List items
		router.NewRoute(http.MethodGet, "/api/items", a.listItems,
			router.WithAccess(router.ACCESS_USER_ROLE),
		),

		// Create item
		router.NewRoute(http.MethodPost, "/api/items", a.createItem),

		// Get item
		router.NewRoute(http.MethodGet, "/api/items/:id", a.getItem),

		// Update item
		router.NewRoute(http.MethodPut, "/api/items/:id", a.updateItem),

		// Delete item
		router.NewRoute(http.MethodDelete, "/api/items/:id", a.deleteItem),
	)

	if err := router.RegisterRoutes(gRouter, accessSvc, a.Subdomain(), routes); err != nil {
		a.Logger().Error("failed to register item routes", zap.Error(err))
	}
}

// listItems handles GET /api/items
func (a *API) listItems(c echo.Context) error {
	ctx := httputil.Context(c)

	return queryHttp.ProcessListRequest(
		ctx.Response().Writer, ctx.Request(),
		"items",
		a.itemSvc.ListItems,
		func(item models.Item) dto.ItemResponse {
			return dto.ItemResponse{
				ID:          item.ID,
				Name:        item.Name,
				Description: item.Description,
				CreatedAt:   item.CreatedAt,
				UpdatedAt:   item.UpdatedAt,
			}
		},
	)
}

// createItem handles POST /api/items
func (a *API) createItem(c echo.Context) error {
	ctx := httputil.Context(c)

	var req dto.CreateItemRequest
	item, ok := httputil.DecodeAndValidateRequest(ctx, &req)
	if !ok {
		return nil
	}

	return httputil.EncodeResponse(ctx, item, &dto.ItemResponse{})
}

// getItem handles GET /api/items/{id}
func (a *API) getItem(c echo.Context) error {
	ctx := httputil.Context(c)

	var req dto.GetItemRequest
	_, ok := httputil.DecodeAndValidateRequest(ctx, &req)
	if !ok {
		return nil
	}

	item, err := a.itemSvc.GetItem(a.Context(), uint64(req.ID))
	if err != nil {
		return ctx.Error(err, http.StatusNotFound)
	}

	return httputil.EncodeResponse(ctx, item, &dto.ItemResponse{})
}

// updateItem handles PUT /api/items/{id}
func (a *API) updateItem(c echo.Context) error {
	ctx := httputil.Context(c)

	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		return ctx.Error(err, http.StatusBadRequest)
	}

	var req dto.UpdateItemRequest
	item, ok := httputil.DecodeAndValidateRequest(ctx, &req)
	if !ok {
		return nil
	}

	err = a.itemSvc.UpdateItem(a.Context(), id, item.Name, item.Description)
	if err != nil {
		return ctx.Error(err, http.StatusInternalServerError)
	}

	return httputil.EncodeResponse(ctx, &dto.StatusResponse{Status: "updated"}, &dto.StatusResponse{})
}

// deleteItem handles DELETE /api/items/{id}
func (a *API) deleteItem(c echo.Context) error {
	ctx := httputil.Context(c)

	var req dto.GetItemRequest
	_, ok := httputil.DecodeAndValidateRequest(ctx, &req)
	if !ok {
		return nil
	}

	if err := a.itemSvc.DeleteItem(a.Context(), uint64(req.ID)); err != nil {
		return ctx.Error(err, http.StatusInternalServerError)
	}

	return httputil.EncodeResponse(ctx, &dto.StatusResponse{Status: "deleted"}, &dto.StatusResponse{})
}
