package ddd_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/vklap.go-ddd/pkg/ddd"
	"strings"
	"testing"
)

type changeEmailCommand struct {
	UserID string
	Email  string
}

func (c *changeEmailCommand) IsValid() error {
	if c.UserID == "" {
		return ddd.NewError("userID cannot be empty", ddd.StatusCodeBadRequest)
	}
	if c.Email == "" {
		return ddd.NewError("email cannot be empty", ddd.StatusCodeBadRequest)
	}
	return nil
}

func (c *changeEmailCommand) CommandName() string {
	return "changeEmailCommand"
}

type notSupportedCommand struct{}

func (c *notSupportedCommand) IsValid() error {
	return nil
}

func (c *notSupportedCommand) CommandName() string {
	return "notSupportedCommand"
}

type emailChangedEvent struct {
	OriginalEmail string
	NewEmail      string
}

func (e *emailChangedEvent) EventName() string {
	return "emailChangedEvent"
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

type user struct {
	ddd.BaseEntity
	email string
}

func (u *user) Email() string {
	return u.email
}

func (u *user) SetEmail(value string) {
	if u.email != "" {
		u.AddEvent(&emailChangedEvent{NewEmail: value, OriginalEmail: u.email})
	}
	u.email = value
}

type repository interface {
	GetUserById(ctx context.Context, id string) (*user, error)
	SaveUser(ctx context.Context, user *user) error
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
	SavedEntities() []ddd.Entity
}

type stubRepository struct {
	user           *user
	entities       []ddd.Entity
	commitCalled   bool
	rollbackCalled bool
	getUserById    func(ctx context.Context, id string) (*user, error)
	saveUser       func(ctx context.Context, user *user) error
	commit         func(ctx context.Context) error
	rollback       func(ctx context.Context) error
	savedEntities  func() []ddd.Entity
}

func (r *stubRepository) GetUserById(ctx context.Context, id string) (*user, error) {
	return r.getUserById(ctx, id)
}

func (r *stubRepository) SaveUser(ctx context.Context, user *user) error {
	return r.saveUser(ctx, user)
}
func (r *stubRepository) Commit(ctx context.Context) error {
	return r.commit(ctx)
}
func (r *stubRepository) Rollback(ctx context.Context) error {
	return r.rollback(ctx)
}
func (r *stubRepository) SavedEntities() []ddd.Entity {
	return r.savedEntities()
}

type changeEmailCommandHandler struct {
	r             repository
	savedEntities []ddd.Entity
}

func (h *changeEmailCommandHandler) Handle(ctx context.Context, command ddd.Command) (any, error) {
	if err := command.IsValid(); err != nil {
		return nil, err
	}
	changeEmailCommand, ok := command.(*changeEmailCommand)
	if ok == false {
		return nil, fmt.Errorf("changeEmailCommandHandler expects a command of type %T", changeEmailCommand)
	}

	user, err := h.r.GetUserById(ctx, changeEmailCommand.UserID)
	if err != nil {
		return nil, err
	}
	user.SetEmail(changeEmailCommand.Email)
	if err = h.r.SaveUser(ctx, user); err != nil {
		return nil, err
	}
	h.savedEntities = append(h.savedEntities, user)
	return nil, nil
}

func (h *changeEmailCommandHandler) Commit(ctx context.Context) error {
	return h.r.Commit(ctx)
}

func (h *changeEmailCommandHandler) Rollback(ctx context.Context) error {
	return h.r.Rollback(ctx)
}

func (h *changeEmailCommandHandler) SavedEntities() []ddd.Entity {
	return h.savedEntities
}

type fakeBootstrapper struct {
	b                              *ddd.Bootstrapper
	ChangeEmailCommandHandler      *changeEmailCommandHandler
	EmailChangedEventHandler       *stubEventHandler
	MossadEmailCreatedEventHandler *stubEventHandler
	Repository                     *stubRepository
}

type stubEventHandler struct {
	event          ddd.Event
	commitCalled   bool
	rollbackCalled bool
	entities       []ddd.Entity

	handle        func(ctx context.Context, event ddd.Event) error
	commit        func(ctx context.Context) error
	rollback      func(ctx context.Context) error
	savedEntities func() []ddd.Entity
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

func (h *stubEventHandler) SavedEntities() []ddd.Entity {
	return h.savedEntities()
}

func newFakeBootstrapper() *fakeBootstrapper {
	repo := &stubRepository{}

	fb := &fakeBootstrapper{
		b:                              ddd.NewBootstrapper(),
		Repository:                     repo,
		ChangeEmailCommandHandler:      &changeEmailCommandHandler{r: repo},
		EmailChangedEventHandler:       &stubEventHandler{},
		MossadEmailCreatedEventHandler: &stubEventHandler{},
	}
	fb.b.RegisterCommandHandlerFactory(&changeEmailCommand{}, func() (ddd.CommandHandler, error) {
		return fb.ChangeEmailCommandHandler, nil
	})
	fb.b.RegisterEventHandlerFactory(&emailChangedEvent{}, func() (ddd.EventHandler, error) {
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
	aUser := &user{email: originalEmail}
	aUser.SetID(userID)

	data := []struct {
		name          string
		command       *changeEmailCommand
		getUserById   func(ctx context.Context, id string) (*user, error)
		failed        bool
		expectedError *ddd.Error
	}{
		{
			name:          "user exists",
			command:       &changeEmailCommand{Email: newEmail, UserID: userID},
			getUserById:   nil,
			failed:        false,
			expectedError: nil,
		},
		{
			name:    "user does not exist",
			command: &changeEmailCommand{Email: newEmail, UserID: userID},
			getUserById: func(ctx context.Context, id string) (*user, error) {
				err := ddd.NewError(fmt.Sprintf("user with id %q does not exist", id), ddd.StatusCodeNotFound)
				return nil, err
			},
			failed:        true,
			expectedError: ddd.NewError("does not exist", ddd.StatusCodeNotFound),
		},
		{
			name:          "missing email validation",
			command:       &changeEmailCommand{Email: "", UserID: userID},
			getUserById:   nil,
			failed:        true,
			expectedError: ddd.NewError("email", ddd.StatusCodeBadRequest),
		},
		{
			name:          "missing user validation",
			command:       &changeEmailCommand{Email: newEmail, UserID: ""},
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
	setupRepository(fb, &user{})

	result, err := fb.b.HandleCommand(context.Background(), command)

	if result != nil {
		t.Errorf("want resut nil, got %v", result)
	}
	if err == nil {
		t.Error("want error not nil, got nil")
	}
	expectedCommand := &changeEmailCommand{}
	if strings.Contains(err.Error(), expectedCommand.CommandName()) != true {
		t.Errorf("want error with %q, got %q", expectedCommand.CommandName(), err.Error())
	}
}

func TestHandleEventFailure(t *testing.T) {
	fb := newFakeBootstrapper()
	const userID = "1"
	const originalEmail = "kamel.amit@thaabet.sy"
	const newEmail = "eli.cohen@mossad.gov.il"
	aUser := &user{email: originalEmail}
	aUser.SetID(userID)
	setupRepository(fb, aUser)
	setupEmailChangedEventHandler(fb)
	setupEmailChangedHandledEventHandler(fb)
	fb.EmailChangedEventHandler.handle = func(ctx context.Context, event ddd.Event) error {
		return errors.New("the spy that did not come home")
	}
	command := &changeEmailCommand{Email: newEmail, UserID: userID}

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
		t.Errorf("want repository commitCalled true, got %v", fb.Repository.commitCalled)
	}
	savedEntities := fb.Repository.SavedEntities()
	if len(savedEntities) != 1 {
		t.Errorf("want 1 repository savedEntites, got %v", len(savedEntities))
	}
	savedEntity := savedEntities[0]
	if savedEntity.ID() != userID {
		t.Errorf("want repository savedEnttiy id %q, got %q", userID, savedEntity.ID())
	}
	if fb.EmailChangedEventHandler.commitCalled != true {
		t.Errorf("want event handler commitCalled true, got %v", fb.EmailChangedEventHandler.commitCalled)
	}
	event, ok := fb.EmailChangedEventHandler.event.(*emailChangedEvent)
	if ok == false {
		t.Errorf("want event %T, got %T", emailChangedEvent{}, event)
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
	command       *changeEmailCommand
	getUserById   func(ctx context.Context, id string) (*user, error)
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
			t.Errorf("expected repository rollbackCalled to be true, got false")
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
	fb.MossadEmailCreatedEventHandler.savedEntities = func() []ddd.Entity {
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
	fb.EmailChangedEventHandler.savedEntities = func() []ddd.Entity {
		entity := &user{}
		entity.AddEvent(&mossadEmailCreatedEvent{})
		// eventWithoutHandler should be ignored silently
		entity.AddEvent(&eventWithoutHandler{})
		return []ddd.Entity{entity}
	}
}

func setupRepository(fb *fakeBootstrapper, aUser *user) {
	fb.Repository.user = aUser
	fb.Repository.saveUser = func(ctx context.Context, user *user) error {
		fb.Repository.user = user
		fb.Repository.entities = append(fb.Repository.entities, user)
		return nil
	}
	fb.Repository.getUserById = func(ctx context.Context, id string) (*user, error) {
		if fb.Repository.user.ID() != id {
			return nil, fmt.Errorf("user not found (id = %q)", id)
		}
		return fb.Repository.user, nil
	}
	fb.Repository.savedEntities = func() []ddd.Entity {
		return fb.Repository.entities
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
