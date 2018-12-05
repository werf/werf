package stage

import (
	"github.com/flant/dapp/pkg/build/builder"
	"github.com/flant/dapp/pkg/config"
	"github.com/flant/dapp/pkg/image"
	"github.com/flant/dapp/pkg/util"
)

func GenerateBeforeSetupStage(dimgBaseConfig *config.DimgBase, extra *builder.Extra, baseStageOptions *NewBaseStageOptions) *BeforeSetupStage {
	b := getBuilder(dimgBaseConfig, extra)
	if b != nil && !b.IsBeforeSetupEmpty() {
		return newBeforeSetupStage(b, baseStageOptions)
	}

	return nil
}

func newBeforeSetupStage(builder builder.Builder, baseStageOptions *NewBaseStageOptions) *BeforeSetupStage {
	s := &BeforeSetupStage{}
	s.UserWithGAPatchStage = newUserWithGAPatchStage(builder, baseStageOptions)
	return s
}

type BeforeSetupStage struct {
	*UserWithGAPatchStage
}

func (s *BeforeSetupStage) Name() StageName {
	return BeforeSetup
}

func (s *BeforeSetupStage) GetDependencies(_ Conveyor, _ image.Image) (string, error) {
	stageDependenciesChecksum, err := s.GetStageDependenciesChecksum(BeforeSetup)
	if err != nil {
		return "", err
	}

	return util.Sha256Hash(s.builder.BeforeSetupChecksum(), stageDependenciesChecksum), nil
}

func (s *BeforeSetupStage) PrepareImage(c Conveyor, prevBuiltImage, image image.Image) error {
	if err := s.UserWithGAPatchStage.PrepareImage(c, prevBuiltImage, image); err != nil {
		return nil
	}

	if err := s.builder.BeforeSetup(image.BuilderContainer()); err != nil {
		return err
	}

	return nil
}
