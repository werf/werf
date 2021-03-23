package cleaning

import (
	"context"

	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/style"
	"github.com/werf/logboek/pkg/types"

	"github.com/werf/werf/pkg/storage"
	"github.com/werf/werf/pkg/storage/manager"
)

type PurgeOptions struct {
	ImagesPurgeOptions
	StagesPurgeOptions
}

func Purge(ctx context.Context, projectName string, storageManager *manager.StorageManager, storageLockManager storage.LockManager, options PurgeOptions) error {
	m := newPurgeManager(projectName, storageManager, options)

	if err := logboek.Context(ctx).Default().LogProcess("Running images purge").DoError(func() error {
		return m.imagesPurgeManager.run(ctx)
	}); err != nil {
		return err
	}

	if err := logboek.Context(ctx).Default().LogProcess("Running stages purge").
		Options(func(options types.LogProcessOptionsInterface) {
			options.Style(style.Highlight())
		}).
		DoError(func() error {
			return m.stagesPurgeManager.run(ctx)
		}); err != nil {
		return err
	}

	return nil
}

func newPurgeManager(projectName string, storageManager *manager.StorageManager, options PurgeOptions) *purgeManager {
	return &purgeManager{
		imagesPurgeManager: newImagesPurgeManager(storageManager, options.ImagesPurgeOptions),
		stagesPurgeManager: newStagesPurgeManager(projectName, storageManager, options.StagesPurgeOptions),
	}
}

type purgeManager struct {
	*imagesPurgeManager
	*stagesPurgeManager
}
