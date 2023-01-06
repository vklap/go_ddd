package ddd_test

import (
	"context"
	"github.com/vklap/go_ddd/internal/domain/command_model"
	"github.com/vklap/go_ddd/internal/entrypoints/boostrapper"
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

type eventWithoutHandler struct{}

func (e *eventWithoutHandler) EventName() string {
	return "eventWithoutHandler"
}

func TestChangeEmail(t *testing.T) {
	const userID = "1"
	const originalEmail = "kamel.amit@thaabet.sy"
	const newEmail = "eli.cohen@mossad.gov.il"

	data := []struct {
		command            *command_model.ChangeEmailCommand
		commitCalled       bool
		commitShouldFail   bool
		errorMessage       string
		errorStatusCode    string
		failed             bool
		name               string
		rollbackCalled     bool
		rollbackShouldFail bool
		userExists         bool
	}{
		{
			command:            &command_model.ChangeEmailCommand{NewEmail: newEmail, UserID: userID},
			commitCalled:       true,
			commitShouldFail:   false,
			errorMessage:       "",
			errorStatusCode:    "",
			failed:             false,
			name:               "change email succeeds",
			rollbackCalled:     false,
			rollbackShouldFail: false,
			userExists:         true,
		},
		{
			command:            &command_model.ChangeEmailCommand{NewEmail: newEmail, UserID: "i-do-not-exist"},
			commitCalled:       false,
			commitShouldFail:   false,
			errorMessage:       "",
			errorStatusCode:    ddd.StatusCodeNotFound,
			failed:             true,
			name:               "user does not exist",
			rollbackCalled:     true,
			rollbackShouldFail: false,
			userExists:         false,
		},
		{
			command:            &command_model.ChangeEmailCommand{NewEmail: "", UserID: userID},
			commitCalled:       false,
			commitShouldFail:   false,
			errorMessage:       "email cannot be empty",
			errorStatusCode:    ddd.StatusCodeBadRequest,
			failed:             true,
			name:               "missing email validation",
			rollbackCalled:     false,
			rollbackShouldFail: false,
			userExists:         true,
		},
		{
			command:            &command_model.ChangeEmailCommand{NewEmail: newEmail, UserID: ""},
			commitCalled:       false,
			commitShouldFail:   false,
			errorMessage:       "user ID cannot be empty",
			errorStatusCode:    ddd.StatusCodeBadRequest,
			failed:             true,
			name:               "missing userID validation",
			rollbackCalled:     false,
			rollbackShouldFail: false,
			userExists:         true,
		},
		{
			command:            &command_model.ChangeEmailCommand{NewEmail: newEmail, UserID: userID},
			commitCalled:       false,
			commitShouldFail:   false,
			errorMessage:       "does not exist",
			errorStatusCode:    ddd.StatusCodeNotFound,
			failed:             true,
			name:               "user does not exist and rollback does not fail",
			rollbackCalled:     true,
			rollbackShouldFail: false,
			userExists:         false,
		},
		{
			command:            &command_model.ChangeEmailCommand{NewEmail: newEmail, UserID: userID},
			commitCalled:       false,
			commitShouldFail:   false,
			errorMessage:       "rollback failed",
			errorStatusCode:    "",
			failed:             true,
			name:               "user does not exist and rollback failed",
			rollbackCalled:     true,
			rollbackShouldFail: true,
			userExists:         false,
		},
		{
			command:            &command_model.ChangeEmailCommand{NewEmail: newEmail, UserID: userID},
			commitCalled:       true,
			commitShouldFail:   true,
			errorMessage:       "commit failed",
			errorStatusCode:    "",
			failed:             true,
			name:               "commit failed",
			rollbackCalled:     false,
			rollbackShouldFail: false,
			userExists:         true,
		},
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			fb := boostrapper.New()
			aUser := &command_model.User{}
			aUser.SetEmail(originalEmail)
			aUser.SetID(userID)
			if d.userExists {
				fb.Repository.UsersById[aUser.ID()] = aUser
			}
			fb.Repository.RollbackShouldFail = d.rollbackShouldFail
			fb.Repository.CommitShouldFail = d.commitShouldFail

			result, err := fb.Bootstrapper.HandleCommand(context.Background(), d.command)

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
	fb := boostrapper.New()
	command := &notSupportedCommand{}

	_, _ = fb.Bootstrapper.HandleCommand(context.Background(), command)
}

func TestHandlerReceivedCommandOfWrongType(t *testing.T) {
	fb := boostrapper.New()
	command := &notSupportedCommand{}
	fb.Bootstrapper.RegisterCommandHandlerFactory(command, func() (ddd.CommandHandler, error) {
		return command_handlers.NewChangeEmailCommandHandler(fb.Repository), nil
	})

	result, err := fb.Bootstrapper.HandleCommand(context.Background(), command)

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
	data := []struct {
		commitCalled   bool
		commitFailed   bool
		name           string
		rollbackCalled bool
		rollbackFailed bool
	}{
		{
			commitCalled:   false,
			commitFailed:   false,
			name:           "regular failure",
			rollbackCalled: true,
			rollbackFailed: false,
		},
		{
			commitCalled:   false,
			commitFailed:   false,
			name:           "rollback failure",
			rollbackCalled: true,
			rollbackFailed: true,
		},
		{
			commitCalled:   false,
			commitFailed:   true,
			name:           "commit failure",
			rollbackCalled: true,
			rollbackFailed: false,
		},
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			fb := boostrapper.New()
			const userID = "1"
			const originalEmail = "kamel.amit@thaabet.sy"
			const newEmail = "eli.cohen@mossad.gov.il"
			aUser := &command_model.User{}
			aUser.SetEmail(originalEmail)
			aUser.SetID(userID)
			fb.PubSubClient.RollbackShouldFail = d.rollbackFailed
			fb.PubSubClient.CommitCalled = d.commitCalled
			fb.PubSubClient.CommitShouldFail = d.commitFailed
			fb.PubSubClient.RollbackCalled = d.rollbackCalled

			command := &command_model.ChangeEmailCommand{NewEmail: newEmail, UserID: userID}

			_, err := fb.Bootstrapper.HandleCommand(context.Background(), command)

			if err == nil {
				t.Error("want error, got nil")
			}

			if fb.PubSubClient.CommitCalled != d.commitCalled {
				t.Errorf("want pubsub commit called %v, got %v", d.commitCalled, fb.PubSubClient.CommitCalled)
			}
			if fb.PubSubClient.RollbackCalled != d.rollbackCalled {
				t.Errorf("want pubsub rollback called %v, got %v", d.rollbackCalled, fb.PubSubClient.RollbackCalled)
			}
		})
	}
}

func assertSuccess(t *testing.T, err error, result any, fb *boostrapper.DemoBootstrapper, newEmail string, userID string, originalEmail string) {
	if err != nil {
		t.Errorf("want no error, got %v", err)
	}
	if result != nil {
		t.Errorf("want result nil, got %v", result)
	}
	user := fb.Repository.UsersById[userID]
	if user.Email() != newEmail {
		t.Errorf("want email %q, got %q", newEmail, user.Email())
	}
	if fb.Repository.CommitCalled != true {
		t.Errorf("want adapters.Repository commitCalled true, got %v", fb.Repository.CommitCalled)
	}
	if !fb.PubSubClient.NotifyEmailChangedCalled {
		t.Error("want notify email changed to be called")
	}
	if !fb.PubSubClient.NotifySlackCalled {
		t.Error("want notify slack to be called")
	}
}

func assertFailure(t *testing.T, err error, d struct {
	command            *command_model.ChangeEmailCommand
	commitCalled       bool
	commitShouldFail   bool
	errorMessage       string
	errorStatusCode    string
	failed             bool
	name               string
	rollbackCalled     bool
	rollbackShouldFail bool
	userExists         bool
}, fb *boostrapper.DemoBootstrapper) {
	dddError, ok := err.(*ddd.Error)
	if ok == true {
		if dddError.StatusCode() != d.errorStatusCode {
			t.Errorf("want err status code %q, got %q", d.errorStatusCode, dddError.StatusCode())
		}
		if strings.Contains(dddError.Error(), d.errorMessage) == false {
			t.Errorf("want %q in dddError, got %q", d.errorMessage, dddError.Error())
		}
	}
	if fb.Repository.RollbackCalled != d.rollbackCalled {
		t.Errorf("want repository rollback to be %v, got %v", d.rollbackCalled, fb.Repository.RollbackCalled)
	}
	if fb.Repository.CommitCalled != d.commitCalled {
		t.Errorf("want repository commit to be %v, got %v", d.commitCalled, fb.Repository.CommitCalled)
	}
	if d.rollbackShouldFail {
		if !strings.Contains(err.Error(), d.errorMessage) {
			t.Errorf("want error with %q, got %q", d.errorMessage, err.Error())
		}
	}
	if d.commitShouldFail {
		if !strings.Contains(err.Error(), "commit failed") {
			t.Errorf("want error with \"commit failed\", got: %q", err.Error())
		}
	}
}
