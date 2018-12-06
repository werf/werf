package cleanup

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/docker/docker/api/types"

	"github.com/flant/dapp/pkg/build"
	"github.com/flant/dapp/pkg/docker_registry"
	"github.com/flant/dapp/pkg/lock"
)

const syncIgnoreProjectDimgstagePeriod = 2 * 60 * 60

func ProjectDimgstagesSync(commonProjectOptions CommonProjectOptions, commonRepoOptions CommonRepoOptions) error {
	projectImagesLockName := fmt.Sprintf("%s.images", commonProjectOptions.ProjectName)
	err := lock.WithLock(projectImagesLockName, lock.LockOptions{Timeout: time.Second * 600}, func() error {
		if commonRepoOptions.Repository != "" {
			err := lock.WithLock(commonRepoOptions.Repository, lock.LockOptions{ReadOnly: true, Timeout: time.Second * 600}, func() error {
				if err := projectDimgstagesSyncByRepoDimgs(commonProjectOptions, commonRepoOptions); err != nil {
					return err
				}

				return nil
			})

			if err != nil {
				return err
			}
		}

		if err := projectDimgstagesSyncByCacheVersion(commonProjectOptions); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	if err := projectCleanup(commonProjectOptions); err != nil {
		return err
	}

	return nil
}

func repoDimgstagesSyncByRepoDimgs(repoDimgs []docker_registry.RepoImage, options CommonRepoOptions) error {
	repoDimgstages, err := repoDimgstageImages(options)
	if err != nil {
		return err
	}

	if len(repoDimgstages) == 0 {
		return nil
	}

	for _, repoDimg := range repoDimgs {
		parentId, err := repoImageParentId(repoDimg)
		if err != nil {
			return err
		}

		repoDimgstages, err = exceptRepoDimgstagesByImageId(repoDimgstages, parentId)
		if err != nil {
			return err
		}
	}

	err = repoImagesRemove(repoDimgstages, options)
	if err != nil {
		return err
	}

	return nil
}

func repoDimgstagesSyncByCacheVersion(options CommonRepoOptions) error {
	repoDimgstages, err := repoDimgstageImages(options)
	if err != nil {
		return err
	}

	var repoImagesToDelete []docker_registry.RepoImage
	for _, repoDimgstage := range repoDimgstages {
		labels, err := repoImageLabels(repoDimgstage)
		if err != nil {
			return err
		}

		version, ok := labels[build.DappCacheVersionLabel]
		if !ok || (version != build.BuildCacheVersion) {
			fmt.Printf("%s %s %s\n", repoDimgstage.Tag, version, build.BuildCacheVersion)
			repoImagesToDelete = append(repoImagesToDelete, repoDimgstage)
		}
	}

	if err := repoImagesRemove(repoImagesToDelete, options); err != nil {
		return err
	}

	return nil
}

func exceptRepoDimgstagesByImageId(repoDimgstages []docker_registry.RepoImage, imageId string) ([]docker_registry.RepoImage, error) {
	repoDimgstage, err := findRepoDimgstageByImageId(repoDimgstages, imageId)
	if repoDimgstage == nil {
		return repoDimgstages, nil
	}

	repoDimgstages, err = exceptRepoDimgstagesByRepoDimgstage(repoDimgstages, *repoDimgstage)
	if err != nil {
		return nil, err
	}

	return repoDimgstages, nil
}

func findRepoDimgstageByImageId(repoDimgstages []docker_registry.RepoImage, imageId string) (*docker_registry.RepoImage, error) {
	for _, repoDimgstage := range repoDimgstages {
		manifest, err := repoDimgstage.Manifest()
		if err != nil {
			return nil, err
		}

		repoDimgstageImageId := manifest.Config.Digest.String()
		if repoDimgstageImageId == imageId {
			return &repoDimgstage, nil
		}
	}

	return nil, nil
}

func exceptRepoDimgstagesByRepoDimgstage(repoDimgstages []docker_registry.RepoImage, repoDimgstage docker_registry.RepoImage) ([]docker_registry.RepoImage, error) {
	labels, err := repoImageLabels(repoDimgstage)
	if err != nil {
		return nil, err
	}

	for label, value := range labels {
		if strings.HasPrefix(label, "dapp-artifact") {
			repoDimgstages, err = exceptRepoDimgstagesByImageId(repoDimgstages, value)
			if err != nil {
				return nil, err
			}
		}
	}

	currentRepoDimgstage := &repoDimgstage
	for {
		repoDimgstages = exceptRepoImages(repoDimgstages, *currentRepoDimgstage)

		parentId, err := repoImageParentId(*currentRepoDimgstage)
		if err != nil {
			return nil, err
		}

		currentRepoDimgstage, err = findRepoDimgstageByImageId(repoDimgstages, parentId)
		if err != nil {
			return nil, err
		}

		if currentRepoDimgstage == nil {
			break
		}
	}

	return repoDimgstages, nil
}

func repoImageParentId(repoImage docker_registry.RepoImage) (string, error) {
	configFile, err := repoImage.Image.ConfigFile()
	if err != nil {
		return "", err
	}

	return configFile.ContainerConfig.Image, nil
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

func projectDimgstagesSyncByRepoDimgs(commonProjectOptions CommonProjectOptions, commonRepoOptions CommonRepoOptions) error {
	repoDimgs, err := repoDimgImages(commonRepoOptions)
	if err != nil {
		return err
	}

	dimgstages, err := projectDimgstages(commonProjectOptions)
	if err != nil {
		return err
	}

	for _, repoDimg := range repoDimgs {
		parentId, err := repoImageParentId(repoDimg)
		if err != nil {
			return err
		}

		dimgstages, err = exceptDimgstagesByImageId(dimgstages, parentId)
		if err != nil {
			return err
		}
	}

	if os.Getenv("DAPP_STAGES_SYNC_LOCAL_DISABLED_DATE_POLICY") == "" {
		for _, dimgstage := range dimgstages {
			if time.Now().Unix()-dimgstage.Created < syncIgnoreProjectDimgstagePeriod {
				dimgstages = exceptImage(dimgstages, dimgstage)
			}
		}
	}

	err = imagesRemove(dimgstages, commonProjectOptions.CommonOptions)
	if err != nil {
		return err
	}

	return nil
}

func exceptDimgstagesByImageId(dimgstages []types.ImageSummary, imageId string) ([]types.ImageSummary, error) {
	dimgstage := findDimgstageByImageId(dimgstages, imageId)
	if dimgstage == nil {
		return dimgstages, nil
	}

	dimgstages, err := exceptDimgstagesByDimgstage(dimgstages, *dimgstage)
	if err != nil {
		return nil, err
	}

	return dimgstages, nil
}

func exceptDimgstagesByDimgstage(dimgstages []types.ImageSummary, dimgstage types.ImageSummary) ([]types.ImageSummary, error) {
	var err error
	for label, value := range dimgstage.Labels {
		if strings.HasPrefix(label, "dapp-artifact") {
			dimgstages, err = exceptDimgstagesByImageId(dimgstages, value)
			if err != nil {
				return nil, err
			}
		}
	}

	currentDimgstage := &dimgstage
	for {
		dimgstages = exceptImage(dimgstages, *currentDimgstage)
		currentDimgstage = findDimgstageByImageId(dimgstages, currentDimgstage.ParentID)
		if currentDimgstage == nil {
			break
		}
	}

	return dimgstages, nil
}

func findDimgstageByImageId(dimgstages []types.ImageSummary, imageId string) *types.ImageSummary {
	for _, dimgstage := range dimgstages {
		if dimgstage.ID == imageId {
			return &dimgstage
		}
	}

	return nil
}

func projectDimgstages(options CommonProjectOptions) ([]types.ImageSummary, error) {
	images, err := dappImagesByFilterSet(projectDimgstageFilterSet(options))
	if err != nil {
		return nil, err
	}

	return images, nil
}

func projectDimgstagesSyncByCacheVersion(options CommonProjectOptions) error {
	return dappDimgstagesFlushByCacheVersion(projectDimgstageFilterSet(options), options.CommonOptions)
}
