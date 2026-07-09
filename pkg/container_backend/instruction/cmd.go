package instruction

import (
	"github.com/moby/buildkit/frontend/dockerfile/instructions"
)

type Cmd struct {
	instructions.CmdCommand
}

func NewCmd(i instructions.CmdCommand) *Cmd {
	return &Cmd{CmdCommand: i}
}

func (i *Cmd) UsesBuildContext() bool {
	return false
}
