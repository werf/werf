package instruction

import (
	"github.com/moby/buildkit/frontend/dockerfile/instructions"
)

type Add struct {
	instructions.AddCommand
}

func NewAdd(i instructions.AddCommand) *Add {
	return &Add{AddCommand: i}
}

func (i *Add) UsesBuildContext() bool {
	return true
}
