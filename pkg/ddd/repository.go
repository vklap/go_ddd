package ddd

import (
	"context"
)

type RepositoryCommitterRollbacker interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
	SavedEntities() []Entity
}
