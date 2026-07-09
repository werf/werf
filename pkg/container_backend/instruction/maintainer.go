package instruction

import (
	"github.com/moby/buildkit/frontend/dockerfile/instructions"
)

type Maintainer struct {
	instructions.MaintainerCommand
}

func NewMaintainer(i instructions.MaintainerCommand) *Maintainer {
	return &Maintainer{MaintainerCommand: i}
}

func (i *Maintainer) UsesBuildContext() bool {
	return false
}
