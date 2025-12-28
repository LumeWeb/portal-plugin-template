package events_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.lumeweb.com/portal-plugin-template/internal/db/models"
	"go.lumeweb.com/portal-plugin-template/internal/events"
	coreTesting "go.lumeweb.com/portal/core/testing"
)

func TestItemNotificationListener_HandleItemCreated(t *testing.T) {
	coreTesting.RunTestCase(t, func(tb coreTesting.TB, ctx coreTesting.TestContext) {
		// Create listener
		listener := events.NewItemNotificationListener(ctx)

		// Create test item
		testItem := &models.Item{
			Name:        "test-item",
			Description: "test description",
		}

		// Test handling item created event
		err := listener.HandleItemCreated(context.Background(), testItem)
		assert.NoError(t, err)
	})
}

func TestFireItemCreated(t *testing.T) {
	coreTesting.RunTestCase(t, func(tb coreTesting.TB, ctx coreTesting.TestContext) {
		// Create test item
		testItem := &models.Item{
			Name:        "test-item",
			Description: "test description",
		}

		// Fire the event
		err := events.FireItemCreated(ctx, testItem)
		assert.NoError(t, err)
	})
}

func TestFireItemCreated_WithListener(t *testing.T) {
	coreTesting.RunTestCase(t, func(tb coreTesting.TB, ctx coreTesting.TestContext) {
		// Track if listener was called
		listenerCalled := false
		var receivedItem *models.Item

		// Register a custom listener
		events.ListenForItemCreated(ctx, func(ctx context.Context, item *models.Item) error {
			listenerCalled = true
			receivedItem = item
			return nil
		})

		// Create test item
		testItem := &models.Item{
			Name:        "test-item",
			Description: "test description",
		}

		// Fire the event
		err := events.FireItemCreated(ctx, testItem)
		assert.NoError(t, err)

		// Verify listener was called
		assert.True(t, listenerCalled)
		assert.NotNil(t, receivedItem)
		assert.Equal(t, testItem.Name, receivedItem.Name)
	})
}

func TestFireItemCreated_MultipleListeners(t *testing.T) {
	coreTesting.RunTestCase(t, func(tb coreTesting.TB, ctx coreTesting.TestContext) {
		// Track listener calls
		listener1Called := false
		listener2Called := false

		// Register multiple listeners
		events.ListenForItemCreated(ctx, func(ctx context.Context, item *models.Item) error {
			listener1Called = true
			return nil
		})

		events.ListenForItemCreated(ctx, func(ctx context.Context, item *models.Item) error {
			listener2Called = true
			return nil
		})

		// Create test item
		testItem := &models.Item{
			Name:        "test-item",
			Description: "test description",
		}

		// Fire the event
		err := events.FireItemCreated(ctx, testItem)
		assert.NoError(t, err)

		// Verify all listeners were called
		assert.True(t, listener1Called)
		assert.True(t, listener2Called)
	})
}

func TestFireItemCreated_ListenerError(t *testing.T) {
	coreTesting.RunTestCase(t, func(tb coreTesting.TB, ctx coreTesting.TestContext) {
		// Register a listener that returns an error
		events.ListenForItemCreated(ctx, func(ctx context.Context, item *models.Item) error {
			return assert.AnError
		})

		// Create test item
		testItem := &models.Item{
			Name:        "test-item",
			Description: "test description",
		}

		// Fire the event - listener errors are logged but don't fail the event
		err := events.FireItemCreated(ctx, testItem)
		// The event system may return an error if a listener fails
		// This is expected behavior - listeners should handle their own errors
		assert.Error(t, err)
	})
}

func TestItemNotificationListener_Context(t *testing.T) {
	coreTesting.RunTestCase(t, func(tb coreTesting.TB, ctx coreTesting.TestContext) {
		// Create listener
		listener := events.NewItemNotificationListener(ctx)

		// Verify listener has access to context
		assert.NotNil(t, listener)
	})
}
