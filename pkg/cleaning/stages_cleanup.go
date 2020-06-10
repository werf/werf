package cleaning

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/flant/lockgate"
	"github.com/werf/werf/pkg/werf"

	"github.com/werf/werf/pkg/stages_manager"

	"github.com/werf/logboek"

	"github.com/werf/werf/pkg/docker_registry"
	"github.com/werf/werf/pkg/image"
	"github.com/werf/werf/pkg/storage"
)

const stagesCleanupDefaultIgnorePeriodPolicy = 2 * 60 * 60

type StagesCleanupOptions struct {
	ImageNameList []string
	DryRun        bool
}

func StagesCleanup(projectName string, imagesRepo storage.ImagesRepo, storageLockManager storage.LockManager, stagesManager *stages_manager.StagesManager, options StagesCleanupOptions) error {
	m := newStagesCleanupManager(projectName, imagesRepo, stagesManager, options)

	if lock, err := storageLockManager.LockStagesAndImages(projectName, storage.LockStagesAndImagesOptions{GetOrCreateImagesOnly: false}); err != nil {
		return fmt.Errorf("unable to lock stages and images: %s", err)
	} else {
		defer storageLockManager.Unlock(lock)
	}

	return logboek.Default.LogProcess(
		"Running stages cleanup",
		logboek.LevelLogProcessOptions{Style: logboek.HighlightStyle()},
		m.run,
	)
}

func newStagesCleanupManager(projectName string, imagesRepo storage.ImagesRepo, stagesManager *stages_manager.StagesManager, options StagesCleanupOptions) *stagesCleanupManager {
	return &stagesCleanupManager{
		ImagesRepo:    imagesRepo,
		ImageNameList: options.ImageNameList,
		StagesManager: stagesManager,
		ProjectName:   projectName,
		DryRun:        options.DryRun,
	}
}

type stagesCleanupManager struct {
	imagesRepoImageList *[]*image.Info

	ImagesRepo    storage.ImagesRepo
	ImageNameList []string
	StagesManager *stages_manager.StagesManager
	ProjectName   string
	DryRun        bool
}

func (m *stagesCleanupManager) initImagesRepoImageList() error {
	repoImages, err := selectRepoImagesFromImagesRepo(m.ImagesRepo, m.ImageNameList)
	if err != nil {
		return err
	}

	m.setImagesRepoImageList(flattenRepoImages(repoImages))

	return nil
}

func (m *stagesCleanupManager) setImagesRepoImageList(repoImageList []*image.Info) {
	m.imagesRepoImageList = &repoImageList
}

func (m *stagesCleanupManager) getOrInitImagesRepoImageList() ([]*image.Info, error) {
	if m.imagesRepoImageList == nil {
		if err := m.initImagesRepoImageList(); err != nil {
			return nil, err
		}
	}

	return *m.imagesRepoImageList, nil
}

func (m *stagesCleanupManager) run() error {
	deleteImageOptions := storage.DeleteImageOptions{
		RmiForce:      false,
		SkipUsedImage: true,
		RmForce:       false,
	}

	lockName := fmt.Sprintf("stages-cleanup.%s-%s", m.StagesManager.StagesStorage.String(), m.ProjectName)
	return werf.WithHostLock(lockName, lockgate.AcquireOptions{Timeout: time.Second * 600}, func() error {
		stages, err := m.StagesManager.GetAllStages()
		if err != nil {
			return err
		}

		var stagesImageList []*image.Info
		stagesByImageName := map[string]*image.StageDescription{}
		for _, stageDesc := range stages {
			stagesImageList = append(stagesImageList, stageDesc.Info)
			stagesByImageName[stageDesc.Info.Name] = stageDesc
		}

		repoImageList, err := m.getOrInitImagesRepoImageList()
		if err != nil {
			return err
		}

		for _, repoImage := range repoImageList {
			stagesImageList = exceptRepoImageAndRelativesByImageID(stagesImageList, repoImage.ParentID)
		}

		var repoImageListToExcept []*image.Info
		if os.Getenv("WERF_DISABLE_STAGES_CLEANUP_DATE_PERIOD_POLICY") == "" {
			for _, repoImage := range stagesImageList {
				if time.Now().Unix()-repoImage.GetCreatedAt().Unix() < stagesCleanupDefaultIgnorePeriodPolicy {
					repoImageListToExcept = append(repoImageListToExcept, repoImage)
				}
			}
		}

		stagesImageList = exceptRepoImageList(stagesImageList, repoImageListToExcept...)

		var stagesToDeleteList []*image.StageDescription
		for _, imgInfo := range stagesImageList {
			if stagesByImageName[imgInfo.Name] == nil || stagesByImageName[imgInfo.Name].Info != imgInfo {
				panic(fmt.Sprintf("inconsistent state detected: %#v != %#v", stagesByImageName[imgInfo.Name].Info, imgInfo))
			}
			stagesToDeleteList = append(stagesToDeleteList, stagesByImageName[imgInfo.Name])
		}

		if err := deleteStageInStagesStorage(m.StagesManager, deleteImageOptions, m.DryRun, stagesToDeleteList...); err != nil {
			return err
		}

		return nil
	})
}

func exceptRepoImageAndRelativesByImageID(repoImageList []*image.Info, imageID string) []*image.Info {
	repoImage := findRepoImageByImageID(repoImageList, imageID)
	if repoImage == nil {
		return repoImageList
	}

	return exceptRepoImageAndRelativesByRepoImage(repoImageList, repoImage)
}

func findRepoImageByImageID(repoImageList []*image.Info, imageID string) *image.Info {
	for _, repoImage := range repoImageList {
		if repoImage.ID == imageID {
			return repoImage
		}
	}

	return nil
}

func exceptRepoImageAndRelativesByRepoImage(repoImageList []*image.Info, repoImage *image.Info) []*image.Info {
	for label, imageID := range repoImage.Labels {
		if strings.HasPrefix(label, image.WerfImportLabelPrefix) {
			repoImageList = exceptRepoImageAndRelativesByImageID(repoImageList, imageID)
		}
	}

	currentRepoImage := repoImage
	for {
		repoImageList = exceptRepoImageList(repoImageList, currentRepoImage)
		currentRepoImage = findRepoImageByImageID(repoImageList, currentRepoImage.ParentID)
		if currentRepoImage == nil {
			break
		}
	}

	return repoImageList
}

func exceptRepoImageList(repoImageList []*image.Info, repoImageListToExcept ...*image.Info) []*image.Info {
	var updatedRepoImageList []*image.Info

loop:
	for _, repoImage := range repoImageList {
		for _, repoImageToExcept := range repoImageListToExcept {
			if repoImage == repoImageToExcept {
				continue loop
			}
		}

		updatedRepoImageList = append(updatedRepoImageList, repoImage)
	}

	return updatedRepoImageList
}

func imagesRepoImageList(imagesRepo storage.ImagesRepo, imageNameList []string) ([]*image.Info, error) {
	repoImages, err := selectRepoImagesFromImagesRepo(imagesRepo, imageNameList)
	if err != nil {
		return nil, err
	}

	return flattenRepoImages(repoImages), nil
}

func flattenRepoImages(repoImages map[string][]*image.Info) (repoImageList []*image.Info) {
	for imageName := range repoImages {
		repoImageList = append(repoImageList, repoImages[imageName]...)
	}

	return
}

func deleteStageInStagesStorage(stagesManager *stages_manager.StagesManager, options storage.DeleteImageOptions, dryRun bool, stages ...*image.StageDescription) error {
	if err := deleteStage(stagesManager.DeleteStages, options, dryRun, stages...); err != nil {
		switch err.(type) {
		case docker_registry.DockerHubUnauthorizedError:
			return fmt.Errorf(`%s
You should specify Docker Hub token or username and password to remove tags with Docker Hub API.
Check --repo-docker-hub-token/username/password --stages-storage-repo-docker-hub-token/username/password options.
Be aware that access to the resource is forbidden with personal access token.
Read more details here https://werf.io/documentation/reference/working_with_docker_registries.html#docker-hub`, err)
		case docker_registry.GitHubPackagesUnauthorizedError:
			return fmt.Errorf(`%s
You should specify a token with the read:packages, write:packages, delete:packages and repo scopes to remove package versions.
Check --repo-github-token and --stages-storage-repo-github-token options.
Read more details here https://werf.io/documentation/reference/working_with_docker_registries.html#github-packages`, err)
		default:
			return err
		}
	}

	return nil
}

func deleteStage(f func(options storage.DeleteImageOptions, stages ...*image.StageDescription) error, options storage.DeleteImageOptions, dryRun bool, stages ...*image.StageDescription) error {
	for _, stageDesc := range stages {
		if !dryRun {
			if err := f(options, stageDesc); err != nil {
				return err
			}
		}

		logboek.Default.LogFDetails("  tag: %s\n", stageDesc.Info.Tag)
		logboek.LogOptionalLn()
	}

	return nil
}

func deleteRepoImage(f func(options storage.DeleteImageOptions, repoImageList ...*image.Info) error, options storage.DeleteImageOptions, dryRun bool, repoImageList ...*image.Info) error {
	for _, repoImage := range repoImageList {
		if !dryRun {
			if err := f(options, repoImage); err != nil {
				return err
			}
		}

		logboek.Default.LogFDetails("  tag: %s\n", repoImage.Tag)
		logboek.LogOptionalLn()
	}

	return nil
}
