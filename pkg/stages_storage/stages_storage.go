package stages_storage

import (
	"time"

	"github.com/flant/werf/pkg/image"
)

type ImageInfo struct {
	Signature         string            `json:"signature"`
	ImageName         string            `json:"imageName"`
	Labels            map[string]string `json:"labels"`
	CreatedAtUnixNano int64             `json:"createdAtUnixNano"`
}

func (info *ImageInfo) CreatedAt() time.Time {
	return time.Unix(info.CreatedAtUnixNano/1000_000_000, info.CreatedAtUnixNano%1000_000_000)
}

type StagesStorage interface {
	// TODO cleanup GetAllImages() ([]StageImage, error)
	GetImagesBySignature(projectName, signature string) ([]*ImageInfo, error)

	// в том числе docker pull из registry + image.SyncDockerState
	// lock по имени image чтобы не делать 2 раза pull одновременно
	SyncStageImage(stageImage image.ImageInterface) error
	StoreStageImage(stageImage image.ImageInterface) error

	String() string
}
