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
				if err := handleDeleteStageOrImageError(err, repoImage.Name); err != nil {
					return err
				}
			}
		}

		logboek.Default.LogFDetails("  tag: %s\n", repoImage.Tag)
		logboek.LogOptionalLn()
	}

	return nil
}
