package service

import (
	pluginCore "go.lumeweb.com/portal-plugin-template/core"
	"go.lumeweb.com/portal-plugin-template/internal/db/models"
	"go.lumeweb.com/portal-plugin-template/internal/events"
	"go.lumeweb.com/portal/core"
	"go.lumeweb.com/portal/db"
	"go.lumeweb.com/queryutil"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var _ pluginCore.ItemService = (*ItemServiceDefault)(nil)

// ItemServiceDefault provides the default implementation of ItemService.
type ItemServiceDefault struct {
	*core.BaseComponent
}

func NewItemService() (core.Service, []core.ContextBuilderOption, error) {
	svc := &ItemServiceDefault{}

	return svc, core.ContextOptions(
		core.ContextWithStartupComponent(svc),
	), nil
}

func (s *ItemServiceDefault) ID() string {
	return pluginCore.ITEM_SERVICE
}

func (s *ItemServiceDefault) ListItems(filters []queryutil.CrudFilter, sorts []queryutil.Sort, pagination queryutil.Pagination) ([]models.Item, int64, error) {
	var items []models.Item
	var total int64

	err := db.RetryableComponentTransaction(s, s.Context(), func(tx *gorm.DB) *gorm.DB {
		query := tx.Model(&models.Item{})

		query = queryutil.ApplyFilters(query, filters, nil)
		query = queryutil.ApplySort(query, sorts)

		if err := query.Count(&total).Error; err != nil {
			return nil
		}

		query = queryutil.ApplyPagination(query, pagination)

		if err := query.Find(&items).Error; err != nil {
			return nil
		}

		return query
	})
	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (s *ItemServiceDefault) CreateItem(ctx core.Context, name string, description string) (*models.Item, error) {
	item := &models.Item{
		Name:        name,
		Description: description,
	}

	if err := db.RetryableComponentTransaction(s, ctx, func(tx *gorm.DB) *gorm.DB {
		return tx.Create(item)
	}); err != nil {
		return nil, err
	}

	// Fire event for item creation
	if err := events.FireItemCreated(ctx, item); err != nil {
		s.Logger().Error("failed to fire item_created event", zap.Error(err))
	}

	return item, nil
}

func (s *ItemServiceDefault) GetItem(ctx core.Context, id uint64) (*models.Item, error) {
	var item models.Item

	if err := db.RetryableComponentTransaction(s, ctx, func(tx *gorm.DB) *gorm.DB {
		return tx.First(&item, id)
	}); err != nil {
		return nil, err
	}

	return &item, nil
}

func (s *ItemServiceDefault) UpdateItem(ctx core.Context, id uint64, name string, description string) error {
	return s.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var item models.Item
		if err := tx.First(&item, id).Error; err != nil {
			return err
		}

		item.Name = name
		item.Description = description

		return tx.Save(&item).Error
	})
}

func (s *ItemServiceDefault) DeleteItem(ctx core.Context, id uint64) error {
	return db.RetryableComponentTransaction(s, ctx, func(tx *gorm.DB) *gorm.DB {
		return tx.WithContext(ctx).Delete(&models.Item{}, id)
	})
}
