package stage

import (
	"github.com/flant/dapp/pkg/build/builder"
	"github.com/flant/dapp/pkg/config"
)

func GenerateBuildArtifactStage(dimgConfig config.DimgInterface, extra *builder.Extra) Interface {
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

func (s *BuildArtifactStage) Name() StageName {
	return BuildArtifact
}

func (s *BuildArtifactStage) GetContext(_ Cache) string {
	return s.builder.BuildArtifactChecksum() // TODO: git
}
