package go_ddd

import (
	"context"
	"log"
)

type commandUnitOfWork struct {
	handler CommandHandler
}

func (uow *commandUnitOfWork) HandleCommand(ctx context.Context, command Command) (result any, err error) {
	if err = command.IsValid(); err != nil {
		return result, err
	}
	result, err = uow.handler.Handle(ctx, command)
	if err != nil {
		rollbackErr := uow.handler.Rollback(ctx)
		if rollbackErr != nil {
			log.Printf("failed to rollback commandUnitOfWork: %v", rollbackErr)
		}
		return result, err
	}
	err = uow.handler.Commit(ctx)
	if err != nil {
		return result, err
	}
	return result, nil
}

type eventUnitOfWork struct {
	handler EventHandler
}

func (uow *eventUnitOfWork) HandleEvent(ctx context.Context, event Event) (err error) {
	err = uow.handler.Handle(ctx, event)
	if err != nil {
		rollbackErr := uow.handler.Rollback(ctx)
		if rollbackErr != nil {
			log.Printf("failed to rollback eventUnitOfWork: %v", rollbackErr)
		}
		return err
	}
	err = uow.handler.Commit(ctx)
	if err != nil {
		return err
	}
	return nil
}
