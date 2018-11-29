package stage

type Interface interface {
	Name() StageName
	GetDependencies(Cache) string   // dependencies + builder_checksum
	GetContext(Cache) string        // context
	GetRelatedStageName() StageName // -> related_stage.context должен влиять на сигнатуру стадии

	SetSignature(signature string)
	GetSignature() string

	SetImage(Image)
	GetImage() Image

	SetGitArtifacts([]*GitArtifact)
	GetGitArtifacts() []*GitArtifact
}
