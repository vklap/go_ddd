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
		return ddd.NewError("user ID cannot be empty", ddd.StatusCodeBadRequest)
	}
	if c.NewEmail == "" {
		return ddd.NewError("new email cannot be empty", ddd.StatusCodeBadRequest)
	}
	return nil
}

func (c *ChangeEmailCommand) CommandName() string {
	return "ChangeEmailCommand"
}

// The below line ensures at compile time that ChangeEmailCommand adheres to the ddd.Command interface
var _ ddd.Command = (*ChangeEmailCommand)(nil)
