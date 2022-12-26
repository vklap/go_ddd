package entrypoints

import (
	"context"
	"github.com/vklap.go-ddd/pkg/domain"
	sl "github.com/vklap.go-ddd/pkg/servicelayer"
)

type Bootstrapper struct {
	commandHandlerFactory *sl.CommandHandlerFactory
	eventHandlersFactory  *sl.EventHandlersFactory
}

func New() *Bootstrapper {
	return &Bootstrapper{
		commandHandlerFactory: &sl.CommandHandlerFactory{},
		eventHandlersFactory:  &sl.EventHandlersFactory{},
	}
}

func (b *Bootstrapper) RegisterCommandHandlerFactory(command domain.Command, factory sl.CreateCommandHandler) {
	b.commandHandlerFactory.Register(command, factory)
}

func (b *Bootstrapper) RegisterEventHandlerFactory(event domain.Event, factory sl.CreateEventHandler) {
	b.eventHandlersFactory.Register(event, factory)
}

func (b *Bootstrapper) HandleCommand(ctx context.Context, command domain.Command) (any, error) {
	mb := sl.NewMessageBus(b.commandHandlerFactory, b.eventHandlersFactory)
	result, err := mb.Handle(ctx, command)
	return result, err
}
