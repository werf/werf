package stage

import (
	"github.com/flant/dapp/pkg/build/builder"
	"github.com/flant/dapp/pkg/config"
	"github.com/flant/dapp/pkg/image"
	"github.com/flant/dapp/pkg/util"
)

func GenerateSetupStage(dimgConfig config.DimgInterface, extra *builder.Extra, baseStageOptions *NewBaseStageOptions) *SetupStage {
	b := getBuilder(dimgConfig, extra)
	if b != nil && !b.IsSetupEmpty() {
		return newSetupStage(b, baseStageOptions)
	}

	return nil
}

func newSetupStage(builder builder.Builder, baseStageOptions *NewBaseStageOptions) *SetupStage {
	s := &SetupStage{}
	s.UserStage = newUserStage(builder, baseStageOptions)
	return s
}

type SetupStage struct {
	*UserStage
}

func (s *SetupStage) Name() StageName {
	return Setup
}

func (s *SetupStage) GetContext(_ Conveyor) (string, error) {
	stageDependenciesChecksum, err := s.GetStageDependenciesChecksum(Setup)
	if err != nil {
		return "", err
	}

	return util.Sha256Hash(s.builder.SetupChecksum(), stageDependenciesChecksum), nil
}

func (s *SetupStage) PrepareImage(prevBuiltImage, image image.Image) error {
	if err := s.BaseStage.PrepareImage(prevBuiltImage, image); err != nil {
		return err
	}

	if err := s.builder.Setup(image.BuilderContainer()); err != nil {
		return err
	}

	return nil
}
