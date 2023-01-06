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
