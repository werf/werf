package build

import (
	"fmt"

	"github.com/flant/logboek"
	"github.com/flant/werf/pkg/build/stage"
)

type ShouldBeBuiltPhase struct{}

func NewShouldBeBuiltPhase() *ShouldBeBuiltPhase {
	return &ShouldBeBuiltPhase{}
}

func (p *ShouldBeBuiltPhase) Run(c *Conveyor) error {
	logProcessOptions := logboek.LogProcessOptions{ColorizeMsgFunc: logboek.ColorizeHighlight}
	return logboek.LogProcess("Checking built stages cache", logProcessOptions, func() error {
		return p.run(c)
	})
}

func (p *ShouldBeBuiltPhase) run(c *Conveyor) error {
	var badImages []*Image

	for _, image := range c.imagesInOrder {
		var badStages []stage.Interface

		for _, s := range image.GetStages() {
			image := s.GetImage()
			if image.IsExists() {
				continue
			}
			badStages = append(badStages, s)
		}

		for _, s := range badStages {
			logboek.LogErrorF("%s %s is not exist in stages storage\n", image.LogDetailedName(), s.LogDetailedName())
		}

		if len(badStages) > 0 {
			badImages = append(badImages, image)
		}
	}

	if len(badImages) > 0 {
		return fmt.Errorf("stages required")
	}

	return nil
}
