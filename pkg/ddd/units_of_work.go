package ddd

import (
	"context"
	"fmt"
)

type commandUnitOfWork struct {
	handler CommandHandler
}

func (uow *commandUnitOfWork) HandleCommand(ctx context.Context, command Command) (result any, err error) {
	result, err = uow.handler.Handle(ctx, command)
	if err != nil {
		rollbackErr := uow.handler.Rollback(ctx)
		if rollbackErr != nil {
			return nil, fmt.Errorf("rollback failed with %q after getting %q", rollbackErr, err)
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
			return fmt.Errorf("rollback failed with %v after getting %v", rollbackErr, err)
		}
		return err
	}
	err = uow.handler.Commit(ctx)
	if err != nil {
		return err
	}
	return nil
}
