package cleaning

import (
	"fmt"

	"github.com/werf/logboek"

	"github.com/werf/werf/pkg/docker_registry"
	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/storage"
)

type ImagesPurgeOptions struct {
	ImageNameList []string
	DryRun        bool
}

func ImagesPurge(projectName string, imagesRepo storage.ImagesRepo, storageLockManager storage.LockManager, options ImagesPurgeOptions) error {
	m := newImagesPurgeManager(imagesRepo, options)

	if lock, err := storageLockManager.LockStagesAndImages(projectName, storage.LockStagesAndImagesOptions{GetOrCreateImagesOnly: false}); err != nil {
		return fmt.Errorf("unable to lock stages and images: %s", err)
	} else {
		defer storageLockManager.Unlock(lock)
	}

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

func selectRepoImagesFromImagesRepo(imagesRepo storage.ImagesRepo, imageNameList []string) (map[string][]*image.Info, error) {
	return imagesRepo.SelectRepoImages(imageNameList, func(reference string, info *image.Info, err error) (bool, error) {
		if err != nil && docker_registry.IsManifestUnknownError(err) {
			logboek.Warn.LogF("Skip image %s: %s\n", reference, err)
			return false, nil
		}

		return true, err
	})
}

func deleteRepoImageInImagesRepo(imagesRepo storage.ImagesRepo, dryRun bool, repoImageList ...*image.Info) error {
	if err := deleteRepoImage(imagesRepo.DeleteRepoImage, storage.DeleteImageOptions{}, dryRun, repoImageList...); err != nil {
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
