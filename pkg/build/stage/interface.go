package stage

import "github.com/flant/dapp/pkg/build"

type Interface interface {
	ReadDockerState() error
	IsImageExists() bool

	Name() StageName
	GetDependencies(Cache) string   // dependencies + builder_checksum
	GetContext(Cache) string        // context
	GetRelatedStageName() StageName // -> related_stage.context должен влиять на сигнатуру стадии

	SetSignature(signature string)
	GetSignature() string

	SetImage(Image)
	GetImage() Image

	SetGitArtifacts([]*build.GitArtifact)
	GetGitArtifacts() []*build.GitArtifact
}
