package instruction

import (
	"github.com/moby/buildkit/frontend/dockerfile/instructions"
)

type Shell struct {
	instructions.ShellCommand
}

func NewShell(i instructions.ShellCommand) *Shell {
	return &Shell{ShellCommand: i}
}

func (i *Shell) UsesBuildContext() bool {
	return false
}
