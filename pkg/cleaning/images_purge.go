package cleaning

import "github.com/flant/logboek"

type ImagesPurgeOptions struct {
	ImagesRepoManager ImagesRepoManager
	ImagesNames       []string
	DryRun            bool
}

func ImagesPurge(options ImagesPurgeOptions) error {
	logProcessOptions := logboek.LogProcessOptions{ColorizeMsgFunc: logboek.ColorizeHighlight}
	return logboek.LogProcess("Running images purge", logProcessOptions, func() error {
		return imagesPurge(options)
	})
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
