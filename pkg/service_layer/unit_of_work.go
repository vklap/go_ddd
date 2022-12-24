package service_layer

import (
	"context"
	"fmt"
	"github.com/vklap.go-ddd/pkg/adapters"
	"github.com/vklap.go-ddd/pkg/domain"
)

type Handler[Command domain.Command, Result *any] interface {
	handle(ctx context.Context, command Command) (*Result, error)
}

type UnitOfWork[Command domain.Command, Result *any, Entity domain.Entity] struct {
	repository adapters.Repository[Entity]
	handler    Handler[Command, Result]
}

func (uow *UnitOfWork[Command, Result, Entity]) Events() []domain.Event {
	result := make([]domain.Event, 0)
	for _, entity := range uow.repository.SavedEntities() {
		result = append(result, entity.Events()...)
	}
	return result
}

func (uow *UnitOfWork[Command, Result, Entity]) HandleCommand(ctx context.Context, command Command) (result *Result, err error) {
	result, err = uow.handler.handle(ctx, command)
	if err != nil {
		rollbackErr := uow.repository.Rollback(ctx)
		if rollbackErr != nil {
			return nil, fmt.Errorf("uow rollback failed: %w", rollbackErr)
		}
		return nil, err
	}
	err = uow.repository.Commit(ctx)
	if err != nil {
		return nil, err
	}
	return result, nil
}
