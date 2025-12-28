package events

import (
	"context"

	"go.lumeweb.com/portal-plugin-template/internal/db/models"
	"go.lumeweb.com/portal/core"
	"go.uber.org/zap"
)

// ItemNotificationListener demonstrates how to listen for events and send notifications.
// This is a sample implementation that logs the notification.
type ItemNotificationListener struct {
	ctx    core.Context
	logger *core.Logger
}

// NewItemNotificationListener creates a new listener for item events.
func NewItemNotificationListener(ctx core.Context) *ItemNotificationListener {
	listener := &ItemNotificationListener{
		ctx:    ctx,
		logger: ctx.Logger(),
	}

	// Register the event listener
	ListenForItemCreated(ctx, listener.HandleItemCreated)

	return listener
}

// HandleItemCreated is called when an item is created.
// In a real implementation, this could send an email, push notification, etc.
func (l *ItemNotificationListener) HandleItemCreated(ctx context.Context, item *models.Item) error {
	l.logger.Info("Item notification",
		zap.Uint("item_id", item.ID),
		zap.String("item_name", item.Name),
		zap.String("message", "New item created successfully"),
	)

	// Example: Send notification (pseudocode)
	// notificationSvc := core.GetService[core.NotificationService](l.ctx, core.NOTIFICATION_SERVICE)
	// return notificationSvc.Send(ctx, &core.Notification{
	//     Type: core.NotificationTypeInfo,
	//     Title: "New Item Created",
	//     Body: fmt.Sprintf("Item '%s' has been created", item.Name),
	// })

	return nil
}
