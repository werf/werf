package stage

type Interface interface {
	Name() StageName
	// TODO: + error
	GetDependencies(Conveyor, Image) string // dependencies + builder_checksum
	IsEmpty(Conveyor, Image) (bool, error)
	GetContext(Conveyor) string     // context
	GetRelatedStageName() StageName // -> related_stage.context должен влиять на сигнатуру стадии

	SetSignature(signature string)
	GetSignature() string

	SetImage(Image)
	GetImage() Image

	SetGitArtifacts([]*GitArtifact)
	GetGitArtifacts() []*GitArtifact
}
