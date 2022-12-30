package go_ddd

// Command interface that should be implemented by commands.
type Command interface {
	CommandName() string
	IsValid() error
}

// Event interface that should be implemented by events.
type Event interface {
	EventName() string
}

// Entity interface that should be implemented by entities.
type Entity interface {
	ID() string
	SetID(id string)
	Events() []Event
}

// BaseEntity struct that can be used in Entity compositions, to prevent repetitive boilerplate code.
type BaseEntity struct {
	id     string
	events []Event
}

// ID returns the entity's ID.
func (e *BaseEntity) ID() string {
	return e.id
}

// SetID sets the entity's ID.
func (e *BaseEntity) SetID(id string) {
	e.id = id
}

// Events exposes the events registered by the entity.
func (e *BaseEntity) Events() []Event {
	return e.events
}

// AddEvent registers events.
func (e *BaseEntity) AddEvent(event Event) {
	e.events = append(e.events, event)
}

var _ Entity = (*BaseEntity)(nil)
