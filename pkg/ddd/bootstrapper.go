package ddd

import (
	"context"
	"github.com/vklap.go-ddd/pkg/ddd/internal"
)

type Bootstrapper struct {
	commandHandlerFactory *internal.CommandHandlerFactory
	eventHandlersFactory  *internal.EventHandlersFactory
}

func New() *Bootstrapper {
	return &Bootstrapper{
		commandHandlerFactory: &internal.CommandHandlerFactory{},
		eventHandlersFactory:  &internal.EventHandlersFactory{},
	}
}

func (b *Bootstrapper) RegisterCommandHandlerFactory(command Command, factory CreateCommandHandler) {
	b.commandHandlerFactory.Register(command, factory)
}

func (b *Bootstrapper) RegisterEventHandlerFactory(event Event, factory CreateEventHandler) {
	b.eventHandlersFactory.Register(event, factory)
}

func (b *Bootstrapper) HandleCommand(ctx context.Context, command Command) (any, error) {
	mb := internal.NewMessageBus(b.commandHandlerFactory, b.eventHandlersFactory)
	result, err := mb.Handle(ctx, command)
	return result, err
}
