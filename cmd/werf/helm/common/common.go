package common

func GetImagesRepoOrStub(imagesRepoOption string) string {
	if imagesRepoOption == "" {
		return "IMAGES_REPO"
	}
	return imagesRepoOption
}

func GetEnvironmentOrStub(environmentOption string) string {
	if environmentOption == "" {
		return "ENV"
	}
	return environmentOption
}
