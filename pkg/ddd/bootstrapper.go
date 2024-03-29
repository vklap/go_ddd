// Package ddd is a Domain-Driven Design (DDD) Framework.
//
// To install, `run go get github.com/vklap/go_ddd`.
//
// Documentation can be found on github (https://github.com/vklap/go_ddd/blob/main/README.md)
// and on pkg.go.dev (https://pkg.go.dev/github.com/vklap/go_ddd).
package ddd

import (
	"context"
)

// Bootstrapper registers command and event handlers.
type Bootstrapper struct {
	commandHandlerFactory *commandHandlerFactory
	eventHandlersFactory  *eventHandlersFactory
}

// NewBootstrapper initializes a new Bootstrapper instance.
func NewBootstrapper() *Bootstrapper {
	return &Bootstrapper{
		commandHandlerFactory: newCommandHandlerFactory(),
		eventHandlersFactory:  newEventHandlersFactory(),
	}
}

// RegisterCommandHandlerFactory registers a function based create command handler factory.
func (b *Bootstrapper) RegisterCommandHandlerFactory(command Command, factory CreateCommandHandler) {
	b.commandHandlerFactory.Register(command, factory)
}

// RegisterEventHandlerFactory registers a function based create event handler factory.
func (b *Bootstrapper) RegisterEventHandlerFactory(event Event, factory CreateEventHandler) {
	b.eventHandlersFactory.Register(event, factory)
}

// HandleCommand is the facade handling Domain Commands, that will eventually trigger registered Event handlers.
func (b *Bootstrapper) HandleCommand(ctx context.Context, command Command) (any, error) {
	mb := newMessageBus(b.commandHandlerFactory, b.eventHandlersFactory)
	result, err := mb.Publish(ctx, command)
	return result, err
}
