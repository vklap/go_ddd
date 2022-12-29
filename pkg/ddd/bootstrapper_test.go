package ddd_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/vklap/go_ddd/internal/domain/command_model"
	"github.com/vklap/go_ddd/internal/service_layer/command_handlers"
	"github.com/vklap/go_ddd/pkg/ddd"
	"strings"
	"testing"
)

type notSupportedCommand struct{}

func (c *notSupportedCommand) IsValid() error {
	return nil
}

func (c *notSupportedCommand) CommandName() string {
	return "notSupportedCommand"
}

// Triggered by emailChangedEventHandler
type mossadEmailCreatedEvent struct{}

func (e *mossadEmailCreatedEvent) EventName() string {
	return "mossadEmailCreatedEvent"
}

type eventWithoutHandler struct{}

func (e *eventWithoutHandler) EventName() string {
	return "eventWithoutHandler"
}

type stubRepository struct {
	user           *command_model.User
	entities       []ddd.Entity
	commitCalled   bool
	rollbackCalled bool
	getUserById    func(ctx context.Context, id string) (*command_model.User, error)
	saveUser       func(ctx context.Context, user *command_model.User) error
	commit         func(ctx context.Context) error
	rollback       func(ctx context.Context) error
}

func (r *stubRepository) GetUserById(ctx context.Context, id string) (*command_model.User, error) {
	return r.getUserById(ctx, id)
}

func (r *stubRepository) SaveUser(ctx context.Context, user *command_model.User) error {
	return r.saveUser(ctx, user)
}
func (r *stubRepository) Commit(ctx context.Context) error {
	return r.commit(ctx)
}
func (r *stubRepository) Rollback(ctx context.Context) error {
	return r.rollback(ctx)
}

type fakeBootstrapper struct {
	b                              *ddd.Bootstrapper
	ChangeEmailCommandHandler      *command_handlers.ChangeEmailCommandHandler
	EmailChangedEventHandler       *stubEventHandler
	MossadEmailCreatedEventHandler *stubEventHandler
	Repository                     *stubRepository
}

type stubEventHandler struct {
	event          ddd.Event
	commitCalled   bool
	rollbackCalled bool
	entities       []ddd.Entity

	handle   func(ctx context.Context, event ddd.Event) error
	commit   func(ctx context.Context) error
	rollback func(ctx context.Context) error
	events   func() []ddd.Event
}

func (h *stubEventHandler) Handle(ctx context.Context, event ddd.Event) error {
	return h.handle(ctx, event)
}

func (h *stubEventHandler) Commit(ctx context.Context) error {
	return h.commit(ctx)
}

func (h *stubEventHandler) Rollback(ctx context.Context) error {
	return h.rollback(ctx)
}

func (h *stubEventHandler) Events() []ddd.Event {
	return h.events()
}

func newFakeBootstrapper() *fakeBootstrapper {
	repo := &stubRepository{}

	fb := &fakeBootstrapper{
		b:                              ddd.NewBootstrapper(),
		Repository:                     repo,
		ChangeEmailCommandHandler:      command_handlers.NewChangeEmailCommandHandler(repo),
		EmailChangedEventHandler:       &stubEventHandler{},
		MossadEmailCreatedEventHandler: &stubEventHandler{},
	}
	fb.b.RegisterCommandHandlerFactory(&command_model.ChangeEmailCommand{}, func() (ddd.CommandHandler, error) {
		return fb.ChangeEmailCommandHandler, nil
	})
	fb.b.RegisterEventHandlerFactory(&command_model.EmailChangedEvent{}, func() (ddd.EventHandler, error) {
		return fb.EmailChangedEventHandler, nil
	})
	fb.b.RegisterEventHandlerFactory(&mossadEmailCreatedEvent{}, func() (ddd.EventHandler, error) {
		return fb.MossadEmailCreatedEventHandler, nil
	})
	return fb
}

func TestChangeEmail(t *testing.T) {
	const userID = "1"
	const originalEmail = "kamel.amit@thaabet.sy"
	const newEmail = "eli.cohen@mossad.gov.il"
	aUser := &command_model.User{}
	aUser.SetEmail(originalEmail)
	aUser.SetID(userID)

	data := []struct {
		name          string
		command       *command_model.ChangeEmailCommand
		getUserById   func(ctx context.Context, id string) (*command_model.User, error)
		failed        bool
		expectedError *ddd.Error
	}{
		{
			name:          "command_model.User exists",
			command:       &command_model.ChangeEmailCommand{NewEmail: newEmail, UserID: userID},
			getUserById:   nil,
			failed:        false,
			expectedError: nil,
		},
		{
			name:    "command_model.User does not exist",
			command: &command_model.ChangeEmailCommand{NewEmail: newEmail, UserID: userID},
			getUserById: func(ctx context.Context, id string) (*command_model.User, error) {
				err := ddd.NewError(fmt.Sprintf("command_model.User with id %q does not exist", id), ddd.StatusCodeNotFound)
				return nil, err
			},
			failed:        true,
			expectedError: ddd.NewError("does not exist", ddd.StatusCodeNotFound),
		},
		{
			name:          "missing email validation",
			command:       &command_model.ChangeEmailCommand{NewEmail: "", UserID: userID},
			getUserById:   nil,
			failed:        true,
			expectedError: ddd.NewError("email", ddd.StatusCodeBadRequest),
		},
		{
			name:          "missing command_model.User validation",
			command:       &command_model.ChangeEmailCommand{NewEmail: newEmail, UserID: ""},
			getUserById:   nil,
			failed:        true,
			expectedError: ddd.NewError("userID", ddd.StatusCodeBadRequest),
		},
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			fb := newFakeBootstrapper()
			setupRepository(fb, aUser)
			if d.getUserById != nil {
				fb.Repository.getUserById = d.getUserById
			}
			setupEmailChangedEventHandler(fb)
			setupEmailChangedHandledEventHandler(fb)

			result, err := fb.b.HandleCommand(context.Background(), d.command)

			if d.failed {
				assertFailure(t, err, d, fb)
			} else {
				assertSuccess(t, err, result, fb, newEmail, userID, originalEmail)
			}
		})
	}
}

func TestCommandWithoutRegisteredHandler(t *testing.T) {
	defer func() {
		if v := recover(); v == nil {
			t.Error("want panic, got not error")
		}
	}()
	fb := newFakeBootstrapper()
	command := &notSupportedCommand{}

	_, _ = fb.b.HandleCommand(context.Background(), command)
}

func TestHandlerReceivedCommandOfWrongType(t *testing.T) {
	fb := newFakeBootstrapper()
	command := &notSupportedCommand{}
	fb.b.RegisterCommandHandlerFactory(command, func() (ddd.CommandHandler, error) {
		return fb.ChangeEmailCommandHandler, nil
	})
	setupRepository(fb, &command_model.User{})

	result, err := fb.b.HandleCommand(context.Background(), command)

	if result != nil {
		t.Errorf("want resut nil, got %v", result)
	}
	if err == nil {
		t.Error("want error not nil, got nil")
	}
	expectedCommand := &command_model.ChangeEmailCommand{}
	if strings.Contains(err.Error(), expectedCommand.CommandName()) != true {
		t.Errorf("want error with %q, got %q", expectedCommand.CommandName(), err.Error())
	}
}

func TestHandleEventFailure(t *testing.T) {
	fb := newFakeBootstrapper()
	const userID = "1"
	const originalEmail = "kamel.amit@thaabet.sy"
	const newEmail = "eli.cohen@mossad.gov.il"
	aUser := &command_model.User{}
	aUser.SetEmail(originalEmail)
	aUser.SetID(userID)
	setupRepository(fb, aUser)
	setupEmailChangedEventHandler(fb)
	setupEmailChangedHandledEventHandler(fb)
	fb.EmailChangedEventHandler.handle = func(ctx context.Context, event ddd.Event) error {
		return errors.New("the spy that did not come home")
	}
	command := &command_model.ChangeEmailCommand{NewEmail: newEmail, UserID: userID}

	_, err := fb.b.HandleCommand(context.Background(), command)

	if err == nil {
		t.Error("want error, got nil")
	}

	if fb.EmailChangedEventHandler.commitCalled == true {
		t.Error("want email changed event handler to not be committed")
	}
	if fb.EmailChangedEventHandler.rollbackCalled != true {
		t.Error("want email changed event handler to be rolled back")
	}
	if fb.MossadEmailCreatedEventHandler.commitCalled == true {
		t.Error("want mossad email created event handler to not be called")
	}
}

func assertSuccess(t *testing.T, err error, result any, fb *fakeBootstrapper, newEmail string, userID string, originalEmail string) {
	if err != nil {
		t.Errorf("want no error, got %v", err)
	}
	if result != nil {
		t.Errorf("want result nil, got %v", result)
	}
	if fb.Repository.user.Email() != newEmail {
		t.Errorf("want email %q, got %q", newEmail, fb.Repository.user.Email())
	}
	if fb.Repository.commitCalled != true {
		t.Errorf("want adapters.Repository commitCalled true, got %v", fb.Repository.commitCalled)
	}
	if fb.EmailChangedEventHandler.commitCalled != true {
		t.Errorf("want event handler commitCalled true, got %v", fb.EmailChangedEventHandler.commitCalled)
	}
	event, ok := fb.EmailChangedEventHandler.event.(*command_model.EmailChangedEvent)
	if ok == false {
		t.Errorf("want event %T, got %T", command_model.EmailChangedEvent{}, event)
	}
	if event.UserID != userID {
		t.Errorf("want event UserID %q, got %q", userID, event.UserID)
	}
	if event.OriginalEmail != originalEmail {
		t.Errorf("want event OriginalEmail %q, got %q", originalEmail, event.OriginalEmail)
	}
	if event.NewEmail != newEmail {
		t.Errorf("want event newEmail %q, got %q", newEmail, event.NewEmail)
	}
	if fb.MossadEmailCreatedEventHandler.commitCalled != true {
		t.Errorf("expected EmailChangedHandledEvent commitCalled %v, got %v", true, fb.MossadEmailCreatedEventHandler.commitCalled)
	}
}

func assertFailure(t *testing.T, err error, d struct {
	name          string
	command       *command_model.ChangeEmailCommand
	getUserById   func(ctx context.Context, id string) (*command_model.User, error)
	failed        bool
	expectedError *ddd.Error
}, fb *fakeBootstrapper) {
	dddError, ok := err.(*ddd.Error)
	if ok == true {
		if dddError.StatusCode() != d.expectedError.StatusCode() {
			t.Errorf("want err status code %q, got %q", d.expectedError.StatusCode(), dddError.StatusCode())
		}
		if strings.Contains(dddError.Error(), d.expectedError.Error()) == false {
			t.Errorf("want %q in dddError, got %q", d.expectedError.Error(), dddError.Error())
		}
	} else {
		t.Errorf("want err %T, got %T", d.expectedError, dddError)
	}

	if d.expectedError.StatusCode() != ddd.StatusCodeBadRequest {
		if fb.Repository.rollbackCalled != true {
			t.Errorf("expected adapters.Repository rollbackCalled to be true, got false")
		}
	}
}

func setupEmailChangedHandledEventHandler(fb *fakeBootstrapper) {
	fb.MossadEmailCreatedEventHandler.handle = func(ctx context.Context, event ddd.Event) error {
		fb.MossadEmailCreatedEventHandler.event = event
		return nil
	}
	fb.MossadEmailCreatedEventHandler.rollback = func(ctx context.Context) error {
		fb.MossadEmailCreatedEventHandler.rollbackCalled = true
		return nil
	}
	fb.MossadEmailCreatedEventHandler.commit = func(ctx context.Context) error {
		fb.MossadEmailCreatedEventHandler.commitCalled = true
		return nil
	}
	fb.MossadEmailCreatedEventHandler.events = func() []ddd.Event {
		return nil
	}
}

func setupEmailChangedEventHandler(fb *fakeBootstrapper) {
	fb.EmailChangedEventHandler.handle = func(ctx context.Context, event ddd.Event) error {
		fb.EmailChangedEventHandler.event = event
		return nil
	}
	fb.EmailChangedEventHandler.rollback = func(ctx context.Context) error {
		fb.EmailChangedEventHandler.rollbackCalled = true
		return nil
	}
	fb.EmailChangedEventHandler.commit = func(ctx context.Context) error {
		fb.EmailChangedEventHandler.commitCalled = true
		return nil
	}
	fb.EmailChangedEventHandler.events = func() []ddd.Event {
		// eventWithoutHandler should be ignored silently
		return []ddd.Event{&mossadEmailCreatedEvent{}, &eventWithoutHandler{}}
	}
}

func setupRepository(fb *fakeBootstrapper, aUser *command_model.User) {
	fb.Repository.user = aUser
	fb.Repository.saveUser = func(ctx context.Context, user *command_model.User) error {
		fb.Repository.user = user
		fb.Repository.entities = append(fb.Repository.entities, user)
		return nil
	}
	fb.Repository.getUserById = func(ctx context.Context, id string) (*command_model.User, error) {
		if fb.Repository.user.ID() != id {
			return nil, fmt.Errorf("command_model.User not found (id = %q)", id)
		}
		return fb.Repository.user, nil
	}
	fb.Repository.commit = func(ctx context.Context) error {
		fb.Repository.commitCalled = true
		return nil
	}
	fb.Repository.rollback = func(ctx context.Context) error {
		fb.Repository.rollbackCalled = true
		return nil
	}
}
