package worker

import (
	"context"
	"encoding/json"
	"github.com/vklap/go_ddd/internal/domain/command_model"
	"github.com/vklap/go_ddd/internal/entrypoints/boostrapper"
	"log"
)

// Start listens to the message broker and dispatches commands to be handled.
func Start() {
	pubSubClient := boostrapper.GetPubSubClientInstance()
	messages, err := pubSubClient.GetMessages(context.Background())
	if err != nil {
		panic(err)
	}
	for message := range messages {
		var command command_model.ChangeEmailCommand
		err = json.Unmarshal(message, &command)
		if err != nil {
			log.Printf("failed to unmarshal ChangeEmailCommand: %v (message: %v)", err, message)
			continue
		}
		_, err = boostrapper.HandleCommand[*command_model.ChangeEmailCommand](context.Background(), &command)
		if err != nil {
			log.Printf("handle ChangeEmailCommand failed: %v", err)
		}
	}
}
