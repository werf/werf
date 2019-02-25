package cleaning

func ImagesPurge(options CommonRepoOptions) error {
	if err := repoImagesFlush(options); err != nil {
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
