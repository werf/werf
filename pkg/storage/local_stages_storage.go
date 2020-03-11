package storage

import (
	"fmt"
	"strings"

	"github.com/flant/logboek"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"

	"github.com/flant/werf/pkg/docker"
	"github.com/flant/werf/pkg/image"
)

type LocalStagesStorage struct {
	StagesStorage // FIXME
}

const NamelessImageRecordTag = "__nameless__"

func makeConfigImageRecordImageName(projectName, imageName string) string {
	tag := imageName
	if imageName == "" {
		tag = NamelessImageRecordTag
	}
	return fmt.Sprintf(image.ManagedImageRecord_ImageFormat, projectName, tag)
}

func (storage *LocalStagesStorage) AddManagedImage(projectName, imageName string) error {
	logboek.Debug.LogF("-- LocalStagesStorage.AddManagedImage %s %s\n", projectName, imageName)

	fullImageName := makeConfigImageRecordImageName(projectName, imageName)

	if exsts, err := docker.ImageExist(fullImageName); err != nil {
		return fmt.Errorf("unable to check existence of image %q: %s", fullImageName, err)
	} else if exsts {
		return nil
	}

	if err := docker.CreateImage(fullImageName); err != nil {
		return fmt.Errorf("unable to create image %q: %s", fullImageName, err)
	}
	return nil
}

func (storage *LocalStagesStorage) RmManagedImage(projectName, imageName string) error {
	logboek.Debug.LogF("-- LocalStagesStorage.RmManagedImage %s %s\n", projectName, imageName)

	fullImageName := makeConfigImageRecordImageName(projectName, imageName)

	if exsts, err := docker.ImageExist(fullImageName); err != nil {
		return fmt.Errorf("unable to check existence of image %q: %s", fullImageName, err)
	} else if !exsts {
		return nil
	}

	if err := docker.CliRmi("--force", fullImageName); err != nil {
		return fmt.Errorf("unable to remove image %q: %s", fullImageName, err)
	}
	return nil
}

func (storage *LocalStagesStorage) GetManagedImages(projectName string) ([]string, error) {
	logboek.Debug.LogF("-- LocalStagesStorage.GetManagedImages %s\n", projectName)

	filterSet := filters.NewArgs()
	filterSet.Add("reference", fmt.Sprintf(image.ManagedImageRecord_ImageNameFormat, projectName))

	images, err := docker.Images(types.ImageListOptions{Filters: filterSet})
	if err != nil {
		return nil, fmt.Errorf("unable to get docker images: %s", err)
	}

	res := []string{}
	for _, img := range images {
		for _, repoTag := range img.RepoTags {
			tag := strings.SplitN(repoTag, ":", 2)[1]

			if tag == NamelessImageRecordTag {
				res = append(res, "")
			} else {
				res = append(res, tag)
			}
		}
	}
	return res, nil
}

func (storage *LocalStagesStorage) GetRepoImagesBySignature(projectName, signature string) ([]*image.Info, error) {
	filterSet := filters.NewArgs()
	filterSet.Add("reference", fmt.Sprintf(image.LocalImageStageImageNameFormat, projectName))
	filterSet.Add("label", fmt.Sprintf("%s=%s", image.WerfStageSignatureLabel, signature))

	images, err := docker.Images(types.ImageListOptions{Filters: filterSet})
	if err != nil {
		return nil, fmt.Errorf("unable to get docker images: %s", err)
	}

	res := []*image.Info{}
	for _, img := range images {
		for _, repoTag := range img.RepoTags {
			res = append(res, &image.Info{
				ImageName:         repoTag,
				Signature:         signature,
				Labels:            img.Labels,
				CreatedAtUnixNano: img.Created * 1000_000_000,
			})
		}
	}

	return res, nil
}

func (storage *LocalStagesStorage) SyncStageImage(stageImage image.ImageInterface) error {
	return stageImage.SyncDockerState()
}

func (storage *LocalStagesStorage) StoreStageImage(stageImage image.ImageInterface) error {
	if err := stageImage.TagBuiltImage(stageImage.Name()); err != nil {
		return fmt.Errorf("unable to save image %s: %s", stageImage.Name(), err)
	}
	if err := stageImage.SyncDockerState(); err != nil {
		return fmt.Errorf("unable to sync docker state of image %s: %s", stageImage.Name(), err)
	}
	return nil
}

func (storage *LocalStagesStorage) String() string {
	return ":local"
}
