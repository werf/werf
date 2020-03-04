package cleaning

import (
	"fmt"

	"github.com/docker/docker/api/types"

	"github.com/docker/docker/api/types/filters"
	"github.com/flant/logboek"
	"github.com/flant/werf/pkg/docker"
	"github.com/flant/werf/pkg/image"
)

type StagesPurgeOptions struct {
	ProjectName                   string
	DryRun                        bool
	RmContainersThatUseWerfImages bool
}

func StagesPurge(options StagesPurgeOptions) error {
	return logboek.Default.LogProcess(
		"Running stages purge",
		logboek.LevelLogProcessOptions{Style: logboek.HighlightStyle()},
		func() error {
			return stagesPurge(options)
		},
	)
}

func stagesPurge(options StagesPurgeOptions) error {
	var commonProjectOptions CommonProjectOptions
	commonProjectOptions.ProjectName = options.ProjectName
	commonProjectOptions.CommonOptions = CommonOptions{
		RmiForce:                      true,
		RmForce:                       options.RmContainersThatUseWerfImages,
		RmContainersThatUseWerfImages: options.RmContainersThatUseWerfImages,
		DryRun:                        options.DryRun,
	}

	if err := projectStagesPurge(commonProjectOptions); err != nil {
		return err
	}

	return nil
}

func projectStagesPurge(options CommonProjectOptions) error {
	if err := werfImagesFlushByFilterSet(projectImageStageFilterSet(options), options.CommonOptions); err != nil {
		return err
	}

	if err := purgeManagedImages(options); err != nil {
		return fmt.Errorf("unable to purge managed images: %s", err)
	}

	return nil
}

func purgeManagedImages(options CommonProjectOptions) error {
	filterSet := filters.NewArgs()
	filterSet.Add("reference", fmt.Sprintf(image.ManagedImageRecord_ImageNameFormat, options.ProjectName))

	images, err := docker.Images(types.ImageListOptions{Filters: filterSet})
	if err != nil {
		return err
	}

	images, err = processUsedImages(images, options.CommonOptions)
	if err != nil {
		return err
	}

	if err := imagesRemove(images, options.CommonOptions); err != nil {
		return err
	}

	return nil
}
