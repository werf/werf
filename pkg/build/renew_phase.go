package build

import "fmt"

func NewRenewPhase() *RenewPhase {
	return &RenewPhase{}
}

type RenewPhase struct{}

func (p *RenewPhase) Run(c *Conveyor) error {
	if debug() {
		fmt.Printf("RenewPhase.Run\n")
	}
	return nil
}
