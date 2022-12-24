package domain

type EventType string
type CommandType string
type StatusCode string

type Error struct {
	message    string
	statusCode StatusCode
}

func (e *Error) Error() string {
	return e.message
}

func (e *Error) StatusCode() StatusCode {
	return e.statusCode
}

type Command interface {
	Type() CommandType
	IsValid() (error, bool)
}

type Event interface {
	Type() EventType
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
