package stage

import (
	"github.com/flant/dapp/pkg/build/builder"
)

func GenerateBuildArtifactStage(dimgConfig interface{}, extra *builder.Extra) Interface {
	b := getBuilder(dimgConfig, extra)
	if b != nil && !b.IsBuildArtifactEmpty() {
		return newBuildArtifactStage(b)
	}

	return nil
}

func newBuildArtifactStage(builder builder.Builder) *BuildArtifactStage {
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
