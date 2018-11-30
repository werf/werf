package stage

type Interface interface {
	Name() StageName

	IsEmpty(Conveyor, Image) (bool, error)

	GetDependencies(Conveyor, Image) (string, error)

	GetContext(Conveyor) (string, error)
	GetRelatedStageName() StageName

	PrepareImage(prevBuiltImage, image Image) error

	SetSignature(signature string)
	GetSignature() string

	SetImage(Image)
	GetImage() Image

	SetGitArtifacts([]*GitArtifact)
	GetGitArtifacts() []*GitArtifact
}
