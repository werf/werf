package instruction

import "github.com/docker/docker/api/types/container"

type Healthcheck struct {
	Type   HealthcheckType
	Config *container.HealthConfig
}

type HealthcheckType string

var (
	HealthcheckTypeNone     HealthcheckType = "NONE"
	HealthcheckTypeCmd      HealthcheckType = "CMD"
	HealthcheckTypeCmdShell HealthcheckType = "CMD-SHELL"
)

func NewHealthcheck(t HealthcheckType, cfg *container.HealthConfig) *Healthcheck {
	return &Healthcheck{Type: t, Config: cfg}
}

func (i *Healthcheck) Name() string {
	return "HEALTHCHECK"
}
