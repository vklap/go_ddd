# A Domain-Driven Design (DDD) Framework for Go Developers <img src="https://s3-us-west-2.amazonaws.com/s.cdpn.io/66955/gopher.svg" alt="gopher" width="25"/><img src="https://s3-us-west-2.amazonaws.com/s.cdpn.io/66955/gopher.svg" alt="gopher" width="23"/><img src="https://s3-us-west-2.amazonaws.com/s.cdpn.io/66955/gopher.svg" alt="gopher" width="21"/>
[![Go Reference](https://pkg.go.dev/badge/github.com/vklap/ddd.svg)](https://pkg.go.dev/github.com/vklap/go_ddd)

## What is this library good for?
This is a lightweight framework that provides a quick setup for
[Domain-Driven](https://en.wikipedia.org/wiki/Domain-driven_design) designed apps that
are easy to unit test - and is based on battle tested DDD Design Patterns, such as:

1. `Domain Layer` entities with domain events for handling side effects (or even support Event-Driven Architectures) 
2. `Application Service Layer` flow handlers that are being executed by units of work (to commit/rollback operations)
3. `Infrastructure/Adapters Layer` to external resources (such as: database repositories, web service clients, etc.)
4. `CQRS` (Command Query Responsibility Separation) with domain commands


This library has no external dependencies :beers:

## Installation

```shell
go get -u github.com/vklap/go_ddd
```

## Import

```go

import "github.com/vklap/go_ddd/pkg/ddd"

func main() {
    b := ddd.NewBootstrapper()
}

```

## How to implement it?

A sample implementation is provided within the [cmd](https://github.com/vklap/go_ddd/tree/main/cmd/worker) 
and [internal](https://github.com/vklap/go_ddd/tree/main/internal) folders of the source code.

The below explanations are based on this sample implementation.

## Sample Implementation

Imagine a simple background job for changing a user's email :email: that consists of:

1. The main flow: modify the user's email, and persist it in the database.
2. A side effect flow: send the user a notification email about the above modification.

Applying DDD in the code consists of the following steps:
- **Step 1**: `Domain Modeling` of the `Commands`, `Events`, and `Entities`
- **Step 2**: Add the `Adapters` required by the `Service Layer`
- **Step 3**: Implement the `Command` and `Event` handlers in the `Service Layer` for managing the applicative flows
- **Step 4**: Create the `Bootrapper` that connects all the above pieces together
- **Step 5**: Add an `Entrypoint` listener that receives messages from a message broker (such as RabbitMQ, Amazon SQS, etc.)
- **Step 6**: Specify the golang `main` function  

### Step 1: Domain Modeling

The `Domain Layer` should contain only in memory business logic and business rules,
and should not be aware of external dependencies, such as databases, web services, file system, etc.

Let's first implement the [ChangeEmailCommand](https://github.com/vklap/go_ddd/blob/main/internal/domain/command_model/change_email_command.go) 
and the [EmailChangedEvent](https://github.com/vklap/go_ddd/blob/main/internal/domain/command_model/email_changed_event.go):

#### ChangeEmailCommand
```go
package command_model

import "github.com/vklap/go_ddd/pkg/ddd"

// ChangeEmailCommand contains the data required to change the user's email.
// Besides this, it also represents a main flow.
type ChangeEmailCommand struct {
	UserID   string `json:"user_id"`
	NewEmail string `json:"new_email"`
}

func (c *ChangeEmailCommand) IsValid() error {
	if c.UserID == "" {
		return ddd.NewError("userID cannot be empty", ddd.StatusCodeBadRequest)
	}
	if c.NewEmail == "" {
		return ddd.NewError("email cannot be empty", ddd.StatusCodeBadRequest)
	}
	return nil
}

func (c *ChangeEmailCommand) CommandName() string {
	return "ChangeEmailCommand"
}

// The below line ensures at compile time that ChangeEmailCommand adheres to the ddd.Command interface 
var _ ddd.Command = (*ChangeEmailCommand)(nil)
```

#### EmailChangedEvent
```go
package command_model

import "github.com/vklap/go_ddd/pkg/ddd"

// EmailChangedEvent contains the data required to notify about the email modification.
// Besides this, it also represents a side effect flow that should be implemented.
type EmailChangedEvent struct {
	UserID        string
	OriginalEmail string
	NewEmail      string
}

func (e *EmailChangedEvent) EventName() string {
	return "EmailChangedEvent"
}

// The below line ensures at compile time that EmailChangedEvent adheres to the ddd.Event interface
var _ ddd.Event = (*EmailChangedEvent)(nil)
```

Next, let's implement the [User](https://github.com/vklap/go_ddd/blob/main/internal/domain/command_model/user.go) Entity:
```go
package command_model

import "github.com/vklap/go_ddd/pkg/ddd"

// User is composed of ddd.BaseEntity which exposes the entity's ID and Events,
// and the user's Email.
type User struct {
	ddd.BaseEntity
	email string
}

func (u *User) Email() string {
	return u.email
}

func (u *User) SetEmail(value string) {
	if u.email != "" && u.email != value {
		// Record the EmailChangedEvent 
		u.AddEvent(&EmailChangedEvent{UserID: u.ID(), NewEmail: value, OriginalEmail: u.email})
	}
	u.email = value
}

// The below line ensures at compile time that User adheres to the ddd.Entity interface
var _ ddd.Entity = (*User)(nil)
```

### Step 2: Adapters

The `Adapters Layer` is responsible for communicating with external resources only, 
and should not contain any business logic.

- [Repository](https://github.com/vklap/go_ddd/blob/main/internal/adapters/repository.go)
- [EmailClient](https://github.com/vklap/go_ddd/blob/main/internal/adapters/email_client.go)
- [PubSubClient](https://github.com/vklap/go_ddd/blob/main/internal/adapters/pubsub_client.go)

#### Repository

```go
package adapters

import (
	"context"
	"fmt"
	"github.com/vklap/go_ddd/internal/domain/command_model"
	"github.com/vklap/go_ddd/pkg/ddd"
)

// usersById is used solely for demo purposes, to support the InMemoryRepository
var usersById = make(map[string]*command_model.User)

func init() {
	user := &command_model.User{}
	user.SetID("1")
	user.SetEmail("kamel.amit@thaabet")
	usersById[user.ID()] = user
}

type Repository interface {
	GetUserById(ctx context.Context, id string) (*command_model.User, error)
	SaveUser(ctx context.Context, user *command_model.User) error
	ddd.RollbackCommitter
}

// InMemoryRepository is used for demo purposes.
// In the real world it might be a MongoDBRepository, PostgresqlRepository, etc.
type InMemoryRepository struct{}

func (r *InMemoryRepository) GetUserById(ctx context.Context, id string) (*command_model.User, error) {
	user, ok := usersById[id]
	if ok == false {
		return nil, ddd.NewError(fmt.Sprintf("user with id %q does not exist", id), ddd.StatusCodeNotFound)
	}
	return user, nil
}

func (r *InMemoryRepository) SaveUser(ctx context.Context, user *command_model.User) error {
	usersById[user.ID()] = user
	return nil
}

func (r *InMemoryRepository) Commit(ctx context.Context) error {
	return nil
}

func (r *InMemoryRepository) Rollback(ctx context.Context) error {
	return nil
}

var _ Repository = (*InMemoryRepository)(nil)
```

#### EmailClient
```go
package adapters

import "log"

type EmailClient interface {
	SendEmail(from, to string, title string, message string) error
}

// InMemoryEmailClient used for demo purposes
type InMemoryEmailClient struct{}

func (c *InMemoryEmailClient) SendEmail(from, to string, title string, message string) error {
	log.Printf("Sent email from %q to %q with title %q and message: %q\n", from, to, title, message)
	return nil
}

var _ EmailClient = (*InMemoryEmailClient)(nil)
```

#### PubSubClient
```go
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
```

### Step 3: The `Command` and `Event` Handlers

The `Service Layer` is responsible for managing applicative flows and as such contains 
references to both the `Domain Layer` and the `Adapters Layer`.

In our use case, the **main flow** which is `change email`, is handled by [ChangeEmailCommandHandler](https://github.com/vklap/go_ddd/blob/main/internal/service_layer/command_handlers/change_email_command_handler.go) -
which registers events that will eventually be handled by event handlers (if they exist). 

In the below implementation, the `change email` flow registers an `EmailChangedEvent`, which will be handled
by [EmailChangedEventHandler](https://github.com/vklap/go_ddd/blob/main/internal/service_layer/event_handlers/email_changed_event_handler.go).

#### ChangeEmailCommandHandler

```go
package command_handlers

import (
	"context"
	"fmt"
	"github.com/vklap/go_ddd/internal/adapters"
	"github.com/vklap/go_ddd/internal/domain/command_model"
	"github.com/vklap/go_ddd/pkg/ddd"
)

// ChangeEmailCommandHandler implements ddd.CommandHandler.
type ChangeEmailCommandHandler struct {
	repository adapters.Repository
	events     []ddd.Event
}

// NewChangeEmailCommandHandler is a constructor function to be used by the Bootstrapper.
func NewChangeEmailCommandHandler(repository adapters.Repository) *ChangeEmailCommandHandler {
	return &ChangeEmailCommandHandler{repository: repository}
}

// Handle manages the business logic flow, and is the glue between the Domain and the Adapters.
func (h *ChangeEmailCommandHandler) Handle(ctx context.Context, command ddd.Command) (any, error) {
	changeEmailCommand, ok := command.(*command_model.ChangeEmailCommand)
	if ok == false {
		return nil, fmt.Errorf("ChangeEmailCommandHandler expects a command of type %T", changeEmailCommand)
	}
	
	// No need to call changeEmailCommand.IsValid() - as it's being called by the framework.

	// Delegate fetching data to the repository, which belongs to the Adapters Layer.
	user, err := h.repository.GetUserById(ctx, changeEmailCommand.UserID)
	if err != nil {
		return nil, err
	}

	// Delegate updating the email to the user, which is a Domain Entity.
	// The SetEmail email method is responsible to detect if the email was changed,
	// and if so, then it will record an EmailChangedEvent. 
	user.SetEmail(changeEmailCommand.NewEmail)

	// Delegate storing data to the repository, which belongs to the Adapters Layer.
	if err = h.repository.SaveUser(ctx, user); err != nil {
		return nil, err
	}

	// This is where Domain events are being registered by the handler,
	// so they can eventually be dispatched to event handlers (if they exist). 
	// In our use case the events will be dispatched to the EmailChangedEventHandler.
	h.events = append(h.events, user.Events()...)

	return nil, nil
}

// Commit is responsible for committing the changes performed by the Handle method, such as
// committing a database transaction managed by the repository.
// This method is being called by the framework, so it should not be called from within the Handle method.
func (h *ChangeEmailCommandHandler) Commit(ctx context.Context) error {
	return h.repository.Commit(ctx)
}

// Rollback is responsible to rollback changes performed by the Handle method, such as
// rollback a database transaction managed by the repository.
// This method is being called by the framework, so it should not be called from within the Handle method.
func (h *ChangeEmailCommandHandler) Rollback(ctx context.Context) error {
	return h.repository.Rollback(ctx)
}

// Events reports about events. 
// These events will be handled by the DDD framework if appropriate event handlers were registered by the bootstrapper.
// This method is being called by the framework, so it should not be called from within the Handle method.
func (h *ChangeEmailCommandHandler) Events() []ddd.Event {
	return h.events
}

var _ ddd.CommandHandler = (*ChangeEmailCommandHandler)(nil)
```

#### EmailChangedEventHandler

```go
package event_handlers

import (
	"context"
	"fmt"
	"github.com/vklap/go_ddd/internal/adapters"
	"github.com/vklap/go_ddd/internal/domain/command_model"
	"github.com/vklap/go_ddd/pkg/ddd"
)

// EmailChangedEventHandler implements ddd.EventHandler.
type EmailChangedEventHandler struct {
	emailClient adapters.EmailClient
}

// NewEmailChangedEventHandler is a constructor function to be used by the Bootstrapper.
func NewEmailChangedEventHandler(emailClient adapters.EmailClient) *EmailChangedEventHandler {
	return &EmailChangedEventHandler{emailClient: emailClient}
}

// Handle manages the business logic flow, and is the glue between the Domain and the Adapters.
func (h *EmailChangedEventHandler) Handle(ctx context.Context, event ddd.Event) error {
	emailChangedEvent, ok := event.(*command_model.EmailChangedEvent)
	if ok == false {
		panic(fmt.Sprintf("failed to handle email changed: want %T, got %T", &command_model.EmailChangedEvent{}, emailChangedEvent))
	}
	from := "noreply@example.com"
	to := emailChangedEvent.OriginalEmail
	title := "NewEmail Changed Notification"
	message := fmt.Sprintf("Your email was changed to %v", emailChangedEvent.NewEmail)
	return h.emailClient.SendEmail(from, to, title, message)
}

// Commit is responsible for committing the changes performed by the Handle method, such as
// committing a database transaction managed by the repository.
// This method is being called by the framework, so it should not be called from within the Handle method.
func (h *EmailChangedEventHandler) Commit(ctx context.Context) error {
	return nil
}

// Rollback is responsible to rollback changes performed by the Handle method, such as
// rollback a database transaction managed by the repository.
// This method is being called by the framework, so it should not be called from within the Handle method.
func (h *EmailChangedEventHandler) Rollback(ctx context.Context) error {
	return nil
}

// Events reports about events.
// These events will be handled by the DDD framework if appropriate event handlers were registered by the bootstrapper.
// This method is being called by the framework, so it should not be called from within the Handle method.
func (h *EmailChangedEventHandler) Events() []ddd.Event {
	return nil
}

var _ ddd.EventHandler = (*EmailChangedEventHandler)(nil)
```

### Step 4: The `Bootrapper`

The [Bootstrapper](https://github.com/vklap/go_ddd/blob/main/internal/entrypoints/boostrapper/bootsrapper.go)
registers the `command` and `event` handlers with their `adapter dependencies`. 
Besides this, it acts as a facade for receiving commands. 

```go
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
```

### Step 5: The Entry Point
The [worker](https://github.com/vklap/go_ddd/blob/main/internal/entrypoints/worker/worker.go)
starts listening for notifications from a fake message broker with requests to change the user's email.

```go
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
			panic(err)
		}
		_, err = boostrapper.HandleCommand[*command_model.ChangeEmailCommand](context.Background(), &command)
		if err != nil {
			log.Printf("handle ChangeEmailCommand failed: %v", err)
		}
	}
}
```

### Step 6: The main function


You can execute and eventually debug the code by running the following command:
`go run ./cmd/worker/main.go`

```go
package main

import "github.com/vklap/go_ddd/internal/entrypoints/worker"

func main() {
	worker.Start()
}
```

## Links

- [pkg.go.dev](https://pkg.go.dev/github.com/vklap/go_ddd/pkg/ddd)
- [README.md](https://github.com/vklap/go_ddd/blob/main/README.md)