package cleaning

import (
	"context"
	"fmt"

	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/style"
	"github.com/werf/logboek/pkg/types"

	"github.com/werf/werf/pkg/docker_registry"
	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/logging"
	"github.com/werf/werf/pkg/storage"
)

type ImagesPurgeOptions struct {
	ImageNameList []string
	DryRun        bool
}

func ImagesPurge(ctx context.Context, projectName string, imagesRepo storage.ImagesRepo, storageLockManager storage.LockManager, options ImagesPurgeOptions) error {
	m := newImagesPurgeManager(imagesRepo, options)

	if lock, err := storageLockManager.LockStagesAndImages(ctx, projectName, storage.LockStagesAndImagesOptions{GetOrCreateImagesOnly: false}); err != nil {
		return fmt.Errorf("unable to lock stages and images: %s", err)
	} else {
		defer storageLockManager.Unlock(ctx, lock)
	}

	return logboek.Context(ctx).Default().LogProcess("Running images purge").
		Options(func(options types.LogProcessOptionsInterface) {
			options.Style(style.Highlight())
		}).
		DoError(func() error {
			return m.run(ctx)
		})
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

func (m *imagesPurgeManager) run(ctx context.Context) error {
	repoImages, err := selectRepoImagesFromImagesRepo(ctx, m.ImagesRepo, m.ImageNameList)
	if err != nil {
		return err
	}

	for imageName, repoImageList := range repoImages {
		if err := logboek.Context(ctx).Default().LogProcess(logging.ImageLogProcessName(imageName, false)).DoError(func() error {
			return deleteRepoImageInImagesRepo(ctx, m.ImagesRepo, m.DryRun, repoImageList...)
		}); err != nil {
			return err
		}
	}

	return nil
}

func selectRepoImagesFromImagesRepo(ctx context.Context, imagesRepo storage.ImagesRepo, imageNameList []string) (map[string][]*image.Info, error) {
	return imagesRepo.SelectRepoImages(ctx, imageNameList, func(reference string, info *image.Info, err error) (bool, error) {
		if err != nil && docker_registry.IsManifestUnknownError(err) {
			logboek.Context(ctx).Warn().LogF("Skip image %s: %s\n", reference, err)
			return false, nil
		}

		return true, err
	})
}

func deleteRepoImageInImagesRepo(ctx context.Context, imagesRepo storage.ImagesRepo, dryRun bool, repoImageList ...*image.Info) error {
	for _, repoImage := range repoImageList {
		if !dryRun {
			if err := imagesRepo.DeleteRepoImage(ctx, repoImage); err != nil {
				if err := handleDeleteStageOrImageError(ctx, err, repoImage.Name); err != nil {
					return err
				}

				continue
			}
		}

		logboek.Context(ctx).Default().LogFDetails("  tag: %s\n", repoImage.Tag)
		logboek.Context(ctx).LogOptionalLn()
	}

	return nil
}
