package adapters

import (
	"context"
	"github.com/vklap.go-ddd/pkg/domain"
)

type CommitterRollbacker interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
	SavedEntities() []domain.Entity
}

type Repository interface {
	CommitterRollbacker
}
