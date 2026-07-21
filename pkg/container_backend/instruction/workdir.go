package instruction

import (
	"github.com/moby/buildkit/frontend/dockerfile/instructions"
)

type Workdir struct {
	instructions.WorkdirCommand
}

func NewWorkdir(i instructions.WorkdirCommand) *Workdir {
	return &Workdir{WorkdirCommand: i}
}

func (i *Workdir) UsesBuildContext() bool {
	return false
}
