package instruction

import (
	"github.com/moby/buildkit/frontend/dockerfile/instructions"
)

type User struct {
	instructions.UserCommand
}

func NewUser(i instructions.UserCommand) *User {
	return &User{UserCommand: i}
}

func (i *User) UsesBuildContext() bool {
	return false
}
