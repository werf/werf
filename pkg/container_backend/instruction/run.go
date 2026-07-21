package instruction

import (
	"github.com/moby/buildkit/frontend/dockerfile/instructions"
)

type Run struct {
	instructions.RunCommand
	Envs    []string
	Secrets []string
	SSH     string
}

func NewRun(i instructions.RunCommand, envs, secrets []string, ssh string) *Run {
	return &Run{RunCommand: i, Envs: envs, Secrets: secrets, SSH: ssh}
}

func (i *Run) UsesBuildContext() bool {
	for _, mount := range i.GetMounts() {
		if mount.Type == instructions.MountTypeBind && mount.From == "" {
			return true
		}
	}

	return false
}

func (i *Run) GetMounts() []*instructions.Mount {
	return instructions.GetMounts(&i.RunCommand)
}

func (i *Run) GetSecurity() string {
	return instructions.GetSecurity(&i.RunCommand)
}

func (i *Run) GetNetwork() string {
	return instructions.GetNetwork(&i.RunCommand)
}
