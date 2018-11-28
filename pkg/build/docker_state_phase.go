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

	for _, dimg := range c.GetDimgsInOrder() {
		for _, stage := range dimg.GetStages() {
			err := stage.GetImage().ReadDockerState()
			if err != nil {
				return fmt.Errorf("error reading docker state of stage %s: %s", stage.Name(), err)
			}
		}
	}

	return nil
}
