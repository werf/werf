package build

import (
	"errors"
	"fmt"

	"github.com/flant/werf/pkg/lock"
)

func NewRenewPhase() *RenewPhase {
	return &RenewPhase{}
}

type RenewPhase struct{}

func (p *RenewPhase) Run(c *Conveyor) error {
	if debug() {
		fmt.Printf("RenewPhase.Run\n")
	}

	var conveyorShouldBeReset bool
	for _, image := range c.imagesInOrder {
		if debug() {
			fmt.Printf("  image: '%s'\n", image.GetName())
		}

		var acquiredLocks []string

		unlockLocks := func() {
			for len(acquiredLocks) > 0 {
				var lockName string
				lockName, acquiredLocks = acquiredLocks[0], acquiredLocks[1:]
				lock.Unlock(lockName)
			}
		}

		defer unlockLocks()

		// lock
		for _, stage := range image.GetStages() {
			img := stage.GetImage()
			if !img.IsExists() {
				continue
			}

			imageLockName := fmt.Sprintf("%s.image.%s", c.projectName(), img.Name())
			err := lock.Lock(imageLockName, lock.LockOptions{})
			if err != nil {
				return fmt.Errorf("failed to lock %s: %s", imageLockName, err)
			}

			if err := img.SyncDockerState(); err != nil {
				return err
			}
		}

		// build
		for _, s := range image.GetStages() {
			img := s.GetImage()
			if img.IsExists() {
				if stageShouldBeReset, err := s.ShouldBeReset(img); err != nil {
					return err
				} else if stageShouldBeReset {
					conveyorShouldBeReset = true

					if image.GetName() == "" {
						fmt.Printf("# Reseting image %s for image %s\n", img.Name(), fmt.Sprintf("stage/%s", s.Name()))
					} else {
						fmt.Printf("# Reseting image %s for image/%s %s\n", img.Name(), image.GetName(), fmt.Sprintf("stage/%s", s.Name()))
					}

					if err := img.Untag(); err != nil {
						return err
					}
				}
			}
		}

		unlockLocks()
	}

	if conveyorShouldBeReset {
		return ConveyorShouldBeResetError()
	} else {
		return nil
	}
}

func ConveyorShouldBeResetError() error {
	return errors.New("conveyor should be reset")
}

func isConveyorShouldBeResetError(err error) bool {
	return err.Error() == ConveyorShouldBeResetError().Error()
}
