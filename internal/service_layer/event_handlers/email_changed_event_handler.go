package event_handlers

import (
	"context"
	"fmt"
	"github.com/vklap/go_ddd/internal/adapters"
	"github.com/vklap/go_ddd/internal/domain/command_model"
	"github.com/vklap/go_ddd/pkg/ddd"
)

// EmailChangedEventHandler implements ddd.EventHandler.
type EmailChangedEventHandler struct {
	emailClient adapters.EmailClient
}

// NewEmailChangedEventHandler is a constructor function to be used by the Bootstrapper.
func NewEmailChangedEventHandler(emailClient adapters.EmailClient) *EmailChangedEventHandler {
	return &EmailChangedEventHandler{emailClient: emailClient}
}

// Handle manages the business logic flow, and is the glue between the Domain and the Adapters.
func (h *EmailChangedEventHandler) Handle(ctx context.Context, event ddd.Event) error {
	emailChangedEvent, ok := event.(*command_model.EmailChangedEvent)
	if ok == false {
		panic(fmt.Sprintf("failed to handle email changed: want %T, got %T", &command_model.EmailChangedEvent{}, emailChangedEvent))
	}
	from := "noreply@example.com"
	to := emailChangedEvent.OriginalEmail
	title := "NewEmail Changed Notification"
	message := fmt.Sprintf("Your email was changed to %v", emailChangedEvent.NewEmail)
	return h.emailClient.SendEmail(from, to, title, message)
}

// Commit is responsible for committing the changes performed by the Handle method, such as
// committing a database transaction managed by the repository.
// This method is being called by the framework, so it should not be called from within the Handle method.
func (h *EmailChangedEventHandler) Commit(ctx context.Context) error {
	return nil
}

// Rollback is responsible to rollback changes performed by the Handle method, such as
// rollback a database transaction managed by the repository.
// This method is being called by the framework, so it should not be called from within the Handle method.
func (h *EmailChangedEventHandler) Rollback(ctx context.Context) error {
	return nil
}

// Events reports about events.
// These events will be handled by the DDD framework if appropriate event handlers were registered by the bootstrapper.
// This method is being called by the framework, so it should not be called from within the Handle method.
func (h *EmailChangedEventHandler) Events() []ddd.Event {
	return nil
}

var _ ddd.EventHandler = (*EmailChangedEventHandler)(nil)
