package dockerfile

import "fmt"

type InstructionEnv struct {
	Envs map[string]string
}

type InstructionCopy struct {
	From string
	Src  []string
	Dst  string
}

type InstructionAdd struct {
	Src []string
	Dst string
}

type InstructionRun struct {
	Command []string
}

type InstructionEntrypoint struct {
	Entrypoint []string
}

type InstructionCmd struct {
	Cmd []string
}

type InstructionUser struct {
	User string
}

type InstructionWorkdir struct {
	Workdir string
}

type InstructionExpose struct {
	Ports []string
}

type InstructionVolume struct {
	Volumes []string
}

type InstructionOnBuild struct {
	Instruction string
}

type InstructionStopSignal struct {
	Signal string
}

type InstructionShell struct {
	Shell []string
}

type InstructionHealthcheck struct {
	Type    HealthcheckType
	Command string
}

type HealthcheckType string

var (
	HealthcheckTypeNone     HealthcheckType = "NONE"
	HealthcheckTypeCmd      HealthcheckType = "CMD"
	HealthcheckTypeCmdShell HealthcheckType = "CMD-SHELL"
)

type InstructionLabel struct {
	Labels map[string]string
}

func (c *InstructionLabel) LabelsAsList() []string {
	var labels []string
	for k, v := range c.Labels {
		labels = append(labels, fmt.Sprintf("%s=%s", k, v))
	}

	return labels
}
