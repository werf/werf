package instruction

type Healthcheck struct {
	Type    HealthcheckType
	Command string
}

type HealthcheckType string

var (
	HealthcheckTypeNone     HealthcheckType = "NONE"
	HealthcheckTypeCmd      HealthcheckType = "CMD"
	HealthcheckTypeCmdShell HealthcheckType = "CMD-SHELL"
)
