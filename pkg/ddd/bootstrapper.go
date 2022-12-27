package ddd

import (
	"context"
)

type Bootstrapper struct {
	commandHandlerFactory *commandHandlerFactory
	eventHandlersFactory  *eventHandlersFactory
}

func NewBootstrapper() *Bootstrapper {
	return &Bootstrapper{
		commandHandlerFactory: newCommandHandlerFactory(),
		eventHandlersFactory:  newEventHandlersFactory(),
	}
}

func (b *Bootstrapper) RegisterCommandHandlerFactory(command Command, factory CreateCommandHandler) {
	b.commandHandlerFactory.Register(command, factory)
}

func (b *Bootstrapper) RegisterEventHandlerFactory(event Event, factory CreateEventHandler) {
	b.eventHandlersFactory.Register(event, factory)
}

func (b *Bootstrapper) HandleCommand(ctx context.Context, command Command) (any, error) {
	mb := newMessageBus(b.commandHandlerFactory, b.eventHandlersFactory)
	result, err := mb.Handle(ctx, command)
	return result, err
}
