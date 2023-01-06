package command_model

import (
	"github.com/vklap/go_ddd/pkg/ddd"
)

// KPIEvent contains data for KPI (Key Performance Indicators) metrics.
type KPIEvent struct {
	Action string
	Data   string
}

func (e *KPIEvent) EventName() string {
	return "KPIEvent"
}

// The below line ensures at compile time that KPIEvent adheres to the ddd.Event interface
var _ ddd.Event = (*KPIEvent)(nil)
