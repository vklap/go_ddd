package command_model

import "github.com/vklap/go_ddd/pkg/go_ddd"

// ChangeEmailCommand contains the data required to change the user's email.
// Besides this, it also represents a main flow.
type ChangeEmailCommand struct {
	UserID   string `json:"user_id"`
	NewEmail string `json:"new_email"`
}

func (c *ChangeEmailCommand) IsValid() error {
	if c.UserID == "" {
		return go_ddd.NewError("userID cannot be empty", go_ddd.StatusCodeBadRequest)
	}
	if c.NewEmail == "" {
		return go_ddd.NewError("email cannot be empty", go_ddd.StatusCodeBadRequest)
	}
	return nil
}

func (c *ChangeEmailCommand) CommandName() string {
	return "ChangeEmailCommand"
}

// The below line ensures at compile time that ChangeEmailCommand adheres to the go_ddd.Command interface
var _ go_ddd.Command = (*ChangeEmailCommand)(nil)
