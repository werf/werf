package build

import (
	"fmt"
	"os"

	"github.com/flant/werf/pkg/build/stage"
)

type ShouldBeBuiltPhase struct{}

func NewShouldBeBuiltPhase() *ShouldBeBuiltPhase {
	return &ShouldBeBuiltPhase{}
}

func (p *ShouldBeBuiltPhase) Run(c *Conveyor) error {
	if debug() {
		fmt.Printf("ShouldBeBuiltPhase.Run\n")
	}

	var badImages []*Image

	for _, image := range c.imagesInOrder {
		if debug() {
			fmt.Printf("  image: '%s'\n", image.GetName())
		}

		var badStages []stage.Interface

		for _, s := range image.GetStages() {
			image := s.GetImage()
			if image.IsExists() {
				continue
			}
			badStages = append(badStages, s)
		}

		for _, s := range badStages {
			if image.GetName() != "" {
				fmt.Fprintf(os.Stderr, "Image '%s' stage '%s' is not built\n", image.GetName(), s.Name())
			} else {
				fmt.Fprintf(os.Stderr, "Image stage '%s' is not built\n", s.Name())
			}
		}

		if len(badStages) > 0 {
			badImages = append(badImages, image)
		}
	}

	if len(badImages) > 0 {
		return fmt.Errorf("images should be built")
	}

	return nil
}
