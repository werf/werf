package cleaning

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/docker/docker/api/types"

	"github.com/flant/logboek"
	"github.com/flant/shluz"

	"github.com/flant/werf/pkg/docker_registry"
	"github.com/flant/werf/pkg/image"
)

const stagesCleanupDefaultIgnorePeriodPolicy = 2 * 60 * 60

type StagesCleanupOptions struct {
	ProjectName       string
	ImagesRepoManager ImagesRepoManager
	StagesStorage     string
	ImagesNames       []string
	DryRun            bool
}

func StagesCleanup(options StagesCleanupOptions) error {
	return logboek.Default.LogProcess(
		"Running stages cleanup",
		logboek.LevelLogProcessOptions{Style: logboek.HighlightStyle()},
		func() error {
			return stagesCleanup(options)
		},
	)
}

func stagesCleanup(options StagesCleanupOptions) error {
	commonProjectOptions := CommonProjectOptions{
		ProjectName: options.ProjectName,
		CommonOptions: CommonOptions{
			SkipUsedImages: true,
			RmiForce:       false,
			RmForce:        false,
			DryRun:         options.DryRun,
		},
	}

	commonRepoOptions := CommonRepoOptions{
		ImagesRepoManager: options.ImagesRepoManager,
		StagesStorage:     options.StagesStorage,
		ImagesNames:       options.ImagesNames,
		DryRun:            options.DryRun,
	}

	projectStagesCleanupLockName := fmt.Sprintf("stages-cleanup.%s", commonProjectOptions.ProjectName)
	return shluz.WithLock(projectStagesCleanupLockName, shluz.LockOptions{Timeout: time.Second * 600}, func() error {
		repoImages, err := repoImages(commonRepoOptions)
		if err != nil {
			return err
		}

		if len(repoImages) != 0 {
			if commonRepoOptions.StagesStorage == localStagesStorage {
				if err := projectImageStagesSyncByRepoImages(repoImages, commonProjectOptions); err != nil {
					return err
				}
			} else {
				if err := repoImageStagesSyncByRepoImages(repoImages, commonRepoOptions); err != nil {
					return err
				}
			}
		} else {
			if err := projectStagesPurge(commonProjectOptions); err != nil {
				return err
			}
		}

		return nil
	})
}

func repoImageStagesSyncByRepoImages(repoImages []docker_registry.RepoImage, options CommonRepoOptions) error {
	repoImageStages, err := repoImageStagesImages(options)
	if err != nil {
		return err
	}

	if len(repoImageStages) == 0 {
		return nil
	}

	for _, repoImage := range repoImages {
		parentId, err := repoImageParentId(repoImage)
		if err != nil {
			return err
		}

		repoImageStages, err = exceptRepoImageStagesByImageId(repoImageStages, parentId)
		if err != nil {
			return err
		}
	}

	err = repoImagesRemove(repoImageStages, options)
	if err != nil {
		return err
	}

	return nil
}

func exceptRepoImageStagesByImageId(repoImageStages []docker_registry.RepoImage, imageId string) ([]docker_registry.RepoImage, error) {
	repoImageStage, err := findRepoImageStageByImageId(repoImageStages, imageId)
	if err != nil {
		return nil, err
	} else if repoImageStage == nil {
		return repoImageStages, nil
	}

	repoImageStages, err = exceptRepoImageStagesByRepoImageStage(repoImageStages, *repoImageStage)
	if err != nil {
		return nil, err
	}

	return repoImageStages, nil
}

func findRepoImageStageByImageId(repoImageStages []docker_registry.RepoImage, imageId string) (*docker_registry.RepoImage, error) {
	for _, repoImageStage := range repoImageStages {
		manifest, err := repoImageStage.Manifest()
		if err != nil {
			return nil, err
		}

		repoImageStageImageId := manifest.Config.Digest.String()
		if repoImageStageImageId == imageId {
			return &repoImageStage, nil
		}
	}

	return nil, nil
}

func exceptRepoImageStagesByRepoImageStage(repoImageStages []docker_registry.RepoImage, repoImageStage docker_registry.RepoImage) ([]docker_registry.RepoImage, error) {
	labels, err := repoImageLabels(repoImageStage)
	if err != nil {
		return nil, err
	}

	for label, imageID := range labels {
		if strings.HasPrefix(label, image.WerfImportLabelPrefix) {
			repoImageStages, err = exceptRepoImageStagesByImageId(repoImageStages, imageID)
			if err != nil {
				return nil, err
			}
		}
	}

	currentRepoImageStage := &repoImageStage
	for {
		repoImageStages = exceptRepoImages(repoImageStages, *currentRepoImageStage)

		parentId, err := repoImageParentId(*currentRepoImageStage)
		if err != nil {
			return nil, err
		}

		currentRepoImageStage, err = findRepoImageStageByImageId(repoImageStages, parentId)
		if err != nil {
			return nil, err
		}

		if currentRepoImageStage == nil {
			break
		}
	}

	return repoImageStages, nil
}

func repoImageParentId(repoImage docker_registry.RepoImage) (string, error) {
	configFile, err := repoImage.Image.ConfigFile()
	if err != nil {
		return "", err
	}

	return configFile.Config.Image, nil
}

func repoImageLabels(repoImage docker_registry.RepoImage) (map[string]string, error) {
	configFile, err := repoImage.Image.ConfigFile()
	if err != nil {
		return nil, err
	}

	return configFile.Config.Labels, nil
}

func repoImageCreated(repoImage docker_registry.RepoImage) (time.Time, error) {
	configFile, err := repoImage.Image.ConfigFile()
	if err != nil {
		return time.Time{}, err
	}

	return configFile.Created.Time, nil
}

func projectImageStagesSyncByRepoImages(repoImages []docker_registry.RepoImage, options CommonProjectOptions) error {
	imageStages, err := projectImageStages(options)
	if err != nil {
		return err
	}

	for _, repoImage := range repoImages {
		parentId, err := repoImageParentId(repoImage)
		if err != nil {
			return err
		}

		imageStages, err = exceptImageStagesByImageId(imageStages, parentId, options)
		if err != nil {
			return err
		}
	}

	if os.Getenv("WERF_DISABLE_STAGES_CLEANUP_DATE_PERIOD_POLICY") == "" {
		for _, imageStage := range imageStages {
			if time.Now().Unix()-imageStage.Created < stagesCleanupDefaultIgnorePeriodPolicy {
				imageStages = exceptImage(imageStages, imageStage)
			}
		}
	}

	imageStages, err = processUsedImages(imageStages, options.CommonOptions)
	if err != nil {
		return err
	}

	err = imagesRemove(imageStages, options.CommonOptions)
	if err != nil {
		return err
	}

	return nil
}

func exceptImageStagesByImageId(imageStages []types.ImageSummary, imageId string, options CommonProjectOptions) ([]types.ImageSummary, error) {
	imageStage := findImageStageByImageId(imageStages, imageId)
	if imageStage == nil {
		return imageStages, nil
	}

	imageStages, err := exceptImageStagesByImageStage(imageStages, *imageStage, options)
	if err != nil {
		return nil, err
	}

	return imageStages, nil
}

func exceptImageStagesByImageStage(imageStages []types.ImageSummary, imageStage types.ImageSummary, commonProjectOptions CommonProjectOptions) ([]types.ImageSummary, error) {
	var err error
	for label, imageID := range imageStage.Labels {
		if strings.HasPrefix(label, image.WerfImportLabelPrefix) {
			imageStages, err = exceptImageStagesByImageId(imageStages, imageID, commonProjectOptions)
			if err != nil {
				return nil, err
			}
		}
	}

	currentImageStage := &imageStage
	for {
		imageStages = exceptImage(imageStages, *currentImageStage)
		currentImageStage = findImageStageByImageId(imageStages, currentImageStage.ParentID)
		if currentImageStage == nil {
			break
		}
	}

	return imageStages, nil
}

func findImageStageByImageId(imageStages []types.ImageSummary, imageId string) *types.ImageSummary {
	for _, imageStage := range imageStages {
		if imageStage.ID == imageId {
			return &imageStage
		}
	}

	return nil
}

func projectImageStages(options CommonProjectOptions) ([]types.ImageSummary, error) {
	images, err := werfImagesByFilterSet(projectImageStageFilterSet(options))
	if err != nil {
		return nil, err
	}

	return images, nil
}
