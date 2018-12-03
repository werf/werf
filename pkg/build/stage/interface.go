package stage

import "github.com/flant/dapp/pkg/image"

type Interface interface {
	Name() StageName

	IsEmpty(c Conveyor, prevBuiltImage image.Image) (bool, error)

	GetDependencies(c Conveyor, prevImage image.Image) (string, error)

	GetContext(Conveyor) (string, error)
	GetRelatedStageName() StageName

	PrepareImage(c Conveyor, prevBuiltImage, image image.Image) error

	SetSignature(signature string)
	GetSignature() string

	SetImage(image.Image)
	GetImage() image.Image

	SetGitArtifacts([]*GitArtifact)
	GetGitArtifacts() []*GitArtifact
}
