package stage

import "github.com/flant/dapp/pkg/image"

type Interface interface {
	Name() StageName

	IsEmpty(c Conveyor, prevBuiltImage image.Image) (bool, error)
	ShouldBeReset(builtImage image.Image) (bool, error)

	GetDependencies(c Conveyor, prevImage image.Image) (string, error)

	PrepareImage(c Conveyor, prevBuiltImage, image image.Image) error

	AfterImageSyncDockerStateHook(Conveyor) error
	PreRunHook(Conveyor) error

	SetSignature(signature string)
	GetSignature() string

	SetImage(image.Image)
	GetImage() image.Image

	SetGitPaths([]*GitPath)
	GetGitPaths() []*GitPath
}
