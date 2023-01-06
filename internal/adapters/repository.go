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
