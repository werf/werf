package instruction

import "github.com/docker/docker/api/types/container"

type Healthcheck struct {
	// TODO(ilya-lesikov): isn't this should be a part of a Config.Test?
	Type   HealthcheckType
	Config *container.HealthConfig
}

type HealthcheckType string

var (
	HealthcheckTypeInherit  HealthcheckType = ""
	HealthcheckTypeNone     HealthcheckType = "NONE"
	HealthcheckTypeCmd      HealthcheckType = "CMD"
	HealthcheckTypeCmdShell HealthcheckType = "CMD-SHELL"
)

func NewHealthcheckType(cfg *container.HealthConfig) HealthcheckType {
	if len(cfg.Test) == 0 {
		return HealthcheckTypeInherit
	} else {
		return HealthcheckType(cfg.Test[0])
	}
}

func NewHealthcheck(cfg *container.HealthConfig) *Healthcheck {
	return &Healthcheck{Type: NewHealthcheckType(cfg), Config: cfg}
}

func (i *Healthcheck) Name() string {
	return "HEALTHCHECK"
}
