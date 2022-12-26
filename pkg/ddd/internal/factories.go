package internal

import (
	"fmt"
	"github.com/vklap.go-ddd/pkg/ddd"
	"sync"
)

type CommandHandlerFactory struct {
	mu               sync.Mutex
	handlerFactories map[string]ddd.CreateCommandHandler
}

func (f *CommandHandlerFactory) Register(command ddd.Command, factory ddd.CreateCommandHandler) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.handlerFactories[command.CommandName()] = factory
}

func (f *CommandHandlerFactory) CreateHandler(command ddd.Command) (ddd.CommandHandler, error) {
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
	handlerFactories map[string][]ddd.CreateEventHandler
}

func (f *EventHandlersFactory) Register(event ddd.Event, factory ddd.CreateEventHandler) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.handlerFactories[event.EventName()] = append(f.handlerFactories[event.EventName()], factory)
}

func (f *EventHandlersFactory) CreateHandlers(event ddd.Event) ([]ddd.EventHandler, error) {
	factories := f.handlerFactories[event.EventName()]
	handlers := make([]ddd.EventHandler, len(factories))
	for _, factory := range factories {
		handler, err := factory()
		if err != nil {
			return nil, err
		}
		handlers = append(handlers, handler)
	}
	return handlers, nil
}
