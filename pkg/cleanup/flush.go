package cleanup

import (
	"fmt"
	"time"

	"github.com/flant/werf/pkg/image"
	"github.com/flant/werf/pkg/lock"
)

func RepoImagesFlush(withImages bool, options CommonRepoOptions) error {
	err := lock.WithLock(options.Repository, lock.LockOptions{Timeout: time.Second * 600}, func() error {
		if withImages {
			if err := repoImagesFlush(options); err != nil {
				return err
			}
		}

		if err := repoImageStagesFlush(options); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func ProjectImagesFlush(withImages bool, options CommonProjectOptions) error {
	projectImagesLockName := fmt.Sprintf("%s.images", options.ProjectName)
	err := lock.WithLock(projectImagesLockName, lock.LockOptions{Timeout: time.Second * 600}, func() error {
		if withImages {
			if err := projectImagesFlush(options); err != nil {
				return err
			}
		}

		if err := projectCleanup(options); err != nil {
			return err
		}

		if err := projectImageStagesFlush(options); err != nil {
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

func repoImageStagesFlush(options CommonRepoOptions) error {
	imageStagesImages, err := repoImageStagesImages(options)
	if err != nil {
		return err
	}

	err = repoImagesRemove(imageStagesImages, options)
	if err != nil {
		return err
	}

	return nil
}

func projectImagesFlush(options CommonProjectOptions) error {
	filterSet := projectFilterSet(options)
	filterSet.Add("label", fmt.Sprintf("%s=true", image.WerfImageLabel))
	if err := werfImagesFlushByFilterSet(filterSet, options.CommonOptions); err != nil {
		return err
	}

	return nil
}

func projectImageStagesFlush(options CommonProjectOptions) error {
	if err := werfImagesFlushByFilterSet(projectImageStageFilterSet(options), options.CommonOptions); err != nil {
		return err
	}

	return nil
}
