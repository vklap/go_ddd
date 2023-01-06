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
	// Setup InMemory fake data
	bs := boostrapper.Instance
	user := &command_model.User{}
	user.SetEmail("kamel.amin@thaabet.sy")
	user.SetID("1")
	bs.Repository.UsersById[user.ID()] = user

	fakePubSubMessage := &command_model.SaveUserCommand{
		Email:  "eli.cohen@mossad.gov.il",
		UserID: "1",
	}
	bs.PubSubClient.Commands = append(bs.PubSubClient.Commands, fakePubSubMessage)

	// Start listening for messages from fake in memory PubSub
	messages, err := bs.PubSubClient.GetSaveUserMessages(context.Background())
	if err != nil {
		panic(err)
	}
	for message := range messages {
		var command command_model.SaveUserCommand
		err = json.Unmarshal(message, &command)
		if err != nil {
			log.Printf("failed to unmarshal SaveUserCommand: %v (message: %v)", err, message)
			continue
		}
		_, err = boostrapper.HandleCommand[*command_model.SaveUserCommand](context.Background(), &command)
		if err != nil {
			log.Printf("handle SaveUserCommand failed: %v", err)
		}
	}
}
