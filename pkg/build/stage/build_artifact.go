package stage

import (
	"github.com/flant/dapp/pkg/build/builder"
	"github.com/flant/dapp/pkg/config"
	"github.com/flant/dapp/pkg/util"
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

func (s *BuildArtifactStage) GetContext(_ Conveyor) (string, error) {
	stageDependenciesChecksum, err := s.GetStageDependenciesChecksum(BuildArtifact)
	if err != nil {
		return "", err
	}

	return util.Sha256Hash(s.builder.BuildArtifactChecksum(), stageDependenciesChecksum), nil
}
