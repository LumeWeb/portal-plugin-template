package models_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.lumeweb.com/portal-plugin-template/internal"
	"go.lumeweb.com/portal-plugin-template/internal/db/migrations"
	"go.lumeweb.com/portal-plugin-template/internal/db/models"
	coreTesting "go.lumeweb.com/portal/core/testing"
	"gorm.io/gorm"
)

func TestMain(m *testing.M) {
	coreTesting.WithDBAndOptions(m,
		coreTesting.WithSQLitePluginMigrations(internal.PLUGIN_NAME, migrations.GetSQLite()),
	)
}

func TestItemModel_Create(t *testing.T) {
	coreTesting.RunTestCaseWithDB(t, func(tb coreTesting.TB, ctx coreTesting.TestContext) {
		db := ctx.DB()

		item := &models.Item{
			Name:        "test-item",
			Description: "test description",
		}

		result := db.Create(item)
		require.NoError(t, result.Error)
		assert.NotZero(t, item.ID)
		assert.NotZero(t, item.CreatedAt)
		assert.NotZero(t, item.UpdatedAt)
		assert.False(t, item.DeletedAt.Valid)
	})
}

func TestItemModel_Read(t *testing.T) {
	coreTesting.RunTestCaseWithDB(t, func(tb coreTesting.TB, ctx coreTesting.TestContext) {
		db := ctx.DB()

		// Create item
		created := &models.Item{
			Name:        "test-item",
			Description: "test description",
		}
		require.NoError(t, db.Create(created).Error)

		// Read item
		var fetched models.Item
		err := db.First(&fetched, created.ID).Error
		require.NoError(t, err)
		assert.Equal(t, created.ID, fetched.ID)
		assert.Equal(t, "test-item", fetched.Name)
		assert.Equal(t, "test description", fetched.Description)
	})
}

func TestItemModel_Update(t *testing.T) {
	coreTesting.RunTestCaseWithDB(t, func(tb coreTesting.TB, ctx coreTesting.TestContext) {
		db := ctx.DB()

		// Create item
		item := &models.Item{
			Name:        "original-name",
			Description: "original description",
		}
		require.NoError(t, db.Create(item).Error)

		// Update item
		item.Name = "updated-name"
		item.Description = "updated description"
		result := db.Save(item)
		require.NoError(t, result.Error)

		// Verify update
		var fetched models.Item
		err := db.First(&fetched, item.ID).Error
		require.NoError(t, err)
		assert.Equal(t, "updated-name", fetched.Name)
		assert.Equal(t, "updated description", fetched.Description)
		assert.True(t, fetched.UpdatedAt.After(item.CreatedAt))
	})
}

func TestItemModel_Delete(t *testing.T) {
	coreTesting.RunTestCaseWithDB(t, func(tb coreTesting.TB, ctx coreTesting.TestContext) {
		db := ctx.DB()

		// Create item
		item := &models.Item{
			Name:        "test-item",
			Description: "test description",
		}
		require.NoError(t, db.Create(item).Error)

		// Delete item (soft delete)
		result := db.Delete(item)
		require.NoError(t, result.Error)

		// Verify soft delete - should not be found with normal query
		var fetched models.Item
		err := db.First(&fetched, item.ID).Error
		assert.Error(t, err)
		assert.Equal(t, gorm.ErrRecordNotFound, err)

		// Verify record still exists with Unscoped
		var deleted models.Item
		err = db.Unscoped().First(&deleted, item.ID).Error
		require.NoError(t, err)
		assert.NotNil(t, deleted.DeletedAt)
	})
}

func TestItemModel_Validation(t *testing.T) {
	coreTesting.RunTestCaseWithDB(t, func(tb coreTesting.TB, ctx coreTesting.TestContext) {
		db := ctx.DB()

		// Test creating item with valid data
		item := &models.Item{
			Name:        "test-item",
			Description: "test description",
		}
		result := db.Create(item)
		assert.NoError(t, result.Error)
		assert.NotZero(t, item.ID)
	})
}

func TestItemModel_List(t *testing.T) {
	coreTesting.RunTestCaseWithDB(t, func(tb coreTesting.TB, ctx coreTesting.TestContext) {
		db := ctx.DB()

		// Create multiple items
		items := []*models.Item{
			{Name: "item-1", Description: "description 1"},
			{Name: "item-2", Description: "description 2"},
			{Name: "item-3", Description: "description 3"},
		}
		for _, item := range items {
			require.NoError(t, db.Create(item).Error)
		}

		// List all items
		var fetched []models.Item
		result := db.Find(&fetched)
		require.NoError(t, result.Error)
		assert.Len(t, fetched, 3)
	})
}

func TestItemModel_QueryByName(t *testing.T) {
	coreTesting.RunTestCaseWithDB(t, func(tb coreTesting.TB, ctx coreTesting.TestContext) {
		db := ctx.DB()

		// Create items
		items := []*models.Item{
			{Name: "apple", Description: "fruit"},
			{Name: "banana", Description: "fruit"},
			{Name: "carrot", Description: "vegetable"},
		}
		for _, item := range items {
			require.NoError(t, db.Create(item).Error)
		}

		// Query by name
		var fetched models.Item
		err := db.Where("name = ?", "apple").First(&fetched).Error
		require.NoError(t, err)
		assert.Equal(t, "apple", fetched.Name)
	})
}

func TestItemModel_Ordering(t *testing.T) {
	coreTesting.RunTestCaseWithDB(t, func(tb coreTesting.TB, ctx coreTesting.TestContext) {
		db := ctx.DB()

		// Create items
		items := []*models.Item{
			{Name: "zebra", Description: "last"},
			{Name: "apple", Description: "first"},
			{Name: "middle", Description: "middle"},
		}
		for _, item := range items {
			require.NoError(t, db.Create(item).Error)
		}

		// Query with ordering
		var fetched []models.Item
		err := db.Order("name asc").Find(&fetched).Error
		require.NoError(t, err)
		assert.Len(t, fetched, 3)
		assert.Equal(t, "apple", fetched[0].Name)
		assert.Equal(t, "middle", fetched[1].Name)
		assert.Equal(t, "zebra", fetched[2].Name)
	})
}

func TestItemModel_Pagination(t *testing.T) {
	coreTesting.RunTestCaseWithDB(t, func(tb coreTesting.TB, ctx coreTesting.TestContext) {
		db := ctx.DB()

		// Create 10 items
		for i := 1; i <= 10; i++ {
			item := &models.Item{
				Name:        "item-" + string(rune('0'+i)),
				Description: "description",
			}
			require.NoError(t, db.Create(item).Error)
		}

		// Query first page (5 items)
		var page1 []models.Item
		err := db.Offset(0).Limit(5).Find(&page1).Error
		require.NoError(t, err)
		assert.Len(t, page1, 5)

		// Query second page (5 items)
		var page2 []models.Item
		err = db.Offset(5).Limit(5).Find(&page2).Error
		require.NoError(t, err)
		assert.Len(t, page2, 5)
	})
}

func TestItemModel_Count(t *testing.T) {
	coreTesting.RunTestCaseWithDB(t, func(tb coreTesting.TB, ctx coreTesting.TestContext) {
		db := ctx.DB()

		// Create items
		for i := 1; i <= 5; i++ {
			item := &models.Item{
				Name:        "item-" + string(rune('0'+i)),
				Description: "description",
			}
			require.NoError(t, db.Create(item).Error)
		}

		// Count items
		var count int64
		err := db.Model(&models.Item{}).Count(&count).Error
		require.NoError(t, err)
		assert.Equal(t, int64(5), count)
	})
}

func TestItemModel_Transaction(t *testing.T) {
	coreTesting.RunTestCaseWithDB(t, func(tb coreTesting.TB, ctx coreTesting.TestContext) {
		db := ctx.DB()

		// Test successful transaction
		err := db.Transaction(func(tx *gorm.DB) error {
			item := &models.Item{
				Name:        "transaction-item",
				Description: "created in transaction",
			}
			return tx.Create(item).Error
		})
		require.NoError(t, err)

		// Verify item was created
		var count int64
		db.Model(&models.Item{}).Where("name = ?", "transaction-item").Count(&count)
		assert.Equal(t, int64(1), count)

		// Test failed transaction (rollback)
		err = db.Transaction(func(tx *gorm.DB) error {
			item := &models.Item{
				Name:        "rollback-item",
				Description: "should be rolled back",
			}
			if err := tx.Create(item).Error; err != nil {
				return err
			}
			// Return error to trigger rollback
			return assert.AnError
		})
		assert.Error(t, err)

		// Verify item was not created
		db.Model(&models.Item{}).Where("name = ?", "rollback-item").Count(&count)
		assert.Equal(t, int64(0), count)
	})
}

func TestItemModel_TextField(t *testing.T) {
	coreTesting.RunTestCaseWithDB(t, func(tb coreTesting.TB, ctx coreTesting.TestContext) {
		db := ctx.DB()

		// Test with long description (text field)
		longDescription := "This is a very long description that exceeds the typical varchar length. " +
			"It should be stored in a text field which can handle much larger amounts of data. " +
			"This is useful for storing detailed information, descriptions, or any other text content " +
			"that might be too large for a standard varchar column."

		item := &models.Item{
			Name:        "long-description-item",
			Description: longDescription,
		}
		result := db.Create(item)
		require.NoError(t, result.Error)

		// Verify long description was stored
		var fetched models.Item
		err := db.First(&fetched, item.ID).Error
		require.NoError(t, err)
		assert.Equal(t, longDescription, fetched.Description)
	})
}
