package container_backend

import "fmt"

type CommandEnv struct {
	Envs map[string]string
}

type CommandCopy struct {
	From string
	Src  []string
	Dst  string
}

type CommandAdd struct {
	Src []string
	Dst string
}

type CommandRun struct {
	Command []string
}

type CommandEntrypoint struct {
	Entrypoint []string
}

type CommandCmd struct {
	Cmd []string
}

type CommandUser struct {
	User string
}

type CommandWorkdir struct {
	Workdir string
}

type CommandExpose struct {
	Ports []string
}

type CommandVolume struct {
	Volumes []string
}

type CommandOnBuild struct {
	Instruction string
}

type CommandStopSignal struct {
	Signal string
}

type CommandShell struct {
	Shell []string
}

type CommandHealthcheck struct {
	Type    HealthcheckType
	Command string
}

type HealthcheckType string

var (
	HealthcheckTypeNone     HealthcheckType = "NONE"
	HealthcheckTypeCmd      HealthcheckType = "CMD"
	HealthcheckTypeCmdShell HealthcheckType = "CMD-SHELL"
)

type CommandLabel struct {
	Labels map[string]string
}

func (c *CommandLabel) LabelsAsList() []string {
	var labels []string
	for k, v := range c.Labels {
		labels = append(labels, fmt.Sprintf("%s=%s", k, v))
	}

	return labels
}
