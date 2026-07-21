package instruction

import (
	"github.com/moby/buildkit/frontend/dockerfile/instructions"
)

type OnBuild struct {
	instructions.OnbuildCommand
}

func NewOnBuild(i instructions.OnbuildCommand) *OnBuild {
	return &OnBuild{OnbuildCommand: i}
}

func (i *OnBuild) UsesBuildContext() bool {
	return false
}
