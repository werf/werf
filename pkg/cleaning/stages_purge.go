package cleaning

import (
	"fmt"
	"time"

	"github.com/flant/logboek"
	"github.com/flant/shluz"

	"github.com/flant/werf/pkg/storage"
)

type StagesPurgeOptions struct {
	RmContainersThatUseWerfImages bool
	DryRun                        bool
}

func StagesPurge(projectName string, stagesStorage storage.StagesStorage, options StagesPurgeOptions) error {
	m := newStagesPurgeManager(projectName, stagesStorage, options)

	return logboek.Default.LogProcess(
		"Running stages purge",
		logboek.LevelLogProcessOptions{Style: logboek.HighlightStyle()},
		m.run,
	)
}

func newStagesPurgeManager(projectName string, stagesStorage storage.StagesStorage, options StagesPurgeOptions) *stagesPurgeManager {
	return &stagesPurgeManager{
		StagesStorage:                 stagesStorage,
		ProjectName:                   projectName,
		RmContainersThatUseWerfImages: options.RmContainersThatUseWerfImages,
		DryRun:                        options.DryRun,
	}
}

type stagesPurgeManager struct {
	StagesStorage                 storage.StagesStorage
	ProjectName                   string
	RmContainersThatUseWerfImages bool
	DryRun                        bool
}

func (m *stagesPurgeManager) run() error {
	deleteImageOptions := storage.DeleteRepoImageOptions{
		RmiForce:                 true,
		SkipUsedImage:            false,
		RmForce:                  m.RmContainersThatUseWerfImages,
		RmContainersThatUseImage: m.RmContainersThatUseWerfImages,
	}

	lockName := fmt.Sprintf("stages-purge.%s-%s", m.StagesStorage.String(), m.ProjectName)
	return shluz.WithLock(lockName, shluz.LockOptions{Timeout: time.Second * 600}, func() error {
		logboek.Default.LogProcessStart("Deleting stages", logboek.LevelLogProcessStartOptions{})
		stagesRepoImageList, err := m.StagesStorage.GetRepoImages(m.ProjectName)
		if err != nil {
			logboek.Default.LogProcessFail(logboek.LevelLogProcessFailOptions{})
			return err
		}

		if err := deleteRepoImageInStagesStorage(m.StagesStorage, deleteImageOptions, m.DryRun, stagesRepoImageList...); err != nil {
			logboek.Default.LogProcessFail(logboek.LevelLogProcessFailOptions{})
			return err
		}
		logboek.Default.LogProcessEnd(logboek.LevelLogProcessEndOptions{})

		logboek.Default.LogProcessStart("Deleting managed images", logboek.LevelLogProcessStartOptions{})
		managedImages, err := m.StagesStorage.GetManagedImages(m.ProjectName)
		if err != nil {
			logboek.Default.LogProcessFail(logboek.LevelLogProcessFailOptions{})
			return err
		}

		for _, managedImage := range managedImages {
			if !m.DryRun {
				if err := m.StagesStorage.RmManagedImage(m.ProjectName, managedImage); err != nil {
					return err
				}
			}

			logboek.Default.LogFDetails("  tag: %s\n", managedImage)
			logboek.LogOptionalLn()
		}
		logboek.Default.LogProcessEnd(logboek.LevelLogProcessEndOptions{})

		return nil
	})
}
