package stage

import (
	"github.com/flant/dapp/pkg/build/builder"
)

func generateBuildArtifactStage(dimgConfig interface{}) Interface {
	b := getBuilder(dimgConfig)
	if b != nil && !b.IsBuildArtifactEmpty() {
		return NewBuildArtifactStage(b)
	}

	return nil
}

func NewBuildArtifactStage(builder builder.Builder) *BuildArtifactStage {
	s := &BuildArtifactStage{}
	s.UserStage = newUserStage(builder)
	return s
}

type BuildArtifactStage struct {
	*UserStage
}

func (s *BuildArtifactStage) Name() string {
	return "buildArtifact"
}

func (s *BuildArtifactStage) GetContext(_ Cache) string {
	return s.builder.BuildArtifactChecksum() // TODO: git
}
