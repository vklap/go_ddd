package service_layer

import (
	"context"
	"github.com/vklap.go-ddd/pkg/adapters"
	"github.com/vklap.go-ddd/pkg/domain"
)

type CommandHandler interface {
	handle(ctx context.Context, command domain.Command) (any, error)
	adapters.CommitterRollbacker
}

type EventHandler interface {
	handle(ctx context.Context, event domain.Event) error
	adapters.CommitterRollbacker
}
