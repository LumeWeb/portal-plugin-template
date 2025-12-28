package service_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	pluginCore "go.lumeweb.com/portal-plugin-template/core"
	"go.lumeweb.com/portal-plugin-template/internal"
	"go.lumeweb.com/portal-plugin-template/internal/db/migrations"
	"go.lumeweb.com/portal-plugin-template/internal/db/models"
	"go.lumeweb.com/portal-plugin-template/internal/service"
	"go.lumeweb.com/portal/core"
	coreTesting "go.lumeweb.com/portal/core/testing"
	"go.lumeweb.com/queryutil"
)

func TestMain(m *testing.M) {
	coreTesting.WithDBAndOptions(m,
		coreTesting.WithSQLitePluginMigrations(internal.PLUGIN_NAME, migrations.GetSQLite()),
	)
}

func TestItemService_ListItems(t *testing.T) {
	coreTesting.RunTestCaseWithDB(t, func(tb coreTesting.TB, ctx coreTesting.TestContext) {
		itemSvc := core.GetService[pluginCore.ItemService](ctx, pluginCore.ITEM_SERVICE)

		// Create test items
		item1 := &models.Item{Name: "item-1", Description: "description 1"}
		item2 := &models.Item{Name: "item-2", Description: "description 2"}
		item3 := &models.Item{Name: "item-3", Description: "description 3"}

		require.NoError(t, ctx.DB().Create(item1).Error)
		require.NoError(t, ctx.DB().Create(item2).Error)
		require.NoError(t, ctx.DB().Create(item3).Error)

		// Test listing all items
		items, total, err := itemSvc.ListItems(nil, nil, queryutil.Pagination{Start: 0, End: 10})
		require.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, items, 3)

		// Test with sorting
		sorts := []queryutil.Sort{{Field: "name", Order: "desc"}}
		items, total, err = itemSvc.ListItems(nil, sorts, queryutil.Pagination{Start: 0, End: 10})
		require.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Equal(t, "item-3", items[0].Name)
	},
		coreTesting.WithService(pluginCore.ITEM_SERVICE, service.NewItemService),
	)
}

func TestItemService_CreateItem(t *testing.T) {
	coreTesting.RunTestCaseWithDB(t, func(tb coreTesting.TB, ctx coreTesting.TestContext) {
		itemSvc := core.GetService[pluginCore.ItemService](ctx, pluginCore.ITEM_SERVICE)

		// Test creating an item
		item, err := itemSvc.CreateItem(ctx, "test-item", "test description")
		require.NoError(t, err)
		assert.NotNil(t, item)
		assert.Equal(t, "test-item", item.Name)
		assert.Equal(t, "test description", item.Description)
		assert.NotZero(t, item.ID)
		assert.NotZero(t, item.CreatedAt)
		assert.NotZero(t, item.UpdatedAt)

		// Verify item was saved to database
		var fetched models.Item
		err = ctx.DB().First(&fetched, item.ID).Error
		require.NoError(t, err)
		assert.Equal(t, "test-item", fetched.Name)
	},
		coreTesting.WithService(pluginCore.ITEM_SERVICE, service.NewItemService),
	)
}

func TestItemService_GetItem(t *testing.T) {
	coreTesting.RunTestCaseWithDB(t, func(tb coreTesting.TB, ctx coreTesting.TestContext) {
		itemSvc := core.GetService[pluginCore.ItemService](ctx, pluginCore.ITEM_SERVICE)

		// Create a test item
		created, err := itemSvc.CreateItem(ctx, "test-item", "test description")
		require.NoError(t, err)

		// Test getting existing item
		item, err := itemSvc.GetItem(ctx, uint64(created.ID))
		require.NoError(t, err)
		assert.NotNil(t, item)
		assert.Equal(t, created.ID, item.ID)
		assert.Equal(t, "test-item", item.Name)

		// Test getting non-existent item
		_, err = itemSvc.GetItem(ctx, 999)
		assert.Error(t, err)
	},
		coreTesting.WithService(pluginCore.ITEM_SERVICE, service.NewItemService),
	)
}

func TestItemService_UpdateItem(t *testing.T) {
	coreTesting.RunTestCaseWithDB(t, func(tb coreTesting.TB, ctx coreTesting.TestContext) {
		itemSvc := core.GetService[pluginCore.ItemService](ctx, pluginCore.ITEM_SERVICE)

		// Create a test item
		created, err := itemSvc.CreateItem(ctx, "original-name", "original description")
		require.NoError(t, err)

		// Test updating the item
		err = itemSvc.UpdateItem(ctx, uint64(created.ID), "updated-name", "updated description")
		require.NoError(t, err)

		// Verify update
		item, err := itemSvc.GetItem(ctx, uint64(created.ID))
		require.NoError(t, err)
		assert.Equal(t, "updated-name", item.Name)
		assert.Equal(t, "updated description", item.Description)

		// Test updating non-existent item
		err = itemSvc.UpdateItem(ctx, 999, "name", "description")
		assert.Error(t, err)
	},
		coreTesting.WithService(pluginCore.ITEM_SERVICE, service.NewItemService),
	)
}

func TestItemService_DeleteItem(t *testing.T) {
	coreTesting.RunTestCaseWithDB(t, func(tb coreTesting.TB, ctx coreTesting.TestContext) {
		itemSvc := core.GetService[pluginCore.ItemService](ctx, pluginCore.ITEM_SERVICE)

		// Create a test item
		created, err := itemSvc.CreateItem(ctx, "test-item", "test description")
		require.NoError(t, err)

		// Test deleting the item
		err = itemSvc.DeleteItem(ctx, uint64(created.ID))
		require.NoError(t, err)

		// Verify deletion (soft delete - should not be found)
		_, err = itemSvc.GetItem(ctx, uint64(created.ID))
		assert.Error(t, err)

		// Verify record still exists with DeletedAt set
		var deleted models.Item
		err = ctx.DB().Unscoped().First(&deleted, created.ID).Error
		require.NoError(t, err)
		assert.NotNil(t, deleted.DeletedAt)

		// Test deleting non-existent item (GORM doesn't error on soft delete of non-existent)
		err = itemSvc.DeleteItem(ctx, 999)
		assert.NoError(t, err)
	},
		coreTesting.WithService(pluginCore.ITEM_SERVICE, service.NewItemService),
	)
}
