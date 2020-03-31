package cleaning

import (
	"fmt"
	"time"

	"github.com/flant/werf/pkg/stages_manager"

	"github.com/flant/logboek"
	"github.com/flant/shluz"

	"github.com/flant/werf/pkg/storage"
)

type StagesPurgeOptions struct {
	RmContainersThatUseWerfImages bool
	DryRun                        bool
}

func StagesPurge(projectName string, stagesManager *stages_manager.StagesManager, options StagesPurgeOptions) error {
	m := newStagesPurgeManager(projectName, stagesManager, options)

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
	return shluz.WithLock(lockName, shluz.LockOptions{Timeout: time.Second * 600}, func() error {
		logboek.Default.LogProcessStart("Deleting stages", logboek.LevelLogProcessStartOptions{})
		stagesImageList, err := m.StagesManager.GetAllStages()
		if err != nil {
			logboek.Default.LogProcessFail(logboek.LevelLogProcessFailOptions{})
			return err
		}

		if err := deleteStageInStagesStorage(m.StagesManager, deleteImageOptions, m.DryRun, stagesImageList...); err != nil {
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
