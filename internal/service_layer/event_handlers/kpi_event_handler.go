package event_handlers

import (
	"context"
	"fmt"
	"github.com/vklap/go_ddd/internal/adapters"
	"github.com/vklap/go_ddd/internal/domain/command_model"
	"github.com/vklap/go_ddd/pkg/ddd"
)

// KPIEventHandler implements ddd.EventHandler.
type KPIEventHandler struct {
	pubSubClient adapters.PubSubClient
}

// NewKPIEventHandler is a constructor function to be used by the Bootstrapper.
func NewKPIEventHandler(pubSubClient adapters.PubSubClient) *KPIEventHandler {
	return &KPIEventHandler{pubSubClient: pubSubClient}
}

// Handle manages the business logic flow, and is the glue between the Domain and the Adapters.
func (h *KPIEventHandler) Handle(ctx context.Context, event ddd.Event) error {
	e, ok := event.(*command_model.KPIEvent)
	if ok == false {
		panic(fmt.Sprintf("failed to handle KPI event: want %T, got %T", &command_model.KPIEvent{}, e))
	}
	return h.pubSubClient.NotifyKPIService(ctx, e)
}

// Commit is responsible for committing the changes performed by the Handle method, such as
// committing a database transaction managed by the repository.
// This method is being called by the framework, so it should not be called from within the Handle method.
func (h *KPIEventHandler) Commit(ctx context.Context) error {
	return h.pubSubClient.Commit(ctx)
}

// Rollback is responsible to rollback changes performed by the Handle method, such as
// rollback a database transaction managed by the repository.
// This method is being called by the framework, so it should not be called from within the Handle method.
func (h *KPIEventHandler) Rollback(ctx context.Context) error {
	return h.pubSubClient.Rollback(ctx)
}

// Events reports about events.
// These events will be handled by the DDD framework if appropriate event handlers were registered by the bootstrapper.
// This method is being called by the framework, so it should not be called from within the Handle method.
func (h *KPIEventHandler) Events() []ddd.Event {
	return nil
}

var _ ddd.EventHandler = (*KPIEventHandler)(nil)
