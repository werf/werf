package instruction

import (
	"github.com/moby/buildkit/frontend/dockerfile/instructions"
)

type StopSignal struct {
	instructions.StopSignalCommand
}

func NewStopSignal(i instructions.StopSignalCommand) *StopSignal {
	return &StopSignal{StopSignalCommand: i}
}

func (i *StopSignal) UsesBuildContext() bool {
	return false
}
