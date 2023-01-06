package command_model

import (
	"github.com/vklap/go_ddd/pkg/ddd"
)

// NotifySlackEvent contains the slack message.
type NotifySlackEvent struct {
	Message string
}

func (e *NotifySlackEvent) EventName() string {
	return "NotifySlackEvent"
}

// The below line ensures at compile time that NotifySlackEvent adheres to the ddd.Event interface
var _ ddd.Event = (*NotifySlackEvent)(nil)
