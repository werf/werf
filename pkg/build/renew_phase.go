package build

import (
	"errors"
	"fmt"

	"github.com/flant/logboek"
	imagePkg "github.com/flant/werf/pkg/image"
)

var (
	ErrConveyorShouldBeReset = errors.New("conveyor should be reset")
)

func NewRenewPhase() *RenewPhase {
	return &RenewPhase{}
}

type RenewPhase struct{}

func (p *RenewPhase) Run(c *Conveyor) error {
	var resErr error

	logProcessOptions := logboek.LogProcessOptions{ColorizeMsgFunc: logboek.ColorizeHighlight}
	err := logboek.LogProcess("Checking invalid stages cache", logProcessOptions, func() error {
		err := p.run(c)

		if isConveyorShouldBeResetError(err) {
			resErr = err
			return nil
		}

		return err
	})

	if err != nil {
		return err
	}

	return resErr
}

func (p *RenewPhase) run(c *Conveyor) error {
	var conveyorShouldBeReset bool

	for _, image := range c.imagesInOrder {
		shouldResetAllNextStages := false
		for _, s := range image.GetStages() {
			img := s.GetImage()
			if img.IsExists() {
				if stageShouldBeReset, err := s.ShouldBeReset(img); err != nil {
					return err
				} else if stageShouldBeReset || shouldResetAllNextStages {
					conveyorShouldBeReset = true

					logboek.LogF("Untag %s for %s/%s\n", img.Name(), image.LogName(), s.Name())

					if err := img.Untag(); err != nil {
						return err
					}
				}

				imageLockName := imagePkg.ImageLockName(img.Name())
				if err := c.ReleaseGlobalLock(imageLockName); err != nil {
					return fmt.Errorf("failed to unlock %s: %s", imageLockName, err)
				}
			} else {
				shouldResetAllNextStages = true
			}
		}
	}

	if conveyorShouldBeReset {
		return ErrConveyorShouldBeReset
	} else {
		return nil
	}
}

func isConveyorShouldBeResetError(err error) bool {
	return err == ErrConveyorShouldBeReset
}
