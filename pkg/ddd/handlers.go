package ddd

import (
	"context"
)

// CommandHandler interface that should be implemented by command handlers.
type CommandHandler interface {
	Handle(ctx context.Context, command Command) (any, error)
	Events() []Event
	RollbackCommitter
}

// EventHandler interface that should be implemented by event handlers.
type EventHandler interface {
	Handle(ctx context.Context, event Event) error
	Events() []Event
	RollbackCommitter
}
