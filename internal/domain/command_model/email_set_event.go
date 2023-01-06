package command_model

import "github.com/vklap/go_ddd/pkg/ddd"

// EmailSetEvent contains the data required to notify about the email modification.
type EmailSetEvent struct {
	UserID        string
	OriginalEmail string
	NewEmail      string
}

func (e *EmailSetEvent) EventName() string {
	return "EmailSetEvent"
}

// The below line ensures at compile time that EmailSetEvent adheres to the ddd.Event interface
var _ ddd.Event = (*EmailSetEvent)(nil)
