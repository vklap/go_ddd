package servicelayer

import (
	"fmt"
	"github.com/vklap.go-ddd/pkg/domain"
	"sync"
)

type CreateCommandHandler func() (CommandHandler, error)
type CreateEventHandler func() (EventHandler, error)

type CommandHandlerFactory struct {
	mu               sync.Mutex
	handlerFactories map[string]CreateCommandHandler
}

func (f *CommandHandlerFactory) Register(command domain.Command, factory CreateCommandHandler) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.handlerFactories[command.CommandName()] = factory
}

func (f *CommandHandlerFactory) CreateHandler(command domain.Command) (CommandHandler, error) {
	factory, ok := f.handlerFactories[command.CommandName()]
	if ok == false {
		panic(fmt.Sprintf("command is not registered in executor: %q", command.CommandName()))
	}
	handler, err := factory()
	if err != nil {
		return nil, err
	}
	return handler, nil
}

type EventHandlersFactory struct {
	mu               sync.Mutex
	handlerFactories map[string][]CreateEventHandler
}

func (f *EventHandlersFactory) Register(event domain.Event, factory CreateEventHandler) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.handlerFactories[event.EventName()] = append(f.handlerFactories[event.EventName()], factory)
}

func (f *EventHandlersFactory) CreateHandlers(event domain.Event) ([]EventHandler, error) {
	factories := f.handlerFactories[event.EventName()]
	handlers := make([]EventHandler, len(factories))
	for _, factory := range factories {
		handler, err := factory()
		if err != nil {
			return nil, err
		}
		handlers = append(handlers, handler)
	}
	return handlers, nil
}
