package build

import "fmt"

func NewPrepareImagesPhase() *PrepareImagesPhase {
	return &PrepareImagesPhase{}
}

type PrepareImagesPhase struct{}

func (p *PrepareImagesPhase) Run(c *Conveyor) error {
	if debug() {
		fmt.Printf("PrepareImagesPhase.Run\n")
	}
	return nil
}
