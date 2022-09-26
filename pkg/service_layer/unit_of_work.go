package service_layer

import (
	"github.com/vklap.go-ddd/pkg/adapters"
	"github.com/vklap.go-ddd/pkg/domain"
)

type Handler[TEntity domain.Entity, TResult any] interface {
	handle(command *domain.Command, repo adapters.Repository[TEntity]) (*TResult, error)
}

type UnitOfWork[TEntity domain.Entity, TResult any] struct {
	Command   *domain.Command
	GetEvents []domain.Event
	entities  []TEntity
}

func (uow *UnitOfWork[TEntity, TRes]) Execute(h Handler[TEntity, TRes], repo adapters.Repository[TEntity]) (*TRes, error) {
	result, err := h.handle(uow.Command, repo)
	if err != nil {
		return nil, err
	}
	savedEntities, err := repo.Commit()
	if err != nil {
		return nil, err
	}
	uow.entities = append(uow.entities, savedEntities...)
	return result, nil
}
