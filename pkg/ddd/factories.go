package ddd

import (
	"fmt"
	"sync"
)

type CreateCommandHandler func() (CommandHandler, error)
type CreateEventHandler func() (EventHandler, error)

type commandHandlerFactory struct {
	mu               sync.Mutex
	handlerFactories map[string]CreateCommandHandler
}

func (f *commandHandlerFactory) Register(command Command, factory CreateCommandHandler) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.handlerFactories[command.CommandName()] = factory
}

func (f *commandHandlerFactory) CreateHandler(command Command) (CommandHandler, error) {
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

func newCommandHandlerFactory() *commandHandlerFactory {
	return &commandHandlerFactory{
		handlerFactories: make(map[string]CreateCommandHandler),
	}
}

type eventHandlersFactory struct {
	mu               sync.Mutex
	handlerFactories map[string][]CreateEventHandler
}

func (f *eventHandlersFactory) Register(event Event, factory CreateEventHandler) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.handlerFactories[event.EventName()] = append(f.handlerFactories[event.EventName()], factory)
}

func (f *eventHandlersFactory) CreateHandlers(event Event) ([]EventHandler, error) {
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

func newEventHandlersFactory() *eventHandlersFactory {
	return &eventHandlersFactory{
		handlerFactories: make(map[string][]CreateEventHandler),
	}
}
