package cleanup

import (
	"fmt"
	"time"

	"github.com/flant/werf/pkg/lock"
)

func RepoImagesFlush(withDimgs bool, options CommonRepoOptions) error {
	err := lock.WithLock(options.Repository, lock.LockOptions{Timeout: time.Second * 600}, func() error {
		if withDimgs {
			if err := repoDimgsFlush(options); err != nil {
				return err
			}
		}

		if err := repoDimgstagesFlush(options); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func ProjectImagesFlush(withDimgs bool, options CommonProjectOptions) error {
	projectImagesLockName := fmt.Sprintf("%s.images", options.ProjectName)
	err := lock.WithLock(projectImagesLockName, lock.LockOptions{Timeout: time.Second * 600}, func() error {
		if withDimgs {
			if err := projectDimgsFlush(options); err != nil {
				return err
			}
		}

		if err := projectCleanup(options); err != nil {
			return err
		}

		if err := projectDimgstagesFlush(options); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func repoDimgsFlush(options CommonRepoOptions) error {
	dimgImages, err := repoDimgImages(options)
	if err != nil {
		return err
	}

	err = repoImagesRemove(dimgImages, options)
	if err != nil {
		return err
	}

	return nil
}

func repoDimgstagesFlush(options CommonRepoOptions) error {
	dimgstageImages, err := repoDimgstageImages(options)
	if err != nil {
		return err
	}

	err = repoImagesRemove(dimgstageImages, options)
	if err != nil {
		return err
	}

	return nil
}

func projectDimgsFlush(options CommonProjectOptions) error {
	filterSet := projectFilterSet(options)
	filterSet.Add("label", "werf-dimg=true")
	if err := werfImagesFlushByFilterSet(filterSet, options.CommonOptions); err != nil {
		return err
	}

	return nil
}

func projectDimgstagesFlush(options CommonProjectOptions) error {
	if err := werfImagesFlushByFilterSet(projectDimgstageFilterSet(options), options.CommonOptions); err != nil {
		return err
	}

	return nil
}
