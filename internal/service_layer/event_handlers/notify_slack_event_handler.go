package event_handlers

import (
	"context"
	"fmt"
	"github.com/vklap/go_ddd/internal/adapters"
	"github.com/vklap/go_ddd/internal/domain/command_model"
	"github.com/vklap/go_ddd/pkg/ddd"
)

// NotifySlackEventHandler implements ddd.EventHandler.
type NotifySlackEventHandler struct {
	pubSubClient adapters.PubSubClient
}

// NewNotifySlackEventHandler is a constructor function to be used by the Bootstrapper.
func NewNotifySlackEventHandler(pubSubClient adapters.PubSubClient) *NotifySlackEventHandler {
	return &NotifySlackEventHandler{pubSubClient: pubSubClient}
}

// Handle manages the business logic flow, and is the glue between the Domain and the Adapters.
func (h *NotifySlackEventHandler) Handle(ctx context.Context, event ddd.Event) error {
	e, ok := event.(*command_model.NotifySlackEvent)
	if ok == false {
		panic(fmt.Sprintf("failed to handle email changed: want %T, got %T", &command_model.NotifySlackEvent{}, e))
	}
	return h.pubSubClient.NotifySlack(ctx, e.Message)
}

// Commit is responsible for committing the changes performed by the Handle method, such as
// committing a database transaction managed by the repository.
// This method is being called by the framework, so it should not be called from within the Handle method.
func (h *NotifySlackEventHandler) Commit(ctx context.Context) error {
	return h.pubSubClient.Commit(ctx)
}

// Rollback is responsible to rollback changes performed by the Handle method, such as
// rollback a database transaction managed by the repository.
// This method is being called by the framework, so it should not be called from within the Handle method.
func (h *NotifySlackEventHandler) Rollback(ctx context.Context) error {
	return h.pubSubClient.Rollback(ctx)
}

// Events reports about events.
// These events will be handled by the DDD framework if appropriate event handlers were registered by the bootstrapper.
// This method is being called by the framework, so it should not be called from within the Handle method.
func (h *NotifySlackEventHandler) Events() []ddd.Event {
	return nil
}

var _ ddd.EventHandler = (*NotifySlackEventHandler)(nil)
