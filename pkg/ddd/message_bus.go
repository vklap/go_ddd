package ddd

import (
	"context"
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

func (m *messageBus) Publish(ctx context.Context, command Command) (any, error) {
	if err := command.IsValid(); err != nil {
		return nil, err
	}
	handler, err := m.commandHandlerFactory.CreateHandler(command)
	if err != nil {
		return nil, err
	}

	uow := commandUnitOfWork{handler}
	result, err := uow.HandleCommand(ctx, command)
	if err != nil {
		return nil, err
	}

	m.events = append(m.events, handler.Events()...)

	if err = m.handleEvents(ctx); err != nil {
		return nil, err
	}

	return result, nil
}

func (m *messageBus) handleEvents(ctx context.Context) error {
	for len(m.events) > 0 {
		var event Event
		event, m.events = m.events[0], m.events[1:]
		handlers, err := m.eventHandlersFactory.CreateHandlers(event)
		if err != nil {
			return err
		}
		for _, handler := range handlers {
			uow := eventUnitOfWork{handler}
			err = uow.HandleEvent(ctx, event)
			if err != nil {
				return err
			}
			m.events = append(m.events, handler.Events()...)
		}
	}
	return nil
}
