package cleaning

import (
	"github.com/flant/logboek"
	"github.com/flant/werf/pkg/stages_manager"

	"github.com/flant/werf/pkg/storage"
)

type CleanupOptions struct {
	ImagesCleanupOptions
	StagesCleanupOptions
}

func Cleanup(projectName string, imagesRepo storage.ImagesRepo, stagesManager *stages_manager.StagesManager, options CleanupOptions) error {
	m := newCleanupManager(projectName, imagesRepo, stagesManager, options)

	if err := logboek.Default.LogProcess(
		"Running images cleanup",
		logboek.LevelLogProcessOptions{Style: logboek.HighlightStyle()},
		m.imagesCleanupManager.run,
	); err != nil {
		return err
	}

	repoImages := m.imagesCleanupManager.getImagesRepoImages()
	m.stagesCleanupManager.setImagesRepoImageList(flattenRepoImages(repoImages))

	if err := logboek.Default.LogProcess(
		"Running stages cleanup",
		logboek.LevelLogProcessOptions{Style: logboek.HighlightStyle()},
		m.stagesCleanupManager.run,
	); err != nil {
		return err
	}

	return nil
}

func newCleanupManager(projectName string, imagesRepo storage.ImagesRepo, stagesManager *stages_manager.StagesManager, options CleanupOptions) *cleanupManager {
	return &cleanupManager{
		imagesCleanupManager: newImagesCleanupManager(imagesRepo, options.ImagesCleanupOptions),
		stagesCleanupManager: newStagesCleanupManager(projectName, imagesRepo, stagesManager, options.StagesCleanupOptions),
	}
}

type cleanupManager struct {
	*imagesCleanupManager
	*stagesCleanupManager
}
