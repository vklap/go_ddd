package domain

const (
	NotAllowed string = "not_allowed"
)

type Identifier struct {
	int64
	string
}

type Event interface {
}

type Command interface {
	validate() error
}

type MessageType interface {
	Command
	Event
}

type Auth struct {
	UserID    Identifier
	SuperUser bool
}

type Entity struct {
	ID     Identifier
	Events []Event
}

type Permission struct {
	Action string
	ID     Identifier
}

type BoundedContextError struct {
	Status  string
	Message string
}

func (e BoundedContextError) Error() string {
	return e.Message
}

func (e BoundedContextError) Is(target error) bool {
	if other, ok := target.(BoundedContextError); ok {
		return e.Status == other.Status
	}
	return false
}
