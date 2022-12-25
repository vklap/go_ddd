package service_layer

import (
	"sync"
)

type CommandHandlerFactory func() (CommandHandler, error)
type EventHandlerFactory func() (EventHandler, error)

type CommandDispatcher struct {
	mu               sync.Mutex
	handlerFactories map[string]CommandHandlerFactory
}

type EventDispatcher struct {
	mu               sync.Mutex
	handlerFactories map[string][]EventHandlerFactory
}
