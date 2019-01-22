package stage

import (
	"github.com/flant/werf/pkg/build/builder"
	"github.com/flant/werf/pkg/config"
	"github.com/flant/werf/pkg/image"
	"github.com/flant/werf/pkg/util"
)

func GenerateBeforeSetupStage(imageBaseConfig *config.ImageBase, gitPatchStageOptions *NewGitPatchStageOptions, baseStageOptions *NewBaseStageOptions) *BeforeSetupStage {
	b := getBuilder(imageBaseConfig, baseStageOptions)
	if b != nil && !b.IsBeforeSetupEmpty() {
		return newBeforeSetupStage(b, gitPatchStageOptions, baseStageOptions)
	}

	return nil
}

func newBeforeSetupStage(builder builder.Builder, gitPatchStageOptions *NewGitPatchStageOptions, baseStageOptions *NewBaseStageOptions) *BeforeSetupStage {
	s := &BeforeSetupStage{}
	s.UserWithGitPatchStage = newUserWithGitPatchStage(builder, BeforeSetup, gitPatchStageOptions, baseStageOptions)
	return s
}

type BeforeSetupStage struct {
	*UserWithGitPatchStage
}

func (s *BeforeSetupStage) GetDependencies(_ Conveyor, _ image.ImageInterface) (string, error) {
	stageDependenciesChecksum, err := s.getStageDependenciesChecksum(BeforeSetup)
	if err != nil {
		return "", err
	}

	return util.Sha256Hash(s.builder.BeforeSetupChecksum(), stageDependenciesChecksum), nil
}

func (s *BeforeSetupStage) PrepareImage(c Conveyor, prevBuiltImage, image image.ImageInterface) error {
	if err := s.UserWithGitPatchStage.PrepareImage(c, prevBuiltImage, image); err != nil {
		return nil
	}

	if err := s.builder.BeforeSetup(image.BuilderContainer()); err != nil {
		return err
	}

	return nil
}
