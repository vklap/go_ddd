package command_model

import "github.com/vklap/go_ddd/pkg/ddd"

// EmailChangedEvent contains the data required to notify about the email modification.
// Besides this, it also represents a side effect flow that should be implemented.
type EmailChangedEvent struct {
	UserID        string
	OriginalEmail string
	NewEmail      string
}

func (e *EmailChangedEvent) EventName() string {
	return "EmailChangedEvent"
}

// The below line ensures at compile time that EmailChangedEvent adheres to the ddd.Event interface
var _ ddd.Event = (*EmailChangedEvent)(nil)
