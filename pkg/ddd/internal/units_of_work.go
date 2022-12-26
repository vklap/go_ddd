package internal

import (
	"context"
	"github.com/vklap.go-ddd/pkg/ddd"
	"log"
)

type CommandUnitOfWork struct {
	handler ddd.CommandHandler
}

func (uow *CommandUnitOfWork) Events() []ddd.Event {
	result := make([]ddd.Event, 0)
	for _, entity := range uow.handler.SavedEntities() {
		result = append(result, entity.Events()...)
	}
	return result
}

func (uow *CommandUnitOfWork) HandleCommand(ctx context.Context, command ddd.Command) (result any, err error) {
	if err = command.IsValid(); err != nil {
		return result, err
	}
	result, err = uow.handler.Handle(ctx, command)
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
	handler ddd.EventHandler
}

func (uow *EventUnitOfWork) Events() []ddd.Event {
	result := make([]ddd.Event, 0)
	for _, entity := range uow.handler.SavedEntities() {
		result = append(result, entity.Events()...)
	}
	return result
}

func (uow *EventUnitOfWork) HandleEvent(ctx context.Context, event ddd.Event) (err error) {
	err = uow.handler.Handle(ctx, event)
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
