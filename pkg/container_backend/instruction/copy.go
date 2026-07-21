package instruction

import (
	"github.com/moby/buildkit/frontend/dockerfile/instructions"
)

type Copy struct {
	instructions.CopyCommand
}

func NewCopy(i instructions.CopyCommand) *Copy {
	return &Copy{CopyCommand: i}
}

func (i *Copy) UsesBuildContext() bool {
	return i.From == ""
}
