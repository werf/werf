package cleaning

import (
	"context"

	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/style"
	"github.com/werf/logboek/pkg/types"

	"github.com/werf/werf/pkg/storage"
	"github.com/werf/werf/pkg/storage/manager"
)

type StagesPurgeOptions struct {
	RmContainersThatUseWerfImages bool
	DryRun                        bool
}

func StagesPurge(ctx context.Context, projectName string, storageLockManager storage.LockManager, storageManager *manager.StorageManager, options StagesPurgeOptions) error {
	m := newStagesPurgeManager(projectName, storageManager, options)

	return logboek.Context(ctx).Default().LogProcess("Running stages purge").
		Options(func(options types.LogProcessOptionsInterface) {
			options.Style(style.Highlight())
		}).
		DoError(func() error {
			return m.run(ctx)
		})
}

func newStagesPurgeManager(projectName string, storageManager *manager.StorageManager, options StagesPurgeOptions) *stagesPurgeManager {
	return &stagesPurgeManager{
		StorageManager:                storageManager,
		ProjectName:                   projectName,
		RmContainersThatUseWerfImages: options.RmContainersThatUseWerfImages,
		DryRun:                        options.DryRun,
	}
}

type stagesPurgeManager struct {
	StorageManager                *manager.StorageManager
	ProjectName                   string
	RmContainersThatUseWerfImages bool
	DryRun                        bool
}

func (m *stagesPurgeManager) run(ctx context.Context) error {
	deleteStageOptions := manager.ForEachDeleteStageOptions{
		DeleteImageOptions: storage.DeleteImageOptions{
			RmiForce: true,
		},
		FilterStagesAndProcessRelatedDataOptions: storage.FilterStagesAndProcessRelatedDataOptions{
			SkipUsedImage:            false,
			RmForce:                  m.RmContainersThatUseWerfImages,
			RmContainersThatUseImage: m.RmContainersThatUseWerfImages,
		},
	}

	logProcess := logboek.Context(ctx).Default().LogProcess("Deleting stages")
	logProcess.Start()

	stages, err := m.StorageManager.GetAllStages(ctx)
	if err != nil {
		logProcess.Fail()
		return err
	}

	if err := deleteStageInStagesStorage(ctx, m.StorageManager, deleteStageOptions, m.DryRun, stages...); err != nil {
		logProcess.Fail()
		return err
	} else {
		logProcess.End()
	}

	logProcess = logboek.Context(ctx).Default().LogProcess("Deleting managed images")
	logProcess.Start()

	managedImages, err := m.StorageManager.StagesStorage.GetManagedImages(ctx, m.ProjectName)
	if err != nil {
		logProcess.Fail()
		return err
	}

	for _, managedImage := range managedImages {
		if !m.DryRun {
			if err := m.StorageManager.StagesStorage.RmManagedImage(ctx, m.ProjectName, managedImage); err != nil {
				return err
			}
		}

		logTag := managedImage
		if logTag == "" {
			logTag = storage.NamelessImageRecordTag
		}

		logboek.Context(ctx).Default().LogFDetails("  tag: %s\n", logTag)
		logboek.Context(ctx).LogOptionalLn()
	}

	logProcess.End()

	return nil
}
