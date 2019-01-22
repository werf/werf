package build

import (
	"fmt"

	imagePkg "github.com/flant/werf/pkg/image"
	"github.com/flant/werf/pkg/lock"
)

func NewBuildPhase(opts BuildOptions) *BuildPhase {
	return &BuildPhase{opts}
}

type BuildOptions struct {
	ImageBuildOptions imagePkg.BuildOptions
}

type BuildPhase struct {
	BuildOptions
}

func (p *BuildPhase) Run(c *Conveyor) error {
	if debug() {
		fmt.Printf("BuildPhase.Run\n")
	}

	for _, image := range c.imagesInOrder {
		if debug() {
			fmt.Printf("  image: '%s'\n", image.GetName())
		}

		var acquiredLocks []string

		unlockLock := func() {
			var lockName string
			lockName, acquiredLocks = acquiredLocks[0], acquiredLocks[1:]
			lock.Unlock(lockName)
		}

		unlockLocks := func() {
			for len(acquiredLocks) > 0 {
				unlockLock()
			}
		}

		defer unlockLocks()

		// lock
		for _, stage := range image.GetStages() {
			img := stage.GetImage()
			if img.IsExists() {
				continue
			}

			imageLockName := fmt.Sprintf("%s.image.%s", c.projectName(), img.Name())
			err := lock.Lock(imageLockName, lock.LockOptions{})
			if err != nil {
				return fmt.Errorf("failed to lock %s: %s", imageLockName, err)
			}

			acquiredLocks = append(acquiredLocks, imageLockName)

			if err := img.SyncDockerState(); err != nil {
				return err
			}
		}

		// build
		for _, s := range image.GetStages() {
			img := s.GetImage()
			if img.IsExists() {
				if image.GetName() == "" {
					fmt.Printf("# Using cached image %s for image %s\n", img.Name(), fmt.Sprintf("stage/%s", s.Name()))
				} else {
					fmt.Printf("# Using cached image %s for image/%s %s\n", img.Name(), image.GetName(), fmt.Sprintf("stage/%s", s.Name()))
				}

				continue
			}

			if image.GetName() == "" {
				fmt.Printf("# Building image %s for image %s\n", img.Name(), fmt.Sprintf("stage/%s", s.Name()))
			} else {
				fmt.Printf("# Building image %s for image/%s %s\n", img.Name(), image.GetName(), fmt.Sprintf("stage/%s", s.Name()))
			}

			if debug() {
				fmt.Printf("    %s\n", s.Name())
			}

			if err := s.PreRunHook(c); err != nil {
				return fmt.Errorf("stage '%s' preRunHook failed: %s", s.Name(), err)
			}

			if err := img.Build(p.ImageBuildOptions); err != nil {
				return fmt.Errorf("failed to build %s: %s", img.Name(), err)
			}

			err := img.SaveInCache()
			if err != nil {
				return fmt.Errorf("failed to save in cache image %s: %s", img.Name(), err)
			}

			unlockLock()
		}
	}

	return nil
}
