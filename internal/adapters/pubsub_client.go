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
	GetSaveUserMessages(ctx context.Context) (chan []byte, error)
	NotifyEmailChanged(ctx context.Context, userId string, newEmail string, oldEmail string) error
	NotifyKPIService(ctx context.Context, e *command_model.KPIEvent) error
	ddd.RollbackCommitter
}

// InMemoryPubSubClient is used for demo purposes.
type InMemoryPubSubClient struct {
	Commands                 []*command_model.SaveUserCommand
	CommitCalled             bool
	CommitShouldFail         bool
	MailSent                 bool
	NotifyEmailSetCalled     bool
	NotifyEmailSetFailed     bool
	NotifyEmailSetNewEmail   string
	NotifyEmailSetOldEmail   string
	NotifyEmailSetShouldFail bool
	NotifyEmailSetUserId     string
	NotifyKPICalled          bool
	KPIEvent                 *command_model.KPIEvent
	RollbackCalled           bool
	RollbackShouldFail       bool
	KPIEventSent             bool
}

func NewInMemoryPubSubClient() *InMemoryPubSubClient {
	return &InMemoryPubSubClient{Commands: make([]*command_model.SaveUserCommand, 0)}
}

func (c *InMemoryPubSubClient) GetSaveUserMessages(ctx context.Context) (chan []byte, error) {
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
	if c.NotifyEmailSetFailed {
		return errors.New("notify email changed failed")
	}
	c.NotifyEmailSetCalled = true
	if c.NotifyEmailSetShouldFail {
		return errors.New("notify email changed has failed")
	}
	c.NotifyEmailSetNewEmail = newEmail
	c.NotifyEmailSetOldEmail = oldEmail
	c.NotifyEmailSetUserId = userId
	log.Printf("requested to send EmailChanged notification: userID=%q, oldEmail=%q, newEmail=%q", userId, oldEmail, newEmail)
	return nil
}

func (c *InMemoryPubSubClient) NotifyKPIService(ctx context.Context, e *command_model.KPIEvent) error {
	c.NotifyKPICalled = true
	c.KPIEvent = e
	log.Printf("notfiied KPI service: %v", e)
	return nil
}

func (c *InMemoryPubSubClient) Commit(ctx context.Context) error {
	c.CommitCalled = true
	if c.CommitShouldFail {
		return errors.New("commit failed")
	}
	if c.NotifyEmailSetCalled {
		c.MailSent = true
	}
	if c.NotifyKPICalled {
		c.KPIEventSent = true
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
