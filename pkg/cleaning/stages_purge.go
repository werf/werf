package cleaning

import (
	"fmt"
	"time"

	"github.com/flant/lockgate"
	"github.com/werf/werf/pkg/werf"

	"github.com/werf/werf/pkg/stages_manager"

	"github.com/flant/logboek"
	"github.com/werf/werf/pkg/storage"
)

type StagesPurgeOptions struct {
	RmContainersThatUseWerfImages bool
	DryRun                        bool
}

func StagesPurge(projectName string, storageLockManager storage.LockManager, stagesManager *stages_manager.StagesManager, options StagesPurgeOptions) error {
	m := newStagesPurgeManager(projectName, stagesManager, options)

	if lock, err := storageLockManager.LockStagesAndImages(projectName, storage.LockStagesAndImagesOptions{GetOrCreateImagesOnly: false}); err != nil {
		return fmt.Errorf("unable to lock stages and images: %s", err)
	} else {
		defer storageLockManager.Unlock(lock)
	}

	return logboek.Default.LogProcess(
		"Running stages purge",
		logboek.LevelLogProcessOptions{Style: logboek.HighlightStyle()},
		m.run,
	)
}

func newStagesPurgeManager(projectName string, stagesManager *stages_manager.StagesManager, options StagesPurgeOptions) *stagesPurgeManager {
	return &stagesPurgeManager{
		StagesManager:                 stagesManager,
		ProjectName:                   projectName,
		RmContainersThatUseWerfImages: options.RmContainersThatUseWerfImages,
		DryRun:                        options.DryRun,
	}
}

type stagesPurgeManager struct {
	StagesManager                 *stages_manager.StagesManager
	ProjectName                   string
	RmContainersThatUseWerfImages bool
	DryRun                        bool
}

func (m *stagesPurgeManager) run() error {
	deleteImageOptions := storage.DeleteImageOptions{
		RmiForce:                 true,
		SkipUsedImage:            false,
		RmForce:                  m.RmContainersThatUseWerfImages,
		RmContainersThatUseImage: m.RmContainersThatUseWerfImages,
	}

	lockName := fmt.Sprintf("stages-purge.%s", m.ProjectName)
	return werf.WithHostLock(lockName, lockgate.AcquireOptions{Timeout: time.Second * 600}, func() error {
		logboek.Default.LogProcessStart("Deleting stages", logboek.LevelLogProcessStartOptions{})

		stages, err := m.StagesManager.GetAllStages()
		if err != nil {
			logboek.Default.LogProcessFail(logboek.LevelLogProcessFailOptions{})
			return err
		}

		if err := deleteStageInStagesStorage(m.StagesManager, deleteImageOptions, m.DryRun, stages...); err != nil {
			logboek.Default.LogProcessFail(logboek.LevelLogProcessFailOptions{})
			return err
		}
		logboek.Default.LogProcessEnd(logboek.LevelLogProcessEndOptions{})

		logboek.Default.LogProcessStart("Deleting managed images", logboek.LevelLogProcessStartOptions{})
		managedImages, err := m.StagesManager.StagesStorage.GetManagedImages(m.ProjectName)
		if err != nil {
			logboek.Default.LogProcessFail(logboek.LevelLogProcessFailOptions{})
			return err
		}

		for _, managedImage := range managedImages {
			if !m.DryRun {
				if err := m.StagesManager.StagesStorage.RmManagedImage(m.ProjectName, managedImage); err != nil {
					return err
				}
			}

			logTag := managedImage
			if logTag == "" {
				logTag = storage.NamelessImageRecordTag
			}

			logboek.Default.LogFDetails("  tag: %s\n", logTag)
			logboek.LogOptionalLn()
		}
		logboek.Default.LogProcessEnd(logboek.LevelLogProcessEndOptions{})

		return nil
	})
}
