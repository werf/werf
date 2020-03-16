package cleaning

import (
	"github.com/flant/logboek"

	"github.com/flant/werf/pkg/storage"
)

type PurgeOptions struct {
	ImagesPurgeOptions
	StagesPurgeOptions
}

func Purge(projectName string, imagesRepo storage.ImagesRepo, stagesStorage storage.StagesStorage, options PurgeOptions) error {
	m := newPurgeManager(projectName, imagesRepo, stagesStorage, options)

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

func newPurgeManager(projectName string, imagesRepo storage.ImagesRepo, stagesStorage storage.StagesStorage, options PurgeOptions) *purgeManager {
	return &purgeManager{
		imagesPurgeManager: newImagesPurgeManager(imagesRepo, options.ImagesPurgeOptions),
		stagesPurgeManager: newStagesPurgeManager(projectName, stagesStorage, options.StagesPurgeOptions),
	}
}

type purgeManager struct {
	*imagesPurgeManager
	*stagesPurgeManager
}
