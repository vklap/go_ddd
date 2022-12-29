package boostrapper

import (
	"context"
	"github.com/vklap/go_ddd/internal/adapters"
	"github.com/vklap/go_ddd/internal/domain/command_model"
	"github.com/vklap/go_ddd/internal/service_layer/command_handlers"
	"github.com/vklap/go_ddd/internal/service_layer/event_handlers"
	"github.com/vklap/go_ddd/pkg/ddd"
)

var b *ddd.Bootstrapper
var pubSubClient adapters.PubSubClient

// init creates the bootstrapper instance and registers the command and event handlers.
func init() {
	b = ddd.NewBootstrapper()
	b.RegisterCommandHandlerFactory(&command_model.ChangeEmailCommand{}, func() (ddd.CommandHandler, error) {
		return command_handlers.NewChangeEmailCommandHandler(&adapters.InMemoryRepository{}), nil
	})
	b.RegisterEventHandlerFactory(&command_model.EmailChangedEvent{}, func() (ddd.EventHandler, error) {
		return event_handlers.NewEmailChangedEventHandler(&adapters.InMemoryEmailClient{}), nil
	})
	pubSubClient = &adapters.InMemoryPubSubClient{}
}

// GetPubSubClientInstance returns an instance of the pubSubClient
func GetPubSubClientInstance() adapters.PubSubClient {
	return pubSubClient
}

// HandleCommand encapsulates the Bootstrapper HandleCommand, and gives a strongly typed interface
// provided by go's generics.
func HandleCommand[Command ddd.Command](ctx context.Context, command Command) (any, error) {
	return b.HandleCommand(ctx, command)
}
