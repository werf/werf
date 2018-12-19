package build

import (
	"fmt"
	"os"

	"github.com/flant/dapp/pkg/build/stage"
)

type ShouldBeBuiltPhase struct{}

func NewShouldBeBuiltPhase() *ShouldBeBuiltPhase {
	return &ShouldBeBuiltPhase{}
}

func (p *ShouldBeBuiltPhase) Run(c *Conveyor) error {
	if debug() {
		fmt.Printf("ShouldBeBuiltPhase.Run\n")
	}

	var badDimgs []*Dimg

	for _, dimg := range c.dimgsInOrder {
		if debug() {
			fmt.Printf("  dimg: '%s'\n", dimg.GetName())
		}

		var badStages []stage.Interface

		for _, s := range dimg.GetStages() {
			image := s.GetImage()
			if image.IsExists() {
				continue
			}
			badStages = append(badStages, s)
		}

		for _, s := range badStages {
			if dimg.GetName() != "" {
				fmt.Fprintf(os.Stderr, "Dimg '%s' stage '%s' is not built\n", dimg.GetName(), s.Name())
			} else {
				fmt.Fprintf(os.Stderr, "Dimg stage '%s' is not built\n", s.Name())
			}
		}

		if len(badStages) > 0 {
			badDimgs = append(badDimgs, dimg)
		}
	}

	if len(badDimgs) > 0 {
		return fmt.Errorf("dimgs should be built")
	}

	return nil
}
