package cleaning

import "github.com/flant/werf/pkg/logger"

func ImagesPurge(options CommonRepoOptions) error {
	return logger.LogProcess("Running images purge", logger.LogProcessOptions{}, func() error {
		return imagesPurge(options)
	})
}

func imagesPurge(options CommonRepoOptions) error {
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
