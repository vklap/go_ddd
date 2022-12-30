package ddd

import (
	"context"
)

type eventsReporter interface {
	Events() []Event
}

// CommandHandler interface that should be implemented by command handlers.
type CommandHandler interface {
	Handle(ctx context.Context, command Command) (any, error)
	RollbackCommitter
	eventsReporter
}

// EventHandler interface that should be implemented by event handlers.
type EventHandler interface {
	Handle(ctx context.Context, event Event) error
	RollbackCommitter
	eventsReporter
}
