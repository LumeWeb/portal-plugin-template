package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	pluginCore "go.lumeweb.com/portal-plugin-template/core"
	"go.lumeweb.com/portal-plugin-template/internal"
	"go.lumeweb.com/portal-plugin-template/internal/api"
	"go.lumeweb.com/portal-plugin-template/internal/api/dto"
	"go.lumeweb.com/portal-plugin-template/internal/db/models"
	"go.lumeweb.com/portal-plugin-template/internal/service"
	"go.lumeweb.com/portal/core"
	coreTesting "go.lumeweb.com/portal/core/testing"
	"gorm.io/gorm"
)

func TestMain(m *testing.M) {
	coreTesting.RunTests(m, coreTesting.TestMainOpts{
		CustomSetup: func() {
			coreTesting.AddGlobalTestContextOptions(
				coreTesting.WithAPI(internal.PLUGIN_NAME, api.NewAPI),
				coreTesting.WithAPIID(internal.PLUGIN_NAME),
				coreTesting.WithMockServiceFactory(pluginCore.ITEM_SERVICE, service.NewMockItemService),
			)
		},
	})
}

// Helper to create a test item model
func createTestItem(id uint, name, description string) *models.Item {
	return &models.Item{
		Model: gorm.Model{
			ID:        id,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name:        name,
		Description: description,
	}
}

// TestAPI_ListItems tests the GET /api/items endpoint.
func TestAPI_ListItems(t *testing.T) {
	coreTesting.RunTestCase(t, func(tb coreTesting.TB, ctx coreTesting.TestContext) {
		mockItemSvc := core.GetService[*service.MockItemService](ctx, pluginCore.ITEM_SERVICE)

		testItems := []models.Item{
			*createTestItem(1, "test-item-1", "test description 1"),
			*createTestItem(2, "test-item-2", "test description 2"),
		}

		mockItemSvc.EXPECT().ListItems(mock.Anything, mock.Anything, mock.Anything).Return(
			testItems, int64(len(testItems)), nil,
		).Maybe()

		req := ctx.NewAPIRequest(http.MethodGet, "/api/items", nil)
		resp := httptest.NewRecorder()
		ctx.Router().ServeHTTP(resp, req)

		if resp.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.Code)
		}

		var result struct {
			Data  []dto.ItemResponse `json:"data"`
			Total int64              `json:"total"`
		}
		if err := json.Unmarshal(resp.Body.Bytes(), &result); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if len(result.Data) == 0 {
			t.Error("expected at least one item")
		}
	})
}

// TestAPI_GetItem tests the GET /api/items/{id} endpoint.
func TestAPI_GetItem(t *testing.T) {
	coreTesting.RunTestCase(t, func(tb coreTesting.TB, ctx coreTesting.TestContext) {
		mockItemSvc := core.GetService[*service.MockItemService](ctx, pluginCore.ITEM_SERVICE)

		testItem := createTestItem(1, "test-item", "test description")

		mockItemSvc.EXPECT().GetItem(mock.Anything, uint64(1)).Return(
			testItem, nil,
		).Maybe()

		req := ctx.NewAPIRequest(http.MethodGet, "/api/items/1", nil)
		resp := httptest.NewRecorder()
		ctx.Router().ServeHTTP(resp, req)

		if resp.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d: %s", resp.Code, resp.Body.String())
		}

		var result dto.ItemResponse
		if err := json.Unmarshal(resp.Body.Bytes(), &result); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if result.Name != "test-item" {
			t.Errorf("expected name 'test-item', got '%s'", result.Name)
		}
	})
}

// TestAPI_GetItem_NotFound tests the GET /api/items/{id} endpoint when item doesn't exist.
func TestAPI_GetItem_NotFound(t *testing.T) {
	coreTesting.RunTestCase(t, func(tb coreTesting.TB, ctx coreTesting.TestContext) {
		mockItemSvc := core.GetService[*service.MockItemService](ctx, pluginCore.ITEM_SERVICE)

		mockItemSvc.EXPECT().GetItem(mock.Anything, uint64(999)).Return(
			nil, gorm.ErrRecordNotFound,
		).Maybe()

		req := ctx.NewAPIRequest(http.MethodGet, "/api/items/999", nil)
		resp := httptest.NewRecorder()
		ctx.Router().ServeHTTP(resp, req)

		if resp.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d: %s", resp.Code, resp.Body.String())
		}
	})
}

// TestAPI_CreateItem tests the POST /api/items endpoint.
func TestAPI_CreateItem(t *testing.T) {
	coreTesting.RunTestCase(t, func(tb coreTesting.TB, ctx coreTesting.TestContext) {
		mockItemSvc := core.GetService[*service.MockItemService](ctx, pluginCore.ITEM_SERVICE)

		newItem := createTestItem(1, "new-item", "new description")

		mockItemSvc.EXPECT().CreateItem(mock.Anything, "new-item", "new description").Return(
			newItem, nil,
		).Maybe()

		body := dto.CreateItemRequest{
			Name:        "new-item",
			Description: "new description",
		}
		bodyBytes, _ := json.Marshal(body)

		req := ctx.NewAPIRequest(http.MethodPost, "/api/items", bodyBytes)
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		ctx.Router().ServeHTTP(resp, req)

		if resp.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d: %s", resp.Code, resp.Body.String())
		}

		var result dto.ItemResponse
		if err := json.Unmarshal(resp.Body.Bytes(), &result); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if result.Name != "new-item" {
			t.Errorf("expected name 'new-item', got '%s'", result.Name)
		}
	})
}

// TestAPI_UpdateItem tests the PUT /api/items/{id} endpoint.
func TestAPI_UpdateItem(t *testing.T) {
	coreTesting.RunTestCase(t, func(tb coreTesting.TB, ctx coreTesting.TestContext) {
		mockItemSvc := core.GetService[*service.MockItemService](ctx, pluginCore.ITEM_SERVICE)

		mockItemSvc.EXPECT().UpdateItem(mock.Anything, uint64(1), "updated-item", "updated description").Return(
			nil,
		).Maybe()

		body := dto.UpdateItemRequest{
			Name:        "updated-item",
			Description: "updated description",
		}
		bodyBytes, _ := json.Marshal(body)

		req := ctx.NewAPIRequest(http.MethodPut, "/api/items/1", bodyBytes)
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		ctx.Router().ServeHTTP(resp, req)

		if resp.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d: %s", resp.Code, resp.Body.String())
		}
	})
}

// TestAPI_DeleteItem tests the DELETE /api/items/{id} endpoint.
func TestAPI_DeleteItem(t *testing.T) {
	coreTesting.RunTestCase(t, func(tb coreTesting.TB, ctx coreTesting.TestContext) {
		mockItemSvc := core.GetService[*service.MockItemService](ctx, pluginCore.ITEM_SERVICE)

		mockItemSvc.EXPECT().DeleteItem(mock.Anything, uint64(1)).Return(
			nil,
		).Maybe()

		req := ctx.NewAPIRequest(http.MethodDelete, "/api/items/1", nil)
		resp := httptest.NewRecorder()
		ctx.Router().ServeHTTP(resp, req)

		if resp.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d: %s", resp.Code, resp.Body.String())
		}
	})
}

// TestAPI_Validation tests request validation for the item API.
func TestAPI_Validation(t *testing.T) {
	coreTesting.RunTestCase(t, func(tb coreTesting.TB, ctx coreTesting.TestContext) {
		// Test: Create item with empty name (should fail validation)
		body := dto.CreateItemRequest{
			Name:        "",
			Description: "test",
		}
		bodyBytes, _ := json.Marshal(body)

		req := ctx.NewAPIRequest(http.MethodPost, "/api/items", bodyBytes)
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		ctx.Router().ServeHTTP(resp, req)

		if resp.Code != http.StatusUnprocessableEntity {
			t.Errorf("expected status 422, got %d: %s", resp.Code, resp.Body.String())
		}
	})
}
