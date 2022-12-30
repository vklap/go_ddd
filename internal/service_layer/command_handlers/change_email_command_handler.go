package command_handlers

import (
	"context"
	"fmt"
	"github.com/vklap/go_ddd/internal/adapters"
	"github.com/vklap/go_ddd/internal/domain/command_model"
	"github.com/vklap/go_ddd/pkg/ddd"
)

// ChangeEmailCommandHandler implements ddd.CommandHandler.
type ChangeEmailCommandHandler struct {
	repository adapters.Repository
	events     []ddd.Event
}

// NewChangeEmailCommandHandler is a constructor function to be used by the Bootstrapper.
func NewChangeEmailCommandHandler(repository adapters.Repository) *ChangeEmailCommandHandler {
	return &ChangeEmailCommandHandler{repository: repository}
}

// Handle manages the business logic flow, and is the glue between the Domain and the Adapters.
func (h *ChangeEmailCommandHandler) Handle(ctx context.Context, command ddd.Command) (any, error) {
	changeEmailCommand, ok := command.(*command_model.ChangeEmailCommand)
	if ok == false {
		return nil, fmt.Errorf("ChangeEmailCommandHandler expects a command of type %T", changeEmailCommand)
	}

	// No need to call changeEmailCommand.IsValid() - as it's being called by the framework.

	user, err := h.repository.GetUserById(ctx, changeEmailCommand.UserID)
	if err != nil {
		return nil, err
	}

	user.SetEmail(changeEmailCommand.NewEmail)

	if err = h.repository.SaveUser(ctx, user); err != nil {
		return nil, err
	}

	// This is where Domain events are being registered by the handler,
	// so they can eventually be dispatched to event handlers - such as:
	// EmailChangedEventHandler
	h.events = append(h.events, user.Events()...)

	return nil, nil
}

// Commit is responsible for committing the changes performed by the Handle method, such as
// committing a database transaction managed by the repository.
// This method is being called by the framework, so it should not be called from within the Handle method.
func (h *ChangeEmailCommandHandler) Commit(ctx context.Context) error {
	return h.repository.Commit(ctx)
}

// Rollback is responsible to rollback changes performed by the Handle method, such as
// rollback a database transaction managed by the repository.
// This method is being called by the framework, so it should not be called from within the Handle method.
func (h *ChangeEmailCommandHandler) Rollback(ctx context.Context) error {
	return h.repository.Rollback(ctx)
}

// Events reports about events.
// These events will be handled by the DDD framework if appropriate event handlers were registered by the bootstrapper.
// This method is being called by the framework, so it should not be called from within the Handle method.
func (h *ChangeEmailCommandHandler) Events() []ddd.Event {
	return h.events
}

var _ ddd.CommandHandler = (*ChangeEmailCommandHandler)(nil)
