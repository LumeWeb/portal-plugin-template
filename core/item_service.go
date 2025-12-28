package core

import (
	"go.lumeweb.com/portal-plugin-template/internal/db/models"
	"go.lumeweb.com/portal/core"
	"go.lumeweb.com/queryutil"
)

const ITEM_SERVICE = "item"

// ItemService defines the interface for managing items.
type ItemService interface {
	core.Service
	ListItems(filters []queryutil.CrudFilter, sorts []queryutil.Sort, pagination queryutil.Pagination) ([]models.Item, int64, error)
	CreateItem(ctx core.Context, name string, description string) (*models.Item, error)
	GetItem(ctx core.Context, id uint64) (*models.Item, error)
	UpdateItem(ctx core.Context, id uint64, name string, description string) error
	DeleteItem(ctx core.Context, id uint64) error
}
