package build

import "fmt"

func NewBuildPhase() *BuildPhase {
	return &BuildPhase{}
}

type BuildPhase struct{}

func (p *BuildPhase) Run(c *Conveyor) error {
	if debug() {
		fmt.Printf("BuildPhase.Run\n")
	}
	return nil
}
