package instruction

import (
	"github.com/moby/buildkit/frontend/dockerfile/instructions"
)

type Volume struct {
	instructions.VolumeCommand
}

func NewVolume(i instructions.VolumeCommand) *Volume {
	return &Volume{VolumeCommand: i}
}

func (i *Volume) UsesBuildContext() bool {
	return false
}
