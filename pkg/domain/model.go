package domain

type Error struct {
	Message    string
	StatusCode string
}

func (e *Error) Error() string {
	return e.Message
}

type Identifier interface {
	~int | ~string
}

type Entity interface {
	ID() Identifier
	setID(id Identifier)
}

type Command interface {
	Type() string
}

type Event interface {
	Type() string
}
