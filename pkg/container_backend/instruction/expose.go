package instruction

import (
	"github.com/moby/buildkit/frontend/dockerfile/instructions"
)

type Expose struct {
	instructions.ExposeCommand
}

func NewExpose(i instructions.ExposeCommand) *Expose {
	return &Expose{ExposeCommand: i}
}

func (i *Expose) UsesBuildContext() bool {
	return false
}
