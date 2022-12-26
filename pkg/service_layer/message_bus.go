package service_layer

import (
	"context"
	"github.com/vklap.go-ddd/pkg/domain"
	"log"
)

type MessageBus struct {
	commandHandlerFactory CommandHandlerFactory
	eventHandlersFactory  EventHandlersFactory
	events                []domain.Event
}

func (m *MessageBus) Handle(ctx context.Context, command domain.Command) (any, error) {
	handler, err := m.commandHandlerFactory.CreateHandler(command)
	if err != nil {
		return nil, err
	}

	uow := CommandUnitOfWork{handler}
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

func (m *MessageBus) handleEvents(ctx context.Context) {
	for len(m.events) > 0 {
		var event domain.Event
		event, m.events = m.events[0], m.events[1:]
		handlers, err := m.eventHandlersFactory.CreateHandlers(event)
		if err != nil {
			log.Printf("failed to create event handler for %q: %v", event.EventName(), err)
		}
		for _, handler := range handlers {
			uow := EventUnitOfWork{handler}
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
