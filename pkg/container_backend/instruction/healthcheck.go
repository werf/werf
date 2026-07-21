package instruction

import (
	"github.com/moby/buildkit/frontend/dockerfile/instructions"
)

type Healthcheck struct {
	instructions.HealthCheckCommand
}

func NewHealthcheck(i instructions.HealthCheckCommand) *Healthcheck {
	return &Healthcheck{HealthCheckCommand: i}
}

func (i *Healthcheck) UsesBuildContext() bool {
	return false
}
