package cleaning

import (
	"fmt"

	"github.com/werf/logboek"

	"github.com/werf/werf/pkg/docker_registry"
	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/logging"
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
	repoImages, err := selectRepoImagesFromImagesRepo(m.ImagesRepo, m.ImageNameList)
	if err != nil {
		return err
	}

	for imageName, repoImageList := range repoImages {
		if err := logboek.Default.LogProcess(logging.ImageLogProcessName(imageName, false), logboek.LevelLogProcessOptions{}, func() error {
			return deleteRepoImageInImagesRepo(m.ImagesRepo, m.DryRun, repoImageList...)
		}); err != nil {
			return err
		}
	}

	return nil
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
	for _, repoImage := range repoImageList {
		if !dryRun {
			if err := imagesRepo.DeleteRepoImage(storage.DeleteImageOptions{}, repoImage); err != nil {
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
					logboek.Warn.LogF("WARNING: Image %s deletion failed: %s\n", repoImage.Name, err)
					logboek.LogOptionalLn()
					return nil
				}
			}
		}

		logboek.Default.LogFDetails("  tag: %s\n", repoImage.Tag)
		logboek.LogOptionalLn()
	}

	return nil
}
