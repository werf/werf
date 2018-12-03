package stage

import (
	"github.com/flant/dapp/pkg/build/builder"
	"github.com/flant/dapp/pkg/config"
	"github.com/flant/dapp/pkg/image"
	"github.com/flant/dapp/pkg/util"
)

func GenerateBuildArtifactStage(dimgConfig config.DimgInterface, extra *builder.Extra, baseStageOptions *NewBaseStageOptions) *BuildArtifactStage {
	b := getBuilder(dimgConfig, extra)
	if b != nil && !b.IsBuildArtifactEmpty() {
		return newBuildArtifactStage(b, baseStageOptions)
	}

	return nil
}

func newBuildArtifactStage(builder builder.Builder, baseStageOptions *NewBaseStageOptions) *BuildArtifactStage {
	s := &BuildArtifactStage{}
	s.UserStage = newUserStage(builder, baseStageOptions)
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

func (s *BuildArtifactStage) PrepareImage(prevBuiltImage, image image.Image) error {
	if err := s.BaseStage.PrepareImage(prevBuiltImage, image); err != nil {
		return err
	}

	if err := s.builder.BuildArtifact(image.BuilderContainer()); err != nil {
		return err
	}

	return nil
}
