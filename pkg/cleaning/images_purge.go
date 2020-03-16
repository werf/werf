package cleaning

import (
	"github.com/flant/logboek"

	"github.com/flant/werf/pkg/image"
	"github.com/flant/werf/pkg/storage"
)

type ImagesPurgeOptions struct {
	ImageNameList []string
	DryRun        bool
}

func ImagesPurge(imagesRepo storage.ImagesRepo, options ImagesPurgeOptions) error {
	m := newImagesPurgeManager(imagesRepo, options)

	return logboek.Default.LogProcess(
		"Running images purge",
		logboek.LevelLogProcessOptions{Style: logboek.HighlightStyle()},
		m.run,
	)
}

func newImagesPurgeManager(imagesRepo storage.ImagesRepo, options ImagesPurgeOptions) *imagesPurgeManager {
	return &imagesPurgeManager{
		ImagesRepo:    imagesRepo,
		ImageNameList: options.ImageNameList,
		DryRun:        options.DryRun,
	}
}

type imagesPurgeManager struct {
	ImagesRepo    storage.ImagesRepo
	ImageNameList []string
	DryRun        bool
}

func (m *imagesPurgeManager) run() error {
	repoImageList, err := imagesRepoImageList(m.ImagesRepo, m.ImageNameList)
	if err != nil {
		return err
	}

	return deleteRepoImageInImagesRepo(m.ImagesRepo, m.DryRun, repoImageList...)
}

func deleteRepoImageInImagesRepo(imagesRepo storage.ImagesRepo, dryRun bool, repoImageList ...*image.Info) error {
	return deleteRepoImage(imagesRepo.DeleteRepoImage, storage.DeleteRepoImageOptions{}, dryRun, repoImageList...)
}
