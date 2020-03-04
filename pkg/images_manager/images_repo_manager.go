package images_manager

type ImagesRepoManager interface {
	ImagesRepo() string
	ImageRepo(imageName string) string
	ImageRepoWithTag(imageName, tag string) string
}
