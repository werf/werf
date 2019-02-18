package cleaning

import (
	"time"

	"github.com/flant/werf/pkg/lock"
)

func ImagesPurge(options CommonRepoOptions) error {
	err := lock.WithLock(options.ImagesRepo, lock.LockOptions{Timeout: time.Second * 600}, func() error {
		if err := repoImagesFlush(options); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func repoImagesFlush(options CommonRepoOptions) error {
	imageImages, err := repoImages(options)
	if err != nil {
		return err
	}

	err = repoImagesRemove(imageImages, options)
	if err != nil {
		return err
	}

	return nil
}
