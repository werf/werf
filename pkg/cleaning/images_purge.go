package cleaning

import (
	"fmt"

	"github.com/flant/logboek"

	"github.com/flant/werf/pkg/docker_registry"
	"github.com/flant/werf/pkg/image"
	"github.com/flant/werf/pkg/storage"
)

type ImagesPurgeOptions struct {
	ImageNameList []string
	DryRun        bool
}

func ImagesPurge(imagesRepo storage.ImagesRepo, options ImagesPurgeOptions) error {
	m := newImagesPurgeManager(imagesRepo, options)

	return logboek.Default.LogProcess(
		"Running images purge",
		logboek.LevelLogProcessOptions{Style: logboek.HighlightStyle()},
		m.run,
	)
}

func newImagesPurgeManager(imagesRepo storage.ImagesRepo, options ImagesPurgeOptions) *imagesPurgeManager {
	return &imagesPurgeManager{
		ImagesRepo:    imagesRepo,
		ImageNameList: options.ImageNameList,
		DryRun:        options.DryRun,
	}
}

type imagesPurgeManager struct {
	ImagesRepo    storage.ImagesRepo
	ImageNameList []string
	DryRun        bool
}

func (m *imagesPurgeManager) run() error {
	repoImageList, err := imagesRepoImageList(m.ImagesRepo, m.ImageNameList)
	if err != nil {
		return err
	}

	return deleteRepoImageInImagesRepo(m.ImagesRepo, m.DryRun, repoImageList...)
}

func deleteRepoImageInImagesRepo(imagesRepo storage.ImagesRepo, dryRun bool, repoImageList ...*image.Info) error {
	if err := deleteRepoImage(imagesRepo.DeleteRepoImage, storage.DeleteRepoImageOptions{}, dryRun, repoImageList...); err != nil {
		switch err.(type) {
		case docker_registry.DockerHubUnauthorizedError:
			return fmt.Errorf(`%s
You should specify Docker Hub token or username and password to remove tags with Docker Hub API.
Check --repo-docker-hub-token/username/password --images-repo-docker-hub-token/username/password options.
Be aware that access to the resource is forbidden with personal access token.
Read more details here https://werf.io/documentation/reference/working_with_docker_registries.html#docker-hub`, err)
		case docker_registry.GitHubPackagesUnauthorizedError:
			return fmt.Errorf(`%s
You should specify a token with the read:packages, write:packages, delete:packages and repo scopes to remove package versions.
Check --repo-github-token and --images-repo-github-token options.
Read more details here https://werf.io/documentation/reference/working_with_docker_registries.html#github-packages`, err)
		default:
			return err
		}
	}

	return nil
}
