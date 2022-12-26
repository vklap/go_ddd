package ddd

import (
	"context"
)

type CommandHandler interface {
	handle(ctx context.Context, command Command) (any, error)
	RepositoryCommitterRollbacker
}

type EventHandler interface {
	handle(ctx context.Context, event Event) error
	RepositoryCommitterRollbacker
}
