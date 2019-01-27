package build

import (
	"fmt"

	"github.com/flant/werf/pkg/build/stage"
	"github.com/flant/werf/pkg/logger"
)

type ShouldBeBuiltPhase struct{}

func NewShouldBeBuiltPhase() *ShouldBeBuiltPhase {
	return &ShouldBeBuiltPhase{}
}

func (p *ShouldBeBuiltPhase) Run(c *Conveyor) error {
	return logger.LogServiceProcess("Check built stages cache", "", func() error {
		return p.run(c)
	})
}

func (p *ShouldBeBuiltPhase) run(c *Conveyor) error {
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
			logger.LogWarningF("%s %s cache should be built\n", image.LogName(), s.Name())
		}

		if len(badStages) > 0 {
			badImages = append(badImages, image)
		}
	}

	if len(badImages) > 0 {
		return fmt.Errorf("images stages cache should be built")
	}

	return nil
}
