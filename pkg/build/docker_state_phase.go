package build

import "fmt"

func NewDockerStatePhase() *DockerStatePhase {
	return &DockerStatePhase{}
}

type DockerStatePhase struct{}

func (p *DockerStatePhase) Run(c *Conveyor) error {
	if debug() {
		fmt.Printf("DockerStatePhase.Run\n")
	}
	return nil
}
