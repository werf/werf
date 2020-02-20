package cleaning

import (
	"github.com/flant/logboek"
)

type ImagesPurgeOptions struct {
	ImagesRepoManager ImagesRepoManager
	ImagesNames       []string
	DryRun            bool
}

func ImagesPurge(options ImagesPurgeOptions) error {
	return logboek.Default.LogProcess(
		"Running images purge",
		logboek.LevelLogProcessOptions{Style: logboek.HighlightStyle()},
		func() error {
			return imagesPurge(options)
		},
	)
}

func imagesPurge(options ImagesPurgeOptions) error {
	commonRepoOptions := CommonRepoOptions{
		ImagesRepoManager: options.ImagesRepoManager,
		ImagesNames:       options.ImagesNames,
		DryRun:            options.DryRun,
	}

	imageImages, err := repoImages(commonRepoOptions)
	if err != nil {
		return err
	}

	err = repoImagesRemove(imageImages, commonRepoOptions)
	if err != nil {
		return err
	}

	return nil
}
