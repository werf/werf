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
		var acquiredLocks []string

		unlockLocks := func() {
			locks := acquiredLocks
			for len(acquiredLocks) > 0 {
				var lockName string
				lockName, locks = locks[0], locks[1:]
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
		for _, stage := range dimg.GetStages() {
			img := stage.GetImage()
			if img.IsExists() {
				continue
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
