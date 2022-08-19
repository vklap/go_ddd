package ddd

import (
	"context"
)

type CommandHandler interface {
	Handle(ctx context.Context, command Command) (any, error)
	RepositoryCommitterRollbacker
}

type EventHandler interface {
	Handle(ctx context.Context, event Event) error
	RepositoryCommitterRollbacker
}
