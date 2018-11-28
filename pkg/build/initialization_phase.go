package build

import (
	"github.com/flant/dapp/pkg/build/stage"
)

type InitializationPhase struct{}

func NewInitializationPhase() *InitializationPhase {
	return &InitializationPhase{}
}

func (p *InitializationPhase) Run(c *Conveyor) error {
	c.DimgsInOrder = stage.GenerateDimgsInOrder(c.Dappfile)
	return nil
}
