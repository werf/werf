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
	s.UserWithGAPatchStage = newUserWithGAPatchStage(builder, baseStageOptions)
	return s
}

type SetupStage struct {
	*UserWithGAPatchStage
}

func (s *SetupStage) Name() StageName {
	return Setup
}

func (s *SetupStage) GetDependencies(_ Conveyor, _ image.Image) (string, error) {
	stageDependenciesChecksum, err := s.GetStageDependenciesChecksum(Setup)
	if err != nil {
		return "", err
	}

	return util.Sha256Hash(s.builder.SetupChecksum(), stageDependenciesChecksum), nil
}

func (s *SetupStage) PrepareImage(c Conveyor, prevBuiltImage, image image.Image) error {
	if err := s.UserWithGAPatchStage.PrepareImage(c, prevBuiltImage, image); err != nil {
		return nil
	}

	if err := s.builder.Setup(image.BuilderContainer()); err != nil {
		return err
	}

	return nil
}
