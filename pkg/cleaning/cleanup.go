package cleaning

import (
	"context"

	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/style"
	"github.com/werf/logboek/pkg/types"

	"github.com/werf/werf/pkg/storage"
	"github.com/werf/werf/pkg/storage/manager"
)

type CleanupOptions struct {
	ImagesCleanupOptions
	StagesCleanupOptions
}

func Cleanup(ctx context.Context, projectName string, storageManager *manager.StorageManager, storageLockManager storage.LockManager, options CleanupOptions) error {
	m := newCleanupManager(projectName, storageManager, options)

	if err := logboek.Context(ctx).Default().LogProcess("Running images cleanup").
		Options(func(options types.LogProcessOptionsInterface) {
			options.Style(style.Highlight())
		}).
		DoError(func() error {
			return m.imagesCleanupManager.run(ctx)
		}); err != nil {
		return err
	}

	repoImages := m.imagesCleanupManager.getImageRepoImageList()
	m.stagesCleanupManager.setImagesRepoImageList(flattenRepoImages(repoImages))

	if err := logboek.Context(ctx).Default().LogProcess("Running stages cleanup").
		Options(func(options types.LogProcessOptionsInterface) {
			options.Style(style.Highlight())
		}).
		DoError(func() error {
			return m.stagesCleanupManager.run(ctx)
		}); err != nil {
		return err
	}

	return nil
}

func newCleanupManager(projectName string, storageManager *manager.StorageManager, options CleanupOptions) *cleanupManager {
	return &cleanupManager{
		imagesCleanupManager: newImagesCleanupManager(projectName, storageManager, options.ImagesCleanupOptions),
		stagesCleanupManager: newStagesCleanupManager(projectName, storageManager, options.StagesCleanupOptions),
	}
}

type cleanupManager struct {
	*imagesCleanupManager
	*stagesCleanupManager
}
