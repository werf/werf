package storage

import (
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"

	"github.com/flant/werf/pkg/docker"
	"github.com/flant/werf/pkg/image"
)

type LocalStagesStorage struct{}

func (storage *LocalStagesStorage) GetImagesBySignature(projectName, signature string) ([]*ImageInfo, error) {
	filterSet := filters.NewArgs()
	filterSet.Add("reference", fmt.Sprintf(image.LocalImageStageImageNameFormat, projectName))
	filterSet.Add("label", fmt.Sprintf("%s=%s", image.WerfStageSignatureLabel, signature))

	images, err := docker.Images(types.ImageListOptions{Filters: filterSet})
	if err != nil {
		return nil, fmt.Errorf("unable to get docker images: %s", err)
	}

	res := []*ImageInfo{}
	for _, img := range images {
		for _, repoTag := range img.RepoTags {
			res = append(res, &ImageInfo{
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
