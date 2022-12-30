package adapters

import (
	"context"
	"fmt"
	"github.com/vklap/go_ddd/internal/domain/command_model"
	"github.com/vklap/go_ddd/pkg/go_ddd"
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
	go_ddd.RollbackCommitter
}

// InMemoryRepository is used for demo purposes.
// In the real world it might be a MongoDBRepository, PostgresqlRepository, etc.
type InMemoryRepository struct{}

func (r *InMemoryRepository) GetUserById(ctx context.Context, id string) (*command_model.User, error) {
	user, ok := usersById[id]
	if ok == false {
		return nil, go_ddd.NewError(fmt.Sprintf("user with id %q does not exist", id), go_ddd.StatusCodeNotFound)
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
