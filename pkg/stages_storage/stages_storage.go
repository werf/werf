package stages_storage

import (
	"time"

	"github.com/flant/werf/pkg/image"
)

type ImageInfo struct {
	Signature string
	ImageName string
	Labels    map[string]string
	CreatedAt time.Time
}

type StagesStorage interface {
	// TODO cleanup GetAllImages() ([]StageImage, error)
	GetImagesBySignature(projectName, signature string) ([]*ImageInfo, error)

	// в том числе docker pull из registry + image.SyncDockerState
	// lock по имени image чтобы не делать 2 раза pull одновременно
	SyncStageImage(stageImage image.ImageInterface) error
	StoreStageImage(stageImage image.ImageInterface) error
	//InspectImage(imageName string) (*ImageInspect, error) ???

	String() string
}
