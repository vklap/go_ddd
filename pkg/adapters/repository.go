package adapters

import "github.com/vklap.go-ddd/pkg/domain"

type EntityFinder[TEntity domain.Entity] interface {
	Find(id domain.Identifier) (*TEntity, error)
}

type EntitySaver[TEntity domain.Entity] interface {
	Save(entity *TEntity) error
}

type Committer[TEntity domain.Entity] interface {
	Commit() (savedEntities []TEntity, err error)
}

type Rollbacker interface {
	Rollback() error
}

type Repository[TEntity domain.Entity] interface {
	Committer[TEntity]
	Rollbacker
}
