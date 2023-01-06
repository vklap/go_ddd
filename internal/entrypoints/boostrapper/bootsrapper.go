package boostrapper

import (
	"context"
	"github.com/vklap/go_ddd/internal/adapters"
	"github.com/vklap/go_ddd/internal/domain/command_model"
	"github.com/vklap/go_ddd/internal/service_layer/command_handlers"
	"github.com/vklap/go_ddd/internal/service_layer/event_handlers"
	"github.com/vklap/go_ddd/pkg/ddd"
)

var Instance *DemoBootstrapper

// init creates the Bootstrapper instance and registers the command and event handlers.
func init() {
	Instance = New()
}

type DemoBootstrapper struct {
	PubSubClient *adapters.InMemoryPubSubClient
	Repository   *adapters.InMemoryRepository
	Bootstrapper *ddd.Bootstrapper
}

func New() *DemoBootstrapper {
	bs := &DemoBootstrapper{
		PubSubClient: adapters.NewInMemoryPubSubClient(),
		Repository:   adapters.NewInMemoryRepository(),
		Bootstrapper: ddd.NewBootstrapper(),
	}
	bs.Bootstrapper.RegisterCommandHandlerFactory(&command_model.ChangeEmailCommand{}, func() (ddd.CommandHandler, error) {
		return command_handlers.NewChangeEmailCommandHandler(bs.Repository), nil
	})
	bs.Bootstrapper.RegisterEventHandlerFactory(&command_model.EmailChangedEvent{}, func() (ddd.EventHandler, error) {
		return event_handlers.NewEmailChangedEventHandler(bs.PubSubClient), nil
	})
	bs.Bootstrapper.RegisterEventHandlerFactory(&command_model.NotifySlackEvent{}, func() (ddd.EventHandler, error) {
		return event_handlers.NewNotifySlackEventHandler(bs.PubSubClient), nil
	})
	return bs
}

// HandleCommand encapsulates the Bootstrapper HandleCommand, and gives a strongly typed interface
// provided by go's generics.
func HandleCommand[Command ddd.Command](ctx context.Context, command Command) (any, error) {
	return Instance.Bootstrapper.HandleCommand(ctx, command)
}
