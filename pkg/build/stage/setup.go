package stage

import (
	"github.com/flant/werf/pkg/build/builder"
	"github.com/flant/werf/pkg/config"
	"github.com/flant/werf/pkg/image"
	"github.com/flant/werf/pkg/util"
)

func GenerateSetupStage(dimgBaseConfig *config.DimgBase, gitPatchStageOptions *NewGitPatchStageOptions, baseStageOptions *NewBaseStageOptions) *SetupStage {
	b := getBuilder(dimgBaseConfig, baseStageOptions)
	if b != nil && !b.IsSetupEmpty() {
		return newSetupStage(b, gitPatchStageOptions, baseStageOptions)
	}

	return nil
}

func newSetupStage(builder builder.Builder, gitPatchStageOptions *NewGitPatchStageOptions, baseStageOptions *NewBaseStageOptions) *SetupStage {
	s := &SetupStage{}
	s.UserWithGitPatchStage = newUserWithGitPatchStage(builder, Setup, gitPatchStageOptions, baseStageOptions)
	return s
}

type SetupStage struct {
	*UserWithGitPatchStage
}

func (s *SetupStage) GetDependencies(_ Conveyor, _ image.Image) (string, error) {
	stageDependenciesChecksum, err := s.getStageDependenciesChecksum(Setup)
	if err != nil {
		return "", err
	}

	return util.Sha256Hash(s.builder.SetupChecksum(), stageDependenciesChecksum), nil
}

func (s *SetupStage) PrepareImage(c Conveyor, prevBuiltImage, image image.Image) error {
	if err := s.UserWithGitPatchStage.PrepareImage(c, prevBuiltImage, image); err != nil {
		return nil
	}

	if err := s.builder.Setup(image.BuilderContainer()); err != nil {
		return err
	}

	return nil
}
