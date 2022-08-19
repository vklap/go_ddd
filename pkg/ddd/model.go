package ddd

type Command interface {
	CommandName() string
	IsValid() error
}

type Event interface {
	EventName() string
}

type Entity interface {
	ID() string
	SetID(id string)
	Events() []Event
}

type BaseEntity struct {
	id     string
	events []Event
}

func (e *BaseEntity) ID() string {
	return e.id
}

func (e *BaseEntity) SetID(id string) {
	e.id = id
}

func (e *BaseEntity) Events() []Event {
	return e.events
}

func (e *BaseEntity) AddEvent(event Event) {
	e.events = append(e.events, event)
}

var _ Entity = (*BaseEntity)(nil)
