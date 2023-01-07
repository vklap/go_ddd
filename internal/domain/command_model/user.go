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
