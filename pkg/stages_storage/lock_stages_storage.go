package stages_storage

import (
	"fmt"
	"time"

	"github.com/docker/docker/api/types/filters"

	"github.com/docker/docker/api/types"

	"github.com/flant/werf/pkg/docker"
	"github.com/flant/werf/pkg/image"
)

type LocalStagesStorage struct{}

func (storage *LocalStagesStorage) GetImagesBySignature(projectName, signature string) ([]*ImageInfo, error) {
	fmt.Printf("-- GetImagesBySignature %s\n", signature)

	filterSet := filters.NewArgs()
	filterSet.Add("reference", fmt.Sprintf("werf-stages-storage/%s:%s-*", projectName, signature))
	images, err := docker.Images(types.ImageListOptions{Filters: filterSet})
	if err != nil {
		return nil, fmt.Errorf("unable to get docker images: %s", err)
	}

	res := []*ImageInfo{}
	for _, img := range images {
		for _, repoTag := range img.RepoTags {
			res = append(res, &ImageInfo{
				ImageName: repoTag,
				Signature: signature,
				Labels:    img.Labels,
				CreatedAt: time.Unix(img.Created, 0),
			})
		}
	}

	return res, nil
}

func (storage *LocalStagesStorage) SyncStageImage(stageImage image.ImageInterface) error {
	fmt.Printf("-- SyncStageImage %s\n", stageImage.Name())
	return stageImage.SyncDockerState()
}

func (storage *LocalStagesStorage) StoreStageImage(stageImage image.ImageInterface) error {
	fmt.Printf("-- StoreImage %s\n", stageImage.Name())
	return nil
}

func (storage *LocalStagesStorage) String() string {
	return ":local"
}
