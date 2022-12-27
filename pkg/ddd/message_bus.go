package ddd

import (
	"context"
	"log"
)

type messageBus struct {
	commandHandlerFactory *commandHandlerFactory
	eventHandlersFactory  *eventHandlersFactory
	events                []Event
}

func newMessageBus(commandHandlerFactory *commandHandlerFactory, eventHandlersFactory *eventHandlersFactory) *messageBus {
	return &messageBus{
		commandHandlerFactory: commandHandlerFactory,
		eventHandlersFactory:  eventHandlersFactory,
	}
}

func (m *messageBus) Handle(ctx context.Context, command Command) (any, error) {
	handler, err := m.commandHandlerFactory.CreateHandler(command)
	if err != nil {
		return nil, err
	}

	uow := commandUnitOfWork{handler}
	result, err := uow.HandleCommand(ctx, command)
	if err != nil {
		return nil, err
	}

	for _, entity := range handler.SavedEntities() {
		m.events = append(m.events, entity.Events()...)
	}

	m.handleEvents(ctx)

	return result, nil
}

func (m *messageBus) handleEvents(ctx context.Context) {
	for len(m.events) > 0 {
		var event Event
		event, m.events = m.events[0], m.events[1:]
		handlers, err := m.eventHandlersFactory.CreateHandlers(event)
		if err != nil {
			log.Printf("failed to create event handler for %q: %v", event.EventName(), err)
		}
		for _, handler := range handlers {
			uow := eventUnitOfWork{handler}
			err = uow.HandleEvent(ctx, event)
			if err != nil {
				log.Printf("failed to handle event %q: %v", event.EventName(), err)
			}
			for _, entity := range handler.SavedEntities() {
				m.events = append(m.events, entity.Events()...)
			}
		}
	}
}
