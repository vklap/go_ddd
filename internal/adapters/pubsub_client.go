package adapters

import (
	"context"
	"encoding/json"
	"github.com/vklap/go_ddd/internal/domain/command_model"
)

type PubSubClient interface {
	GetMessages(ctx context.Context) (chan []byte, error)
}

// InMemoryPubSubClient is used for demo purposes.
type InMemoryPubSubClient struct{}

func (c *InMemoryPubSubClient) GetMessages(ctx context.Context) (chan []byte, error) {
	messages := make(chan []byte)
	go func() {
		for {
			command := &command_model.ChangeEmailCommand{UserID: "1", NewEmail: "eli.cohen@mossad.gov.il"}
			data, err := json.Marshal(command)
			if err != nil {
				panic(err)
			}
			messages <- data
			close(messages)
			break
		}
	}()
	return messages, nil
}

var _ PubSubClient = (*InMemoryPubSubClient)(nil)
