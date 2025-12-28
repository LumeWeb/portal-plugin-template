package events

import (
	"context"

	"go.lumeweb.com/portal-plugin-template/internal/db/models"
	"go.lumeweb.com/portal/core"
)

// Event constants for the template plugin.
const (
	EventItemCreated = "template.item_created"
)

// ItemCreatedEvent represents the data sent when an item is created.
type ItemCreatedEvent struct {
	Item *models.Item
	Ctx  context.Context
}

// ListenForItemCreated registers a handler to run when an item is created.
func ListenForItemCreated(ctx core.Context, handler func(context.Context, *models.Item) error) {
	core.Listen[ItemCreatedEvent](ctx, EventItemCreated, func(e *core.CoreEvent[ItemCreatedEvent]) error {
		return handler(e.Data.Ctx, e.Data.Item)
	})
}

// FireItemCreated fires the item_created event.
func FireItemCreated(ctx core.Context, item *models.Item) error {
	return core.Fire(ctx, EventItemCreated, &ItemCreatedEvent{
		Item: item,
		Ctx:  ctx.GetContext(),
	})
}
