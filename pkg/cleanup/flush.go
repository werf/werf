package cleanup

import (
	"fmt"
	"time"

	"github.com/flant/werf/pkg/image"
	"github.com/flant/werf/pkg/lock"
)

func ImagesFlush(options CommonRepoOptions) error {
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

func StagesFlush(options CommonProjectOptions) error {
	projectImagesLockName := fmt.Sprintf("%s.images", options.ProjectName)
	err := lock.WithLock(projectImagesLockName, lock.LockOptions{Timeout: time.Second * 600}, func() error {
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
