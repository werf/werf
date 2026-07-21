package instruction

import (
	"github.com/moby/buildkit/frontend/dockerfile/instructions"
)

type Entrypoint struct {
	instructions.EntrypointCommand
	EntrypointResetCMD bool
}

func NewEntrypoint(i instructions.EntrypointCommand, entrypointResetCMD bool) *Entrypoint {
	return &Entrypoint{EntrypointCommand: i, EntrypointResetCMD: entrypointResetCMD}
}

func (i *Entrypoint) UsesBuildContext() bool {
	return false
}
