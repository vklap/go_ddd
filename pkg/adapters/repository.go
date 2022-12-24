package adapters

import (
	"context"
	"github.com/vklap.go-ddd/pkg/domain"
)

type Repository[Entity domain.Entity] interface {
	Find(ctx context.Context, id string) (Entity, error)
	Save(ctx context.Context, entity Entity) error
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
	SavedEntities() []Entity
}
