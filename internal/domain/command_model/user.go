package command_model

import "github.com/vklap/go_ddd/pkg/go_ddd"

// User is composed of go_ddd.BaseEntity which exposes the entity's ID and Events,
// and the user's Email.
type User struct {
	go_ddd.BaseEntity
	email string
}

func (u *User) Email() string {
	return u.email
}

func (u *User) SetEmail(value string) {
	if u.email != "" && u.email != value {
		// Record the EmailChangedEvent
		u.AddEvent(&EmailChangedEvent{UserID: u.ID(), NewEmail: value, OriginalEmail: u.email})
	}
	u.email = value
}

// The below line ensures at compile time that User adheres to the go_ddd.Entity interface
var _ go_ddd.Entity = (*User)(nil)
