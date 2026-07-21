package instruction

import (
	"github.com/moby/buildkit/frontend/dockerfile/instructions"
)

type Env struct {
	instructions.EnvCommand
}

func NewEnv(i instructions.EnvCommand) *Env {
	return &Env{EnvCommand: i}
}

func (i *Env) UsesBuildContext() bool {
	return false
}
