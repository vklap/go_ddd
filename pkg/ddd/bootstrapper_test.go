package ddd_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/vklap.go-ddd/pkg/ddd"
	"testing"
)

type User struct {
	ddd.BaseEntity
	email string
}

func (u *User) Email() string {
	return u.email
}

func (u *User) SetEmail(value string) {
	u.email = value
}

type ChangeEmailCommand struct {
	UserID string
	Email  string
}

func (c *ChangeEmailCommand) IsValid() error {
	if c.UserID == "" {
		return errors.New("userID cannot be empty")
	}
	if c.Email == "" {
		return errors.New("email cannot be empty")
	}
	return nil
}

func (c *ChangeEmailCommand) CommandName() string {
	return "ChangeEmailCommand"
}

type Repository interface {
	GetUserById(ctx context.Context, id string) (*User, error)
	SaveUser(ctx context.Context, user *User) error
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
	SavedEntities() []ddd.Entity
}

type StubRepository struct {
	user          *User
	entities      []ddd.Entity
	commitCalled  bool
	getUserById   func(ctx context.Context, id string) (*User, error)
	saveUser      func(ctx context.Context, user *User) error
	commit        func(ctx context.Context) error
	rollback      func(ctx context.Context) error
	savedEntities func() []ddd.Entity
}

func (r *StubRepository) GetUserById(ctx context.Context, id string) (*User, error) {
	return r.getUserById(ctx, id)
}

func (r *StubRepository) SaveUser(ctx context.Context, user *User) error {
	return r.saveUser(ctx, user)
}
func (r *StubRepository) Commit(ctx context.Context) error {
	return r.commit(ctx)
}
func (r *StubRepository) Rollback(ctx context.Context) error {
	return r.rollback(ctx)
}
func (r *StubRepository) SavedEntities() []ddd.Entity {
	return r.savedEntities()
}

type ChangeEmailCommandHandler struct {
	r             Repository
	savedEntities []ddd.Entity
}

func (h *ChangeEmailCommandHandler) Handle(ctx context.Context, command ddd.Command) (any, error) {
	if err := command.IsValid(); err != nil {
		return nil, err
	}
	changeEmailCommand, ok := command.(*ChangeEmailCommand)
	if ok == false {
		return nil, errors.New("ChangeEmailCommandHandler expects a command of type ChangeEmailCommand")
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

func (h *ChangeEmailCommandHandler) Commit(ctx context.Context) error {
	return h.r.Commit(ctx)
}

func (h *ChangeEmailCommandHandler) Rollback(ctx context.Context) error {
	return h.r.Rollback(ctx)
}

func (h *ChangeEmailCommandHandler) SavedEntities() []ddd.Entity {
	return h.savedEntities
}

type FakeBootstrapper struct {
	b                         *ddd.Bootstrapper
	ChangeEmailCommandHandler *ChangeEmailCommandHandler
	Repository                *StubRepository
}

func NewFakeBootstrapper() *FakeBootstrapper {
	repo := &StubRepository{}

	fb := &FakeBootstrapper{
		b:                         ddd.NewBootstrapper(),
		Repository:                repo,
		ChangeEmailCommandHandler: &ChangeEmailCommandHandler{r: repo},
	}
	fb.b.RegisterCommandHandlerFactory(&ChangeEmailCommand{}, func() (ddd.CommandHandler, error) {
		return fb.ChangeEmailCommandHandler, nil
	})
	return fb
}

func TestChangeEmail(t *testing.T) {
	fb := NewFakeBootstrapper()
	const userID = "1"
	const originalEmail = "kamel.amit@thaabet.sy"
	const newEmail = "eli.cohen@mossad.gov.il"
	fb.Repository.user = &User{email: originalEmail}
	fb.Repository.user.SetID(userID)

	fb.Repository.saveUser = func(ctx context.Context, user *User) error {
		fb.Repository.user = user
		fb.Repository.entities = append(fb.Repository.entities, user)
		return nil
	}
	fb.Repository.getUserById = func(ctx context.Context, id string) (*User, error) {
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

	command := &ChangeEmailCommand{Email: newEmail, UserID: userID}

	result, err := fb.b.HandleCommand(context.Background(), command)
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
		t.Errorf("want commitCalled true, got %v", fb.Repository.commitCalled)
	}
	savedEntities := fb.Repository.SavedEntities()
	if len(savedEntities) != 1 {
		t.Errorf("want 1 savedEntites, got %v", len(savedEntities))
	}
	savedEntity := savedEntities[0]
	if savedEntity.ID() != userID {
		t.Errorf("want savedEnttiy id %q, got %q", userID, savedEntity.ID())
	}
}
