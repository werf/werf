package cleaning

import (
	"context"

	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/style"
	"github.com/werf/logboek/pkg/types"

	"github.com/werf/werf/pkg/docker_registry"
	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/logging"
	"github.com/werf/werf/pkg/storage"
	"github.com/werf/werf/pkg/storage/manager"
)

type ImagesPurgeOptions struct {
	ImageNameList []string
	DryRun        bool
}

func ImagesPurge(ctx context.Context, projectName string, storageManager *manager.StorageManager, storageLockManager storage.LockManager, options ImagesPurgeOptions) error {
	m := newImagesPurgeManager(storageManager, options)

	return logboek.Context(ctx).Default().LogProcess("Running images purge").
		Options(func(options types.LogProcessOptionsInterface) {
			options.Style(style.Highlight())
		}).
		DoError(func() error {
			return m.run(ctx)
		})
}

func newImagesPurgeManager(storageManager *manager.StorageManager, options ImagesPurgeOptions) *imagesPurgeManager {
	return &imagesPurgeManager{
		StorageManager: storageManager,
		ImageNameList:  options.ImageNameList,
		DryRun:         options.DryRun,
	}
}

type imagesPurgeManager struct {
	StorageManager *manager.StorageManager
	ImageNameList  []string
	DryRun         bool
}

func (m *imagesPurgeManager) run(ctx context.Context) error {
	repoImages, err := selectRepoImagesFromImagesRepo(ctx, m.StorageManager, m.ImageNameList)
	if err != nil {
		return err
	}

	for imageName, repoImageList := range repoImages {
		if err := logboek.Context(ctx).Default().LogProcess(logging.ImageLogProcessName(imageName, false)).DoError(func() error {
			return deleteRepoImageInImagesRepo(ctx, m.StorageManager, m.DryRun, repoImageList...)
		}); err != nil {
			return err
		}
	}

	return nil
}

func selectRepoImagesFromImagesRepo(ctx context.Context, storageManager *manager.StorageManager, imageNameList []string) (map[string][]*image.Info, error) {
	return storageManager.SelectRepoImages(ctx, imageNameList, func(reference string, _ *image.Info, err error) (bool, error) {
		if err != nil && docker_registry.IsManifestUnknownError(err) {
			logboek.Context(ctx).Warn().LogF("Skip image %s: %s\n", reference, err)
			return false, nil
		}

		return true, err
	})
}

func deleteRepoImageInImagesRepo(ctx context.Context, storageManager *manager.StorageManager, dryRun bool, repoImageList ...*image.Info) error {
	if dryRun {
		for _, repoImage := range repoImageList {
			logboek.Context(ctx).Default().LogFDetails("  tag: %s\n", repoImage.Tag)
			logboek.Context(ctx).LogOptionalLn()
		}
		return nil
	}

	return storageManager.ForEachDeleteRepoImage(ctx, repoImageList, func(repoImage *image.Info, err error) error {
		if err != nil {
			if err := handleDeleteStageOrImageError(ctx, err, repoImage.Name); err != nil {
				return err
			}
			return nil
		}

		logboek.Context(ctx).Default().LogFDetails("  tag: %s\n", repoImage.Tag)
		logboek.Context(ctx).LogOptionalLn()

		return nil
	})
}
