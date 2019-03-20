package stage

import "github.com/flant/werf/pkg/image"

type Interface interface {
	Name() StageName
	LogDetailedName() string

	IsEmpty(c Conveyor, prevBuiltImage image.ImageInterface) (bool, error)
	ShouldBeReset(builtImage image.ImageInterface) (bool, error)

	GetDependencies(c Conveyor, prevImage image.ImageInterface) (string, error)

	PrepareImage(c Conveyor, prevBuiltImage, image image.ImageInterface) error

	AfterImageSyncDockerStateHook(Conveyor) error
	PreRunHook(Conveyor) error

	SetSignature(signature string)
	GetSignature() string

	SetImage(image.ImageInterface)
	GetImage() image.ImageInterface

	SetGitMappings([]*GitMapping)
	GetGitMappings() []*GitMapping
}
