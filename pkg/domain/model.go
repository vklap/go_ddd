package domain

type Error struct {
	message    string
	statusCode string
}

func (e *Error) Error() string {
	return e.message
}

func (e *Error) StatusCode() string {
	return e.statusCode
}

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

var _ Entity = (*BaseEntity)(nil)
