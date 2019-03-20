package cleaning

import "github.com/flant/logboek"

func ImagesPurge(options CommonRepoOptions) error {
	return logboek.LogProcess("Running images purge", logboek.LogProcessOptions{}, func() error {
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
