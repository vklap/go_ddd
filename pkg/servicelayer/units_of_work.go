package servicelayer

import (
	"context"
	"log"

	"github.com/vklap.go-ddd/pkg/domain"
)

type CommandUnitOfWork struct {
	handler CommandHandler
}

func (uow *CommandUnitOfWork) Events() []domain.Event {
	result := make([]domain.Event, 0)
	for _, entity := range uow.handler.SavedEntities() {
		result = append(result, entity.Events()...)
	}
	return result
}

func (uow *CommandUnitOfWork) HandleCommand(ctx context.Context, command domain.Command) (result any, err error) {
	if err = command.IsValid(); err != nil {
		return result, err
	}
	result, err = uow.handler.handle(ctx, command)
	if err != nil {
		rollbackErr := uow.handler.Rollback(ctx)
		if rollbackErr != nil {
			log.Printf("failed to rollback CommandUnitOfWork: %v", rollbackErr)
		}
		return result, err
	}
	err = uow.handler.Commit(ctx)
	if err != nil {
		return result, err
	}
	return result, nil
}

type EventUnitOfWork struct {
	handler EventHandler
}

func (uow *EventUnitOfWork) Events() []domain.Event {
	result := make([]domain.Event, 0)
	for _, entity := range uow.handler.SavedEntities() {
		result = append(result, entity.Events()...)
	}
	return result
}

func (uow *EventUnitOfWork) HandleEvent(ctx context.Context, event domain.Event) (err error) {
	err = uow.handler.handle(ctx, event)
	if err != nil {
		rollbackErr := uow.handler.Rollback(ctx)
		if rollbackErr != nil {
			log.Printf("failed to rollback EventUnitOfWork: %v", rollbackErr)
		}
		return err
	}
	err = uow.handler.Commit(ctx)
	if err != nil {
		return err
	}
	return nil
}
