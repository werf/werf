package build

import (
	"fmt"

	"github.com/flant/dapp/pkg/image"
	"github.com/flant/dapp/pkg/lock"
)

func NewBuildPhase() *BuildPhase {
	return &BuildPhase{}
}

type BuildPhase struct{}

func (p *BuildPhase) Run(c *Conveyor) error {
	if debug() {
		fmt.Printf("BuildPhase.Run\n")
	}

	for _, dimg := range c.DimgsInOrder {
		if debug() {
			fmt.Printf("  dimg: '%s'\n", dimg.GetName())
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
		for _, stage := range dimg.GetStages() {
			img := stage.GetImage()
			if img.IsExists() {
				continue
			}

			imageLockName := fmt.Sprintf("%s.image.%s", c.ProjectName, img.Name())
			err := lock.Lock(imageLockName, lock.LockOptions{})
			if err != nil {
				return fmt.Errorf("failed to lock %s: %s", imageLockName, err)
			}
		}

		// build
		for _, s := range dimg.GetStages() {
			img := s.GetImage()
			if img.IsExists() {
				continue
			}

			if debug() {
				fmt.Printf("    %s\n", s.Name())
			}

			err := img.Build2(image.BuildOptions{})
			if err != nil {
				return fmt.Errorf("failed to build %s: %s", img.Name(), err)
			}
		}

		// save in cache
		for _, stage := range dimg.GetStages() {
			img := stage.GetImage()
			if img.IsExists() {
				continue
			}

			err := img.SaveInCache()
			if err != nil {
				return fmt.Errorf("failed to save in cache image %s: %s", img.Name(), err)
			}
		}

		unlockLocks()
	}

	return nil
}
