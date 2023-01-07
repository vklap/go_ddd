package command_handlers

import (
	"context"
	"fmt"
	"github.com/vklap/go_ddd/internal/adapters"
	"github.com/vklap/go_ddd/internal/domain/command_model"
	"github.com/vklap/go_ddd/pkg/ddd"
)

// SaveUserCommandHandler implements ddd.CommandHandler.
type SaveUserCommandHandler struct {
	repository adapters.Repository
	events     []ddd.Event
}

// NewSaveUserCommandHandler is a constructor function to be used by the Bootstrapper.
func NewSaveUserCommandHandler(repository adapters.Repository) *SaveUserCommandHandler {
	return &SaveUserCommandHandler{repository: repository}
}

// Handle manages the business logic flow, and is the glue between the Domain and the Adapters.
func (h *SaveUserCommandHandler) Handle(ctx context.Context, command ddd.Command) (any, error) {
	saveUserCommand, ok := command.(*command_model.SaveUserCommand)
	if ok == false {
		return nil, fmt.Errorf("SaveUserCommandHandler expects a command of type %T", saveUserCommand)
	}

	// No need to call saveUserCommand.IsValid() - as it's being called by the framework.

	// Delegate fetching data to the repository, which belongs to the Adapters Layer.
	user, err := h.repository.GetUserById(ctx, saveUserCommand.UserID)
	if err != nil {
		return nil, err
	}

	// Delegate updating the email to the user, which is a Domain Entity.
	// The SetEmail method is responsible to detect if a new email was set,
	// and if so, then it will record an EmailSetEvent.
	user.SetEmail(saveUserCommand.Email)

	// Delegate storing data to the repository.
	if err = h.repository.SaveUser(ctx, user); err != nil {
		return nil, err
	}

	// This is where Domain events are being registered,
	// so they can eventually be dispatched to event handlers (if they exist).
	// In our use case the events will be dispatched to the EmailChangedEventHandler.
	h.events = append(h.events, user.Events()...)

	return nil, nil
}

// Commit is responsible for committing the changes performed by the Handle method, such as
// committing a database transaction managed by the repository.
// This method is being called by the framework, so it should not be called from within the Handle method.
func (h *SaveUserCommandHandler) Commit(ctx context.Context) error {
	return h.repository.Commit(ctx)
}

// Rollback is responsible to rollback changes performed by the Handle method, such as
// rollback a database transaction managed by the repository.
// This method is being called by the framework, so it should not be called from within the Handle method.
func (h *SaveUserCommandHandler) Rollback(ctx context.Context) error {
	return h.repository.Rollback(ctx)
}

// Events reports about events.
// These events will be handled by the DDD framework if appropriate event handlers were registered by the bootstrapper.
// This method is being called by the framework, so it should not be called from within the Handle method.
func (h *SaveUserCommandHandler) Events() []ddd.Event {
	return h.events
}

var _ ddd.CommandHandler = (*SaveUserCommandHandler)(nil)
