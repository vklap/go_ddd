# A Domain-Driven Design (DDD) Framework for Go Developers <img src="https://s3-us-west-2.amazonaws.com/s.cdpn.io/66955/gopher.svg" alt="gopher" width="25"/><img src="https://s3-us-west-2.amazonaws.com/s.cdpn.io/66955/gopher.svg" alt="gopher" width="23"/><img src="https://s3-us-west-2.amazonaws.com/s.cdpn.io/66955/gopher.svg" alt="gopher" width="21"/>
[![Go Reference](https://pkg.go.dev/badge/github.com/vklap/ddd.svg)](https://pkg.go.dev/github.com/vklap/go_ddd)

## What is this library good for?
This is a lightweight framework that provides a quick and simple setup for
[Domain-Driven](https://en.wikipedia.org/wiki/Domain-driven_design) designed apps
that are a pleasure to maintain and easy to unit test.

These are the main features that are supported by the framework:
1. **Unit of Work** with a **commit** and **rollback** mechanism for application layer handlers
2. Definition of **Domain Commands** in the domain layer and their **Command Handlers** in the application layer
3. Definition of **Domain Events** in the domain layer and their **Event Handlers** in the application layer
4. **Event-Driven Architecture** based on **Domain Events**

This library has no external dependencies :beers: and hence should be easy to add to any project that can benefit from 
DDD.

Many concepts used in this framework are based on the author's own experience and greatly inspired by amazing DDD books,
such as:

* [Domain-Driven Design: Tackling Complexity in the Heart of Software](https://www.oreilly.com/library/view/domain-driven-design-tackling/0321125215/)
* [Architecture Patterns with Python](https://www.oreilly.com/library/view/architecture-patterns-with/9781492052197/)
* [Event-Driven Architecture in Golang](https://www.packtpub.com/product/event-driven-architecture-in-golang/9781803238012)


## Installation

```shell
go get github.com/vklap/go_ddd
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

The below explanation is based on this sample implementation.

## Sample Implementation

Let's imagine a simplified background job for saving a user's details that consists 
of the following steps within a unit of work:

1. Get the new user's data from a **PubSub message broker** (such as Amazon SQS, RabbitMQ, etc.) 
   and transform it into a **command** object that can be handled by the **application layer**
2. Perform basic validations on the **command**'s data
3. Get the existing **user entity** data from the database, via a **repository**
4. Update the **user entity** with the data stored in the **command** object
5. **Save** the updated user entity **in the repository**
6. Either **commit** (and store the new data in the database) 
   or **rollback** (and thus discard the changes recorded in the previous steps)

Steps 2 (command validation) and 6 (commit or rollback) are triggered by the framework.

### How the code looks like?

#### Domain Layer

##### User Entity

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
	if value != "" && u.email != value {
		u.AddEvent(&EmailSetEvent{UserID: u.ID(), NewEmail: value, OriginalEmail: u.email})
	}
	u.email = value
}

// The below line ensures at compile time that User adheres to the ddd.Entity interface
var _ ddd.Entity = (*User)(nil)
```

##### SaveUserCommand

```go
package command_model

import "github.com/vklap/go_ddd/pkg/ddd"

// SaveUserCommand contains the data required to store a user's details.
type SaveUserCommand struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
}

func (c *SaveUserCommand) IsValid() error {
	if c.UserID == "" {
		return ddd.NewError("user ID cannot be empty", ddd.StatusCodeBadRequest)
	}
	if c.Email == "" {
		return ddd.NewError("email cannot be empty", ddd.StatusCodeBadRequest)
	}
	return nil
}

func (c *SaveUserCommand) CommandName() string {
	return "SaveUserCommand"
}

// The below line ensures at compile time that SaveUserCommand adheres to the ddd.Command interface
var _ ddd.Command = (*SaveUserCommand)(nil)
```

##### Repository

Please note that we're using an in memory repository for demo purposes 
(and also for the [unit tests](https://github.com/vklap/go_ddd/blob/main/pkg/ddd/bootstrapper_test.go))

```go
package adapters

import (
	"context"
	"errors"
	"fmt"
	"github.com/vklap/go_ddd/internal/domain/command_model"
	"github.com/vklap/go_ddd/pkg/ddd"
)

type Repository interface {
	GetUserById(ctx context.Context, id string) (*command_model.User, error)
	SaveUser(ctx context.Context, user *command_model.User) error
	ddd.RollbackCommitter
}

// InMemoryRepository is used for demo purposes.
// In the real world it might be a MongoDBRepository, PostgresqlRepository, etc.
type InMemoryRepository struct {
	CommitCalled       bool
	CommitShouldFail   bool
	RollbackCalled     bool
	RollbackShouldFail bool
	UsersById          map[string]*command_model.User
	savedUsers         []*command_model.User
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{UsersById: make(map[string]*command_model.User)}
}

func (r *InMemoryRepository) GetUserById(ctx context.Context, id string) (*command_model.User, error) {
	user, ok := r.UsersById[id]
	if ok == false {
		return nil, ddd.NewError(fmt.Sprintf("user with id %q does not exist", id), ddd.StatusCodeNotFound)
	}
	return user, nil
}

func (r *InMemoryRepository) SaveUser(ctx context.Context, user *command_model.User) error {
	r.savedUsers = append(r.savedUsers, user)
	return nil
}

func (r *InMemoryRepository) Commit(ctx context.Context) error {
	r.CommitCalled = true
	if r.CommitShouldFail {
		return errors.New("commit failed")
	}
	for _, user := range r.savedUsers {
		r.UsersById[user.ID()] = user
	}
	return nil
}

func (r *InMemoryRepository) Rollback(ctx context.Context) error {
	r.RollbackCalled = true
	if r.RollbackShouldFail {
		return errors.New("rollback failed")
	}
	r.savedUsers = make([]*command_model.User, 0)
	return nil
}

var _ Repository = (*InMemoryRepository)(nil)
```

##### SaveUserCommandHandler

This is the application layer flow that is triggered by the framework's unit of work - 
in order to either commit or rollback the changes. 

This handler is registered to the above defined `SaveUserCommand` - so that whenever this command is received, 
then this handler will be executed. The registration is handled by the `Bootstrapper` which will be shown later.

```go
package command_handlers

import (
	"context"
	"fmt"
	"github.com/vklap/go_ddd/internal/adapters"
	"github.com/vklap/go_ddd/internal/domain/command_model"
	"github.com/vklap/go_ddd/pkg/ddd"
)

// SaveUserCommandHandler implements ddd.CommandHandler.
type SaveUserCommandHandler struct {
	repository adapters.Repository
	events     []ddd.Event
}

// NewSaveUserCommandHandler is a constructor function to be used by the Bootstrapper.
func NewSaveUserCommandHandler(repository adapters.Repository) *SaveUserCommandHandler {
	return &SaveUserCommandHandler{repository: repository}
}

// Handle manages the business logic flow, and is the glue between the Domain and the Adapters.
func (h *SaveUserCommandHandler) Handle(ctx context.Context, command ddd.Command) (any, error) {
	saveUserCommand, ok := command.(*command_model.SaveUserCommand)
	if ok == false {
		return nil, fmt.Errorf("SaveUserCommandHandler expects a command of type %T", saveUserCommand)
	}

	// No need to call saveUserCommand.IsValid() - as it's being called by the framework.

	// Delegate fetching data to the repository, which belongs to the Adapters Layer.
	user, err := h.repository.GetUserById(ctx, saveUserCommand.UserID)
	if err != nil {
		return nil, err
	}

	// Delegate updating the email to the user, which is a Domain Entity.
	// The SetEmail method is responsible to detect if a new email was set,
	// and if so, then it will record an EmailSetEvent.
	user.SetEmail(saveUserCommand.Email)

	// Delegate storing data to the repository.
	if err = h.repository.SaveUser(ctx, user); err != nil {
		return nil, err
	}

	// This is where Domain events are being registered,
	// so they can eventually be dispatched to event handlers (if they exist).
	// In our use case the events will be dispatched to the EmailChangedEventHandler.
	h.events = append(h.events, user.Events()...)

	return nil, nil
}

// Commit is responsible for committing the changes performed by the Handle method, such as
// committing a database transaction managed by the repository.
// This method is being called by the framework, so it should not be called from within the Handle method.
func (h *SaveUserCommandHandler) Commit(ctx context.Context) error {
	return h.repository.Commit(ctx)
}

// Rollback is responsible to rollback changes performed by the Handle method, such as
// rollback a database transaction managed by the repository.
// This method is being called by the framework, so it should not be called from within the Handle method.
func (h *SaveUserCommandHandler) Rollback(ctx context.Context) error {
	return h.repository.Rollback(ctx)
}

// Events reports about events.
// These events will be handled by the DDD framework if appropriate event handlers were registered by the bootstrapper.
// This method is being called by the framework, so it should not be called from within the Handle method.
func (h *SaveUserCommandHandler) Events() []ddd.Event {
	return h.events
}

var _ ddd.CommandHandler = (*SaveUserCommandHandler)(nil)
```

##### Registration of the SaveUserCommand with its handler: SaveUserCommandHandler
This happens within the bootstrapper, like so:

```go
package boostrapper

import (
	"context"
	"github.com/vklap/go_ddd/internal/adapters"
	"github.com/vklap/go_ddd/internal/domain/command_model"
	"github.com/vklap/go_ddd/internal/service_layer/command_handlers"
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

// New creates and initializes the bootstrapper.
// Please notice that in a real world scenario you may not require such a custom "DemoBootstrapper",
// and that the adapter instances should most probably be created within the register callbacks, like so:
// 	bs.Bootstrapper.RegisterCommandHandlerFactory(&command_model.SaveUserCommand{}, func() (ddd.CommandHandler, error) {
//		return command_handlers.NewSaveUserCommandHandler(adapters.NewInMemoryRepository()), nil
//	})
func New() *DemoBootstrapper {
	bs := &DemoBootstrapper{
		PubSubClient: adapters.NewInMemoryPubSubClient(),
		Repository:   adapters.NewInMemoryRepository(),
		Bootstrapper: ddd.NewBootstrapper(),
	}
	bs.Bootstrapper.RegisterCommandHandlerFactory(&command_model.SaveUserCommand{}, func() (ddd.CommandHandler, error) {
		return command_handlers.NewSaveUserCommandHandler(bs.Repository), nil
	})
	return bs
}
```

##### Handling the SaveUserCommand by the framework

Based on the above created bootstrapper singleton Instance variable, 
this is how the command should be propagated into the framework: 
```go
var command command_model.SaveUserCommand
...
bootstrapper.Instance.HandleCommand(context.Background(), &command)
```

### But wait, isn't this code over-engineered?

Basically, if this is all the code should do, then this code is arguably too complex.
Yet, what happens when the requirements grow, and you need to handle other tasks, such as:
1. Trigger a verification email to validate the provided email?
2. Notify a KPI Service about the changes - for further analysis
3. Handle other changes, as in a real world scenario the user entity should have much more properties - 
   where each property change might require triggering other actions (a.k.a. `Domain Events`)

The code might quickly look like this:
```go

func (h *SaveUserCommandHandler) Handle(ctx context.Context, command ddd.Command) (any, error) {
    saveUserCommand, ok := command.(*command_model.SaveUserCommand)
    if ok == false {
        return nil, fmt.Errorf("SaveUserCommandHandler expects a command of type %T", saveUserCommand)
    }

	user, err := h.repository.GetUserById(ctx, saveUserCommand.UserID)
    if err != nil {
        return nil, err
    }

    user.SetEmail(saveUserCommand.Email)
    if err = h.repository.SaveUser(ctx, user); err != nil {
        return nil, err
    }

    // Side effects...
    if user.EmailChanged() {
        if err := h.PubSub.requestEmailVerification(user); err != nil {
            ...
        }
        if err := h.PubSub.requestEmailVerification(user); err != nil {
            ...
        }       
    }

    // Side effects...
    user.SetPhoneNumber(...)
    if user.PhoneChanged() {
        if err := h.PubSub.requestPhoneVerification(user); err != nil {
            ...
        }
        if err := h.PubSub.requestPhoneVerification(user); err != nil {
            ...
        }       
    }

    ...

    return nil, nil
}

```

The above code will contain lots side effects, and will defeat the SRP (Single Responsibility Principle) 
for which it was created - which is to save the new user details.
Even worse, it will sooner than later become spaghetti code - that will be a nightmare to maintain and unit test. 

### Event-Driven Architecture with EventHandlers to the Rescue
All the above side effects should best be extracted out of the above code, and handled within other handlers. 
These handlers will be handled in the same way as the command handler, 
i.e. within units of work of their own - and may trigger other events which will be handled by the framework.

Here are 2 sample event handlers:

#### EmailSetEventHandler that will trigger a KPIEvent that will be handled by the KPIEventHandler
```go
package event_handlers

import (
	"context"
	"fmt"
	"github.com/vklap/go_ddd/internal/adapters"
	"github.com/vklap/go_ddd/internal/domain/command_model"
	"github.com/vklap/go_ddd/pkg/ddd"
)

// EmailSetEventHandler implements ddd.EventHandler.
type EmailSetEventHandler struct {
	pubSubClient adapters.PubSubClient
	events       []ddd.Event
}

// NewEmailSetEventHandler is a constructor function to be used by the Bootstrapper.
func NewEmailSetEventHandler(pubSubClient adapters.PubSubClient) *EmailSetEventHandler {
	return &EmailSetEventHandler{pubSubClient: pubSubClient, events: make([]ddd.Event, 0)}
}

// Handle manages the business logic flow, and is the glue between the Domain and the Adapters.
func (h *EmailSetEventHandler) Handle(ctx context.Context, event ddd.Event) error {
	e, ok := event.(*command_model.EmailSetEvent)
	if ok == false {
		panic(fmt.Sprintf("failed to handle email set: want %T, got %T", &command_model.EmailSetEvent{}, e))
	}
	if err := h.pubSubClient.NotifyEmailChanged(ctx, e.UserID, e.NewEmail, e.OriginalEmail); err != nil {
		return err
	}
	h.events = append(h.events, &command_model.KPIEvent{Action: e.EventName(), Data: fmt.Sprintf("%v", e)})
	return nil
}

// Commit is responsible for committing the changes performed by the Handle method, such as
// committing a database transaction managed by the repository.
// This method is being called by the framework, so it should not be called from within the Handle method.
func (h *EmailSetEventHandler) Commit(ctx context.Context) error {
	return h.pubSubClient.Commit(ctx)
}

// Rollback is responsible to rollback changes performed by the Handle method, such as
// rollback a database transaction managed by the repository.
// This method is being called by the framework, so it should not be called from within the Handle method.
func (h *EmailSetEventHandler) Rollback(ctx context.Context) error {
	return h.pubSubClient.Rollback(ctx)
}

// Events reports about events.
// These events will be handled by the DDD framework if appropriate event handlers were registered by the bootstrapper.
// This method is being called by the framework, so it should not be called from within the Handle method.
func (h *EmailSetEventHandler) Events() []ddd.Event {
	return h.events
}

var _ ddd.EventHandler = (*EmailSetEventHandler)(nil)
```

##### KPIEventHandler
```go
package event_handlers

import (
	"context"
	"fmt"
	"github.com/vklap/go_ddd/internal/adapters"
	"github.com/vklap/go_ddd/internal/domain/command_model"
	"github.com/vklap/go_ddd/pkg/ddd"
)

// KPIEventHandler implements ddd.EventHandler.
type KPIEventHandler struct {
	pubSubClient adapters.PubSubClient
}

// NewKPIEventHandler is a constructor function to be used by the Bootstrapper.
func NewKPIEventHandler(pubSubClient adapters.PubSubClient) *KPIEventHandler {
	return &KPIEventHandler{pubSubClient: pubSubClient}
}

// Handle manages the business logic flow, and is the glue between the Domain and the Adapters.
func (h *KPIEventHandler) Handle(ctx context.Context, event ddd.Event) error {
	e, ok := event.(*command_model.KPIEvent)
	if ok == false {
		panic(fmt.Sprintf("failed to handle KPI event: want %T, got %T", &command_model.KPIEvent{}, e))
	}
	return h.pubSubClient.NotifyKPIService(ctx, e)
}

// Commit is responsible for committing the changes performed by the Handle method, such as
// committing a database transaction managed by the repository.
// This method is being called by the framework, so it should not be called from within the Handle method.
func (h *KPIEventHandler) Commit(ctx context.Context) error {
	return h.pubSubClient.Commit(ctx)
}

// Rollback is responsible to rollback changes performed by the Handle method, such as
// rollback a database transaction managed by the repository.
// This method is being called by the framework, so it should not be called from within the Handle method.
func (h *KPIEventHandler) Rollback(ctx context.Context) error {
	return h.pubSubClient.Rollback(ctx)
}

// Events reports about events.
// These events will be handled by the DDD framework if appropriate event handlers were registered by the bootstrapper.
// This method is being called by the framework, so it should not be called from within the Handle method.
func (h *KPIEventHandler) Events() []ddd.Event {
	return nil
}

var _ ddd.EventHandler = (*KPIEventHandler)(nil)

```

### Advantages of applying the above-mentioned Domain-Driven Design Tactical Patterns

- A clear separation of concerns between the business rules (which reside solely inside the domain layer), 
  the application flows (which reside in the service layer) and the IO related operations - 
  such as communication with databases/web services/file system (which reside in the adapters layer)

- This separation of concerns make this kind of code very suitable for unit & integration tests - 
  the service & domain layers can be fully unit tested and the adapter layer can easily 
  be integration tested (without being concerned with any business logic leaking from the other layers - 
  so that the integration tests can remain simple)
  
- A common code base structure makes it much easier for other developers, 
  who are aware of this structure, to get into the code.

## Links

- [pkg.go.dev](https://pkg.go.dev/github.com/vklap/go_ddd)
- [README.md](https://github.com/vklap/go_ddd/blob/main/README.md)