package cleaning

import (
	"fmt"

	"github.com/flant/werf/pkg/image"
	"github.com/flant/werf/pkg/logger"
)

func StagesPurge(options CommonProjectOptions) error {
	return logger.LogProcess("Running stages purge", logger.LogProcessOptions{}, func() error {
		return stagesPurge(options)
	})
}

func stagesPurge(options CommonProjectOptions) error {
	options.CommonOptions.SkipUsedImages = false
	options.CommonOptions.RmiForce = true
	options.CommonOptions.RmForce = false

	if err := projectStagesPurge(options); err != nil {
		return err
	}

	return nil
}

func projectStagesPurge(options CommonProjectOptions) error {
	images, err := werfImagesByFilterSet(projectImageStageFilterSet(options))
	if err != nil {
		return err
	}

	if err := imagesRemove(images, options.CommonOptions); err != nil {
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
