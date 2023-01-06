package adapters

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/vklap/go_ddd/internal/domain/command_model"
	"github.com/vklap/go_ddd/pkg/ddd"
	"log"
)

type PubSubClient interface {
	GetChangeEmailMessages(ctx context.Context) (chan []byte, error)
	NotifyEmailChanged(ctx context.Context, userId string, newEmail string, oldEmail string) error
	NotifySlack(ctx context.Context, message string) error
	ddd.RollbackCommitter
}

// InMemoryPubSubClient is used for demo purposes.
type InMemoryPubSubClient struct {
	Commands                     []*command_model.ChangeEmailCommand
	CommitCalled                 bool
	CommitShouldFail             bool
	MailSent                     bool
	NotifyEmailChangedCalled     bool
	NotifyEmailChangedFailed     bool
	NotifyEmailChangedNewEmail   string
	NotifyEmailChangedOldEmail   string
	NotifyEmailChangedShouldFail bool
	NotifyEmailChangedUserId     string
	NotifySlackCalled            bool
	NotifySlackMessage           string
	RollbackCalled               bool
	RollbackShouldFail           bool
	SlackMessageSent             bool
}

func NewInMemoryPubSubClient() *InMemoryPubSubClient {
	return &InMemoryPubSubClient{Commands: make([]*command_model.ChangeEmailCommand, 0)}
}

func (c *InMemoryPubSubClient) GetChangeEmailMessages(ctx context.Context) (chan []byte, error) {
	messages := make(chan []byte)
	go func() {
		for _, command := range c.Commands {
			data, err := json.Marshal(command)
			if err != nil {
				panic(err)
			}
			messages <- data
		}
		close(messages)
	}()
	return messages, nil
}

func (c *InMemoryPubSubClient) NotifyEmailChanged(ctx context.Context, userId string, newEmail, oldEmail string) error {
	if c.NotifyEmailChangedFailed {
		return errors.New("notify email changed failed")
	}
	c.NotifyEmailChangedCalled = true
	if c.NotifyEmailChangedShouldFail {
		return errors.New("notify email changed has failed")
	}
	c.NotifyEmailChangedNewEmail = newEmail
	c.NotifyEmailChangedOldEmail = oldEmail
	c.NotifyEmailChangedUserId = userId
	log.Printf("requested to send EmailChanged notification: userID=%q, oldEmail=%q, newEmail=%q", userId, oldEmail, newEmail)
	return nil
}

func (c *InMemoryPubSubClient) NotifySlack(ctx context.Context, message string) error {
	c.NotifySlackCalled = true
	c.NotifySlackMessage = message
	log.Printf("requested to send Slack message: %q", message)
	return nil
}

func (c *InMemoryPubSubClient) Commit(ctx context.Context) error {
	c.CommitCalled = true
	if c.CommitShouldFail {
		return errors.New("commit failed")
	}
	if c.NotifyEmailChangedCalled {
		c.MailSent = true
	}
	if c.NotifySlackCalled {
		c.SlackMessageSent = true
	}
	return nil
}

func (c *InMemoryPubSubClient) Rollback(ctx context.Context) error {
	c.RollbackCalled = true
	if c.RollbackShouldFail {
		return errors.New("rollback failed")
	}
	return nil
}

var _ PubSubClient = (*InMemoryPubSubClient)(nil)
