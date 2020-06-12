package cleaning

import (
	"fmt"

	"github.com/werf/logboek"

	"github.com/werf/werf/pkg/stages_manager"
	"github.com/werf/werf/pkg/storage"
)

type PurgeOptions struct {
	ImagesPurgeOptions
	StagesPurgeOptions
}

func Purge(projectName string, imagesRepo storage.ImagesRepo, storageLockManager storage.LockManager, stagesManager *stages_manager.StagesManager, options PurgeOptions) error {
	m := newPurgeManager(projectName, imagesRepo, stagesManager, options)

	if lock, err := storageLockManager.LockStagesAndImages(projectName, storage.LockStagesAndImagesOptions{GetOrCreateImagesOnly: false}); err != nil {
		return fmt.Errorf("unable to lock stages and images: %s", err)
	} else {
		defer storageLockManager.Unlock(lock)
	}

	if err := logboek.Default.LogProcess(
		"Running images purge",
		logboek.LevelLogProcessOptions{Style: logboek.HighlightStyle()},
		m.imagesPurgeManager.run,
	); err != nil {
		return err
	}

	if err := logboek.Default.LogProcess(
		"Running stages purge",
		logboek.LevelLogProcessOptions{Style: logboek.HighlightStyle()},
		m.stagesPurgeManager.run,
	); err != nil {
		return err
	}

	return nil
}

func newPurgeManager(projectName string, imagesRepo storage.ImagesRepo, stagesManager *stages_manager.StagesManager, options PurgeOptions) *purgeManager {
	return &purgeManager{
		imagesPurgeManager: newImagesPurgeManager(imagesRepo, options.ImagesPurgeOptions),
		stagesPurgeManager: newStagesPurgeManager(projectName, stagesManager, options.StagesPurgeOptions),
	}
}

type purgeManager struct {
	*imagesPurgeManager
	*stagesPurgeManager
}
