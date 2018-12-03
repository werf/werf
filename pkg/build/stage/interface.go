package stage

import "github.com/flant/dapp/pkg/image"

type Interface interface {
	Name() StageName

	IsEmpty(Conveyor, image.Image) (bool, error)

	GetDependencies(Conveyor, image.Image) (string, error)

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
